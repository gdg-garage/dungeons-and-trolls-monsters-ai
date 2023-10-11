package bot

import swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"

func (b *Bot) bestSkill() *swagger.DungeonsandtrollsCommandsBatch {
	allSkills := b.filterActiveSkills(getAllSkills(b.Details.Monster.EquippedItems))
	b.Logger.Infow("All skills",
		"skills", allSkills,
		"numSkills", len(allSkills),
	)
	reqSkills := b.filterRequirementsMetSkills(allSkills)
	// TODO: big drop -> rest ??? (maybe not really)

	oocSkills := b.filterCastableWithOOCSkills(reqSkills)
	// Add default move skill
	oocSkills = append(oocSkills, getDefaultMoveSkill())
	skillNames := []string{}
	for s := range oocSkills {
		skill := oocSkills[s]
		skillNames = append(skillNames, skill.Name)
	}
	b.Logger.Infow("Castable skills",
		"skillNames", skillNames,
		"skills", oocSkills,
		"numSkills", len(oocSkills),
	)
	// TODO: big drop -> move to safety
	skillsByRange := map[int][]swagger.DungeonsandtrollsSkill{}

	maxRange := 0
	for i := range oocSkills {
		skill := oocSkills[i]
		range_ := b.calculateAttributesValue(*skill.Range_)
		skillsByRange[range_] = append(skillsByRange[range_], skill)
		if range_ > maxRange {
			maxRange = range_
		}
	}

	targetsByRange := b.findTargetsInRangeAsMap(*b.Details.Position, int32(maxRange))
	b.Logger.Infow("Max range",
		"maxRange", maxRange,
	)
	emptyTargets := b.getEmptyPositionsAsTargets(int32(maxRange))
	for t := range emptyTargets {
		target := emptyTargets[t]
		dist := b.BotState.MapExtended[*target.MapObjects.Position].distance
		targetsByRange[dist] = append(targetsByRange[dist], target)
		b.Logger.Infow("Adding empty target",
			"position", target.MapObjects.Position,
			"myPosition", b.Details.Position,
			"distance", dist,
		)
	}

	allTargets := []MapObject{}
	targetNames := []string{}
	for _, targets := range targetsByRange {
		allTargets = append(allTargets, targets...)
	}
	for _, target := range allTargets {
		targetNames = append(targetNames, target.GetName())
	}
	b.Logger.Infow("All targets in range",
		"rangeBucketsCount", len(targetsByRange),
		"allTargetsCount", len(allTargets),
		"targetNames", targetNames,
		"maxRange", maxRange,
	)

	bestResult := SkillResult{}
	var bestSkill *swagger.DungeonsandtrollsSkill
	var bestTarget *MapObject

	for skillRange, skills := range skillsByRange {
		for targetRange, targets := range targetsByRange {
			if targetRange > skillRange {
				continue
			}
			for s := range skills {
				skill := skills[s]
				if *skill.Target == swagger.NONE_SkillTarget {
					target := b.BotState.Self
					if !b.isLegalSkillTargetCombination(skill, target) {
						continue
					}
					result := b.evaluateSkill(skill, target)
					b.Logger.Infow("Skill (target none) evaluated",
						"skillName", skill.Name,
						"result", result,
						"result.VitalsSelf", result.VitalsSelf,
						"result.Random", result.Random,
						"resultsCombinedScore", b.getCombinedVitalsScore(result),
					)
					if b.isBetterThanSkillResult(result, bestResult) {
						b.Logger.Infow("New best skill (target none).",
							"skillName", skill.Name,
							"result", result,
							"result.VitalsSelf", result.VitalsSelf,
							"result.Random", result.Random,
							"resultsCombinedScore", b.getCombinedVitalsScore(result),
						)
						bestResult = result
						bestSkill = &skill
						bestTarget = &b.BotState.Self
					}
					break
				}
				for t := range targets {
					target := targets[t]
					if !b.isLegalSkillTargetCombination(skill, target) {
						continue
					}
					result := b.evaluateSkill(skill, target)
					b.Logger.Infow("Skill + target evaluated",
						"skillName", skill.Name,
						"targetName", target.GetName(),
						"targetPosition", target.MapObjects.Position,
						"myPosition", b.Details.Position,
						"result", result,
						"result.VitalsSelf", result.VitalsSelf,
						"result.VitalsFriendly", result.VitalsFriendly,
						"result.VitalsHostile", result.VitalsHostile,
						"result.MovementSelf", result.MovementSelf,
						"result.Random", result.Random,
						"resultsCombinedScore", b.getCombinedVitalsScore(result),
					)
					if b.isBetterThanSkillResult(result, bestResult) {
						b.Logger.Infow("New best skill + target combination.",
							"skillName", skill.Name,
							"targetName", target.GetName(),
							"targetPosition", target.MapObjects.Position,
							"myPosition", b.Details.Position,
							"result", result,
							"result.VitalsSelf", result.VitalsSelf,
							"result.VitalsFriendly", result.VitalsFriendly,
							"result.VitalsHostile", result.VitalsHostile,
							"result.MovementSelf", result.MovementSelf,
							"result.Random", result.Random,
							"resultCombinedScore", b.getCombinedVitalsScore(result),
						)
						bestResult = result
						bestSkill = &skill
						bestTarget = &target
					}
				}
			}
		}
	}

	if bestSkill == nil {
		b.Logger.Warnw("No skill chosen")
		move := b.moveTowardsEnemy(b.BotState.Objects.Hostile)
		if move != nil {
			return move
		}
		return nil //b.randomWalk()
	}
	b.Logger.Infow("Best skill + target combination!!!",
		"skillName", bestSkill.Name,
		"skill", bestSkill,
		"result", bestResult,
		"resultCombinedScore", b.getCombinedVitalsScore(bestResult),
		"targetId", bestTarget.GetId(),
		"targetName", bestTarget.GetName(),
		"targetFaction", bestTarget.GetFaction(),
		"position", bestTarget.MapObjects.Position,
		"myPosition", b.Details.Position,
	)
	return b.useSkill(*bestSkill, *bestTarget)
}

func (b *Bot) getCombinedVitalsScore(s SkillResult) float32 {
	// XXX: Coefficients here can be tweaked for aggression vs. survival preference
	return b.Config.Preservation*s.VitalsSelf +
		b.Config.Support*s.VitalsFriendly +
		-b.Config.Aggression*s.VitalsHostile +
		s.MovementSelf +
		b.Config.Randomness*s.Random
}

func (b *Bot) isBetterThanSkillResult(sk1, sk2 SkillResult) bool {
	return b.getCombinedVitalsScore(sk1) > b.getCombinedVitalsScore(sk2)
}

func (b *Bot) getBetterSkillResult(s1, s2 *SkillResult) SkillResult {
	if b.getCombinedVitalsScore(*s1) > b.getCombinedVitalsScore(*s2) {
		return *s1
	}
	return *s2
}
