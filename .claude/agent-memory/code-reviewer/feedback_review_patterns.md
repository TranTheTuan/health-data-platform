---
name: recurring review patterns in this project
description: Issues and patterns that recur or are worth watching for in future reviews
type: feedback
---

Watch for `innerHTML` with server-derived data — this project stores raw TCP payloads verbatim; XSS via `renderTable`-style innerHTML is a real risk whenever packet data is rendered in the browser.

**Why:** TCP payload content is arbitrary bytes from the device; a compromised or malicious device can craft XSS payloads.

**How to apply:** Flag any `el.innerHTML = ...` where the content includes `command_code`, `raw_payload`, or any server-returned string. Recommend `textContent` or DOM construction.

RLock → conn mutation pattern: `sync.RWMutex` RLock is used to read a session pointer, then the conn is mutated outside the lock. This is a TOCTOU/race pattern. Always check whether operations after the unlock mutate shared state.

**Why:** `SendBurst` reads session under RLock but calls `conn.Write` and `conn.SetDeadline` outside any lock. Race detector will flag concurrent `StopSession`.

**How to apply:** When reviewing SessionManager-style structs, check that all I/O operations on the stored net.Conn are either fully under a lock or protected by a per-session mutex.
