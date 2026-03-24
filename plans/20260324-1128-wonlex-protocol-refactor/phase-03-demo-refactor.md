# Phase 03 — Demo Generator + Service Refactor

## Context Links
- Demo generator: `internal/demo/packet_generator.go` (62 LOC, old IW format)
- Demo service: `internal/service/demo_service.go` (137 LOC)
- New parser API: Phase 01 output

## Overview
- **Priority:** P2
- **Status:** Complete
- **Description:** Rewrite the demo packet generator and demo service to produce and consume Wonlex-format frames instead of old IW frames.

## Key Insights
- `packet_generator.go` currently builds `IW<CMD><PAYLOAD>#` — must become `[3G*<deviceID>*<LEN>*<CMD>,<PAYLOAD>]`
- `demo_service.go` login flow currently sends `IW` frame and expects `IWBP00` ack — must send Wonlex `LK` frame and expect `[CS*...*...*LK]` ack
- `demo_service.go` burst reads ack up to `#` — must read up to `]` now
- `LoginFrame` becomes `LinkFrame` — sends a `LK` keep-alive to authenticate
- `RandomFrame` needs to accept `deviceID` param to include in frame header

## Requirements

### Functional
- `BuildWonlexFrame(deviceID, cmd, payload)` constructs `[3G*deviceID*LEN*CMD,PAYLOAD]`
- `LinkFrame(deviceID)` constructs `[3G*deviceID*0002*LK]`
- `RandomFrame(deviceID)` returns random data frame with device ID embedded
- Demo service login sends `LinkFrame`, expects `[CS*...*LK]` ack
- Demo service burst reads ack up to `]` instead of `#`

### Non-functional
- `packet_generator.go` ≤ 100 LOC
- `demo_service.go` ≤ 140 LOC

## Architecture

### packet_generator.go rewrite

```go
package demo

import (
    "fmt"
    "math/rand"
    "time"

    "github.com/TranTheTuan/health-data-platform/internal/tcp/protocol"
)

var PersistableCommands = []string{
    protocol.CmdLocation, // UD
    protocol.CmdAlarm,    // AL
}

func BuildWonlexFrame(deviceID, cmd, payload string) string {
    content := cmd
    if payload != "" {
        content = cmd + "," + payload
    }
    lenHex := fmt.Sprintf("%04X", len(content))
    return fmt.Sprintf("[3G*%s*%s*%s]", deviceID, lenHex, content)
}

func LinkFrame(deviceID string) string {
    return BuildWonlexFrame(deviceID, protocol.CmdLink, "")
}

func RandomFrame(deviceID string) string {
    cmd := PersistableCommands[rand.Intn(len(PersistableCommands))]
    return BuildWonlexFrame(deviceID, cmd, randomPayload(cmd))
}

func randomPayload(cmd string) string {
    switch cmd {
    case protocol.CmdLocation:
        // Simplified UD location payload (Hanoi area)
        ts := time.Now().UTC().Format("060102,150405")
        lat := 22.0 + float64(rand.Intn(1000))/10000.0
        lng := 113.0 + float64(rand.Intn(10000))/10000.0
        return fmt.Sprintf("%s,A,%.6f,N,%.6f,E,0.00,0.0,0.0,6,100,51,14188,0,00010010,6,255,460,0,9360,5081,156,...", ts, lat, lng)
    case protocol.CmdAlarm:
        // Simplified AL payload (location + alarm flags)
        ts := time.Now().UTC().Format("060102,150405")
        return fmt.Sprintf("%s,A,22.570720,N,113.8620167,E,0.00,188.6,0.0,9,100,51,14188,0,00200000,...", ts)
    default:
        return ""
    }
}
```

### demo_service.go changes

```go
// StartSession: change login to link
loginFrame := demo.LinkFrame(imei) // imei here is the 10-digit device ID
conn.Write([]byte(loginFrame))

// Read ack — change delimiter from '#' to ']' 
ack, err := reader.ReadString(']')
if err != nil || !strings.Contains(ack, "LK") {
    conn.Close()
    return errors.New("demo: link rejected by TCP server")
}

// SendBurst: change ack delimiter from '#' to ']'
sess.reader.ReadString(']')
```

## Related Code Files

| File | Action | Description |
|------|--------|-------------|
| `internal/demo/packet_generator.go` | **REWRITE** | Wonlex frame builder |
| `internal/service/demo_service.go` | **MODIFY** | Update login/ack protocol |

## Implementation Steps

1. Rewrite `packet_generator.go`:
   - `BuildWonlexFrame(deviceID, cmd, payload)` calculates hex length
   - `LinkFrame(deviceID)` = `[3G*deviceID*0002*LK]`
   - `RandomFrame(deviceID)` picks random command and generates payload
   - Update `randomPayload` for Wonlex command types (UD, AL formats)
2. Modify `demo_service.go`:
   - `LoginFrame(imei)` → `LinkFrame(imei)` (imei is actually 10-digit device ID now)
   - Ack delimiter: `ReadString('#')` → `ReadString(']')`
   - Ack validation: `HasPrefix(ack, "IWBP00")` → `Contains(ack, "LK")`
   - `RandomFrame()` → `RandomFrame(deviceID)` — need to store deviceID in session
3. Update `demoSession` struct:
   - Add `deviceID string` field so `SendBurst` can pass it to `RandomFrame`
4. Run `go build ./...` to verify compile

## Todo List

- [x] Rewrite `internal/demo/packet_generator.go` with Wonlex frames
- [x] Modify `internal/service/demo_service.go` for new protocol
- [x] Add `deviceID` field to `demoSession` struct
- [x] Verify compile: `go build ./...`

## Success Criteria

- `LinkFrame("8800000015")` → `[3G*8800000015*0002*LK]`
- `RandomFrame("8800000015")` produces valid Wonlex frame
- Demo start/stop/burst works end-to-end with new protocol
- `go build ./...` passes

## Risk Assessment

| Risk | Mitigation |
|------|-----------|
| UD payload format doesn't match real device exactly | Acceptable — demo generates synthetic data, raw payload is stored as-is |
| Double-closing session on burst error | Existing `closeAndRemove` handles this safely |
