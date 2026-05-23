package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dominickvaske/ev-telemetry/internal/fleet"
	"github.com/dominickvaske/ev-telemetry/internal/sim"
	"github.com/dominickvaske/ev-telemetry/internal/store"
)

// Ingest is a function that reads from a telemetry channel and then checks
// for a battery alert
func Ingest(fs *store.FleetStore, telemetryCh <-chan fleet.Vehicle, done chan struct{}) {
	for {
		select {
		case v := <-telemetryCh:
			if err := fs.Set(context.Background(), v); err != nil {
				log.Printf("ERR: Vehicle %s not found", v.ID)
			} else if v.BatteryPct < 10.0 {
				log.Printf("ALERT: vehicle %s battery at %.1f%%", v.ID, v.BatteryPct)
			}
		case <-done:
			return
		}
	}
}

func main() {
	v1 := fleet.Vehicle{ID: "V-001", BatteryPct: 86.0, SpeedKPH: 0.0, TempC: 21.0, IsCharging: true, Timestamp: time.Now()}
	v2 := fleet.Vehicle{ID: "V-002", BatteryPct: 70.0, SpeedKPH: 66.0, TempC: 24.0, IsCharging: false, Timestamp: time.Now()}
	v3 := fleet.Vehicle{ID: "V-003", BatteryPct: 43.0, SpeedKPH: 32.0, TempC: 22.0, IsCharging: false, Timestamp: time.Now()}

	fs := store.NewFleetStore()

	fs.Add(context.Background(), v1)
	fs.Add(context.Background(), v2)
	fs.Add(context.Background(), v3)

	for _, vehicle := range fs.List(context.Background()) {
		fmt.Println(vehicle)
	}

	v4 := fleet.Vehicle{ID: "V-004", BatteryPct: 10.8, SpeedKPH: 32.0, TempC: 22.0, IsCharging: false, Timestamp: time.Now()}

	fs.Add(context.Background(), v4)
	//alerts := sim.SimulateTick(fs)

	//for _, a := range alerts {
	//	fmt.Println(a)
	//}

	telemetryCh := make(chan fleet.Vehicle)
	done := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(5)

	// launch goroutines for each vehicle and ingest
	go func() {
		defer wg.Done()
		sim.SimulateVehicle(v1, telemetryCh, done)
	}()
	go func() {
		defer wg.Done()
		sim.SimulateVehicle(v2, telemetryCh, done)
	}()
	go func() {
		defer wg.Done()
		sim.SimulateVehicle(v3, telemetryCh, done)
	}()
	go func() {
		defer wg.Done()
		sim.SimulateVehicle(v4, telemetryCh, done)
	}()

	// launch ingest goroutine to sit and wait for passed in vehicles
	go func() {
		defer wg.Done()
		Ingest(fs, telemetryCh, done)
	}()

	// wait for signal.Notify and then close done if received
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	close(done)
	wg.Wait()
	fmt.Println("\nAll routines closed. Exiting....")
}
