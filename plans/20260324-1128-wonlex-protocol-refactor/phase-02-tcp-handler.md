# Phase 02 — TCP Handler Refactor

## Context Links
- Current handler: `internal/handler/tcp/handler.go` (124 LOC)
- New parser API: `protocol.ParseFrame(raw)` returns `Frame{Manufacturer, DeviceID, Length, Cmd, Payload}`
- New reply API: `protocol.BuildReply(deviceID, cmd)` returns reply string or `""`

## Overview
- **Priority:** P1
- **Status:** Complete
- **Description:** Refactor `HandleConnection` to use the new Wonlex `Frame` struct, authenticate by 10-digit device ID, and handle the new reply format.

## Key Insights
- Old flow: first frame must be `AP00` login with 15-digit IMEI → lookup by exact IMEI match
- New flow: **every frame contains `DeviceID`** (10 digits) in its header → no special "login" handshake needed
- First frame: use `frame.DeviceID` to lookup device in DB → this IS the authentication
- `LK` (link/keep-alive) IS the equivalent of old login — device sends it periodically and expects a reply
- For `UD` and `UD2`: save packet but do NOT send reply
- For `AL`: save packet and send `[3G*deviceID*LEN*AL]`
- For `CONFIG`: save packet and send `[CS*deviceID*LEN*CONFIG,1]`
- `shouldPersist`: all commands are persisted (save raw payload + command for every packet received)

## Requirements

### Functional
- First frame: parse → extract `DeviceID` → `repo.LookupDeviceByIMEI(deviceID)` → if not found, close
- After auth: loop frames, persist ALL packets, conditionally send reply
- `reply := protocol.BuildReply(frame.DeviceID, frame.Cmd)` — if non-empty, write to conn
- Save raw payload including the IMEI (device ID) which is now 10 chars

### Non-functional
- `handler.go` ≤ 130 LOC
- Use `slog` for error logging (not `log.Printf`)

## Architecture

### Updated HandleConnection pseudo-flow

```go
func (h *TCPConnectHandler) HandleConnection(conn net.Conn) {
    defer conn.Close()
    ctx := context.Background()

    scanner := bufio.NewScanner(conn)
    scanner.Split(protocol.ScanFrame)
    scanner.Buffer(make([]byte, maxFrameBytes), maxFrameBytes)
    
    // Auth: read first frame, extract deviceID
    conn.SetDeadline(time.Now().Add(authTimeout))
    if !scanner.Scan() { return }
    
    frame, err := protocol.ParseFrame(scanner.Text())
    if err != nil { return }
    
    device, err := h.svc.LookupDeviceByIMEI(ctx, frame.DeviceID)
    if err != nil { return }
    
    h.svc.UpdateLastSeen(ctx, device.ID)
    
    // Reply to first frame if needed
    if reply := protocol.BuildReply(frame.DeviceID, frame.Cmd); reply != "" {
        conn.Write([]byte(reply))
    }
    
    // Persist first frame
    h.persistPacket(ctx, device, frame)
    
    // Main loop
    conn.SetDeadline(time.Now().Add(idleTimeout))
    for scanner.Scan() {
        conn.SetDeadline(time.Now().Add(idleTimeout))
        
        frame, err := protocol.ParseFrame(scanner.Text())
        if err != nil { continue }
        
        h.persistPacket(ctx, device, frame)
        
        if reply := protocol.BuildReply(frame.DeviceID, frame.Cmd); reply != "" {
            if _, err := conn.Write([]byte(reply)); err != nil { return }
        }
    }
}

func (h *TCPConnectHandler) persistPacket(ctx context.Context, device domain.Device, frame protocol.Frame) {
    req := dto.IngestPacketRequest{
        DeviceID:    device.ID,
        UserID:      device.UserID,
        CommandCode: frame.Cmd,
        RawPayload:  frame.Payload,
    }
    if err := h.svc.ProcessPacket(ctx, req); err != nil {
        slog.Error("tcp: insert packet failed", slog.String("device_id", device.ID), slog.String("cmd", frame.Cmd), slog.Any("error", err))
    }
}
```

### LookupDeviceByIMEI adaptation
- Currently queries `WHERE imei = $1` with a 15-digit string
- Need to support 10-digit device ID lookup
- Option: `WHERE imei LIKE '%' || $1` (suffix match) — fragile
- **Recommended:** Store the 10-digit device ID alongside the IMEI, OR change registration to accept 10 digits. This is handled in Phase 04.
- **For now in Phase 02:** change the repo query to match the LAST 10 digits: `WHERE RIGHT(imei, 10) = $1`
  - This is safe because device IDs are unique within the last 10 digits

## Related Code Files

| File | Action | Description |
|------|--------|-------------|
| `internal/handler/tcp/handler.go` | **REWRITE** | New Wonlex connection handler |
| `internal/repository/device.go` | **MODIFY** | `LookupDeviceByIMEI` to support 10-digit suffix match |

## Implementation Steps

1. Rewrite `handler.go`:
   - Remove `ParseAP00` call (no longer exists)
   - Use `frame.DeviceID` for auth instead of payload-based IMEI
   - Remove `shouldPersist` filter — persist everything
   - Use `protocol.BuildReply(deviceID, cmd)` and only write if non-empty
   - Switch from `log.Printf` to `slog.Error`
2. Update `LookupDeviceByIMEI` in `repository/device.go`:
   - Change query to `WHERE RIGHT(imei, 10) = $1` for 10-digit match
3. Run `go build ./...` to verify compile

## Todo List

- [x] Rewrite `internal/handler/tcp/handler.go`
- [x] Modify `LookupDeviceByIMEI` query in `internal/repository/device.go`
- [x] Verify compile: `go build ./...`

## Success Criteria

- Handler correctly parses Wonlex frames
- Device lookup works with 10-digit ID matching last 10 digits of stored IMEI  
- Replies only sent for commands that require them
- All packets persisted with correct command code and raw payload
- `go build ./...` passes

## Risk Assessment

| Risk | Mitigation |
|------|-----------|
| `RIGHT(imei, 10)` won't use index | Acceptable for current scale; add functional index in phase 04 if needed |
| Device sends unknown command | Parser returns error, handler skips with `continue` |
