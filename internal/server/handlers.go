package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dominickvaske/ev-telemetry/internal/alert"
	"github.com/dominickvaske/ev-telemetry/internal/fleet"
	"github.com/dominickvaske/ev-telemetry/internal/store"
	"github.com/go-chi/chi/v5"
)

// handleGetFleet for route GET /fleet
func (s *Server) handleGetFleet(w http.ResponseWriter, r *http.Request) {
	// grab current list of vehicles and check if empty
	vehicles := s.store.List(r.Context())
	if vehicles == nil {
		writeJSON(w, http.StatusOK, []fleet.Vehicle{})
		return
	}
	writeJSON(w, http.StatusOK, vehicles)
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

// handleGetAlerts for route GET /alerts
func (s *Server) handleGetAlerts(w http.ResponseWriter, r *http.Request) {
	alerts := s.alerts.All()
	if alerts == nil {
		writeJSON(w, http.StatusOK, []alert.Alert{})
		return
	}
	writeJSON(w, http.StatusOK, alerts)
}

// handlePostTelemetry for route POST /telemetry to update a vehicle
func (s *Server) handlePostTelemetry(w http.ResponseWriter, r *http.Request) {
	var vehicle fleet.Vehicle

	// decode the request body into a vehicle struct
	if err := json.NewDecoder(r.Body).Decode(&vehicle); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// check if vehicle exists in store
	if vehicle.ID == "" {
		writeError(w, "id is required", http.StatusBadRequest)
		return
	}

	// set the vehicle in the store
	if err := s.store.Set(r.Context(), vehicle); err != nil {
		if errors.Is(err, store.ErrVehicleNotFound) {
			writeError(w, "vehicle not found", http.StatusNotFound)
		} else {
			writeError(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, vehicle)
}
