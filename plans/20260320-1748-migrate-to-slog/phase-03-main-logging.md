# Phase 3: Core Refactor (cmd/api)

## Overview
- **Priority:** P1
- **Status:** Pending
- Replace all standard `log` package calls within `cmd/api/main.go` with structured `slog` calls.

## Requirements
- No structured `log` package calls remaining in `main.go`.
- Use correct log levels:
    - `log.Fatalf` -> `slog.Error` + `os.Exit(1)`.
    - `log.Println` / `log.Printf` (Startup) -> `slog.Info`.
    - `log.Println` / `log.Printf` (Errors) -> `slog.Error`.

## Related Code Files
- [cmd/api/main.go](file:///home/tuantt/projects/health-data-platform/cmd/api/main.go)

## Implementation Steps
1. Scan `main.go` for all `log.` references.
2. Replace startup logs: `"HTTP/TCP server starting..."` with `slog.Info`.
3. Replace fatal logs: `"Failed to connect to DB..."` with `slog.Error` and `os.Exit(1)`.
4. Ensure error context is passed: `slog.Error("Database connection failed", "err", err)`.

## Success Criteria
- Native `log` is completely removed from `main.go` imports.
- Logs include identifiable context (e.g., `addr`, `err`).
