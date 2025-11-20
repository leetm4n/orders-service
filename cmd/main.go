package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/leetm4n/orders-service/config"
	"github.com/leetm4n/orders-service/internal/application"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	slog.Info("application starting")

	if err := run(); err != nil {
		slog.Error("run failed", "error", err)

		os.Exit(1)
	}

	os.Exit(0)
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg := config.MustLoadConfig()

	if err := application.Run(ctx, cfg); err != nil {
		return err
	}

	return nil
}
