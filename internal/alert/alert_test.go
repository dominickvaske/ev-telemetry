package alert

import (
	"testing"
	"time"

	"github.com/dominickvaske/ev-telemetry/internal/fleet"
)

func TestCheckBattery(t *testing.T) {
	prev := fleet.Vehicle{
		ID:         "V-test",
		BatteryPct: 11.0,
		SpeedKPH:   75.0,
		TempC:      45.0,
		IsCharging: false,
		Timestamp:  time.Now(),
	}

	curr := fleet.Vehicle{
		ID:         "V-test",
		BatteryPct: 9.0,
		SpeedKPH:   75.0,
		TempC:      45.0,
		IsCharging: false,
		Timestamp:  time.Now(),
	}

	if a := CheckBattery(prev, curr); a == nil {
		t.Error("Battery alert returned nil on change 11->9 percent")
	}

	prev = curr
	curr = fleet.Vehicle{
		ID:         "V-test",
		BatteryPct: 8.0,
		SpeedKPH:   75.0,
		TempC:      45.0,
		IsCharging: false,
		Timestamp:  time.Now(),
	}
	if a := CheckBattery(prev, curr); a != nil {
		t.Error("Battery alert fired on change 9->8 percent")
	}
}
