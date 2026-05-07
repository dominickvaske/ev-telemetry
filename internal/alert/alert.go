package alert

import (
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
