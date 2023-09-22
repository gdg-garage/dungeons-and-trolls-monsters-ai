package bot

import (
	"math/rand"
	"time"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
	"go.uber.org/zap"
)

type BotMemory map[string]*Bot

type BotState struct {
	State        string
	TargetObject swagger.DungeonsandtrollsMapObjects
	Target       swagger.DungeonsandtrollsMonster
	Mood         Mood
}

type Bot struct {
	BotState BotState
	// PrevBotState  BotState
	GameState     *swagger.DungeonsandtrollsGameState
	PrevGameState *swagger.DungeonsandtrollsGameState
	// We can add more fields here
	Logger *zap.SugaredLogger
}

func New(state *swagger.DungeonsandtrollsGameState, botID string, existingBots BotMemory, logger *zap.SugaredLogger) *Bot {
	loggerWID := logger.With(zap.String("botID", botID))
	// check if bot exists
	if bot, ok := existingBots[botID]; ok {
		// bot exists, update state, return bot
		bot.PrevGameState = bot.GameState
		bot.GameState = state
		// Maybe don't update logger every tick
		// Update tick value instead
		bot.Logger = loggerWID
		return bot
	}

	// create new bot
	bot := &Bot{
		BotState:      BotState{},
		GameState:     state,
		PrevGameState: nil,
		Logger:        loggerWID,
	}
	existingBots[botID] = bot
	return bot
}

func (b *Bot) Run3() *swagger.DungeonsandtrollsCommandsBatch {
	b.updateMood()

	state := *b.GameState
	score := state.Score
	b.Logger.Infow("Game state: Character info",
		"Name", state.Character.Name,
		"Score", score,
		"Money", state.Character.Money,
		"CurrentLevel", state.CurrentLevel,
		"CurrentPosition", state.CurrentPosition,
	)

	var mainHandItem *swagger.DungeonsandtrollsItem
	for _, item := range state.Character.Equip {
		if *item.Slot == swagger.MAIN_HAND_DungeonsandtrollsItemType {
			mainHandItem = &item
			break
		}
	}

	if mainHandItem == nil {
		b.Logger.Debug("Looking for items to buy ...")
		item := b.shop()
		if item != nil {
			return &swagger.DungeonsandtrollsCommandsBatch{
				Buy: &swagger.DungeonsandtrollsIdentifiers{Ids: []string{item.Id}},
			}
		}
		b.Logger.Warn("ERROR: Found no item to buy!")
	}

	objects := b.getMapObjectsByCategory()
	b.Logger.Debugw("Stairs position ...",
		"stairsPosition", objects.Stairs.Position,
	)

	// Add seed
	rand.Seed(time.Now().UnixNano())
	random := rand.Intn(8)
	if random <= 1 {
		b.Logger.Debug("Picking a random yell ...")
		randomYell := rand.Intn(8)
		var yells []string = []string{
			"Anybody home?",
			"What was that?",
			"Hey! Show yourself!",
			"Yo!",
			"500 error",
			"Your mom is a nice lady",
			"You will never find me",
			"Come and get me",
			"8 ball says: no",
		}
		yell := yells[randomYell]
		b.Logger.Debugw("Yelling ...",
			"yell", yell,
		)
		return &swagger.DungeonsandtrollsCommandsBatch{
			Yell: &swagger.DungeonsandtrollsMessage{
				Text: yell,
			},
		}
	}

	if len(objects.Monsters) > 0 {
		// for _, monster := range objects.Monsters {
		// 	log.Printf("Monster: %+v\n", monster)
		// }
		b.Logger.Debug("Let's fight!")
		b.Logger.Debug("Picking a target ...")
		target := b.pickTarget(&objects)
		if target != nil {
			// log.Printf("Target: %+v\n", target)
			b.Logger.Debugw("Picked target",
				"targetPosition", target.MapObjects.Position,
			)

			if target.MapObjects.Position.PositionX == state.CurrentPosition.PositionX && target.MapObjects.Position.PositionY == state.CurrentPosition.PositionY {
				b.Logger.Debug("Picking a skill ...")
				skill := b.pickSkill()
				b.Logger.Debugw("Picked skill",
					"skillName", skill.Name,
					"skill", skill,
				)

				b.Logger.Debugw("Attacking target ...",
					"targetPosition", target.MapObjects.Position,
					"targetName", target.GetName(),
				)

				return useSkill(*skill, *target)
			} else {
				b.Logger.Debugw("Moving towards target ...",
					"targetPosition", target.MapObjects.Position,
					"targetName", target.GetName(),
				)
				return &swagger.DungeonsandtrollsCommandsBatch{
					Move: target.MapObjects.Position,
				}
			}
		}
	}

	if objects.Stairs == nil {
		b.Logger.Warn("Can't find stairs!")
		return &swagger.DungeonsandtrollsCommandsBatch{
			Yell: &swagger.DungeonsandtrollsMessage{
				Text: "Where are the stairs? I can't find them!",
			},
		}
	}

	b.Logger.Infow("Moving towards stairs ...",
		"stairsPosition", objects.Stairs.Position,
	)
	return &swagger.DungeonsandtrollsCommandsBatch{
		Move: objects.Stairs.Position,
	}
}

func (b *Bot) shop() *swagger.DungeonsandtrollsItem {
	shop := b.GameState.ShopItems
	money := b.GameState.Character.Money
	for _, item := range shop {
		if item.Price <= money {
			if *item.Slot == swagger.MAIN_HAND_DungeonsandtrollsItemType {
				if len(item.Skills) > 0 {
					if item.Skills[0].DamageAmount.Scalar > 0 {
						b.Logger.Infow("Found item to buy ...",
							"itemName", item.Name,
						)
						return &item
					}
				}
			}
		}
	}
	return nil
}
