package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leetm4n/orders-service/api"
	openapiTypes "github.com/oapi-codegen/runtime/types"
)

func (s *ServerImpl) GetOrderById(w http.ResponseWriter, r *http.Request, orderId openapiTypes.UUID) {
	uuid := pgtype.UUID{}
	if err := uuid.Scan(orderId.String()); err != nil {
		slog.Error("failed to convert id to UUID", "error", err)

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	order, err := s.queries.GetOrderByID(r.Context(), uuid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(api.ErrorResponse{
				Error: "order not found",
				Code:  http.StatusNotFound,
			})
			return
		}

		slog.Error("failed to get order by id", "error", err)
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
}
