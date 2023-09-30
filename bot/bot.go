package bot

import (
	"math/rand"
	"time"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
	"go.uber.org/zap"
)

type BotState struct {
	Objects     MapObjectsByCategory
	MapExtended map[swagger.DungeonsandtrollsPosition]MapCellExt
	Self        MapObject

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
		return nil
	}
	if monster.Attributes.Life <= 0 {
		b.Logger.Warnw("Skipping DEAD monster ☠")
		return nil
	}
	attrs := monster.Attributes
	// calculate distance and line of sight
	b.BotState.MapExtended = b.calculateDistanceAndLineOfSight(level, *position)
	b.BotState.Objects = b.getMapObjectsByCategoryForLevel(level)

	if len(b.BotState.Objects.Hostile) > 0 {
		skills := getAllSkills(monster.EquippedItems)
		b.Logger.Infow("All skills",
			"skills", skills,
			"numSkills", len(skills),
		)
		dmgSkills := b.filterDamageSkills2(*attrs, skills)
		b.Logger.Infow("Dmg skills",
			"skills", dmgSkills,
			"numSkills", len(dmgSkills),
		)
		skills2 := b.filterRequirementsMetSkills2(*attrs, dmgSkills)
		b.Logger.Infow("Requirements met skills",
			"skills", skills2,
			"numSkills", len(skills2),
		)
		if len(skills2) == 0 {
			b.Logger.Errorw("No skills available")
			return nil
		}
		enemies := b.BotState.Objects.Hostile
		for _, enemy := range enemies {
			enemyDistance := b.BotState.MapExtended[*enemy.MapObjects.Position].distance
			for _, skill := range skills2 {
				if enemyDistance <= calculateAttributesValue(*attrs, *skill.Range_) {
					b.Logger.Warnw("Attacking enemy",
						"enemyName", enemy.GetName(),
						"enemyID", enemy.GetId(),
						"skillName", skill.Name,
						"skill", skill,
					)
					return useSkill(skill, enemy)
				}
			}
		}
		// Go to player
		magicDistance := 7 // distance threshold
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
		return &swagger.DungeonsandtrollsCommandsBatch{
			Move: closeEnemies[rp].MapObjects.Position,
		}
	}

	random := rand.Intn(4)
	if random < 2 {
		b.Logger.Infow("Picking a support skill ...")
		// Rest & Heal, etc.
		skills := getAllSkills(b.Details.Monster.EquippedItems)
		skills2 := b.filterRequirementsMetSkills2(*attrs, skills)
		if len(skills2) > 0 {
			supportSkills := []swagger.DungeonsandtrollsSkill{}
			for _, skill := range skills2 {
				b.Logger.Infow("Checking skill",
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
				b.Logger.Infow("Skill attributes",
					"skillName", skill.Name,
					"skill", skill,
					"casterEffects", casterEffects,
					"skillAttributes", skillAttributes,
					"skillLife", skillLife,
					"skillMana", skillMana,
					"skillStamina", skillStamina,
				)
				if (skillLife != nil && calculateAttributesValue(*attrs, *skillLife) > 0) ||
					(skillMana != nil && calculateAttributesValue(*attrs, *skillMana) > 0) ||
					(skillStamina != nil && calculateAttributesValue(*attrs, *skillStamina) > 0) {
					supportSkills = append(supportSkills, skill)
				}
			}
			b.Logger.Infow("Support skills",
				"skills", supportSkills,
			)
			if len(supportSkills) > 0 {
				rp := rand.Intn(len(supportSkills))
				b.Logger.Infow("Using support skill",
					"skillName", supportSkills[rp].Name,
					"skill", supportSkills[rp],
				)
				// TODO: Don't attack yourself 🙈
				return useSkill(supportSkills[rp], b.BotState.Self)
			}

			// Idle
			if random < 3 {
				return &swagger.DungeonsandtrollsCommandsBatch{
					Yell: &swagger.DungeonsandtrollsMessage{
						Text: "I'm a monster!",
					},
				}
			}
			return b.randomWalkFromPositionExt(level, *b.Details.Position, b.BotState.MapExtended)
		}
	}
	return nil
}

func (b *Bot) Run3() *swagger.DungeonsandtrollsCommandsBatch {
	b.updateMood()

	state := *b.GameState
	score := state.Score
	b.Logger.Infow("Game state: Character info",
		"Name", state.Character.Name,
		"Score", score,
		"Money", state.Character.Money,
		"CurrentLevel", state.CurrentLevel,
		"CurrentPosition", state.CurrentPosition,
	)

	var mainHandItem *swagger.DungeonsandtrollsItem
	for _, item := range state.Character.Equip {
		if *item.Slot == swagger.MAIN_HAND_DungeonsandtrollsItemType {
			mainHandItem = &item
			break
		}
	}

	b.calculateDistanceAndLineOfSight(state.CurrentLevel, *state.CurrentPosition)

	if mainHandItem == nil {
		b.Logger.Debug("Looking for items to buy ...")
		item := b.shop()
		if item != nil {
			return &swagger.DungeonsandtrollsCommandsBatch{
				Buy: &swagger.DungeonsandtrollsIdentifiers{Ids: []string{item.Id}},
			}
		}
		b.Logger.Warn("ERROR: Found no item to buy!")
	}

	objects := b.getMapObjectsByCategory()
	b.Logger.Debugw("Stairs position ...",
		"stairsPosition", objects.Stairs.Position,
	)

	// Add seed
	rand.Seed(time.Now().UnixNano())
	random := rand.Intn(8)
	if random <= 1 {
		b.Logger.Debug("Picking a random yell ...")
		randomYell := rand.Intn(8)
		var yells []string = []string{
			"Anybody home?",
			"What was that?",
			"Hey! Show yourself!",
			"Yo!",
			"500 error",
			"Your mom is a nice lady",
			"You will never find me",
			"Come and get me",
			"8 ball says: no",
		}
		yell := yells[randomYell]
		b.Logger.Debugw("Yelling ...",
			"yell", yell,
		)
		return &swagger.DungeonsandtrollsCommandsBatch{
			Yell: &swagger.DungeonsandtrollsMessage{
				Text: yell,
			},
		}
	}

	if len(objects.Monsters) > 0 {
		// for _, monster := range objects.Monsters {
		// 	log.Printf("Monster: %+v\n", monster)
		// }
		b.Logger.Debug("Let's fight!")
		b.Logger.Debug("Picking a target ...")
		target := b.pickTarget(&objects)
		if target != nil {
			// log.Printf("Target: %+v\n", target)
			b.Logger.Debugw("Picked target",
				"targetPosition", target.MapObjects.Position,
			)

			if target.MapObjects.Position.PositionX == state.CurrentPosition.PositionX && target.MapObjects.Position.PositionY == state.CurrentPosition.PositionY {
				b.Logger.Debug("Picking a skill ...")
				skill := b.pickSkill()
				b.Logger.Debugw("Picked skill",
					"skillName", skill.Name,
					"skill", skill,
				)

				b.Logger.Debugw("Attacking target ...",
					"targetPosition", target.MapObjects.Position,
					"targetName", target.GetName(),
				)

				return useSkill(*skill, *target)
			} else {
				b.Logger.Debugw("Moving towards target ...",
					"targetPosition", target.MapObjects.Position,
					"targetName", target.GetName(),
				)
				return &swagger.DungeonsandtrollsCommandsBatch{
					Move: target.MapObjects.Position,
				}
			}
		}
	}

	if objects.Stairs == nil {
		b.Logger.Warn("Can't find stairs!")
		return &swagger.DungeonsandtrollsCommandsBatch{
			Yell: &swagger.DungeonsandtrollsMessage{
				Text: "Where are the stairs? I can't find them!",
			},
		}
	}

	b.Logger.Infow("Moving towards stairs ...",
		"stairsPosition", objects.Stairs.Position,
	)
	return &swagger.DungeonsandtrollsCommandsBatch{
		Move: objects.Stairs.Position,
	}
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
