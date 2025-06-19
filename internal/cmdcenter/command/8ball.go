package command

import (
	"math/rand"
	"strings"

	"github.com/arnokay/arnobot-shared/apperror"
)

var answers []string = []string{
	"it is certain",
	"it is decidedly so",
	"without a doubt",
	"Yes definitely",
	"You may rely on it",

	"as I see it, yes",
	"most likely",
	"outlook good",
	"yes",
	"signs point to yes",

	"reply hazy, try again",
	"ask again later",
	"better not tell you now",
	"cannot predict now",
	"concentrate and ask again",

	"don't count on it",
	"my reply is no",
	"my sources say no",
	"outlook not so good",
	"very doubtful",
}

var answersLength = len(answers)

type EightBall struct {
	CommonCommand
}

func NewEightBall() *EightBall {
	return &EightBall{}
}

func (c *EightBall) Name() string {
	return "8ball"
}

func (c *EightBall) Execute(ctx CommandContext) (CommandResponse, error) {
	if strings.TrimSpace(ctx.Command().Args) == "" {
		return CommandResponse{}, apperror.ErrNoAction
	}

	answerIndex := rand.Intn(answersLength - 1)
	answer := answers[answerIndex]

	response := CommandResponse{
		Message: "🎱: " + answer,
		ReplyTo: ctx.Message().ID,
	}

	return response, nil
}
