# HTTP Server Design — ev-telemetry

**Date:** 2026-05-21
**Status:** Approved

## Goal

Add a JSON REST API to ev-telemetry so that an external dashboard (or curl) can read fleet state and alert history in real time. The simulation goroutines continue running internally; HTTP is the external interface.

## Architecture

```
sim goroutines (4 vehicles, 1s tick)
       │
       ▼
   Ingest()  ──── checks alert thresholds ──▶ AlertLog.Append()
       │
       ▼
  FleetStore
       │
       ▼
  server.Server (chi router, :8080)
  ├── POST /telemetry
  ├── GET  /fleet
  ├── GET  /vehicle/{id}
  └── GET  /alerts
```

`FleetStore` and `AlertLog` are created once in `main.go` and shared between the simulation path and HTTP path. Both are safe for concurrent access via `sync.RWMutex`.

## Changes to Existing Packages

### `internal/fleet/fleet.go`
- Add snake_case JSON struct tags to `Vehicle` and `Summary`
- All fields must serialize/deserialize correctly with `encoding/json`

### `internal/store/store.go`
- Add `context.Context` as the first parameter to public methods used by handlers: `Get`, `List`, `Set`, `Add`
- Handlers will pass `r.Context()` into these calls
- Store does not need to act on context for now — accepting it builds the habit and prepares for database integration

### `internal/alert/alert.go`
- Add a new exported `AlertLog` struct: a `sync.RWMutex`-protected `[]Alert` slice
- Two methods: `Append(a Alert)` and `All() []Alert`
- `Ingest()` in `main.go` will call `Append` when a threshold is crossed

## New Package: `internal/server`

### `server.go`

**`Server` struct:**
```
store   *store.FleetStore
alerts  *alert.AlertLog
router  *chi.Mux
```

**Constructor:** `NewServer(store, alerts) *Server`
- Registers all routes on a new chi.Mux
- Returns a ready-to-use Server

**`Server` implements `http.Handler`** by delegating `ServeHTTP` to its router.

**Helper functions (private):**
- `writeJSON(w, statusCode, v)` — sets Content-Type, writes status, encodes v as JSON
- `writeError(w, statusCode, message)` — writes `{"error": "<message>"}` envelope

### Endpoints

#### `POST /telemetry`
- Decode request body as `fleet.Vehicle`
- Validate: ID must not be empty
- Call `store.Set(ctx, v)` — returns 404 JSON if vehicle not found, 400 if body is malformed
- Returns 200 with the updated vehicle as JSON on success

#### `GET /fleet`
- Call `store.List(ctx)`
- Return JSON array of all vehicles
- Empty fleet returns `[]` (not null)

#### `GET /vehicle/{id}`
- Extract `{id}` from URL using chi's `URLParam`
- Call `store.Get(ctx, id)`
- Return 404 JSON error if not found
- Return 200 with vehicle JSON on success

#### `GET /alerts`
- Call `alertLog.All()`
- Return JSON array of all alerts fired since server start
- Empty returns `[]` (not null)

### HTTP Conventions (apply to all endpoints)
- `Content-Type: application/json` on every response
- Error envelope: `{"error": "human readable message"}`
- Status codes: 200 OK, 400 Bad Request, 404 Not Found, 405 Method Not Allowed (handled by chi)

## `main.go` Changes

1. Add `--port` flag (default `8080`) for configurable port
2. Create `AlertLog` alongside `FleetStore`
3. Pass both into `server.NewServer`
4. Start `http.ListenAndServe` in its own goroutine
5. Update `Ingest()` signature to accept `*alert.AlertLog` and call `Append` when battery < 10%

## What Comes After (out of scope for this spec)

- Frontend dashboard (terminal UI or web UI) built on top of this API
- Database integration (PostgreSQL) — context is already threaded through for this
- Additional alert types (speed, temperature)
- WebSocket or SSE for live push to dashboard
- Tests