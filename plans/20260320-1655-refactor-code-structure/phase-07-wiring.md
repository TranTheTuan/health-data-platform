# Phase 7: Dependency Injection Wiring

## Overview
- **Priority:** P1
- **Status:** Pending
- Build the final executable composition inside `cmd/api/main.go`.

## Requirements
- Sequentially inject the dependency graph from bottom-up: `Repo -> Service -> Handler -> Delivery`.

## Implementation Steps
1. Establish Postgres `sql.DB`.
2. Map Repositories:
   ```go
   devRepo := repository.NewDeviceRepository(db)
   ```
3. Map Services injecting Repos:
   ```go
   devSvc := service.NewDeviceService(devRepo)
   ```
4. Map Handlers injecting Services:
   ```go
   devHttpHandler := http_handler.NewDeviceHandler(devSvc)
   devTcpHandler := tcp_handler.NewConnectHandler(devSvc)
   ```
5. Pass to Deliveries and Boot:
   ```go
   e := echo.New()
   http_delivery.RegisterRoutes(e, devHttpHandler)

   tcpSrvr := tcp_delivery.NewServer(":9090", devTcpHandler)
   ```

## Success Criteria
- Zero cyclic dependencies.
- Dual-server runs perfectly.
- Full DTO-to-Domain isolation achieved via the layering chain.
