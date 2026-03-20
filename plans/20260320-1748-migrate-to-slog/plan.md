---
title: "Structured Logging Migration (slog)"
description: "Migrating from standard 'log' to Go 1.21+ 'log/slog' with environment-based LOG_LEVEL and mandatory JSON format."
status: pending
priority: P2
effort: 2h
tags: [logging, slog, observability]
created: 2026-03-20
---

# Structured Logging Migration Plan

## Overview
Implement structured logging using `log/slog`. All logs will be in **JSON format**, but the **Log Level** will vary based on the `ENVIRONMENT`.

## Phases

| # | Phase | Status | Effort | Link |
|---|-------|--------|--------|------|
| 1 | Configuration Update | Done | 0.5h | [phase-01](./phase-01-config-logging.md) |
| 2 | Logger Initialization | Done | 0.5h | [phase-02-init-slog.md](./phase-02-init-slog.md) |
| 3 | Core Refactor (cmd/api) | Done | 0.5h | [phase-03-main-logging.md](./phase-03-main-logging.md) |
| 4 | Service Layer Insight | Done | 1h | [phase-04-service-logging.md](./phase-04-service-logging.md) |

## Key Rules
- **Layer Restriction**: Only refactor `cmd/api` and `internal/service`.
- **Mandatory Format**: Always use **JSON**.
- **Log Levels** (by `ENVIRONMENT` env var):
    - `DEV` (default): Level is `DEBUG`.
    - `PRODUCT`: Level is `INFO`.
