package controller

import (
	

	"github.com/arnokay/arnobot-shared/applog"
	"github.com/arnokay/arnobot-shared/topics"
	"github.com/nats-io/nats.go"

	"github.com/arnokay/arnobot-core/internal/app/service"
)

type UserCommandController struct {
	userCommandService *service.UserCommandService
	logger             applog.Logger
}

func NewUserCommandController(
	userCommandService *service.UserCommandService,
) *UserCommandController {
	logger := applog.NewServiceLogger("user-command-controller")

	return &UserCommandController{
		userCommandService: userCommandService,
		logger:             logger,
	}
}

func (c *UserCommandController) Connect(conn *nats.Conn) {
	conn.QueueSubscribe(topics.CoreUserCommandCreate, topics.CoreUserCommandCreate, c.Create)
	conn.QueueSubscribe(topics.CoreUserCommandUpdate, topics.CoreUserCommandUpdate, c.Update)
	conn.QueueSubscribe(topics.CoreUserCommandDelete, topics.CoreUserCommandDelete, c.Delete)
	conn.QueueSubscribe(topics.CoreUserCommandGetOne, topics.CoreUserCommandGetOne, c.GetOne)
	conn.QueueSubscribe(topics.CoreUserCommandGetByUserID, topics.CoreUserCommandGetByUserID, c.GetByUserID)
}

func (c *UserCommandController) GetByUserID(msg *nats.Msg) {
	handleRequest(msg, c.userCommandService.GetByUserID)
}

func (c *UserCommandController) GetOne(msg *nats.Msg) {
	handleRequest(msg, c.userCommandService.GetOne)
}

func (c *UserCommandController) Create(msg *nats.Msg) {
	handleRequest(msg, c.userCommandService.Create)
}

func (c *UserCommandController) Update(msg *nats.Msg) {
	handleRequest(msg, c.userCommandService.Update)
}

func (c *UserCommandController) Delete(msg *nats.Msg) {
	handleRequest(msg, c.userCommandService.Delete)
}
