# ev-telemetry

A real-time electric vehicle fleet telemetry system built in Go. Simulates a fleet of EVs reporting live telemetry — battery level, speed, temperature — ingests that data into PostgreSQL, fires alerts on critical thresholds, and displays everything on a live web dashboard.

## Demo

Open `http://localhost:8080` after starting the server to see the live fleet dashboard.

![Dashboard showing fleet status, battery levels, and alert log]

## What it does

- Simulates N vehicles concurrently (configurable via `--vehicles` flag), each running in its own goroutine reporting telemetry every second
- Ingests vehicle updates into a thread-safe in-memory store and persists every event to PostgreSQL
- Fires alerts when battery, speed, or temperature cross critical thresholds — persisted to the database and shown in the live dashboard
- Exposes fleet state, alert history, and server metrics over HTTP
- Serves a live web dashboard that polls the API and updates in real time
- Exposes a Prometheus-compatible `/metrics` endpoint (event counters, latency histogram, goroutine count)
- JSON-structured logs via Go's `slog` package
- CI runs on every push via GitHub Actions

## Running locally

**Prerequisites:** Docker Desktop

```bash
# Start PostgreSQL
docker compose up -d

# Run database migrations
goose -dir migrations postgres "host=127.0.0.1 port=5432 user=ev-user password=ev-password dbname=ev-telemetry sslmode=disable" up

# Copy environment file and fill in values
cp .env.example .env

# Run the server (default: 4 vehicles on :8080)
go run ./cmd/server/

# Run with more vehicles
go run ./cmd/server/ --vehicles=50

# Custom port
go run ./cmd/server/ --port=9090
```

Open `http://localhost:8080` in your browser.

## API

| Method | Route | Description |
|--------|-------|-------------|
| `GET` | `/fleet` | All vehicles and their current state |
| `GET` | `/vehicle/{id}` | Single vehicle by ID |
| `GET` | `/alerts` | All alerts fired since startup |
| `POST` | `/telemetry` | Update a vehicle's state |
| `GET` | `/metrics` | Prometheus metrics endpoint |
| `GET` | `/status` | Server uptime, event count, goroutine count |

### Example requests

```bash
curl http://localhost:8080/fleet
curl http://localhost:8080/alerts
curl http://localhost:8080/status

curl -X POST http://localhost:8080/telemetry \
  -H "Content-Type: application/json" \
  -d '{"id":"V-001","battery_pct":50.0,"speed_kph":30.0,"temp_c":22.0,"is_charging":false}'
```

All errors return a consistent JSON envelope:
```json
{"error": "vehicle not found"}
```

## Project structure

```
ev-telemetry/
├── cmd/
│   └── server/
│       └── main.go             # Entry point: simulation, ingest, HTTP server
├── internal/
│   ├── alert/
│   │   ├── alert.go            # Alert types, CheckBattery/Speed/Temp, thread-safe Log
│   │   └── alert_test.go
│   ├── fleet/
│   │   └── fleet.go            # Core types: Vehicle, VehicleUpdate, Summary
│   ├── metrics/
│   │   └── metrics.go          # Prometheus counters, histogram, atomic event counter
│   ├── server/
│   │   ├── server.go           # Chi router, /status handler, static file serving
│   │   └── handlers.go         # Route handlers
│   ├── sim/
│   │   └── sim.go              # Vehicle simulation goroutines
│   └── store/
│       ├── store.go            # Thread-safe in-memory fleet store (RWMutex)
│       ├── store_test.go
│       └── errors.go
├── migrations/
│   ├── 00001_create_telemetry_events.sql
│   └── 00002_create_alerts_table.sql
├── static/
│   ├── index.html              # Dashboard layout
│   ├── style.css               # Dark terminal theme
│   └── app.js                  # Polling logic, DOM updates
├── .github/
│   └── workflows/
│       └── ci.yml              # Build and test on every push
├── docker-compose.yml
└── .env                        # Local secrets (not committed)
```

## Concepts covered

- **Goroutines and channels** — concurrent vehicle simulation with fan-in ingestion; buffered channel sized to fleet
- **Mutexes** — `sync.RWMutex` for safe concurrent store access; `sync/atomic` for lock-free counters
- **PostgreSQL** — pgx connection pool, parameterised queries, Goose migrations
- **Prometheus** — Counter, Histogram metric types; `/metrics` endpoint for scraping
- **Structured logging** — JSON logs via `slog` with key-value fields
- **HTTP servers** — `net/http` with chi router, handler methods, static file serving
- **GitHub Actions CI** — build and test on every push to main
- **Docker Compose** — containerised PostgreSQL with named volume for persistence
- **Context** — `context.Context` threaded through handlers and store operations
- **Package design** — isolated packages with clear responsibilities
