package bot

import (
	"context"
	"sync"
	"time"

	"github.com/antihax/optional"
	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
	"github.com/gdg-garage/dungeons-and-trolls-monsters-ai/swaggerutil"
	"go.uber.org/zap"
)

type MonsterDetails struct {
	Id         string
	Name       string
	Level      int32
	Index      int
	Position   *swagger.DungeonsandtrollsPosition
	Monster    *swagger.DungeonsandtrollsMonster
	MapObjects *swagger.DungeonsandtrollsMapObjects
	CurrentMap *swagger.DungeonsandtrollsLevel
}

type BotDispatcher struct {
	Client        *swagger.APIClient
	Ctx           context.Context
	Bots          map[string]*Bot
	BotsLock      sync.Mutex
	Logger        *zap.SugaredLogger
	LoggerWTick   *zap.SugaredLogger
	TickStartTime time.Time
}

func NewBotDispatcher(client *swagger.APIClient, ctx context.Context, logger *zap.SugaredLogger) *BotDispatcher {
	return &BotDispatcher{
		Client:   client,
		Ctx:      ctx,
		Bots:     make(map[string]*Bot),
		BotsLock: sync.Mutex{},
		Logger:   logger,
	}
}

func (d *BotDispatcher) HandleTick(gameState *swagger.DungeonsandtrollsGameState, tickStartTime time.Time) error {
	d.TickStartTime = tickStartTime
	d.LoggerWTick = d.Logger.With(
		"tick", gameState.Tick,
		"tickStartTime", tickStartTime,
	)

	for _, level := range gameState.Map_.Levels {
		// go d.HandleLevel(gameState, level)
		err := d.HandleLevel(gameState, level)
		if err != nil {
			d.LoggerWTick.Error("Error when running monster AI for level",
				zap.Error(err),
				zap.Int32("mapLevel", level.Level),
			)
		}
	}
	return nil
}

func (d *BotDispatcher) HandleLevel(gameState *swagger.DungeonsandtrollsGameState, level swagger.DungeonsandtrollsLevel) error {
	monsters := getMonstersDetailsForLevel(gameState, &level)
	d.LoggerWTick.Infow("Handling level",
		"mapLevel", level.Level,
		"monstersCount", len(monsters),
	)
	commands := swagger.DungeonsandtrollsCommandsForMonsters{}
	commands.Commands = make(map[string]swagger.DungeonsandtrollsCommandsBatch)
	for i := range monsters {
		monster := monsters[i]
		botLogger := d.LoggerWTick.With(
			"monsterId", monster.Id,
			"monsterName", monster.Name,
			"mapLevel", monster.Level,
		)
		d.BotsLock.Lock()
		bot, found := d.Bots[monster.Id]
		if !found {
			// initialize bot / new monster
			bot = &Bot{
				MonsterId: monster.Id,
				BotState:  BotState{},
			}
			d.Bots[monster.Id] = bot
		} else {
			// copy previous state
			bot.PrevBotState = bot.BotState
			bot.PrevGameState = bot.GameState
			bot.PrevDetails = bot.Details
		}
		d.BotsLock.Unlock()
		bot.Logger = botLogger
		bot.GameState = gameState
		bot.Details = monster
		cmd := bot.Run()
		cmd = bot.constructYellCommand(cmd)
		if cmd != nil {
			commands.Commands[monster.Id] = *cmd
			// XXX: send individually
			sendIndividually := false
			sendAsynchronously := true
			if sendIndividually {
				commandsCopy := swagger.DungeonsandtrollsCommandsForMonsters{
					Commands: make(map[string]swagger.DungeonsandtrollsCommandsBatch),
				}
				for k, v := range commands.Commands {
					commandsCopy.Commands[k] = v
				}
				if sendAsynchronously {
					go d.sendMonsterCommands(commandsCopy, botLogger)
				} else {
					d.sendMonsterCommands(commandsCopy, botLogger)
				}
				commands.Commands = make(map[string]swagger.DungeonsandtrollsCommandsBatch)
			}
		}
	}
	if len(commands.Commands) > 0 {
		loggerWLevel := d.LoggerWTick.With(
			"mapLevel", level,
		)
		go d.sendMonsterCommands(commands, loggerWLevel)
	}
	return nil
}

func getMonstersDetailsForLevel(state *swagger.DungeonsandtrollsGameState, level *swagger.DungeonsandtrollsLevel) []MonsterDetails {
	currentMap := level
	monsters := []MonsterDetails{}
	for i := range currentMap.Objects {
		object := currentMap.Objects[i]
		for j := range object.Monsters {
			details := MonsterDetails{
				Id:         object.Monsters[j].Id,
				Name:       object.Monsters[j].Name,
				Level:      level.Level,
				CurrentMap: currentMap,
				Index:      j,
				Position:   object.Position,
				Monster:    &object.Monsters[j],
				MapObjects: &object,
			}
			monsters = append(monsters, details)
		}
	}
	return monsters
}

func (d *BotDispatcher) sendMonsterCommands(cmds swagger.DungeonsandtrollsCommandsForMonsters, logger *zap.SugaredLogger) error {
	opts := swagger.DungeonsAndTrollsApiDungeonsAndTrollsMonstersCommandsOpts{
		Blocking: optional.NewBool(false),
	}
	_, httpResp, err := d.Client.DungeonsAndTrollsApi.DungeonsAndTrollsMonstersCommands(d.Ctx, cmds, &opts)
	tickDuration := time.Since(d.TickStartTime)
	logger2 := logger.With(
		"tickDurationSeconds", tickDuration,
	)
	swaggerutil.LogResponse(logger2, err, httpResp, "MonsterCommands", cmds)
	return nil
}
