package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"arnobot-shared/applog"
	"arnobot-shared/cache"
	"arnobot-shared/cache/mapcacher"
	mbControllers "arnobot-shared/controllers/mb"
	"arnobot-shared/db"
	"arnobot-shared/pkg/assert"
	"arnobot-shared/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"arnobot-core/internal/app/config"
	"arnobot-core/internal/app/service"
	"arnobot-core/internal/mb/controller"
)

const APP_NAME = "core"

type application struct {
	logger *slog.Logger

	db        *pgxpool.Pool
	queries   db.Querier
	cache     cache.Cacher
	msgBroker *nats.Conn
	storage   storage.Storager

	services      *service.Services
	mbControllers mbControllers.NatsController
}

func main() {
	var app application

	// load config
	cfg := config.Load()

	// load logger
	logger := applog.Init(APP_NAME, os.Stdout, cfg.Global.LogLevel)
	app.logger = logger

	// load db
	dbConn := openDB()
	app.db = dbConn

	// load cache
	cache := mapcacher.New()
	app.cache = &cache

	// load message broker
	mbConn, _, _ := openMB()
	app.msgBroker = mbConn

	// load storage
	store := storage.NewStorage(app.db)
	app.storage = store

	// load services
	services := &service.Services{}
	app.services = services

	// load message broker controllers
	app.mbControllers = &controller.Controllers{}

	app.Start()
}

func openDB() *pgxpool.Pool {
	cfg, err := pgxpool.ParseConfig(config.Config.DB.DSN)
	assert.NoError(err, "openDB: cannot open database connection")

	cfg.MaxConns = int32(config.Config.DB.MaxOpenConns)
	cfg.MinConns = int32(config.Config.DB.MaxIdleConns)

	duration, err := time.ParseDuration(config.Config.DB.MaxIdleTime)
	assert.NoError(err, "openDB: cannot parse duration")

	cfg.MaxConnIdleTime = duration

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	assert.NoError(err, "openDB: cannot open database connection")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = pool.Ping(ctx)
	assert.NoError(err, "openDB: cannot ping")

	return pool
}

func openMB() (*nats.Conn, jetstream.JetStream, jetstream.KeyValue) {
	nc, err := nats.Connect(config.Config.MB.URL)
	assert.NoError(err, "openMB: cannot open message broker connection")

	js, err := jetstream.New(nc)
	assert.NoError(err, "openMB: cannot open jetstream")
	kv, err := js.CreateKeyValue(context.Background(), jetstream.KeyValueConfig{
		Bucket: "default-core",
	})
	assert.NoError(err, "openMB: cannot create KVstore")

  v, err := kv.Get(context.Background(), "kek")

	return nc, js, kv
}
