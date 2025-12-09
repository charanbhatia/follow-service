package health

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type HealthChecker struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewHealthChecker(db *sql.DB, logger *zap.Logger) *HealthChecker {
	return &HealthChecker{
		db:     db,
		logger: logger,
	}
}

func (h *HealthChecker) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *HealthChecker) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		h.logger.Error("readiness check failed", zap.Error(err))
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Database unavailable"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}

func (h *HealthChecker) StartHealthServer(port string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health/live", h.LivenessHandler)
	mux.HandleFunc("/health/ready", h.ReadinessHandler)

	h.logger.Info("starting health check server", zap.String("port", port))
	return http.ListenAndServe(":"+port, mux)
}
