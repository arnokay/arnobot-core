package command

import (
	"math/rand/v2"
	"time"
)

type coinCommand struct{}

func NewCoinCommand() coinCommand {
	return coinCommand{}
}

func (c coinCommand) Name() string {
	return "coin"
}

func (c coinCommand) Aliases() []string {
	return nil
}

func (c coinCommand) Description() string {
	return "get heads or tails by throwing this coin"
}

func (c coinCommand) Cooldown() time.Duration {
	return time.Second * 5
}

func (c coinCommand) Execute(ctx CommandContext) (CommandResponse, error) {
	random := rand.IntN(6000)

	side := "edge"

	if random < 2999 {
		side = "heads"
	}

	if random > 2999 {
		side = "tails"
	}

	response := CommandResponse{
		Message: "🪙: " + side,
		ReplyTo: ctx.Message().ID,
	}

	return response, nil
}
