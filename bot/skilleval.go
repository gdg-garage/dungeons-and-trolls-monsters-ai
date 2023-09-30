package bot

import (
	"math"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

type SkillResult struct {
	Damage int32
}

func (b *Bot) getSkillTargetPosition(skill *swagger.DungeonsandtrollsSkill, target *MapObject) *swagger.DungeonsandtrollsPosition {
	switch *skill.Target {
	case swagger.NONE_SkillTarget:
		return b.Details.Position
	case swagger.POSITION_SkillTarget:
		return target.MapObjects.Position
	case swagger.CHARACTER_SkillTarget:
		return target.MapObjects.Position
	}
	b.Logger.Errorw("PANIC: Unknown target type for skill",
		"skill", skill,
	)
	return nil
}

func (b *Bot) evaluateSkill(skill swagger.DungeonsandtrollsSkill, target MapObject) *SkillResult {
	if !b.areAttributeRequirementMet(*skill.Cost) {
		b.Logger.Infow("Skill attributes cost requirement not met")
		return nil
	}
	// TODO: check out of combat

	casterPostion := b.Details.Position
	targetPosition := b.getSkillTargetPosition(&skill, &target)
	b.Logger.Infow("Checking distance",
		"casterPostion", casterPostion,
		"targetPosition", targetPosition,
		"manhattanDistance", manhattanDistance(*casterPostion, *targetPosition),
	)
	if manhattanDistance(*casterPostion, *targetPosition) > int32(b.calculateAttributesValue(*skill.Range_)) {
		b.Logger.Infow("Enemy out of skill range")
		return nil
	}

	damage := b.calculateAttributesValue(*skill.DamageAmount)
	switch *skill.Target {
	case swagger.NONE_SkillTarget:
		return &SkillResult{}
	case swagger.POSITION_SkillTarget:
		return &SkillResult{
			Damage: int32(damage),
		}
	case swagger.CHARACTER_SkillTarget:
		return &SkillResult{
			Damage: int32(damage),
		}
	}
	return &SkillResult{}
}

func (b *Bot) findTargetsInRange(position swagger.DungeonsandtrollsPosition, dist int32) []MapObject {
	xStart := position.PositionX - dist
	yStart := position.PositionY - dist
	xEnd := position.PositionX + dist
	yEnd := position.PositionY + dist

	targets := []MapObject{}
	for y := yStart; y < yEnd; y++ {
		for x := xStart; x < xEnd; x++ {
			pos := makePosition(x, y)
			if !b.isInBounds(b.Details.Level, pos) || manhattanDistance(pos, position) > dist {
				continue
			}
			targets = append(targets, extractTargets(b.BotState.MapExtended[pos].mapObjects)...)
		}
	}
	return targets
}

func (b *Bot) findTargetsInRadius(position swagger.DungeonsandtrollsPosition, dist int32) []MapObject {
	xStart := position.PositionX - dist
	yStart := position.PositionY - dist
	xEnd := position.PositionX + dist
	yEnd := position.PositionY + dist

	targets := []MapObject{}
	for y := yStart; y < yEnd; y++ {
		for x := xStart; x < xEnd; x++ {
			pos := makePosition(x, y)
			if !b.isInBounds(b.Details.Level, pos) || euclidDistance(pos, position) > dist {
				continue
			}
			targets = append(targets, extractTargets(b.BotState.MapExtended[pos].mapObjects)...)
		}
	}
	return targets
}

func extractTargets(mapObjects swagger.DungeonsandtrollsMapObjects) []MapObject {
	targets := []MapObject{}
	for i, _ := range mapObjects.Players {
		targets = append(targets, NewPlayerMapObject(mapObjects, i))
	}
	for i, _ := range mapObjects.Monsters {
		targets = append(targets, NewMonsterMapObject(mapObjects, i))
	}
	return targets
}

func manhattanDistance(a swagger.DungeonsandtrollsPosition, b swagger.DungeonsandtrollsPosition) int32 {
	return int32(math.Abs(float64(a.PositionX-b.PositionX)) + math.Abs(float64(a.PositionY-b.PositionY)))
}

func euclidDistance(a swagger.DungeonsandtrollsPosition, b swagger.DungeonsandtrollsPosition) int32 {
	return int32(math.Floor(math.Sqrt(math.Pow(float64(a.PositionX-b.PositionX), 2) + math.Pow(float64(a.PositionY-b.PositionY), 2))))
}
