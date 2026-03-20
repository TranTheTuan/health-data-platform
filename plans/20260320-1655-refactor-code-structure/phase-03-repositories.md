# Phase 3: Data Repositories

## Overview
- **Priority:** P1
- **Status:** Pending
- Consolidate Postgres logic inside `internal/repository`. Define the repository interfaces inside this package.

## Requirements
- Defines the `DeviceRepository` interface that other packages can reference.
- Implements the interface privately (e.g., `type pgDeviceRepo struct{}`).
- Methods translate SQL tuples directly into `internal/domain` structs (Entities).
- **CRITICAL**: Repositories MUST only work with Domain Entities, never DTOs.

## Related Code Files
- Delete: `internal/tcp/repository.go`
- Delete: `internal/device/repository.go`
- Create: `internal/repository/device.go`
- Create: `internal/repository/user.go`

## Implementation Steps
1. Create `internal/repository` folder.
2. `type DeviceRepository interface { GetByIMEI(...) (domain.Device, error) }`.
3. Provide an instantiation factory: `func NewDeviceRepository(db *sql.DB) DeviceRepository`.
4. Copy existing queries (`db.QueryRow()`, `db.Exec()`) from legacy scattered paths into this consolidated package.
