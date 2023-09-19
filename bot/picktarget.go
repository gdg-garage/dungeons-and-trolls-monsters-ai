package bot

import (
	"math/rand"
)

// Enemies

func (b *Bot) getEnemies(mapObjects *MapObjectsByCategory) []MapObject {
	enemies := []MapObject{}
	for _, mapObject := range mapObjects.Players {
		if !b.IsFriendly(mapObject) {
			enemies = append(enemies, mapObject)
		}
	}
	for _, mapObject := range mapObjects.Monsters {
		if !b.IsFriendly(mapObject) {
			enemies = append(enemies, mapObject)
		}
	}
	return enemies
}

// Friendly

func (b *Bot) getFriendly(objects *MapObjectsByCategory) []MapObject {
	friends := []MapObject{}
	for _, mapObject := range objects.Players {
		if b.IsFriendly(mapObject) {
			friends = append(friends, mapObject)
		}
	}
	for _, mapObject := range objects.Monsters {
		if b.IsFriendly(mapObject) {
			friends = append(friends, mapObject)
		}
	}
	return friends
}

func (b *Bot) pickTarget(objects *MapObjectsByCategory) *MapObject {
	return b.pickClosestTarget(b.getEnemies(objects))
}

func (b *Bot) pickRandomTarget(enemies []MapObject) *MapObject {
	// get random object
	x := rand.Intn(len(enemies))
	return &enemies[x]
}

func (b *Bot) pickClosestTarget(enemies []MapObject) *MapObject {
	// TODO: implement
	return &enemies[0]
}
