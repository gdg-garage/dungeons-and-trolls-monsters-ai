package bot

import (
	"log"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
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
	return calculateAttributesValue(*b.GameState.Character.Attributes, attrs)
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
	value += attrs.Scalar
	return int(value)
}

func (b *Bot) areAttributeRequirementMet(attrs swagger.DungeonsandtrollsAttributes) bool {
	return areAttributeRequirementMet(*b.GameState.Character.Attributes, attrs)
}

func areAttributeRequirementMet(myAttrs swagger.DungeonsandtrollsAttributes, attrs swagger.DungeonsandtrollsAttributes) bool {
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
		return true
	}
	// I bet this will be useful
	if myAttrs.Strength < attrs.Strength {
		log.Printf("Attribute check failed: Strength (have: %v < need: %v)\n", myAttrs.Strength, attrs.Strength)
		return false
	}
	if myAttrs.Dexterity < attrs.Dexterity {
		log.Printf("Attribute check failed: Dexterity (have: %v < need: %v)\n", myAttrs.Dexterity, attrs.Dexterity)
		return false
	}
	if myAttrs.Intelligence < attrs.Intelligence {
		log.Printf("Attribute check failed: Intelligence (have: %v < need: %v)\n", myAttrs.Intelligence, attrs.Intelligence)
		return false
	}
	if myAttrs.Willpower < attrs.Willpower {
		log.Printf("Attribute check failed: Willpower (have: %v < need: %v)\n", myAttrs.Willpower, attrs.Willpower)
		return false
	}
	if myAttrs.Constitution < attrs.Constitution {
		log.Printf("Attribute check failed: Constitution (have: %v < need: %v)\n", myAttrs.Constitution, attrs.Constitution)
		return false
	}
	if myAttrs.SlashResist < attrs.SlashResist {
		log.Printf("Attribute check failed: SlashResist (have: %v < need: %v)\n", myAttrs.SlashResist, attrs.SlashResist)
		return false
	}
	if myAttrs.PierceResist < attrs.PierceResist {
		log.Printf("Attribute check failed: PierceResist (have: %v < need: %v)\n", myAttrs.PierceResist, attrs.PierceResist)
		return false
	}
	if myAttrs.FireResist < attrs.FireResist {
		log.Printf("Attribute check failed: FireResist (have: %v < need: %v)\n", myAttrs.FireResist, attrs.FireResist)
		return false
	}
	if myAttrs.PoisonResist < attrs.PoisonResist {
		log.Printf("Attribute check failed: PoisonResist (have: %v < need: %v)\n", myAttrs.PoisonResist, attrs.PoisonResist)
		return false
	}
	if myAttrs.ElectricResist < attrs.ElectricResist {
		log.Printf("Attribute check failed: ElectricResist (have: %v < need: %v)\n", myAttrs.ElectricResist, attrs.ElectricResist)
		return false
	}
	if myAttrs.Life < attrs.Life {
		log.Printf("Attribute check failed: Life (have: %v < need: %v)\n", myAttrs.Life, attrs.Life)
		return false
	}
	if myAttrs.Stamina < attrs.Stamina {
		log.Printf("Attribute check failed: Stamina (have: %v < need: %v)\n", myAttrs.Stamina, attrs.Stamina)
		return false
	}
	if myAttrs.Mana < attrs.Mana {
		log.Printf("Attribute check failed: Mana (have: %v < need: %v)\n", myAttrs.Mana, attrs.Mana)
		return false
	}
	return false
}

func useSkill(skill swagger.DungeonsandtrollsSkill, target MapObject) *swagger.DungeonsandtrollsCommandsBatch {
	if *skill.Target == swagger.CHARACTER_SkillTarget {
		return &swagger.DungeonsandtrollsCommandsBatch{
			Skill: &swagger.DungeonsandtrollsSkillUse{
				SkillId:  skill.Id,
				TargetId: target.GetId(),
			},
		}
	}
	if *skill.Target == swagger.POSITION_SkillTarget {
		return &swagger.DungeonsandtrollsCommandsBatch{
			Skill: &swagger.DungeonsandtrollsSkillUse{
				SkillId:  skill.Id,
				Location: target.MapObjects.Position,
			},
		}
	}
	if *skill.Target == swagger.NONE_SkillTarget {
		return &swagger.DungeonsandtrollsCommandsBatch{
			Skill: &swagger.DungeonsandtrollsSkillUse{
				SkillId: skill.Id,
			},
		}
	}
	log.Panicln("ERROR: Unknown skill target:", *skill.Target)
	return nil
}
