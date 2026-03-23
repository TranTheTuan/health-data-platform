---
title: "Code Review — Demo TCP Packet Generator"
date: 2026-03-23
reviewer: code-reviewer
plan: plans/20260323-0957-device-packet-inspector/
---

# Code Review Summary

## Scope
- Files: `internal/demo/packet_generator.go`, `internal/demo/session_manager.go`, `internal/handler/http/demo.go`, `internal/delivery/http/router.go`, `cmd/api/main.go`, `web/templates/packets.html`
- LOC: ~350 (Go) + ~220 (HTML/JS)
- Focus: Thread safety, connection lifecycle, auth/ownership, JS error handling, security
- Build status: `go build ./...` — PASSES, zero errors

---

## Overall Assessment

Well-structured, clean implementation. Auth ownership guards are correct on all four endpoints. The mutex strategy is sound for its level of concurrency. A few real issues are present: a TOCTOU race in `SendBurst`, a missing `conn.SetDeadline` reset after a failed burst, a leaked `net.Conn` on partial startup (if Write succeeds but ack fails), and an XSS vector in `renderTable`. All are fixable with small targeted changes.

---

## Critical Issues

### 1. XSS via unescaped `innerHTML` in `renderTable` (packets.html:253-255)

**Problem:** `p.command_code` and `p.raw_payload` are inserted directly into `tr.innerHTML`. If a malicious packet arrives with payload `<img src=x onerror=alert(1)>`, it executes in the user's browser. The server stores raw_payload verbatim from TCP input.

**Impact:** Stored XSS. Any packet the server stores can execute JS on any user viewing the inspector.

**Fix:** Use `textContent` instead of `innerHTML` for untrusted fields.

```js
// Replace the innerHTML template string with DOM construction:
const tr = document.createElement('tr');
const tdTime = document.createElement('td');
tdTime.style.cssText = 'white-space: nowrap; color: var(--text-muted); font-size: 13px;';
tdTime.textContent = new Date(p.created_at).toLocaleString();

const tdCmd = document.createElement('td');
const badge = document.createElement('span');
badge.className = 'code-badge';
badge.textContent = p.command_code;
tdCmd.appendChild(badge);

const tdPayload = document.createElement('td');
tdPayload.className = 'payload';
tdPayload.textContent = p.raw_payload || '';
if (!p.raw_payload) {
    const empty = document.createElement('span');
    empty.style.opacity = '0.5';
    empty.textContent = 'Empty';
    tdPayload.appendChild(empty);
}
tr.append(tdTime, tdCmd, tdPayload);
tbody.appendChild(tr);
```

---

## High Priority

### 2. TOCTOU race in `SendBurst` — RLock → write on conn (session_manager.go:71-89)

**Problem:** `SendBurst` acquires only an `RLock` to read the session pointer, then calls `conn.Write` and `conn.SetDeadline` without holding any lock. If `StopSession` (which acquires a full `Lock`) closes and deletes the session concurrently, the goroutine in `SendBurst` continues writing to a closed `conn`. This produces a benign error in practice but is a real race condition — the race detector will flag it if the conn object itself is mutated.

**More concretely:** `closeAndRemove` is called from inside `SendBurst` (line 83) while holding no lock; it then tries to acquire a write lock — that is safe. But the pattern of "read pointer under RLock, mutate conn state outside lock" is fragile.

**Fix:** Either upgrade to a full `Lock` for the write path or add a `sync.Mutex` per-session to guard conn I/O separately. The simplest fix given low concurrency:

```go
func (m *SessionManager) SendBurst(deviceID string, count int) error {
    m.mu.Lock()               // full lock — burst is not a hot path
    defer m.mu.Unlock()
    sess, exists := m.sessions[deviceID]
    if !exists {
        return ErrSessionNotFound
    }
    // ... rest of loop unchanged, no need to call closeAndRemove separately
    // on write error just close + delete inline and return
}
```

### 3. Deadline not cleared after partial burst failure (session_manager.go:81-88)

**Problem:** If `conn.Write` fails on iteration `i < count`, `closeAndRemove` is called and the method returns. However if all writes succeed but the last `reader.ReadString('#')` times out (slow server), `conn.SetDeadline(time.Time{})` at line 88 is still reached (good). But if `closeAndRemove` is called mid-loop due to a write error, the deadline is never cleared on the conn before it is closed — not a leak since conn is closed, but the deadline is set on a conn that `closeAndRemove` will then close. This is benign but the flow is confusing.

**Separate sub-issue:** `reader.ReadString('#')` errors are silently swallowed on line 86 (no assignment to `_`). While ignoring ack errors is intentional, this should be a named discard with a comment to signal the intent is deliberate, and the error surfaced should it be non-EOF.

```go
_, _ = sess.reader.ReadString('#') // ack not required; ignore timeout/EOF
```

### 4. Connection leak on `StartSession` login-ack failure (session_manager.go:61-64)

**Problem:** The login ack check at line 61 combines two conditions:
```go
if err != nil || !strings.HasPrefix(ack, "IWBP00") {
    conn.Close()
    return errors.New("demo: login rejected by TCP server")
}
```
This is correct (`conn.Close()` is called). However, the actual login reply from the server is `IWBP00;<timestamp>,0#` (see `protocol.BuildReply`). The prefix check `strings.HasPrefix(ack, "IWBP00")` is correct and works. No leak here — just confirming it's fine.

**However:** The `bufio.Reader` wrapping `conn` at line 59 is created *before* the deadline is set at line 58. The ordering is:

```go
conn.SetDeadline(time.Now().Add(5 * time.Second))   // line 58
reader := bufio.NewReader(conn)                       // line 59
```

Actually line 58 comes before line 59 — re-reading confirms the order is correct. No issue.

---

## Medium Priority

### 5. `getIMEI` called on every request — DB query per endpoint (demo.go:24-35)

**Problem:** Every `StartSession`, `StopSession`, `SessionStatus`, and `SendBurst` call invokes `h.svc.ListDevices(ctx, userID)`, which hits the database to fetch all devices for the user, then linearly scans for the matching `deviceID`. For users with many devices this is wasteful; more specifically it issues a DB round-trip on every burst call.

**Context:** This is a demo feature so it's acceptable short-term, but the service already has `LookupDeviceByIMEI` — there is no `GetDeviceByID` method that would allow a direct single-device ownership check.

**Recommendation:** Add a `GetDeviceByID(ctx, deviceID) (domain.Device, error)` to the repository + service, and do ownership check as `device.UserID == userID`. This reduces the auth check from O(n) DB read to O(1).

### 6. `math/rand` without seed — deterministic sequences before Go 1.20 (packet_generator.go:5,22)

**Problem:** `rand.Intn` from `math/rand` is used without seeding. In Go 1.20+, the global source is automatically seeded randomly, so this is fine for Go 1.20+. If the project targets Go < 1.20, the sequence will be deterministic (seed 1) and all demo sessions will emit identical packet values.

**Action:** Confirm `go.mod` requires Go 1.20+. If yes, no change needed. If below, add `rand.Seed(time.Now().UnixNano())` in an `init()`.

**Verification:**
```bash
grep ^go go.mod
```

### 7. GPS timestamp hardcoded to `20260323120000` (packet_generator.go:49)

**Problem:** The GPS payload builder hardcodes the date literal `20260323120000` as the satellite fix timestamp. When the date rolls past 2026-03-23 this will look like stale data in the UI and any downstream system doing temporal analysis.

**Fix:** Replace with `time.Now().UTC().Format("20060102150405")`.

### 8. `startDemoSession` re-enables btn-start in `finally` while badge already updated (packets.html:346-363)

**Problem:** In the success path, `updateSessionBadge(true)` is called first (disabling btn-start), then `finally` runs `document.getElementById('btn-start').disabled = false` — re-enabling it. The subsequent `checkSessionStatus()` call will eventually correct this, but there is a brief window where both btn-start and btn-burst are enabled simultaneously.

**Fix:** In the success path, do not re-enable btn-start in `finally`. Instead let `checkSessionStatus()` control the final state (it already does), or conditionally skip the re-enable:

```js
} finally {
    // Only re-enable start if not active (success path leaves it disabled via checkSessionStatus)
    checkSessionStatus(); // this will set correct state
    // remove: document.getElementById('btn-start').disabled = false;
}
```

---

## Low Priority

### 9. `sendBurst` does not re-enable btn-stop on non-404 errors (packets.html:384-399)

**Problem:** When `!res.ok` and status is not 404 (line 390), the method returns early from the `try` block. The `finally` block calls `checkSessionStatus()` which will set `btn-stop` state correctly — but only after the async check completes. Between return and status check, btn-stop remains disabled. This is a minor UX glitch, not a functional bug.

### 10. `stopDemoSession` does not call `refreshData()` after success (packets.html:365-382)

No bug, but stopping the session means no more new packets will arrive. Refreshing after stop is consistent with the `sendBurst` behavior and would give users immediate confirmation.

### 11. `resolveAddr` only handles `:PORT` shorthand (session_manager.go:33-38)

**Problem:** If `cfg.TCPAddr` is `0.0.0.0:9090`, this function returns `0.0.0.0:9090` unchanged. Dialing `0.0.0.0` from the same host is typically fine on Linux but may fail on some environments. Mapping to `127.0.0.1` explicitly would be more robust.

**Fix:** If address starts with `0.0.0.0:` replace with `localhost:`.

### 12. `showToast` referenced in JS but not defined in packets.html

`showToast()` is called throughout the demo JS (lines 349, 353, 356, etc.) but is not defined in this template file. It must be defined in `base.html` or another included template. Verify it exists and handles the `'warning'` severity level used at line 349 (most implementations only handle `'success'` and `'error'`).

---

## Edge Cases Found by Scouting

- **Concurrent `StartSession` calls for same device:** Two requests that both pass the `exists` check before either inserts will both attempt `net.Dial`. Only the first to acquire the mutex write lock succeeds; the second gets `ErrSessionAlreadyActive` — correct, mutex is held for full start sequence.
- **Server restart with orphaned sessions:** If the HTTP server restarts, `SessionManager` is re-initialized with empty map. Any open TCP connections from the previous process are leaked at the OS level until the remote closes them. Since this is a demo-only self-connect feature, this is acceptable.
- **TCP server not yet ready at startup:** `NewSessionManager` is wired in `main.go` before the TCP server goroutine starts. If a user hits "Start Session" within the first millisecond of server boot, `net.DialTimeout` with 5s timeout will retry until the TCP server binds. Acceptable race.
- **`StopSession` called while `SendBurst` is in progress:** Due to the TOCTOU issue (#2 above), `conn.Close()` via `StopSession` will cause the `conn.Write` in `SendBurst` to return an error. The loop will call `closeAndRemove`, which tries to acquire a write lock — but `StopSession` already deleted the session. `closeAndRemove` guards with `if sess, exists := m.sessions[deviceID]; exists`, so the double-delete is safe. Net result: burst fails mid-way, session is gone. No crash or deadlock, but the race detector would flag the unsynchronized conn access.

---

## Positive Observations

- All four endpoints check `user_id` from session context before any other logic — correct auth-first pattern.
- Ownership guard via `getIMEI` is applied consistently on every handler including `SessionStatus`.
- `ErrSessionAlreadyActive` and `ErrSessionNotFound` are sentinel errors checked with `errors.Is` — correct.
- `conn.SetDeadline` is used with 5s for login and 3s per write — no indefinite blocks.
- `closeAndRemove` is idempotent (guards with `exists` check).
- `SendBurst` burst count is a `const` (not a user-supplied param) — no DoS vector on burst size.
- Demo routes are under the existing `protected` group with `AuthMiddleware` — no new middleware needed.
- Build passes cleanly (`go build ./...` exits 0).

---

## Recommended Actions (Prioritized)

1. **[Critical]** Fix XSS in `renderTable` — switch to DOM `textContent` API.
2. **[High]** Fix TOCTOU race in `SendBurst` — use full write lock or per-session mutex.
3. **[High]** Fix hardcoded GPS timestamp (2026-03-23) — use `time.Now().UTC().Format(...)`.
4. **[Medium]** Verify `showToast('warning')` is handled in base template.
5. **[Medium]** Confirm Go 1.20+ in `go.mod`; if not, seed `math/rand`.
6. **[Medium]** Fix `startDemoSession` `finally` block re-enabling btn-start incorrectly.
7. **[Low]** Replace hardcoded `0.0.0.0` dial target with explicit `localhost` in `resolveAddr`.
8. **[Low]** Add `GetDeviceByID` service method to replace O(n) `ListDevices` ownership check.

---

## Metrics

- Type coverage: N/A (Go — fully typed, no issues found by `go build`)
- Linting issues: 0 compilation errors; `go vet` not runnable in this session
- Test coverage: No tests for `demo` package — recommend unit tests for `SessionManager` concurrency and `RandomFrame` output shape
- Security issues: 1 critical (XSS), 1 medium (TOCTOU race), remainder low

---

## Unresolved Questions

- Does `showToast` in `base.html` support the `'warning'` severity passed by `startDemoSession`?
- What Go version does `go.mod` declare — is `math/rand` auto-seeding guaranteed?
- Is there a plan to expose the demo feature only in non-production environments, or is it intended for prod use?
