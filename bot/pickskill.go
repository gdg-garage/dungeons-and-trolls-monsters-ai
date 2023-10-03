package bot

import (
	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

func getAllSkills(equip []swagger.DungeonsandtrollsItem) []swagger.DungeonsandtrollsSkill {
	skills := []swagger.DungeonsandtrollsSkill{}
	for _, item := range equip {
		skills = append(skills, item.Skills...)
	}
	return skills
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

func (b *Bot) filterCastableWithOOCSkills(skills []swagger.DungeonsandtrollsSkill) []swagger.DungeonsandtrollsSkill {
	outOfCombatTurnsConstant := int32(2)
	if b.Details.Monster.LastDamageTaken > outOfCombatTurnsConstant {
		return skills
	}
	filtered := []swagger.DungeonsandtrollsSkill{}
	for _, skill := range skills {
		if !skill.Flags.RequiresOutOfCombat {
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
