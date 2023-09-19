package bot

import (
	"math/rand"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

func filterOutFriendly(objects *objectsByCategory) []swagger.DungeonsandtrollsMapObjects {
	// TODO: pick Players plus other monsters based on factions
	return objects.Monsters
}

func pickTarget(objects *objectsByCategory) *swagger.DungeonsandtrollsMapObjects {
	return pickClosestTarget(filterOutFriendly(objects))
}

func pickRandomTarget(enemies []swagger.DungeonsandtrollsMapObjects) *swagger.DungeonsandtrollsMapObjects {
	// get random object
	x := rand.Intn(len(enemies))
	return &enemies[x]
}

func pickClosestTarget(enemies []swagger.DungeonsandtrollsMapObjects) *swagger.DungeonsandtrollsMapObjects {
	// TODO: implement
	return &enemies[0]
}
