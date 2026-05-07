package main

import (
	"fmt"
	"time"

	"github.com/dominickvaske/ev-telemetry/internal/fleet"
	"github.com/dominickvaske/ev-telemetry/internal/sim"
	"github.com/dominickvaske/ev-telemetry/internal/store"
)

func main() {
	v1 := fleet.Vehicle{ID: "V-001", BatteryPct: 86.0, SpeedKPH: 0.0, TempC: 21.0, IsCharging: true, Timestamp: time.Now()}
	v2 := fleet.Vehicle{ID: "V-002", BatteryPct: 70.0, SpeedKPH: 66.0, TempC: 24.0, IsCharging: false, Timestamp: time.Now()}
	v3 := fleet.Vehicle{ID: "V-003", BatteryPct: 43.0, SpeedKPH: 32.0, TempC: 22.0, IsCharging: false, Timestamp: time.Now()}

	fs := store.NewFleetStore()

	fs.Add(v1)
	fs.Add(v2)
	fs.Add(v3)

	for _, vehicle := range fs.List() {
		fmt.Println(vehicle)
	}

	// test update battery
	//err := store.UpdateBattery("Bad-ID", 20)
	//if err != nil {
	//	fmt.Println("Update Battery error: ", err)
	//}
	//
	//// test remove function
	//_, err = store.Remove("Bad-ID")
	//if err != nil {
	//	fmt.Println("Remove error: ", err)
	//}

	// test remove function
	//v, err := store.Remove("V-002")
	//if err != nil {
	//	fmt.Println("Remove error: ", err)
	//} else {
	//	fmt.Printf("Removed Vehicle: %s: ", v.ID)
	//	fmt.Println(v)
	//}

	// test list charging function
	//charging := store.ListCharging()
	//fmt.Println("Vehicles charging: ")
	//for _, v := range charging {
	//	fmt.Println(v)
	//}

	//summary := store.Summary()
	//fmt.Println("Number of Vehicles: ", summary.TotalVehicles)
	//fmt.Println("Number Charging: ", summary.ChargingCount)
	//fmt.Printf("Average Battery Percent: %f\n", summary.AvgBatteryPct)
	//fmt.Printf("Average Speed: %f\n", summary.AvgSpeedKPH)

	v4 := fleet.Vehicle{ID: "V-004", BatteryPct: 10.8, SpeedKPH: 32.0, TempC: 22.0, IsCharging: false, Timestamp: time.Now()}

	fs.Add(v4)
	alerts := sim.SimulateTick(fs)

	for _, a := range alerts {
		fmt.Println(a)
	}

}
