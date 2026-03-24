# Phase 01 — Protocol Parser Rewrite

## Context Links
- Current parser: `internal/tcp/protocol/parser.go` (147 LOC, IW format)
- Current tests: `internal/tcp/protocol/parser_test.go` (238 LOC)
- Protocol doc: Wonlex frame format `[manufacturer*deviceID*contentLength*content]`

## Overview
- **Priority:** P1
- **Status:** Complete
- **Description:** Completely rewrite `parser.go` and `parser_test.go` to handle the Wonlex frame format.

## Key Insights
- Frame delimiters change from `IW...#` to `[...]`
- `ScanFrame` needs to split on `]` instead of `#`, and strip leading `[`
- 4 fields separated by `*`: manufacturer, deviceID (10 chars), contentLength (4-char hex), content
- Content is: `CMD` alone or `CMD,payload...` (comma-separated)
- Server replies use `CS` manufacturer prefix instead of `3G`
- Some commands (`UD`, `UD2`) don't need server reply — `BuildReply` must return empty string for these

## Requirements

### Functional
- `ScanFrame`: split on `]`, strip leading `[`
- `ParseFrame`: split `raw` by `*` into 4 fields → `Frame{Manufacturer, DeviceID, Length, Cmd, Payload}`
- `BuildReply`: construct `[CS*deviceID*LEN*CMD]` for commands that need reply; return `""` for no-reply commands
- Validate `DeviceID` is exactly 10 decimal digits
- Compute content length as `fmt.Sprintf("%04X", len(content))`

### Non-functional
- `parser.go` ≤ 150 LOC
- `parser_test.go` rewritten to test new format

## Architecture

### New Frame struct
```go
type Frame struct {
    Manufacturer string // "3G"
    DeviceID     string // 10-digit device identifier
    Length       string // 4-char hex content length
    Cmd          string // "UD", "AL", "LK", "UD2", "CONFIG", etc.
    Payload      string // everything after CMD comma, or "" if no payload
}
```

### New command constants
```go
const (
    CmdLink     = "LK"     // Keep-alive/link
    CmdLocation = "UD"     // GPS+LBS+WiFi location report
    CmdBlind    = "UD2"    // Blind spot data fill-in
    CmdAlarm    = "AL"     // Alarm data report
    CmdConfig   = "CONFIG" // Device configuration report
)
```

### Reply rules
```go
var replyRules = map[string]struct {
    prefix string // manufacturer prefix in reply ("CS" or "3G")
    needsReply bool
}{
    CmdLink:     {"CS", true},
    CmdLocation: {"",   false}, // no reply needed
    CmdBlind:    {"",   false}, // no reply
    CmdAlarm:    {"3G", true},
    CmdConfig:   {"CS", true},  // reply with result
}
```

### ScanFrame change
```go
func ScanFrame(data []byte, atEOF bool) (advance int, token []byte, err error) {
    // Find '[' start
    start := bytes.IndexByte(data, '[')
    if start < 0 {
        if atEOF { return 0, nil, nil }
        return 0, nil, nil
    }
    // Find ']' end after start
    end := bytes.IndexByte(data[start:], ']')
    if end < 0 {
        if atEOF { return 0, nil, nil }
        return 0, nil, nil
    }
    // Return content between [ and ]
    return start + end + 1, data[start+1 : start+end], nil
}
```

### ParseFrame
```go
func ParseFrame(raw string) (Frame, error) {
    parts := strings.SplitN(raw, "*", 4)
    if len(parts) != 4 { return Frame{}, ErrInvalidFrame }
    
    manufacturer := parts[0]
    deviceID := parts[1]
    length := parts[2]
    content := parts[3]
    
    if len(deviceID) != 10 { return Frame{}, ErrInvalidDeviceID }
    
    // Split content into CMD and payload at first comma
    cmd, payload := content, ""
    if idx := strings.IndexByte(content, ','); idx >= 0 {
        cmd = content[:idx]
        payload = content[idx+1:]
    }
    
    return Frame{Manufacturer: manufacturer, DeviceID: deviceID, Length: length, Cmd: cmd, Payload: payload}, nil
}
```

### BuildReply
```go
func BuildReply(deviceID, cmd string) string {
    rule, ok := replyRules[cmd]
    if !ok || !rule.needsReply { return "" }
    
    var content string
    if cmd == CmdConfig {
        content = "CONFIG,1" // acknowledge with result=1 (OK)
    } else {
        content = cmd
    }
    
    lenHex := fmt.Sprintf("%04X", len(content))
    return fmt.Sprintf("[%s*%s*%s*%s]", rule.prefix, deviceID, lenHex, content)
}
```

## Related Code Files

| File | Action | Description |
|------|--------|-------------|
| `internal/tcp/protocol/parser.go` | **REWRITE** | New Wonlex frame parser, reply builder |
| `internal/tcp/protocol/parser_test.go` | **REWRITE** | Tests for new format |

## Implementation Steps

1. Rewrite `parser.go`:
   - Replace all old constants with new Wonlex command constants
   - Replace `Frame` struct fields
   - Rewrite `ScanFrame` to use `[` and `]` delimiters
   - Rewrite `ParseFrame` to split by `*` and extract CMD/payload from content
   - Remove `ParseAP00` (no longer needed — device ID is in frame header)
   - Rewrite `BuildReply` to accept `(deviceID, cmd)` and construct `[CS*...*...*...]`
2. Rewrite `parser_test.go` with new test cases for Wonlex format
3. Run `go test ./internal/tcp/protocol/...`

## Todo List

- [x] Rewrite `parser.go` with Wonlex protocol
- [x] Rewrite `parser_test.go` with new test cases
- [x] Verify: `go test ./internal/tcp/protocol/...`
- [x] Verify: `go build ./...` (will fail until phase 02 updates handler)

## Success Criteria

- `parser.go` ≤ 150 LOC
- All new tests pass
- Frame `[3G*8800000015*00BC*UD,120118,...]` parses correctly
- `BuildReply("8800000015", "AL")` → `[3G*8800000015*0002*AL]`
- `BuildReply("8800000015", "UD")` → `""` (no reply)
- `BuildReply("8800000015", "CONFIG")` → `[CS*8800000015*0008*CONFIG,1]`

## Risk Assessment

| Risk | Mitigation |
|------|-----------|
| Breaking `go build` after phase 01 | Expected — handler.go still references old API; fixed in phase 02 |
| Content length calc off-by-one | Test with known examples from protocol doc |
