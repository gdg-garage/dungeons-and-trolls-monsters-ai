package bot

import (
	"math/rand"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

const skillDefaultMove = "DEFAULT_MOVE"

func getDefaultMoveSkill() swagger.DungeonsandtrollsSkill {
	targetType := swagger.POSITION_SkillTarget
	return swagger.DungeonsandtrollsSkill{
		Name:   skillDefaultMove,
		Target: &targetType,
		CasterEffects: &swagger.DungeonsandtrollsSkillEffect{
			Flags: &swagger.DungeonsandtrollsSkillSpecificFlags{
				Movement: true,
			},
		},
		Range_: &swagger.DungeonsandtrollsAttributes{
			Constant: 1,
		},

		TargetEffects: &swagger.DungeonsandtrollsSkillEffect{
			Flags: &swagger.DungeonsandtrollsSkillSpecificFlags{},
		},
		Cost:         &swagger.DungeonsandtrollsAttributes{},
		Radius:       &swagger.DungeonsandtrollsAttributes{},
		Duration:     &swagger.DungeonsandtrollsAttributes{},
		DamageAmount: &swagger.DungeonsandtrollsAttributes{},
		Flags:        &swagger.DungeonsandtrollsSkillGenericFlags{},
	}
}

func isDefaultMoveSkill(skill swagger.DungeonsandtrollsSkill) bool {
	return skill.Name == skillDefaultMove
}

func (b *Bot) getNeighborPositions() []MapObject {
	x := b.Details.Position.PositionX
	y := b.Details.Position.PositionY
	positions := []swagger.DungeonsandtrollsPosition{
		makePosition(x-1, y),
		makePosition(x+1, y),
		makePosition(x, y-1),
		makePosition(x, y+1),
	}
	targets := []MapObject{}
	for _, pos := range positions {
		if b.BotState.MapExtended[pos].mapObjects.IsFree {
			mapObject := NewEmptyMapObject(pos)
			targets = append(targets, mapObject)
		}
	}
	return targets
}

func (b *Bot) getTargetPositions(maxRange int32) []MapObject {
	b.Logger.Infow("Getting target positions up to max range",
		"maxRange", maxRange,
	)
	positions := []MapObject{}
	for d := 1; d < int(maxRange); d++ {
		for i := 0; i < 6; i++ {
			distanceX := rand.Intn(2*d+2) - d
			distanceY := rand.Intn(2*d+2) - d
			newX := b.Details.Position.PositionX + int32(distanceX)
			newY := b.Details.Position.PositionY + int32(distanceY)

			position := makePosition(newX, newY)
			tileInfo, found := b.BotState.MapExtended[position]
			if !found || !tileInfo.mapObjects.IsFree {
				// unreachable or not free
				continue
			}
			obj := NewEmptyMapObject(position)
			positions = append(positions, obj)
			if len(positions) >= d*2 {
				break
			}
		}
	}
	return positions
}
