package bot

import (
	"log"
	"math/rand"
	"time"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

const (
	IdleState = "idle"
	// ExploreState = "explore"
	AgroState    = "agro"
	AttackState  = "attack"
	DefendState  = "defend"
	SupportState = "support"
	FleeState    = "flee"
)

type BotState struct {
	State        string
	TargetCoords swagger.DungeonsandtrollsCoordinates
	// ?
	// TargetId
}

var botState = BotState{
	State: IdleState,
}

func Run2(state swagger.DungeonsandtrollsGameState) *swagger.DungeonsandtrollsCommandsBatch {
	// id := "TODO"

	switch botState.State {
	case IdleState:
		// = no enemies
		// enemies nearby -> agro
	case AgroState:
		// = aware of enemies
		// attempt to target enemy -> attack
		// no enemies -> idle
		// maybe -> support
		// timeout -> idle
	case AttackState:
		// = melee
		// target dead, no enemies -> idle
		// timeout -> change target
		// better target -> change target
		// ally in need -> support / defend
		// low on health -> flee
		// no allies -> flee
	case DefendState:
		// = fight alongside ally
	case SupportState:
		// = ranged, heal, buff, etc.
	case FleeState:
		// = run away
		// no enemies -> idle
		// allies nearby -> support
		// timeout -> idle
	}
	return nil
}

func Run(state swagger.DungeonsandtrollsGameState) *swagger.DungeonsandtrollsCommandsBatch {
	score := state.Score
	log.Println("Score:", score)
	log.Println("Character.Money:", state.Character.Money)
	log.Println("CurrentPosition:", state.CurrentPosition)
	log.Println("Character.Equip:", state.Character.Equip)

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

	objects := getObjectsByCategory(&state)
	log.Println("Stairs position:", objects.Stairs.Position)

	if len(objects.Monsters) > 0 {
		for _, monster := range objects.Monsters {
			log.Printf("Monster: %+v\n", monster)
		}
		log.Println("Let's fight!")
		// TODO: Use Skill if monster in range
		return &swagger.DungeonsandtrollsCommandsBatch{
			Move: objects.Monsters[0].Position,
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

	// Add seed
	rand.Seed(time.Now().UnixNano())
	random := rand.Intn(6)
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
		Move: objects.Stairs.Position,
	}
}

func shop(state *swagger.DungeonsandtrollsGameState) *swagger.DungeonsandtrollsItem {
	shop := state.ShopItems
	for _, item := range shop {
		if item.Price <= state.Character.Money {
			if *item.Slot == swagger.MAIN_HAND_DungeonsandtrollsItemType {
				log.Println("Chosen item:", item.Name)
				return &item
			}
		}
	}
	return nil
}

func getObjectsByCategory(state *swagger.DungeonsandtrollsGameState) objectsByCategory {
	level := state.CurrentPosition.Level
	currentMap := state.Map_.Levels[level]
	objects := objectsByCategory{}
	for i := range currentMap.Objects {
		// get references to objects
		object := currentMap.Objects[i]
		if object.IsStairs {
			log.Printf("Found stairs: %+v\n", object)
			log.Printf("Stairs coords: %+v\n", object.Position)
			objects.Stairs = &object
		}
		if object.IsSpawn {
			objects.Spawn = &object
		}
		if len(object.Players) > 0 {
			objects.Players = append(objects.Players, object)
		}
		if len(object.Monsters) > 0 {
			objects.Monsters = append(objects.Monsters, object)
		}
	}
	return objects
}

type objectsByCategory struct {
	Spawn    *swagger.DungeonsandtrollsMapObjects
	Stairs   *swagger.DungeonsandtrollsMapObjects
	Players  []swagger.DungeonsandtrollsMapObjects
	Monsters []swagger.DungeonsandtrollsMapObjects
	// TODO: split into monster factions
	// Add portals, effects, etc.
}
