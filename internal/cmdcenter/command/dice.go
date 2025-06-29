package command

import (
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/arnokay/arnobot-shared/apperror"
)

const (
	minSides     = 2
	defaultSides = 6
)

func randRange(min, max int) int {
	return rand.IntN(max-min+1) + min
}

type diceCommand struct{}

func NewDiceCommand() diceCommand {
	return diceCommand{}
}

func (c diceCommand) Name() string {
	return "dice"
}

func (c diceCommand) Aliases() []string {
	return nil
}

func (c diceCommand) Description() string {
	return "roll a dice with N sides (default N=6, min N=2, max N=100)"
}

func (c diceCommand) Cooldown() time.Duration {
	return time.Second * 5
}

func (c diceCommand) Execute(ctx CommandContext) (CommandResponse, error) {
	sides := defaultSides
	if ctx.Command().Args != "" {
    convSides, err := strconv.Atoi(ctx.Command().Args)
		if err != nil {
			return CommandResponse{}, apperror.New(apperror.CodeInvalidInput, "sides are not integer", err)
		}
    sides = convSides
	}

	if sides > 100 || sides < 2 {
		return CommandResponse{}, apperror.New(apperror.CodeInvalidInput, "sides are limited to 100", nil)
	}

	side := randRange(minSides, sides)

	response := CommandResponse{
		Message: "🎲: " + strconv.Itoa(side),
		ReplyTo: ctx.Message().ID,
	}

	return response, nil
}
