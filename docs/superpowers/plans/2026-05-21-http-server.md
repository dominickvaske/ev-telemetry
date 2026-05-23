# HTTP Server Implementation Plan

> **Note:** This plan is written for the learner to implement themselves. Steps describe *what* to do and *why*, with small hints — not complete solutions. Work through one task at a time and ask your tutor when stuck.

**Goal:** Add a JSON REST API to ev-telemetry using chi and an `internal/server` package, running alongside the existing simulation goroutines.

**Architecture:** A `Server` struct in `internal/server` holds references to a `FleetStore` and `AlertLog` and registers four chi routes. `main.go` creates both stores, passes them to `NewServer`, and starts the HTTP server in its own goroutine alongside the existing simulation.

**Tech Stack:** Go stdlib `net/http`, `github.com/go-chi/chi/v5`, `encoding/json`

**Spec:** `docs/superpowers/specs/2026-05-21-http-server-design.md`

---

## File Map

| Action | File | Responsibility |
|--------|------|----------------|
| Modify | `internal/fleet/fleet.go` | Add JSON struct tags to `Vehicle` and `Summary` |
| Modify | `internal/store/store.go` | Add `context.Context` param to `Get`, `List`, `Set`, `Add` |
| Modify | `internal/alert/alert.go` | Add `AlertLog` struct with `Append` and `All` methods |
| Modify | `internal/sim/sim.go` | Update any `store` calls to match new context signatures |
| Create | `internal/server/server.go` | `Server` struct, constructor, route registration, helpers |
| Create | `internal/server/handlers.go` | The four handler methods |
| Modify | `cmd/server/main.go` | Port flag, create `AlertLog`, wire `NewServer`, start HTTP |

---

## Task 1: Add JSON tags to `fleet.Vehicle`

**File:** `internal/fleet/fleet.go`

**What:** Add a backtick struct tag to every field in `Vehicle` and `Summary` so `encoding/json` serializes them to snake_case.

**Why:** Without tags, Go exports `BatteryPct` as `"BatteryPct"` in JSON. With the tag `` `json:"battery_pct"` `` it becomes `"battery_pct"`. This is your API contract — pick it now and don't change it.

**Hint — tag syntax:**
```go
FieldName Type `json:"field_name"`
```

**Steps:**
- [ ] Open `internal/fleet/fleet.go`
- [ ] Add a snake_case `json:` tag to every field in `Vehicle` (ID, BatteryPct, SpeedKPH, TempC, IsCharging, Timestamp)
- [ ] Add tags to `Summary` fields as well
- [ ] Run `go build ./...` — it should compile cleanly with no errors

---

## Task 2: Install chi

**What:** Add chi as a dependency.

**Steps:**
- [ ] In your terminal, from the project root, run:
  ```
  go get github.com/go-chi/chi/v5
  ```
- [ ] Verify `go.mod` and `go.sum` were updated
- [ ] Run `go build ./...` to confirm the dependency resolves

---

## Task 3: Add `context.Context` to store methods

**File:** `internal/store/store.go`

**What:** Add `ctx context.Context` as the first parameter to the four methods that HTTP handlers will call: `Add`, `Get`, `Set`, and `List`.

**Why:** This is the Go convention for any function that could block or call external resources. It threads cancellation signals from the HTTP request all the way down. See spec for full reasoning.

**You'll need to import:** `"context"` at the top of the file.

**Hint — what the signature change looks like for `Get`:**
```go
// Before
func (fs *FleetStore) Get(id string) (fleet.Vehicle, bool)

// After
func (fs *FleetStore) Get(ctx context.Context, id string) (fleet.Vehicle, bool)
```

The body of each method does not change — just the signature.

**Steps:**
- [ ] Update `Add`, `Get`, `Set`, and `List` signatures to accept `ctx context.Context` as the first argument
- [ ] Add `"context"` to the import block
- [ ] Run `go build ./...` — it will fail because `main.go` and `sim.go` call these methods. That's expected; you'll fix callers next.
- [ ] In `cmd/server/main.go`, update the calls to `fs.Add`, `fs.Set`, `fs.Get`, `fs.List` to pass `context.Background()` for now (handlers will pass the real context later)
- [ ] In `internal/sim/sim.go`, do the same for any store method calls
- [ ] Run `go build ./...` — should compile cleanly now

---

## Task 4: Create `AlertLog` in the alert package

**File:** `internal/alert/alert.go`

**What:** Add a new exported struct `AlertLog` that safely accumulates `Alert` values from multiple goroutines.

**Why:** Both `Ingest` (running in a goroutine) and the HTTP handler (running on an HTTP goroutine) will access this — so it needs a mutex, just like `FleetStore`.

**What it needs:**
- A `sync.RWMutex` field
- A `[]Alert` slice field
- A constructor `NewAlertLog() *AlertLog`
- An `Append(a Alert)` method — uses a write lock
- An `All() []Alert` method — uses a read lock, returns a copy of the slice

**Hint — why return a copy in `All()`:** If you return the underlying slice directly, the caller could modify it, corrupting the log. Return `append([]Alert{}, log.alerts...)` to hand back a safe copy.

**Steps:**
- [ ] Add `AlertLog` struct to `internal/alert/alert.go`
- [ ] Add `NewAlertLog`, `Append`, and `All` methods
- [ ] Add `"sync"` to the import block if it isn't already there
- [ ] Run `go build ./...` — should compile cleanly

---

## Task 5: Create `internal/server/server.go`

**File:** `internal/server/server.go` (new file, new package)

**What:** The `Server` struct, its constructor, and two private helper functions.

**Steps:**
- [ ] Create the directory `internal/server/`
- [ ] Create `server.go` with `package server`
- [ ] Define the `Server` struct with three fields: `store *store.FleetStore`, `alerts *alert.AlertLog`, `router *chi.Mux`
- [ ] Write `NewServer(store *store.FleetStore, alerts *alert.AlertLog) *Server`
  - Creates a `chi.NewRouter()`
  - Registers the four routes (you haven't written the handlers yet — you can register them as `nil` placeholders or skip until Task 6, then come back)
  - Returns the populated struct
- [ ] Add a `ServeHTTP(w http.ResponseWriter, r *http.Request)` method that delegates to `s.router` — this makes `*Server` implement `http.Handler`
- [ ] Write `writeJSON(w http.ResponseWriter, status int, v any)`:
  - Sets `Content-Type: application/json`
  - Calls `w.WriteHeader(status)`
  - Uses `json.NewEncoder(w).Encode(v)`
- [ ] Write `writeError(w http.ResponseWriter, status int, message string)`:
  - Calls `writeJSON` with a struct like `struct{ Error string \`json:"error"\` }{message}`
- [ ] Run `go build ./...`

---

## Task 6: Implement `GET /fleet`

**File:** `internal/server/handlers.go` (new file)

**What:** The simplest handler — calls `store.List` and encodes the result.

**Steps:**
- [ ] Create `handlers.go` with `package server`
- [ ] Write method `func (s *Server) handleGetFleet(w http.ResponseWriter, r *http.Request)`
  - Call `s.store.List(r.Context())`
  - Pass the result to `writeJSON` with status 200
  - If the slice is nil (empty store), make sure you pass an empty slice `[]fleet.Vehicle{}` not `nil` — JSON encodes `nil` slices as `null`, not `[]`
- [ ] Register this handler in `NewServer`: `r.Get("/fleet", s.handleGetFleet)`
- [ ] Run `go build ./...`

---

## Task 7: Implement `GET /vehicle/{id}`

**File:** `internal/server/handlers.go`

**What:** Extract the URL parameter, look up the vehicle, return 404 if not found.

**Hint — how chi URL params work:**
```go
id := chi.URLParam(r, "id")
```

**Steps:**
- [ ] Write `func (s *Server) handleGetVehicle(w http.ResponseWriter, r *http.Request)`
  - Extract `id` using `chi.URLParam`
  - Call `s.store.Get(r.Context(), id)`
  - If not found (`ok == false`), call `writeError` with status 404 and message `"vehicle not found"`
  - If found, call `writeJSON` with status 200 and the vehicle
- [ ] Register in `NewServer`: `r.Get("/vehicle/{id}", s.handleGetVehicle)`
- [ ] Run `go build ./...`

---

## Task 8: Implement `GET /alerts`

**File:** `internal/server/handlers.go`

**What:** Returns the full alert log accumulated since server start.

**Steps:**
- [ ] Write `func (s *Server) handleGetAlerts(w http.ResponseWriter, r *http.Request)`
  - Call `s.alerts.All()`
  - Same empty-slice guard as `handleGetFleet`
  - `writeJSON` with status 200
- [ ] Register in `NewServer`: `r.Get("/alerts", s.handleGetAlerts)`
- [ ] Run `go build ./...`

---

## Task 9: Implement `POST /telemetry`

**File:** `internal/server/handlers.go`

**What:** The only write endpoint. Decodes a JSON body into a `Vehicle`, validates it, and updates the store.

**Why validate ID:** An empty ID would silently corrupt the store — you'd update the vehicle keyed to `""`.

**Hint — decoding JSON from a request body:**
```go
var v fleet.Vehicle
if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
    // bad JSON — return 400
}
```

**Steps:**
- [ ] Write `func (s *Server) handlePostTelemetry(w http.ResponseWriter, r *http.Request)`
  - Decode the body into `fleet.Vehicle`; on error return 400 with `writeError`
  - Check `v.ID == ""`; if so return 400 with message `"id is required"`
  - Call `s.store.Set(r.Context(), v)`
  - If `Set` returns `store.ErrVehicleNotFound`, return 404
  - On success, return 200 with the vehicle as JSON
- [ ] Register in `NewServer`: `r.Post("/telemetry", s.handlePostTelemetry)`
- [ ] Run `go build ./...`

---

## Task 10: Update `Ingest` to use `AlertLog`

**File:** `cmd/server/main.go`

**What:** The existing `Ingest` function logs a `Printf` when battery is low. Wire it to actually append to the `AlertLog` instead.

**Steps:**
- [ ] Update `Ingest` signature to accept `*alert.AlertLog` as a second parameter
- [ ] Replace the `log.Printf` battery alert with a call to `alertLog.Append(...)` that creates a proper `alert.Alert` struct (fill in ID, AlertVEHID, Type, Value, Message, TimeStamp)
- [ ] Note: `GlobalAlertID` in `alert.go` is flagged as not thread-safe. For now, keep using it — you'll address this later. Just be aware.
- [ ] Update the call site in `main()` to pass the alert log
- [ ] Run `go build ./...`

---

## Task 11: Wire everything in `main.go`

**File:** `cmd/server/main.go`

**What:** Add a port flag, create the `AlertLog`, create the server, and start listening.

**Steps:**
- [ ] Add a port flag at the top of `main` using `flag.String` or `flag.Int`:
  ```go
  port := flag.String("port", "8080", "HTTP server port")
  flag.Parse()
  ```
- [ ] Create `alertLog := alert.NewAlertLog()`
- [ ] Create `srv := server.NewServer(fs, alertLog)`
- [ ] Start the HTTP server in its own goroutine so it doesn't block the simulation:
  ```go
  go func() {
      log.Printf("listening on :%s", *port)
      if err := http.ListenAndServe(":"+*port, srv); err != nil {
          log.Fatal(err)
      }
  }()
  ```
- [ ] Run `go build ./...` — full clean build
- [ ] Run the server: `go run ./cmd/server/`

---

## Task 12: Verify with curl

**What:** Manually test each endpoint to confirm the server works end-to-end.

**Steps:**
- [ ] Start the server in one terminal: `go run ./cmd/server/`
- [ ] In a second terminal, test each endpoint:

  ```bash
  # All vehicles
  curl localhost:8080/fleet

  # Single vehicle
  curl localhost:8080/vehicle/V-001

  # Vehicle that doesn't exist
  curl localhost:8080/vehicle/V-999

  # Alert log (wait a few seconds for battery to drop below 10% on any vehicle)
  curl localhost:8080/alerts

  # Post a telemetry update (update V-001's battery)
  curl -X POST localhost:8080/telemetry \
    -H "Content-Type: application/json" \
    -d '{"id":"V-001","battery_pct":50.0,"speed_kph":30.0,"temp_c":22.0,"is_charging":false}'
  ```

- [ ] Confirm each returns valid JSON with the correct status codes
- [ ] Confirm `GET /vehicle/V-999` returns `{"error":"vehicle not found"}` with a 404
- [ ] Once everything works, commit all changes together

---

## What's Next (future sessions)

- Writing Go tests for handlers using `httptest`
- PostgreSQL database integration (context is already threaded through)
- Docker: containerising the server
- Frontend dashboard on top of this API