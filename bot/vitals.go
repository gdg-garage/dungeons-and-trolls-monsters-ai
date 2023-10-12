package bot

import (
	"math"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

func (b *Bot) getResistForDamageType(target *MapObject, damageType swagger.DungeonsandtrollsDamageType) float32 {
	attrs := target.GetAttributes()
	switch damageType {
	case swagger.SLASH_DungeonsandtrollsDamageType:
		return attrs.SlashResist
	case swagger.PIERCE_DungeonsandtrollsDamageType:
		return attrs.PierceResist
	case swagger.FIRE_DungeonsandtrollsDamageType:
		return attrs.FireResist
	case swagger.POISON_DungeonsandtrollsDamageType:
		return attrs.PoisonResist
	case swagger.ELECTRIC_DungeonsandtrollsDamageType:
		return attrs.ElectricResist
	case swagger.NONE_DungeonsandtrollsDamageType:
		return 0
	}
	b.Logger.Error("FATAL: getResistForDamageType(): Unknown damage type!",
		"damageType", damageType,
	)
	return 0
}

func (b *Bot) calculateDamage(target *MapObject, skill *swagger.DungeonsandtrollsSkill) float32 {
	power := b.calculateAttributesValue(*skill.DamageAmount)
	resist := b.getResistForDamageType(target, *skill.DamageType)
	damage := float32(float64(power*10) / (float64(10) + math.Max(float64(resist), -5)))
	damageFinal := randomizeScoreN(damage, 20)
	b.Logger.Infow("Damage calculated",
		"targetName", target.GetName(),
		"power", power,
		"resist", resist,
		"damage", damage,
		"damageRandomized", damageFinal,
	)
	return damageFinal
}

func (b *Bot) scoreVitalsWithDamage(target *MapObject, skillAttributes *swagger.DungeonsandtrollsSkillAttributes, skill *swagger.DungeonsandtrollsSkill) (float32, float32, float32) {
	damage := b.calculateDamage(target, skill)
	damageAttrs := &swagger.DungeonsandtrollsAttributes{
		Life:    float32(damage),
		Stamina: 0,
		Mana:    0,
	}
	return b.scoreVitalsFor(target, skillAttributes, damageAttrs, -1, skill)
}

func (b *Bot) scoreVitalsWithCost(skillAttributes *swagger.DungeonsandtrollsSkillAttributes, skill *swagger.DungeonsandtrollsSkill) (float32, float32, float32) {
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
func (b *Bot) scoreVitalsFor(target *MapObject, skillAttributes *swagger.DungeonsandtrollsSkillAttributes, extraAttributes *swagger.DungeonsandtrollsAttributes, extraSign float32, skill *swagger.DungeonsandtrollsSkill) (float32, float32, float32) {
	targetAttrs := target.GetAttributes()
	targetMaxAttrs := target.GetMaxAttributes()
	skillAttributes = fillSkillAttributes(*skillAttributes)

	b.Logger.Infow("Debug scoreVitalsFor",
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

	lifePercentage, lifePercentageAfter := b.calculateAttributePercentages(targetAttrs.Life, targetMaxAttrs.Life, skillLifeGain)
	staminaPercentage, staminaPercentageAfter := b.calculateAttributePercentages(targetAttrs.Stamina, targetMaxAttrs.Stamina, skillStaminaGain)
	manaPercentage, manaPercentageAfter := b.calculateAttributePercentages(targetAttrs.Mana, targetMaxAttrs.Mana, skillManaGain)

	score := b.scoreVitalsFunc(lifePercentage, staminaPercentage, manaPercentage)
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
	scoreBuffsDiff, scoreResistsDiff := b.scoreBuffs(target, skillAttributes, skill)
	return scoreDiff, scoreBuffsDiff, scoreResistsDiff
}

func (b *Bot) scorePercentageOnACurve(percentage, curve, killingBlowBonus float32) float32 {
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
			res = -killingBlowBonus
		}
		return res
	}
	return f(percentage, curve)
}

func (b *Bot) scoreVitalsFunc(lifePercentage, staminaPercentage, manaPercentage float32) float32 {
	return 7.5*b.scorePercentageOnACurve(lifePercentage, 75, 0.5) +
		1.75*b.scorePercentageOnACurve(staminaPercentage, 25, 0.2) +
		1.2*b.scorePercentageOnACurve(manaPercentage, 10, 0.2)
}

func (b *Bot) scoreBuffsFunc(strPercentage, dexPercentage, intPercentage, willPercentage, consPercentage float32) float32 {
	f := func(percentage float32) float32 {
		return b.scorePercentageOnACurve(percentage, 25, 0.2)
	}
	return 4*f(strPercentage) + 2*f(dexPercentage) + 2*f(intPercentage) + 1*f(willPercentage) + 1*f(consPercentage)
}

func (b *Bot) scoreResistFunc(slashPercentage, piercePercentage, firePercentage, poisonPercentage, electricPercentage float32) float32 {
	f := func(percentage float32) float32 {
		return b.scorePercentageOnACurve(percentage, 25, 0.2)
	}
	return 2.5*f(slashPercentage) + 4*f(piercePercentage) + 1.5*f(firePercentage) + f(poisonPercentage) + f(electricPercentage)
}

func (b *Bot) calculateAttributePercentages(value, maxValue, gain float32) (float32, float32) {
	percentage := value / maxValue
	if math.IsNaN(float64(percentage)) {
		percentage = 0
	}
	percentageAfter := (value + gain) / maxValue
	if math.IsNaN(float64(percentageAfter)) {
		percentageAfter = 0
	}
	return percentage, percentageAfter
}

func (b *Bot) scoreBuffs(target *MapObject, skillAttributes *swagger.DungeonsandtrollsSkillAttributes, skill *swagger.DungeonsandtrollsSkill) (float32, float32) {
	strengthGain := b.calculateAttributesValue(*skillAttributes.Strength)
	dexterityGain := b.calculateAttributesValue(*skillAttributes.Dexterity)
	intelligenceGain := b.calculateAttributesValue(*skillAttributes.Intelligence)
	willpowerGain := b.calculateAttributesValue(*skillAttributes.Willpower)
	constitutionGain := b.calculateAttributesValue(*skillAttributes.Constitution)

	slashResistGain := b.calculateAttributesValue(*skillAttributes.SlashResist)
	pierceResistGain := b.calculateAttributesValue(*skillAttributes.PierceResist)
	fireResistGain := b.calculateAttributesValue(*skillAttributes.FireResist)
	poisonResistGain := b.calculateAttributesValue(*skillAttributes.PoisonResist)
	electricResistGain := b.calculateAttributesValue(*skillAttributes.ElectricResist)

	strengthPercentage, strengthPercentageAfter := b.calculateAttributePercentages(target.GetAttributes().Strength, target.GetMaxAttributes().Strength, strengthGain)
	dexterityPercentage, dexterityPercentageAfter := b.calculateAttributePercentages(target.GetAttributes().Dexterity, target.GetMaxAttributes().Dexterity, dexterityGain)
	intelligencePercentage, intelligencePercentageAfter := b.calculateAttributePercentages(target.GetAttributes().Intelligence, target.GetMaxAttributes().Intelligence, intelligenceGain)
	willpowerPercentage, willpowerPercentageAfter := b.calculateAttributePercentages(target.GetAttributes().Willpower, target.GetMaxAttributes().Willpower, willpowerGain)
	constitutionPercentage, constitutionPercentageAfter := b.calculateAttributePercentages(target.GetAttributes().Constitution, target.GetMaxAttributes().Constitution, constitutionGain)

	slashResistPercentage, slashResistPercentageAfter := b.calculateAttributePercentages(target.GetAttributes().SlashResist, target.GetMaxAttributes().SlashResist, slashResistGain)
	pierceResistPercentage, pierceResistPercentageAfter := b.calculateAttributePercentages(target.GetAttributes().PierceResist, target.GetMaxAttributes().PierceResist, pierceResistGain)
	fireResistPercentage, fireResistPercentageAfter := b.calculateAttributePercentages(target.GetAttributes().FireResist, target.GetMaxAttributes().FireResist, fireResistGain)
	poisonResistPercentage, poisonResistPercentageAfter := b.calculateAttributePercentages(target.GetAttributes().PoisonResist, target.GetMaxAttributes().PoisonResist, poisonResistGain)
	electricResistPercentage, electricResistPercentageAfter := b.calculateAttributePercentages(target.GetAttributes().ElectricResist, target.GetMaxAttributes().ElectricResist, electricResistGain)

	scoreBuffs := b.scoreBuffsFunc(strengthPercentage, dexterityPercentage, intelligencePercentage, willpowerPercentage, constitutionPercentage)
	scoreBuffsAfter := b.scoreBuffsFunc(strengthPercentageAfter, dexterityPercentageAfter, intelligencePercentageAfter, willpowerPercentageAfter, constitutionPercentageAfter)
	scoreResists := b.scoreResistFunc(slashResistPercentage, pierceResistPercentage, fireResistPercentage, poisonResistPercentage, electricResistPercentage)
	scoreResistsAfter := b.scoreResistFunc(slashResistPercentageAfter, pierceResistPercentageAfter, fireResistPercentageAfter, poisonResistPercentageAfter, electricResistPercentageAfter)

	return scoreBuffsAfter - scoreBuffs, scoreResistsAfter - scoreResists
}
