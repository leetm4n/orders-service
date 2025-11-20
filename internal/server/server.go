package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/leetm4n/orders-service/api"
	"github.com/leetm4n/orders-service/internal/repo"
	"github.com/leetm4n/orders-service/pkg/middlewares"
	validationMw "github.com/oapi-codegen/nethttp-middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/sync/errgroup"
)

var _ api.ServerInterface = (*ServerImpl)(nil)

type ServerImpl struct {
	queries *repo.Queries
}

type Server struct {
	port                       int
	host                       string
	gracefulShutdownTimeoutSec int
	server                     *http.Server
}

type ServerOptions struct {
	Port                       int
	Host                       string
	GracefulShutdownTimeoutSec int
	Queries                    *repo.Queries
}

func New(opts ServerOptions) *Server {
	mux := http.NewServeMux()
	spec, err := openapi3.NewLoader().LoadFromData([]byte(api.APIYaml))
	if err != nil {
		slog.Error("failed to load openapi spec", "error", err)
		panic("failed to load openapi spec")
	}

	validationMW := validationMw.OapiRequestValidatorWithOptions(spec, &validationMw.Options{
		ErrorHandler: func(w http.ResponseWriter, message string, statusCode int) {
			w.WriteHeader(statusCode)
			_ = json.NewEncoder(w).Encode(api.ErrorResponse{
				Error: message,
				Code:  statusCode,
			})
		},
	})

	s := &ServerImpl{
		queries: opts.Queries,
	}

	isNotHealthzPath := func(r *http.Request) bool {
		return r.URL.Path != "/healthz"
	}

	otelMux := otelhttp.NewHandler(api.HandlerFromMux(s, mux), "orders-service-server", otelhttp.WithFilter(isNotHealthzPath))

	handler := middlewares.ContentTypeSetterMW(validationMW(otelMux))

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", opts.Host, opts.Port),
		Handler: handler,
	}

	return &Server{
		port:                       opts.Port,
		host:                       opts.Host,
		gracefulShutdownTimeoutSec: opts.GracefulShutdownTimeoutSec,
		server:                     server,
	}
}

func (s *Server) Start(ctx context.Context) error {
	slog.Info(fmt.Sprintf("server starting on %s:%d", s.host, s.port))

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("err running server: %w", err)

		}

		return nil
	})

	eg.Go(func() error {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(),
			time.Duration(s.gracefulShutdownTimeoutSec)*time.Second)
		defer cancel()

		if err := s.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("err shutting down server: %w", err)
		}

		return nil
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	slog.Info("server stopped gracefully")

	return nil
}
