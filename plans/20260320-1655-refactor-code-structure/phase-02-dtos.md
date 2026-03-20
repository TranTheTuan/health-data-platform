# Phase 2: Data Transfer Objects (DTOs)

## Overview
- **Priority:** P1
- **Status:** Pending
- Establish `internal/dto` package to define Data Transfer Objects (DTOs).
- These will represent the data entering the system (Requests) and leaving the system (Responses) via APIs.

## Requirements
- DTOs should have JSON/XML/etc. tags as needed for transport serialization.
- Must **not** be tied to database schemas. (Independent from `internal/domain`).
- No methods on DTOs, except maybe for validation helper methods.

## Related Code Files
- Create: `internal/dto/device_dto.go`
- Create: `internal/dto/auth_dto.go`
- Create: `internal/dto/packet_dto.go`

## Implementation Steps
1. Create `internal/dto` directory.
2. Define Request DTOs: e.g., `type RegisterDeviceRequest struct { IMEI string }`.
3. Define Response DTOs: e.g., `type DeviceResponse struct { ID string, IMEI string, LastSeen string }`.
4. Ensure transport-specific metadata (JSON tags) is correctly applied here, not in `internal/domain`.

## Success Criteria
- Request/Response structures are isolated from Core Domain Entities.
- No direct references to `*sql.DB` or other transport frameworks here.
