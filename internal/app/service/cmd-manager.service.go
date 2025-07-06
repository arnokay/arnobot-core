package service

import (
	"context"
	"errors"
	"log/slog"

	"strings"

	"github.com/arnokay/arnobot-shared/apperror"
	"github.com/arnokay/arnobot-shared/applog"
	"github.com/arnokay/arnobot-shared/data"
	"github.com/arnokay/arnobot-shared/events"
	"github.com/arnokay/arnobot-shared/platform"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/arnokay/arnobot-core/internal/commands/cmdtypes"
)

type CmdManagerService struct {
	cache jetstream.KeyValue

	commands     map[string]cmdtypes.Command
	commandNames map[string]bool

	logger applog.Logger
}

func NewCmdManagerService(cache jetstream.KeyValue) *CmdManagerService {
	logger := applog.NewServiceLogger("cmd-manager-service")

	return &CmdManagerService{
		cache:  cache,
		logger: logger,

		commands:     map[string]cmdtypes.Command{},
		commandNames: map[string]bool{},
	}
}

func (m *CmdManagerService) IsCommand(cmdName string) bool {
	hasPrefix := strings.HasPrefix(cmdName, cmdtypes.CommandPrefix)
	if !hasPrefix {
		return false
	}
	cmdName = strings.TrimPrefix(cmdName, cmdtypes.CommandPrefix)
	_, ok := m.commandNames[cmdName]
	return ok
}

func (m *CmdManagerService) IsCommandEvent(event events.Message) bool {
	hasPrefix := strings.HasPrefix(event.Message, cmdtypes.CommandPrefix)
	if !hasPrefix {
		return false
	}
	cmd := m.parseCommand(event.Message)
	_, ok := m.commandNames[cmd.Command]
	return ok
}

func (m *CmdManagerService) parseCommand(message string) cmdtypes.ParsedCommand {
	cmd, rest, _ := strings.Cut(message, " ")
	name := strings.TrimPrefix(cmd, cmdtypes.CommandPrefix)
	return cmdtypes.ParsedCommand{
		Prefix:  cmdtypes.CommandPrefix,
		Command: name,
		Args:    rest,
	}
}

func (m *CmdManagerService) setBroadcasterCommandCooldown(ctx context.Context, platform platform.Platform, broadcasterID string, cmd cmdtypes.Command) error {
	_, err := m.cache.Create(ctx, m.getCacheKey(platform, broadcasterID, cmd), []byte{}, jetstream.KeyTTL(cmd.Cooldown()))
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyExists) {
			m.logger.DebugContext(
				ctx,
				"command in cooldown",
				"platform", platform,
				"broadcasterID", broadcasterID,
				"cmd", m.getCommandLog(cmd),
			)
		} else {
			m.logger.ErrorContext(
				ctx,
				"cannot cache command cooldown",
				"err", err,
				"platform", platform,
				"broadcasterID", broadcasterID,
				"cmd", m.getCommandLog(cmd),
			)
			return apperror.ErrExternal
		}
	}

	return nil
}

func (m *CmdManagerService) getCacheKey(platform platform.Platform, broadcasterID string, cmd cmdtypes.Command) string {
	key := "cmdm." + platform.String() + "." + broadcasterID + "." + cmd.Name()
	return key
}

func (m *CmdManagerService) isBroadcasterCommandInCooldown(ctx context.Context, platform platform.Platform, broadcasterID string, cmd cmdtypes.Command) bool {
	_, err := m.cache.Get(ctx, m.getCacheKey(platform, broadcasterID, cmd))
	if err != nil {
		if !errors.Is(err, jetstream.ErrNoKeysFound) {
			m.logger.ErrorContext(
				ctx,
				"cannot get command cooldown from cache",
				"err", err,
				"broadcasterID", broadcasterID,
				"cmd", m.getCommandLog(cmd),
			)
		}
		return false
	}

	return true
}

func (m *CmdManagerService) Execute(ctx context.Context, event events.Message) (*events.MessageSend, error) {
	parsedMessage := m.parseCommand(event.Message)

	cmdCtx := cmdtypes.CommandContext{
		Context: ctx,
		Chatter: cmdtypes.PlatformUser{
			ID:       event.ChatterID,
			Name:     event.ChatterName,
			Login:    event.ChatterLogin,
			Role:     event.ChatterRole,
			Platform: event.Platform,
		},
		Channel: cmdtypes.PlatformUser{
			ID:       event.BroadcasterID,
			Name:     event.BroadcasterName,
			Login:    event.BroadcasterLogin,
			UserID:   event.UserID,
			Role:     data.ChatterBroadcaster,
			Platform: event.Platform,
		},
		Bot: cmdtypes.PlatformUser{
			ID: event.BotID,
		},
		Message: cmdtypes.Message{
			ID:      event.MessageID,
			Message: event.Message,
			ReplyTo: event.ReplyTo,
		},
		Command: parsedMessage,
	}

	cmd, ok := m.commands[parsedMessage.Command]

	if !ok {
		m.logger.ErrorContext(ctx, "there is no command to execute", "command", parsedMessage.Command)
		return nil, apperror.ErrInternal
	}

	if m.isBroadcasterCommandInCooldown(ctx, event.Platform, event.BroadcasterID, cmd) {
		m.logger.DebugContext(
			ctx,
			"command in cooldown",
			"platform", event.Platform,
			"broadcasterID", event.BroadcasterID,
			"cmd", m.getCommandLog(cmd),
		)
		return nil, apperror.ErrForbidden
	}

	err := m.setBroadcasterCommandCooldown(ctx, event.Platform, event.BroadcasterID, cmd)
	if err != nil {
		m.logger.DebugContext(
			ctx,
			"cannot set cmd cooldown or cache error",
			"err", err,
			"cmd", m.getCommandLog(cmd),
		)
	}
	cmdResponse, err := cmd.Execute(cmdCtx)
	if err != nil {
		if errors.Is(err, apperror.ErrNoAction) {
			return nil, err
		}
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

func (m *CmdManagerService) add(ctx context.Context, name string, cmd cmdtypes.Command) {
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

func (m *CmdManagerService) Add(ctx context.Context, cmd cmdtypes.Command) {
	m.add(ctx, cmd.Name(), cmd)

	for _, cmdName := range cmd.Aliases() {
		m.add(ctx, cmdName, cmd)
	}
}

func (m *CmdManagerService) getCommandLog(cmd cmdtypes.Command) slog.Value {
	return slog.GroupValue(
		slog.String("name", cmd.Name()),
		slog.String("description", cmd.Description()),
		slog.String("aliases", strings.Join(cmd.Aliases(), ",")),
	)
}
