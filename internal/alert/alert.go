package alert

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dominickvaske/ev-telemetry/internal/fleet"
)

type AlertType string

const (
	SpeedAlert   AlertType = "SPEED"
	BatteryAlert AlertType = "BATTERY"
	TempAlert    AlertType = "TEMPERATURE"
)

// GlobalAlertID is concurrent safe
var GlobalAlertID atomic.Int32

type Alert struct {
	ID        string    `json:"id"`         // Unique ID for alert event
	VehicleID string    `json:"vehicle_id"` // ID of the vehicle with the alert
	Type      AlertType `json:"type"`       // AlertType (string) for reason for alert
	Value     float64   `json:"value"`      // raw value causing alert
	Message   string    `json:"message"`    // human-readable error message
	TimeStamp time.Time `json:"timestamp"`  // timestamp of alert creation
}

type Log struct {
	lock   sync.RWMutex
	alerts []Alert
}

func NewAlertLog() *Log {
	return &Log{
		lock:   sync.RWMutex{},
		alerts: make([]Alert, 0),
	}
}

// Append alert to the overall log
func (l *Log) Append(a Alert) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.alerts = append(l.alerts, a)
}

// All returns a copy of the current slice of alerts from log
func (l *Log) All() []Alert {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return append(make([]Alert, 0), l.alerts...)
}

// CheckBattery checks if a vehicle meets alert criteria for battery alert
func CheckBattery(prev, curr fleet.Vehicle) *Alert {
	if prev.BatteryPct >= 10.0 && curr.BatteryPct < 10 {
		id := int(GlobalAlertID.Add(1))
		return &Alert{
			ID:        "A-" + strconv.Itoa(id),
			VehicleID: curr.ID,
			Type:      BatteryAlert,
			Value:     curr.BatteryPct,
			Message:   "Battery less than 10 percent",
			TimeStamp: time.Now(),
		}
	}
	return nil
}

func CheckTemp(prev, curr fleet.Vehicle) *Alert {
	if prev.TempC < 50.0 && curr.TempC >= 50.0 {
		id := int(GlobalAlertID.Add(1))
		return &Alert{
			ID:        "A-" + strconv.Itoa(id),
			VehicleID: curr.ID,
			Type:      TempAlert,
			Value:     curr.TempC,
			Message:   "Vehicle Temperature greater than 50C",
			TimeStamp: time.Now(),
		}
	}
	return nil
}

func CheckSpeed(prev, curr fleet.Vehicle) *Alert {
	if prev.SpeedKPH < 100.0 && curr.SpeedKPH >= 100.0 {
		id := int(GlobalAlertID.Add(1))
		return &Alert{
			ID:        "A-" + strconv.Itoa(id),
			VehicleID: curr.ID,
			Type:      SpeedAlert,
			Value:     curr.SpeedKPH,
			Message:   "Vehicle Speed greater than 100KPH",
			TimeStamp: time.Now(),
		}
	}
	return nil
}
