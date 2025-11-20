package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leetm4n/orders-service/config"
	"github.com/leetm4n/orders-service/internal/server"
	"github.com/quantumsheep/otelpgxpool"
	"golang.org/x/sync/errgroup"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	slog.Info("starting")

	if err := run(); err != nil {
		slog.Error("run failed", "error", err)

		os.Exit(1)
	}

	os.Exit(0)
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	otelShutdown, err := InitTracer(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	cfg := config.MustLoadConfig()

	poolConfig, err := pgxpool.ParseConfig(cfg.DATABASE_URL)
	if err != nil {
		return err
	}

	poolConfig.ConnConfig.Tracer = otelpgxpool.NewTracer()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("new pool error: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping database error: %w", err)
	}

	server := server.NewServer(server.ServerOptions{
		Port:                       cfg.Port,
		Host:                       cfg.Host,
		GracefulShutdownTimeoutSec: cfg.GracefulShutdownTimeoutSec,
	})

	eG, ctx := errgroup.WithContext(ctx)
	eG.Go(func() error {
		return server.Start(ctx)
	})

	if err := eG.Wait(); err != nil {
		return err
	}

	return nil
}
