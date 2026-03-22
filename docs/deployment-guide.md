# Deployment Guide
## Health Data Platform

### Overview
This document describes the procedures for deploying the Health Data Platform (HDP) using containerized services via Docker Compose for an on-premise server. The architecture includes a dual-port server providing both HTTP/API and TCP/Ingestion services.

### Prerequisites
- Docker and Docker Compose installed on the target server.
- Go version 1.22+ installed locally for compilation.
- Gitea Server with Container Registry enabled.
- PostgreSQL database accessible from the container network.
- Google OAuth 2.0 Credentials (client ID, client secret).

### Ports & Connectivity
- **HTTP Port 8080**: Public-facing dashboard and REST API.
- **TCP Port 9090**: Private or public-facing smartwatch data ingestion (depends on network topology).
- **PostgreSQL 5432**: Database connectivity.

### Containerization
1. **Dockerfiles**: The primary multi-stage `Dockerfile` is located at the root of the project.
2. **Building the Image**:
   ```bash
   docker build -t hdp-api:latest .
   ```

### Configuration & Environment Variables
The following environment variables MUST be set for the server to operate:
- `SESSION_SECRET`: Random string for HMAC signing.
- `GOOGLE_CLIENT_ID`: OAuth Client ID.
- `GOOGLE_CLIENT_SECRET`: OAuth Client Secret.
- `GOOGLE_CALLBACK_URL`: Typically `http://localhost:8080/auth/google/callback` for dev.
- `DATABASE_URL`: `postgres://user:pass@host:port/dbname`.

### Deploying via Docker Compose (On-Premise)
1. **Configuring the server**: Create a directory such as `~/homelab/hdp` on the target server.
2. **Docker Compose**: Create a `docker-compose.yml` file pointing to your Gitea Registry image.
3. **Environment**: Create a `.env` file within the same directory ensuring all environment variables mentioned above are populated.
4. **Running**:
   ```bash
   docker compose pull && docker compose up -d
   ```

### CI/CD
Continuous Deployment is orchestrated using Gitea Actions. 
The workflow configuration is defined in `.gitea/workflows/deploy.yml`. 
Each published git tag matching `v*` triggers a process that:
1. Builds the Docker Image via a locally hosted action runner.
2. Pushes the Docker image to the Gitea Container Registry over the LAN.
3. Deploys the container to the destination VM by establishing an SSH connection and running `docker compose up -d`.
