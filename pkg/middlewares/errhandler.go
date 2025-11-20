package middlewares

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/leetm4n/orders-service/api"
)

func ErrorHandlerMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("recovered from panic: %v", r)

				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(api.ErrorResponse{
					Error: "Internal Server Error",
					Code:  http.StatusInternalServerError,
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}
