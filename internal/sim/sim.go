package sim

import (
	"strconv"
	"time"

	"github.com/dominickvaske/ev-telemetry/internal/alert"
	"github.com/dominickvaske/ev-telemetry/internal/store"
)

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
