package model

import (
	"time"

	"github.com/leetm4n/orders-service/pkg/tracing"
)

type Order struct {
	ID              string    `json:"id"`
	Quantity        int       `json:"quantity"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	Status          string    `json:"status"`
	ShippingAddress string    `json:"shippingAddress"`
	Sku             string    `json:"sku"`
}

type OrderCreatedEvent struct {
	Order Order                 `json:"order"`
	Trace tracing.TraceEnvelope `json:"trace"`
}
