CREATE TABLE IF NOT EXISTS device_status_sync (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    last_sync_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    devices_deactivated INTEGER NOT NULL DEFAULT 0 CHECK (devices_deactivated >= 0)
);

INSERT INTO device_status_sync (id)
VALUES (1)
ON CONFLICT (id) DO NOTHING;

CREATE OR REPLACE FUNCTION synchronize_stale_devices()
RETURNS INTEGER
LANGUAGE plpgsql
AS $$
DECLARE
    updated_count INTEGER;
BEGIN
    UPDATE devices
    SET is_active = FALSE,
        updated_at = clock_timestamp()
    WHERE is_active = TRUE
      AND (
          last_seen_at IS NULL
          OR last_seen_at < clock_timestamp() - INTERVAL '5 minutes'
      );

    GET DIAGNOSTICS updated_count = ROW_COUNT;
    RETURN updated_count;
END;
$$;

CREATE OR REPLACE FUNCTION process_device_status_sync_update()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    deactivated_count INTEGER;
BEGIN
    deactivated_count := synchronize_stale_devices();

    UPDATE device_status_sync
    SET devices_deactivated = deactivated_count
    WHERE id = NEW.id;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_process_device_status_sync_update ON device_status_sync;
CREATE TRIGGER trg_process_device_status_sync_update
AFTER UPDATE OF last_sync_at
ON device_status_sync
FOR EACH ROW
WHEN (OLD.last_sync_at IS DISTINCT FROM NEW.last_sync_at)
EXECUTE FUNCTION process_device_status_sync_update();