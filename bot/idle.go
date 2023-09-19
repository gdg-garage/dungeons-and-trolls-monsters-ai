package bot

import (
	"math/rand"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

func (b *Bot) randomWalk() *swagger.DungeonsandtrollsCommandsBatch {
	// get random direction
	for {
		distanceX := rand.Intn(4)
		distanceY := rand.Intn(4)
		newX := int(b.GameState.CurrentPosition.PositionX) + distanceX
		newY := int(b.GameState.CurrentPosition.PositionY) + distanceY
		currentLevel := b.GameState.CurrentLevel
		currentMap := b.GameState.Map_.Levels[currentLevel]

		for _, objects := range currentMap.Objects {
			if int(objects.Position.PositionX) == newX && int(objects.Position.PositionY) == newY {
				if objects.IsFree {
					return &swagger.DungeonsandtrollsCommandsBatch{
						Move: &swagger.DungeonsandtrollsCoordinates{
							PositionX: int32(newX),
							PositionY: int32(newY),
						},
					}
				}
			}
		}
	}
}
