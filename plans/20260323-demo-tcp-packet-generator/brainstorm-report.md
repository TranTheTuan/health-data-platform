# Brainstorm: Demo TCP Packet Generator

**Date:** 2026-03-23
**Feature:** Auto-send random TCP packets to targeted device for demo

---

## Problem Statement

Manual TCP packet sending via Packet Sender is tedious for demos. Need a one-click way to simulate a smartwatch session — auth, health data bursts, controlled teardown — all from the existing web UI.

---

## Requirements

- "Start Session" button: connect to TCP server (`:9090`), send `AP00` auth using device's IMEI
- "Send Burst" button: send 5–10 random health packets on active connection
- "Stop Session" button: gracefully close TCP connection
- Session status indicator in UI (active / inactive)
- Backend self-connects (loopback to localhost:9090)
- Device IMEI must be registered in DB (already true on packet inspector page)

---

## Approaches Evaluated

### Option A — HTTP endpoints + in-memory session map ✅ SELECTED
- 4 REST endpoints to manage session lifecycle
- Backend holds `map[deviceID]*DemoSession{conn, imei}` in memory
- Pros: simple, zero deps, follows existing patterns, session persists between bursts
- Cons: not horizontally scalable (irrelevant for demo)

### Option B — Fire-and-forget per click
- Connect → auth → burst → disconnect on each click
- Pros: no state management
- Cons: doesn't support "terminate session" UX; re-auth overhead per click

### Option C — WebSocket with live packet feed
- Real-time stream of sent packets pushed to UI
- Pros: great UX
- Cons: significant complexity (WebSocket goroutines, protocol framing) — overkill

---

## Final Solution: Option A

### API Design

| Method | Endpoint | Action |
|--------|----------|--------|
| POST | `/protected/devices/:id/demo/session` | Connect + AP00 auth |
| POST | `/protected/devices/:id/demo/packets` | Send burst (5–10 random) |
| DELETE | `/protected/devices/:id/demo/session` | Close TCP conn |
| GET | `/protected/devices/:id/demo/session` | Check session status |

### New Files

```
internal/demo/
├── packet_generator.go   # Random IW-format payload per packet type
└── session_manager.go    # In-memory session map, Start/SendBurst/Stop
internal/handler/http/demo.go  # 4 HTTP handlers
```

### Modified Files

```
internal/delivery/http/router.go   # Register new routes
web/templates/packets.html         # Add demo control buttons + status badge
```

No DB changes needed. No new dependencies.

### Packet Types in Burst

Random mix from: `AP49` (HR), `APHT` (HR+BP), `AP50` (temp), `AP01` (GPS), `AP97` (sleep)

Excluded: `AP00` (login — only on session start), `AP03`/`APWT` (not persisted — useless for demo)

### Random Payload Ranges

| Type | Fields | Range |
|------|--------|-------|
| AP49 | heart_rate | 55–105 bpm |
| APHT | hr, systolic, diastolic | 60–100, 100–140, 60–90 |
| AP50 | temp (scaled ×100) | 3600–3750 (36.00–37.50°C) |
| AP01 | lat, lng (fixed area), speed | Hanoi vicinity, 0–5 km/h |
| AP97 | sleep stages (simplified) | Random durations |

### Session State Machine

```
[idle] → START → [connecting] → AP00 sent → [active]
[active] → SEND BURST → sends 5-10 packets → [active]
[active] → STOP → conn.Close() → [idle]
[active] → TCP error/timeout → [idle] (auto-cleanup)
```

### UI Changes (packets.html)

- Demo control panel below the packet filter bar
- "Start Session" → grayed out when active
- "Send Burst" + "Stop Session" → visible only when active
- Status badge: green "Session Active" / gray "No Session"
- JS polls `GET .../demo/session` on page load to restore badge state

---

## Implementation Considerations

- `session_manager.go` uses `sync.RWMutex` for concurrent-safe map access
- TCP client must wait for AP00 ack before marking session active
- `session_manager` reads `TCP_HOST`/`TCP_PORT` from config (already in `configs/config.go`)
- `demo.go` handler injects `DeviceService` (to fetch IMEI by device ID) + `DemoSessionManager`
- Respect file size limit: keep each file under 200 lines

---

## Risks

| Risk | Mitigation |
|------|-----------|
| Device IMEI not in DB | Handler validates device ownership first (already done by existing middleware pattern) |
| TCP server rejects AP00 (device not found) | Return clear error to UI from start endpoint |
| Stale sessions (user closes browser) | 5-min idle timeout on TCP server auto-kills; also expose stop on page unload |
| Multiple sessions per device | session_manager rejects start if session already active (return 409) |

---

## Success Criteria

- Click "Start" → session active badge appears, packet count in inspector increases
- Click "Send Burst" → 5–10 new rows in packet inspector after refresh
- Click "Stop" → TCP conn closed, badge goes gray
- No code changes needed to existing TCP server or device service

---

## Next Steps

1. Implement `packet_generator.go`
2. Implement `session_manager.go`
3. Implement `demo.go` HTTP handlers
4. Register routes in `router.go`
5. Update `packets.html` with demo controls
6. Wire dependencies in `cmd/api/main.go`
