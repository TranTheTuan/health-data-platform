# Phase 3: Integration & Config

**Context Links**
- [Main Plan](./plan.md)
- [Phase 2: TCP Server](./phase-02-tcp-server.md)

## Overview
- Priority: P1
- Effort: 1h
- Status: Pending
- Wire the TCP server into `main.go` alongside the existing Echo HTTP server. Both run concurrently.

## Key Insights
- Both servers share the same `*sql.DB` pool — open once, pass to both.
- Use `errgroup` to run both servers and propagate shutdown on error.
- On `SIGINT`/`SIGTERM`, cancel the root context → both servers shut down gracefully.

## Requirements
- TCP server and HTTP/Echo server start concurrently from `main.go`.
- Shared DB connection pool.
- `os/signal` + context cancellation for clean shutdown.

## Related Code Files
- `[MODIFY]` `cmd/api/main.go`
- `[MODIFY]` `configs/config.go` — `TCPAddr`, `DatabaseURL` already added in prior phases.
- `[NEW]` `internal/db/connection.go`

## Implementation Steps

### `internal/db/connection.go`
```go
package db

import (
    "database/sql"
    _ "github.com/jackc/pgx/v5/stdlib"
)

func NewPool(databaseURL string) (*sql.DB, error) {
    db, err := sql.Open("pgx", databaseURL)
    if err != nil {
        return nil, err
    }
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    return db, db.Ping()
}
```

### `cmd/api/main.go` changes
1. Load `cfg` as before.
2. Call `db.NewPool(cfg.DatabaseURL)` → `pool`.
3. Init Google OAuth as before.
4. Create `tcp.NewServer(cfg.TCPAddr, pool)`.
5. Set up `signal.NotifyContext` for SIGINT/SIGTERM.
6. Use `errgroup.Group` (from `golang.org/x/sync/errgroup`) to run:
   - Echo server: `e.Start(...)`
   - TCP server: `tcpSrv.Start(ctx)`
7. On any error or signal → cancel context → both shut down.

## Todo List
- [ ] Create `internal/db/connection.go`
- [ ] Refactor `cmd/api/main.go` to add signal handling, errgroup, DB pool, TCP server
- [ ] `go get golang.org/x/sync`

## Success Criteria
- `go build -o /tmp/hdp ./cmd/api/main.go` succeeds.
- Server starts with both `:8080` (HTTP) and `:9090` (TCP) listening.
- Ctrl+C causes both to shut down cleanly (no hanging goroutines).

## Security Considerations
- DB credentials in `DATABASE_URL` — loaded from env, never hardcoded.

## Next Steps
- [Phase 4: Device Registration API](./phase-04-device-api.md)
