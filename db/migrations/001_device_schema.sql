CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'alarm_status') THEN
        CREATE TYPE alarm_status AS ENUM ('ACTIVE', 'ACKNOWLEDGED', 'RESOLVED');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    last_seen_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS alarms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    status alarm_status NOT NULL DEFAULT 'ACTIVE',
    reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    acknowledged_at TIMESTAMPTZ NULL,
    resolved_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_devices_is_active ON devices (is_active);
CREATE INDEX IF NOT EXISTS idx_devices_last_seen_at ON devices (last_seen_at);
CREATE INDEX IF NOT EXISTS idx_alarms_device_id ON alarms (device_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_active_alarm_per_device
    ON alarms (device_id)
    WHERE status = 'ACTIVE';
