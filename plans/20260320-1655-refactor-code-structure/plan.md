---
title: "Clean Architecture Refactoring Plan (Revised v2)"
description: "Refactor codebase: Domain, DTOs, Repo, Service, Handler, Delivery."
status: pending
priority: P1
effort: 7h
tags: [refactor, architecture, structural, dto]
created: 2026-03-20
---

# Feature Implementation Plan

## Overview

Refining the architecture with a dedicated DTO layer to ensure Domain Objects are never exposed directly to external clients.

1. `internal/domain`: Pure data structs (Entities).
2. `internal/dto`: Request and Response structures for external communication.
3. `internal/repository`: Data access interfaces and implementations (Entity based).
4. `internal/service`: Business logic (translates between Entities and DTOs).
5. `internal/handler`: Transport-specific logic (uses DTOs).
6. `internal/delivery`: Infrastructure/Framework binding.

## Phases

| # | Phase | Status | Effort | Link |
|---|-------|--------|--------|------|
| 1 | Core Domain Models | Done | 0.5h | [phase-01](./phase-01-core-domain.md) |
| 2 | Data Transfer Objects (DTOs) | Done | 0.5h | [phase-02-dtos.md](./phase-02-dtos.md) |
| 3 | Data Repositories | Done | 1h | [phase-03-repositories](./phase-03-repositories.md) |
| 4 | Business Services | Done | 1.5h | [phase-04-services](./phase-04-services.md) |
| 5 | Request/Response Handlers | Done | 1h | [phase-05-handlers](./phase-05-handlers.md) |
| 6 | Transport Deliveries | Done | 1h | [phase-06-delivery](./phase-06-delivery.md) |
| 7 | Dependency Injection Wiring | Done | 1h | [phase-07-wiring](./phase-07-wiring.md) |

## Dependencies
- Handlers use DTOs.
- Services use Domain Entities internally but return/accept DTOs for external layers.
- Repositories strictly use Domain Entities.
