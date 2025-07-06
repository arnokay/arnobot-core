package controller

import (
	

	"github.com/arnokay/arnobot-shared/applog"
	"github.com/arnokay/arnobot-shared/pkg/assert"
	"github.com/arnokay/arnobot-shared/topics"
	"github.com/nats-io/nats.go"

	"github.com/arnokay/arnobot-core/internal/app/service"
)

type MessageController struct {
	messageService *service.MessageService
	logger         applog.Logger
}

func NewMessageController(messageService *service.MessageService) *MessageController {
	logger := applog.NewServiceLogger("message-controller")

	return &MessageController{
		messageService: messageService,
		logger:         logger,
	}
}

func (c *MessageController) Connect(conn *nats.Conn) {
	topic := topics.TopicBuilder(topics.PlatformBroadcasterChatMessageNotify).
		BroadcasterID(topics.Any).
		Platform(topics.Any).
		Build()
	_, err := conn.QueueSubscribe(topic, topic, c.NewChatMessage)
	assert.NoError(err, "cannot start: "+topic)
}

func (c *MessageController) NewChatMessage(msg *nats.Msg) {
	handlePublish(msg, c.messageService.HandleNewMessage)
}
