---
title: "Gitea CI/CD Docker Deployment"
description: "Deploy Health Data Platform to an on-premise server using Gitea Registry, Gitea Actions, and Docker Compose."
status: in-progress
priority: P1
effort: 4h
tags: [devops, gitea, docker, ci-cd, deployment]
created: 2026-03-22
---

# Feature Implementation Plan

## Overview
Automate the deployment process for the Health Data Platform (HDP) using Gitea's internal ecosystem. The deployment targets an on-premise VM via local network (LAN) communication to bypass Cloudflare Zero Trust for the Docker Registry.

1. **Build Phase**: Build a multi-stage Docker image and push to Gitea Container Registry via local LAN.
2. **Deploy Phase**: SSH into the target VM to pull and restart the stack.
3. **Secret Management**: Use a local `.env` on the VM for sensitive data.
4. **Manual Migrations**: User handles database migrations manually before/after deployment.

## Phases

| # | Phase | Status | Effort | Link |
|---|-------|--------|--------|------|
| 1 | Dockerization | done | 1h | [phase-01](./phase-01-dockerization.md) |
| 2 | Gitea Action Workflow | done | 1h | [phase-02-gitea-workflow.md](./phase-02-gitea-workflow.md) |
| 3 | Server Environment Setup | pending | 1h | [phase-03-server-setup.md](./phase-03-server-setup.md) |
| 4 | Verification & Handoff | pending | 1h | [phase-04-verification.md](./phase-04-verification.md) |

## Dependencies
- **Gitea Actions Runner**: Must be local (same machine or LAN as Gitea server).
- **Target VM**: Must be accessible via SSH from the Gitea Runner.
- **Docker**: Installed on both builder and target.
