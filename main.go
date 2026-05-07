package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dominickvaske/ev-telemetry/internal/fleet"
)

// FleetStore : Fleet store struct followed by corresponding constructor
type FleetStore struct {
	vehicles map[string]fleet.Vehicle
}

func newFleetStore() *FleetStore {
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

// UpdateVehicle updates a vehicle based on ID with package updates
func (fs *FleetStore) UpdateVehicle(id string, update fleet.VehicleUpdate) error {
	// grab the proper vehicle to update
	val, ok := fs.vehicles[id]

	if !ok {
		return fmt.Errorf("vehicle %s not found: %w", id, ErrVehicleNotFound)
	}

	// perform nil checks on fields of update Vehicle
	var retErr error = nil
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

// ListCharging : Returns slice of all vehicle structs currently charging
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

// SimulateTick runs through all vehicles and drops battery
// percentage by one point to simulate a drive time instance
func (fs *FleetStore) SimulateTick() []Alert {
	alerts := make([]Alert, 0)

	for id, vehicle := range fs.vehicles {
		vehicle.BatteryPct--

		if vehicle.BatteryPct < 10.0 {
			alert := Alert{
				ID:         "A-" + strconv.Itoa(globalAlertID),
				alertVEHID: id,
				Type:       BatteryAlert,
				Value:      vehicle.BatteryPct,
				Message:    "Battery less than 10 percent",
				TimeStamp:  time.Now(),
			}
			globalAlertID++
			alerts = append(alerts, alert)
		}
		err := fs.UpdateBattery(id, vehicle.BatteryPct)
		if err != nil { // vehicle not found
			continue
		}
	}
	return alerts
}

type AlertType string

const (
	TempAlert    AlertType = "TEMPERATURE"
	SpeedAlert   AlertType = "SPEED"
	BatteryAlert AlertType = "BATTERY"
)

// TODO: not thread-safe, replace with atomic or mutex in Phase 2
var globalAlertID = 1

type Alert struct {
	ID         string    // Unique ID for alert event
	alertVEHID string    // ID of the vehicle with the alert
	Type       AlertType // AlertType (string) for reason for alert
	Value      float64   // raw value causing alert
	Message    string    // human-readable error message
	TimeStamp  time.Time // timestamp of alert creation
}

func main() {
	v1 := fleet.Vehicle{ID: "V-001", BatteryPct: 86.0, SpeedKPH: 0.0, TempC: 21.0, IsCharging: true, Timestamp: time.Now()}
	v2 := fleet.Vehicle{ID: "V-002", BatteryPct: 70.0, SpeedKPH: 66.0, TempC: 24.0, IsCharging: false, Timestamp: time.Now()}
	v3 := fleet.Vehicle{ID: "V-003", BatteryPct: 43.0, SpeedKPH: 32.0, TempC: 22.0, IsCharging: false, Timestamp: time.Now()}

	store := newFleetStore()

	store.Add(v1)
	store.Add(v2)
	store.Add(v3)

	for _, vehicle := range store.List() {
		fmt.Println(vehicle)
	}

	// test update battery
	//err := store.UpdateBattery("Bad-ID", 20)
	//if err != nil {
	//	fmt.Println("Update Battery error: ", err)
	//}
	//
	//// test remove function
	//_, err = store.Remove("Bad-ID")
	//if err != nil {
	//	fmt.Println("Remove error: ", err)
	//}

	// test remove function
	//v, err := store.Remove("V-002")
	//if err != nil {
	//	fmt.Println("Remove error: ", err)
	//} else {
	//	fmt.Printf("Removed Vehicle: %s: ", v.ID)
	//	fmt.Println(v)
	//}

	// test list charging function
	//charging := store.ListCharging()
	//fmt.Println("Vehicles charging: ")
	//for _, v := range charging {
	//	fmt.Println(v)
	//}

	//summary := store.Summary()
	//fmt.Println("Number of Vehicles: ", summary.TotalVehicles)
	//fmt.Println("Number Charging: ", summary.ChargingCount)
	//fmt.Printf("Average Battery Percent: %f\n", summary.AvgBatteryPct)
	//fmt.Printf("Average Speed: %f\n", summary.AvgSpeedKPH)

	v4 := fleet.Vehicle{ID: "V-004", BatteryPct: 10.8, SpeedKPH: 32.0, TempC: 22.0, IsCharging: false, Timestamp: time.Now()}

	store.Add(v4)
	alerts := store.SimulateTick()

	for _, a := range alerts {
		fmt.Println(a)
	}

}
