# Phase 04 — UI Demo Controls (packets.html)

## Context Links
- Template: `web/templates/packets.html`
- CSS design system: `web/templates/base.html` (CSS variables: `--accent`, `--surface`, `--border`, etc.)
- Demo endpoints: `POST/DELETE/GET /protected/devices/:id/demo/session`, `POST /protected/devices/:id/demo/packets`

## Overview
- **Priority:** P2
- **Status:** Pending (blocked by Phase 03)
- **Description:** Add a "Demo Controls" card above the packet table in `packets.html` with Start/Stop/Send Burst buttons and a session status badge.

## Key Insights
- `deviceId` is already in JS as `const deviceId = document.querySelector('.layout').dataset.deviceId`
- The page uses `fetch()` for API calls — same pattern for demo endpoints
- Must handle: start/stop button state toggle, loading state during API calls, error feedback via existing `showToast`
- On page load: call `GET .../demo/session` to restore session badge (user may have navigated away and back)
- Session auto-cleanup: add `window.beforeunload` to call `DELETE .../demo/session` (best-effort, no guarantee)
- Use existing CSS variables — no new CSS frameworks

## Requirements

### Functional
- "Start Session" button: calls `POST .../demo/session`; on success shows "Session Active" badge + enables Send/Stop buttons
- "Send Burst" button: calls `POST .../demo/packets`; on success shows toast "Sent 7 packets" + auto-refreshes packet list
- "Stop Session" button: calls `DELETE .../demo/session`; on success resets to inactive state
- Status badge: green "● Session Active" when active, gray "○ No Session" when inactive
- All buttons disabled while request is in-flight (prevent double-click)
- Page load: `GET .../demo/session` → restore badge state

### Non-functional
- Demo card styled consistently with existing `.filter-bar` card
- No new CSS files — inline styles within `{{define "head"}}` block

## Architecture

### CSS additions (in `{{define "head"}}`)

```css
/* Demo Controls */
.demo-bar { background: var(--surface); padding: 16px 24px; border-radius: var(--radius-lg); border: 1px solid var(--border); margin-bottom: 24px; display: flex; gap: 12px; align-items: center; flex-wrap: wrap; }
.demo-bar h3 { margin: 0; font-size: 14px; font-weight: 600; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.5px; flex: 1; }
.session-badge { padding: 4px 10px; border-radius: 20px; font-size: 12px; font-weight: 600; }
.session-badge.active { background: rgba(16,185,129,0.15); color: #10b981; border: 1px solid rgba(16,185,129,0.3); }
.session-badge.inactive { background: var(--bg-mid); color: var(--text-muted); border: 1px solid var(--border); }
.btn-danger { background: rgba(239,68,68,0.15); color: #ef4444; border: 1px solid rgba(239,68,68,0.3); padding: 8px 16px; border-radius: var(--radius-sm); cursor: pointer; font-weight: 500; font-size: 14px; }
.btn-danger:hover:not(:disabled) { background: rgba(239,68,68,0.25); }
button:disabled { opacity: 0.5; cursor: not-allowed; }
```

### HTML additions (in `{{define "content"}}`, after filter-bar div)

```html
<!-- Demo Controls -->
<div class="demo-bar">
    <h3>Demo</h3>
    <span id="session-badge" class="session-badge inactive">○ No Session</span>
    <button id="btn-start" class="btn-primary" onclick="startDemoSession()">Start Session</button>
    <button id="btn-burst" class="btn-secondary" onclick="sendBurst()" disabled>Send Burst</button>
    <button id="btn-stop" class="btn-danger" onclick="stopDemoSession()" disabled>Stop Session</button>
</div>
```

### JS additions (in `{{define "scripts"}}`)

```javascript
// ── Demo Session ─────────────────────────────────────────────────
async function updateSessionBadge(active) {
    const badge = document.getElementById('session-badge');
    const btnStart = document.getElementById('btn-start');
    const btnBurst = document.getElementById('btn-burst');
    const btnStop = document.getElementById('btn-stop');

    if (active) {
        badge.textContent = '● Session Active';
        badge.className = 'session-badge active';
        btnStart.disabled = true;
        btnBurst.disabled = false;
        btnStop.disabled = false;
    } else {
        badge.textContent = '○ No Session';
        badge.className = 'session-badge inactive';
        btnStart.disabled = false;
        btnBurst.disabled = true;
        btnStop.disabled = true;
    }
}

async function startDemoSession() {
    document.getElementById('btn-start').disabled = true;
    try {
        const res = await fetch(`/protected/devices/${deviceId}/demo/session`, { method: 'POST' });
        if (res.status === 409) { showToast('Session already active', 'error'); return; }
        if (!res.ok) {
            const d = await res.json();
            showToast(d.error || 'Failed to start session', 'error');
            return;
        }
        updateSessionBadge(true);
        showToast('Demo session started', 'success');
    } catch {
        showToast('Network error', 'error');
    } finally {
        document.getElementById('btn-start').disabled = false;
    }
}

async function stopDemoSession() {
    document.getElementById('btn-stop').disabled = true;
    try {
        const res = await fetch(`/protected/devices/${deviceId}/demo/session`, { method: 'DELETE' });
        if (!res.ok) {
            const d = await res.json();
            showToast(d.error || 'Failed to stop session', 'error');
            return;
        }
        updateSessionBadge(false);
        showToast('Demo session stopped', 'success');
    } catch {
        showToast('Network error', 'error');
    } finally {
        document.getElementById('btn-stop').disabled = false;
    }
}

async function sendBurst() {
    document.getElementById('btn-burst').disabled = true;
    try {
        const res = await fetch(`/protected/devices/${deviceId}/demo/packets`, { method: 'POST' });
        if (res.status === 404) { updateSessionBadge(false); showToast('Session expired', 'error'); return; }
        if (!res.ok) { showToast('Failed to send packets', 'error'); return; }
        const d = await res.json();
        showToast(`Sent ${d.sent} packets`, 'success');
        // Auto-refresh packet list to show new data
        refreshData();
    } catch {
        showToast('Network error', 'error');
    } finally {
        document.getElementById('btn-burst').disabled = !document.getElementById('btn-stop').disabled === false;
        // Re-enable if session still active
        const active = document.getElementById('session-badge').classList.contains('active');
        document.getElementById('btn-burst').disabled = !active;
    }
}

// Restore session state on page load
async function checkSessionStatus() {
    try {
        const res = await fetch(`/protected/devices/${deviceId}/demo/session`);
        if (res.ok) {
            const d = await res.json();
            updateSessionBadge(d.active);
        }
    } catch { /* silent */ }
}

// Best-effort cleanup on page unload
window.addEventListener('beforeunload', () => {
    navigator.sendBeacon(`/protected/devices/${deviceId}/demo/session`, JSON.stringify({ _method: 'DELETE' }));
});
```

**Note:** `sendBeacon` doesn't support DELETE. Use a dedicated `POST /protected/devices/:id/demo/session/stop` endpoint alternatively, or accept that sessions will be cleaned up by TCP server's 5-min idle timeout. **Preferred:** skip `beforeunload` complexity; rely on server-side idle timeout. Remove `beforeunload` listener.

**Correction for page load init:** Add `checkSessionStatus()` call inside the existing `DOMContentLoaded` handler.

```javascript
document.addEventListener('DOMContentLoaded', () => {
    // existing code...
    fetchPackets();
    checkSessionStatus(); // add this
});
```

## Related Code Files

| File | Action | Description |
|------|--------|-------------|
| `web/templates/packets.html` | **MODIFY** | Add demo bar CSS, HTML section, JS functions |

## Implementation Steps

1. Open `web/templates/packets.html`
2. In `{{define "head"}}` style block, append demo CSS classes
3. In `{{define "content"}}`, insert `<div class="demo-bar">...</div>` between `.filter-bar` and `.card` divs
4. In `{{define "scripts"}}`, add all demo JS functions after existing functions
5. In `DOMContentLoaded` handler, add `checkSessionStatus()` call
6. **Remove** the `beforeunload` sendBeacon — rely on TCP server idle timeout instead
7. Check that `showToast` function exists in the template (or `base.html`) — if not, implement a simple toast

## Todo List

- [ ] Add demo CSS to `{{define "head"}}` style block
- [ ] Add `<div class="demo-bar">` HTML between filter-bar and card
- [ ] Add JS functions: `updateSessionBadge`, `startDemoSession`, `stopDemoSession`, `sendBurst`, `checkSessionStatus`
- [ ] Add `checkSessionStatus()` to `DOMContentLoaded` handler
- [ ] Verify `showToast` function exists — if not, implement simple toast notification
- [ ] Manual test: start session, send burst, verify packets appear in table, stop session

## Success Criteria

- Start button → "Session Active" badge appears + Send/Stop enabled
- Send Burst → packet table auto-refreshes with new rows
- Stop button → badge resets, Start re-enabled
- Page reload → badge state restored from `GET /demo/session`
- No JS errors in browser console

## Risk Assessment

| Risk | Mitigation |
|------|-----------|
| `showToast` not defined in packets.html | Check base.html; if absent, add simple fixed-position toast div |
| Session state desync (tab closed mid-session) | TCP server idle timeout cleans up after 5 min |
| `sendBurst` re-enable logic is complex | Keep it simple: after any burst call, re-check `session-badge` class to decide button state |

## Security Considerations

- All demo API calls go through auth middleware (protected routes)
- No sensitive data sent in demo payloads — only synthetic health metrics

## Next Steps

- Manual smoke test: register device, open packet inspector, use demo controls
- Verify packets appear in packet inspector after burst
