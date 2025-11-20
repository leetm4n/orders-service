package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/leetm4n/orders-service/api"
)

func (s *ServerImpl) GetHealth(w http.ResponseWriter, r *http.Request) {
	resp := api.GetHealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to write health response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
