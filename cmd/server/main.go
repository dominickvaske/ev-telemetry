package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dominickvaske/ev-telemetry/internal/alert"
	"github.com/dominickvaske/ev-telemetry/internal/fleet"
	"github.com/dominickvaske/ev-telemetry/internal/metrics"
	"github.com/dominickvaske/ev-telemetry/internal/server"
	"github.com/dominickvaske/ev-telemetry/internal/sim"
	"github.com/dominickvaske/ev-telemetry/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// Ingest is a function that reads from a telemetry channel and then checks
// for a possible alert
func Ingest(fs *store.FleetStore,
	telemetryCh <-chan fleet.Vehicle,
	done chan struct{},
	alertLog *alert.Log,
	pool *pgxpool.Pool) {
	for {
		select {
		case v := <-telemetryCh:
			oldV, _ := fs.Get(context.Background(), v.ID)
			if err := fs.Set(context.Background(), v); err != nil {
				log.Printf("ERR: Vehicle %s not found", v.ID)
			}
			// INSERT into database telemetry_events table
			query := `INSERT INTO telemetry_events (vehicle_id, battery_pct, speed_kph, temp_c, is_charging, timestamp) 
					  VALUES ($1, $2, $3, $4, $5, $6)`
			start := time.Now()
			_, err := pool.Exec(context.Background(), query, v.ID, v.BatteryPct, v.SpeedKPH, v.TempC, v.IsCharging, v.Timestamp)
			duration := time.Since(start).Seconds()
			if err != nil {
				log.Printf("Query execution failed: %v\n", err)
			}

			// update matrics tracking for latency
			metrics.LatencyHist.Observe(duration)
			// update counter metric
			metrics.EventsCounter.Inc()

			if a := alert.CheckBattery(oldV, v); a != nil {
				alertLog.Append(*a)
			}
			if a := alert.CheckSpeed(oldV, v); a != nil {
				alertLog.Append(*a)
			}
			if a := alert.CheckTemp(oldV, v); a != nil {
				alertLog.Append(*a)
			}
		case <-done:
			return
		}
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	alertLog := alert.NewAlertLog()

	port := flag.String("port", "8080", "HTTP server port")
	flag.Parse()

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
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Error opening pgx pool")
	}
	go func() {
		defer wg.Done()
		Ingest(fs, telemetryCh, done, alertLog, pool)
	}()

	// start a server in its own go routine
	// not in wait group since it doesn't depend on the done channel
	srv := server.NewServer(fs, alertLog)
	go func() {
		log.Printf("listening on : %s", *port)
		if err := http.ListenAndServe(":"+*port, srv); err != nil {
			log.Fatal(err)
		}
	}()

	// wait for signal.Notify and then close done if received
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	close(done)
	wg.Wait()
	fmt.Println("\nAll routines closed. Exiting....")
}
