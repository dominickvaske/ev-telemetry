package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dominickvaske/ev-telemetry/internal/alert"
	"github.com/dominickvaske/ev-telemetry/internal/fleet"
	"github.com/dominickvaske/ev-telemetry/internal/store"
)

func testServer() *Server {
	return NewServer(store.NewFleetStore(), alert.NewAlertLog())
}

func seedVehicle(s *Server, id string) fleet.Vehicle {
	v := fleet.Vehicle{
		ID:         id,
		BatteryPct: 80.0,
		SpeedKPH:   60.0,
		TempC:      22.0,
		IsCharging: false,
		Timestamp:  time.Now(),
	}
	s.store.Add(nil, v)
	return v
}

func TestHandleGetFleet_Empty(t *testing.T) {
	srv := testServer()
	req := httptest.NewRequest(http.MethodGet, "/fleet", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleGetFleet_WithVehicles(t *testing.T) {
	srv := testServer()
	seedVehicle(srv, "V-001")
	seedVehicle(srv, "V-002")

	req := httptest.NewRequest(http.MethodGet, "/fleet", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	var vehicles []fleet.Vehicle
	if err := json.NewDecoder(w.Body).Decode(&vehicles); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if len(vehicles) != 2 {
		t.Errorf("got %d vehicles, want 2", len(vehicles))
	}
}

func TestHandleGetVehicle_Found(t *testing.T) {
	srv := testServer()
	seedVehicle(srv, "V-001")

	req := httptest.NewRequest(http.MethodGet, "/vehicle/V-001", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleGetVehicle_NotFound(t *testing.T) {
	srv := testServer()

	req := httptest.NewRequest(http.MethodGet, "/vehicle/missing", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandlePostTelemetry_Valid(t *testing.T) {
	srv := testServer()
	seedVehicle(srv, "V-001")

	update := fleet.Vehicle{ID: "V-001", BatteryPct: 50.0, SpeedKPH: 80.0, TempC: 25.0, Timestamp: time.Now()}
	body, _ := json.Marshal(update)

	req := httptest.NewRequest(http.MethodPost, "/telemetry", bytes.NewReader(body))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandlePostTelemetry_MissingID(t *testing.T) {
	srv := testServer()

	body, _ := json.Marshal(fleet.Vehicle{BatteryPct: 50.0})
	req := httptest.NewRequest(http.MethodPost, "/telemetry", bytes.NewReader(body))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleGetAlerts_Empty(t *testing.T) {
	srv := testServer()

	req := httptest.NewRequest(http.MethodGet, "/alerts", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
	}
}
