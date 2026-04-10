package main

import (
	"errors"
	"fmt"
	"time"
)

// Initial package scope variables
var ErrVehicleNotFound = errors.New("vehicle not found")

type FleetStore struct {
	vehicles map[string]Vehicle
}

func newFleetStore() *FleetStore {
	return &FleetStore{
		vehicles: make(map[string]Vehicle),
	}
}

// Add a vehicle to the fleetstore
func (fs *FleetStore) Add(v Vehicle) {
	fs.vehicles[v.ID] = v
}

// Get a vehicle from a string
func (fs *FleetStore) Get(id string) (Vehicle, bool) {
	val, ok := fs.vehicles[id]
	return val, ok
}

// List all vehicles currently in fleet store by returning slice
func (fs *FleetStore) List() []Vehicle {
	var result []Vehicle
	for _, vehicle := range fs.vehicles {
		result = append(result, vehicle)
	}
	return result
}

func (fs *FleetStore) UpdateBattery(id string, pct float64) error {
	val, ok := fs.vehicles[id]
	if ok {
		val.BatteryPct = pct
		return nil
	}
	return ErrVehicleNotFound
}

type Vehicle struct {
	ID         string
	BatteryPct float64
	SpeedKPH   float64
	TempC      float64
	IsCharging bool
	Timestamp  time.Time
}

func main() {
	v1 := Vehicle{ID: "V-001", BatteryPct: 86.0, SpeedKPH: 0.0, TempC: 21.0, IsCharging: true, Timestamp: time.Now()}
	v2 := Vehicle{ID: "V-002", BatteryPct: 70.0, SpeedKPH: 66.0, TempC: 24.0, IsCharging: false, Timestamp: time.Now()}
	v3 := Vehicle{ID: "V-003", BatteryPct: 43.0, SpeedKPH: 32.0, TempC: 22.0, IsCharging: false, Timestamp: time.Now()}

	store := newFleetStore()

	store.Add(v1)
	store.Add(v2)
	store.Add(v3)

	for _, vehicle := range store.List() {
		fmt.Println(vehicle)
	}

	err := store.UpdateBattery("Bad-ID", 20)
	if err != nil {
		fmt.Println("error: ", err)
	}
}
