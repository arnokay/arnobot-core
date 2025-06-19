package controller

import (
	"errors"
	"log/slog"

	"github.com/arnokay/arnobot-shared/apperror"
	"github.com/arnokay/arnobot-shared/applog"
	"github.com/arnokay/arnobot-shared/apptype"
	"github.com/arnokay/arnobot-shared/platform"
	"github.com/arnokay/arnobot-shared/topics"

	"github.com/nats-io/nats.go"

	"github.com/arnokay/arnobot-core/internal/app/service"
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
