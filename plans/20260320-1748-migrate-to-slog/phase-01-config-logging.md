# Phase 1: Configuration Update

## Overview
- **Priority:** P1
- **Status:** Pending
- Extend the application configuration to support the environment modes: `DEV` and `PRODUCT`.

## Requirements
- Support `ENVIRONMENT` environment variable:
    - `DEV`: Development mode.
    - `PRODUCT`: Production mode.
- Default to `DEV` if not specified.
- High consensus for JSON-only formatting; no format environment variable needed.

## Related Code Files
- [configs/config.go](file:///home/tuantt/projects/health-data-platform/configs/config.go)

## Implementation Steps
1. Add `Environment` field to the `Config` struct.
2. In `LoadConfig`, fetch `ENVIRONMENT` using the `getEnv` helper.

## Success Criteria
- Environment variable `ENVIRONMENT` successfully mapped to the `Config` struct.
