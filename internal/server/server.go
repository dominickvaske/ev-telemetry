package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/dominickvaske/ev-telemetry/internal/alert"
	"github.com/dominickvaske/ev-telemetry/internal/metrics"
	"github.com/dominickvaske/ev-telemetry/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	store     *store.FleetStore
	alerts    *alert.Log
	router    *chi.Mux
	startTime time.Time
	port      string
}

func NewServer(s *store.FleetStore, a *alert.Log, port string) *Server {
	srv := Server{
		store:     s,
		alerts:    a,
		router:    chi.NewRouter(),
		startTime: time.Now(),
		port:      port,
	}

	srv.router.Get("/fleet", srv.handleGetFleet)
	srv.router.Get("/vehicle/{id}", srv.handleGetVehicle)
	srv.router.Get("/alerts", srv.handleGetAlerts)
	srv.router.Post("/telemetry", srv.handlePostTelemetry)
	srv.router.Handle("/metrics", promhttp.Handler())
	srv.router.Get("/status", srv.handleGetStatus)

	fileServer := http.FileServer(http.Dir("static"))
	srv.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})
	srv.router.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	return &srv
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	elapsed := time.Since(s.startTime)
	h := int(elapsed.Hours())
	m := int(elapsed.Minutes()) % 60
	sec := int(elapsed.Seconds()) % 60

	writeJSON(w, http.StatusOK, map[string]any{
		"uptime":       fmt.Sprintf("%02d:%02d:%02d", h, m, sec),
		"events_total": metrics.TotalEvents.Load(),
		"goroutines":   runtime.NumGoroutine(),
		"port":         s.port,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, msg string, code int) {
	writeJSON(w, code, struct {
		Error string `json:"error"`
	}{msg})
}
