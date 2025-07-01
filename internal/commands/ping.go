package commands

import (
	"time"

	"github.com/arnokay/arnobot-core/internal/commands/cmdtypes"
)

type pingCommand struct{}

func NewPingCommand() pingCommand {
	return pingCommand{}
}

func (c pingCommand) Name() string {
	return "ping"
}

func (c pingCommand) Aliases() []string {
	return nil
}

func (c pingCommand) Description() string {
	return "im gonna pong"
}

func (c pingCommand) Cooldown() time.Duration {
	return time.Second * 5
}

func (c pingCommand) Execute(ctx cmdtypes.CommandContext) (cmdtypes.CommandResponse, error) {
	response := cmdtypes.CommandResponse{
		Message: "pong",
		ReplyTo: ctx.Message.ID,
	}

	return response, nil
}
