# Codebase Summary
## Health Data Platform

### Repository Structure
This project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout).

- `/cmd`: Main entry points for the applications.
  - `/cmd/api`: The main combined HTTP and TCP server binary.
- `/internal`: Private library code, the core of the business logic.
  - `/internal/api`: Echo HTTP server handlers, renderer, and routes.
    - `/internal/api/handlers`: Auth (Google OAuth) and Device management endpoints.
  - `/internal/auth`: Session signing and Google UserInfo handling.
  - `/internal/db`: PostgreSQL database connection management.
  - `/internal/demo`: Demo TCP packet generator for testing and live demonstrations.
    - `packet_generator.go`: Generates random IW protocol frames (login, health metrics, GPS, etc.).
    - `session_manager.go`: Manages persistent TCP connections for demo sessions.
  - `/internal/tcp`: TCP server for smartwatch data ingestion.
    - `/internal/tcp/protocol`: Robust frame scanner and parser for the `IW` protocol.
- `/web`: Web resources and templates.
  - `/web/templates`: Professional, responsive HTML templates (`base.html`, `dashboard.html`, `login.html`) utilizing Vanilla CSS and modern JavaScript.
- `/configs`: Configuration templates and environment variable management.
- `/docs`: Detailed system documentation (this directory).
- `/plans`: Implementation plans and archived blueprints for specific features.

### Current State
As of March 23, 2026, the Health Data Platform is a functional, dual-server application:
1.  **HTTP/API Server**: Fully functional Google OAuth login flow and a modern Dashboard for user-device registration with a Packet Inspector page.
2.  **TCP Ingestion Server**: A robust, noise-resilient TCP server that authenticates smartwatches via IMEI (AP00) and persists 13+ types of health/location packets.
3.  **Persistence**: Uses PostgreSQL to store user-to-device mappings and a flexible `device_packets` table for raw and parsed protocol data.
4.  **Demo Feature**: Built-in TCP packet generator for testing, demonstrations, and development without physical smartwatch hardware.
