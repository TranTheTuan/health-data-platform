# Deployment Guide
## Health Data Platform

### Overview
This document describes the procedures for deploying the Health Data Platform (HDP) using containerized services orchestrated by Kubernetes. The architecture includes a dual-port server providing both HTTP/API and TCP/Ingestion services.

### Prerequisites
- Docker or compatible container runtime.
- Go version 1.22+ installed locally for compilation.
- Kubernetes cluster (e.g., Minikube, GKE, EKS).
- PostgreSQL database accessible from the cluster.
- Google OAuth 2.0 Credentials (client ID, client secret).

### Ports & Connectivity
- **HTTP Port 8080**: Public-facing dashboard and REST API.
- **TCP Port 9090**: Private or public-facing smartwatch data ingestion (depends on network topology).
- **PostgreSQL 5432**: Database connectivity.

### Containerization
1. **Dockerfiles**: The primary `Dockerfile` configurations are located in `/build/package`.
2. **Building the Image**:
   ```bash
   docker build -t hdp-api:latest -f build/package/Dockerfile .
   ```

### Configuration & Environment Variables
The following environment variables MUST be set for the server to operate:
- `SESSION_SECRET`: Random string for HMAC signing.
- `GOOGLE_CLIENT_ID`: OAuth Client ID.
- `GOOGLE_CLIENT_SECRET`: OAuth Client Secret.
- `GOOGLE_CALLBACK_URL`: Typically `http://localhost:8080/auth/google/callback` for dev.
- `DATABASE_URL`: `postgres://user:pass@host:port/dbname`.

### Deploying to Kubernetes
1. Review the Helm charts located in `/deployments/k8s`.
2. Create Kubernetes `Secrets` for database credentials and OAuth secrets.
3. Apply the deployment:
   ```bash
   kubectl apply -f deployments/k8s/
   ```
4. **Service Exposure**:
   - Create a LoadBalancer or NodePort service for both 8080 and 9090.
   - Note: Some Cloud Load Balancers may require specific configurations for persistent TCP connections on port 9090.

### CI/CD
Continuous Integration is configured in `/build/ci`. Each commit to the `main` branch triggers unit tests, linting (`golangci-lint`), and security scans. Tagged releases automatically trigger image building and pushing to the container registry.
