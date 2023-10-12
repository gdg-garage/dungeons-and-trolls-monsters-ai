package bot

import (
	"math"
	"math/rand"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

const skillDefaultMove = "DEFAULT_MOVE"

func getDefaultMoveSkill() swagger.DungeonsandtrollsSkill {
	targetType := swagger.POSITION_SkillTarget
	damageType := swagger.NONE_DungeonsandtrollsDamageType
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
		DamageType:   &damageType,
		Flags:        &swagger.DungeonsandtrollsSkillGenericFlags{},
	}
}

func isDefaultMoveSkill(skill swagger.DungeonsandtrollsSkill) bool {
	return skill.Name == skillDefaultMove
}

func (b *Bot) getEmptyPositionsAsTargets(maxRange int32) []MapObject {
	return b.getEmptyPositionsAsTargetsFromPosition(*b.Details.Position, maxRange)
}

func (b *Bot) getEmptyPositionsAsTargetsFromPosition(position swagger.DungeonsandtrollsPosition, dist int32) []MapObject {
	xStart := position.PositionX - dist
	yStart := position.PositionY - dist
	xEnd := position.PositionX + dist
	yEnd := position.PositionY + dist

	targets := []MapObject{}
	for y := yStart; y <= yEnd; y++ {
		for x := xStart; x <= xEnd; x++ {
			pos := makePosition(x, y)
			if !b.isInBounds(b.Details.Level, pos) || manhattanDistance(pos, position) > dist {
				continue
			}
			if pos == *b.Details.Position {
				// Skip current position
				continue
			}
			tileInfo, found := b.BotState.MapExtended[pos]
			if found && (!tileInfo.mapObjects.IsFree || len(tileInfo.mapObjects.Monsters) > 0 || len(tileInfo.mapObjects.Players) > 0) {
				// Skip non-free tiles or tiles with monsters or players
				continue
			}
			targets = append(targets, NewEmptyMapObject(pos))
		}
	}
	return targets
}

func (b *Bot) stretchMovePosition(position swagger.DungeonsandtrollsPosition) swagger.DungeonsandtrollsPosition {
	x := position.PositionX - b.Details.Position.PositionX
	y := position.PositionY - b.Details.Position.PositionY
	x = b.Details.Position.PositionX + x*8 + rand.Int31n(4) - rand.Int31n(4)
	y = b.Details.Position.PositionY + y*8 + rand.Int31n(4) - rand.Int31n(4)
	candidatePosition := makePosition(x, y)

	dist := int32(8)
	xStart := x - dist
	yStart := y - dist
	xEnd := x + dist
	yEnd := y + dist

	bestPosition := swagger.DungeonsandtrollsPosition{}
	var bestDistance float32
	bestDistance = math.MaxFloat32
	for y := yStart; y < yEnd; y++ {
		for x := xStart; x < xEnd; x++ {
			pos := makePosition(x, y)
			distanceToCandidate := manhattanDistance(pos, candidatePosition)
			if !b.isInBounds(b.Details.Level, pos) || distanceToCandidate > dist {
				continue
			}
			tileInfo, found := b.BotState.MapExtended[pos]
			if found && tileInfo.mapObjects.IsFree {
				distance := float32(int(4*distanceToCandidate)+tileInfo.distance) + rand.Float32()/100
				if distance < bestDistance {
					bestDistance = distance
					bestPosition = pos
				}
			}
		}
	}
	return bestPosition
}
