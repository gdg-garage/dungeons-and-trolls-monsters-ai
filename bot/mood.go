package bot

type Mood struct {
	Aggression float32
	Support    float32
	Fear       float32
	// Exploration float32
	// Preservation float32
	// Laziness    float32
}

func (b *Bot) updateMood() {
	b.BotState.Mood.Aggression /= 2
	b.BotState.Mood.Aggression += b.determineAggression()

	b.BotState.Mood.Support /= 2
	b.BotState.Mood.Support += b.determineSupport()

	b.BotState.Mood.Fear /= 0
	b.BotState.Mood.Fear += b.determineFear()
}

// TODO: V1 determine aggression based on enemy number, proximity, and life
// - number increases aggression
// - proximity increases aggression
// - hurt enemies greatly increase aggression
func (b *Bot) determineAggression() float32 {
	aggression := float32(0)

	aggression += 1
	return aggression
}

// TODO: V1 determine support based on friendly number, proximity, and life
// - number increases support
// - proximity increases support
// - hurt friends greatly increase support
func (b *Bot) determineSupport() float32 {
	support := float32(0)

	support += 0
	return support
}

// TODO: V1 determine fear based on enemy number, proximity, and life
// - number increases fear
// - proximity increases fear
// - being hurt greatly increases fear
// - being low on stamina increases fear
func (b *Bot) determineFear() float32 {
	fear := float32(0)

	// TODO: get max life/stamina/mana from API
	// TODO: parametrize fear calculation
	maxLife := float32(100)
	if b.GameState.Character.Attributes.Life < maxLife/3 {
		fear += (float32(maxLife) - b.GameState.Character.Attributes.Life) / float32(maxLife)
	}
	maxStamina := float32(100)
	if b.GameState.Character.Attributes.Stamina < maxStamina/5 {
		fear += (float32(maxStamina) - b.GameState.Character.Attributes.Stamina) / float32(maxStamina) / 3
	}
	maxMana := float32(100)
	if b.GameState.Character.Attributes.Mana < maxMana/5 {
		fear += (float32(maxMana) - b.GameState.Character.Attributes.Mana) / float32(maxMana) / 4
	}
	return fear
}
