---
name: health-data-platform project conventions
description: Key architectural patterns, security practices, and conventions observed in this codebase
type: project
---

Auth ownership guard pattern: every protected handler reads `user_id` from `c.Get("user_id").(string)` first, then verifies device ownership via `DeviceService.ListDevices`. All four demo endpoints follow this correctly.

TCP protocol frame format: `IW<4-char-CMD><PAYLOAD>#` — no newline terminator, `#` is delimiter. Server replies use `IWBP<CMD>` prefix. Login ack is `IWBP00;<timestamp>,0#`.

**Why:** IW smartwatch protocol; ScanFrame in protocol/parser.go splits on `#`.

**How to apply:** When reviewing TCP-related code, verify frame construction matches `IW+CMD+PAYLOAD+#` and ack checks use `strings.HasPrefix(ack, "IWBP00")`.

Routes under `protected` group automatically get `AuthMiddleware`. Demo routes correctly placed there — no new middleware needed for new demo-style endpoints.

`math/rand` global source is auto-seeded in Go 1.20+. Check `go.mod` Go version before flagging unseeded rand usage.
