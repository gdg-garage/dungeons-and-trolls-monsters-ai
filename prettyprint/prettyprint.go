package prettyprint

import (
	"fmt"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
)

func Command(cmd *swagger.DungeonsandtrollsCommandsBatch) {
	fmt.Printf("Command full: %+v\n", cmd)
	// &{Buy:<nil> PickUp:<nil> Move:0x1400021c324 Skill:<nil> Yell:<nil>}
	if cmd.Buy != nil {
		fmt.Printf("  -> Buy: %+v\n", cmd.Buy)
	}
	if cmd.PickUp != nil {
		fmt.Printf("  -> PickUp: %+v\n", cmd.PickUp)
	}
	if cmd.Move != nil {
		fmt.Printf("  -> Move: %+v\n", cmd.Move)
	}
	if cmd.Skill != nil {
		fmt.Printf("  -> Skill: %+v\n", cmd.Skill)
	}
	if cmd.Yell != nil {
		fmt.Printf("  -> Yell: %+v\n", cmd.Yell)
	}
}
