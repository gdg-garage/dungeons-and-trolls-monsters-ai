package bot

import (
	"context"
	"sync"

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
}

type BotDispatcher struct {
	Client      *swagger.APIClient
	Ctx         context.Context
	Bots        map[string]*Bot
	BotsLock    sync.Mutex
	Logger      *zap.SugaredLogger
	LoggerWTick *zap.SugaredLogger
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

func (d *BotDispatcher) HandleTick(gameState *swagger.DungeonsandtrollsGameState) error {
	d.LoggerWTick = d.Logger.With("tick", gameState.Tick)

	levels := d.getLevels(gameState)
	for _, level := range levels {
		// go d.HandleLevel(gameState, level)
		err := d.HandleLevel(gameState, level)
		if err != nil {
			d.LoggerWTick.Error("Error when running monster AI for level",
				zap.Error(err),
				zap.Int32("mapLevel", level),
			)
		}
	}
	return nil
}

func (d *BotDispatcher) HandleLevel(gameState *swagger.DungeonsandtrollsGameState, level int32) error {
	monsters := getMonstersDetailsForLevel(gameState, level)
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
		cmd := bot.Run5()
		if cmd != nil {
			commands.Commands[monster.Id] = *cmd
		}
	}
	if len(commands.Commands) > 0 {
		return d.sendMonsterCommands(commands)
	}
	return nil
}

func (d *BotDispatcher) getLevels(gameState *swagger.DungeonsandtrollsGameState) []int32 {
	var levels []int32
	for _, level := range gameState.Map_.Levels {
		levels = append(levels, level.Level)
	}
	return reverseListInt32(levels)
}

func reverseListInt32(a []int32) []int32 {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
	return a
}

func getMonstersDetailsForLevel(state *swagger.DungeonsandtrollsGameState, level int32) []MonsterDetails {
	currentMap := state.Map_.Levels[level]
	monsters := []MonsterDetails{}
	for i := range currentMap.Objects {
		object := currentMap.Objects[i]
		for j := range object.Monsters {
			details := MonsterDetails{
				Id:         object.Monsters[j].Id,
				Name:       object.Monsters[j].Name,
				Level:      level,
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

func (d *BotDispatcher) sendMonsterCommands(cmds swagger.DungeonsandtrollsCommandsForMonsters) error {
	opts := swagger.DungeonsAndTrollsApiDungeonsAndTrollsMonstersCommandsOpts{
		Blocking: optional.NewBool(false),
	}
	_, httpResp, err := d.Client.DungeonsAndTrollsApi.DungeonsAndTrollsMonstersCommands(d.Ctx, cmds, &opts)
	swaggerutil.LogResponse(d.LoggerWTick, err, httpResp, "MonsterCommands", cmds)
	return nil
}