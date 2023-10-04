package bot

import (
	"math"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

type SkillResult struct {
	VitalsHostile  float32
	VitalsFriendly float32
	VitalsSelf     float32
}

func (sr *SkillResult) Add(other SkillResult) *SkillResult {
	sr.VitalsHostile += other.VitalsHostile
	sr.VitalsFriendly += other.VitalsFriendly
	sr.VitalsSelf += other.VitalsSelf
	return sr
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
		b.Logger.Debugw("Skill attributes cost requirement not met")
		return nil
	}
	// TODO: check out of combat

	casterPostion := b.Details.Position
	targetPosition := b.getSkillTargetPosition(&skill, &target)
	b.Logger.Debug("Checking distance",
		"casterPostion", casterPostion,
		"targetPosition", targetPosition,
		"manhattanDistance", manhattanDistance(*casterPostion, *targetPosition),
	)
	if manhattanDistance(*casterPostion, *targetPosition) > int32(b.calculateAttributesValue(*skill.Range_)) {
		b.Logger.Debugw("Enemy out of skill range")
		return nil
	}
	if !b.BotState.MapExtended[*targetPosition].lineOfSight {
		b.Logger.Debugw("Enemy is not in line of sight")
		return nil
	}
	// Check out of combat

	// Eval yourself
	result := b.evalEffectFor(&b.BotState.Self, skill.CasterEffects, &skill)
	// Eval target(s)
	switch *skill.Target {
	case swagger.NONE_SkillTarget:
		break
	case swagger.POSITION_SkillTarget:
		targets := b.findTargetsInRadius(*targetPosition, int32(b.calculateAttributesValue(*skill.Radius)))
		for i, _ := range targets {
			target_ := targets[i]
			result.Add(b.evalEffectFor(&target_, skill.TargetEffects, &skill))
		}
		break
	case swagger.CHARACTER_SkillTarget:
		result.Add(b.evalEffectFor(&target, skill.TargetEffects, &skill))
		break
	}
	return &result
}

func (b *Bot) evalEffectFor(target *MapObject, effect *swagger.DungeonsandtrollsSkillEffect, skill *swagger.DungeonsandtrollsSkill) SkillResult {
	if effect == nil || effect.Attributes == nil || b.IsNeutral(*target) {
		return SkillResult{}
	}
	var vitalsScore float32
	if target.GetId() == b.BotState.Self.GetId() {
		vitalsScore = b.scoreVitalsWithCost(effect.Attributes, skill)
	} else {
		vitalsScore = b.scoreVitalsWithDamage(target, effect.Attributes, skill)
	}
	vitalsScore *= float32(b.calculateAttributesValue(*skill.Duration))
	if effect.Flags.Stun {
		if b.IsHostile(*target) {
			// We assume we will have time to deal twice the damage
			// because of stun next turn
			// Plus 0.2 ~ 10% of HP worth of damage
			vitalsScore -= 0.2
		} else {
			// Minus 0.4 ~ 20% of HP worth of damage
			vitalsScore -= 0.4
		}
	}
	if effect.Flags.Knockback {
		vitalsScore -= 0.1
	}
	// XXX: Maybe make bigger targets worth more
	//      Not relevant because players are on the same level
	// vitalsScore *= target.GetMaxAttributes().Life
	if b.IsHostile(*target) {
		return SkillResult{
			VitalsHostile: vitalsScore,
		}
	} else {
		return SkillResult{
			VitalsFriendly: vitalsScore,
		}
	}
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
