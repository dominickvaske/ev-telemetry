package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"math/rand"
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

// saveAlert saves a passed in alert to the database referenced by the passed in pool
func saveAlert(ctx context.Context, pool *pgxpool.Pool, a alert.Alert) {
	query := `INSERT INTO alerts (vehicle_id, alert_type, value, message, timestamp)
              VALUES ($1, $2, $3, $4, $5)`
	_, err := pool.Exec(ctx, query, a.VehicleID, a.Type, a.Value, a.Message, a.TimeStamp)
	if err != nil {
		slog.Error("failed to save alert", "err", err, "vehicle_id", a.VehicleID)
	}
}

func checkAlerts(pool *pgxpool.Pool, alertLog *alert.Log, oldV fleet.Vehicle, v fleet.Vehicle) {
	if a := alert.CheckBattery(oldV, v); a != nil {
		alertLog.Append(*a)
		saveAlert(context.Background(), pool, *a)
	}
	if a := alert.CheckSpeed(oldV, v); a != nil {
		alertLog.Append(*a)
		saveAlert(context.Background(), pool, *a)
	}
	if a := alert.CheckTemp(oldV, v); a != nil {
		alertLog.Append(*a)
		saveAlert(context.Background(), pool, *a)
	}
}

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
				slog.Error("vehicle not found", "vehicle_id", v.ID)
			}
			// INSERT into database telemetry_events table
			query := `INSERT INTO telemetry_events (vehicle_id, battery_pct, speed_kph, temp_c, is_charging, timestamp) 
					  VALUES ($1, $2, $3, $4, $5, $6)`
			start := time.Now()
			_, err := pool.Exec(context.Background(), query, v.ID, v.BatteryPct, v.SpeedKPH, v.TempC, v.IsCharging, v.Timestamp)
			duration := time.Since(start).Seconds()
			if err != nil {
				slog.Error("query execution failed", "err", err)
			}

			// update matrics tracking for latency
			metrics.LatencyHist.Observe(duration)
			// update counter metric
			metrics.EventsCounter.Inc()
			metrics.TotalEvents.Add(1)

			// check possible alerts and log
			checkAlerts(pool, alertLog, oldV, v)
		case <-done:
			return
		}
	}
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := godotenv.Load(); err != nil {
		slog.Error("error loading .env file", "err", err)
		os.Exit(1)
	}

	alertLog := alert.NewAlertLog()

	port := flag.String("port", "8080", "HTTP server port")
	n := flag.Int("vehicles", 4, "number of vehicles to simulate")
	flag.Parse()

	fs := store.NewFleetStore()
	vehicles := make([]fleet.Vehicle, *n)
	for i := range *n {
		v := fleet.Vehicle{
			ID:         fmt.Sprintf("V-%03d", i+1),
			BatteryPct: 20 + rand.Float64()*80,
			SpeedKPH:   rand.Float64() * 80,
			TempC:      18 + rand.Float64()*10,
			IsCharging: rand.Float64() < 0.2,
			Timestamp:  time.Now(),
		}
		vehicles[i] = v
		fs.Add(context.Background(), v)
	}

	telemetryCh := make(chan fleet.Vehicle, *n)
	done := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(*n + 1)

	for _, v := range vehicles {
		v := v
		go func() {
			defer wg.Done()
			sim.SimulateVehicle(v, telemetryCh, done)
		}()
	}

	// launch ingest goroutine to sit and wait for passed in vehicles
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("error opening pgx pool", "err", err)
		os.Exit(1)
	}
	go func() {
		defer wg.Done()
		Ingest(fs, telemetryCh, done, alertLog, pool)
	}()

	// start a server in its own go routine
	// not in wait group since it doesn't depend on the done channel
	srv := server.NewServer(fs, alertLog, *port)
	go func() {
		slog.Info("server starting", "port", *port)
		if err := http.ListenAndServe(":"+*port, srv); err != nil {
			slog.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	// wait for signal.Notify and then close done if received
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	close(done)
	wg.Wait()
	slog.Info("all routines closed, exiting....")
}
