# Phase 3: API Delivery

## Overview
- **Priority:** P1
- **Status:** Pending
- **Goal:** Expose the paginated `DevicePacket` endpoint on `Echo` via a secure path and attach the required HTTP handlers.

## Core Requirements
1. New endpoint: `GET /protected/devices/:id/packets` that returns JSON `PaginatedPacketResponse`.
2. New page endpoint: `GET /dashboard/devices/:id/packets` that renders the HTML view.

## Related Code Files
- [internal/handler/http/device.go](file:///home/tuantt/projects/health-data-platform/internal/handler/http/device.go)
- [internal/delivery/http/router.go](file:///home/tuantt/projects/health-data-platform/internal/delivery/http/router.go)

## Implementation Steps
1. In `device.go`, add `ListPacketsAPI(c echo.Context) error`.
   - Parse `deviceID` from `:id` param.
   - Bind `c` into `ListPacketsRequest` implicitly or manually (Limit, Offset, From, To, PacketType).
   - Get authenticated user from `c.Get("user_id")` context (AuthMiddleware).
   - Call `devSvc.ListDevicePackets` using the struct.
   - If success, `c.JSON(http.StatusOK, response)`.
2. In `device.go`, add `PacketInspectPage(c echo.Context) error`.
   - Parse `deviceID` from `:id`.
   - Get user ID. Verify they own the device by calling `devSvc.LookupDeviceByIMEI` OR add an explicit `LookupDeviceByID` that does the validation. Wait, best logic:
      - Call `LookupDeviceByID` to ensure it exists and matches the user. If unauth, `Redirect` to `/dashboard`.
   - Render the `packets.html` template.
3. In `router.go`, update routes:
   ```go
   dashboardRoute.GET("/devices/:id/packets", dh.PacketInspectPage) // Use AuthMiddleware
   
   protected.GET("/devices/:id/packets", dh.ListPacketsAPI)
   ```

## Success Criteria
- Requesting an invalid device ID returns 404 or 403 on the API.
- Query parameters bind properly (i.e. `?type=AP01&limit=20&offset=0`).
- Viewing the page securely routes based on `AuthMiddleware`.
