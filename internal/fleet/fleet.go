// Package fleet is a package that contains all non-dependent
// types for managing the fleet
package fleet

import "time"

// Vehicle is a struct to simulate all vehicles
// in the fleet
type Vehicle struct {
	ID         string
	BatteryPct float64
	SpeedKPH   float64
	TempC      float64
	IsCharging bool
	Timestamp  time.Time
}

// VehicleUpdate is a type that includes all information
// relating to a possible vehicle update
type VehicleUpdate struct {
	NewTemp     *float64
	NewSpd      *float64
	NewBatPct   *float64
	NewChgState *bool
	UpdateTime  *time.Time
}

// Summary is a type used to summarize
// all information relating to a fleet
type Summary struct {
	TotalVehicles int
	ChargingCount int
	AvgBatteryPct float64
	AvgSpeedKPH   float64
}
