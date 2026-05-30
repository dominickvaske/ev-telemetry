# ev-telemetry

A real-time electric vehicle fleet telemetry system built in Go. This project simulates a fleet of EVs reporting live telemetry data — battery level, speed, temperature — and exposes that data through a JSON REST API.

The goal is to build toward a live fleet monitoring dashboard: a single screen showing every vehicle's status in real time, with an alert log that surfaces critical events as they happen.

## What it does

- Simulates multiple vehicles concurrently, each running in its own goroutine and reporting telemetry every second
- Ingests vehicle updates into a thread-safe in-memory store protected by `sync.RWMutex`
- Fires alerts when battery levels cross critical thresholds and accumulates them in a thread-safe log
- Exposes the fleet state and alert history over HTTP via a JSON REST API

## API

The server runs on `:8080` by default. The port is configurable with the `--port` flag.

### `GET /fleet`
Returns the current state of all vehicles as a JSON array.

```bash
curl http://localhost:8080/fleet
```

### `GET /vehicle/{id}`
Returns a single vehicle by ID. Returns a `404` with a JSON error if not found.

```bash
curl http://localhost:8080/vehicle/V-001
```

### `GET /alerts`
Returns all alerts that have fired since the server started.

```bash
curl http://localhost:8080/alerts
```

### `POST /telemetry`
Accepts a JSON vehicle payload and updates that vehicle's state in the store. The vehicle must already exist (identified by `id`).

```bash
curl -X POST http://localhost:8080/telemetry \
  -H "Content-Type: application/json" \
  -d '{"id":"V-001","battery_pct":50.0,"speed_kph":30.0,"temp_c":22.0,"is_charging":false}'
```

### Error responses
All errors return a consistent JSON envelope:
```json
{"error": "vehicle not found"}
```

## Project structure

```
ev-telemetry/
├── cmd/
│   └── server/
│       └── main.go         # Entry point: wires simulation, store, alert log, and HTTP server
├── internal/
│   ├── fleet/
│   │   └── fleet.go        # Core types: Vehicle, VehicleUpdate, Summary
│   ├── store/
│   │   ├── store.go        # Thread-safe in-memory fleet store (RWMutex)
│   │   └── errors.go       # Sentinel errors
│   ├── alert/
│   │   └── alert.go        # Alert types and thread-safe AlertLog
│   ├── sim/
│   │   └── sim.go          # Vehicle simulation goroutines
│   └── server/
│       ├── server.go       # HTTP server, chi router, response helpers
│       └── handlers.go     # Route handlers
```

## Running locally

```bash
go run ./cmd/server/
```

With a custom port:
```bash
go run ./cmd/server/ --port 9090
```

## Concepts covered

This project was built as a learning exercise covering:

- **Goroutines and channels** — concurrent vehicle simulation with fan-in ingestion
- **Mutexes** — `sync.RWMutex` for safe concurrent store access; `sync/atomic` for a lock-free counter
- **HTTP servers** — `net/http` with the chi router, handler methods on a Server struct
- **JSON** — struct tags, encoding/decoding request and response bodies
- **Context** — `context.Context` threaded through handlers into store operations
- **Package design** — isolated packages with clear responsibilities and well-defined interfaces

## Roadmap

- [ ] Richer simulation (speed variation, temperature, location)
- [ ] Handler tests with `httptest`
- [ ] Docker and docker-compose
- [ ] PostgreSQL persistence
- [ ] Live dashboard frontend
