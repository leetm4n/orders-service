package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leetm4n/orders-service/config"
	"github.com/leetm4n/orders-service/internal/model"
	"github.com/leetm4n/orders-service/internal/repo"
	"github.com/leetm4n/orders-service/internal/server"
	"github.com/leetm4n/orders-service/internal/worker"
	"github.com/leetm4n/orders-service/pkg/tracing"
	"github.com/quantumsheep/otelpgxpool"
	"golang.org/x/sync/errgroup"
)

func Run(ctx context.Context, cfg config.Config) error {
	otelShutdown, err := tracing.InitTracer(ctx, "orders-service", cfg.TracingExporterEndpoint)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
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

	orderCreatedCh := make(chan model.OrderCreatedEvent, 100)
	defer close(orderCreatedCh)

	server := server.New(server.ServerOptions{
		Port:                       cfg.Port,
		Host:                       cfg.Host,
		GracefulShutdownTimeoutSec: cfg.GracefulShutdownTimeoutSec,
		Queries:                    repo.New(pool),
		OrderCreatedChan:           orderCreatedCh,
	})

	eG, ctx := errgroup.WithContext(ctx)

	// start server
	eG.Go(func() error {
		return server.Start(ctx)
	})

	// start worker
	worker := worker.New(orderCreatedCh)
	eG.Go(func() error {
		return worker.Start(ctx)
	})

	if err := eG.Wait(); err != nil {
		return err
	}

	return nil
}
