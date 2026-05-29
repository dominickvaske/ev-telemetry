package server

import (
	"net/http"

	"github.com/dominickvaske/ev-telemetry/internal/fleet"
)

// handleGetFleet for route GET /fleet
func (s *Server) handleGetFleet(w http.ResponseWriter, r *http.Request) {
	store := s.store.List(r.Context())
	if store == nil {
		writeJSON(w, http.StatusOK, []fleet.Vehicle{})
		return
	}
	writeJSON(w, http.StatusOK, store)
}
