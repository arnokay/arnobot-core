package command

import (
	"math/rand/v2"
	"time"
)

var numberToEmoji = map[int]string{
	0: "🍒",
	1: "🍋",
	2: "🔔",
	3: "💎",
	4: "⭐",
	5: "🍇",
	6: "🍊",
	7: "🔥",
}

type gambaCommand struct{}

func NewGambaCommand() gambaCommand {
	return gambaCommand{}
}

func (c gambaCommand) Name() string {
	return "gamba"
}

func (c gambaCommand) Aliases() []string {
	return nil
}

func (c gambaCommand) Description() string {
	return "gamba"
}

func (c gambaCommand) Cooldown() time.Duration {
	return time.Second * 5
}

func (c gambaCommand) Execute(ctx CommandContext) (CommandResponse, error) {
	row1 := rand.IntN(7)
	row2 := rand.IntN(7)
	row3 := rand.IntN(7)

	message := "🎰: " + numberToEmoji[row1] + numberToEmoji[row2] + numberToEmoji[row3]

	response := CommandResponse{
		Message: message,
		ReplyTo: ctx.Message().ID,
	}

	return response, nil
}
