# Phase 6: Transport Deliveries

## Overview
- **Priority:** P1
- **Status:** Pending
- Establish `internal/delivery` focused exclusively on network server lifetimes, framework configuration, and socket generation.

## Requirements
- Extremely "dumb" infrastructural code.
- Contains the Echo App initialization and middleware attachments for HTTP HTTP.
- Contains the `net.Listen` and graceful context cancellation logic for TCP loop.

## Related Code Files
- Migrate: `internal/api/routes.go` to `internal/delivery/http_router.go`.
- Migrate: `internal/tcp/server.go` to `internal/delivery/tcp_server.go`.

## Implementation Steps
1. Create `internal/delivery/http` and `internal/delivery/tcp`.
2. Build `RegisterHttpRoutes(app *echo.Echo, authHandler, devHandler)`.
3. Adapt the old TCP Server struct to accept the new `handler.TCPConnectHandler`. Once `Accept()` passes a connection socket, it injects it into the handler to decode packets.
