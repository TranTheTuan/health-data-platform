# Phase 1: Database & Repository

## Overview
- **Priority:** P1
- **Status:** Pending
- **Goal:** Enable pagination, filtering, and counting for a device's packets while securing the queries by User ID.

## Core Requirements
1. The queries must restrict to both `device_id` and `user_id` to ensure an attacker cannot guess a device ID to view its packets.
2. The index `(device_id, packet_type, created_at DESC)` should be created if filtering by type is expected often.

## Related Code Files
- [internal/repository/device.go](file:///home/tuantt/projects/health-data-platform/internal/repository/device.go)

## Implementation Steps
1. Add to `DeviceRepository` interface: 
   ```go
   ListPackets(ctx context.Context, userID, deviceID string, packetType *string, from, to *time.Time, limit, offset int) ([]domain.Packet, int, error)
   ```
   *Note: Returning both `(`data`, `total_count`, `error`)` in one go via two queries (`SELECT count(*)` and `SELECT * ... LIMIT OFFSET`) is often acceptable given the pagination use case, or we use separate interfaces. Let's return total count alongside the packets natively.*
2. Inside `pgDeviceRepo.ListPackets` implementation:
   - Construct standard base query: `WHERE user_id = $1 AND device_id = $2`.
   - Append conditionally: `AND packet_type = $X` (if provided).
   - Append conditionally: `AND created_at >= $Y` (if `from` provided).
   - Append conditionally: `AND created_at <= $Z` (if `to` provided).
   - Make sure to count total first using the same conditions.
   - Run the final `ORDER BY created_at DESC LIMIT $L OFFSET $O`.
3. Optionally, write an initialization script to create composite database indices on `device_packets(device_id, created_at DESC)`.

## Success Criteria
- The repo safely blocks access if `userID` does not match the device owner.
- Filtering by packet type works correctly (including optional nil behavior).
- Time limits and offset/limit pagination fetch properly.
- Total count matches the filtered dataset length.
