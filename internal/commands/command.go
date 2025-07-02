package commands

import (
	"slices"
	"strings"
	"time"

	"github.com/arnokay/arnobot-shared/apperror"
	"github.com/arnokay/arnobot-shared/data"

	"github.com/arnokay/arnobot-core/internal/app/service"
	"github.com/arnokay/arnobot-core/internal/commands/cmdtypes"
)

const (
	createOp = "add"
	updateOp = "edit"
	deleteOp = "del"
)

type cmdCommand struct {
	userCommandService *service.UserCommandService
}

func NewCmdCommand(
	userCommandService *service.UserCommandService,
) cmdCommand {
	return cmdCommand{
		userCommandService: userCommandService,
	}
}

func (c cmdCommand) Name() string {
	return "cmd"
}

func (c cmdCommand) Aliases() []string {
	return []string{"cmd" + createOp, "cmd" + deleteOp, "cmd" + updateOp}
}

func (c cmdCommand) Description() string {
	return "example: !cmd (add|edit|del) command_name text of command (only for add or edit)"
}

func (c cmdCommand) OpDescription(op string) string {
	switch op {
	case createOp, updateOp:
		return op + " example: !cmd " + op + " !customcommand Response to custom command! PogChamp"
	case deleteOp:
		return op + " example: !cmd " + op + " !customcommand"
	default:
		return c.Description()
	}
}

func (c cmdCommand) Cooldown() time.Duration {
	return time.Second * 5
}

func (c cmdCommand) Execute(ctx cmdtypes.CommandContext) (cmdtypes.CommandResponse, error) {
	var response cmdtypes.CommandResponse

	if ctx.Chatter.Role < data.ChatterModerator {
		return cmdtypes.CommandResponse{}, apperror.ErrNoAction
	}

	response.ReplyTo = ctx.Message.ID

	var operation string
	var rest string

	if slices.Contains(c.Aliases(), ctx.Command.Command) {
		operation = strings.TrimPrefix(ctx.Command.Command, "cmd")
		rest = ctx.Command.Args
	} else {
		operation, rest, _ = strings.Cut(ctx.Command.Args, " ")
	}

	switch operation {
	case createOp:
		name, text, _ := strings.Cut(rest, " ")
		name = strings.TrimSpace(name)
		text = strings.TrimSpace(text)
		if text == "" || name == "" {
			response.Message = c.Description()
			break
		}
		_, err := c.userCommandService.Create(ctx.Context, data.UserCommandCreate{
			UserID: ctx.Channel.UserID,
			Name:   name,
			Text:   text,
			Reply:  false,
		})
		if err != nil {
			response.Message = "couldnt create command, got error: " + err.Error()
			break
		}
		response.Message = "command created!"
	case deleteOp:
		name, _, _ := strings.Cut(rest, " ")
		if name == "" {
			response.Message = c.OpDescription(operation)
			break
		}
		_, err := c.userCommandService.Delete(ctx.Context, data.UserCommandDelete{
			UserID: ctx.Channel.UserID,
			Name:   name,
		})
		if err != nil {
			response.Message = "couldnt delete command, got error: " + err.Error()
			break
		}
		response.Message = "command deleted!"
	case updateOp:
		name, text, _ := strings.Cut(rest, " ")
		if text == "" || name == "" {
			response.Message = c.OpDescription(operation)
			break
		}
		_, err := c.userCommandService.Update(ctx.Context, data.UserCommandUpdate{
			UserID: ctx.Channel.UserID,
			Name:   name,
			Text:   &text,
		})
		if err != nil {
			response.Message = "couldnt update command, got error: " + err.Error()
			break
		}
		response.Message = "command updated!"
	default:
		response.Message = c.Description()
	}
	return response, nil
}
