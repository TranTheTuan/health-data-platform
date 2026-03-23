# Project Roadmap
## Health Data Platform

### Phase 1: Foundation (Completed)
- [x] Initialize repository using standard Go project layout.
- [x] Document project directory structure, standards, and architecture.
- [x] Finalize initial database schema (Users, Devices, Packets) in `/internal/migrations`.
- [x] Integrated Google OAuth 2.0 authentication flow and secure session management.

### Phase 2: Core Platform & Ingestion (In-Progress)
- [x] **Smartwatch Data Ingestion**: Robust TCP server (port 9090) with IW protocol scanner.
- [x] **IMEI Authentication**: Support for device login (AP00) and ownership check.
- [x] **Web Portal**: Responsive dashboard (8080) for device management and registration.
- [x] **Inline Validation**: High-quality user experience with real-time field validation for registration.
- [x] **Demo TCP Packet Generator**: Built-in demo feature for testing and live demonstrations without physical smartwatches.
- [ ] **Heartrate & Blood Pressure Parsing**: Expand the TCP handler to parse and store specific health metrics from APHP/APHT packets.
- [ ] Create detailed data models and database repositories in `/internal/domain`.

### Phase 3: Advanced Features & Integrations (Pending)
- [ ] **GPS/LBS Location Normalization**: Implement a service to convert raw hex/text location data into GeoJSON format.
- [ ] **Audit Logging & Security**: Implement detailed audit logging for PHI-related actions.
- [ ] **Data Anonymization**: Develop anonymization logic in `/pkg/security`.
- [ ] Prepare analytics dashboards (using web clients connecting to the platform).

### Phase 4: Production Deployment (Upcoming)
- [ ] Finalize end-to-end testing and performance benchmarking.
- [ ] Deploy initial release to production cluster using Helm in `/deployments/k8s`.
- [ ] Monitor platform health and optimize database queries based on metrics.
- [ ] Prepare user and API documentation for external consumption.
