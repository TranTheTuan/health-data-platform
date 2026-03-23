# Phase 2: Service Layer & DTOs

## Overview
- **Priority:** P1
- **Status:** Pending
- **Goal:** Add DTO requests and responses specifically for the new `ListDevicePackets` API endpoint. Extend `DeviceService` to handle parsing requirements and formatting the payload for UI delivery.

## Core Requirements
1. DTO `ListPacketsRequest` to encapsulate pagination data, time filters, and packet types.
2. DTO `PacketResponse` to map domain packets to stringified dates and cleanly exposed raw commands.
3. DTO `PaginatedPacketResponse` wrapping the `PacketResponse` array and providing a total count.

## Related Code Files
- [internal/dto/packet_dto.go](file:///home/tuantt/projects/health-data-platform/internal/dto/packet_dto.go)
- [internal/service/device_service.go](file:///home/tuantt/projects/health-data-platform/internal/service/device_service.go)

## Implementation Steps
1. Add to `packet_dto.go`:
   ```go
   type ListPacketsRequest struct {
       DeviceID   string `query:"-"` // Passed from path param
       PacketType string `query:"type"`
       From       string `query:"from"` // expected time ISO string
       To         string `query:"to"`
       Limit      int    `query:"limit"`
       Offset     int    `query:"offset"`
   }
   
   type PacketResponse struct {
       ID          string    `json:"id"`
       CommandCode string    `json:"command_code"`
       RawPayload  string    `json:"raw_payload"`
       CreatedAt   string    `json:"created_at"`
   }

   type PaginatedPacketResponse struct {
       Packets []PacketResponse `json:"packets"`
       Total   int              `json:"total"`
   }
   ```
2. In `DeviceService` interface: add `ListDevicePackets(ctx context.Context, userID string, req dto.ListPacketsRequest) (dto.PaginatedPacketResponse, error)`.
3. In `deviceService.ListDevicePackets` implementation:
   - Apply logical defaults: limit 10, max limit 100.
   - Parse `From` and `To` time strings (if provided) into `time.Time`.
   - Forward values into `repo.ListPackets`.
   - Log errors via `slog.Error` on parse or query failure.
   - Map returned `[]domain.Packet` into `dto.PaginatedPacketResponse`.

## Success Criteria
- Time strings are safely ignored or rejected natively if badly formatted.
- Limits are clamped (limit > 100 becomes 100, limit <= 0 becomes 10).
- Array slices map safely, avoiding panic conditions on empty results.
