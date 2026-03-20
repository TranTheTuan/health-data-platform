---
title: "UI Redesign & Device Registration Dashboard"
description: "Replace inline HTML strings with a polished, professional web UI including a dashboard with Google login, device registration, and device list pages"
status: completed
priority: P1
effort: 4h
branch: main
tags: [frontend, ui, feature, backend]
created: 2026-03-20
---

# UI Redesign & Device Registration Dashboard

## Overview

The current UI is raw inline Go HTML strings — a plain `<a>Login</a>` link and bare text responses. This plan replaces them with a professional, self-contained HTML/CSS/JS web interface served by Echo:

- **Landing page** (`/`) — polished login page with "Sign in with Google" button
- **Dashboard** (`/dashboard`) — authenticated users see their devices and can add new ones
- **Device registration form** — inline form to register a smartwatch IMEI
- **Logout** — clear session cookie and redirect to landing

All pages are **server-rendered HTML** served directly by Go/Echo using `html/template`. No JS framework, no build step. CSS embedded in each template.

## Architecture

```
/                    → Home handler → login.html template
/login               → GoogleLogin (redirect, existing)
/auth/google/callback → GoogleCallback (existing) → redirect to /dashboard
/dashboard           → DashboardHandler (protected) → renders dashboard.html
/dashboard/devices   → DeviceHandler.Register POST (JSON API, existing)
/logout              → LogoutHandler → clear cookie → redirect to /
```

The existing JSON API endpoints (`POST /protected/devices`, `GET /protected/devices`) are **reused** via `fetch()` calls from the dashboard JS.

## Phases

| # | Phase | Status | Effort | Link |
|---|-------|--------|--------|------|
| 1 | HTML Templates | Completed | 1.5h | [phase-01](./phase-01-html-templates.md) |
| 2 | Backend Wiring | Completed | 1h | [phase-02-backend-wiring.md](./phase-02-backend-wiring.md) |
| 3 | Polish & Logout | Completed | 1h | [phase-03-polish-and-logout.md](./phase-03-polish-and-logout.md) |

## Key Dependencies

- Go `html/template` stdlib — secure, no new deps
- Echo `c.Render()` — requires implementing `echo.Renderer` interface
- CSS: modern design via embedded styles in templates (no CDN required for core layout)
- Google Fonts Inter (CDN) for typography
- Existing auth + device API handlers — no changes needed to business logic
