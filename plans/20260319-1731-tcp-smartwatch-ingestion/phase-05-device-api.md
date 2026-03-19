# Phase 5: Device Registration API (HTTP/Echo)

**Context Links**
- [Main Plan](./plan.md)
- [Phase 4: Integration](./phase-04-integration.md)

## Overview
- Priority: P2
- Effort: 2h
- Status: Pending
- Add authenticated HTTP API endpoints allowing users to register their smartwatch by IMEI and link it to their account, and to list their registered devices.

## Key Insights
- **IMEI is pre-known** — the user looks up their watch IMEI from the device itself (printed on back or visible in watch settings). They submit it to `POST /devices` to register it under their account.
- No token generation needed (unlike old plan). IMEI is the identity. Server just stores the association `(imei, user_id)`.
- Endpoints are **protected by `AuthMiddleware`** (session cookie from Google OAuth).
- `DELETE /devices/:id` is optional (Phase 5 scope stretch goal).

## Endpoints

```
POST   /devices         Register a device by IMEI
GET    /devices         List devices owned by current user
DELETE /devices/:id     (optional) unlink a device
```

### POST /devices — Request body
```json
{ "imei": "353456789012345", "name": "My Garmin Watch" }
```
### POST /devices — Response
```json
{ "id": "<uuid>", "imei": "353456789012345", "name": "My Garmin Watch", "created_at": "..." }
```

### GET /devices — Response
```json
[{ "id": "...", "imei": "...", "name": "...", "last_seen_at": "...", "created_at": "..." }]
```

## Related Code Files
- `[NEW]` `internal/device/repository.go` — `RegisterDevice`, `ListDevices`
- `[NEW]` `internal/api/handlers/device.go` — `DeviceHandler.Register`, `DeviceHandler.List`
- `[MODIFY]` `internal/api/routes.go` — add device routes under protected group

## Implementation Steps

### `internal/device/repository.go`
```go
type DeviceRow struct { ID, IMEI, UserID, Name string; LastSeenAt, CreatedAt time.Time }

func RegisterDevice(ctx, db, userID, imei, name string) (DeviceRow, error)
// INSERT INTO devices (imei, user_id, name) VALUES ($1, $2, $3)
// ON CONFLICT (imei) DO NOTHING — reject duplicate IMEI across users

func ListDevices(ctx, db, userID string) ([]DeviceRow, error)
// SELECT id, imei, name, last_seen_at, created_at FROM devices WHERE user_id = $1
```

### `internal/api/handlers/device.go`
1. Parse + validate IMEI from JSON body (must be 15 digits).
2. Call `device.RegisterDevice`.
3. On duplicate IMEI conflict → return HTTP 409 Conflict.
4. On success → return HTTP 201 with device JSON.

### `internal/api/routes.go`
```go
devHandler := handlers.NewDeviceHandler(db)
protected.POST("/devices", devHandler.Register)
protected.GET("/devices", devHandler.List)
```

## Todo List
- [ ] Create `internal/device/repository.go`
- [ ] Create `internal/api/handlers/device.go`
- [ ] Update `internal/api/routes.go` with device routes
- [ ] Validate IMEI format in handler (15 digits regex)

## Success Criteria
- `POST /devices` with valid session + `{"imei":"353456789012345","name":"Watch"}` → 201, device row in DB
- `GET /devices` returns list of user's devices
- TCP server accepts `IWAP00353456789012345#` and successfully looks up the registered device
- `POST /devices` with duplicate IMEI → 409 Conflict

## Security Considerations
- Validate IMEI is exactly 15 numeric characters — reject everything else.
- On duplicate IMEI: if IMEI already registered to a different user, return 409 (do not reveal whose).
- Route is under `AuthMiddleware` — all unauthenticated requests already return 401 before reaching handler.
