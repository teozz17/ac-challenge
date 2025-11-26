package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

var startTime = time.Now()

type Check struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
	Uptime string            `json:"uptime"`
}

type Handler struct {
	db *mongo.Database
}

func NewHandler(db *mongo.Database) *Handler {
	return &Handler{db: db}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	checks := make(map[string]string)
	overallStatus := "healthy"

	if err := h.db.Client().Ping(ctx, nil); err != nil {
		checks["mongodb"] = "unhealthy: " + err.Error()
		overallStatus = "unhealthy"
	} else {
		checks["mongodb"] = "ok"
	}

	if os.Getenv("OPENAI_API_KEY") == "" {
		checks["openai"] = "warning: API key not set"
		if overallStatus == "healthy" {
			overallStatus = "degraded"
		}
	} else {
		checks["openai"] = "ok"
	}

	if os.Getenv("WEATHER_API_KEY") == "" {
		checks["weather_api"] = "warning: API key not set"
		if overallStatus == "healthy" {
			overallStatus = "degraded"
		}
	} else {
		checks["weather_api"] = "ok"
	}

	uptime := time.Since(startTime)
	response := Check{
		Status: overallStatus,
		Checks: checks,
		Uptime: formatDuration(uptime),
	}

	w.Header().Set("Content-Type", "application/json")

	switch overallStatus {
	case "healthy", "degraded":
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(response)
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
