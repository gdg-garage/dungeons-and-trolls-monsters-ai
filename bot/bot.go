package bot

import (
	"log"
	"math/rand"
	"time"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
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
}

func New(state *swagger.DungeonsandtrollsGameState, botID string, existingBots BotMemory) *Bot {
	// check if bot exists
	if bot, ok := existingBots[botID]; ok {
		// bot exists, update state, return bot
		bot.PrevGameState = bot.GameState
		bot.GameState = state
		return bot
	}

	// create new bot
	bot := &Bot{
		BotState:      BotState{},
		GameState:     state,
		PrevGameState: nil,
	}
	existingBots[botID] = bot
	return bot
}

func (b *Bot) Run3() *swagger.DungeonsandtrollsCommandsBatch {
	b.updateMood()

	state := *b.GameState
	score := state.Score
	log.Println("Score:", score)
	log.Println("Character.Money:", state.Character.Money)
	log.Println("CurrentPosition:", state.CurrentPosition)
	if len(state.Character.Equip) > 0 {
		log.Printf("Character.Equip: %+v\n", state.Character.Equip[0])
	}

	var mainHandItem *swagger.DungeonsandtrollsItem
	for _, item := range state.Character.Equip {
		if *item.Slot == swagger.MAIN_HAND_DungeonsandtrollsItemType {
			mainHandItem = &item
			break
		}
	}

	if mainHandItem == nil {
		log.Println("Looking for items to buy ...")
		item := shop(&state)
		if item != nil {
			return &swagger.DungeonsandtrollsCommandsBatch{
				Buy: &swagger.DungeonsandtrollsIdentifiers{Ids: []string{item.Id}},
			}
		}
		log.Println("ERROR: Found no item to buy!")
	}

	objects := b.getMapObjectsByCategory()
	log.Println("Stairs position:", objects.Stairs.Position)

	// Add seed
	rand.Seed(time.Now().UnixNano())
	random := rand.Intn(8)
	log.Println("Random:", random)
	if random <= 1 {
		log.Println("Picking a random yell ...")
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
		log.Println("Yelling:", yell)
		return &swagger.DungeonsandtrollsCommandsBatch{
			Yell: &swagger.DungeonsandtrollsMessage{
				Text: yell,
			},
		}
	}

	// TODO: Get all enemies
	// TODO: Get all friends
	// TODO: Get all neutral

	if len(objects.Monsters) > 0 {
		// for _, monster := range objects.Monsters {
		// 	log.Printf("Monster: %+v\n", monster)
		// }
		log.Println("Let's fight!")
		log.Println("Picking a target ...")
		target := b.pickTarget(&objects)
		if target != nil {
			// log.Printf("Target: %+v\n", target)
			log.Printf("Target coords: %+v\n", target.MapObjects.Position)

			if target.MapObjects.Position.PositionX == state.CurrentPosition.PositionX && target.MapObjects.Position.PositionY == state.CurrentPosition.PositionY {
				log.Println("Picking a skill ...")
				skill := b.pickSkill()
				log.Printf("Picked skill: %+v\n", skill)

				log.Println("Attacking target ...")

				return useSkill(*skill, *target)
			} else {
				log.Println("Moving towards target ...")
				return &swagger.DungeonsandtrollsCommandsBatch{
					Move: target.MapObjects.Position,
				}
			}
		}
	}

	if objects.Stairs == nil {
		log.Println("Can't find stairs")
		return &swagger.DungeonsandtrollsCommandsBatch{
			Yell: &swagger.DungeonsandtrollsMessage{
				Text: "Where are the stairs? I can't find them!",
			},
		}
	}

	log.Println("Moving towards stairs ...")
	return &swagger.DungeonsandtrollsCommandsBatch{
		Move: objects.Stairs.Position,
	}
}

func shop(state *swagger.DungeonsandtrollsGameState) *swagger.DungeonsandtrollsItem {
	shop := state.ShopItems
	for _, item := range shop {
		if item.Price <= state.Character.Money {
			if *item.Slot == swagger.MAIN_HAND_DungeonsandtrollsItemType {
				if len(item.Skills) > 0 {
					if item.Skills[0].DamageAmount.Scalar > 0 {
						log.Println("Chosen item:", item.Name)
						return &item
					}
				}
			}
		}
	}
	return nil
}
