package bot

import (
	"math"
	"math/rand"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

func (b *Bot) randomWalk() *swagger.DungeonsandtrollsCommandsBatch {
	pos := b.Details.Position
	// get random direction
	for i := 0; i < 20; i++ {
		distanceX := rand.Intn(8) - 4
		distanceY := rand.Intn(8) - 4
		newX := int(pos.PositionX) + distanceX
		newY := int(pos.PositionY) + distanceY

		tileInfo, found := b.BotState.MapExtended[makePosition(int32(newX), int32(newY))]
		if !found || !tileInfo.mapObjects.IsFree || tileInfo.distance == math.MaxInt32 {
			// unreachable or not free
			continue
		}
		if len(tileInfo.mapObjects.Monsters) > 0 && i < 14 {
			// Prefer not to walk into other monsters
			continue
		}
		return &swagger.DungeonsandtrollsCommandsBatch{
			Move: &swagger.DungeonsandtrollsPosition{
				PositionX: int32(newX),
				PositionY: int32(newY),
			},
		}
	}
	b.Logger.Warnw("randomWalkFromPosition: No free position found")
	return &swagger.DungeonsandtrollsCommandsBatch{
		Move: &swagger.DungeonsandtrollsPosition{
			PositionX: int32(pos.PositionX),
			PositionY: int32(pos.PositionY),
		},
	}
}
