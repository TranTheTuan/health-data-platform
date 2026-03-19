# Phase 4: Integration & Config

**Context Links**
- [Main Plan](./plan.md)
- [Phase 3: TCP Server](./phase-03-tcp-server.md)

## Overview
- Priority: P1
- Effort: 1h
- Status: Pending
- Wire the TCP server and DB pool into `main.go` alongside the existing Echo HTTP server. Both run concurrently, share the DB pool, and shut down gracefully on SIGINT.

## Key Insights
- Both servers share a single `*sql.DB` pool — open once at startup, pass to both.
- Use `context` + `os/signal` for graceful shutdown (SIGINT/SIGTERM cancels root context → both servers stop).
- `golang.org/x/sync/errgroup` cleanly runs two blocking servers and propagates the first error.
- The current `main.go` uses `middleware.RequestLogger()` (user changed it). Keep that.

## Related Code Files
- `[MODIFY]` `cmd/api/main.go`
- `[MODIFY]` `configs/config.go` — `DatabaseURL`, `TCPAddr` (added in Phase 1)
- `[NEW]` `internal/db/connection.go` (added in Phase 1)

## Implementation Steps

### `cmd/api/main.go` revised structure:
```go
func main() {
    cfg := configs.LoadConfig()
    auth.InitGoogleOAuth(cfg)

    // DB pool (shared by HTTP & TCP)
    pool, err := db.NewPool(cfg.DatabaseURL)
    if err != nil { log.Fatal(err) }
    defer pool.Close()

    // Echo HTTP server
    e := echo.New()
    e.Use(middleware.RequestLogger())
    e.Use(middleware.Recover())
    api.RegisterRoutes(e, cfg, pool) // pass pool for device handler

    // TCP server
    tcpSrv := tcp.NewServer(cfg.TCPAddr, pool)

    // Graceful shutdown via SIGINT/SIGTERM
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    g, gCtx := errgroup.WithContext(ctx)
    g.Go(func() error { return e.Start(":8080") })
    g.Go(func() error { return tcpSrv.Start(gCtx) })

    if err := g.Wait(); err != nil {
        log.Println("Server stopped:", err)
    }
}
```

### `internal/api/routes.go` signature change:
```go
func RegisterRoutes(e *echo.Echo, cfg *configs.Config, db *sql.DB)
```
- Pass `db` down to device handler (Phase 5).

## Todo List
- [ ] `go get golang.org/x/sync`
- [ ] Refactor `cmd/api/main.go`: add signal handling, errgroup, DB pool, TCP server
- [ ] Update `internal/api/routes.go` signature to accept `db *sql.DB`

## Success Criteria
- `go build -o /tmp/hdp ./cmd/api/main.go` exits 0
- Server starts, logs `:8080` (HTTP) and `:9090` (TCP) listeners
- Ctrl+C triggers clean shutdown (no hanging goroutine or panic)

## Security Considerations
- `DATABASE_URL` loaded from env only — never hardcoded.
- If DB connection fails at startup, the server exits rather than starting in a broken state.

## Next Steps
- [Phase 5: Device Registration API](./phase-05-device-api.md)
