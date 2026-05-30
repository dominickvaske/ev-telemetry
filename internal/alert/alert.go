package alert

import (
	"sync"
	"sync/atomic"
	"time"
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
