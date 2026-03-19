# Phase 2: Auth Hooks & Echo Server

**Context Links**
- [Main Plan](./plan.md)
- [Phase 1: Setup](./phase-01-setup.md)

**Overview**
- Priority: P1
- Current status: Pending
- Brief description: Implement the Echo HTTP server with login handlers, OAuth2 callback handlers, and a protected route.

**Key Insights**
- The callback route must exchange the authorization code for an access token.
- User data can be retrieved from Google's `https://www.googleapis.com/oauth2/v2/userinfo` endpoint using the access token.
- Sessions/cookies need to be used to maintain the logged-in state of the user after the OAuth loop completes.

**Requirements**
- Implement `/login` to redirect to Google's OAuth consent screen.
- Implement `/auth/google/callback` to handle Google's redirect and verify state.
- Implement `/protected` endpoint which verifies if the user is logged in.
- Implement `/` an index to show public info or links to login.

**Architecture**
- Application entry point is `/cmd/api/main.go`.
- Route handlers will live in `/internal/api/handlers/auth.go`.
- Session management (cookie-based for simplicity) built into Echo middleware.

**Related Code Files**
- `[NEW]` `/cmd/api/main.go`
- `[NEW]` `/internal/api/handlers/auth.go`
- `[NEW]` `/internal/api/routes.go`

**Implementation Steps**
1. Initialize the Echo server in `main.go`.
2. Define the HTTP routes in `routes.go` (e.g., `e.GET("/login", handlers.GoogleLogin)`).
3. In `auth.go`, write `GoogleLogin(c echo.Context)` to construct the `oauth2` AuthURL and redirect the user. Generate a random CSRF "state" token and set it in a secure cookie.
4. Write `GoogleCallback(c echo.Context)` to read the state from the query, verify it against the cookie, and then call `Exchange` to get the access token.
5. In `GoogleCallback`, make a standard HTTP request to `https://www.googleapis.com/oauth2/v2/userinfo` adding the `Authorization: Bearer <token>` header to fetch user email/id.
6. Set a session cookie recognizing the user.
7. Implement an Echo middleware to restrict `/protected` route based on the session cookie.

**Todo List**
- [ ] Setup `main.go` basic echo skeleton.
- [ ] Create `/login` redirect handler.
- [ ] Create `/auth/google/callback` handler and verify state.
- [ ] Fetch User Info from Google API.
- [ ] Store session cookie.
- [ ] Create Protect middleware.
- [ ] Create `/protected` test route.

**Success Criteria**
- A user can navigate to `/login`, authorize via Google, and be redirected back.
- The user info is successfully fetched and logged/stored.
- The user can access the `/protected` endpoint only after this flow.

**Security Considerations**
- The OAuth2 `state` parameter must be random and validated upon callback to prevent CSRF.
- Cookies used for sessions MUST be `HttpOnly` and `Secure`.
- Don't hardcode client secrets.

**Next Steps**
- After implementation, test the flow via browser manually.
