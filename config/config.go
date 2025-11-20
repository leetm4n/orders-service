package config

import (
	"log/slog"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port                       int    `envconfig:"PORT" default:"8080"`
	Host                       string `envconfig:"HOST" default:""`
	GracefulShutdownTimeoutSec int    `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT_SEC" default:"2"`
	DatabaseURL                string `envconfig:"DATABASE_URL"`
	TracingExporterEndpoint    string `envconfig:"TRACING_EXPORTER_ENDPOINT" default:"localhost:4317"`
}

func MustLoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		slog.Debug("could not load .env")
	}

	var cfg Config
	envconfig.MustProcess("", &cfg)

	return cfg
}
