package controller

import (
	"fmt"

	"arnobot-shared/apptype"
	"arnobot-shared/topics"

	"github.com/nats-io/nats.go"
)

type MessageController struct{}

func (c *MessageController) Connect(conn *nats.Conn) {
  conn.QueueSubscribe(topics.CoreChatMessageNotify, topics.CoreChatMessageNotify, c.NewChatMessage)
}

func (c *MessageController) NewChatMessage(msg *nats.Msg) {
	var payload apptype.CoreChatMessageNotify

	payload.Decode(msg.Data)

	fmt.Println(payload)
}
