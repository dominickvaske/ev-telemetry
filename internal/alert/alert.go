package alert

import (
	"sync"
	"time"
)

type AlertType string

const (
	SpeedAlert   AlertType = "SPEED"
	BatteryAlert AlertType = "BATTERY"
	TempAlert    AlertType = "TEMPERATURE"
)

// TODO: not thread-safe, replace with atomic or mutex in Phase 2
var GlobalAlertID = 1

type Alert struct {
	ID         string    // Unique ID for alert event
	AlertVEHID string    // ID of the vehicle with the alert
	Type       AlertType // AlertType (string) for reason for alert
	Value      float64   // raw value causing alert
	Message    string    // human-readable error message
	TimeStamp  time.Time // timestamp of alert creation
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
