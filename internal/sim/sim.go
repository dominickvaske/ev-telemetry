package sim

import (
	"strconv"
	"time"

	"github.com/dominickvaske/ev-telemetry/internal/alert"
	"github.com/dominickvaske/ev-telemetry/internal/fleet"
	"github.com/dominickvaske/ev-telemetry/internal/store"
)

// SimulateVehicle runs simulation for a single vehicle v
// v: fleet.Vehicle being simulated on specific goroutine
// telemetryCh: channel shared by ingest function to know when to update
// done: channel shared with main to receive done signal
func SimulateVehicle(v fleet.Vehicle, telemetryCh chan<- fleet.Vehicle, done chan struct{}) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			v.BatteryPct--
			v.Timestamp = time.Now()
			telemetryCh <- v
		case <-done:
			return
		}
	}
}

// SimulateTick runs through all vehicles and drops battery
// percentage by one point to simulate a drive time instance
func SimulateTick(fs *store.FleetStore) []alert.Alert {
	alerts := make([]alert.Alert, 0)

	for _, vehicle := range fs.List() {
		vehicle.BatteryPct--
		id := vehicle.ID
		if vehicle.BatteryPct < 10.0 {
			newAlert := alert.Alert{
				ID:         "A-" + strconv.Itoa(alert.GlobalAlertID),
				AlertVEHID: id,
				Type:       alert.BatteryAlert,
				Value:      vehicle.BatteryPct,
				Message:    "Battery less than 10 percent",
				TimeStamp:  time.Now(),
			}
			alert.GlobalAlertID++
			alerts = append(alerts, newAlert)
		}
		err := fs.UpdateBattery(id, vehicle.BatteryPct)
		if err != nil { // vehicle not found
			continue
		}
	}
	return alerts
}
