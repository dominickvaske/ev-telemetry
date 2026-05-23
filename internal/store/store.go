package store

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dominickvaske/ev-telemetry/internal/fleet"
)

// FleetStore defines struct to hold all vehicles in a fleet
type FleetStore struct {
	vehicles   map[string]fleet.Vehicle
	updateLock sync.RWMutex
}

// NewFleetStore defines the constructor for creating a
// FleetStore struct
func NewFleetStore() *FleetStore {
	return &FleetStore{
		vehicles: make(map[string]fleet.Vehicle),
	}
}

// Add a vehicle to the Fleet Store
func (fs *FleetStore) Add(ctx context.Context, v fleet.Vehicle) {
	fs.updateLock.Lock()
	defer fs.updateLock.Unlock()
	fs.vehicles[v.ID] = v
}

// Get a vehicle from a string
func (fs *FleetStore) Get(ctx context.Context, id string) (fleet.Vehicle, bool) {
	fs.updateLock.RLock()
	defer fs.updateLock.RUnlock()
	val, ok := fs.vehicles[id]
	return val, ok
}

// Set a vehicle by updating the current value assuming it exists
func (fs *FleetStore) Set(ctx context.Context, v fleet.Vehicle) error {
	fs.updateLock.Lock()
	defer fs.updateLock.Unlock()
	if _, ok := fs.vehicles[v.ID]; !ok {
		return ErrVehicleNotFound
	}
	fs.vehicles[v.ID] = v
	return nil
}

// List all vehicles currently in fleet store by returning slice
func (fs *FleetStore) List(ctx context.Context) []fleet.Vehicle {
	fs.updateLock.RLock()
	defer fs.updateLock.RUnlock()
	result := make([]fleet.Vehicle, 0)
	for _, vehicle := range fs.vehicles {
		result = append(result, vehicle)
	}
	return result
}

// updateBattery is unexported version for reference
// handles mutex locking
func (fs *FleetStore) updateBattery(id string, pct float64) error {
	val, ok := fs.vehicles[id]
	if ok {
		val.BatteryPct = pct
		val.Timestamp = time.Now()
		fs.vehicles[id] = val
		return nil
	}
	return fmt.Errorf("vehicle %s not found: %w", id, ErrVehicleNotFound)
}

// UpdateBattery exported version of updateBattery with mutex locking
func (fs *FleetStore) UpdateBattery(id string, pct float64) error {
	fs.updateLock.Lock()
	defer fs.updateLock.Unlock()
	return fs.updateBattery(id, pct)
}

// updateSpeed unexported version for mutex locking
func (fs *FleetStore) updateSpeed(id string, speed float64) error {
	val, ok := fs.vehicles[id]
	if ok {
		val.SpeedKPH = speed
		val.Timestamp = time.Now()
		fs.vehicles[id] = val
		return nil
	}
	return fmt.Errorf("vehicle %s not found: %w", id, ErrVehicleNotFound)
}

// UpdateSpeed of provided vehicle exported version
func (fs *FleetStore) UpdateSpeed(id string, speed float64) error {
	fs.updateLock.Lock()
	defer fs.updateLock.Unlock()
	return fs.updateSpeed(id, speed)
}

// updateTemp of a vehicle unexported version
func (fs *FleetStore) updateTemp(id string, temp float64) error {
	val, ok := fs.vehicles[id]
	if ok {
		val.TempC = temp
		val.Timestamp = time.Now()
		fs.vehicles[id] = val
		return nil
	}
	return fmt.Errorf("vehicle %s not found: %w", id, ErrVehicleNotFound)
}

// UpdateTemp is the exportable version with locking control
func (fs *FleetStore) UpdateTemp(id string, temp float64) error {
	fs.updateLock.Lock()
	defer fs.updateLock.Unlock()
	return fs.updateTemp(id, temp)
}

// updateChgState takes in a vehicle id and updates whether it's charging or not
func (fs *FleetStore) updateChgState(id string) error {
	val, ok := fs.vehicles[id]
	if ok {
		val.IsCharging = !val.IsCharging
		val.Timestamp = time.Now()
		fs.vehicles[id] = val
		return nil
	}
	return fmt.Errorf("vehicle %s not found: %w", id, ErrVehicleNotFound)
}

// UpdateChgState is the exportable, concurrent safe function
func (fs *FleetStore) UpdateChgState(id string) error {
	fs.updateLock.Lock()
	defer fs.updateLock.Unlock()
	return fs.updateChgState(id)
}

// UpdateVehicle updates a vehicle in the store based on ID with package updates
func (fs *FleetStore) UpdateVehicle(id string, update fleet.VehicleUpdate) error {
	// install write lock
	fs.updateLock.Lock()
	defer fs.updateLock.Unlock()

	// grab the proper vehicle to update
	_, ok := fs.vehicles[id]
	if !ok {
		return fmt.Errorf("vehicle %s not found: %w", id, ErrVehicleNotFound)
	}

	// perform nil checks on fields of update Vehicle
	var retErr error

	if update.NewSpd != nil {
		retErr = fs.updateSpeed(id, *update.NewSpd)
		if retErr != nil {
			return retErr
		}
	}

	if update.NewBatPct != nil {
		retErr = fs.updateBattery(id, *update.NewBatPct)
		if retErr != nil {
			return retErr
		}
	}

	if update.NewTemp != nil {
		retErr = fs.updateTemp(id, *update.NewTemp)
		if retErr != nil {
			return retErr
		}
	}

	if update.NewChgState != nil {
		retErr = fs.updateChgState(id)
		if retErr != nil {
			return retErr
		}
	}

	return retErr
}

// Remove removes a vehicle from the fleet store
// returns the removed vehicle or an error if vehicle not found
func (fs *FleetStore) Remove(id string) (fleet.Vehicle, error) {
	// use write lock
	fs.updateLock.Lock()
	defer fs.updateLock.Unlock()

	val, ok := fs.vehicles[id]
	if ok {
		delete(fs.vehicles, id)
		return val, nil
	}
	return fleet.Vehicle{}, fmt.Errorf("vehicle %s not found: %w", id, ErrVehicleNotFound)
}

// ListCharging returns slice of all vehicle structs currently charging
func (fs *FleetStore) ListCharging() []fleet.Vehicle {
	// use a read lock
	fs.updateLock.RLock()
	defer fs.updateLock.RUnlock()

	result := make([]fleet.Vehicle, 0)
	for _, vehicle := range fs.vehicles {
		if vehicle.IsCharging {
			result = append(result, vehicle)
		}
	}
	return result
}

// Summary provides a Summary struct of the fleet store
func (fs *FleetStore) Summary() fleet.Summary {
	// use read lock
	fs.updateLock.RLock()
	defer fs.updateLock.RUnlock()

	var chargingCount, totalVehicles int
	var avgBatteryPct, avgSpeedKPH float64

	for _, v := range fs.vehicles {
		totalVehicles++
		if v.IsCharging {
			chargingCount++
		}
		avgBatteryPct += v.BatteryPct
		avgSpeedKPH += v.SpeedKPH

	}
	if totalVehicles == 0 {
		return fleet.Summary{}
	}
	return fleet.Summary{
		TotalVehicles: totalVehicles,
		ChargingCount: chargingCount,
		AvgBatteryPct: avgBatteryPct / float64(totalVehicles),
		AvgSpeedKPH:   avgSpeedKPH / float64(totalVehicles),
	}
}
