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

	Config Config

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

	if monster.Algorithm == "none" {
		b.Logger.Warnw("Skipping monster with algorithm 'none'")
		return b.Yell("I'm a chest ... I think")
	}
	if monster.Attributes.Life <= 0 {
		b.Logger.Warnw("Skipping DEAD monster")
		return nil
	}
	if monster.Stun.IsStunned {
		b.Logger.Warnw("Skipping stunned monster")
		return b.Yell("STUNNED!")
	}
	b.Logger.Infow("Handling monster",
		"monster", monster,
		"position", position,
	)
	// calculate distance and line of sight
	b.BotState.MapExtended = b.calculateDistanceAndLineOfSight(level, *position)
	b.BotState.Objects = b.getMapObjectsByCategoryForLevel(level)

	// One shot skill eval
	return b.bestSkill()
}

func (b *Bot) moveTowardsEnemy(enemies []MapObject) *swagger.DungeonsandtrollsCommandsBatch {
	// Go to player
	magicDistance := 15 // distance threshold
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
