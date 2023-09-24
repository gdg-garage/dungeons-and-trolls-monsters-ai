package bot

import (
	"math/rand"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

const (
	// 3 and less is melee
	MeleeRange = 3
)

func (b *Bot) pickSkill() *swagger.DungeonsandtrollsSkill {

	skills := getAllSkills(b.GameState.Character.Equip)
	numAll := len(skills)

	skills = b.filterRequirementsMetSkills(skills)
	numRequirementsMet := len(skills)
	skills = b.filterDamageSkills(skills)
	numWithDamage := len(skills)
	skills = b.filterMeleeSkills(skills)
	numMelee := len(skills)

	b.Logger.Debugw("Picking skills ...",
		"1_numAll", numAll,
		"2_numRequirementsMet", numRequirementsMet,
		"3_numWithDamage", numWithDamage,
		"4_numMelee", numMelee,
		"finalSkills", skills,
	)

	x := rand.Intn(len(skills))
	return &skills[x]
}

func getAllSkills(equip []swagger.DungeonsandtrollsItem) []swagger.DungeonsandtrollsSkill {
	skills := []swagger.DungeonsandtrollsSkill{}
	for _, item := range equip {
		skills = append(skills, item.Skills...)
	}
	return skills
}

// Range

func (b *Bot) filterMeleeSkills(skills []swagger.DungeonsandtrollsSkill) []swagger.DungeonsandtrollsSkill {
	filtered := []swagger.DungeonsandtrollsSkill{}
	for _, skill := range skills {
		if b.calculateAttributesValue(*skill.Range_) <= MeleeRange {
			filtered = append(filtered, skill)
		}
	}
	return filtered
}

func (b *Bot) filterRangedSkills(skills []swagger.DungeonsandtrollsSkill) []swagger.DungeonsandtrollsSkill {
	filtered := []swagger.DungeonsandtrollsSkill{}
	for _, skill := range skills {
		if b.calculateAttributesValue(*skill.Range_) > MeleeRange {
			filtered = append(filtered, skill)
		}
	}
	return filtered
}

// Can cast

func (b *Bot) filterRequirementsMetSkills(skills []swagger.DungeonsandtrollsSkill) []swagger.DungeonsandtrollsSkill {
	filtered := []swagger.DungeonsandtrollsSkill{}
	for _, skill := range skills {
		if b.areAttributeRequirementMet(*skill.Cost) {
			filtered = append(filtered, skill)
		}
	}
	return filtered
}

func (b *Bot) filterRequirementsMetSkills2(attrs swagger.DungeonsandtrollsAttributes, skills []swagger.DungeonsandtrollsSkill) []swagger.DungeonsandtrollsSkill {
	filtered := []swagger.DungeonsandtrollsSkill{}
	for _, skill := range skills {
		if areAttributeRequirementMet(attrs, *skill.Cost) {
			filtered = append(filtered, skill)
		}
	}
	return filtered
}

// Damage skills

func (b *Bot) filterDamageSkills(skills []swagger.DungeonsandtrollsSkill) []swagger.DungeonsandtrollsSkill {
	filtered := []swagger.DungeonsandtrollsSkill{}
	for _, skill := range skills {
		if b.calculateAttributesValue(*skill.DamageAmount) > 0 {
			filtered = append(filtered, skill)
		}
	}
	return filtered
}

func (b *Bot) filterDamageSkills2(attrs swagger.DungeonsandtrollsAttributes, skills []swagger.DungeonsandtrollsSkill) []swagger.DungeonsandtrollsSkill {
	filtered := []swagger.DungeonsandtrollsSkill{}
	for _, skill := range skills {
		if calculateAttributesValue(attrs, *skill.DamageAmount) > 0 {
			filtered = append(filtered, skill)
		}
	}
	return filtered
}

// Healing skills

func (b *Bot) filterHealingSkills(skills []swagger.DungeonsandtrollsSkill) []swagger.DungeonsandtrollsSkill {
	filtered := []swagger.DungeonsandtrollsSkill{}
	for _, skill := range skills {
		if b.calculateAttributesValue(*skill.TargetEffects.Attributes.Life) < 0 {
			filtered = append(filtered, skill)
		}
	}
	return filtered
}

// Movement skills

func (b *Bot) filterMovementSkills(skills []swagger.DungeonsandtrollsSkill) []swagger.DungeonsandtrollsSkill {
	filtered := []swagger.DungeonsandtrollsSkill{}
	for _, skill := range skills {
		if skill.CasterEffects.Flags.Movement {
			filtered = append(filtered, skill)
		}
	}
	return filtered
}

// Other skills ???

func (b *Bot) filterNoDamageSkills(skills []swagger.DungeonsandtrollsSkill) []swagger.DungeonsandtrollsSkill {
	filtered := []swagger.DungeonsandtrollsSkill{}
	for _, skill := range skills {
		if b.calculateAttributesValue(*skill.DamageAmount) == 0 {
			filtered = append(filtered, skill)
		}
	}
	return filtered
}
