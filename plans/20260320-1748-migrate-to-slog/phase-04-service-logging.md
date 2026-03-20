# Phase 4: Service Layer Insight

## Overview
- **Priority:** P1
- **Status:** Pending
- Enhance debugging and visibility by adding structured logs to critical business services within `internal/service`.

## Requirements
- Maintain service-level awareness of key milestones (e.g., device registration, successful authentication).
- Use `slog.Debug` for verbose development-only details. (Only visible in `DEV`).
- Use `slog.Info` for production-ready business events. (Visible in both `DEV` and `PRODUCT`).
- Use `slog.Error` for critical issues.

## Related Code Files
- [internal/service/device_service.go](file:///home/tuantt/projects/health-data-platform/internal/service/device_service.go)
- [internal/service/auth_service.go](file:///home/tuantt/projects/health-data-platform/internal/service/auth_service.go)

## Implementation Steps
1. In `device_service.go`:
    - `slog.Info`: Successful device registration (include `imei`, `userID`).
    - `slog.Debug`: Detailed state updates (e.g., initial validation passed).
    - `slog.Warn`: Invalid registration attempts (IMEI format error).
2. In `auth_service.go`:
    - `slog.Info`: Successful OAuth user exchange.
    - `slog.Error`: Exchange failures with the external Google API.

## Success Criteria
- Services provide structured, searchable logs of core actions.
- Use `slog.Group` where appropriate to group related attributes (e.g., `"device", slog.Group("imei", imei, "name", name))`.
