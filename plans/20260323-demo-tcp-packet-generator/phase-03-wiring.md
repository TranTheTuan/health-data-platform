# Phase 03 — Wiring (Router + main.go)

## Context Links
- Router: `internal/delivery/http/router.go`
- Entry point: `cmd/api/main.go`
- Demo handler: `internal/handler/http/demo.go` (Phase 02)
- Session manager: `internal/demo/session_manager.go` (Phase 01)

## Overview
- **Priority:** P2
- **Status:** Pending (blocked by Phase 02)
- **Description:** Register the 4 demo HTTP routes and wire `DemoHandler` + `SessionManager` into dependency injection in `main.go`.

## Key Insights
- `RegisterRoutes` currently takes `*AuthHandler` and `*DeviceHandler`; must add `*DemoHandler` param
- All demo routes go under `/protected/` group (already auth-guarded via middleware)
- `SessionManager` is application-scoped (singleton), created once in `main()`
- `TCPAddr` from `cfg` is passed to `NewSessionManager`

## Requirements

### Functional
- 4 routes registered under `/protected/devices/:id/demo/...`
- `SessionManager` instantiated once in `main()` before handlers
- `DemoHandler` injected with `SessionManager` + `DeviceService`

## Architecture

### router.go changes

```go
// Change signature to accept DemoHandler
func RegisterRoutes(e *echo.Echo, ah *http_handler.AuthHandler, dh *http_handler.DeviceHandler, dmh *http_handler.DemoHandler) {
    // ... existing routes unchanged ...

    // Demo session routes
    protected.POST("/devices/:id/demo/session", dmh.StartSession)
    protected.DELETE("/devices/:id/demo/session", dmh.StopSession)
    protected.GET("/devices/:id/demo/session", dmh.SessionStatus)
    protected.POST("/devices/:id/demo/packets", dmh.SendBurst)
}
```

### main.go changes

Add after `devSvc` creation:
```go
import "github.com/TranTheTuan/health-data-platform/internal/demo"

// Demo session manager (application-scoped singleton)
demoSessions := demo.NewSessionManager(cfg.TCPAddr)
demoHandler := http_handler.NewDemoHandler(demoSessions, devSvc)
```

Update `RegisterRoutes` call:
```go
http_delivery.RegisterRoutes(e, authHttpHandler, devHttpHandler, demoHandler)
```

## Related Code Files

| File | Action | Description |
|------|--------|-------------|
| `internal/delivery/http/router.go` | **MODIFY** | Add `*DemoHandler` param + 4 new routes |
| `cmd/api/main.go` | **MODIFY** | Instantiate `SessionManager` + `DemoHandler`, update `RegisterRoutes` call |

## Implementation Steps

1. Modify `internal/delivery/http/router.go`:
   - Add `dmh *http_handler.DemoHandler` to `RegisterRoutes` signature
   - Add 4 protected routes for demo endpoints

2. Modify `cmd/api/main.go`:
   - Add import for `internal/demo`
   - After `devSvc` creation, create `demoSessions` and `demoHandler`
   - Update `RegisterRoutes` call to pass `demoHandler`

3. Run `go build ./...` to verify compile

## Todo List

- [ ] Modify `internal/delivery/http/router.go` — add `*DemoHandler` param + 4 routes
- [ ] Modify `cmd/api/main.go` — instantiate `SessionManager`, `DemoHandler`, update `RegisterRoutes`
- [ ] Verify compile: `go build ./...`

## Success Criteria

- `go build ./...` passes with no errors
- `router.go` stays under 50 LOC
- `main.go` stays under 100 LOC

## Risk Assessment

| Risk | Mitigation |
|------|-----------|
| `RegisterRoutes` signature change breaks compile | Only one call site in `main.go`; easy to update |
| `SessionManager` not closed on shutdown | TCP connections have server-side idle timeout; acceptable for demo |

## Next Steps

- Phase 04: Add demo control panel UI to `packets.html`
