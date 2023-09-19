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
	getAllSkills(b.GameState.Character.Equip)

	skills := getAllSkills(b.GameState.Character.Equip)
	skills = b.filterRequirementsMetSkills(skills)
	skills = b.filterDamageSkills(skills)

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
		if b.calculateAttributesValue(*skill.DamageAmount) < 0 {
			filtered = append(filtered, skill)
		}
	}
	return filtered
}

// Movement skills

func (b *Bot) filterMovementSkills(skills []swagger.DungeonsandtrollsSkill) []swagger.DungeonsandtrollsSkill {
	filtered := []swagger.DungeonsandtrollsSkill{}
	for _, skill := range skills {
		for _, flag := range skill.CasterEffects.Flags {
			if flag == "move" {
				filtered = append(filtered, skill)
				break
			}
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
