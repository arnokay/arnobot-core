package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/arnokay/arnobot-shared/apperror"
	"github.com/arnokay/arnobot-shared/applog"
	"github.com/arnokay/arnobot-shared/events"
	"github.com/arnokay/arnobot-shared/service"
)

type MessageService struct {
	cmdManagerService     *CmdManagerService
	userCmdManagerService *UserCmdManagerService
	platformModuleService *service.PlatformModuleIn

	logger *slog.Logger
}

func NewMessageService(
	commandManager *CmdManagerService,
	userCommand *UserCmdManagerService,
	platformModuleService *service.PlatformModuleIn,
) *MessageService {
	logger := applog.NewServiceLogger("message-service")

	return &MessageService{
		cmdManagerService:     commandManager,
		platformModuleService: platformModuleService,
		userCmdManagerService: userCommand,
		logger:                logger,
	}
}

func (s *MessageService) HandleNewMessage(ctx context.Context, event events.Message) error {
	switch {
	case s.cmdManagerService.IsCommandEvent(event):
		s.logger.DebugContext(ctx, "new command", "event", event)
		response, err := s.cmdManagerService.Execute(ctx, event)
		if err != nil {
			if errors.Is(err, apperror.ErrNoAction) {
				s.logger.DebugContext(ctx, "no action is needed")
				return nil
			}
			return err
		}
		err = s.platformModuleService.ChatSendMessage(ctx, *response)
		if err != nil {
			s.logger.ErrorContext(ctx, "cannot send chat message")
			return err
		}
	case s.userCmdManagerService.IsCommandEvent(ctx, event):
		s.logger.DebugContext(ctx, "new user command", "event", event)
		response, err := s.userCmdManagerService.Execute(ctx, event)
		if err != nil {
			return err
		}
		err = s.platformModuleService.ChatSendMessage(ctx, *response)
		if err != nil {
			s.logger.ErrorContext(ctx, "cannot send chat message")
			return err
		}
	}

	return nil
}
