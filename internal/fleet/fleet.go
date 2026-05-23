// Package fleet is a package that contains all non-dependent
// types for managing the fleet
package fleet

import "time"

// Vehicle is a struct to simulate all vehicles
// in the fleet
type Vehicle struct {
	ID         string    `json:"id"`
	BatteryPct float64   `json:"battery_pct"`
	SpeedKPH   float64   `json:"speed_kph"`
	TempC      float64   `json:"temp_c"`
	IsCharging bool      `json:"is_charging"`
	Timestamp  time.Time `json:"timestamp"`
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
	TotalVehicles int     `json:"total_vehicles"`
	ChargingCount int     `json:"charging_count"`
	AvgBatteryPct float64 `json:"avg_battery_pct"`
	AvgSpeedKPH   float64 `json:"avg_speed_kph"`
}
