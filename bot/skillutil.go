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
	)
	if *skill.Target == swagger.CHARACTER_SkillTarget {
		return &swagger.DungeonsandtrollsCommandsBatch{
			Skill: &swagger.DungeonsandtrollsSkillUse{
				SkillId:  skill.Id,
				TargetId: target.GetId(),
			},
			Yell: &swagger.DungeonsandtrollsMessage{
				Text: "Using skill " + skill.Name + "!",
			},
		}
	}
	if *skill.Target == swagger.POSITION_SkillTarget {
		return &swagger.DungeonsandtrollsCommandsBatch{
			Skill: &swagger.DungeonsandtrollsSkillUse{
				SkillId: skill.Id,
				Position: &swagger.DungeonsandtrollsPosition{
					PositionX: target.MapObjects.Position.PositionX,
					PositionY: target.MapObjects.Position.PositionY,
				},
			},
			Yell: &swagger.DungeonsandtrollsMessage{
				Text: "Using skill " + skill.Name + "!",
			},
		}
	}
	if *skill.Target == swagger.NONE_SkillTarget {
		return &swagger.DungeonsandtrollsCommandsBatch{
			Skill: &swagger.DungeonsandtrollsSkillUse{
				SkillId: skill.Id,
			},
			Yell: &swagger.DungeonsandtrollsMessage{
				Text: "Using skill " + skill.Name + "!",
			},
		}
	}
	b.Logger.Errorw("PANIC: Unknown skill target",
		"skillTarget", *skill.Target,
		"skill", skill,
	)
	return nil
}
