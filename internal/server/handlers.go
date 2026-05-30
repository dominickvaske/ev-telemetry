package server

import (
	"net/http"

	"github.com/dominickvaske/ev-telemetry/internal/fleet"
	"github.com/go-chi/chi/v5"
)

// handleGetFleet for route GET /fleet
func (s *Server) handleGetFleet(w http.ResponseWriter, r *http.Request) {
	// grab current list of vehicles and check if empty
	store := s.store.List(r.Context())
	if store == nil {
		writeJSON(w, http.StatusOK, []fleet.Vehicle{})
		return
	}
	writeJSON(w, http.StatusOK, store)
}

// handleGetVehicle for route GET /vehicle/{id}
func (s *Server) handleGetVehicle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	v, ok := s.store.Get(r.Context(), id)
	if !ok {
		writeError(w, "vehicle not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, v)
}
