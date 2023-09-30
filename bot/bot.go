package bot

import (
	"math"
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

func (b *Bot) Run() *swagger.DungeonsandtrollsCommandsBatch {
	b.BotState.Self = NewMonsterMapObject(*b.Details.MapObjects, b.Details.Index)
	b.BotState.Yell = ""
	b.BotState.PrefixYell = ""
	monster := b.Details.Monster
	// monsterTileObjects := b.Details.MapObjects
	level := b.Details.Level
	position := b.Details.Position

	b.Logger.Infow("Handling monster",
		"monster", monster,
		"position", position,
	)
	if monster.Faction == "neutral" {
		b.Logger.Warnw("Skipping neutral monster")
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

	random := rand.Intn(5)
	switch random {
	case 0:
		return b.rest()
	case 1:
		fallthrough
	case 2:
		return b.randomWalk()
	case 3:
		fallthrough
	case 4:
		return b.jumpAway()
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
	dmgSkills := b.filterDamageSkills(skills)
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
		random := rand.Intn(2)
		if random == 0 {
			return b.rest()
		} else {
			return b.jumpAway()
		}
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

	random := rand.Intn(3)
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
	b.Logger.Infow("I'm coming for you!",
		"targetName", closeEnemies[rp].GetName(),
	)
	return &swagger.DungeonsandtrollsCommandsBatch{
		Move: closeEnemies[rp].MapObjects.Position,
	}
}

func (b *Bot) rest() *swagger.DungeonsandtrollsCommandsBatch {
	b.Logger.Debugw("Picking a support skill ...")
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

// Get vitals score for skill
// Tells you how much the skill will improve your resources (life, stamina, mana)
// Can be used for both casterEffect and targetEffect skills
func (b *Bot) scoreVitals(skillAttributes *swagger.DungeonsandtrollsSkillAttributes, skill *swagger.DungeonsandtrollsSkill) float32 {
	skillAttributes = fillSkillAttributes(*skillAttributes)
	skillLifeGain := float32(b.calculateAttributesValue(*skillAttributes.Life)) - skill.Cost.Life
	skillStaminaGain := float32(b.calculateAttributesValue(*skillAttributes.Stamina)) - skill.Cost.Stamina
	skillManaGain := float32(b.calculateAttributesValue(*skillAttributes.Mana)) - skill.Cost.Mana

	lifePercentage := b.Details.Monster.Attributes.Life / b.Details.Monster.MaxAttributes.Life
	staminaPercentage := b.Details.Monster.Attributes.Stamina / b.Details.Monster.MaxAttributes.Stamina
	if math.IsNaN(float64(staminaPercentage)) {
		staminaPercentage = 0
	}
	manaPercentage := b.Details.Monster.Attributes.Mana / b.Details.Monster.MaxAttributes.Mana
	if math.IsNaN(float64(manaPercentage)) {
		manaPercentage = 0
	}
	score := b.scoreVitalsFunc(lifePercentage, staminaPercentage, manaPercentage)

	lifePercentageAfter := (b.Details.Monster.Attributes.Life + skillLifeGain) / b.Details.Monster.MaxAttributes.Life
	staminaPercentageAfter := (b.Details.Monster.Attributes.Stamina + skillStaminaGain) / b.Details.Monster.MaxAttributes.Stamina
	if math.IsNaN(float64(staminaPercentageAfter)) {
		staminaPercentageAfter = 0
	}
	manaPercentageAfter := (b.Details.Monster.Attributes.Mana + skillManaGain) / b.Details.Monster.MaxAttributes.Mana
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
		"life", b.Details.Monster.Attributes.Life,
		"lifeMax", b.Details.Monster.MaxAttributes.Life,
		"staminaPercentage", staminaPercentage,
		"stamina", b.Details.Monster.Attributes.Stamina,
		"staminaMax", b.Details.Monster.MaxAttributes.Stamina,
		"manaPercentage", manaPercentage,
		"mana", b.Details.Monster.Attributes.Mana,
		"manaMax", b.Details.Monster.MaxAttributes.Mana,
		"lifePercentageAfter", lifePercentageAfter,
		"staminaPercentageAfter", staminaPercentageAfter,
		"manaPercentageAfter", manaPercentageAfter,
		"vitalsScore", score,
		"vitalsScoreAfter", scoreAfter,
		"vitalsScoreDiff", scoreDiff,
	)
	return scoreDiff
}

func (b *Bot) scoreVitalsFunc(lifePercentage, staminaPercentage, manaPercentage float32) float32 {
	f := func(x float32) float32 {
		if x > 1 {
			// cap score at 100%
			x = 1
		}
		// adding 2 just to make the score usually positive (50% resource == 0 score)
		return 2 - (float32(1) / x)
	}
	return 4*f(lifePercentage) + 2*f(staminaPercentage) + f(manaPercentage)
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
