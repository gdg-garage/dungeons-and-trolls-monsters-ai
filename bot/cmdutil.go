package bot

import swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"

func (b *Bot) Yell(msg string) *swagger.DungeonsandtrollsCommandsBatch {
	return &swagger.DungeonsandtrollsCommandsBatch{
		Yell: &swagger.DungeonsandtrollsMessage{
			Text: msg,
		},
	}
}
