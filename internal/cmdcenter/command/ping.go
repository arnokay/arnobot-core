package command

import (
	"context"
	"time"
)

type PingCommand struct{}

func (c PingCommand) Name() string {
	return "ping"
}

func (c PingCommand) Description() string {
	return "im gonna ping"
}

func (c PingCommand) Cooldown() time.Duration {
	return time.Second * 5
}

func (c PingCommand) Execute(ctx context.Context) string {
  return "pong"
}
