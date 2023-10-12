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

	BuffsHostile  float32
	BuffsFriendly float32
	BuffsSelf     float32

	ResistsHostile  float32
	ResistsFriendly float32
	ResistsSelf     float32

	// MovementHostile float32
	MovementSelf float32
	// MovementFriendly float32

	Random float32
}

func (sr *SkillResult) Add(other SkillResult) *SkillResult {
	sr.VitalsHostile += other.VitalsHostile
	sr.VitalsFriendly += other.VitalsFriendly
	sr.VitalsSelf += other.VitalsSelf
	sr.BuffsHostile += other.BuffsHostile
	sr.BuffsFriendly += other.BuffsFriendly
	sr.BuffsSelf += other.BuffsSelf
	sr.ResistsHostile += other.ResistsHostile
	sr.ResistsFriendly += other.ResistsFriendly
	sr.ResistsSelf += other.ResistsSelf
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

func (b *Bot) isLegalSkillTargetCombination(skill swagger.DungeonsandtrollsSkill, target MapObject) bool {
	if target.IsEmpty() && *skill.Target != swagger.POSITION_SkillTarget {
		// Target CHARACTER can't be used with empty targets
		// Target NONE is possible but useless - we use "self" to evaluate "target NONE" skills
		b.Logger.Infow("Skipping skill evaluation for empty target")
		return false
	}
	return true
}

func (b *Bot) evaluateSkill(skill swagger.DungeonsandtrollsSkill, target MapObject) SkillResult {
	empty := SkillResult{}
	if !b.areAttributeRequirementMet(*skill.Cost) {
		b.Logger.Infow("Skill attributes cost requirement not met")
		return empty
	}

	casterPosition := b.Details.Position
	targetPosition := b.getSkillTargetPosition(&skill, &target)
	b.Logger.Debug("Checking distance",
		"casterPosition", casterPosition,
		"targetPosition", targetPosition,
		"manhattanDistance", manhattanDistance(*casterPosition, *targetPosition),
	)
	if manhattanDistance(*casterPosition, *targetPosition) > int32(b.calculateAttributesValue(*skill.Range_)) {
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
	b.Logger.Infow("Eval for caster")
	result := b.evalEffectFor(&b.BotState.Self, skill.CasterEffects, &skill, false)
	result.Random = rand.Float32()
	// Eval movement for self
	if skill.CasterEffects.Flags.Movement {
		result.MovementSelf = float32(b.scoreMovementDiff(targetPosition)) / 7
	}
	// Eval ground effect around caster
	if skill.CasterEffects.Flags.GroundEffect {
		targets := b.findTargetsInRadius(*casterPosition, radius)
		for i := range targets {
			target_ := targets[i]
			result.Add(b.evalEffectFor(&target_, skill.TargetEffects, &skill, true))
		}
		return result
	}
	// Radius 0 special case
	if radius <= 0 {
		// Eval target if character
		if *skill.Target == swagger.CHARACTER_SkillTarget {
			b.Logger.Infow("Eval for character target",
				"target", target.GetName(),
				"resultBefore", result,
			)
			result.Add(b.evalEffectFor(&target, skill.TargetEffects, &skill, true))
			b.Logger.Infow("AFTER Eval for character target",
				"target", target.GetName(),
				"resultAfter", result,
			)
		} else if *skill.Target == swagger.POSITION_SkillTarget {
			targets := b.findTargetsInRadius(*targetPosition, radius)
			b.Logger.Infow("Eval for position target",
				"numTargets", len(targets),
			)
			for i := range targets {
				target_ := targets[i]
				if target_.GetId() == b.BotState.Self.GetId() {
					b.Logger.Infow("No eval for self")
					continue
				}
				result.Add(b.evalEffectFor(&target_, skill.TargetEffects, &skill, true))
			}
		} else {
			b.Logger.Infow("No Eval for none target")
		}
		return result
	}
	// Eval AoE / ground effect around target
	targets := b.findTargetsInRadius(*targetPosition, radius)
	for i := range targets {
		b.Logger.Infow("Eval for AoE / ground effect",
			"numTargets", len(targets),
		)
		target_ := targets[i]
		if *skill.Target == swagger.NONE_SkillTarget && target_.GetId() == b.BotState.Self.GetId() {
			continue
		}
		result.Add(b.evalEffectFor(&target_, skill.TargetEffects, &skill, true))
	}
	return result
}

func (b *Bot) evalEffectFor(target *MapObject, effect *swagger.DungeonsandtrollsSkillEffect, skill *swagger.DungeonsandtrollsSkill, withDamage bool) SkillResult {
	if b.IsNeutral(*target) {
		return SkillResult{}
	}
	if effect == nil {
		effect = &swagger.DungeonsandtrollsSkillEffect{}
	}
	if effect.Attributes == nil {
		effect.Attributes = &swagger.DungeonsandtrollsSkillAttributes{}
	}
	var vitalsScore, buffsScore, resistsScore float32
	if withDamage {
		vitalsScore, buffsScore, resistsScore = b.scoreVitalsWithDamage(target, effect.Attributes, skill)
	} else {
		// withCost
		vitalsScore, buffsScore, resistsScore = b.scoreVitalsWithCost(effect.Attributes, skill)
	}
	if effect.Flags.Stun && !b.GetStunInfo(*target).IsImmune {
		if target.GetId() == b.Details.Id {
			vitalsScore -= 0.4
		} else if b.IsHostile(*target) {
			vitalsScore -= 0.3
		} else {
			vitalsScore -= 0.2
		}
	}
	if effect.Flags.Knockback {
		vitalsScore -= 0.1
	}
	vitalsSummons := float32(0)
	if effect.Summons != nil && len(effect.Summons) > 0 {
		vitalsSummons += 0.35
	}
	// XXX: Maybe make bigger targets worth more
	//      Not relevant because players are on the same level
	// vitalsScore *= target.GetMaxAttributes().Life
	if target.GetId() == b.BotState.Self.GetId() {
		return SkillResult{
			VitalsSelf:  vitalsScore,
			BuffsSelf:   buffsScore,
			ResistsSelf: resistsScore,

			VitalsFriendly: vitalsSummons,
		}
	}
	if b.IsHostile(*target) {
		return SkillResult{
			VitalsHostile:  vitalsScore,
			BuffsHostile:   buffsScore,
			ResistsHostile: resistsScore,

			VitalsFriendly: vitalsSummons,
		}
	}
	// Neutral effects are included in friendly for simplicity
	return SkillResult{
		VitalsFriendly:  vitalsScore + vitalsSummons,
		BuffsFriendly:   buffsScore,
		ResistsFriendly: resistsScore,
	}
}

func (b *Bot) findTargetsInRange(position swagger.DungeonsandtrollsPosition, dist int32) []MapObject {
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
		if mapObjects.Monsters[i].Faction == "neutral" {
			// Do not target neutral monsters (chests, etc.)
			continue
		}
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

func (b *Bot) scoreMovementDiff(position *swagger.DungeonsandtrollsPosition) float32 {
	return b.scoreMovement(position) - b.scoreMovement(b.Details.Position)
}

func (b *Bot) scoreMovement(position *swagger.DungeonsandtrollsPosition) float32 {
	distances := b.calculateDistancesForPosition(position)
	distances.NumCloseFriendly += 1
	distances.NumCloseHostiles += 1
	if distances.NumCloseHostiles > 10 {
		distances.NumCloseHostiles = 10
	}
	if distances.NumCloseFriendly > 10 {
		distances.NumCloseFriendly = 10
	}
	scoreClosestHostile := 10 / float32(distances.DistanceToClosestHostile+10)
	if distances.DistanceToClosestHostile < 2 {
		scoreClosestHostile -= 0.08
		if distances.DistanceToClosestHostile == 0 {
			scoreClosestHostile -= 0.06
		}
	}
	scoreClosestFriendly := 10 / float32(distances.DistanceToClosestFriendly+10)
	if distances.DistanceToClosestFriendly < 2 {
		scoreClosestFriendly -= 0.21
		if distances.DistanceToClosestFriendly == 0 {
			scoreClosestFriendly -= 0.31
		}
	}
	scoreTargetPosition := 20 / float32(distances.DistanceToTargetPosition+20)

	scoreDistToSelf := float32(distances.DistanceToSelf) / 10
	scoreNumHostiles := float32(distances.NumCloseHostiles) / 10
	scoreNumFriendly := float32(distances.NumCloseFriendly) / 20

	scorePosition := float32(0)
	tileInfo, found := b.BotState.MapExtended[*position]
	if found {
		if tileInfo.mapObjects.IsStairs || tileInfo.mapObjects.IsSpawn {
			scorePosition -= 0.7
		}
		for _, monster := range tileInfo.mapObjects.Monsters {
			if monster.Id != b.Details.Monster.Id {
				scorePosition -= 0.12
			}
		}
	}

	vitalsSelf := b.getCurrentVitals()
	vitalsCoef := (vitalsSelf - 4) / 5 // assuming 0-10
	// TODO: use distances and vitals
	result := b.Config.Restlessness*scoreDistToSelf +
		scoreClosestHostile*6 +
		scoreClosestFriendly*2 +
		scoreTargetPosition*20 +
		vitalsCoef*scoreNumHostiles*2 +
		scoreNumFriendly*1 +
		scorePosition

	b.Logger.Infow("Evaluated movement score for self",
		"result.MovementSelf", result,
		"distances", distances,
		"myPosition", b.Details.Position,
		"position", position,
		"scoreClosestHostile", scoreClosestHostile,
		"scoreClosestFriendly", scoreClosestFriendly,
		"scoreDistToSelf", scoreDistToSelf,
		"scoreTargetPosition", scoreTargetPosition,
		"scoreNumHostiles", scoreNumHostiles,
		"scoreNumFriendly", scoreNumFriendly,
		"vitalsSelf", vitalsSelf,
		"vitalsCoef", vitalsCoef,
	)
	return result
}

func (b *Bot) getCurrentVitals() float32 {
	monster := b.Details.Monster
	staminaPercentage := monster.Attributes.Stamina / monster.MaxAttributes.Stamina
	manaPercentage := monster.Attributes.Mana / monster.MaxAttributes.Mana

	return b.scoreVitalsFunc(monster.LifePercentage, staminaPercentage, manaPercentage)
}

type Distances struct {
	DistanceToClosestHostile  int32
	DistanceToClosestFriendly int32
	DistanceToSelf            int32
	DistanceToTargetPosition  int32

	NumCloseHostiles int
	NumCloseFriendly int
}

const CLOSE_DISTANCE = 12

func (b *Bot) calculateDistancesForPosition(position *swagger.DungeonsandtrollsPosition) Distances {
	dists := Distances{
		DistanceToSelf:            manhattanDistance(*b.Details.Position, *position),
		DistanceToClosestHostile:  math.MaxInt32 - 1,
		DistanceToClosestFriendly: math.MaxInt32 - 1,
		DistanceToTargetPosition:  math.MaxInt32 - 1,
		NumCloseFriendly:          1, // self
	}
	for _, obj := range b.Details.CurrentMap.Objects {
		if !b.BotState.MapExtended[*obj.Position].lineOfSight {
			// Skip position without line of sight
			continue
		}
		dist := manhattanDistance(*position, *obj.Position)
		if len(obj.Players) > 0 {
			mo := NewPlayerMapObject(obj, 0)
			if b.IsHostile(mo) {
				if dist < dists.DistanceToClosestHostile {
					dists.DistanceToClosestHostile = dist
				}
				dists.NumCloseHostiles += len(obj.Players)
			} else if b.IsFriendly(mo) {
				if dist < dists.DistanceToClosestFriendly {
					dists.DistanceToClosestFriendly = dist
				}
				dists.NumCloseFriendly += len(obj.Players)
			}
		}
		if dist > CLOSE_DISTANCE {
			continue
		}

		for i, monster := range obj.Monsters {
			if monster.Faction == "neutral" {
				continue
			}
			mo := NewMonsterMapObject(obj, i)
			if b.IsHostile(mo) {
				if dist < dists.DistanceToClosestHostile {
					dists.DistanceToClosestHostile = dist
				}
				dists.NumCloseHostiles++
			} else if b.IsFriendly(mo) {
				if dist < dists.DistanceToClosestFriendly {
					dists.DistanceToClosestFriendly = dist
				}
				dists.NumCloseFriendly++
			}
		}
	}
	return dists
}
