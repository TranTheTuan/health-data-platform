# Phase 04 — IMEI/Device ID + UI Updates

## Context Links
- Device domain: `internal/domain/device.go`
- Device DTO: `internal/dto/device_dto.go`
- Service IMEI validation: `internal/service/device_service.go` (line 21: `reIMEI` regex)
- DB migration: `internal/migrations/001_initial_schema.up.sql`
- Packet inspector UI: `web/templates/packets.html` (filter dropdown lines 109-123)
- Dashboard: `web/templates/dashboard.html`

## Overview
- **Priority:** P2
- **Status:** Complete
- **Description:** Update IMEI validation to accept 10-digit device IDs, update the UI filter dropdown to use Wonlex command codes, and create a migration to widen the IMEI column.

## Key Insights
- Currently: IMEI validation enforces exactly 15 decimal digits (`^\d{15}$`)
- New: device ID is 10 digits. User should register using the 10-digit ID printed on the watch
- DB `imei` column is `VARCHAR(15)`, which already fits 10 chars — no schema change needed for width
- But the UNIQUE constraint on `imei` remains valid — that's good
- Domain comment says "15-digit unique identifier" — update to reflect 10-digit
- UI command type filter dropdown still shows old IW commands (AP00, AP01, etc.) — must replace with Wonlex (UD, AL, LK, UD2, CONFIG)

## Requirements

### Functional
- `reIMEI` regex: change from `^\d{15}$` to `^\d{10,15}$` (accept both 10 and 15)
  - Allows backward compatibility if some users already registered 15-digit IMEIs
- `ErrInvalidIMEI` message: update to "device ID must be 10-15 decimal digits"
- Domain comment: update to reflect variable-length ID
- UI filter dropdown: replace old IW command options with Wonlex commands
- DB migration `002_`: optional — add functional index on `RIGHT(imei, 10)` for performance

### Non-functional
- No breaking changes for existing 15-digit IMEIs in DB
- Filter dropdown matches actual protocol commands

## Architecture

### Service changes
```go
// device_service.go
var reIMEI = regexp.MustCompile(`^\d{10,15}$`)
var ErrInvalidIMEI = errors.New("device ID must be 10-15 decimal digits")
```

### UI filter dropdown replacement
```html
<select id="filterType">
    <option value="">All Types</option>
    <option value="LK">LK - Link/Keep-alive</option>
    <option value="UD">UD - Location Report</option>
    <option value="UD2">UD2 - Blind Spot Data</option>
    <option value="AL">AL - Alarm Report</option>
    <option value="CONFIG">CONFIG - Device Config</option>
</select>
```

### Optional migration
```sql
-- 002_add_imei_suffix_index.up.sql
CREATE INDEX IF NOT EXISTS idx_devices_imei_suffix10
    ON devices (RIGHT(imei, 10));
```

## Related Code Files

| File | Action | Description |
|------|--------|-------------|
| `internal/service/device_service.go` | **MODIFY** | Relax IMEI regex to 10-15 digits |
| `internal/domain/device.go` | **MODIFY** | Update comment |
| `web/templates/packets.html` | **MODIFY** | Replace command filter options |
| `internal/migrations/002_add_imei_suffix_index.up.sql` | **CREATE** | Functional index for 10-digit lookup |

## Implementation Steps

1. Update `internal/service/device_service.go`:
   - `reIMEI` regex → `^\d{10,15}$`
   - `ErrInvalidIMEI` message → "device ID must be 10-15 decimal digits"
2. Update `internal/domain/device.go`:
   - Comment on `IMEI` field → "10-15 digit device identifier"
3. Update `web/templates/packets.html`:
   - Replace all `<option>` values in `#filterType` dropdown with Wonlex command codes
4. Create `internal/migrations/002_add_imei_suffix_index.up.sql`:
   - Add functional index for `RIGHT(imei, 10)`
5. Run `go build ./...` and `go test ./...`

## Todo List

- [x] Relax IMEI validation in `device_service.go`
- [x] Update domain model comment
- [x] Replace UI filter dropdown with Wonlex commands
- [x] Create migration `002_add_imei_suffix_index.up.sql`
- [x] Verify: `go build ./...` and `go test ./...`
- [x] Manual test: register device with 10-digit ID, open packet inspector

## Success Criteria

- Can register a device with 10-digit ID via web dashboard
- Packet inspector filter dropdown shows Wonlex commands
- `RIGHT(imei, 10)` index exists for TCP handler lookups
- All existing 15-digit IMEI devices continue working
- `go build ./...` and `go test ./...` pass

## Risk Assessment

| Risk | Mitigation |
|------|-----------|
| Existing 15-digit devices can't connect | TCP handler uses `RIGHT(imei, 10)` match — works for both |
| 10-digit ID collides with last 10 of existing 15-digit | Highly unlikely — IMEI structure prevents this |
