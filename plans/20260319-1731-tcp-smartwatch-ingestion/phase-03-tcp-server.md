# Phase 3: TCP Server Core

**Context Links**
- [Main Plan](./plan.md)
- [Phase 2: Protocol Parser](./phase-02-protocol-parser.md)

## Overview
- Priority: P1
- Effort: 2.5h
- Status: Pending
- Implement the TCP listener, per-connection goroutines, IMEI auth using AP00, packet routing, DB writes, and reply sending.

## Key Insights
- One goroutine per connection: acceptable for wearable-scale (hundreds concurrent, not millions).
- Authentication enforced: first frame MUST be `AP00`. Any other first frame → close connection.
- After auth, loop: read frame → parse → lookup/insert → reply. No request-response pairing complexity (one-frame-at-a-time).
- DB write is synchronous per frame. Batch optimization is a future concern (YAGNI).
- Must set **read deadline** on every frame read to kill zombie connections.

## Packet Routing Table

| Cmd | Handler action | DB write | Reply |
|-----|---------------|----------|-------|
| `AP00` | Lookup device by IMEI, update `last_seen_at` | UPDATE | `IWBP00;<ts>,0#` |
| `AP01`, `AP02` | Insert location packet | INSERT | `IWBPxx#` |
| `AP03` | Heartbeat — no DB write needed | — | `IWBP03#` |
| `AP49`, `APHT`, `APHP`, `AP50` | Insert health metric packet | INSERT | `IWBPxx#` |
| `AP97` | Insert sleep packet | INSERT | `IWBP97#` |
| `AP07` | Insert audio packet (raw payload stored) | INSERT | `IWBP07#` |
| `AP10` | Insert alarm packet | INSERT | `IWBP10#` |
| `APWT` | Weather sync — reply only (no data to store) | — | `IWBPWT#` |
| `APHD` | Insert ECG packet | INSERT | `IWBPHD#` |
| Unknown | Log warning | — | close |

## Related Code Files
- `[NEW]` `internal/tcp/repository.go` — DB queries: `LookupDeviceByIMEI`, `UpdateLastSeen`, `InsertPacket`
- `[NEW]` `internal/tcp/handler.go` — per-connection auth+routing logic
- `[NEW]` `internal/tcp/server.go` — listener + graceful shutdown

## File Size Note
Three files, each ≤200 LOC. Clean concern separation: repo ↔ handler ↔ server.

## Implementation Steps

### `internal/tcp/repository.go`
```go
type DeviceRecord struct { ID, UserID string }

// LookupDeviceByIMEI finds device by IMEI; returns ErrNotFound if absent
func LookupDeviceByIMEI(ctx, db, imei) (DeviceRecord, error)

// UpdateLastSeen sets last_seen_at = NOW() for a device
func UpdateLastSeen(ctx, db, deviceID string) error

// InsertPacket stores a device packet
func InsertPacket(ctx, db, deviceID, userID, packetType, rawPayload string, parsedData interface{}) error
```

Key query for InsertPacket:
```sql
INSERT INTO device_packets (device_id, user_id, packet_type, raw_payload, parsed_data)
VALUES ($1, $2, $3, $4, $5::jsonb)
```

### `internal/tcp/handler.go`
```go
func HandleConnection(conn net.Conn, db *sql.DB) {
    defer conn.Close()
    scanner := bufio.NewScanner(conn)
    scanner.Split(protocol.ScanFrame)

    // 1. Auth: first frame must be AP00
    conn.SetDeadline(time.Now().Add(10 * time.Second))
    if !scanner.Scan() { return }
    frame, _ := protocol.ParseFrame(scanner.Text())
    if frame.Cmd != protocol.CmdLogin { conn.Write([]byte("IWBP00ERROR#")); return }
    imei, _ := protocol.ParseAP00(frame.Payload)

    device, err := repo.LookupDeviceByIMEI(ctx, db, imei)
    if err != nil { return } // unknown IMEI = silent drop

    repo.UpdateLastSeen(ctx, db, device.ID)
    reply := protocol.BuildReply(frame.Cmd)
    conn.Write([]byte(reply))

    // 2. Data loop
    conn.SetDeadline(time.Now().Add(5 * time.Minute)) // idle timeout
    for scanner.Scan() {
        conn.SetDeadline(time.Now().Add(5 * time.Minute)) // reset on each frame
        frame, err := protocol.ParseFrame(scanner.Text())
        if err != nil { continue }
        if frame.Cmd != protocol.CmdHeartbeat && frame.Cmd != protocol.CmdWeather {
            repo.InsertPacket(ctx, db, device.ID, device.UserID, frame.Cmd, frame.Payload, nil)
        }
        conn.Write([]byte(protocol.BuildReply(frame.Cmd)))
    }
}
```

### `internal/tcp/server.go`
```go
type Server struct { addr string; db *sql.DB }

func (s *Server) Start(ctx context.Context) error {
    ln, _ := net.Listen("tcp", s.addr)
    go func() { <-ctx.Done(); ln.Close() }()
    for {
        conn, err := ln.Accept()
        if err != nil { return err } // ctx cancel closes ln → returns here
        go HandleConnection(conn, s.db)
    }
}
```

## Todo List
- [ ] Create `internal/tcp/repository.go`
- [ ] Create `internal/tcp/handler.go`
- [ ] Create `internal/tcp/server.go`

## Success Criteria
- `go build ./...` succeeds
- Manual test with `netcat`:
  - `echo -n "IWAP00INVALID#" | nc localhost 9090` → server silently drops
  - `echo -n "IWAP00<valid_imei>#" | nc localhost 9090` → server replies `IWBP00;...#`
  - After login, `echo -n "IWAP03#" | nc localhost 9090` → replies `IWBP03#` (heartbeat)
  - After login, `echo -n "IWAP49<data>#" | nc localhost 9090` → row in `device_packets`

## Risk Assessment
- **Partial reads**: handled by `bufio.Scanner` + `ScanFrame` (`#`-split).
- **Unknown IMEI**: silent drop (no error reply that leaks info).
- **Frame size DoS**: limit scanner buffer to 64KB max.
- **Concurrent DB writes**: `database/sql` pool handles safely.

## Security Considerations
- Never log raw IMEI in plaintext (it's a hardware ID, not a secret, but still PII).
- Auth timeout: 10s for first frame to prevent slow-open attacks.
- Idle timeout: 5 minutes resets on each valid frame.

## Next Steps
- [Phase 4: Integration & Config](./phase-04-integration.md)
