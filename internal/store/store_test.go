package store

import (
	"context"
	"testing"
	"time"

	"github.com/dominickvaske/ev-telemetry/internal/fleet"
)

func testVehicle(id string) fleet.Vehicle {
	return fleet.Vehicle{
		ID:         id,
		BatteryPct: 80.0,
		SpeedKPH:   60.0,
		TempC:      22.0,
		IsCharging: false,
		Timestamp:  time.Now(),
	}
}

func TestAdd(t *testing.T) {
	fs := NewFleetStore()
	v := testVehicle("V-001")
	fs.Add(context.Background(), v)

	got, ok := fs.Get(context.Background(), "V-001")
	if !ok {
		t.Fatal("vehicle not found after Add")
	}
	if got.ID != v.ID {
		t.Errorf("got ID %s, want %s", got.ID, v.ID)
	}
}

func TestGet_NotFound(t *testing.T) {
	fs := NewFleetStore()
	_, ok := fs.Get(context.Background(), "missing")
	if ok {
		t.Error("expected not found, got ok=true")
	}
}

func TestSet(t *testing.T) {
	fs := NewFleetStore()
	fs.Add(context.Background(), testVehicle("V-001"))

	updated := testVehicle("V-001")
	updated.BatteryPct = 50.0
	if err := fs.Set(context.Background(), updated); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := fs.Get(context.Background(), "V-001")
	if got.BatteryPct != 50.0 {
		t.Errorf("got battery %v, want 50.0", got.BatteryPct)
	}
}

func TestSet_NotFound(t *testing.T) {
	fs := NewFleetStore()
	err := fs.Set(context.Background(), testVehicle("missing"))
	if err == nil {
		t.Error("expected error for missing vehicle, got nil")
	}
}

func TestList(t *testing.T) {
	fs := NewFleetStore()
	fs.Add(context.Background(), testVehicle("V-001"))
	fs.Add(context.Background(), testVehicle("V-002"))

	vehicles := fs.List(context.Background())
	if len(vehicles) != 2 {
		t.Errorf("got %d vehicles, want 2", len(vehicles))
	}
}