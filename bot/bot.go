package bot

import (
	"math/rand"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
	"go.uber.org/zap"
)

type BotState struct {
	Objects     MapObjectsByCategory
	MapExtended map[swagger.DungeonsandtrollsPosition]MapCellExt
	Self        MapObject
	Yell        string
	PrefixYell  string

	State        string
	TargetObject swagger.DungeonsandtrollsMapObjects
	Target       swagger.DungeonsandtrollsMonster
	Mood         Mood
}

// BotState is managed by bot algorithm
// GameState is the current state of the game
// MonsterDetails are parts of game state passed from dispatcher for convenience
type Bot struct {
	MonsterId string

	BotState  BotState
	GameState *swagger.DungeonsandtrollsGameState
	Details   MonsterDetails

	PrevBotState  BotState
	PrevGameState *swagger.DungeonsandtrollsGameState
	PrevDetails   MonsterDetails

	Logger *zap.SugaredLogger
}

func (b *Bot) Run5() *swagger.DungeonsandtrollsCommandsBatch {
	b.BotState.Self = NewMonsterMapObject(*b.Details.MapObjects, b.Details.Index)
	monster := b.Details.Monster
	// monsterTileObjects := b.Details.MapObjects
	level := b.Details.Level
	position := b.Details.Position

	b.Logger.Infow("Handling monster",
		"monster", monster,
	)
	if monster.Faction == "neutral" {
		b.Logger.Warnw("Skipping neutral monster")
		return b.Yell("I'm a neutral monster!")
	}
	if monster.Attributes.Life <= 0 {
		b.Logger.Warnw("Skipping DEAD monster ☠")
		return nil
	}
	// calculate distance and line of sight
	b.BotState.MapExtended = b.calculateDistanceAndLineOfSight(level, *position)
	b.BotState.Objects = b.getMapObjectsByCategoryForLevel(level)

	combatCmd := b.combat()
	if combatCmd != nil {
		return combatCmd
	}

	random := rand.Intn(2)
	switch random {
	case 0:
		return b.rest()
	case 1:
		return b.randomWalkFromPositionExt(level, *b.Details.Position, b.BotState.MapExtended)
	default:
		return b.Yell("Nothing to do ...")
	}
}

func (b *Bot) combat() *swagger.DungeonsandtrollsCommandsBatch {
	if len(b.BotState.Objects.Hostile) <= 0 {
		return nil
	}
	skills := getAllSkills(b.Details.Monster.EquippedItems)
	b.Logger.Infow("All skills",
		"skills", skills,
		"numSkills", len(skills),
	)
	dmgSkills := b.filterDamageSkills2(*b.Details.Monster.Attributes, skills)
	if len(skills) > 0 && len(dmgSkills) == 0 {
		b.BotState.PrefixYell = "NO DMG SKILLS!"
		b.Logger.Errorw("NO DMG SKILLS!",
			"allSkills", skills,
			"monsterAttributes", b.Details.Monster.Attributes,
			"monster", b.Details.Monster,
		)
	}
	skills2 := b.filterRequirementsMetSkills(dmgSkills)
	if len(skills2) == 0 {
		b.Logger.Errorw("No skills available")
		return nil
	}
	if len(dmgSkills)-len(skills2) >= 2 {
		b.BotState.PrefixYell = "Combat rest"
		return b.rest()
	}
	if len(dmgSkills) > 0 && len(skills2) == 0 {
		b.BotState.PrefixYell = "Out of stamina (?)"
		b.Logger.Errorw("Can't damage because I'm out of stamina and/or other resources",
			"allSkills", skills,
			"monsterAttributes", b.Details.Monster.Attributes,
			"monster", b.Details.Monster,
		)
	}
	enemies := b.BotState.Objects.Hostile
	if len(enemies) <= 0 {
		b.Logger.Warnw("No enemies found")
		return nil
	}
	b.Logger.Infow("Choosing skills & enemies",
		"skills", skills2,
		"numSkills", len(skills),
		// "enemies", enemies,
	)
	var bestSkill *swagger.DungeonsandtrollsSkill
	bestDmg := int32(0)
	bestEnemy := &MapObject{}
	for _, enemy := range enemies {
		for _, skill := range skills2 {
			skillResult := b.evaluateSkill(skill, enemy)
			b.Logger.Infow("Skill evaluated",
				"skillName", skill.Name,
				"skill", skill,
				zap.Any("skillResult", skillResult),
				"position", enemy.MapObjects.Position,
			)
			if skillResult != nil && skillResult.Damage > bestDmg {
				b.Logger.Infow("Found better skill",
					"skillName", skill.Name,
					"skill", skill,
					"damage", skillResult.Damage,
					"position", enemy.MapObjects.Position,
				)
				bestSkill = &skill
				bestDmg = skillResult.Damage
				bestEnemy = &enemy
			}
		}
	}
	if bestSkill != nil {
		b.Logger.Infow("Using best skill available",
			"skillName", bestSkill.Name,
			"skill", bestDmg,
			"damage", bestDmg,
			"targetName", bestEnemy.GetName(),
		)
		return b.useSkill(*bestSkill, *bestEnemy)
	}
	b.Logger.Warnw("No skill chosen")

	random := rand.Intn(2)
	switch random {
	case 0:
		return b.rest()
	default:
		return b.moveTowardsEnemy(enemies)
	}
}

func (b *Bot) moveTowardsEnemy(enemies []MapObject) *swagger.DungeonsandtrollsCommandsBatch {
	// Go to player
	magicDistance := 13 // distance threshold
	closeEnemies := []MapObject{}
	for _, enemy := range enemies {
		if b.BotState.MapExtended[*enemy.MapObjects.Position].distance < magicDistance {
			closeEnemies = append(closeEnemies, enemy)
		}
	}
	if len(closeEnemies) == 0 {
		return nil
	}
	rp := rand.Intn(len(closeEnemies))
	b.BotState.Yell = "I'm coming for you " + closeEnemies[rp].GetName() + "!"
	return &swagger.DungeonsandtrollsCommandsBatch{
		Move: closeEnemies[rp].MapObjects.Position,
	}
}

func (b *Bot) jumpAway() *swagger.DungeonsandtrollsCommandsBatch {
	return nil
}

func (b *Bot) rest() *swagger.DungeonsandtrollsCommandsBatch {
	b.Logger.Debugw("Picking a support skill ...")
	// Rest & Heal, etc.
	skills := getAllSkills(b.Details.Monster.EquippedItems)
	skills2 := b.filterRequirementsMetSkills(skills)
	if len(skills2) <= 0 {
		return nil
	}
	supportSkills := []swagger.DungeonsandtrollsSkill{}
	for _, skill := range skills2 {
		b.Logger.Debugw("Checking skill",
			"skillName", skill.Name,
			"skill", skill,
		)
		casterEffects := skill.CasterEffects
		if casterEffects == nil {
			continue
		}
		skillAttributes := casterEffects.Attributes
		if skillAttributes == nil {
			continue
		}
		skillLife := skillAttributes.Life
		skillMana := skillAttributes.Mana
		skillStamina := skillAttributes.Stamina
		b.Logger.Debugw("Skill attributes",
			"skillName", skill.Name,
			"skill", skill,
			"casterEffects", casterEffects,
			"skillAttributes", skillAttributes,
			"skillLife", skillLife,
			"skillMana", skillMana,
			"skillStamina", skillStamina,
		)
		attrs := b.Details.Monster.Attributes
		if (skillLife != nil && calculateAttributesValue(*attrs, *skillLife) > 0) ||
			(skillMana != nil && calculateAttributesValue(*attrs, *skillMana) > 0) ||
			(skillStamina != nil && calculateAttributesValue(*attrs, *skillStamina) > 0) {
			supportSkills = append(supportSkills, skill)
		}
	}
	if len(supportSkills) <= 0 {
		return nil
	}
	rp := rand.Intn(len(supportSkills))
	// TODO: Don't attack yourself 🙈
	b.Logger.Infow("Picked support skill",
		"allSupportSkills", supportSkills,
		"pickedSkill", supportSkills[rp],
	)
	return b.useSkill(supportSkills[rp], b.BotState.Self)
}

func (b *Bot) shop() *swagger.DungeonsandtrollsItem {
	shop := b.GameState.ShopItems
	money := b.GameState.Character.Money
	for _, item := range shop {
		if item.Price <= money {
			if *item.Slot == swagger.MAIN_HAND_DungeonsandtrollsItemType {
				if len(item.Skills) > 0 {
					if item.Skills[0].DamageAmount.Constant > 0 {
						b.Logger.Infow("Found item to buy ...",
							"itemName", item.Name,
						)
						return &item
					}
				}
			}
		}
	}
	return nil
}
