# Phase 3: Polish & Logout

**Context Links**
- [Main Plan](./plan.md)
- [Phase 2: Backend Wiring](./phase-02-backend-wiring.md)

## Overview
- Priority: P2
- Effort: 1h
- Status: Pending
- Add the logout endpoint, polish small UX details (loading states, empty state, toast notifications), and verify overall flow.

## Related Code Files
- `[MODIFY]` `internal/api/handlers/auth.go` — add `Logout` handler
- `[MODIFY]` `internal/api/routes.go` — add `GET /logout` route
- `[MODIFY]` `web/templates/dashboard.html` — loading spinner, empty state, toast

## Logout Handler
```go
func (h *AuthHandler) Logout(c echo.Context) error {
    // Expire the session cookie immediately
    cookie := new(http.Cookie)
    cookie.Name = "session"
    cookie.Value = ""
    cookie.Expires = time.Unix(0, 0)
    cookie.MaxAge = -1
    cookie.HttpOnly = true
    cookie.Path = "/"
    c.SetCookie(cookie)
    return c.Redirect(http.StatusTemporaryRedirect, "/")
}
```

Route: `e.GET("/logout", ah.Logout)` (public, no middleware needed — just clears cookie).

## Dashboard UX Polish

### Loading State
```js
// On page load, show spinner while fetching device list
async function loadDevices() {
    showSpinner();
    const resp = await fetch('/protected/devices');
    const devices = await resp.json();
    renderDevices(devices);
    hideSpinner();
}
```

### Empty State
When `devices` array is empty, render an empty state card:
```html
<div class="empty-state">
  <div class="empty-icon">📡</div>
  <p>No devices registered yet</p>
  <button onclick="toggleAddForm()">Register your first device</button>
</div>
```

### Toast Notifications
After `POST /protected/devices`:
- **Success**: green toast "Device registered successfully!" (auto-dismiss 3s)
- **Conflict**: orange toast "IMEI already registered"
- **Error**: red toast "Registration failed. Please try again."

### IMEI Input Formatting
- Show character count on IMEI input: `12 / 15`
- Client-side validation: disable submit if not exactly 15 digits
- Auto-strip non-digit chars on paste

### Last Seen Indicator
In device table:
- If `last_seen_at` is null → grey badge "Never connected"
- If < 5 min ago → green badge "Online"
- If < 1 hour ago → blue badge "Recently"
- Else → grey text formatted date

## Todo List
- [ ] Add `Logout` handler to `auth.go`
- [ ] Register `GET /logout` in `routes.go`
- [ ] Add loading spinner styles + JS to `dashboard.html`
- [ ] Add empty state component
- [ ] Add toast notification system (pure CSS + JS, ~30 lines)
- [ ] IMEI input character counter + client-side validation
- [ ] Last-seen status badge logic in JS

## Success Criteria
- `/logout` clears cookie and redirects to login page
- Dashboard shows spinner while loading, empty state when no devices
- Toasts appear and auto-dismiss correctly
- IMEI input rejects non-numeric and non-15-digit submissions client-side

## Security Considerations
- Logout **must** clear the cookie server-side (not just client-side JS delete).
- Toast messages never render raw server error text to the user — only safe static messages.
