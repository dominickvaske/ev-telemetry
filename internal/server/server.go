package server

import (
	"encoding/json"
	"net/http"

	"github.com/dominickvaske/ev-telemetry/internal/alert"
	"github.com/dominickvaske/ev-telemetry/internal/store"
	"github.com/go-chi/chi/v5"
)

// Server is a struct that the main.go package can use to
// create a server for the eventual frontend to communicate with
type Server struct {
	store  *store.FleetStore
	alerts *alert.Log
	router *chi.Mux
}

func NewServer(s *store.FleetStore, a *alert.Log) *Server {
	srv := Server{
		store:  s,
		alerts: a,
		router: chi.NewRouter(),
	}

	srv.router.Get("/fleet", srv.handleGetFleet)
	return &srv
}

// ServeHTTP allows Server to meet Handler interface spec
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// writeJSON serves as a utility function to remove repetition of initial response to client
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError allows for consistent error writing by creating a struct of {"error": "msg"}
func writeError(w http.ResponseWriter, msg string, code int) {
	writeJSON(w, code, struct {
		Error string `json:"error"`
	}{msg})
}
