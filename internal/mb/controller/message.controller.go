package controller

import (
	"fmt"

	"arnobot-shared/mbtypes"
	"arnobot-shared/topics"

	"github.com/nats-io/nats.go"
)

type MessageController struct{}

func (c *MessageController) Connect(conn *nats.Conn) {
  conn.QueueSubscribe(topics.CoreChatMessageNotify, topics.CoreChatMessageNotify, c.NewChatMessage)
}

func (c *MessageController) NewChatMessage(msg *nats.Msg) {
	var payload mbtypes.CoreChatMessageNotify

	payload.Decode(msg.Data)

	fmt.Println(payload)
}
