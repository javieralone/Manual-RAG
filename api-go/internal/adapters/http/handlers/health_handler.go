package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"api-go/internal/adapters/observability"
)

type DependencyChecker interface {
	Check(ctx context.Context) error
}

type HealthHandler struct {
	checks  []DependencyChecker
	metrics *observability.Metrics
}

func NewHealthHandler(metrics *observability.Metrics, checks ...DependencyChecker) *HealthHandler {
	return &HealthHandler{checks: checks, metrics: metrics}
}

func (h *HealthHandler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "UP"})
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	if !h.UpdateReadiness(ctx) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "NOT_READY"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "READY"})
}

func (h *HealthHandler) UpdateReadiness(ctx context.Context) bool {
	ready := true
	for _, check := range h.checks {
		if err := check.Check(ctx); err != nil {
			ready = false
			break
		}
	}
	if h.metrics != nil {
		if ready {
			h.metrics.Readiness.Set(1)
		} else {
			h.metrics.Readiness.Set(0)
		}
	}
	return ready
}

func (h *HealthHandler) StartReadinessMonitor(ctx context.Context, interval time.Duration) {
	go func() {
		readinessTicker := time.NewTicker(interval)
		defer readinessTicker.Stop()

		for {
			checkContext, cancel := context.WithTimeout(ctx, 3*time.Second)
			h.UpdateReadiness(checkContext)
			cancel()

			select {
			case <-readinessTicker.C:
			case <-ctx.Done():
				return
			}
		}
	}()
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
