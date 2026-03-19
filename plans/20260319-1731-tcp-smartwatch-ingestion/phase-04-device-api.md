# Phase 4: Device Registration API (Echo)

**Context Links**
- [Main Plan](./plan.md)
- [Phase 3: Integration](./phase-03-integration.md)

## Overview
- Priority: P2
- Effort: 1.5h
- Status: Pending
- Add an authenticated HTTP endpoint for users to register a new smartwatch device and receive a `device_token`.

## Key Insights
- Only logged-in users (session cookie set by Google OAuth) can register devices.
- The token generated server-side is a cryptographically random UUID or 32-byte hex string.
- Store `device_token` hashed in the DB (like a password) to prevent token exposure from DB leaks.

## Requirements
- `POST /devices` — protected by `AuthMiddleware`, creates device record, returns `device_token`.
- `GET /devices` — lists devices owned by the current user.

## Architecture
```
POST /devices
  └── AuthMiddleware (session check)
        └── DeviceHandler.Create
              ├── generate random token
              ├── INSERT INTO devices (user_id, device_token, name)
              └── return {id, name, device_token}   ← only shown once!
```

## Related Code Files
- `[NEW]` `internal/api/handlers/device.go`
- `[NEW]` `internal/device/repository.go`
- `[MODIFY]` `internal/api/routes.go` — add device routes

## Implementation Steps

### `internal/device/repository.go`
1. `CreateDevice(ctx, db, userID, name, tokenHash string) (deviceID string, err error)`
2. `ListDevices(ctx, db, userID string) ([]DeviceRow, error)` — returns id + name only (no token).

### `internal/api/handlers/device.go`
1. `DeviceHandler struct { db *sql.DB }`, constructor `NewDeviceHandler(db)`.
2. `Create(c echo.Context)`:
   - Get `user_id` from context (set by middleware).
   - Parse optional `name` from JSON body.
   - Generate token: `crypto/rand` → 32 bytes → hex string.
   - Hash token: `sha256.Sum256([]byte(token))` → store hex of hash.
   - Call `CreateDevice`.
   - Return `{"device_token": "<raw_token>"}` — clarify in response this is shown once.
3. `List(c echo.Context)`:
   - Get `user_id` from context.
   - Call `ListDevices`.
   - Return JSON array.

### `internal/api/routes.go`
Add under protected group:
```go
devHandler := handlers.NewDeviceHandler(db)
protected.POST("/devices", devHandler.Create)
protected.GET("/devices", devHandler.List)
```

Requires passing `db` into `RegisterRoutes(e, cfg, db)`.

## Todo List
- [ ] Create `internal/device/repository.go`
- [ ] Create `internal/api/handlers/device.go`
- [ ] Update `internal/api/routes.go` to register device routes
- [ ] Update `cmd/api/main.go` to pass `db` to `RegisterRoutes`

## Success Criteria
- `POST /devices` with a valid session cookie returns a `device_token`.
- `GET /devices` returns the list of the user's devices (no token, just id/name).
- TCP server accepts `CONNECT <token>` using the raw token and finds the matching device.

## Security Considerations
- Raw `device_token` shown only once at creation. Store only the hash.
- TCP auth: hash the incoming CONNECT token and compare with stored hash.
- HTTPS required in production — device_token exposure over plain HTTP is a risk in plaintext.
