# Phase 01 — Demo Package

## Context Links
- Protocol constants: `internal/tcp/protocol/parser.go` (CmdLogin, CmdGPSLoc, etc.)
- Frame format: `IW<CMD><PAYLOAD>#` (device → server)
- Brainstorm: `plans/20260323-demo-tcp-packet-generator/brainstorm-report.md`

## Overview
- **Priority:** P2
- **Status:** Pending
- **Description:** Create `internal/demo/` package with two files:
  1. `packet_generator.go` — random IW-format payload generator per packet type
  2. `session_manager.go` — in-memory TCP session lifecycle manager

## Key Insights
- The IW frame format sent by the *client* (device) is: `"IW" + cmd + payload + "#"`
- `BuildReply` in protocol package is for server → client; we construct client frames manually
- `TCPAddr` in config may be `":9090"` (listen-only addr); must normalize to `"localhost:9090"` for dialing
- `AP00` ack response from server: `"IWBP00;yyyyMMddHHmmss,0#"` — must read and validate before marking session active
- Session map needs `sync.RWMutex` for safe concurrent HTTP handler access
- Payload format for each type is defined by how `shouldPersist` and TCP handler processes them — raw payload is stored as-is, so any reasonable format works

## Requirements

### Functional
- `packet_generator.go` generates realistic random payloads for: `AP49`, `APHT`, `AP50`, `AP01`, `AP97`
- `session_manager.go` manages: `StartSession(deviceID, imei)`, `SendBurst(deviceID, count)`, `StopSession(deviceID)`, `IsActive(deviceID)`
- `StartSession` dials TCP, sends AP00 login, reads ack, stores session — returns error if device not found on TCP server
- `SendBurst` sends `count` random packets on the open connection, reads acks
- `StopSession` closes connection and removes from map
- Duplicate `StartSession` on same deviceID returns error (409 semantics)

### Non-functional
- Each file ≤ 200 LOC
- Thread-safe via `sync.RWMutex`
- Connections cleaned up on error (defer conn.Close only on failure path; on success caller owns conn)

## Architecture

```
internal/demo/
├── packet_generator.go   (~80 LOC)
└── session_manager.go    (~110 LOC)
```

### packet_generator.go

```go
package demo

import (
    "fmt"
    "math/rand"
    "github.com/TranTheTuan/health-data-platform/internal/tcp/protocol"
)

// PersistableCommands lists packet types valid for demo bursts.
var PersistableCommands = []string{
    protocol.CmdHeartRate,   // AP49
    protocol.CmdHRAndBP,     // APHT
    protocol.CmdHRBPSPO2,    // APHP
    protocol.CmdTemperature, // AP50
    protocol.CmdGPSLoc,      // AP01
    protocol.CmdSleep,       // AP97
}

// RandomFrame returns a complete IW frame string for a random packet type.
func RandomFrame() string {
    cmd := PersistableCommands[rand.Intn(len(PersistableCommands))]
    return BuildFrame(cmd, randomPayload(cmd))
}

// BuildFrame constructs a client-side IW frame: "IW<CMD><PAYLOAD>#"
func BuildFrame(cmd, payload string) string {
    return "IW" + cmd + payload + "#"
}

// LoginFrame constructs the AP00 login frame for the given IMEI.
func LoginFrame(imei string) string {
    return BuildFrame(protocol.CmdLogin, imei)
}

func randomPayload(cmd string) string {
    switch cmd {
    case protocol.CmdHeartRate:   // AP49: single HR value
        return fmt.Sprintf("%d", randRange(55, 105))
    case protocol.CmdHRAndBP:     // APHT: hr,systolic,diastolic
        return fmt.Sprintf("%d,%d,%d", randRange(60,100), randRange(100,140), randRange(60,90))
    case protocol.CmdHRBPSPO2:    // APHP: hr,systolic,diastolic,spo2,glucose
        return fmt.Sprintf("%d,%d,%d,%d,%.1f", randRange(60,100), randRange(100,140), randRange(60,90), randRange(95,100), float32(randRange(45,65))/10.0)
    case protocol.CmdTemperature: // AP50: temp scaled ×100 (3600=36.00°C)
        return fmt.Sprintf("%d", randRange(3600, 3750))
    case protocol.CmdGPSLoc:      // AP01: simplified GPS (Hanoi area)
        lat := 21.0 + float64(rand.Intn(100))/1000.0
        lng := 105.8 + float64(rand.Intn(100))/1000.0
        return fmt.Sprintf("%.4f,%.4f,10,%d,6,20260323120000", lat, lng, rand.Intn(5))
    case protocol.CmdSleep:       // AP97: simplified sleep stages
        return fmt.Sprintf("%d,%d,%d", randRange(60,180), randRange(30,120), randRange(10,60))
    default:
        return ""
    }
}

func randRange(min, max int) int {
    return min + rand.Intn(max-min+1)
}
```

### session_manager.go

```go
package demo

import (
    "bufio"
    "errors"
    "net"
    "strings"
    "sync"
    "time"
)

var ErrSessionAlreadyActive = errors.New("demo session already active for this device")
var ErrSessionNotFound = errors.New("no active demo session for this device")

type demoSession struct {
    conn   net.Conn
    reader *bufio.Reader
}

type SessionManager struct {
    mu       sync.RWMutex
    sessions map[string]*demoSession // key: deviceID
    tcpAddr  string
}

func NewSessionManager(tcpAddr string) *SessionManager {
    return &SessionManager{
        sessions: make(map[string]*demoSession),
        tcpAddr:  resolveAddr(tcpAddr),
    }
}

// resolveAddr converts ":9090" → "localhost:9090" for dialing
func resolveAddr(addr string) string {
    if strings.HasPrefix(addr, ":") {
        return "localhost" + addr
    }
    return addr
}

func (m *SessionManager) StartSession(deviceID, imei string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    if _, exists := m.sessions[deviceID]; exists {
        return ErrSessionAlreadyActive
    }

    conn, err := net.DialTimeout("tcp", m.tcpAddr, 5*time.Second)
    if err != nil {
        return err
    }

    // Send AP00 login
    if _, err = conn.Write([]byte(LoginFrame(imei))); err != nil {
        conn.Close()
        return err
    }

    // Read ack (expect IWBP00;...#)
    conn.SetDeadline(time.Now().Add(5 * time.Second))
    reader := bufio.NewReader(conn)
    ack, err := reader.ReadString('#')
    if err != nil || !strings.HasPrefix(ack, "IWBP00") {
        conn.Close()
        return errors.New("demo: login rejected by TCP server")
    }
    conn.SetDeadline(time.Time{}) // clear deadline

    m.sessions[deviceID] = &demoSession{conn: conn, reader: reader}
    return nil
}

func (m *SessionManager) SendBurst(deviceID string, count int) error {
    m.mu.RLock()
    sess, exists := m.sessions[deviceID]
    m.mu.RUnlock()
    if !exists {
        return ErrSessionNotFound
    }

    for i := 0; i < count; i++ {
        frame := RandomFrame()
        sess.conn.SetDeadline(time.Now().Add(3 * time.Second))
        if _, err := sess.conn.Write([]byte(frame)); err != nil {
            m.closeAndRemove(deviceID)
            return err
        }
        // Read ack (discard, just drain)
        sess.reader.ReadString('#')
    }
    sess.conn.SetDeadline(time.Time{})
    return nil
}

func (m *SessionManager) StopSession(deviceID string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    sess, exists := m.sessions[deviceID]
    if !exists {
        return ErrSessionNotFound
    }
    sess.conn.Close()
    delete(m.sessions, deviceID)
    return nil
}

func (m *SessionManager) IsActive(deviceID string) bool {
    m.mu.RLock()
    defer m.mu.RUnlock()
    _, exists := m.sessions[deviceID]
    return exists
}

func (m *SessionManager) closeAndRemove(deviceID string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    if sess, exists := m.sessions[deviceID]; exists {
        sess.conn.Close()
        delete(m.sessions, deviceID)
    }
}
```

## Related Code Files

| File | Action | Description |
|------|--------|-------------|
| `internal/demo/packet_generator.go` | **CREATE** | Random IW frame generator |
| `internal/demo/session_manager.go` | **CREATE** | In-memory TCP session manager |

## Implementation Steps

1. Create `internal/demo/` directory
2. Create `packet_generator.go` with `RandomFrame()`, `LoginFrame()`, `BuildFrame()`, `randomPayload()`
3. Create `session_manager.go` with `SessionManager`, `StartSession()`, `SendBurst()`, `StopSession()`, `IsActive()`
4. Run `go build ./internal/demo/...` to verify compile

## Todo List

- [ ] Create `internal/demo/packet_generator.go`
- [ ] Create `internal/demo/session_manager.go`
- [ ] Verify compile: `go build ./internal/demo/...`

## Success Criteria

- `go build ./internal/demo/...` passes with no errors
- `packet_generator.go` ≤ 100 LOC
- `session_manager.go` ≤ 130 LOC
- `SessionManager` is thread-safe

## Risk Assessment

| Risk | Mitigation |
|------|-----------|
| TCP server rejects AP00 (IMEI not in DB) | `StartSession` returns error; handler returns 400 |
| Connection leak on burst error | `closeAndRemove` called on any write error |
| Stale sessions (browser closed) | TCP server's 5-min idle timeout auto-closes; also `StopSession` in page unload JS |

## Security Considerations

- Only authenticated users can trigger demo sessions (auth middleware in phase 3)
- DeviceID ownership validated before starting session (phase 2)
- No user data exposed by the generator — only synthetic payloads

## Next Steps

- Phase 02: HTTP handler that wraps `SessionManager` with ownership validation
