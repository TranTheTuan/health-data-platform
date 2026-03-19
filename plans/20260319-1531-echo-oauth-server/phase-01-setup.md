# Phase 1: Setup & Configuration

**Context Links**
- [Main Plan](./plan.md)

**Overview**
- Priority: P1
- Current status: Pending
- Brief description: Setup dependencies, environment variables, and initial OAuth2 Google configuration.

**Key Insights**
- We need the `golang.org/x/oauth2` library to manage the OAuth flow and `github.com/labstack/echo/v4` to host the endpoints.
- Environment variables must be securely injected via `.env` or application configs.

**Requirements**
- Add Go modules for Echo and OAuth2.
- Support configuring Google Client ID and Secret from the environment.

**Architecture**
- Configuration loading will happen cleanly, keeping credentials out of the source code.
- Auth config will be scoped in the `/internal/auth` package to prevent leaking secrets.

**Related Code Files**
- `[MODIFY]` `go.mod`
- `[MODIFY]` `go.sum`
- `[NEW]` `/configs/config.go`
- `[NEW]` `/internal/auth/google.go`

**Implementation Steps**
1. Run `go get github.com/labstack/echo/v4` and `go get golang.org/x/oauth2`.
2. Create `/configs/config.go` to load environment variables.
3. Create `/internal/auth/google.go` which initializes the `oauth2.Config` and provides URLs.

**Todo List**
- [ ] Install dependencies
- [ ] Create config loader
- [ ] Setup OAuth config object
- [ ] Add env vars to `.env` example

**Success Criteria**
- Libraries are downloaded successfully.
- Code compiles without errors.
- `oauth2.Config` can be correctly instantiated via environment variables.

**Security Considerations**
- Add `.env` to `.gitignore` to prevent secret leakage.

**Next Steps**
- Move to [Phase 2](./phase-02-implementation.md) to implement the server handlers.
