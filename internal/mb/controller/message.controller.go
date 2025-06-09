package controller

import (
	"log/slog"

	"arnobot-shared/applog"
	"arnobot-shared/apptype"
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
	conn.QueueSubscribe(topics.CoreChatMessageNotify, topics.CoreChatMessageNotify, c.NewChatMessage)
}

func (c *MessageController) NewChatMessage(msg *nats.Msg) {
	var payload apptype.CoreChatMessageNotify

	payload.Decode(msg.Data)

	ctx, cancel := newControllerContext(payload.TraceID)
	defer cancel()

	err := c.messageService.HandleNewMessage(ctx, payload.Data)
	if err != nil {
		c.logger.ErrorContext(ctx, "cannot handle message")
		return
	}
}
