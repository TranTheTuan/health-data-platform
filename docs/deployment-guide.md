# Deployment Guide
## Health Data Platform

### Overview
This document describes the intended procedures for deploying the Health Data Platform. The system is designed to be deployed as containerized services, orchestrated by Kubernetes (or similar container platforms).

### Prerequisites
- Docker or compatible container runtime.
- Go version 1.22+ installed locally for compilation.
- Kubernetes cluster (e.g., Minikube, GKE, EKS) if testing orchestration.
- PostgreSQL database accessible from the cluster.

### Containerization
1. **Dockerfiles**: The primary `Dockerfile` configurations are located in `/build/package`.
2. **Building the Image**: Run `docker build -t health-data-platform:latest -f build/package/Dockerfile .` from the root directory.

### Configuration
- All configurations should be provided via environment variables or loaded from configuration files specified in the `/configs` directory.
- Use `confd` or Kubernetes `ConfigMaps` to mount environment settings securely into the running container.

### Deploying to Kubernetes
1. Review the Helm charts or manifest files located in the `/deployments/k8s` directory.
2. Ensure you have the `kubectl` CLI configured correctly.
3. Validate connection secrets (e.g., db passwords, API keys) are secured using Kubernetes `Secrets`.
4. Apply the deployment:
   ```bash
   kubectl apply -f deployments/k8s/
   ```

### CI/CD
Continuous Integration is configured in `/build/ci`. Each commit to the `main` branch triggers unit tests, linting (`golangci-lint`), and security scans. Tagged releases automatically trigger image building and pushing to the container registry.
