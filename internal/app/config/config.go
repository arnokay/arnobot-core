package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/arnokay/arnobot-shared/pkg/assert"
)

const (
	ENV_PORT                 = "PORT"
	ENV_DB_DSN               = "DB_DSN"
	ENV_TWITCH_CLIENT_ID     = "TWITCH_CLIENT_ID"
	ENV_TWITCH_CLIENT_SECRET = "TWITCH_CLIENT_SECRET"
	ENV_TWITCH_REDIRECT_URI  = "TWITCH_REDIRECT_URI"
)

type config struct {
	Global    GlobalConfig
	DB        DBConfig
	MB        MBConfig
}

type GlobalConfig struct {
	Env      string
	Port     int
	LogLevel int
}

type MBConfig struct {
	URL string
}

type DBConfig struct {
	DSN          string
	MaxIdleConns int
	MaxOpenConns int
	MaxIdleTime  string
}

type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

var Config *config

func Load() *config {
	Config = &config{
		Global: GlobalConfig{
			Port:     3000,
			LogLevel: -4,
		},
	}

	if os.Getenv(ENV_PORT) != "" {
		port, err := strconv.Atoi(os.Getenv(ENV_PORT))
		assert.NoError(err, fmt.Sprintf("%v: not a number", ENV_PORT))
		Config.Global.Port = port
	}

	flag.StringVar(&Config.Global.Env, "env", "development", "Environment (development|staging|production)")

	flag.IntVar(&Config.Global.Port, "port", Config.Global.Port, "Server Port")
	flag.IntVar(&Config.Global.LogLevel, "log-level", Config.Global.LogLevel, "Minimal Log Level (default: -4)")

	flag.StringVar(&Config.DB.DSN, "db-dsn", os.Getenv(ENV_DB_DSN), "DB DSN")
	flag.IntVar(&Config.DB.MaxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.IntVar(&Config.DB.MaxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.StringVar(&Config.DB.MaxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.Parse()

	return Config
}
