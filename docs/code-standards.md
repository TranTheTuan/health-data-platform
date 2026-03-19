# Code Standards & Guidelines
## Health Data Platform

### General Guidelines
- Code should be written in clean, idiomatic Go.
- Use `gofmt` to format all Go source code.
- Ensure `go vet` and static analysis tools (like `staticcheck` or `golangci-lint`) pass cleanly without warnings.
- Keep functions small, focused, and testable.
- Document exported functions, types, constants, and variables using standard Go comments.

### Modularization Principles
- **200 LOC Rule**: Consider modularizing code files that exceed 200 lines of code.
- Analyze logical separation boundaries (functions, classes, concerns).
- Before creating a new module, check if an existing module can be reused or extended.
- Use explicit and kebab-case naming conventions for non-Go files or long descriptive names to ensure they are self-documenting.

### Package Structure
- Follow the boundary rules established by the Standard Go Layout:
  - Do not put reusable business logic in `/cmd`.
  - Prefer `/internal` for business logic to prevent external dependencies from tightly coupling to your private code.
  - Only use `/pkg` for libraries explicitly designed to be imported by third parties.

### Testing
- Write table-driven unit tests for all business logic.
- Use integration tests for database and external API interactions.
- Strive for high test coverage (>80%) on core libraries and internal application logic.
