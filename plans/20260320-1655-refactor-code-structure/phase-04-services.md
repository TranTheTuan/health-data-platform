# Phase 4: Business Services

## Overview
- **Priority:** P1
- **Status:** Pending
- Extract scattered business logic purely into `internal/service`. Define the service interfaces inside this package.
- Services will act as the "coordinator" layer, mapping between entities (domain) and DTOs (dto).

## Requirements
- Defines the `DeviceService` interface representing strict business use cases.
- Depends strictly on the `repository` interfaces to load and save Domain Entities.
- Performs mapping logic (e.g., from `domain.Device` to `dto.DeviceResponse`).
- Independent of presentation formats (HTTP context, TCP packet formatting).

## Related Code Files
- Create: `internal/service/device_service.go`
- Create: `internal/service/auth_service.go`
- Extract logic from `internal/api/handlers/device.go` and `internal/tcp/handler.go`.

## Implementation Steps
1. Create `internal/service`.
2. Define `type DeviceService interface { ... }` explicitly.
3. Provide a factory: `func NewDeviceService(repo repository.DeviceRepository) DeviceService`.
4. Methods accept/return DTOs from `internal/dto` as appropriate.
5. Mapping between DTOs and Domain Entities should happen within the methods here, never in the repository.
6. Validation rules (like confirming an IMEI length) are contained here before calling the repository.

## Success Criteria
- Services are entirely testable with standard Go mocks.
- Mapping logic is properly isolated within the service layer.
