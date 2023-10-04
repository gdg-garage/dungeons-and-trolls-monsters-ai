package bot

import (
	"math"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

func (b *Bot) evaluateHealSkill(skill *swagger.DungeonsandtrollsSkill, target *MapObject) float32 {
	if *skill.Target == swagger.NONE_SkillTarget {
		return 0
	}
	range_ := b.calculateAttributesValue(*skill.Range_)
	if b.BotState.MapExtended[*target.MapObjects.Position].distance > range_ {
		return 0
	}
	return b.scoreVitalsWithDamage(target, skill.TargetEffects.Attributes, skill)
}

func (b *Bot) scoreVitalsWithDamage(target *MapObject, skillAttributes *swagger.DungeonsandtrollsSkillAttributes, skill *swagger.DungeonsandtrollsSkill) float32 {
	damage := b.calculateAttributesValue(*skill.DamageAmount)
	damageAttrs := &swagger.DungeonsandtrollsAttributes{
		Life: float32(damage),
	}
	return b.scoreVitalsFor(target, skillAttributes, damageAttrs, -1, skill)
}

func (b *Bot) scoreVitalsWithCost(skillAttributes *swagger.DungeonsandtrollsSkillAttributes, skill *swagger.DungeonsandtrollsSkill) float32 {
	// Adjust cost by duration
	// It will be mulitplied by duration again
	duration := float32(b.calculateAttributesValue(*skill.Duration))
	costAttrs := &swagger.DungeonsandtrollsAttributes{
		Life:    skill.Cost.Life / duration,
		Stamina: skill.Cost.Stamina / duration,
		Mana:    skill.Cost.Mana / duration,
	}
	return b.scoreVitalsFor(&b.BotState.Self, skillAttributes, costAttrs, -1, skill)
}

// Get vitals score for skill
// Tells you how much the skill will improve your resources (life, stamina, mana)
// Can be used for both casterEffect and targetEffect skills
func (b *Bot) scoreVitalsFor(target *MapObject, skillAttributes *swagger.DungeonsandtrollsSkillAttributes, extraAttributes *swagger.DungeonsandtrollsAttributes, extraSign float32, skill *swagger.DungeonsandtrollsSkill) float32 {
	targetAttrs := target.GetAttributes()
	targetMaxAttrs := target.GetMaxAttributes()
	skillAttributes = fillSkillAttributes(*skillAttributes)

	skillLifeGain := float32(calculateAttributesValue(*targetAttrs, *skillAttributes.Life)) + extraSign*extraAttributes.Life
	skillStaminaGain := float32(calculateAttributesValue(*targetAttrs, *skillAttributes.Stamina)) + extraSign*extraAttributes.Stamina
	skillManaGain := float32(calculateAttributesValue(*targetAttrs, *skillAttributes.Mana)) + extraSign*extraAttributes.Mana

	// Apply duration
	// XXX: This might make over time skills too significant
	duration := float32(b.calculateAttributesValue(*skill.Duration))
	skillLifeGain *= duration
	skillStaminaGain *= duration
	skillManaGain *= duration

	lifePercentage := targetAttrs.Life / targetMaxAttrs.Life
	staminaPercentage := targetAttrs.Stamina / targetMaxAttrs.Stamina
	if math.IsNaN(float64(staminaPercentage)) {
		staminaPercentage = 0
	}
	manaPercentage := targetAttrs.Mana / targetMaxAttrs.Mana
	if math.IsNaN(float64(manaPercentage)) {
		manaPercentage = 0
	}
	score := b.scoreVitalsFunc(lifePercentage, staminaPercentage, manaPercentage)

	lifePercentageAfter := (targetAttrs.Life + skillLifeGain) / targetMaxAttrs.Life
	staminaPercentageAfter := (targetAttrs.Stamina + skillStaminaGain) / targetMaxAttrs.Stamina
	if math.IsNaN(float64(staminaPercentageAfter)) {
		staminaPercentageAfter = 0
	}
	manaPercentageAfter := (targetAttrs.Mana + skillManaGain) / targetMaxAttrs.Mana
	if math.IsNaN(float64(manaPercentageAfter)) {
		manaPercentageAfter = 0
	}
	scoreAfter := b.scoreVitalsFunc(lifePercentageAfter, staminaPercentageAfter, manaPercentageAfter)

	scoreDiff := scoreAfter - score

	b.Logger.Debugw("Skill vitals score",
		"skillName", skill.Name,
		"skill", skill,
		"skillAttributes", skillAttributes,
		"lifeGain", skillLifeGain,
		"staminaGain", skillStaminaGain,
		"manaGain", skillManaGain,
		"lifePercentage", lifePercentage,
		"life", targetAttrs.Life,
		"lifeMax", targetMaxAttrs.Life,
		"staminaPercentage", staminaPercentage,
		"stamina", targetAttrs.Stamina,
		"staminaMax", targetMaxAttrs.Stamina,
		"manaPercentage", manaPercentage,
		"mana", targetAttrs.Mana,
		"manaMax", targetMaxAttrs.Mana,
		"lifePercentageAfter", lifePercentageAfter,
		"staminaPercentageAfter", staminaPercentageAfter,
		"manaPercentageAfter", manaPercentageAfter,
		"vitalsScore", score,
		"vitalsScoreAfter", scoreAfter,
		"vitalsScoreDiff", scoreDiff,
	)
	return scoreDiff
}

func (b *Bot) scoreVitalsFunc(lifePercentage, staminaPercentage, manaPercentage float32) float32 {
	f := func(x float32) float32 {
		if x > 1 {
			// cap score at 100%
			x = 1
		}
		// adding 2 just to make the score usually positive (50% resource == 0 score)
		return 2 - (float32(1) / x)
	}
	return 4*f(lifePercentage) + 2*f(staminaPercentage) + f(manaPercentage)
}
