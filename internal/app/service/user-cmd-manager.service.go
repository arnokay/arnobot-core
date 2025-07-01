package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/arnokay/arnobot-shared/apperror"
	"github.com/arnokay/arnobot-shared/applog"
	"github.com/arnokay/arnobot-shared/data"
	"github.com/arnokay/arnobot-shared/events"
	"github.com/arnokay/arnobot-shared/platform"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
)

type UserCmdManagerService struct {
	cache              jetstream.KeyValue
	cmdManagerService     *CmdManagerService
	userCommandService *UserCommandService

	logger *slog.Logger
}

func NewUserCmdManagerService(
	cache jetstream.KeyValue,
	commandManager *CmdManagerService,
	userCommandService *UserCommandService,
) *UserCmdManagerService {
	logger := applog.NewServiceLogger("user-cmd-manager-service")

	return &UserCmdManagerService{
		cache:              cache,
		cmdManagerService:     commandManager,
		userCommandService: userCommandService,

		logger: logger,
	}
}

func (s *UserCmdManagerService) getCommandKVKey(userID uuid.UUID, name string) string {
	return "ucs." + userID.String() + "." + name
}

func (s *UserCmdManagerService) getCommandCooldownKVKey(platform platform.Platform, broadcasterID string, cmd data.UserCommand) string {
	key := "ucs." + platform.String() + "." + broadcasterID + "." + cmd.Name
	return key
}

func (s *UserCmdManagerService) parseCommand(message string) string {
	cmd, _, _ := strings.Cut(message, " ")
	return cmd
}

func (s *UserCmdManagerService) IsCommandEvent(ctx context.Context, event events.Message) bool {
	_, err := s.userCommandService.GetOne(ctx, data.UserCommandGetOne{
		UserID: event.UserID,
		Name:   s.parseCommand(event.Message),
	})
	return err == nil
}

func (s *UserCmdManagerService) setBroadcasterCommandCooldown(ctx context.Context, platform platform.Platform, broadcasterID string, cmd data.UserCommand) error {
	_, err := s.cache.Create(
		ctx,
		s.getCommandCooldownKVKey(platform, broadcasterID, cmd),
		[]byte{},
		jetstream.KeyTTL(time.Second*10),
	)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyExists) {
			s.logger.DebugContext(
				ctx,
				"user command in cooldown",
				"platform", platform,
				"broadcasterID", broadcasterID,
				"cmd", cmd,
			)
		} else {
			s.logger.ErrorContext(
				ctx,
				"cannot cache command cooldown",
				"err", err,
				"platform", platform,
				"broadcasterID", broadcasterID,
				"cmd", cmd,
			)
			return apperror.ErrExternal
		}
	}

	return nil
}

func (s *UserCmdManagerService) isBroadcasterCommandInCooldown(ctx context.Context, platform platform.Platform, broadcasterID string, cmd data.UserCommand) bool {
	_, err := s.cache.Get(ctx, s.getCommandCooldownKVKey(platform, broadcasterID, cmd))
	if err != nil {
		if !errors.Is(err, jetstream.ErrNoKeysFound) {
			s.logger.ErrorContext(
				ctx,
				"cannot get command cooldown from cache",
				"err", err,
				"broadcasterID", broadcasterID,
				"cmd", cmd,
			)
		}
		return false
	}

	return true
}

func (s *UserCmdManagerService) Execute(ctx context.Context, event events.Message) (*events.MessageSend, error) {
	userCommand, err := s.userCommandService.GetOne(ctx, data.UserCommandGetOne{
		UserID: event.UserID,
		Name:   s.parseCommand(event.Message),
	})
	if err != nil {
		return nil, err
	}

	if s.isBroadcasterCommandInCooldown(ctx, event.Platform, event.BroadcasterID, userCommand) {
		s.logger.DebugContext(
			ctx,
			"command in cooldown",
			"platform", event.Platform,
			"broadcasterID", event.BroadcasterID,
			"cmd", userCommand,
		)
		return nil, apperror.ErrForbidden
	}
	err = s.setBroadcasterCommandCooldown(ctx, event.Platform, event.BroadcasterID, userCommand)
	if err != nil {
		s.logger.DebugContext(
			ctx,
			"cannot set cmd cooldown or cache error",
			"err", err,
			"cmd", userCommand,
		)
	}

	resp := events.MessageSend{
		Message: userCommand.Text,
	}

	if userCommand.Reply {
		resp.ReplyTo = event.MessageID
	}

	var response events.MessageSend

	response.BroadcasterID = event.BroadcasterID
	response.BotID = event.BotID
	response.Platform = event.Platform
	response.Message = userCommand.Text
	if userCommand.Reply {
		response.ReplyTo = event.MessageID
	}
	return &response, nil
}
