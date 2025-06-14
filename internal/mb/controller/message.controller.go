package controller

import (
	"errors"
	"log/slog"

	"arnobot-shared/apperror"
	"arnobot-shared/applog"
	"arnobot-shared/apptype"
	"arnobot-shared/platform"
	"arnobot-shared/topics"

	"github.com/nats-io/nats.go"

	"arnobot-core/internal/app/service"
)

type MessageController struct {
	messageService *service.MessageService
	logger         *slog.Logger
}

func NewMessageController(messageService *service.MessageService) *MessageController {
	logger := applog.NewServiceLogger("message-controller")

	return &MessageController{
		messageService: messageService,
		logger:         logger,
	}
}

func (c *MessageController) Connect(conn *nats.Conn) {
	chatMessageNotifyTopic := topics.PlatformBroadcasterChatMessageNotify.Build(platform.All, topics.Any)
	conn.QueueSubscribe(chatMessageNotifyTopic, chatMessageNotifyTopic, c.NewChatMessage)
}

func (c *MessageController) NewChatMessage(msg *nats.Msg) {
	var payload apptype.CoreChatMessageNotify

	payload.Decode(msg.Data)

	ctx, cancel := newControllerContext(payload.TraceID)
	defer cancel()

	err := c.messageService.HandleNewMessage(ctx, payload.Data)
	if err != nil && !errors.Is(err, apperror.ErrNoAction) {
		c.logger.ErrorContext(ctx, "cannot handle message")
		return
	}
}
