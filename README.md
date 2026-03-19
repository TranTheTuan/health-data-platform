# Health Data Platform

This project defines the core architecture and structure for the Health Data Platform. Built natively in Go, the repository leverages the standard Go layout to facilitate modular, scalable, and secure operations within the health data ecosystem.

## Documentation Overview

Detailed documentation is stored in the `/docs` directory. Please review them carefully when joining the project or updating the architecture.

- [Project Overview & PDR](docs/project-overview-pdr.md)
- [System Architecture](docs/system-architecture.md)
- [Codebase Summary](docs/codebase-summary.md)
- [Code Standards](docs/code-standards.md)
- [Project Roadmap](docs/project-roadmap.md)
- [Deployment Guide](docs/deployment-guide.md)
- [Design Guidelines](docs/design-guidelines.md)

## Repository Layout Basics

Based on the [Go standard project layout](https://github.com/golang-standards/project-layout), here's where things are:

- **`/cmd`**: Main application executables.
- **`/internal`**: Private application code and business logic.
- **`/pkg`**: Library code safe for external use.
- **`/api`**: API contracts, swagger files, protobufs.
- **`/build`**: CI/CD and packaging.
- **`/deployments`**: Deployment configurations.
- **`/configs`**: Configuration templates.
- **`/scripts`**: Automation utilities.

## Getting Started

1. Clone the repository and initialize the module.
   `go mod init health-data-platform` (or run `go mod tidy` if already setup).
2. Follow the directory structure to add logic to `/internal` and entrypoints to `/cmd`.
3. Consult `AGENTS.md` and the rules located in `.agent/rules/` for AI integration and team coordination guidelines.
