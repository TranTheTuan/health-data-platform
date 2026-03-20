# System Architecture
## Health Data Platform

### High-Level Architecture
The Health Data Platform is a service-oriented Go backend with a dual-server architecture that provides a separation of concerns between high-throughput data ingestion (TCP) and user dashboard/API management (HTTP).

### Architecture Components
1. **API Gateway / Entrypoints (`/cmd/api`)**: The single Go binary starts both an Echo HTTP server (8080) and a TCP server (9090) in parallel.
2. **Business Logic Layer (`/internal/api/handlers`, `/internal/tcp`)**: Core logic for Google OAuth authentication, session management, and the `IW` protocol for smartwatch data.
3. **Data Access Layer (`/internal/db`, `/internal/device`, `/internal/tcp/repository.go`)**: Abstracts interactions with PostgreSQL for device registration, user data, and the `device_packets` table.
4. **Shared Libraries (`/internal/auth`, `/internal/tcp/protocol`)**: Utilities for Google UserInfo, HMAC signing, and protocol parsing.

### Data Ingestion (TCP Ingestion)
1. **Connection Hook**: `net.Listen("tcp", addr)` accepts persistent connections.
2. **HandleConnection**: Each connection is handled in its own goroutine (state-loop).
3. **ScanFrame**: A custom `bufio.SplitFunc` that handles noise (newlines), delimiters (`[` or `]`), and '#' terminators for robust frame accumulation.
4. **Auth State Machine**: First packet must be `AP00` (Login) with valid IMEI. Connection is rejected if IMEI is not registered to a user.
5. **Persistence**: Valid data packets (GPS, Heart rate, etc.) are asynchronously or synchronously persisted to the `device_packets` table.

### User Flow (HTTP/API Dashboard)
1. **User Sign-In**: Authenticates via Google OAuth 2.0.
2. **Session**: An HMAC-signed cookie is issued, containing user ID and email.
3. **Dashboard**: The user accesses `dashboard.html` which fetches the device list and provides a registration form with real-time IMEI validation.
4. **Device Registration**: The user enters a 15-digit IMEI. The backend registers the device to the user's account in the `devices` table.

### Diagram: High-Level Data Flow
```mermaid
graph LR
    subgraph Clients
        SW[Smartwatch] -- TCP 9090 --> TCPS[TCP Server]
        Browser[User Browser] -- HTTP 8080 --> HTTPS[HTTP Server]
    end

    subgraph "HDP Backend (Go/Echo/TCP)"
        TCPS -- Auth/IMEI --> DB[(PostgreSQL)]
        HTTPS -- Auth/OAuth --> DB
        TCPS -- Ingest Packets --> DB
        HTTPS -- Manage Devices --> DB
    end
```
