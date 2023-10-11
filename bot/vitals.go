package bot

import (
	"math"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

func (b *Bot) scoreVitalsWithDamage(target *MapObject, skillAttributes *swagger.DungeonsandtrollsSkillAttributes, skill *swagger.DungeonsandtrollsSkill) float32 {
	damage := b.calculateAttributesValue(*skill.DamageAmount)
	damageAttrs := &swagger.DungeonsandtrollsAttributes{
		Life:    float32(damage),
		Stamina: 0,
		Mana:    0,
	}
	return b.scoreVitalsFor(target, skillAttributes, damageAttrs, -1, skill)
}

func (b *Bot) scoreVitalsWithCost(skillAttributes *swagger.DungeonsandtrollsSkillAttributes, skill *swagger.DungeonsandtrollsSkill) float32 {
	// Adjust cost by duration
	// It will be multiplied by duration again
	duration := float32(b.calculateAttributesValue(*skill.Duration))
	if duration == 0 {
		duration = 1
	}
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
// XXX: Add other attributes after you test, debug, and balance the current version
func (b *Bot) scoreVitalsFor(target *MapObject, skillAttributes *swagger.DungeonsandtrollsSkillAttributes, extraAttributes *swagger.DungeonsandtrollsAttributes, extraSign float32, skill *swagger.DungeonsandtrollsSkill) float32 {
	targetAttrs := target.GetAttributes()
	targetMaxAttrs := target.GetMaxAttributes()
	skillAttributes = fillSkillAttributes(*skillAttributes)

	b.Logger.Debugw("Debug scoreVitalsFor",
		"extraSign", extraSign,
		"skillAttributes", skillAttributes,
		"extraAttributes", extraAttributes,
		"targetAttrs", targetAttrs,
		"targetMaxAttrs", targetMaxAttrs,
	)

	skillLifeGain := float32(calculateAttributesValue(*targetAttrs, *skillAttributes.Life)) + extraSign*extraAttributes.Life
	skillStaminaGain := float32(calculateAttributesValue(*targetAttrs, *skillAttributes.Stamina)) + extraSign*extraAttributes.Stamina
	skillManaGain := float32(calculateAttributesValue(*targetAttrs, *skillAttributes.Mana)) + extraSign*extraAttributes.Mana

	// Apply duration
	// XXX: This might make over time skills too significant
	duration := float32(b.calculateAttributesValue(*skill.Duration))
	if duration == 0 {
		duration = 1
	}
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

	b.Logger.Infow("Skill vitals score",
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
		"duration", duration,
	)
	return scoreDiff
}

func (b *Bot) scoreVitalsFunc(lifePercentage, staminaPercentage, manaPercentage float32) float32 {
	cleanUp := func(x float32) float32 {
		if math.IsNaN(float64(x)) {
			x = 0
		}
		if x < 0 {
			return 0
		}
		if x > 1 {
			return 1
		}
		return x
	}
	//
	// Check out the curves on Wolfram Alpha:
	// https://www.wolframalpha.com/input?i=log%28%28x*100+%2B+1%29%29+%2F+log%28101%29+for+x+from+0+to+1
	// https://www.wolframalpha.com/input?i=log%28%28x*100+%2B+1%29%29+%2F+log%28101%29+for+x+in+%280%2C+0.1%2C+0.2%2C+0.3%2C+0.4%2C+0.5%2C+0.6%2C+0.7%2C+0.8%2C+0.9%2C+1%29
	//
	// curve aggression for x in {0, 0.1, 0.2, ...}:
	// 10: {0, 0.289065, 0.458157, 0.57813, 0.671188, 0.747222, 0.811508, 0.867194, 0.916314, 0.960253, 1}
	// 25: {0, 0.384508, 0.549941, 0.656846, 0.73598, 0.798837, 0.850984, 0.895545, 0.934448, 0.968971, 1}
	// 50: {0, 0.455707, 0.609868, 0.705166, 0.774328, 0.828647, 0.873382, 0.911413, 0.944491, 0.973757, 1}
	// 75: {0, 0.494158, 0.640212, 0.728976, 0.792934, 0.842965, 0.884063, 0.918939, 0.949233, 0.976009, 1}
	// 100: {0, 0.519574, 0.659684, 0.744073, 0.804653, 0.851944, 0.89074, 0.923633, 0.952185, 0.977409, 1}
	f := func(x float32, curveAggression float32) float32 {
		x = cleanUp(x)
		res := float32(math.Log(float64((x*curveAggression + 1))) / math.Log(float64(curveAggression)+1))
		if res == 0 {
			// Killing blow etc. should always have high score
			res = -0.5
		}
		return res
	}
	return 7*f(lifePercentage, 75) + 1.5*f(staminaPercentage, 25) + 1*f(manaPercentage, 10)
}
