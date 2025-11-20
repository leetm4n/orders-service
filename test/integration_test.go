package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"github.com/davecgh/go-spew/spew"
	"github.com/leetm4n/orders-service/api"
	"github.com/leetm4n/orders-service/config"
	"github.com/leetm4n/orders-service/db"
	"github.com/leetm4n/orders-service/internal/application"
	openapiTypes "github.com/oapi-codegen/runtime/types"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestAppWithPostgres(t *testing.T) {
	ctx := t.Context()

	pgContainer, err := postgres.Run(ctx,
		"postgres:18",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer pgContainer.Terminate(ctx)

	host, _ := pgContainer.Host(ctx)
	port, _ := pgContainer.MappedPort(ctx, "5432/tcp")

	databaseURL := fmt.Sprintf(
		"postgres://testuser:testpass@%s:%s/testdb?sslmode=disable",
		host, port.Port(),
	)

	url, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatalf("failed to parse database URL: %v", err)
	}

	migrator := dbmate.New(url)
	migrator.MigrationsDir = []string{"./migrations"}
	migrator.FS = db.Migrations

	if err := migrator.Migrate(); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	cancellableContext, cancel := context.WithCancel(ctx)
	defer cancel()

	go application.Run(cancellableContext, config.Config{
		DatabaseURL: databaseURL,
		Port:        8085,
		Host:        "",
	})

	// Give the app time to start
	time.Sleep(1 * time.Second)

	// Test healthz first
	resp, err := http.Get("http://localhost:8085/healthz")
	if err != nil {
		t.Fatalf("http request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	// Create Order
	idempotencyKey := openapiTypes.UUID{}
	sku := openapiTypes.UUID{}

	_ = idempotencyKey.UnmarshalText([]byte("3deb76e4-cd89-4aa3-b143-89e9c0ed11db"))
	_ = sku.UnmarshalText([]byte("3deb76e4-cd89-4aa3-b143-89e9c0ed11ad"))

	b, err := json.Marshal(api.CreateOrderRequest{
		Quantity:        4,
		ShippingAddress: "address",
		IdempotencyKey:  &idempotencyKey,
		Sku:             sku,
	})
	if err != nil {
		t.Fatalf("error marshalling request, got %v", err)

	}
	resp, err = http.Post("http://localhost:8085/orders", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("http request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", resp.StatusCode)
	}

	if err != nil {
		t.Fatalf("err reading body: %v", err)
	}

	// Create order again with same idempotency key
	resp, err = http.Post("http://localhost:8085/orders", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("http request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	order := map[string]interface{}{}
	if err := json.Unmarshal(respBody, &order); err != nil {
		t.Fatalf("err reading body: %v", err)
	}

	id, ok := order["id"].(string)
	if !ok {
		t.Fatalf("cannot get order id")
	}

	spew.Dump(fmt.Sprintf("http://localhost:8085/orders/%s", id))

	// Get Order by ID
	resp, err = http.Get(fmt.Sprintf("http://localhost:8085/orders/%s", id))
	if err != nil {
		t.Fatalf("http request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
}
