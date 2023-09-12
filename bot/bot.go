package bot

import (
	"log"
	"math/rand"
	"time"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

func Run(state swagger.DungeonsandtrollsGameState) *swagger.DungeonsandtrollsCommandsBatch {
	score := state.Score
	log.Println("Score:", score)
	log.Println("Character.Money:", state.Character.Money)
	log.Println("CurrentPosition:", state.CurrentPosition)
	log.Println("Character.Equip:", state.Character.Equip)

	// log.Println("Items in shop:")
	shop := state.ShopItems
	for _, item := range shop {
		if item.Price <= state.Character.Money {
			// log.Println("Can afford:", item.Name)
			// log.Printf("\t\tItem slot: %v, Price: %v\n", *item.Slot, item.BuyPrice)
			if *item.Slot == swagger.MAIN_HAND_DungeonsandtrollsItemType && false {
				log.Println("Found main hand item:", item.Name)
			}
		}
	}

	level := state.CurrentPosition.Level
	log.Println("Current level:", level)

	currentMap := state.Map_.Levels[level]
	log.Println("Current map level:", currentMap.Level)
	var stairsPosition *swagger.DungeonsandtrollsCoordinates

	//log.Printf("Current map: %+v\n", currentMap)

	for _, object := range currentMap.Objects {
		if object.IsStairs {
			log.Printf("Found stairs: %+v\n", object)
			log.Printf("Stairs coords: %+v\n", object.Position)
			stairsPosition = object.Position
		}
	}

	if stairsPosition == nil {
		log.Println("Can't find stairs")
		return nil
	}

	// Fix stairs level
	// stairsPosition.Level = level
	// log.Printf("Chosen fixed stairs coords: %+v\n", stairsPosition)

	// Add seed
	rand.Seed(time.Now().UnixNano())
	random := rand.Intn(3)
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

	log.Println("Moving towards stairs ...")
	return &swagger.DungeonsandtrollsCommandsBatch{
		Move: stairsPosition,
	}
	// log.Printf("Map: %+v\n", state.Map_)
	// stairsCoords := state.
}
