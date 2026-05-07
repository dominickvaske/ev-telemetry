package store

import (
	"fmt"
	"time"

	"github.com/dominickvaske/ev-telemetry/internal/fleet"
)

// FleetStore defines struct to hold all vehicles in a fleet
type FleetStore struct {
	vehicles map[string]fleet.Vehicle
}

// NewFleetStore defines the constructor for creating a
// FleetStore struct
func NewFleetStore() *FleetStore {
	return &FleetStore{
		vehicles: make(map[string]fleet.Vehicle),
	}
}

// Add a vehicle to the Fleet Store
func (fs *FleetStore) Add(v fleet.Vehicle) {
	fs.vehicles[v.ID] = v
}

// Get a vehicle from a string
func (fs *FleetStore) Get(id string) (fleet.Vehicle, bool) {
	val, ok := fs.vehicles[id]
	return val, ok
}

// List all vehicles currently in fleet store by returning slice
func (fs *FleetStore) List() []fleet.Vehicle {
	result := make([]fleet.Vehicle, 0)
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
	return fmt.Errorf("vehicle %s not found: %w", id, ErrVehicleNotFound)
}

// UpdateSpeed of provided vehicle
func (fs *FleetStore) UpdateSpeed(id string, speed float64) error {
	val, ok := fs.vehicles[id]
	if ok {
		val.SpeedKPH = speed
		fs.vehicles[id] = val
		return nil
	}
	return fmt.Errorf("vehicle %s not found: %w", id, ErrVehicleNotFound)
}

// UpdateTemp of a vehicle
func (fs *FleetStore) UpdateTemp(id string, temp float64) error {
	val, ok := fs.vehicles[id]
	if ok {
		val.TempC = temp
		fs.vehicles[id] = val
		return nil
	}
	return fmt.Errorf("vehicle %s not found: %w", id, ErrVehicleNotFound)
}

func (fs *FleetStore) UpdateChgState(id string) error {
	val, ok := fs.vehicles[id]
	if ok {
		val.IsCharging = !val.IsCharging
		fs.vehicles[id] = val
		return nil
	}
	return fmt.Errorf("vehicle %s not found: %w", id, ErrVehicleNotFound)
}

// UpdateVehicle updates a vehicle in the store based on ID with package updates
func (fs *FleetStore) UpdateVehicle(id string, update fleet.VehicleUpdate) error {
	// grab the proper vehicle to update
	val, ok := fs.vehicles[id]

	if !ok {
		return fmt.Errorf("vehicle %s not found: %w", id, ErrVehicleNotFound)
	}

	// perform nil checks on fields of update Vehicle
	var retErr error
	updated := false

	if update.NewSpd != nil {
		retErr = fs.UpdateSpeed(id, *update.NewSpd)
		if retErr != nil {
			return retErr
		}
		updated = true
	}

	if update.NewBatPct != nil {
		retErr = fs.UpdateBattery(id, *update.NewBatPct)
		if retErr != nil {
			return retErr
		}
		updated = true
	}

	if update.NewTemp != nil {
		retErr = fs.UpdateTemp(id, *update.NewTemp)
		if retErr != nil {
			return retErr
		}
		updated = true
	}

	if update.NewChgState != nil {
		retErr = fs.UpdateChgState(id)
		if retErr != nil {
			return retErr
		}
		updated = true
	}

	if updated {
		val = fs.vehicles[id]
		val.Timestamp = time.Now()
		fs.vehicles[id] = val
	}

	return retErr
}

// Remove removes a vehicle from the fleet store
// returns the removed vehicle or an error if vehicle not found
func (fs *FleetStore) Remove(id string) (fleet.Vehicle, error) {
	val, ok := fs.vehicles[id]
	if ok {
		delete(fs.vehicles, id)
		return val, nil
	}
	return fleet.Vehicle{}, fmt.Errorf("vehicle %s not found: %w", id, ErrVehicleNotFound)
}

// ListCharging returns slice of all vehicle structs currently charging
func (fs *FleetStore) ListCharging() []fleet.Vehicle {
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
