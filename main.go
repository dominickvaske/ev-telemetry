package main

import (
	"fmt"
	"time"
)

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

	vehicles := []Vehicle{v1, v2, v3}

	for _, vehicle := range vehicles {
		fmt.Println(vehicle)
	}
}
