# Phase 2: Backend Wiring

**Context Links**
- [Main Plan](./plan.md)
- [Phase 1: HTML Templates](./phase-01-html-templates.md)

## Overview
- Priority: P1
- Effort: 1h
- Status: Pending
- Wire Echo to render the HTML templates, update handler methods to use `c.Render()`, and add a `DashboardHandler`.

## Key Insights
- Echo does not have a built-in template renderer — must implement the `echo.Renderer` interface with `html/template`.
- `html/template` loads templates from disk using `template.ParseGlob("web/templates/*.html")`. Call once at startup, pass renderer to Echo.
- The `Home` handler currently returns inline HTML — replace with `c.Render(200, "login.html", nil)`.
- The `ProtectedEndpoint` handler becomes `DashboardHandler` returning `c.Render(200, "dashboard.html", data)`.
- Session needs to store user email as well as user ID, so the dashboard can show the user's email. Two options:
  - **Option A (chosen)**: Store `userID|userEmail` in the session cookie value, split in middleware. Simple, no DB needed.
  - Option B: Lookup Google userinfo on every dashboard load (extra HTTP call).

## Session Cookie Change
Current cookie value: `<google_user_id>` (signed)
Updated cookie value: `<google_user_id>|<email>` (signed together before HMAC)

Auth middleware extracts both fields and sets `user_id` + `user_email` in context.

## Related Code Files
- `[NEW]` `internal/api/renderer.go` — implements `echo.Renderer` using `html/template`
- `[MODIFY]` `internal/api/handlers/auth.go`:
  - `GoogleCallback`: update signed cookie value to `userID|email`
  - `Home`: replace inline HTML with `c.Render(200, "login.html", nil)`
  - `ProtectedEndpoint` → renamed `Dashboard`: render `dashboard.html` with `DashboardData`
  - `AuthMiddleware`: extract `user_email` from cookie and set in context
- `[MODIFY]` `internal/api/routes.go`:
  - Wire renderer onto Echo instance
  - `protected.GET("")` → `ah.Dashboard`
  - Redirect `/protected` → `/dashboard` for cleaner URL (optional)

## Implementation Steps

### `internal/api/renderer.go`
```go
package api

import (
    "html/template"
    "io"
    "github.com/labstack/echo/v4"
)

type TemplateRenderer struct {
    templates *template.Template
}

func NewTemplateRenderer(pattern string) (*TemplateRenderer, error) {
    t, err := template.ParseGlob(pattern)
    if err != nil { return nil, err }
    return &TemplateRenderer{templates: t}, nil
}

func (r *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
    return r.templates.ExecuteTemplate(w, name, data)
}
```

### `cmd/api/main.go` additions
```go
renderer, err := api.NewTemplateRenderer("web/templates/*.html")
if err != nil { log.Fatal(err) }
e.Renderer = renderer
```

### Updated `auth.go` — GoogleCallback
```go
// Store "userID|email" signed together
cookieVal := fmt.Sprintf("%s|%s", userID, email)
signedValue := auth.Sign(cookieVal, h.cfg.SessionSecret)
```

### Updated `AuthMiddleware`
```go
raw, err := auth.Verify(sessionCookie.Value, h.cfg.SessionSecret)
parts := strings.SplitN(raw, "|", 2)
c.Set("user_id", parts[0])
if len(parts) > 1 { c.Set("user_email", parts[1]) }
```

### `Dashboard` handler
```go
type DashboardData struct { UserID, Email string }

func (h *AuthHandler) Dashboard(c echo.Context) error {
    return c.Render(http.StatusOK, "dashboard.html", DashboardData{
        UserID: c.Get("user_id").(string),
        Email:  c.Get("user_email").(string),
    })
}
```

## Todo List
- [ ] Create `internal/api/renderer.go`
- [ ] Update `auth.go`: `Home`, `GoogleCallback`, `AuthMiddleware`, add `Dashboard`
- [ ] Update `cmd/api/main.go`: register renderer
- [ ] Update `routes.go`: point `protected.GET("")` → `Dashboard`

## Success Criteria
- `GET /` returns HTML login page (200, Content-Type: text/html)
- `GET /protected` after login returns HTML dashboard (200)
- User email visible in sidebar
- No `c.String()` / inline HTML left in `Home` or `ProtectedEndpoint`

## Security Considerations
- `html/template` auto-escapes user data — `UserID`, `Email` are safe to render.
- Session cookie still HMAC-signed — the pipe `|` separator is inside the signed payload, so it cannot be forged.

## Next Steps
- [Phase 3: Polish & Logout](./phase-03-polish-and-logout.md)
