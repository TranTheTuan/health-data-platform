# Health Data Platform (HDP)

A robust, high-performance Go-based backend platform designed for secure health data management and smartwatch device ingestion.

## Key Features

- **Dual-server architecture**:
  - **Echo HTTP/API Server (Port 8080)**: Handles user authentication (Google OAuth), RESTful APIs, and a modern web dashboard for device management.
  - **TCP Ingestion Server (Port 9090)**: Implements the `IW` protocol for robust, persistent smartwatch data stream ingestion with IMEI-based authentication.
- **Secure Authentication**:
  - Google OAuth 2.0 integration for seamless user onboarding.
  - HMAC-signed session cookies for persistent, secure user sessions.
- **Smartwatch Protocol (IW)**:
  - Custom `bufio.Scanner` based protocol parser for robust, noise-resilient packet ingestion.
  - Handles 13+ packet types (GPS, Heart rate, BP, SPO2, etc.).
- **Modern Web Dashboard**:
  - Professional, dark-themed UI built with Glassmorphism principles.
  - Device registration with real-time IMEI validation and interactive feedback.
  - Real-time device list and status tracking.
- **Database Persistence**:
  - PostgreSQL with JSONB for flexible, high-performance health packet storage.

## Tech Stack

- **Go (1.22+)**: Primary language for the backend servers.
- **Echo**: High-performance HTTP framework.
- **PostgreSQL**: Robust relational datastore.
- **Vanilla JS/CSS/HTML**: Used for the dashboard (keeping dependencies light and performance high).
- **Google OAuth 2.0**: For user identity management.

## Getting Started

1.  Clone the repository.
2.  Set up your `.env` with Google OAuth credentials and PostgreSQL connection string.
3.  Run the platform:
    ```bash
    go run cmd/api/main.go
    ```
4.  Access the dashboard at `http://localhost:8080/dashboard`.

## Documentation

Comprehensive documentation can be found in the `./docs` directory:
- [Codebase Summary](docs/codebase-summary.md)
- [System Architecture](docs/system-architecture.md)
- [Project Roadmap](docs/project-roadmap.md)
- [Code Standards](docs/code-standards.md)
- [Deployment Guide](docs/deployment-guide.md)
