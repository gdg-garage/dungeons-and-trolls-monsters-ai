package bot

import (
	"log"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

/*
	Position *DungeonsandtrollsCoordinates `json:"position,omitempty"`
	Monsters []DungeonsandtrollsMonster `json:"monsters,omitempty"`
	Players []DungeonsandtrollsCharacter `json:"players,omitempty"`
	IsStairs bool `json:"isStairs,omitempty"`
	Portal *DungeonsandtrollsWaypoint `json:"portal,omitempty"`
	Decorations []DungeonsandtrollsDecoration `json:"decorations,omitempty"`
	Effects []DungeonsandtrollsEffect `json:"effects,omitempty"`
	Items []DungeonsandtrollsItem `json:"items,omitempty"`
	IsFree bool `json:"isFree,omitempty"`
	IsWall bool `json:"isWall,omitempty"`
	IsDoor bool `json:"isDoor,omitempty"`
	IsSpawn bool `json:"isSpawn,omitempty"`
*/

const (
	MapObjectTypePlayer  = "player"
	MapObjectTypeMonster = "monster"

	// MapObjectTypePortal     = "portal"
	// MapObjectTypeDecoration = "decoration"
	MapObjectTypeEffect = "effect"
	// MapObjectTypeItem   = "item"
)

type MapObject struct {
	MapObjects swagger.DungeonsandtrollsMapObjects
	Type       string
	Index      int
}

func (mo MapObject) GetId() string {
	switch mo.Type {
	case MapObjectTypePlayer:
		return mo.MapObjects.Players[mo.Index].Id
	case MapObjectTypeMonster:
		return mo.MapObjects.Monsters[mo.Index].Id
	case MapObjectTypeEffect:
		log.Println("ERROR: MapObject.GetId(): Can't get ID for Effect")
		return ""
	default:
		log.Println("ERROR: MapObject.GetId(): Unknown type")
		return ""
	}

}

func (mo MapObject) GetIdentifier() *swagger.DungeonsandtrollsIdentifier {
	return &swagger.DungeonsandtrollsIdentifier{
		Id: mo.GetId(),
	}
}

func (mo MapObject) GetFaction() string {
	if mo.Type == MapObjectTypePlayer {
		return "player"
	}
	if mo.Type != MapObjectTypeMonster {
		log.Println("ERROR: MapObject.GetFaction(): Unknown type")
		return "unknown"
	}
	return "monster"
	// TODO: get faction of mo
	// return mo.MapObjects.Monsters[mo.Index].Faction
}

func (b *Bot) IsFriendly(mo MapObject) bool {
	// TODO: get my faction
	// myFaction := b.GameState.Character.Faction
	myFaction := "player"
	faction := mo.GetFaction()
	if faction == myFaction || faction == "neutral" {
		return true
	}
	switch myFaction {
	case "player":
		return faction == "templar"
	case "monster":
		return faction == "outlaw" || faction == "horror"
	case "outlaw":
		return faction == "monster"
	case "templar":
		return faction == "player"
	default:
		log.Println("ERROR: IsFriendly(): Unknown faction")
		return true
	}
}

func NewPlayerMapObject(mapObjects swagger.DungeonsandtrollsMapObjects, index int) MapObject {
	if len(mapObjects.Players) <= index {
		log.Println("ERROR: New MapObject: Index out of range for Players")
	}
	return MapObject{
		MapObjects: mapObjects,
		Type:       MapObjectTypePlayer,
		Index:      index,
	}
}

func NewMonsterMapObject(mapObjects swagger.DungeonsandtrollsMapObjects, index int) MapObject {
	if len(mapObjects.Monsters) <= index {
		log.Println("ERROR: New MapObject: Index out of range for Monsters")
	}
	return MapObject{
		MapObjects: mapObjects,
		Type:       MapObjectTypeMonster,
		Index:      index,
	}
}

func NewEffectMapObject(mapObjects swagger.DungeonsandtrollsMapObjects, index int) MapObject {
	if len(mapObjects.Effects) <= index {
		log.Print("ERROR: New MapObject: Index out of range for Effects")
	}
	return MapObject{
		MapObjects: mapObjects,
		Type:       MapObjectTypeEffect,
		Index:      index,
	}
}

func (b *Bot) getMapObjectsByCategory() MapObjectsByCategory {
	state := *b.GameState
	level := state.CurrentPosition.Level
	currentMap := state.Map_.Levels[level]
	objects := MapObjectsByCategory{}
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
			for i := range object.Players {
				objects.Players = append(objects.Players, NewPlayerMapObject(object, i))
			}
		}
		if len(object.Monsters) > 0 {
			for i := range object.Monsters {
				objects.Monsters = append(objects.Monsters, NewMonsterMapObject(object, i))
			}
		}
		if len(object.Effects) > 0 {
			for i := range object.Effects {
				objects.Effects = append(objects.Effects, NewEffectMapObject(object, i))
			}
		}
		// Maybe TODO (e.g. monsters guarding portals)
		// if len(object.Portals) > 0 {
	}
	return objects
}

type MapObjectsByCategory struct {
	Spawn    *swagger.DungeonsandtrollsMapObjects
	Stairs   *swagger.DungeonsandtrollsMapObjects
	Players  []MapObject
	Monsters []MapObject
	Effects  []MapObject
	// Portals  []MapObject
}
