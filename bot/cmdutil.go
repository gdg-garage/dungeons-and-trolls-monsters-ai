package bot

import swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"

func (b *Bot) Yell(msg string) *swagger.DungeonsandtrollsCommandsBatch {
	return &swagger.DungeonsandtrollsCommandsBatch{
		Yell: &swagger.DungeonsandtrollsMessage{
			Text: msg,
		},
	}
}

func (b *Bot) MoveXY(x, y int32) *swagger.DungeonsandtrollsCommandsBatch {
	return &swagger.DungeonsandtrollsCommandsBatch{
		Move: &swagger.DungeonsandtrollsPosition{
			PositionX: x,
			PositionY: y,
		},
	}
}

func (b *Bot) Move(position swagger.DungeonsandtrollsPosition) *swagger.DungeonsandtrollsCommandsBatch {
	return &swagger.DungeonsandtrollsCommandsBatch{
		Move: &position,
	}
}
