# Phase 1: Dockerization

## Goal
Create the container configuration for the Health Data Platform (HDP), supporting both the HTTP (8080) and TCP (9090) servers.

## Tasks
1. Create a multi-stage **Dockerfile** for the Go backend.
2. Create a **docker-compose.yml** for the on-premise VM.

---

### Step 1: Dockerfile (project root)
Use a multi-stage build to minimize the final container size and reduce the attack surface.

```dockerfile
# Build Stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency manifests and install
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go

# Run Stage
FROM alpine:latest

WORKDIR /app

# Ensure logs directory exists
RUN mkdir -p /app/logs

# Copy binary from builder
COPY --from=builder /app/main .
# Copy web templates and static files if any
COPY --from=builder /app/web /app/web

EXPOSE 8080 9090

CMD ["./main"]
```

### Step 2: docker-compose.yml (for ~/homelab/hdp)
This file should be manually created on the target server.

```yaml
version: '3.8'

services:
  hdp:
    image: 192.168.1.222:3000/tuantt/hdp:latest
    container_name: hdp-api
    restart: unless-stopped
    ports:
      - "8080:8080"  # HTTP/API
      - "9090:9090"  # TCP Ingestion
    env_file:
      - .env         # External secret management
    volumes:
      - ./logs:/app/logs
    networks:
      - hdp-network

networks:
  hdp-network:
    driver: bridge
```

---

## Verification
- Run `docker build -t hdp:test .` locally and ensure it starts up correctly.
- Check that the `web/` templates are accessible within the container.
