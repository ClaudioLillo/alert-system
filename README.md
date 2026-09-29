# Alert System

A thin Go backend that connects to PostgreSQL running on Amazon RDS. The API does not own the business logic; instead, it delegates every mutation and validation to PostgreSQL functions and triggers.

## Architecture

- API layer: Go HTTP server
- Persistence layer: PostgreSQL on Amazon RDS
- Business logic: database functions and triggers
- Documentation: Swagger/OpenAPI

## Database model

### devices

Stores connected devices and liveness information.

Fields:

- id: UUID
- name: device name
- is_active: whether the device is currently alive
- last_seen_at: timestamp of the last heartbeat
- created_at: creation timestamp
- updated_at: last update timestamp

### alarms

Stores alarm records raised by devices and their lifecycle state.

Fields:

- id: UUID
- device_id: related device
- status: ACTIVE, ACKNOWLEDGED, RESOLVED
- reason: business explanation
- created_at: created timestamp
- acknowledged_at: ack timestamp
- resolved_at: resolution timestamp

## Database-owned business rules

The following logic lives in PostgreSQL instead of the Go service:

- register_device_heartbeat(device_id): updates the heartbeat and marks the device active
- create_alarm(device_id, reason): verifies device existence and raises a new alarm
- acknowledge_alarm(alarm_id): transitions an alarm to ACKNOWLEDGED
- resolve_alarm(alarm_id): transitions an alarm to RESOLVED
- synchronize_stale_devices(): marks devices inactive when their last heartbeat is older than five minutes
- device status sync trigger: runs stale-device synchronization when the sync marker is updated
- ensure_single_active_alarm trigger: prevents multiple active alarms for the same device

This keeps the API as a pass-through layer while the database enforces the core business rules.

## Project structure

- cmd/server/main.go: HTTP server entry point
- cmd/migrate/main.go: migration runner
- internal/api: HTTP handlers and router
- internal/config: runtime configuration from environment
- internal/db: PostgreSQL connection and migration execution
- internal/model: DTOs and request/response models
- db/migrations: SQL migration files
- docs/swagger: generated Swagger files

## Local setup

1. Configure PostgreSQL access:

   ```bash
   export DB_URL="postgres://postgres:your_password@your-rds-endpoint:5432/alert_system?sslmode=require"
   export HTTP_ADDR="0.0.0.0:8080"
   export MIGRATION_DIR="./db/migrations"
   ```

2. Run the migrations:

   ```bash
   go run ./cmd/migrate
   ```

3. Start the API:

   ```bash
   go run ./cmd/server
   ```

4. Open Swagger:

   ```text
   http://localhost:8080/swagger/index.html
   ```

## API endpoints

- GET /healthz
- GET /api/v1/devices
- GET /api/v1/devices/{id}
- POST /api/v1/devices/{id}/heartbeat
- POST /api/v1/devices/synchronize
- GET /api/v1/alarms
- POST /api/v1/alarms
- POST /api/v1/alarms/{id}/acknowledge
- POST /api/v1/alarms/{id}/resolve

## Example usage

Device heartbeat:

```bash
curl -X POST http://localhost:8080/api/v1/devices/<device-id>/heartbeat
```

Synchronize device statuses:

```bash
curl -X POST http://localhost:8080/api/v1/devices/synchronize
```

Create alarm:

```bash
curl -X POST http://localhost:8080/api/v1/alarms \
  -H "Content-Type: application/json" \
  -d '{"device_id":"<device-id>","reason":"Door opened while armed"}'
```

Acknowledge alarm:

```bash
curl -X POST http://localhost:8080/api/v1/alarms/<alarm-id>/acknowledge
```

Resolve alarm:

```bash
curl -X POST http://localhost:8080/api/v1/alarms/<alarm-id>/resolve
```

## Notes

- The Go service is intentionally thin and acts as a transport layer.
- PostgreSQL enforces validation, lifecycle transitions, and stale-device tracking.
- For production, protect the database connection with AWS IAM or a secure secret manager and add TLS configuration as required.
