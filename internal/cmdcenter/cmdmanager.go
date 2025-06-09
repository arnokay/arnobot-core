package cmdcenter

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"arnobot-shared/apperror"
	"arnobot-shared/applog"
	"arnobot-shared/events"
	"github.com/nats-io/nats.go/jetstream"

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
	cache jetstream.KeyValue

	commands     map[string]command.Command
	commandNames map[string]bool

	logger *slog.Logger
}

func NewCommandManager(cache jetstream.KeyValue) *CommandManager {
	logger := applog.NewServiceLogger("command-manager")

	return &CommandManager{
		cache:  cache,
		logger: logger,

		commands:     map[string]command.Command{},
		commandNames: map[string]bool{},
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
		Prefix:  command.CommandPrefix,
		Command: name,
		Args:    rest,
	}
}

func (m *CommandManager) setBroadcasterCommandCooldown(ctx context.Context, broadcasterID string, cmd command.Command) error {
	cooldown := time.Now().Add(cmd.Cooldown())
	b, _ := cooldown.MarshalBinary()
	_, err := m.cache.Put(ctx, m.getCacheKey(broadcasterID, cmd), b)
	if err != nil {
		m.logger.ErrorContext(
			ctx,
			"cannot cache command cooldown",
			"broadcasterID", broadcasterID,
			"cmd", m.getCommandLog(cmd),
		)
		return apperror.ErrExternal
	}

	return nil
}

func (m *CommandManager) getCacheKey(broadcasterID string, cmd command.Command) string {
	return "channel:" + broadcasterID + ":cmd:" + cmd.Name()
}

func (m *CommandManager) getBroadcasterCommandCooldown(ctx context.Context, broadcasterID string, cmd command.Command) (time.Time, error) {
	value, err := m.cache.Get(ctx, m.getCacheKey(broadcasterID, cmd))
	if err != nil {
		m.logger.ErrorContext(
			ctx,
			"cannot get command cooldown from cache",
			"broadcasterID", broadcasterID,
			"cmd", m.getCommandLog(cmd),
		)
		return time.Time{}, apperror.ErrExternal
	}
	var cooldown time.Time
	err = cooldown.UnmarshalBinary(value.Value())
	if err != nil {
		m.logger.ErrorContext(
			ctx,
			"cannot unmarshal cooldown",
			"err", err,
			"broadcasterID", broadcasterID,
			"cmd", m.getCommandLog(cmd),
		)
		return time.Time{}, apperror.ErrInternal
	}

	return cooldown, nil
}

func (m *CommandManager) Execute(ctx context.Context, event events.Message) (*events.MessageSend, error) {
	parsedMessage := m.parseCommand(event.Message)

	cmdCtx := CommandContext{
		context: ctx,
		chatter: &command.PlatformUser{
			ID:       event.ChatterID,
			Name:     event.ChatterName,
			Login:    event.ChatterLogin,
			Platform: event.Platform,
		},
		channel: &command.PlatformUser{
			ID:       event.BroadcasterID,
			Name:     event.BroadcasterName,
			Login:    event.BroadcasterLogin,
			Platform: event.Platform,
		},
		bot: &command.PlatformUser{
			ID: event.BotID,
		},
		message: &command.Message{
			ID:      event.MessageID,
			Message: event.Message,
			ReplyTo: event.ReplyTo,
		},
		command: &parsedMessage,
	}

	cmd, ok := m.commands[parsedMessage.Command]

	if !ok {
		m.logger.ErrorContext(ctx, "there is no command to execute", "command", parsedMessage.Command)
		return nil, apperror.ErrInternal
	}

	cooldown, err := m.getBroadcasterCommandCooldown(ctx, event.BroadcasterID, cmd)
	if err != nil {
		m.logger.DebugContext(
			ctx,
			"no command cooldown or cache error",
			"err", err,
			"cmd", m.getCommandLog(cmd),
		)
	}

	if time.Now().Before(cooldown) {
		m.logger.DebugContext(
			ctx,
			"command in cooldown",
			"now", time.Now(),
			"cooldown", cooldown,
			"cmd", m.getCommandLog(cmd),
		)
		return nil, apperror.ErrForbidden
	}

	err = m.setBroadcasterCommandCooldown(ctx, event.BroadcasterID, cmd)
	if err != nil {
		m.logger.DebugContext(
			ctx,
			"cannot set cmd cooldown or cache error",
			"err", err,
			"cmd", m.getCommandLog(cmd),
		)
	}
	cmdResponse, err := cmd.Execute(&cmdCtx)
	if err != nil {
		m.logger.ErrorContext(
			ctx,
			"cannot execute command",
			"err", err,
			"cmd", slog.GroupValue(
				slog.String("name", cmd.Name()),
				slog.String("description", cmd.Description()),
				slog.String("aliases", strings.Join(cmd.Aliases(), ",")),
			),
		)
		return nil, apperror.ErrInternal
	}

	if !cmdResponse.ShouldRespond() {
		return nil, apperror.ErrNoAction
	}

	var response events.MessageSend

	response.BroadcasterID = event.BroadcasterID
	response.BotID = event.BotID
	response.Platform = event.Platform
	response.Message = cmdResponse.Message
	response.ReplyTo = cmdResponse.ReplyTo

	return &response, nil
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

func (m *CommandManager) getCommandLog(cmd command.Command) string {
	return slog.GroupValue(
		slog.String("name", cmd.Name()),
		slog.String("description", cmd.Description()),
		slog.String("aliases", strings.Join(cmd.Aliases(), ",")),
	).String()
}
