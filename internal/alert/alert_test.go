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

func TestCheckTemp(t *testing.T) {
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
		BatteryPct: 12.0,
		SpeedKPH:   100.0,
		TempC:      55.0,
		IsCharging: false,
		Timestamp:  time.Now(),
	}

	if a := CheckTemp(prev, curr); a == nil {
		t.Error("Temp alert returned nil on change 45->55 percent")
	}

	prev = curr
	curr = fleet.Vehicle{
		ID:         "V-test",
		BatteryPct: 8.0,
		SpeedKPH:   105.0,
		TempC:      60.0,
		IsCharging: false,
		Timestamp:  time.Now(),
	}
	if a := CheckSpeed(prev, curr); a != nil {
		t.Error("Speed alert fired on change 55->60 percent")
	}
}

func TestCheckSpeed(t *testing.T) {
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
		BatteryPct: 12.0,
		SpeedKPH:   100.0,
		TempC:      45.0,
		IsCharging: false,
		Timestamp:  time.Now(),
	}

	if a := CheckSpeed(prev, curr); a == nil {
		t.Error("Speed alert returned nil on change 75->100 percent")
	}

	prev = curr
	curr = fleet.Vehicle{
		ID:         "V-test",
		BatteryPct: 8.0,
		SpeedKPH:   105.0,
		TempC:      45.0,
		IsCharging: false,
		Timestamp:  time.Now(),
	}
	if a := CheckSpeed(prev, curr); a != nil {
		t.Error("Speed alert fired on change 100->105 percent")
	}
}
