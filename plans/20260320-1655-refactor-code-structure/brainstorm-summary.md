# Brainstorm Summary: Clean Architecture Refactoring

## Problem Statement & Requirements
- **Scattered Domains**: Currently, logic for single entities (e.g., Device) is split across independent `tcp/`, `api/`, and `device/` packages, making lifecycles hard to trace.
- **Dependency Inversion Violations**: Repositories are intertwined with delivery layers (TCP/HTTP) rather than adhering purely to domain models.
- **Goal**: Establish a clear, orderly Clean Architecture (Layer-Oriented) structure where higher layers strictly depend on lower-layer interfaces.

## Evaluated Approaches
1. **Option 1: Clean Architecture (Layer-Oriented)**
   - Folders logically structured by layer (`domain`, `repository`, `service`, `delivery/http`, `delivery/tcp`).
   - *Pros*: Go compiler strictly enforces dependency flow (Delivery -> Service -> Repo). Eliminates cyclic dependencies.
   - *Cons*: Code for a single feature is distributed across multiple layout folders.
2. **Option 2: Clean Architecture (Domain-Oriented)**
   - Folders structured by feature (`device`, `auth`), with layering rules applied within the files of that package.
   - *Pros*: High feature cohesion.
   - *Cons*: Highly vulnerable to cyclic dependency issues in Go, particularly when two domain services need to cross-communicate.

## Final Recommended Solution & Rationale
**Selected: Option 1 (Clean Architecture - Layer-Oriented)**
This is the undeniable optimal choice for Go, based primarily on the user's keen insight. By physically decoupling the Domain (Core Interfaces), Repository (Data Access), Service (Business Logic), and Delivery (Transport) operations into distinct packages, it completely mitigates Go's strict cyclic dependency rules. The unidirectional dependency graph naturally guarantees that lower layers (e.g., Repositories) cannot mistakenly import or rely on higher operations (e.g., Services or Handlers).

## Implementation Considerations & Risks
- **Major File Movement**: Complete structural overhaul. For instance, the existing `tcp` directory must be broken down into `delivery/tcp` (handlers), `repository` (saving packages), and `service` (protocol ingestion rules).
- **Interface Definitions**: Must cleanly extract and define all functional boundaries into a centralized `domain` package (e.g., `DeviceRepository`, `DeviceService`).
- **Dependency Injection**: The `cmd/api/main.go` file will have to take over the responsibility of manually wiring Repositories into Services, and those Services into the Delivery controllers.

## Success Metrics & Validation Criteria
- Go builds perfectly without any import cycle errors.
- Delivery packages (`delivery/http`, `delivery/tcp`) contain 0% business logic and only map data to/from the framework.
- Service packages contain 0% TCP/HTTP request parsing and 0% SQL executions.
- Repository packages securely implement interfaces strictly defined within the `domain` package.

## Next Steps & Dependencies
- Draft concrete models and interfaces within `internal/domain`.
- Migrate SQL interactions into `internal/repository`.
- Re-implement core workflows within `internal/service`.
- Reposition server setups to `internal/delivery/http` and `internal/delivery/tcp`.
