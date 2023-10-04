package bot

import (
	"math"
	"math/rand"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

func (b *Bot) randomWalk() *swagger.DungeonsandtrollsCommandsBatch {
	myPosition := *b.Details.Position
	if b.PrevBotState.State == "move" && b.PrevBotState.TargetPosition != myPosition {
		// Continue moving to target
		b.Logger.Infow("Continue moving to target",
			"targetPosition", b.PrevBotState.TargetPosition,
		)
		return b.Move(b.PrevBotState.TargetPosition)
	}

	for i := 0; i < 16; i++ {
		// get random direction
		distanceX := rand.Intn(7) - 3
		distanceY := rand.Intn(7) - 3
		newX := int(myPosition.PositionX) + distanceX
		newY := int(myPosition.PositionY) + distanceY

		position := makePosition(int32(newX), int32(newY))
		tileInfo, found := b.BotState.MapExtended[position]
		if !found || !tileInfo.mapObjects.IsFree || tileInfo.distance == math.MaxInt32 {
			// unreachable or not free
			continue
		}
		if len(tileInfo.mapObjects.Monsters) > 0 && i < 5 {
			// Prefer not to walk into other monsters
			continue
		}
		return b.Move(position)
	}
	b.Logger.Warnw("randomWalkFromPosition: No free position found")
	return b.Yell("I'm stuck ...")
}

func (b *Bot) jumpAway() *swagger.DungeonsandtrollsCommandsBatch {
	allSkills := getAllSkills(b.Details.Monster.EquippedItems)
	skills := b.filterMovementSkills(allSkills)
	b.Logger.Infow("Jumping away ...",
		"moveSkills", skills,
	)
	if len(skills) == 0 {
		b.Logger.Infow("No movement skills (to jump away)")
		return nil
	}
	reqSkills := b.filterRequirementsMetSkills(skills)
	if len(reqSkills) == 0 {
		b.Logger.Infow("No movement skills that I have resources for (to jump away)")
		random := rand.Intn(2)
		if random == 0 {
			return b.rest()
		}
		return b.randomWalk()
	}

	skillsByRange := map[int][]swagger.DungeonsandtrollsSkill{}
	for _, skill := range reqSkills {
		range_ := b.calculateAttributesValue(*skill.Range_)
		b.Logger.Debugw("Adding skill to range map",
			"skillName", skill.Name,
			"rangeAttrs", skill.Range_,
			"range", range_,
		)
		skillsByRange[range_] = append(skillsByRange[range_], skill)
	}

	pos := b.Details.Position

	for i := 0; i < 20; i++ {
		distanceX := rand.Intn(7) - 3
		distanceY := rand.Intn(7) - 3
		newX := pos.PositionX + int32(distanceX)
		newY := pos.PositionY + int32(distanceY)

		position := makePosition(newX, newY)
		tileInfo, found := b.BotState.MapExtended[position]
		if !found || !tileInfo.mapObjects.IsFree || tileInfo.distance == math.MaxInt32 {
			// unreachable or not free
			continue
		}
		if len(tileInfo.mapObjects.Monsters) > 0 && i < 14 {
			// Prefer not to walk into other monsters
			continue
		}
		b.Logger.Debugw("Testing position for jump",
			"position", position,
			"myPosition", pos,
			"distance", tileInfo.distance,
			"try", i,
		)
		skillsWithRange, found := skillsByRange[tileInfo.distance]
		if found && len(skillsWithRange) > 0 {
			random := rand.Intn(len(skillsWithRange))
			if skillsWithRange[random].Flags.RequiresLineOfSight && !tileInfo.lineOfSight {
				b.Logger.Infow("No line of sight for jump")
				continue
			}
			return b.MoveSkill(&skillsWithRange[random], &position)
		}
		if i < 10 {
			// Prefer full range jumps
			continue
		}
		skillsWithRange, found = skillsByRange[tileInfo.distance+1]
		if found && len(skillsWithRange) > 0 {
			random := rand.Intn(len(skillsWithRange))
			if skillsWithRange[random].Flags.RequiresLineOfSight && !tileInfo.lineOfSight {
				b.Logger.Infow("No line of sight for jump")
				continue
			}
			return b.MoveSkill(&skillsWithRange[random], &position)
		}
	}
	b.Logger.Errorw("No positions found for jump -> fallback to random walk")
	b.addYell("Can't jump anywhere ...")
	return b.randomWalk()
}

func (b *Bot) MoveSkillXY(skill *swagger.DungeonsandtrollsSkill, x, y int32) *swagger.DungeonsandtrollsCommandsBatch {
	pos := makePosition(x, y)
	return b.MoveSkill(skill, &pos)
}

func (b *Bot) MoveSkill(skill *swagger.DungeonsandtrollsSkill, position *swagger.DungeonsandtrollsPosition) *swagger.DungeonsandtrollsCommandsBatch {
	b.Logger.Infow("Jumping! (move skill)",
		"skill", skill,
		"position", position,
		"myPosition", b.Details.Position,
		"range", b.calculateAttributesValue(*skill.Range_),
	)
	// XXX: This is super dumb
	mapObjects := b.BotState.MapExtended[*position].mapObjects
	if len(mapObjects.Players) > 0 {
		b.addFirstYell("HOP :)")
		return b.useSkill(*skill, NewPlayerMapObject(mapObjects, 0))
	} else if len(mapObjects.Monsters) > 0 {
		b.addFirstYell("HOP :)")
		return b.useSkill(*skill, NewMonsterMapObject(mapObjects, 0))
	} else {
		if *skill.Target != swagger.POSITION_SkillTarget {
			b.addFirstYell("NO HOP :(")
			b.Logger.Errorw("Aborting jump :(")
			return nil
		}
		b.addFirstYell("HOP :)")
		return &swagger.DungeonsandtrollsCommandsBatch{
			Skill: &swagger.DungeonsandtrollsSkillUse{
				SkillId:  skill.Id,
				Position: position,
			},
		}
	}
}
