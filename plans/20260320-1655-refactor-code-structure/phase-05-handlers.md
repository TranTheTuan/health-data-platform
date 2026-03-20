# Phase 5: Request/Response Handlers

## Overview
- **Priority:** P1
- **Status:** Pending
- Establish cleanly bounded request `handlers` that sit between `delivery` (network) and `service` (logic).

## Requirements
- Translates specific framework formats (Echo Context params, TCP scanner frames) into variables (DTOs) suitable for `service` methods.
- Invoked explicitly by the `delivery` routes.
- Format exceptions and errors back to outgoing responses (DTOs) for HTTP/TCP.
- **CRITICAL**: Handlers MUST ONLY interact with Services and DTOs. They should NEVER handle Domain Entities directly.

## Related Code Files
- Migrate: `internal/api/handlers/` to `internal/handler/http`
- Migrate: Logic inside `internal/tcp/handler.go` to `internal/handler/tcp`

## Implementation Steps
1. Create `internal/handler/http` and `internal/handler/tcp`.
2. Implement HTTP handlers: `type DeviceHandler struct { svc service.DeviceService }`.
3. Provide Echo methods: `func (h *DeviceHandler) RegisterDevice(c echo.Context) error`.
4. Parse Echo `c.Bind()` results into `dto.RegisterDeviceRequest`.
5. Call `h.svc.RegisterDevice(c.Request().Context(), requestDto)`.
6. Return `c.JSON(http.StatusOK, responseDto)`.
7. Implement TCP handlers similarly by parsing stream into frames and calling services.

## Success Criteria
- Handlers are thin wrappers that convert framework inputs to DTOs and back.
- No direct references to Domain models here.
