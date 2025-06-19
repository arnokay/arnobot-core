package controller

import (
	"context"
	"time"

	"github.com/arnokay/arnobot-shared/trace"
	"github.com/nats-io/nats.go"
)

type Controllers struct {
	MessageController *MessageController
}

func (c *Controllers) Connect(conn *nats.Conn) {
	c.MessageController.Connect(conn)
}

func newControllerContext(traceID string) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	ctx = trace.Context(ctx, traceID)

	return ctx, cancel
}
