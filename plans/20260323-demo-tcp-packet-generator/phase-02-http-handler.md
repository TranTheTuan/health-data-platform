# Phase 02 — HTTP Demo Handler

## Context Links
- Existing handler pattern: `internal/handler/http/device.go`
- Session manager: `internal/demo/session_manager.go` (Phase 01)
- Router: `internal/delivery/http/router.go`

## Overview
- **Priority:** P2
- **Status:** Pending (blocked by Phase 01)
- **Description:** Create `internal/handler/http/demo.go` with 4 HTTP endpoints managing the demo session lifecycle.

## Key Insights
- Reuse `checkOwnership` pattern from `DeviceHandler` — but `checkOwnership` is a method, not exported; duplicate the logic or accept it as a helper call to `svc.ListDevices`
- IMEI retrieval: call `svc.ListDevices(ctx, userID)` and filter by device ID (returns `DeviceResponse` which includes `IMEI`)
- `DemoHandler` needs: `*demo.SessionManager` + `service.DeviceService` (for ownership + IMEI lookup)
- Default burst count: **7** packets (middle of 5–10 range)
- Return JSON for all endpoints: `{"active": true/false}` for status; `{"sent": N}` for burst; `{"error": "..."}` for errors

## Requirements

### Functional
- `POST /protected/devices/:id/demo/session` — start session; 409 if already active, 400 if IMEI not in TCP server
- `DELETE /protected/devices/:id/demo/session` — stop session; 404 if not active
- `GET /protected/devices/:id/demo/session` — returns `{"active": bool}`
- `POST /protected/devices/:id/demo/packets` — sends burst; 404 if session not active; returns `{"sent": N}`
- All endpoints: 401 if unauthenticated, 403 if device not owned by user

### Non-functional
- File ≤ 150 LOC
- Follow exact same auth/ownership pattern as `DeviceHandler`

## Architecture

```go
package http

import (
    "errors"
    "net/http"

    "github.com/labstack/echo/v4"
    "github.com/TranTheTuan/health-data-platform/internal/demo"
    "github.com/TranTheTuan/health-data-platform/internal/service"
)

// DemoHandler exposes HTTP endpoints for the demo TCP session feature.
type DemoHandler struct {
    sessions *demo.SessionManager
    svc      service.DeviceService
}

func NewDemoHandler(sessions *demo.SessionManager, svc service.DeviceService) *DemoHandler {
    return &DemoHandler{sessions: sessions, svc: svc}
}

// getIMEI fetches the IMEI for a device owned by the given user.
// Returns "", false if not found or not owned.
func (h *DemoHandler) getIMEI(c echo.Context, userID, deviceID string) (string, bool) {
    devices, err := h.svc.ListDevices(c.Request().Context(), userID)
    if err != nil {
        return "", false
    }
    for _, d := range devices {
        if d.ID == deviceID {
            return d.IMEI, true
        }
    }
    return "", false
}

func (h *DemoHandler) StartSession(c echo.Context) error {
    userID, ok := c.Get("user_id").(string)
    if !ok || userID == "" {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
    }
    deviceID := c.Param("id")
    imei, ok := h.getIMEI(c, userID, deviceID)
    if !ok {
        return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
    }
    if err := h.sessions.StartSession(deviceID, imei); err != nil {
        if errors.Is(err, demo.ErrSessionAlreadyActive) {
            return c.JSON(http.StatusConflict, map[string]string{"error": "session already active"})
        }
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, map[string]bool{"active": true})
}

func (h *DemoHandler) StopSession(c echo.Context) error {
    userID, ok := c.Get("user_id").(string)
    if !ok || userID == "" {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
    }
    deviceID := c.Param("id")
    if _, ok := h.getIMEI(c, userID, deviceID); !ok {
        return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
    }
    if err := h.sessions.StopSession(deviceID); err != nil {
        return c.JSON(http.StatusNotFound, map[string]string{"error": "no active session"})
    }
    return c.JSON(http.StatusOK, map[string]bool{"active": false})
}

func (h *DemoHandler) SessionStatus(c echo.Context) error {
    userID, ok := c.Get("user_id").(string)
    if !ok || userID == "" {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
    }
    deviceID := c.Param("id")
    if _, ok := h.getIMEI(c, userID, deviceID); !ok {
        return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
    }
    return c.JSON(http.StatusOK, map[string]bool{"active": h.sessions.IsActive(deviceID)})
}

func (h *DemoHandler) SendBurst(c echo.Context) error {
    userID, ok := c.Get("user_id").(string)
    if !ok || userID == "" {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
    }
    deviceID := c.Param("id")
    if _, ok := h.getIMEI(c, userID, deviceID); !ok {
        return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
    }
    const burstCount = 7
    if err := h.sessions.SendBurst(deviceID, burstCount); err != nil {
        if errors.Is(err, demo.ErrSessionNotFound) {
            return c.JSON(http.StatusNotFound, map[string]string{"error": "no active session"})
        }
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, map[string]int{"sent": burstCount})
}
```

## Related Code Files

| File | Action | Description |
|------|--------|-------------|
| `internal/handler/http/demo.go` | **CREATE** | 4 HTTP handlers for demo session |

## Implementation Steps

1. Create `internal/handler/http/demo.go`
2. Implement `DemoHandler` struct with `NewDemoHandler()`
3. Implement `getIMEI()` helper (ownership + IMEI in one call)
4. Implement all 4 handler methods following the pseudocode above
5. Run `go build ./internal/handler/http/...` to verify compile

## Todo List

- [ ] Create `internal/handler/http/demo.go`
- [ ] Implement `DemoHandler` + `NewDemoHandler`
- [ ] Implement `getIMEI`, `StartSession`, `StopSession`, `SessionStatus`, `SendBurst`
- [ ] Verify compile: `go build ./internal/handler/http/...`

## Success Criteria

- File ≤ 150 LOC
- Compiles without errors
- All 4 methods follow auth/ownership guard pattern from `DeviceHandler`

## Risk Assessment

| Risk | Mitigation |
|------|-----------|
| `ListDevices` called twice per request (status + burst) | Acceptable — low traffic demo feature; no caching needed |
| `getIMEI` duplicates `checkOwnership` logic | Intentional — returns IMEI in same pass, avoids second DB call |

## Next Steps

- Phase 03: Register routes in router.go and wire dependencies in main.go
