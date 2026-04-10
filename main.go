package main

import (
	"errors"
	"fmt"
	"time"
)

// ErrVehicleNotFound : a descriptive error to pass when a vehicle not found
// in fleet store
var ErrVehicleNotFound = errors.New("vehicle not found")

// FleetStore : Fleet store struct followed by corresponding constructor
type FleetStore struct {
	vehicles map[string]Vehicle
}

func newFleetStore() *FleetStore {
	return &FleetStore{
		vehicles: make(map[string]Vehicle),
	}
}

// Add a vehicle to the Fleet Store
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

// UpdateBattery of a passed in id for a vehicle
func (fs *FleetStore) UpdateBattery(id string, pct float64) error {
	val, ok := fs.vehicles[id]
	if ok {
		val.BatteryPct = pct
		fs.vehicles[id] = val
		return nil
	}
	return ErrVehicleNotFound
}

// UpdateSpeed of provided vehicle
func (fs *FleetStore) UpdateSpeed(id string, speed float64) error {
	val, ok := fs.vehicles[id]
	if ok {
		val.SpeedKPH = speed
		fs.vehicles[id] = val
		return nil
	}
	return ErrVehicleNotFound
}

// UpdateTemp of a vehicle
func (fs *FleetStore) UpdateTemp(id string, temp float64) error {
	val, ok := fs.vehicles[id]
	if ok {
		val.TempC = temp
		fs.vehicles[id] = val
		return nil
	}
	return ErrVehicleNotFound
}

func (fs *FleetStore) Remove(id string) (Vehicle, error) {
	val, ok := fs.vehicles[id]
	if ok {
		delete(fs.vehicles, id)
		return val, nil
	}
	return Vehicle{}, ErrVehicleNotFound
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
		fmt.Println("Update Battery error: ", err)
	}

	_, err = store.Remove("Bad-ID")
	if err != nil {
		fmt.Println("Remove error: ", err)
	}

	v, err := store.Remove("V-002")
	if err != nil {
		fmt.Println("Remove error: ", err)
	} else {
		fmt.Printf("Removed Vehicle: %s: ", v.ID)
		fmt.Println(v)
	}

}
