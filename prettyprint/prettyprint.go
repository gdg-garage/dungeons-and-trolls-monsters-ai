package prettyprint

import (
	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
	"go.uber.org/zap"
)

func extractCommandType(cmd *swagger.DungeonsandtrollsCommandsBatch) string {
	if cmd.Buy != nil {
		return "Buy"
	}
	if cmd.PickUp != nil {
		return "PickUp"
	}
	if cmd.Move != nil {
		return "Move"
	}
	if cmd.Skill != nil {
		return "Skill"
	}
	if cmd.Yell != nil {
		return "Yell"
	}
	return ""
}

func extractCommand(cmd *swagger.DungeonsandtrollsCommandsBatch) interface{} {
	if cmd.Buy != nil {
		return cmd.Buy
	}
	if cmd.PickUp != nil {
		return cmd.PickUp
	}
	if cmd.Move != nil {
		return cmd.Move
	}
	if cmd.Skill != nil {
		return cmd.Skill
	}
	if cmd.Yell != nil {
		return cmd.Yell
	}
	return nil
}

func Command(logger *zap.SugaredLogger, cmd *swagger.DungeonsandtrollsCommandsBatch) {
	logger.Infow("Command",
		zap.String("commandType", extractCommandType(cmd)),
		zap.Any("command", extractCommand(cmd)),
	)
}
