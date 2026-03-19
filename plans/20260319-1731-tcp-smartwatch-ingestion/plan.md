---
title: "TCP Server for Smartwatch Data Ingestion"
description: "TCP server implementing the IW protocol for smartwatch devices: IMEI auth, 13 packet types, PostgreSQL persistence linked to user accounts"
status: pending
priority: P1
effort: 9h
branch: main
tags: [backend, tcp, database, health-data, protocol, feature]
created: 2026-03-19
---

# TCP Server for Smartwatch Data Ingestion

## Overview

Implement a TCP server using Go's `net` stdlib that handles persistent long connections from GPS/health smartwatch devices over 2G/4G. Devices use a proprietary `IW` protocol with `#` frame terminator. The server authenticates devices via IMEI (AP00 login), parses 13 distinct packet types, stores health/location data in PostgreSQL linked to user accounts, and sends correct `IWBP##` responses to each packet.

Runs alongside the existing Echo HTTP server on a separate port (default `:9090`).

## Protocol Reference

**Frame format:** `IW<COMMAND><PAYLOAD>#`

| Command | Direction | Purpose | Server Reply |
|---------|-----------|---------|--------------|
| `AP00` | Device→Server | **Login / Auth** — sends IMEI (15 digits) | `IWBP00;<UTC_TIMESTAMP>,<TZ>#` |
| `AP01` | Device→Server | GPS+LBS+Status+BT+WiFi location | `IWBP01#` |
| `AP02` | Device→Server | Multi-base LBS location | `IWBP02#` |
| `AP03` | Device→Server | Heartbeat / keepalive | `IWBP03#` |
| `AP07` | Device→Server | Upload audio message | `IWBP07#` |
| `AP10` | Device→Server | Alarm + return address | `IWBP10#` |
| `AP49` | Device→Server | Heart rate | `IWBP49#` |
| `APHT` | Device→Server | Heart rate + blood pressure | `IWBPHT#` |
| `APHP` | Device→Server | HR + BP + SPO2 + blood sugar | `IWBPHP#` |
| `AP50` | Device→Server | Body temperature | `IWBP50#` |
| `AP97` | Device→Server | Sleep data | `IWBP97#` |
| `APWT` | Device→Server | Weather sync request | `IWBPWT#` |
| `APHD` | Device→Server | ECG upload | `IWBPHD#` |

**AP00 Login example:**
```
Device sends:  IWAP00353456789012345#
Server replies: IWBP00;20150101125223,8#
               (UTC timestamp YYYYMMDDHHmmss, server timezone offset)
```

**Auth flow:**
1. Device connects via TCP.
2. First packet MUST be `AP00` with IMEI.
3. Server looks up `devices` table by `imei`. If not found → reject + close.
4. Server responds with current UTC time. Connection is now authenticated.
5. All subsequent packets are processed against that `device_id` / `user_id`.

## Architecture

```
[Smartwatch] --TCP--> [:9090 TCP Server]
                           │
              [Frame reader (read until '#')]
                           │
              [Parser: strip 'IW', extract CMD+PAYLOAD]
                           │
              [Router: dispatch by command code]
                    ┌──────┴──────────┐
             [AP00 handler]    [Health/Location handlers]
                    │                  │
             [IMEI lookup]      [INSERT device_packets]
                    │                  │
             [Reply BP00]       [Reply BWxx#]
```

## Phases

| # | Phase | Status | Effort | Link |
|---|-------|--------|--------|------|
| 1 | Database Schema | Pending | 1.5h | [phase-01](./phase-01-db-schema.md) |
| 2 | Protocol Parser | Pending | 1.5h | [phase-02-protocol-parser.md](./phase-02-protocol-parser.md) |
| 3 | TCP Server Core | Pending | 2.5h | [phase-03-tcp-server.md](./phase-03-tcp-server.md) |
| 4 | Integration & Config | Pending | 1h | [phase-04-integration.md](./phase-04-integration.md) |
| 5 | Device Registration API | Pending | 2h | [phase-05-device-api.md](./phase-05-device-api.md) |

## Key Dependencies

- `github.com/jackc/pgx/v5/stdlib` — PostgreSQL driver
- Go stdlib `net` — TCP listener (no external framework)
- Existing `configs.Config` — extend for `DatabaseURL`, `TCPAddr`

## Key Design Decisions

- **IMEI = device identity**: no separate token; IMEI from AP00 is the auth credential.
- **`device_packets` table with JSONB `parsed_data`**: single table handles all 13 packet types. Flexible for future parsing without schema migration.
- **`#`-terminated frame reader**: reads byte-by-byte until `#`; handles partial TCP reads correctly.
- **Server MUST reply to every packet**: watches will hang or retry if no acknowledgment arrives.
