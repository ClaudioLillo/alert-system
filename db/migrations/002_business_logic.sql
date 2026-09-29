CREATE OR REPLACE FUNCTION create_device(p_name TEXT)
RETURNS devices
LANGUAGE plpgsql
AS $$
DECLARE
    created_device devices;
BEGIN
    IF p_name IS NULL OR btrim(p_name) = '' THEN
        RAISE EXCEPTION 'Device name is required';
    END IF;

    INSERT INTO devices (name, is_active, last_seen_at)
    VALUES (btrim(p_name), FALSE, NULL)
    RETURNING * INTO created_device;

    RETURN created_device;
END;
$$;

CREATE OR REPLACE FUNCTION register_device_heartbeat(p_device_id UUID)
RETURNS devices
LANGUAGE plpgsql
AS $$
DECLARE
    updated_device devices;
BEGIN
    UPDATE devices
    SET is_active = TRUE,
        last_seen_at = NOW(),
        updated_at = NOW()
    WHERE id = p_device_id
    RETURNING * INTO updated_device;

    IF updated_device.id IS NULL THEN
        RAISE EXCEPTION 'Device % was not found', p_device_id;
    END IF;

    RETURN updated_device;
END;
$$;

CREATE OR REPLACE FUNCTION create_alarm(p_device_id UUID, p_reason TEXT)
RETURNS alarms
LANGUAGE plpgsql
AS $$
DECLARE
    new_alarm alarms;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM devices WHERE id = p_device_id) THEN
        RAISE EXCEPTION 'Device % was not found', p_device_id;
    END IF;

    INSERT INTO alarms (device_id, status, reason)
    VALUES (p_device_id, 'ACTIVE', p_reason)
    RETURNING * INTO new_alarm;

    RETURN new_alarm;
END;
$$;

CREATE OR REPLACE FUNCTION acknowledge_alarm(p_alarm_id UUID)
RETURNS alarms
LANGUAGE plpgsql
AS $$
DECLARE
    updated_alarm alarms;
BEGIN
    UPDATE alarms
    SET status = 'ACKNOWLEDGED',
        acknowledged_at = NOW()
    WHERE id = p_alarm_id AND status = 'ACTIVE'
    RETURNING * INTO updated_alarm;

    IF updated_alarm.id IS NULL THEN
        RAISE EXCEPTION 'Alarm % is not active or was not found', p_alarm_id;
    END IF;

    RETURN updated_alarm;
END;
$$;

CREATE OR REPLACE FUNCTION resolve_alarm(p_alarm_id UUID)
RETURNS alarms
LANGUAGE plpgsql
AS $$
DECLARE
    updated_alarm alarms;
BEGIN
    UPDATE alarms
    SET status = 'RESOLVED',
        resolved_at = NOW()
    WHERE id = p_alarm_id AND status IN ('ACTIVE', 'ACKNOWLEDGED')
    RETURNING * INTO updated_alarm;

    IF updated_alarm.id IS NULL THEN
        RAISE EXCEPTION 'Alarm % is not active/acknowledged or was not found', p_alarm_id;
    END IF;

    RETURN updated_alarm;
END;
$$;

CREATE OR REPLACE FUNCTION mark_inactive_devices()
RETURNS INTEGER
LANGUAGE plpgsql
AS $$
DECLARE
    updated_count INTEGER;
BEGIN
    UPDATE devices
    SET is_active = FALSE,
        updated_at = NOW()
    WHERE last_seen_at IS NULL OR last_seen_at < NOW() - INTERVAL '15 seconds';

    GET DIAGNOSTICS updated_count = ROW_COUNT;
    RETURN updated_count;
END;
$$;

CREATE OR REPLACE FUNCTION ensure_single_active_alarm()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM alarms
        WHERE device_id = NEW.device_id
          AND status = 'ACTIVE'
    ) THEN
        RAISE EXCEPTION 'Device % already has an active alarm', NEW.device_id;
    END IF;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_ensure_single_active_alarm ON alarms;
CREATE TRIGGER trg_ensure_single_active_alarm
BEFORE INSERT OR UPDATE OF device_id, status
ON alarms
FOR EACH ROW
WHEN (NEW.status = 'ACTIVE')
EXECUTE FUNCTION ensure_single_active_alarm();
