package bot

import (
	"math"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

// Get vitals score for skill
// Tells you how much the skill will improve your resources (life, stamina, mana)
// Can be used for both casterEffect and targetEffect skills
func (b *Bot) scoreVitals(skillAttributes *swagger.DungeonsandtrollsSkillAttributes, skill *swagger.DungeonsandtrollsSkill) float32 {
	skillAttributes = fillSkillAttributes(*skillAttributes)
	skillLifeGain := float32(b.calculateAttributesValue(*skillAttributes.Life)) - skill.Cost.Life
	skillStaminaGain := float32(b.calculateAttributesValue(*skillAttributes.Stamina)) - skill.Cost.Stamina
	skillManaGain := float32(b.calculateAttributesValue(*skillAttributes.Mana)) - skill.Cost.Mana

	lifePercentage := b.Details.Monster.Attributes.Life / b.Details.Monster.MaxAttributes.Life
	staminaPercentage := b.Details.Monster.Attributes.Stamina / b.Details.Monster.MaxAttributes.Stamina
	if math.IsNaN(float64(staminaPercentage)) {
		staminaPercentage = 0
	}
	manaPercentage := b.Details.Monster.Attributes.Mana / b.Details.Monster.MaxAttributes.Mana
	if math.IsNaN(float64(manaPercentage)) {
		manaPercentage = 0
	}
	score := b.scoreVitalsFunc(lifePercentage, staminaPercentage, manaPercentage)

	lifePercentageAfter := (b.Details.Monster.Attributes.Life + skillLifeGain) / b.Details.Monster.MaxAttributes.Life
	staminaPercentageAfter := (b.Details.Monster.Attributes.Stamina + skillStaminaGain) / b.Details.Monster.MaxAttributes.Stamina
	if math.IsNaN(float64(staminaPercentageAfter)) {
		staminaPercentageAfter = 0
	}
	manaPercentageAfter := (b.Details.Monster.Attributes.Mana + skillManaGain) / b.Details.Monster.MaxAttributes.Mana
	if math.IsNaN(float64(manaPercentageAfter)) {
		manaPercentageAfter = 0
	}
	scoreAfter := b.scoreVitalsFunc(lifePercentageAfter, staminaPercentageAfter, manaPercentageAfter)

	scoreDiff := scoreAfter - score

	b.Logger.Infow("Skill vitals score",
		"skillName", skill.Name,
		"skill", skill,
		"skillAttributes", skillAttributes,
		"lifeGain", skillLifeGain,
		"staminaGain", skillStaminaGain,
		"manaGain", skillManaGain,
		"lifePercentage", lifePercentage,
		"life", b.Details.Monster.Attributes.Life,
		"lifeMax", b.Details.Monster.MaxAttributes.Life,
		"staminaPercentage", staminaPercentage,
		"stamina", b.Details.Monster.Attributes.Stamina,
		"staminaMax", b.Details.Monster.MaxAttributes.Stamina,
		"manaPercentage", manaPercentage,
		"mana", b.Details.Monster.Attributes.Mana,
		"manaMax", b.Details.Monster.MaxAttributes.Mana,
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
