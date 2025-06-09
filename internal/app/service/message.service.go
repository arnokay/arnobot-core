package service

import (
	"context"
	"errors"
	"log/slog"

	"arnobot-shared/apperror"
	"arnobot-shared/applog"
	"arnobot-shared/events"
	"arnobot-shared/service"

	"arnobot-core/internal/cmdcenter"
)

type MessageService struct {
	commandManager        *cmdcenter.CommandManager
	platformModuleService *service.PlatformModuleService

	logger *slog.Logger
}

func NewMessageService(
	commandManager *cmdcenter.CommandManager,
	platformModuleService *service.PlatformModuleService,
) *MessageService {
	logger := applog.NewServiceLogger("message-service")

	return &MessageService{
		commandManager:        commandManager,
		platformModuleService: platformModuleService,
		logger:                logger,
	}
}

func (s *MessageService) HandleNewMessage(ctx context.Context, event events.Message) error {
	if s.commandManager.IsCommand(event) {
		s.logger.DebugContext(ctx, "new command", "event", event)
		response, err := s.commandManager.Execute(ctx, event)
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
	}

	return nil
}
