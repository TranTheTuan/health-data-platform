---
title: "Wonlex Protocol Refactor"
description: "Refactor TCP parser, handler, demo generator, and UI from old IW protocol to Wonlex [3G*deviceID*LEN*CMD,...] format"
status: complete
priority: P1
effort: 6h
issue:
branch: refactor/update-logic-with-new-doc
tags: [refactor, backend, protocol, frontend]
created: 2026-03-24
completed: 2026-03-24
---

# Wonlex Protocol Refactor

## Overview

Replace the entire IW protocol stack (`IW<CMD><PAYLOAD>#`) with the correct Wonlex/Setracker protocol format: `[manufacturer*deviceID*contentLength*content]`.

## Context

- Brainstorm: current conversation (2026-03-24)
- Old protocol parser: `internal/tcp/protocol/parser.go`
- Old TCP handler: `internal/handler/tcp/handler.go`
- Old demo generator: `internal/demo/packet_generator.go`
- Demo service: `internal/service/demo_service.go`
- Device service: `internal/service/device_service.go`
- DB migration: `internal/migrations/001_initial_schema.up.sql`
- UI filter dropdown: `web/templates/packets.html`

## Protocol Mapping

### Old → New Frame Format

| Old (IW) | New (Wonlex) |
|----------|-------------|
| `IW<CMD><PAYLOAD>#` | `[manufacturer*deviceID*contentLength*content]` |
| Delimiter: `#` | Delimiter: `]` (start: `[`) |
| Prefix: `IW` (2 chars) | Manufacturer: `3G` (2 chars) |
| IMEI: 15-digit in AP00 payload | Device ID: 10-digit in frame header |
| CMD: 4-char fixed | CMD: variable length (UD, AL, CONFIG, UD2) |

### Command Mapping

| Old Cmd | New Cmd | Description | Server Reply |
|---------|---------|-------------|--------------|
| AP00 | (login via LK) | Keep-alive/Link | `[CS*deviceID*LEN*LK]` |
| AP01 | UD | Location data (GPS+LBS+WiFi) | No reply needed |
| AP02 | UD2 | Blind spot data fill-in | No reply |
| AP10 | AL | Alarm data report | `[3G*deviceID*LEN*AL]` |
| — | CONFIG | Device config report | `[CS*deviceID*LEN*CONFIG,result]` |

### Key Protocol Rules

1. **Frame structure:** `[manufacturer*deviceID*contentLength*content]`
2. **Manufacturer:** fixed 2 bytes (`3G`, `CS` for server replies)
3. **Device ID:** 10 digits (derived from the original IMEI)
4. **Content length:** 4-byte HEX ASCII (e.g. `00BC` = 188 bytes), high nibble first
5. **Content:** `CMD` or `CMD,payload_data`
6. **Server reply prefix:** `CS` (not `3G`) for config acknowledgments

## Phases

| # | Phase | Status | Effort | Link |
|---|-------|--------|--------|------|
| 1 | Protocol parser rewrite | Complete | 2h | [phase-01-protocol-parser.md](./phase-01-protocol-parser.md) |
| 2 | TCP handler refactor | Complete | 1.5h | [phase-02-tcp-handler.md](./phase-02-tcp-handler.md) |
| 3 | Demo generator + service refactor | Complete | 1.5h | [phase-03-demo-refactor.md](./phase-03-demo-refactor.md) |
| 4 | IMEI/Device ID + UI updates | Complete | 1h | [phase-04-imei-and-ui.md](./phase-04-imei-and-ui.md) |

## Dependencies

- Phases run sequentially: 1 → 2 → 3 → 4
- Phase 1 is the foundation — all other phases depend on the new `Frame` struct
- DB schema change (IMEI column width) in phase 4 must be a migration

## Implementation Rules
- Apply `YAGNI` — only implement commands present in the protocol doc
- Apply `KISS` — one parser, one reply builder, one handler loop
- Apply `DRY` — reuse `BuildReply` for all server responses
- Each file ≤ 200 LOC
- Always log on errors with `slog`
