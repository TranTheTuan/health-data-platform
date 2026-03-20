# Phase 2: Logger Initialization

## Overview
- **Priority:** P1
- **Status:** Pending
- Setup the global `log/slog` logger with **JSON format ALWAYS**, and environment-based levels.

## Requirements
- Format: Always **JSON**.
- Level (Environment-driven):
    - `DEV` (default): `slog.LevelDebug`.
    - `PRODUCT`: `slog.LevelInfo`.

## Related Code Files
- [cmd/api/main.go](file:///home/tuantt/projects/health-data-platform/cmd/api/main.go)

## Implementation Steps
1. Create a `initLogger(cfg *configs.Config)` function in `main.go`.
2. Define the log level mapping:
    - If `cfg.Environment == "PRODUCT"`, `level = slog.LevelInfo`.
    - Else, `level = slog.LevelDebug`.
3. Initialize the handler:
    ```go
    handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: level,
    })
    ```
4. Set the global logger: `slog.SetDefault(slog.New(handler))`.

## Success Criteria
- Global logger correctly set up using JSON format.
- `DEBUG` logs correctly visible in `DEV` mode.
- Only `INFO`+ logs visible in `PRODUCT` mode.
