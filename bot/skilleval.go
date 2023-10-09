package bot

import (
	"math"
	"math/rand"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

type SkillResult struct {
	VitalsHostile  float32
	VitalsFriendly float32
	VitalsSelf     float32

	// MovementHostile float32
	MovementSelf float32
	// MovementFriendly float32

	Random float32
}

func (sr *SkillResult) Add(other SkillResult) *SkillResult {
	sr.VitalsHostile += other.VitalsHostile
	sr.VitalsFriendly += other.VitalsFriendly
	sr.VitalsSelf += other.VitalsSelf
	sr.MovementSelf += other.MovementSelf
	sr.Random += other.Random
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

func (b *Bot) evaluateSkill(skill swagger.DungeonsandtrollsSkill, target MapObject) SkillResult {
	empty := SkillResult{}
	if target.IsEmpty() && *skill.Target != swagger.POSITION_SkillTarget {
		// Target CHARACTER can't be used with empty targets
		// Target NONE is possible but useless - we use "self" to evaluate "target NONE" skills
		b.Logger.Infow("Skipping skill evaluation for empty target")
		return empty
	}
	if !b.areAttributeRequirementMet(*skill.Cost) {
		b.Logger.Infow("Skill attributes cost requirement not met")
		return empty
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
		b.Logger.Infow("Target out of skill range")
		return empty
	}
	if skill.Flags.RequiresLineOfSight && !b.BotState.MapExtended[*targetPosition].lineOfSight {
		b.Logger.Infow("Target is not in line of sight")
		return empty
	}
	// TODO: Check out of combat

	radius := int32(b.calculateAttributesValue(*skill.Radius))

	// Eval yourself
	result := b.evalEffectFor(&b.BotState.Self, skill.CasterEffects, &skill)
	result.Random = rand.Float32()
	// Eval movement for self
	if skill.CasterEffects.Flags.Movement {
		result.MovementSelf = b.scoreMovement(targetPosition)
	}
	// Eval ground effect around caster
	if skill.CasterEffects.Flags.GroundEffect {
		targets := b.findTargetsInRadius(*casterPostion, radius)
		for i := range targets {
			target_ := targets[i]
			result.Add(b.evalEffectFor(&target_, skill.TargetEffects, &skill))
		}
		return result
	}
	// Radius 0 special case
	if radius <= 0 {
		// Eval target if character
		if *skill.Target == swagger.CHARACTER_SkillTarget {
			result.Add(b.evalEffectFor(&target, skill.TargetEffects, &skill))
		} else if *skill.Target == swagger.POSITION_SkillTarget {
			targets := b.findTargetsInRadius(*targetPosition, radius)
			for i := range targets {
				target_ := targets[i]
				if target_.GetId() == b.BotState.Self.GetId() || target_.GetId() == target.GetId() {
					// Exclude caster and target
					continue
				}
				result.Add(b.evalEffectFor(&target_, skill.TargetEffects, &skill))
			}
		}
		return result
	}
	// Eval AoE / ground effect around target
	targets := b.findTargetsInRadius(*targetPosition, radius)
	for i := range targets {
		target_ := targets[i]
		if target_.GetId() == b.BotState.Self.GetId() {
			// Exclude caster and target
			continue
		}
		result.Add(b.evalEffectFor(&target_, skill.TargetEffects, &skill))
	}
	return result
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
	duration := float32(b.calculateAttributesValue(*skill.Duration))
	if duration == 0 {
		duration = 1
	}
	vitalsScore *= duration
	if effect.Flags.Stun {
		if b.IsHostile(*target) {
			if !b.GetStunInfo(*target).IsImmune {
				vitalsScore -= 0.2
			}
		} else {
			if !b.Details.Monster.Stun.IsImmune {
				vitalsScore -= 0.4
			}
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

func (b *Bot) findTargetsInRangeAsMap(position swagger.DungeonsandtrollsPosition, dist int32) map[int][]MapObject {
	xStart := position.PositionX - dist
	yStart := position.PositionY - dist
	xEnd := position.PositionX + dist
	yEnd := position.PositionY + dist

	targets := map[int][]MapObject{}
	targetNames := []string{}
	for y := yStart; y <= yEnd; y++ {
		for x := xStart; x <= xEnd; x++ {
			pos := makePosition(x, y)
			distance := manhattanDistance(pos, position)
			tileInfo, found := b.BotState.MapExtended[pos]
			if !found || distance > dist {
				continue
			}
			extractedTargets := extractTargets(tileInfo.mapObjects)
			for t := range extractedTargets {
				target_ := extractedTargets[t]
				targets[int(distance)] = append(targets[int(distance)], target_)
				targetNames = append(targetNames, target_.GetName())
			}
			if len(targets[int(distance)]) > 0 {
				b.Logger.Infow("Targets added?",
					"position", pos,
					"targetNames", targetNames,
					"targetCount", len(targets[int(distance)]),
					"range", distance,
				)
			}
		}
	}
	return targets
}

func (b *Bot) findTargetsInRadius(position swagger.DungeonsandtrollsPosition, dist int32) []MapObject {
	if dist < 0 {
		dist = 0
	}
	xStart := position.PositionX - dist
	yStart := position.PositionY - dist
	xEnd := position.PositionX + dist
	yEnd := position.PositionY + dist

	targets := []MapObject{}
	for y := yStart; y <= yEnd; y++ {
		for x := xStart; x <= xEnd; x++ {
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
	for i := range mapObjects.Players {
		targets = append(targets, NewPlayerMapObject(mapObjects, i))
	}
	for i := range mapObjects.Monsters {
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

func (b *Bot) scoreMovement(position *swagger.DungeonsandtrollsPosition) float32 {
	dist := b.BotState.MapExtended[*position].distance
	result := float32(dist) / 10
	b.Logger.Infow("Evaluated movement score for self",
		"result.MovementSelf", result,
		"distance", dist,
	)
	return result
}
