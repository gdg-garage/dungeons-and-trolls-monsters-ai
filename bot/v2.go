package bot

import swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"

func (b *Bot) bestSkill() *swagger.DungeonsandtrollsCommandsBatch {
	allSkills := getAllSkills(b.Details.Monster.EquippedItems)
	b.Logger.Infow("All skills",
		"skills", allSkills,
		"numSkills", len(allSkills),
	)
	reqSkills := b.filterRequirementsMetSkills(allSkills)
	// TODO: big drop -> rest ??? (maybe not really)

	oocSkills := b.filterCastableWithOOCSkills(reqSkills)
	b.Logger.Infow("Castable skills",
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
					result := b.evaluateSkill(skill, b.BotState.Self)
					b.Logger.Infow("Skill (target none) evaluated",
						"skillName", skill.Name,
						"result", result,
						"result.VitalsSelf", result.VitalsSelf,
						"resultsCombinedScore", b.getCombinedVitalsScore(result),
					)
					if b.isBetterThanSkillResult(result, bestResult) {
						b.Logger.Infow("New best skill (target none).",
							"skillName", skill.Name,
							"result", result,
							"result.VitalsSelf", result.VitalsSelf,
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
					result := b.evaluateSkill(skill, target)
					b.Logger.Infow("Skill + target evaluated",
						"skillName", skill.Name,
						"targetName", target.GetName(),
						"result", result,
						"result.VitalsSelf", result.VitalsSelf,
						"result.VitalsFriendly", result.VitalsFriendly,
						"result.VitalsHostile", result.VitalsHostile,
						"resultsCombinedScore", b.getCombinedVitalsScore(result),
					)
					if b.isBetterThanSkillResult(result, bestResult) {
						b.Logger.Infow("New best skill + target combination.",
							"skillName", skill.Name,
							"result", result,
							"result.VitalsSelf", result.VitalsSelf,
							"result.VitalsFriendly", result.VitalsFriendly,
							"result.VitalsHostile", result.VitalsHostile,
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
		return b.randomWalk()
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
	return 2*s.VitalsSelf + s.VitalsFriendly - 4*s.VitalsHostile
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
