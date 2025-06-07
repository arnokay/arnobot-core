package service

import (
	"arnobot-core/internal/cmdcenter"
	"arnobot-shared/applog"
	"arnobot-shared/events"
	"context"
	"log/slog"
)

type MessageService struct {
  commandManager *cmdcenter.CommandManager

  logger *slog.Logger
}

func NewMessageService(
  commandManager *cmdcenter.CommandManager,
) *MessageService {
  logger := applog.NewServiceLogger("message-service")

  return &MessageService{
    commandManager: commandManager,
    logger: logger,
  }
}

func (s *MessageService) HandleNewMessage(ctx context.Context, event events.Message) error {
  if s.commandManager.IsCommand(event) {
    s.commandManager.Execute(ctx, event)
  }

  return nil
}
