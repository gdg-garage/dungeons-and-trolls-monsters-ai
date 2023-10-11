package bot

import (
	"fmt"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
	"go.uber.org/zap"
)

/*
	Strength float32 `json:"strength,omitempty"`
	Dexterity float32 `json:"dexterity,omitempty"`
	Intelligence float32 `json:"intelligence,omitempty"`
	Willpower float32 `json:"willpower,omitempty"`
	Constitution float32 `json:"constitution,omitempty"`
	SlashResist float32 `json:"slashResist,omitempty"`
	PierceResist float32 `json:"pierceResist,omitempty"`
	FireResist float32 `json:"fireResist,omitempty"`
	PoisonResist float32 `json:"poisonResist,omitempty"`
	ElectricResist float32 `json:"electricResist,omitempty"`
	Life float32 `json:"life,omitempty"`
	Stamina float32 `json:"stamina,omitempty"`
	Mana float32 `json:"mana,omitempty"`
	Scalar float32 `json:"scalar,omitempty"`
*/

func (b *Bot) calculateAttributesValue(attrs swagger.DungeonsandtrollsAttributes) int {
	return calculateAttributesValue(*b.Details.Monster.Attributes, attrs)
}

func calculateAttributesValue(myAttrs swagger.DungeonsandtrollsAttributes, attrs swagger.DungeonsandtrollsAttributes) int {
	var value float32
	value += myAttrs.Strength * attrs.Strength
	value += myAttrs.Dexterity * attrs.Dexterity
	value += myAttrs.Intelligence * attrs.Intelligence
	value += myAttrs.Willpower * attrs.Willpower
	value += myAttrs.Constitution * attrs.Constitution
	value += myAttrs.SlashResist * attrs.SlashResist
	value += myAttrs.PierceResist * attrs.PierceResist
	value += myAttrs.FireResist * attrs.FireResist
	value += myAttrs.PoisonResist * attrs.PoisonResist
	value += myAttrs.ElectricResist * attrs.ElectricResist
	value += myAttrs.Life * attrs.Life
	value += myAttrs.Stamina * attrs.Stamina
	value += myAttrs.Mana * attrs.Mana
	value += attrs.Constant
	return int(value)
}

func (b *Bot) areAttributeRequirementMet(attrs swagger.DungeonsandtrollsAttributes) bool {
	err := areAttributeRequirementMet(*b.Details.Monster.Attributes, attrs)
	if err != nil {
		b.Logger.Warnw("Attribute requirement not met",
			zap.Error(err),
			"myAttributes", *b.Details.Monster.Attributes,
			"requiredAttributes", attrs,
		)
		return false
	}
	return true
}

func areAttributeRequirementMet(myAttrs swagger.DungeonsandtrollsAttributes, attrs swagger.DungeonsandtrollsAttributes) error {
	if myAttrs.Strength >= attrs.Strength &&
		myAttrs.Dexterity >= attrs.Dexterity &&
		myAttrs.Intelligence >= attrs.Intelligence &&
		myAttrs.Willpower >= attrs.Willpower &&
		myAttrs.Constitution >= attrs.Constitution &&
		myAttrs.SlashResist >= attrs.SlashResist &&
		myAttrs.PierceResist >= attrs.PierceResist &&
		myAttrs.FireResist >= attrs.FireResist &&
		myAttrs.PoisonResist >= attrs.PoisonResist &&
		myAttrs.ElectricResist >= attrs.ElectricResist &&
		myAttrs.Life >= attrs.Life &&
		myAttrs.Stamina >= attrs.Stamina &&
		myAttrs.Mana >= attrs.Mana {
		return nil
	}
	if myAttrs.Strength < attrs.Strength {
		return fmt.Errorf("Attribute check failed: Strength (have: %v < need: %v)\n", myAttrs.Strength, attrs.Strength)
	}
	if myAttrs.Dexterity < attrs.Dexterity {
		return fmt.Errorf("Attribute check failed: Dexterity (have: %v < need: %v)\n", myAttrs.Dexterity, attrs.Dexterity)
	}
	if myAttrs.Intelligence < attrs.Intelligence {
		return fmt.Errorf("Attribute check failed: Intelligence (have: %v < need: %v)\n", myAttrs.Intelligence, attrs.Intelligence)
	}
	if myAttrs.Willpower < attrs.Willpower {
		return fmt.Errorf("Attribute check failed: Willpower (have: %v < need: %v)\n", myAttrs.Willpower, attrs.Willpower)
	}
	if myAttrs.Constitution < attrs.Constitution {
		return fmt.Errorf("Attribute check failed: Constitution (have: %v < need: %v)\n", myAttrs.Constitution, attrs.Constitution)
	}
	if myAttrs.SlashResist < attrs.SlashResist {
		return fmt.Errorf("Attribute check failed: SlashResist (have: %v < need: %v)\n", myAttrs.SlashResist, attrs.SlashResist)
	}
	if myAttrs.PierceResist < attrs.PierceResist {
		return fmt.Errorf("Attribute check failed: PierceResist (have: %v < need: %v)\n", myAttrs.PierceResist, attrs.PierceResist)
	}
	if myAttrs.FireResist < attrs.FireResist {
		return fmt.Errorf("Attribute check failed: FireResist (have: %v < need: %v)\n", myAttrs.FireResist, attrs.FireResist)
	}
	if myAttrs.PoisonResist < attrs.PoisonResist {
		return fmt.Errorf("Attribute check failed: PoisonResist (have: %v < need: %v)\n", myAttrs.PoisonResist, attrs.PoisonResist)
	}
	if myAttrs.ElectricResist < attrs.ElectricResist {
		return fmt.Errorf("Attribute check failed: ElectricResist (have: %v < need: %v)\n", myAttrs.ElectricResist, attrs.ElectricResist)
	}
	if myAttrs.Life < attrs.Life {
		return fmt.Errorf("Attribute check failed: Life (have: %v < need: %v)\n", myAttrs.Life, attrs.Life)
	}
	if myAttrs.Stamina < attrs.Stamina {
		return fmt.Errorf("Attribute check failed: Stamina (have: %v < need: %v)\n", myAttrs.Stamina, attrs.Stamina)
	}
	if myAttrs.Mana < attrs.Mana {
		return fmt.Errorf("Attribute check failed: Mana (have: %v < need: %v)\n", myAttrs.Mana, attrs.Mana)
	}
	return fmt.Errorf("PANIC: Attribute check failed with UNKNOWN REASON!")
}

func (b *Bot) useSkill(skill swagger.DungeonsandtrollsSkill, target MapObject) *swagger.DungeonsandtrollsCommandsBatch {
	b.Logger.Infow("Using skill",
		"skillName", skill.Name,
		"skill", skill,
		"skillTargetType", skill.Target,
		"target", target.GetName(),
		"targetPosition", target.GetPosition(),
	)
	if isDefaultMoveSkill(skill) {
		b.addFirstYell("Catch me :)")
		if b.BotState.TargetPosition == nil {
			stretchedPosition := b.stretchMovePosition(*target.GetPosition())
			b.BotState.TargetPosition = &stretchedPosition
			b.BotState.TargetPositionTimeout = 6
			b.Logger.Infow("Calculated stretched position",
				"movePosition", target.MapObjects.Position,
				"stretchedPosition", stretchedPosition,
				"myPosition", b.Details.Position,
			)
		}
		return &swagger.DungeonsandtrollsCommandsBatch{
			Move: target.MapObjects.Position,
		}
	}
	if *skill.Target == swagger.CHARACTER_SkillTarget {
		msg := fmt.Sprintf(
			"Using skill %s! -> [%d, %d] %s",
			skill.Name,
			target.MapObjects.Position.PositionX,
			target.MapObjects.Position.PositionY,
			target.GetName())
		b.addFirstYell(msg)
		return &swagger.DungeonsandtrollsCommandsBatch{
			Skill: &swagger.DungeonsandtrollsSkillUse{
				SkillId:  skill.Id,
				TargetId: target.GetId(),
			},
		}
	}
	if *skill.Target == swagger.POSITION_SkillTarget {
		msg := fmt.Sprintf(
			"Using skill %s! -> [%d, %d]",
			skill.Name,
			target.MapObjects.Position.PositionX,
			target.MapObjects.Position.PositionY)
		b.addFirstYell(msg)
		return &swagger.DungeonsandtrollsCommandsBatch{
			Skill: &swagger.DungeonsandtrollsSkillUse{
				SkillId: skill.Id,
				Position: &swagger.DungeonsandtrollsPosition{
					PositionX: target.MapObjects.Position.PositionX,
					PositionY: target.MapObjects.Position.PositionY,
				},
			},
		}
	}
	if *skill.Target == swagger.NONE_SkillTarget {
		b.addFirstYell("Using skill " + skill.Name + ".")
		return &swagger.DungeonsandtrollsCommandsBatch{
			Skill: &swagger.DungeonsandtrollsSkillUse{
				SkillId: skill.Id,
			},
		}
	}
	b.Logger.Errorw("PANIC: Unknown skill target",
		"skillTarget", *skill.Target,
		"skill", skill,
	)
	return nil
}

func fillSkillAttributes(skillAttrs swagger.DungeonsandtrollsSkillAttributes) *swagger.DungeonsandtrollsSkillAttributes {
	if skillAttrs.Strength == nil {
		skillAttrs.Strength = &swagger.DungeonsandtrollsAttributes{}
	}
	if skillAttrs.Dexterity == nil {
		skillAttrs.Dexterity = &swagger.DungeonsandtrollsAttributes{}
	}
	if skillAttrs.Intelligence == nil {
		skillAttrs.Intelligence = &swagger.DungeonsandtrollsAttributes{}
	}
	if skillAttrs.Willpower == nil {
		skillAttrs.Willpower = &swagger.DungeonsandtrollsAttributes{}
	}
	if skillAttrs.Constitution == nil {
		skillAttrs.Constitution = &swagger.DungeonsandtrollsAttributes{}
	}
	if skillAttrs.SlashResist == nil {
		skillAttrs.SlashResist = &swagger.DungeonsandtrollsAttributes{}
	}
	if skillAttrs.PierceResist == nil {
		skillAttrs.PierceResist = &swagger.DungeonsandtrollsAttributes{}
	}
	if skillAttrs.FireResist == nil {
		skillAttrs.FireResist = &swagger.DungeonsandtrollsAttributes{}
	}
	if skillAttrs.PoisonResist == nil {
		skillAttrs.PoisonResist = &swagger.DungeonsandtrollsAttributes{}
	}
	if skillAttrs.ElectricResist == nil {
		skillAttrs.ElectricResist = &swagger.DungeonsandtrollsAttributes{}
	}
	if skillAttrs.Life == nil {
		skillAttrs.Life = &swagger.DungeonsandtrollsAttributes{}
	}
	if skillAttrs.Stamina == nil {
		skillAttrs.Stamina = &swagger.DungeonsandtrollsAttributes{}
	}
	if skillAttrs.Mana == nil {
		skillAttrs.Mana = &swagger.DungeonsandtrollsAttributes{}
	}
	return &skillAttrs
}

// Filtering

func getAllSkills(equip []swagger.DungeonsandtrollsItem) []swagger.DungeonsandtrollsSkill {
	skills := []swagger.DungeonsandtrollsSkill{}
	for _, item := range equip {
		skills = append(skills, item.Skills...)
	}
	return skills
}

// Can cast

func (b *Bot) filterActiveSkills(skills []swagger.DungeonsandtrollsSkill) []swagger.DungeonsandtrollsSkill {
	filtered := []swagger.DungeonsandtrollsSkill{}
	for _, skill := range skills {
		if !skill.Flags.Passive {
			filtered = append(filtered, skill)
		}
	}
	return filtered
}

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
