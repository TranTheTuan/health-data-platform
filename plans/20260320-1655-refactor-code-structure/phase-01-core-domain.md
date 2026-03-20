# Phase 1: Core Domain Models

## Overview
- **Priority:** P1
- **Status:** Pending
- Establish `internal/domain` package as a pure, dependency-less struct definition layer.

## Requirements
- Must contain *only* basic Go struct types (e.g., `Device`, `User`).
- Must **not** define interfaces. (Interfaces belong to higher-level operational packages).
- Dependencies allowed: Almost zero. Maybe `time`. No external code or project packages.

## Related Code Files
- Create: `internal/domain/device.go`
- Create: `internal/domain/user.go`
- Create: `internal/domain/auth.go`

## Implementation Steps
1. Create `internal/domain` folder.
2. Relocate data structs currently defined in `internal/tcp/repository.go` and `internal/auth/...` into this neutral layer.
3. Rename the structs to crisp domain nouns (e.g., from `DeviceRecord` to `Device`).
4. Ensure no cross-package imports exist here.
