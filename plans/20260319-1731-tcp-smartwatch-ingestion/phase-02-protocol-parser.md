# Phase 2: Protocol Parser

**Context Links**
- [Main Plan](./plan.md)
- [Phase 1: DB Schema](./phase-01-db-schema.md)

## Overview
- Priority: P1
- Effort: 1.5h
- Status: Pending
- Implement a standalone, testable parser for the `IW` protocol used by smartwatch devices. This is the most security-sensitive and correctness-critical component.

## Key Insights
- **Frame format:** `IW<CMD><PAYLOAD>#` where:
  - `IW` is always the 2-char prefix (constant)
  - `<CMD>` is 4 chars for most commands (`AP00`, `AP01`, `APHT`, `APHP`, `APHD`, `APWT`) — some are 4 chars
  - `<PAYLOAD>` is everything between CMD and `#`
  - `#` is the frame terminator — NOT a newline
- Devices may send fragmented TCP segments. The reader must accumulate bytes until `#` is received. Do NOT assume one `read()` = one full frame.
- The `#` terminator makes `bufio.Scanner` with a custom `ScanFunc` (split on `#`) the cleanest approach.
- Command code determines the reply string. Per-field parsing of payload is done on a best-effort basis into JSONB for Phase 1 (full deep parsing for each packet type can be phased in).

## Protocol Command Codes

```go
const (
    CmdLogin       = "AP00"  // IMEI login       → "IWBP00;<yyyyMMddHHmmss>,<tz>#"
    CmdGPSLoc      = "AP01"  // GPS location      → "IWBP01#"
    CmdLBSLoc      = "AP02"  // LBS location      → "IWBP02#"
    CmdHeartbeat   = "AP03"  // Keepalive         → "IWBP03#"
    CmdAudio       = "AP07"  // Audio upload      → "IWBP07#"
    CmdAlarm       = "AP10"  // Alarm             → "IWBP10#"
    CmdHeartRate   = "AP49"  // Heart rate        → "IWBP49#"
    CmdHRAndBP     = "APHT"  // HR + BP           → "IWBPHT#"
    CmdHRBPSPO2    = "APHP"  // HR + BP + SPO2 + glucose → "IWBPHP#"
    CmdTemperature = "AP50"  // Body temp         → "IWBP50#"
    CmdSleep       = "AP97"  // Sleep data        → "IWBP97#"
    CmdWeather     = "APWT"  // Weather sync      → "IWBPWT#"
    CmdECG         = "APHD"  // ECG upload        → "IWBPHD#"
)
```

## AP00 Login Payload Parsing
```
IWAP00353456789012345#
       └─────────────── IMEI (15 digits)

Server reply: IWBP00;20260319103000,0#
              └─────────────────────── UTC timestamp (YYYYMMDDHHmmss) + timezone offset
```

## Requirements
- `ParseFrame(raw string) (cmd, payload string, err error)` — strips `IW` prefix, extracts 4-char cmd + rest as payload
- `BuildReply(cmd string) string` — returns the correct server acknowledgment string incl. special BP00 timestamp logic
- `ParseAP00(payload string) (imei string, err error)` — extract and validate 15-digit IMEI
- A `Frame` struct: `{ Cmd, Payload string }`
- `ScanFrame` — a `bufio.SplitFunc` that splits on `#` for use with `bufio.Scanner`

## Related Code Files
- `[NEW]` `internal/tcp/protocol/parser.go` — frame parsing and reply building
- `[NEW]` `internal/tcp/protocol/parser_test.go` — table-driven unit tests

## Implementation Steps

### `internal/tcp/protocol/parser.go`
1. Define constants for all 13 command codes.
2. `ScanFrame(data []byte, atEOF bool) (advance, token []byte, err error)` — splits on `#`.
3. `ParseFrame(raw string) (Frame, error)`:
   - Validate `IW` prefix (return `ErrInvalidFrame` if missing).
   - Extract 4-char cmd code.
   - Rest is payload (may be empty for simple commands).
4. `ParseAP00(payload string) (string, error)`:
   - Trim whitespace.
   - Validate length == 15 and all digits.
   - Return IMEI string.
5. `BuildReply(cmd string) string`:
   - For `AP00`: return `IWBP00;<time.Now().UTC().Format("20060102150405")>,0#`
   - For all others: return `IW<cmd mapped to BP variant>#` e.g. `IWBP01#`, `IWBPHT#`

### `internal/tcp/protocol/parser_test.go`
Table-driven tests covering:
- Valid `AP00` login frame parsing
- Valid health frame (`APHT`) parsing
- Frame missing `IW` prefix → error
- Frame with 3-char cmd → error
- `ParseAP00` with non-digit IMEI → error
- `ParseAP00` with 14-digit IMEI → error
- `BuildReply("AP00")` contains `IWBP00;` prefix and ends with `#`

## Todo List
- [ ] Create `internal/tcp/protocol/parser.go` with all constants and functions
- [ ] Create `internal/tcp/protocol/parser_test.go`
- [ ] Run: `go test ./internal/tcp/protocol/... -v`

## Success Criteria
- All tests pass: `ok github.com/TranTheTuan/health-data-platform/internal/tcp/protocol`
- No external dependencies — pure Go stdlib

## Security Considerations
- Reject frames without `IW` prefix immediately.
- Validate IMEI is exactly 15 digits before any DB lookup.
- Maximum frame size guard: if accumulated bytes > 64KB without `#`, close connection (anti-DoS).

## Next Steps
- [Phase 3: TCP Server Core](./phase-03-tcp-server.md)
