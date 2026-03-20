# Phase 1: HTML Templates

**Context Links**
- [Main Plan](./plan.md)

## Overview
- Priority: P1
- Effort: 1.5h
- Status: Pending
- Create beautiful, professional HTML templates for the landing page and authenticated dashboard.

## Key Insights
- **No JS framework** — pure HTML/CSS + minimal vanilla JS for the device registration form. Keeps the project lean (KISS).
- Go's `html/template` automatically escapes XSS — safe for rendering user data.
- Templates live in `web/templates/` directory. Echo renders them via a custom `TemplateRenderer`.
- **Design language**: dark, modern health/tech aesthetic — dark navy background, electric blue accent, glassmorphism cards, smooth hover transitions. Google's Material Design icon font for icons.

## Design Specification

### Color System
```css
--bg:          #0a0f1e  /* deep navy */
--surface:     rgba(255,255,255,0.05)  /* glass card */
--border:      rgba(255,255,255,0.1)
--accent:      #4f8ef7  /* electric blue */
--accent-glow: #4f8ef755
--text:        #e8eaf0
--text-muted:  #8892a4
--success:     #34d399
--error:       #f87171
```

### Typography
- Font: Inter (Google Fonts CDN)
- Headings: 600–700 weight
- Body: 400–500 weight, `--text-muted`

## Templates

### `web/templates/base.html` — shared base layout
Contains: `<head>` with meta tags, font import, shared CSS variables, `{{block "content" .}}` slot.

### `web/templates/login.html` — landing/login page
**Elements:**
- Animated gradient background with floating blobs
- Center card: platform logo + tagline "Health Data Platform"
- Subtitle: "Securely manage your smartwatch health data"
- "Sign in with Google" button (Google icon + styled button, electric blue, hover glow)
- Footer: "Powered by Go & Echo"

**Behavior:** Clicking button → `<a href="/login">` redirect.

### `web/templates/dashboard.html` — authenticated dashboard
**Layout:** Sidebar navigation + main content area

**Sidebar:**
- Platform logo
- Nav items: Dashboard, Devices (active), Settings (placeholder)
- Bottom: user avatar circle + email + "Logout" button

**Main content:**
- Header: "My Devices" + "Add Device" button (opens inline form)
- Stats row: total devices count card, last active card
- Device list table: columns `Name | IMEI | Last Seen | Status | Actions`
- "Add Device" panel (hidden by default, toggled via JS):
  - Input: Device Name (text)
  - Input: IMEI Number (15-digit, validated)
  - Submit button → `fetch POST /protected/devices` → refresh list on success
  - Error/success toast notification

**Template data shape** passed from Go handler:
```go
type DashboardData struct {
    UserID  string
    Email   string   // fetched from session or stored separately
}
```
Device list loaded dynamically via `fetch GET /protected/devices` on page load (no server-side list needed in template).

## Related Code Files
- `[NEW]` `web/templates/base.html`
- `[NEW]` `web/templates/login.html`
- `[NEW]` `web/templates/dashboard.html`

## Implementation Steps
1. Create `web/templates/` directory.
2. Write `base.html` with shared CSS, fonts, CSS variables.
3. Write `login.html` (extends base) — full-page centered login card with animated background.
4. Write `dashboard.html` (extends base) — sidebar layout + devices section + JS for fetch calls.

## Todo List
- [ ] Create `web/templates/base.html`
- [ ] Create `web/templates/login.html`
- [ ] Create `web/templates/dashboard.html`

## Success Criteria
- Templates render without error in browser (no broken CSS, no console errors)
- Login page: "Sign in with Google" button is visible and styled
- Dashboard: device list renders from API, add-device form works end-to-end
- Responsive: looks good at 1280px wide minimum

## Next Steps
- [Phase 2: Backend Wiring](./phase-02-backend-wiring.md)
