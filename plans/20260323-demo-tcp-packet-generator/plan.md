---
title: "Demo TCP Packet Generator"
description: "Add a one-click demo feature to simulate smartwatch TCP sessions from the web UI"
status: completed
priority: P2
effort: 4h
issue:
branch: feat/add-claude-agent
tags: [feature, backend, frontend]
created: 2026-03-23
---

# Demo TCP Packet Generator

## Overview

Add a demo control panel to the Packet Inspector page that lets users simulate a real smartwatch TCP session against the local server — no external tools (e.g. Packet Sender) required.

## Context

- Brainstorm: [brainstorm-report.md](./brainstorm-report.md)
- TCP protocol: `internal/tcp/protocol/parser.go`
- Existing handler pattern: `internal/handler/http/device.go`

## Phases

| # | Phase | Status | Effort | Link |
|---|-------|--------|--------|------|
| 1 | Demo package (generator + session manager) | Done | 1.5h | [phase-01](./phase-01-demo-package.md) |
| 2 | HTTP demo handler | Done | 1h | [phase-02](./phase-02-http-handler.md) |
| 3 | Wiring (router + main) | Done | 30m | [phase-03-wiring.md](./phase-03-wiring.md) |
| 4 | UI — demo controls in packets.html | Done | 1h | [phase-04-ui.md](./phase-04-ui.md) |

## Dependencies

- Phases run sequentially: 1 → 2 → 3 → 4
- TCP server must be running for manual testing (already exists)
- Device must be registered in DB before starting a demo session
