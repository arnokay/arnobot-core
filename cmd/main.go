package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/arnokay/arnobot-shared/applog"
	mbControllers "github.com/arnokay/arnobot-shared/controllers/mb"
	"github.com/arnokay/arnobot-shared/db"
	"github.com/arnokay/arnobot-shared/pkg/assert"
	sharedService "github.com/arnokay/arnobot-shared/service"
	"github.com/arnokay/arnobot-shared/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/arnokay/arnobot-core/internal/app/config"
	"github.com/arnokay/arnobot-core/internal/app/service"
	"github.com/arnokay/arnobot-core/internal/cmdcenter"
	"github.com/arnokay/arnobot-core/internal/cmdcenter/command"
	"github.com/arnokay/arnobot-core/internal/mb/controller"
)

const APP_NAME = "core"

type application struct {
	logger *slog.Logger

	db        *pgxpool.Pool
	queries   db.Querier
	cache     jetstream.KeyValue
	queue     jetstream.JetStream
	msgBroker *nats.Conn
	storage   storage.Storager

	services       *service.Services
	commandManager *cmdcenter.CommandManager
	mbControllers  mbControllers.NatsController
}

func main() {
	var app application

	ctx := context.Background()

	// load config
	cfg := config.Load()

	// load logger
	logger := applog.Init(APP_NAME, os.Stdout, cfg.Global.LogLevel)
	app.logger = logger

	// load db
	dbConn := openDB(ctx)
	app.db = dbConn

	// load message broker, queue and cache
	mbConn, queue, cache := openMB(ctx)
	app.msgBroker = mbConn
	app.queue = queue
	app.cache = cache

	// load storage
	store := storage.NewStorage(app.db)
	app.storage = store

	// load command manager
	app.commandManager = cmdcenter.NewCommandManager(app.cache)

	// load commands
	cmdPing := command.NewPingCommand()
	app.commandManager.Add(ctx, cmdPing)
	eightBall := command.NewEightBall()
	app.commandManager.Add(ctx, eightBall)
	dice := command.NewDiceCommand()
	app.commandManager.Add(ctx, dice)
	coin := command.NewCoinCommand()
	app.commandManager.Add(ctx, coin)
	gamba := command.NewGambaCommand()
	app.commandManager.Add(ctx, gamba)

	// load services
	services := &service.Services{}
	services.PlatformModuleService = sharedService.NewPlatformModuleService(app.msgBroker)
	services.MessageService = service.NewMessageService(app.commandManager, services.PlatformModuleService)
	app.services = services

	// load message broker controllers
	app.mbControllers = &controller.Controllers{
		MessageController: controller.NewMessageController(app.services.MessageService),
	}

	app.Start()
}

func openDB(ctx context.Context) *pgxpool.Pool {
	cfg, err := pgxpool.ParseConfig(config.Config.DB.DSN)
	assert.NoError(err, "openDB: cannot open database connection")

	cfg.MaxConns = int32(config.Config.DB.MaxOpenConns)
	cfg.MinConns = int32(config.Config.DB.MaxIdleConns)

	duration, err := time.ParseDuration(config.Config.DB.MaxIdleTime)
	assert.NoError(err, "openDB: cannot parse duration")

	cfg.MaxConnIdleTime = duration

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	assert.NoError(err, "openDB: cannot open database connection")

	err = pool.Ping(ctx)
	assert.NoError(err, "openDB: cannot ping")

	return pool
}

func openMB(ctx context.Context) (*nats.Conn, jetstream.JetStream, jetstream.KeyValue) {
	nc, err := nats.Connect(config.Config.MB.URL)
	assert.NoError(err, "openMB: cannot open message broker connection")

	js, err := jetstream.New(nc)
	assert.NoError(err, "openMB: cannot open jetstream")
	kv, err := js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket: "default-core",
	})
	assert.NoError(err, "openMB: cannot create KVstore")

	return nc, js, kv
}
