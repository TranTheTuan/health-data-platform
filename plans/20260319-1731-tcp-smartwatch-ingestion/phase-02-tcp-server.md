# Phase 2: TCP Server Core

**Context Links**
- [Main Plan](./plan.md)
- [Phase 1: DB Schema](./phase-01-db-schema.md)

## Overview
- Priority: P1
- Effort: 2.5h
- Status: Pending
- Implement the core TCP listener, connection handler, device auth, and DB write logic.

## Key Insights
- One goroutine per connection — acceptable for wearable device scale (hundreds, not millions).
- Must set read deadlines (`conn.SetDeadline`) to eliminate zombie connections.
- Newline-delimited text protocol — simplest possible parsing, smartwatch-friendly.
- Device lookup is a single indexed query; cache in connection state to avoid per-message DB hit.
- Batch writes are a future optimisation — single inserts are fine for Phase 1 (YAGNI).

## Protocol Spec
```
Client → Server:
  CONNECT <device_token>\n    # first message, required
  DATA <payload_text>\n       # one or more, after auth

Server → Client:
  OK\n                        # auth success, connection open
  ERROR <reason>\n            # auth failure or bad message
```

## Requirements
- Listen on a configurable port (default `:9090`)
- Reject connections that don't send `CONNECT` within 10s
- Authenticate device via `device_token` lookup in DB
- For each `DATA` message, insert a row into `health_records`
- Graceful shutdown via `context.Context` cancel

## Architecture

```
main.go
  └── tcp.Server.Start(ctx)
        └── net.Listen("tcp", addr)
              └── goroutine: handleConnection(conn, db)
                    ├── Read "CONNECT <token>"
                    ├── LookupDevice(token) → device{id, user_id}
                    ├── Reply "OK"
                    └── Loop: Read "DATA <text>" → InsertHealthRecord
```

## Related Code Files
- `[NEW]` `internal/tcp/server.go` — TCP server struct, listener, graceful shutdown
- `[NEW]` `internal/tcp/handler.go` — per-connection logic (auth + data loop)
- `[NEW]` `internal/tcp/repository.go` — DB queries: `LookupDevice`, `InsertHealthRecord`
- `[MODIFY]` `configs/config.go` — add `TCPAddr` field (e.g. `:9090`)
- `[MODIFY]` `.env` — add `TCP_ADDR=:9090`

## File Size Note
Each file must stay ≤ 200 LOC. Splitting `server.go` + `handler.go` + `repository.go` keeps concerns separate.

## Implementation Steps

### `internal/tcp/repository.go`
1. Define `DeviceRecord struct { ID, UserID string }`.
2. `LookupDevice(ctx, db, token) (DeviceRecord, error)` — query `devices` by `device_token`.
3. `InsertHealthRecord(ctx, db, deviceID, userID, data string) error` — insert into `health_records`.

### `internal/tcp/handler.go`
1. `HandleConnection(conn net.Conn, db *sql.DB)` function.
2. Set `conn.SetDeadline(time.Now().Add(10s))` immediately.
3. Read first line → parse `CONNECT <token>`. On failure → write `ERROR\n`, close.
4. Call `LookupDevice`. On failure → write `ERROR unknown device\n`, close.
5. Reset deadline to rolling 5-minute idle timeout.
6. Write `OK\n`.
7. Loop: read line → parse `DATA <text>` → call `InsertHealthRecord` → continue.
8. On any read error or timeout → close connection cleanly.

### `internal/tcp/server.go`
1. `Server struct { addr string; db *sql.DB }`.
2. `NewServer(addr, db)` constructor.
3. `Start(ctx context.Context) error` — `net.Listen`, accept loop, spawn goroutines.
4. On `ctx.Done()` → close listener to break accept loop.

## Todo List
- [ ] Create `internal/tcp/repository.go`
- [ ] Create `internal/tcp/handler.go`
- [ ] Create `internal/tcp/server.go`
- [ ] Update `configs.Config` with `TCPAddr`
- [ ] Update `.env`

## Success Criteria
- `go build ./...` succeeds
- Manual test: `echo -e "CONNECT bad_token\n" | nc localhost 9090` → returns `ERROR`
- Manual test: connect with valid token → returns `OK`, sends `DATA hello` → row appears in DB

## Risk Assessment
- **Zombie connections**: mitigated by read deadline + idle timeout.
- **Concurrent DB writes**: `database/sql` pool handles this safely.
- **Data flooding**: no rate limiting in Phase 1; add in future if needed.

## Security Considerations
- Device tokens treated as secrets — never log the raw token value.
- Read timeout prevents slow-loris attacks.
- Validate `DATA` payload is non-empty text before inserting.

## Next Steps
- [Phase 3: Integration & Config](./phase-03-integration.md)
