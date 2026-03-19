# Phase 1: Database Schema

**Context Links**
- [Main Plan](./plan.md)

## Overview
- Priority: P1
- Effort: 1.5h
- Status: Pending
- Set up PostgreSQL tables for device management, user-device linking, and health/location packet storage.

## Key Insights
- **IMEI is the device identifier** — no custom token. Devices send their IMEI in the `AP00` login packet.
- A `devices` table stores IMEI and the `user_id` it belongs to. Admin/app links IMEI to user via HTTP API.
- **Single `device_packets` table** with `packet_type` + `raw_payload TEXT` + `parsed_data JSONB` handles all 13 command types without per-type tables. Avoids premature schema over-engineering (YAGNI). Specific queries can filter by `packet_type`.
- `pgx/v5/stdlib` driver with `database/sql` for connection pooling and testability.

## Requirements

### `devices` table
| Column | Type | Notes |
|--------|------|-------|
| `id` | UUID | Primary key |
| `imei` | VARCHAR(15) | Unique, device identifier from AP00 |
| `user_id` | TEXT | FK reference to the Google OAuth user ID |
| `name` | TEXT | Optional friendly name (set by user via HTTP API) |
| `last_seen_at` | TIMESTAMPTZ | Updated on each AP00 login |
| `created_at` | TIMESTAMPTZ | Default NOW() |

### `device_packets` table
| Column | Type | Notes |
|--------|------|-------|
| `id` | UUID | Primary key |
| `device_id` | UUID | FK → `devices.id` |
| `user_id` | TEXT | Denormalized for fast queries, no join needed |
| `packet_type` | VARCHAR(10) | e.g. `AP49`, `APHT`, `AP01` |
| `raw_payload` | TEXT | Full raw packet (after stripping `IW` and `#`) |
| `parsed_data` | JSONB | Extracted fields (nullable, populated by parser) |
| `recorded_at` | TIMESTAMPTZ | Default NOW() (server receive time) |

## Architecture
```sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE devices (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    imei        VARCHAR(15) NOT NULL UNIQUE,
    user_id     TEXT NOT NULL,
    name        TEXT,
    last_seen_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE device_packets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id   UUID NOT NULL REFERENCES devices(id),
    user_id     TEXT NOT NULL,
    packet_type VARCHAR(10) NOT NULL,
    raw_payload TEXT NOT NULL,
    parsed_data JSONB,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Query pattern: recent health data per device
CREATE INDEX idx_device_packets_device_type_time
    ON device_packets(device_id, packet_type, recorded_at DESC);

-- Query pattern: all health data for a user
CREATE INDEX idx_device_packets_user_time
    ON device_packets(user_id, recorded_at DESC);
```

## Related Code Files
- `[NEW]` `deployments/db/migrations/001_initial_schema.sql`
- `[MODIFY]` `configs/config.go` — add `DatabaseURL`, `TCPAddr` fields
- `[MODIFY]` `.env` — add `DATABASE_URL`, `TCP_ADDR`
- `[NEW]` `internal/db/connection.go` — open pgx pool via `database/sql`

## Implementation Steps
1. Create `deployments/db/migrations/001_initial_schema.sql` with the SQL above.
2. Add to `.env`:
   ```
   DATABASE_URL=postgres://admin:admin@localhost:5432/hdp?sslmode=disable
   TCP_ADDR=:9090
   ```
3. Extend `configs.Config`:
   ```go
   DatabaseURL string
   TCPAddr     string
   ```
4. Create `internal/db/connection.go`:
   ```go
   func NewPool(url string) (*sql.DB, error) {
       db, err := sql.Open("pgx", url)
       db.SetMaxOpenConns(25)
       db.SetMaxIdleConns(5)
       return db, db.Ping()
   }
   ```

## Todo List
- [ ] Create `deployments/db/migrations/001_initial_schema.sql`
- [ ] Run migration against local DB
- [ ] Update `configs/config.go` with `DatabaseURL`, `TCPAddr`
- [ ] Update `.env` with new vars
- [ ] Create `internal/db/connection.go`

## Success Criteria
- `psql $DATABASE_URL -f deployments/db/migrations/001_initial_schema.sql` exits 0
- Both tables and indexes visible in `\dt` and `\di`
- `internal/db/connection.go` compiles cleanly

## Security Considerations
- IMEI is device identity but NOT a secret (it's hardware-printed). Don't treat it as a password.
- `user_id` denormalized in packets — ensure it's set from the server-side device lookup, not from the device payload.

## Next Steps
- [Phase 2: Protocol Parser](./phase-02-protocol-parser.md)
