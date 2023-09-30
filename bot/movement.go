package bot

import (
	"math"
	"math/rand"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

func (b *Bot) jumpAway() *swagger.DungeonsandtrollsCommandsBatch {
	allSkills := getAllSkills(b.Details.Monster.EquippedItems)
	skills := b.filterMovementSkills(allSkills)
	if len(skills) == 0 {
		b.Logger.Infow("No movement skills (to jump away)")
		return nil
	}
	reqSkills := b.filterRequirementsMetSkills(skills)
	if len(reqSkills) == 0 {
		return b.rest()
		// TODO: implement moveAway ?
	}

	skillsByRange := map[int][]swagger.DungeonsandtrollsSkill{}
	for _, skill := range reqSkills {
		range_ := b.calculateAttributesValue(*skill.Range_)
		skillsByRange[range_] = append(skillsByRange[range_], skill)
	}

	pos := b.Details.Position

	for i := 0; i < 20; i++ {
		distanceX := rand.Intn(8) - 4
		distanceY := rand.Intn(8) - 4
		newX := pos.PositionX + int32(distanceX)
		newY := pos.PositionY + int32(distanceY)

		tileInfo, found := b.BotState.MapExtended[makePosition(newX, newY)]
		if !found || !tileInfo.mapObjects.IsFree || tileInfo.distance == math.MaxInt32 {
			// unreachable or not free
			continue
		}
		if len(tileInfo.mapObjects.Monsters) > 0 && i < 14 {
			// Prefer not to walk into other monsters
			continue
		}
		skillsWithRange, found := skillsByRange[tileInfo.distance]
		if found && len(skillsWithRange) > 0 {
			random := rand.Intn(len(skillsWithRange))
			b.MoveSkillXY(&skillsWithRange[random], newX, newY)
		}
		if i < 10 {
			// Prefer full range jumps
			continue
		}
		skillsWithRange, found = skillsByRange[tileInfo.distance-1]
		if found && len(skillsWithRange) > 0 {
			random := rand.Intn(len(skillsWithRange))
			b.MoveSkillXY(&skillsWithRange[random], newX, newY)
		}
	}
	return nil
}

func (b *Bot) MoveSkillXY(skill *swagger.DungeonsandtrollsSkill, x, y int32) *swagger.DungeonsandtrollsCommandsBatch {
	pos := makePosition(x, y)
	return b.MoveSkill(skill, &pos)
}

func (b *Bot) MoveSkill(skill *swagger.DungeonsandtrollsSkill, position *swagger.DungeonsandtrollsPosition) *swagger.DungeonsandtrollsCommandsBatch {
	return &swagger.DungeonsandtrollsCommandsBatch{
		Skill: &swagger.DungeonsandtrollsSkillUse{
			SkillId:  skill.Id,
			Position: position,
		},
		Yell: &swagger.DungeonsandtrollsMessage{
			Text: "HOP :)",
		},
	}
}
