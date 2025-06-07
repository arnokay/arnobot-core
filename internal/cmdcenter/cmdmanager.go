package cmdcenter

import (
	"context"
	"log/slog"
	"strings"

	"arnobot-shared/apperror"
	"arnobot-shared/applog"
	"arnobot-shared/cache"
	"arnobot-shared/events"

	"arnobot-core/internal/cmdcenter/command"
)

type CommandContext struct {
	context context.Context
	chatter *command.PlatformUser
	channel *command.PlatformUser
	bot     *command.PlatformUser
	message *command.Message
	command *command.ParsedCommand
}

func (c *CommandContext) Chatter() *command.PlatformUser {
	return c.chatter
}

func (c *CommandContext) Channel() *command.PlatformUser {
	return c.channel
}

func (c *CommandContext) Bot() *command.PlatformUser {
	return c.bot
}

func (c *CommandContext) Message() *command.Message {
	return c.message
}

func (c *CommandContext) Command() *command.ParsedCommand {
	return c.command
}

func (c *CommandContext) Context() context.Context {
	return c.context
}

type CommandManager struct {
	cache cache.Cacher

	commands     map[string]command.Command
	commandNames map[string]bool

	logger *slog.Logger
}

func NewCommandManager(cache cache.Cacher) *CommandManager {
	logger := applog.NewServiceLogger("command-manager")

	return &CommandManager{
		cache:  cache,
		logger: logger,
	}
}

func (m *CommandManager) IsCommand(event events.Message) bool {
	hasPrefix := strings.HasPrefix(event.Message, command.CommandPrefix)
	if !hasPrefix {
		return false
	}
	cmd := m.parseCommand(event.Message)
	_, ok := m.commandNames[cmd.Command]
	return ok
}

func (m *CommandManager) parseCommand(message string) command.ParsedCommand {
	cmd, rest, _ := strings.Cut(message, " ")
	name := strings.TrimPrefix(cmd, command.CommandPrefix)
	return command.ParsedCommand{
    Prefix: command.CommandPrefix,
    Command: name,
    Args: rest,
  } 
}

func (m *CommandManager) Execute(ctx context.Context, event events.Message) (events.MessageSend, error) {
  parsedMessage := m.parseCommand(event.Message)

  cmdCtx := CommandContext{
    context: ctx,
    chatter: &command.PlatformUser{
      ID: event.ChatterID,
      Name: event.ChatterName,
      Login: event.ChatterLogin,
      Platform: event.Platform,
    },
    channel: &command.PlatformUser{
      ID: event.BroadcasterID,
      Name: event.BroadcasterName,
      Login: event.BroadcasterLogin,
      Platform: event.Platform,
    },
    bot: &command.PlatformUser{
      ID: event.BotID,
    },
    message: &command.Message{
      ID: event.MessageID,
      Message: event.Message,
      ReplyTo: event.ReplyTo,
    },
    command: &parsedMessage,
  }

  command, ok := m.commands[parsedMessage.Command]
  
  if !ok {
    m.logger.ErrorContext(ctx, "there is no command to execute", "command", parsedMessage.Command)
    return events.MessageSend{}, apperror.ErrInternal
  }

  //TODO: add cooldown
  cmdResponse, err := command.Execute(&cmdCtx)
  if err != nil {
    m.logger.ErrorContext(
      ctx, 
      "cannot execute command", 
      "err", err, 
      "cmd", slog.GroupValue(
        slog.String("name", command.Name()),
        slog.String("description", command.Description()),
        slog.String("aliases", strings.Join(command.Aliases(), ",")),
      ),
    )
    return events.MessageSend{}, apperror.ErrInternal
  }

  var response events.MessageSend

  response.BroadcasterID = event.BroadcasterID
  response.BotID = event.BotID
  response.Platform = event.Platform
  response.Message = cmdResponse.Message
  response.ReplyTo = cmdResponse.ReplyTo

	return response, nil
}

func (m *CommandManager) add(ctx context.Context, name string, cmd command.Command) {
	if _, ok := m.commands[name]; ok {
		m.logger.WarnContext(
			ctx,
			"attempt at setting already existing command",
			"provided_name", name,
			"cmd", slog.GroupValue(
				slog.String("name", cmd.Name()),
				slog.String("description", cmd.Description()),
				slog.String("aliases", strings.Join(cmd.Aliases(), ",")),
			),
		)
		return
	}
	m.commands[name] = cmd
	m.commandNames[name] = true
	m.logger.DebugContext(ctx, "added command to command manager", "cmdName", name)
}

func (m *CommandManager) Add(ctx context.Context, cmd command.Command) {
	m.add(ctx, cmd.Name(), cmd)

	for _, cmdName := range cmd.Aliases() {
		m.add(ctx, cmdName, cmd)
	}
}
