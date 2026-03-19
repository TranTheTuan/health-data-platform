# Design Guidelines
## Health Data Platform

### Architectural Philosophy
1. **API First**: The platform will provide primarily backend APIs. Any web frontend should operate purely as a consumer of these APIs.
2. **Stateless Operations**: Applications deployed through the `/cmd` entry points should maintain stateless operations whenever possible to allow horizontal scaling.
3. **Data Security**: Data handling must strictly validate incoming payloads, avoid injecting user data indiscriminately, and ensure encryption paths are unbroken.

### Database Design
- Tables should maintain tracking timestamps (`created_at`, `updated_at`, `deleted_at` if soft deletion is supported).
- Use UUIDs as primary keys to reduce predictability of identifiers.
- Ensure sensitive Patient Health Information (PHI) is encrypted at rest or anonymized when processed for generic analytics.

### Error Handling & Logging
- **Structured Logging**: Use structured logging (e.g., JSON) with zap or logrus. Ensure log lines contain contextual metadata (request IDs, timestamps, urgency).
- **Error Types**: Standardize error returns. DO NOT leak stack traces to the end users over the API. Instead, map internal errors to user-friendly HTTP statuses.
- **Observability**: Maintain prometheus metrics for API endpoints (latency, 4xx/5xx counts).
