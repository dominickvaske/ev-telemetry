package sim

import (
	"math"
	"math/rand"
	"time"

	"github.com/dominickvaske/ev-telemetry/internal/fleet"
)

const ambientTempC = 22.0

// SimulateVehicle runs a per-vehicle simulation loop, sending state updates
// every second until the done channel is closed.
func SimulateVehicle(v fleet.Vehicle, telemetryCh chan<- fleet.Vehicle, done chan struct{}) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			v = tick(v)
			v.Timestamp = time.Now()
			telemetryCh <- v
		case <-done:
			return
		}
	}
}

// tick advances vehicle state by one second.
func tick(v fleet.Vehicle) fleet.Vehicle {
	if v.IsCharging {
		// Charging: recover battery at 1.5% per second, not moving.
		// Stop charging once battery reaches 80%.
		v.BatteryPct = math.Min(v.BatteryPct+1.5, 100.0)
		v.SpeedKPH = 0
		if v.BatteryPct >= 80.0 {
			v.IsCharging = false
		}
	} else {
		// Driving: drain battery based on speed. Faster driving = faster drain.
		drain := 0.3 + v.SpeedKPH/300.0
		v.BatteryPct = math.Max(v.BatteryPct-drain, 0)

		if v.BatteryPct == 0 {
			// Dead battery: pull over and start charging.
			v.SpeedKPH = 0
			v.IsCharging = true
		} else {
			// Vary speed by up to ±5 kph each tick, clamped to 0–120.
			delta := float64(rand.Intn(11) - 5)
			v.SpeedKPH = math.Max(0, math.Min(120, v.SpeedKPH+delta))
		}
	}

	// Temperature drifts toward a target driven by speed, falls toward ambient when slow.
	targetTemp := ambientTempC + v.SpeedKPH*0.15
	v.TempC += (targetTemp - v.TempC) * 0.1

	// 2% chance per tick of a motor stress spike (+5°C).
	if rand.Float64() < 0.02 {
		v.TempC += 5.0
	}

	return v
}
