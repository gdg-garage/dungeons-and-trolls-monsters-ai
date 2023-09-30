package bot

import (
	"math"
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
						Move: &swagger.DungeonsandtrollsPosition{
							PositionX: int32(newX),
							PositionY: int32(newY),
						},
					}
				}
			}
		}
	}
}

func (b *Bot) randomWalkFromPosition(level int32, pos swagger.DungeonsandtrollsPosition) *swagger.DungeonsandtrollsCommandsBatch {
	// get random direction
	for i := 0; i < 20; i++ {
		distanceX := rand.Intn(10) - 5
		distanceY := rand.Intn(10) - 5
		newX := int(pos.PositionX) + distanceX
		newY := int(pos.PositionY) + distanceY
		currentMap := b.GameState.Map_.Levels[level]

		isFree := true
		for _, objects := range currentMap.Objects {
			if int(objects.Position.PositionX) == newX && int(objects.Position.PositionY) == newY && !objects.IsFree {
				isFree = false
			}
			if !b.isInBounds(level, makePosition(int32(newX), int32(newY))) {
				isFree = false
			}
		}
		if !isFree {
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

func (b *Bot) randomWalkFromPositionExt(level int32, pos swagger.DungeonsandtrollsPosition, mapExt map[swagger.DungeonsandtrollsPosition]MapCellExt) *swagger.DungeonsandtrollsCommandsBatch {
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
