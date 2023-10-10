package bot

import (
	"strings"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

func (b *Bot) addPublicYell(msg string) {
	b.BotState.Yells = append(b.BotState.Yells, msg)
}

func (b *Bot) addYell(msg string) {
	if b.Environment == "dev" {
		b.BotState.Yells = append(b.BotState.Yells, msg)
	}
}

func (b *Bot) addFirstYell(msg string) {
	if b.Environment == "dev" {
		b.BotState.Yells = append([]string{msg}, b.BotState.Yells...)
	}
}

func (b *Bot) constructYellCommand(cmd *swagger.DungeonsandtrollsCommandsBatch) *swagger.DungeonsandtrollsCommandsBatch {
	// Add yell from command
	if cmd != nil && cmd.Yell != nil && cmd.Yell.Text != "" {
		b.addFirstYell(cmd.Yell.Text)
	}
	// Join all yells
	text := strings.Join(b.BotState.Yells, "; ")
	// Do nothing if no text
	if text == "" {
		return cmd
	}
	if cmd == nil {
		cmd = &swagger.DungeonsandtrollsCommandsBatch{}
	}
	cmd.Yell = &swagger.DungeonsandtrollsMessage{
		Text: text,
	}
	return cmd
}
