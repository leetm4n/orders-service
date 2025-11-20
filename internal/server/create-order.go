package server

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leetm4n/orders-service/api"
	"github.com/leetm4n/orders-service/internal/model"
	"github.com/leetm4n/orders-service/internal/repo"
	"github.com/leetm4n/orders-service/pkg/tracing"
	openapiTypes "github.com/oapi-codegen/runtime/types"
)

func (s *ServerImpl) CreateOrder(w http.ResponseWriter, r *http.Request) {
	requestBody := api.CreateOrderRequest{}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	skuUUID := pgtype.UUID{}
	if err := skuUUID.Scan(requestBody.Sku.String()); err != nil {
		slog.Error("failed to convert sku to UUID", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	idempotencyKey := pgtype.UUID{}
	if requestBody.IdempotencyKey != nil {
		_ = idempotencyKey.Scan(requestBody.IdempotencyKey.String())
	}

	createdOrder, err := s.queries.CreateOrder(r.Context(), repo.CreateOrderParams{
		Quantity:        int32(requestBody.Quantity),
		ShippingAddress: requestBody.ShippingAddress,
		Sku:             skuUUID,
		IdempotencyKey:  idempotencyKey,
	})

	if err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code != pgerrcode.UniqueViolation {
			slog.Error("failed to create order", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		order, err := s.queries.GetOrderByIdempotencyKey(r.Context(), idempotencyKey)
		if err != nil {
			slog.Error("failed to get order by idempotency key after unique violation", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		response := api.Order{
			Id:              openapiTypes.UUID(order.ID.Bytes),
			Quantity:        int(order.Quantity),
			CreatedAt:       order.CreatedAt.Time,
			UpdatedAt:       order.UpdatedAt.Time,
			Status:          api.OrderStatus(order.Status),
			ShippingAddress: order.ShippingAddress,
			Sku:             openapiTypes.UUID(order.Sku.Bytes),
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			slog.Error("failed to write order response", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		return
	}

	orderModel := model.Order{
		ID:              createdOrder.ID.String(),
		Quantity:        int(createdOrder.Quantity),
		CreatedAt:       createdOrder.CreatedAt.Time,
		UpdatedAt:       createdOrder.UpdatedAt.Time,
		Status:          string(createdOrder.Status),
		ShippingAddress: createdOrder.ShippingAddress,
		Sku:             createdOrder.Sku.String(),
	}

	func(ctx context.Context) {
		ctx, span := s.tracer.Start(ctx, "emitOrderCreatedEvent")
		defer span.End()
		select {
		case s.orderCreatedChan <- model.OrderCreatedEvent{
			Order: orderModel,
			Trace: tracing.SerializeTraceCtx(ctx),
		}:
			slog.Info("order created event emitted")
		case <-ctx.Done():
			slog.Error("failed to emit order created event: context done", "error", ctx.Err())
		}
	}(r.Context())

	response := api.Order{
		Id:              openapiTypes.UUID(createdOrder.ID.Bytes),
		Quantity:        int(createdOrder.Quantity),
		CreatedAt:       createdOrder.CreatedAt.Time,
		UpdatedAt:       createdOrder.UpdatedAt.Time,
		Status:          api.OrderStatus(createdOrder.Status),
		ShippingAddress: createdOrder.ShippingAddress,
		Sku:             openapiTypes.UUID(createdOrder.Sku.Bytes),
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to write order response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
