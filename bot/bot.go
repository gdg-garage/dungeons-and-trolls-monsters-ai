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
	Yells       []string

	State          string
	TargetPosition swagger.DungeonsandtrollsPosition
	// TargetObject   swagger.DungeonsandtrollsMapObjects
	// Target         swagger.DungeonsandtrollsMonster
	// Mood           Mood
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

func (b *Bot) Run() *swagger.DungeonsandtrollsCommandsBatch {
	b.BotState.Self = NewMonsterMapObject(*b.Details.MapObjects, b.Details.Index)
	b.BotState.Yells = []string{}
	monster := b.Details.Monster
	// monsterTileObjects := b.Details.MapObjects
	level := b.Details.Level
	position := b.Details.Position

	b.Logger.Infow("Handling monster",
		"monster", monster,
		"position", position,
	)
	if monster.Algorithm == "none" {
		b.Logger.Warnw("Skipping monster with algorithm 'none'")
		return b.Yell("I'm a chest ... I think")
	}
	if monster.Attributes.Life <= 0 {
		b.Logger.Warnw("Skipping DEAD monster")
		return nil
	}
	// calculate distance and line of sight
	b.BotState.MapExtended = b.calculateDistanceAndLineOfSight(level, *position)
	b.BotState.Objects = b.getMapObjectsByCategoryForLevel(level)

	combatCmd := b.combat()
	if combatCmd != nil {
		return combatCmd
	}

	random := rand.Intn(7)
	switch random {
	case 0:
		heal := b.heal()
		if heal != nil {
			return heal
		}
		fallthrough
	case 1:
		return b.rest()
	case 2:
		fallthrough
	case 3:
		jump := b.jumpAway()
		if jump != nil {
			return jump
		}
	case 4:
		fallthrough
	case 5:
		fallthrough
	case 6:
		return b.randomWalk()
	default:
		return b.Yell("Nothing to do ...")
	}
	return b.Yell("Nothing to do ...")
}

func (b *Bot) combat() *swagger.DungeonsandtrollsCommandsBatch {
	random := rand.Intn(5)
	if random == 0 {
		heal := b.heal()
		if heal != nil {
			b.addFirstYell("HEAL!")
			return heal
		}
	}
	if len(b.BotState.Objects.Hostile) <= 0 {
		return nil
	}
	skills := getAllSkills(b.Details.Monster.EquippedItems)
	b.Logger.Infow("All skills",
		"skills", skills,
		"numSkills", len(skills),
	)
	dmgSkills := b.filterDamageSkills(skills)
	if len(skills) > 0 && len(dmgSkills) == 0 {
		b.addFirstYell("NO DMG SKILLS!")
		b.Logger.Errorw("NO DMG SKILLS!",
			"allSkills", skills,
			"monsterAttributes", b.Details.Monster.Attributes,
			"monster", b.Details.Monster,
		)
	}
	skills2 := b.filterRequirementsMetSkills(dmgSkills)
	if len(skills2) == 0 {
		b.addYell("No skills available")
		b.Logger.Errorw("No skills available")
	}
	if len(dmgSkills)-len(skills2) >= 2 {
		b.addYell("Combat rest")
		random := rand.Intn(3)
		if random == 0 {
			return b.rest()
		} else {
			return b.jumpAway()
		}
	}
	if len(dmgSkills) > 0 && len(skills2) == 0 {
		b.addYell("Out of stamina (?)")
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
	bestEnemy := MapObject{}
	for j := range enemies {
		enemy := enemies[j]
		for i := range skills2 {
			skill := skills2[i]
			skillResult := b.evaluateSkill(skill, enemy)
			b.Logger.Infow("Skill evaluated",
				"skillName", skill.Name,
				"skill", skill,
				zap.Any("skillResult", skillResult),
				"position", enemy.MapObjects.Position,
				"myPosition", b.Details.Position,
			)
			if skillResult != nil && skillResult.Damage > bestDmg {
				b.Logger.Infow("Found better skill",
					"skillName", skill.Name,
					"skill", skill,
					"damage", skillResult.Damage,
					"position", enemy.MapObjects.Position,
					"myPosition", b.Details.Position,
				)
				bestSkill = &skill
				bestDmg = skillResult.Damage
				bestEnemy = enemy
			}
		}
	}
	if bestSkill != nil {
		b.Logger.Infow("Using best skill available",
			"skillName", bestSkill.Name,
			"skill", bestSkill,
			"damage", bestDmg,
			"targetId", bestEnemy.GetId(),
			"targetName", bestEnemy.GetName(),
			"targetFaction", bestEnemy.GetFaction(),
			"position", bestEnemy.MapObjects.Position,
			"myPosition", b.Details.Position,
			"myFaction", b.Details.Monster.Faction,
		)
		return b.useSkill(*bestSkill, bestEnemy)
	}
	b.Logger.Warnw("No skill chosen")

	random = rand.Intn(3)
	switch random {
	case 0:
		return b.rest()
	case 1:
		return b.heal()
	default:
		return b.moveTowardsEnemy(enemies)
	}
}

func (b *Bot) moveTowardsEnemy(enemies []MapObject) *swagger.DungeonsandtrollsCommandsBatch {
	// Go to player
	magicDistance := 17 // distance threshold
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
	b.addYell("I'm coming for you " + closeEnemies[rp].GetName() + "!")
	b.Logger.Infow("I'm coming for you!",
		"targetName", closeEnemies[rp].GetName(),
	)
	return &swagger.DungeonsandtrollsCommandsBatch{
		Move: closeEnemies[rp].MapObjects.Position,
	}
}

func (b *Bot) heal() *swagger.DungeonsandtrollsCommandsBatch {
	b.Logger.Debugw("Picking a friend-support skill ...")
	// Rest & Heal, etc.
	allSkills := getAllSkills(b.Details.Monster.EquippedItems)
	skills := b.filterRequirementsMetSkills(allSkills)
	if len(skills) <= 0 {
		return nil
	}
	targets := b.BotState.Objects.Friendly
	if len(targets) <= 0 {
		b.Logger.Warnw("No friends found :cry: (to heal)")
		return nil
	}
	b.Logger.Infow("Choosing healing skills & friends",
		"skills", skills,
		"numSkills", len(skills),
	)
	var bestSkill *swagger.DungeonsandtrollsSkill
	bestScore := float32(0)
	bestTarget := MapObject{}
	for j := range targets {
		target := targets[j]
		for i := range skills {
			skill := skills[i]
			skillVitalsScore := b.evaluateHealSkill(&skill, &target)
			b.Logger.Infow("Skill + target evaluated",
				"skillName", skill.Name,
				"skill", skill,
				"vitalsScore", skillVitalsScore,
				"position", target.MapObjects.Position,
				"myPosition", b.Details.Position,
				"targetName", target.GetName(),
				"targetFaction", target.GetFaction(),
			)
			if skillVitalsScore > bestScore {
				b.Logger.Infow("Found better skill + target",
					"skillName", skill.Name,
					"skill", skill,
					"vitalsScore", skillVitalsScore,
					"position", target.MapObjects.Position,
					"myPosition", b.Details.Position,
					"targetName", target.GetName(),
					"targetFaction", target.GetFaction(),
				)
				bestSkill = &skill
				bestScore = skillVitalsScore
				bestTarget = target
			}
		}
	}
	if bestSkill == nil {
		b.Logger.Warnw("No heal/support skill chosen")
		return nil
	}
	b.Logger.Infow("Using best heal skill available",
		"skillName", bestSkill.Name,
		"skill", bestSkill,
		"vitalsScore", bestScore,
		"targetId", bestTarget.GetId(),
		"targetName", bestTarget.GetName(),
		"targetFaction", bestTarget.GetFaction(),
		"position", bestTarget.MapObjects.Position,
		"myPosition", b.Details.Position,
		"myFaction", b.Details.Monster.Faction,
	)
	return b.useSkill(*bestSkill, bestTarget)
}

func (b *Bot) rest() *swagger.DungeonsandtrollsCommandsBatch {
	b.Logger.Debugw("Picking a self-support skill ...")
	// Rest & Heal, etc.
	allSkills := getAllSkills(b.Details.Monster.EquippedItems)
	skills := b.filterRequirementsMetSkills(allSkills)
	if len(skills) <= 0 {
		return nil
	}
	var bestSkill *swagger.DungeonsandtrollsSkill
	bestSkill = nil
	bestScore := float32(0)
	for i := range skills {
		skill := skills[i]
		casterEffects := skill.CasterEffects
		if casterEffects == nil {
			continue
		}
		skillAttributes := casterEffects.Attributes
		if skillAttributes == nil {
			continue
		}
		// TODO: also check target attributes
		score := b.scoreVitals(skillAttributes, &skill)
		if score > bestScore {
			bestScore = score
			bestSkill = &skill
		}
	}
	if bestSkill == nil {
		b.Logger.Warnw("No support skill chosen")
		return nil
	}
	b.Logger.Infow("Best support skill",
		"skillName", bestSkill.Name,
		"skill", bestSkill,
		"vitalsScore", bestScore,
		"skillTargetType", bestSkill.Target,
	)
	if *bestSkill.Target != swagger.NONE_SkillTarget {
		b.Logger.Warnw("Picking self as target for support skill - might not be a good idea")
	}
	return b.useSkill(*bestSkill, b.BotState.Self)
}
