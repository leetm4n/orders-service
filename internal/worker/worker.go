package worker

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/leetm4n/orders-service/internal/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Worker struct {
	orderCreatedCh <-chan model.OrderCreatedEvent
	tracer         trace.Tracer
}

func New(orderCreatedCh <-chan model.OrderCreatedEvent) *Worker {
	return &Worker{
		orderCreatedCh: orderCreatedCh,
		tracer:         otel.Tracer("orders-ms-worker"),
	}
}

var ErrChannelClosed = errors.New("event emitter channel closed")

func (w *Worker) Start(ctx context.Context) error {
	slog.Info("worker starting")

	for {
		select {
		case <-ctx.Done():
			slog.Info("worker stopping due to context cancellation")

			return nil
		case order, ok := <-w.orderCreatedCh:
			if !ok {
				return ErrChannelClosed
			}

			if err := w.processOrderEvent(ctx, order); err != nil {
				slog.Error("failed to process order event", "error", err)
			}
		}
	}
}

func (w *Worker) processOrderEvent(ctx context.Context, event model.OrderCreatedEvent) error {
	slog.Info("processing order event", "orderId", event.Order.ID)

	select {
	case <-ctx.Done():
		slog.Info("worker stopping processing due to context cancellation")
	case <-time.After(2 * time.Second):
		// Simulated processing time
		slog.Info("order processed", "orderId", event.Order.ID)
	}

	return nil
}
