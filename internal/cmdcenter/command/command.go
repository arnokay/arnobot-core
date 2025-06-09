package command

import (
	"context"
	"time"

	"arnobot-shared/platform"
)

const CommandPrefix string = "!"

type Command interface {
	Name() string
	Aliases() []string
	Description() string
	Cooldown() time.Duration
	Execute(ctx CommandContext) (CommandResponse, error)
}

type CommandContext interface {
	Chatter() *PlatformUser
	Channel() *PlatformUser
	Bot() *PlatformUser
	Message() *Message
	Command() *ParsedCommand
	Context() context.Context
}

type PlatformUser struct {
	ID       string
	Name     string
	Login    string
	Platform platform.Platform
}

type Message struct {
	ID      string
	Message string
	ReplyTo string
}

type ParsedCommand struct {
	Prefix  string
	Command string
	Args    string
}

type CommandResponse struct {
	Message string
	ReplyTo string
	Private bool
}

func (c CommandResponse) ShouldRespond() bool {
	return c.Message != ""
}

type CommonCommand struct{}

func (c *CommonCommand) Aliases() []string {
	return []string{}
}

func (c *CommonCommand) Description() string {
	return ""
}

func (c *CommonCommand) Cooldown() time.Duration {
	return time.Second * 5
}
