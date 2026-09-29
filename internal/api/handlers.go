package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"alert-system/internal/model"
)

// Handler owns the database dependency used by HTTP endpoints.
type Handler struct {
	db *sql.DB
}

// Health returns the service health status.
// @Summary Health check
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /healthz [get]
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ListDevices lists every device known to the system.
// @Summary List devices
// @Tags devices
// @Produce json
// @Success 200 {array} model.Device
// @Router /api/v1/devices [get]
func (h *Handler) ListDevices(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), `
		SELECT id, name, is_active, last_seen_at, created_at, updated_at
		FROM devices
		ORDER BY created_at DESC
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch devices")
		return
	}
	defer rows.Close()

	devices := make([]model.Device, 0)
	for rows.Next() {
		var device model.Device
		if err := rows.Scan(&device.ID, &device.Name, &device.IsActive, &device.LastSeenAt, &device.CreatedAt, &device.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to scan device row")
			return
		}
		devices = append(devices, device)
	}

	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "device query error")
		return
	}

	writeJSON(w, http.StatusOK, devices)
}

// GetDevice returns a single device by ID.
// @Summary Get device by ID
// @Tags devices
// @Produce json
// @Param id path string true "Device ID"
// @Success 200 {object} model.Device
// @Failure 404 {object} map[string]string
// @Router /api/v1/devices/{id} [get]
func (h *Handler) GetDevice(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "device ID is required")
		return
	}

	var device model.Device
	err := h.db.QueryRowContext(r.Context(), `
		SELECT id, name, is_active, last_seen_at, created_at, updated_at
		FROM devices
		WHERE id = $1
	`, id).Scan(&device.ID, &device.Name, &device.IsActive, &device.LastSeenAt, &device.CreatedAt, &device.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "device not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load device")
		return
	}

	writeJSON(w, http.StatusOK, device)
}

// CreateDevice registers a new device in the database-owned business logic.
// @Summary Create device
// @Tags devices
// @Accept json
// @Produce json
// @Param request body model.CreateDeviceRequest true "Device creation payload"
// @Success 201 {object} model.Device
// @Failure 400 {object} map[string]string
// @Router /api/v1/devices [post]
func (h *Handler) CreateDevice(w http.ResponseWriter, r *http.Request) {
	var request model.CreateDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	if request.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	var device model.Device
	err := h.db.QueryRowContext(r.Context(), `
		SELECT id, name, is_active, last_seen_at, created_at, updated_at
		FROM create_device($1)
	`, request.Name).Scan(
		&device.ID, &device.Name, &device.IsActive, &device.LastSeenAt, &device.CreatedAt, &device.UpdatedAt,
	)
	if err != nil {
		log.Printf("create device failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create device")
		return
	}

	writeJSON(w, http.StatusCreated, device)
}

// DeviceHeartbeat records a device heartbeat and updates its active status in the database function.
// @Summary Send a device heartbeat
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Success 200 {object} model.Device
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/devices/{id}/heartbeat [post]
func (h *Handler) DeviceHeartbeat(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "device ID is required")
		return
	}

	var device model.Device
	err := h.db.QueryRowContext(r.Context(), `
		SELECT id, name, is_active, last_seen_at, created_at, updated_at
		FROM register_device_heartbeat($1)
	`, id).Scan(
		&device.ID, &device.Name, &device.IsActive, &device.LastSeenAt, &device.CreatedAt, &device.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "device not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "heartbeat registration failed")
		return
	}

	writeJSON(w, http.StatusOK, device)
}

// SynchronizeDeviceStatuses updates the sync marker, which triggers database-side stale-device processing.
// @Summary Synchronize device active statuses
// @Tags devices
// @Produce json
// @Success 200 {object} model.DeviceStatusSyncResult
// @Failure 500 {object} map[string]string
// @Router /api/v1/devices/synchronize [post]
func (h *Handler) SynchronizeDeviceStatuses(w http.ResponseWriter, r *http.Request) {
	if _, err := h.db.ExecContext(r.Context(), `
		UPDATE device_status_sync
		SET last_sync_at = clock_timestamp()
		WHERE id = 1
	`); err != nil {
		log.Printf("device status synchronization failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to synchronize device statuses")
		return
	}

	var result model.DeviceStatusSyncResult
	if err := h.db.QueryRowContext(r.Context(), `
		SELECT last_sync_at, devices_deactivated
		FROM device_status_sync
		WHERE id = 1
	`).Scan(&result.LastSyncAt, &result.DevicesDeactivated); err != nil {
		log.Printf("load device status synchronization result failed: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load synchronization result")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ListAlarms returns all alarm records, including historical entries.
// @Summary List alarms
// @Tags alarms
// @Produce json
// @Success 200 {array} model.Alarm
// @Router /api/v1/alarms [get]
func (h *Handler) ListAlarms(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), `
		SELECT id, device_id, status, reason, created_at, acknowledged_at, resolved_at
		FROM alarms
		ORDER BY created_at DESC
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch alarms")
		return
	}
	defer rows.Close()

	alarms := make([]model.Alarm, 0)
	for rows.Next() {
		var alarm model.Alarm
		if err := rows.Scan(&alarm.ID, &alarm.DeviceID, &alarm.Status, &alarm.Reason, &alarm.CreatedAt, &alarm.AcknowledgedAt, &alarm.ResolvedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to scan alarm row")
			return
		}
		alarms = append(alarms, alarm)
	}

	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "alarm query error")
		return
	}

	writeJSON(w, http.StatusOK, alarms)
}

// CreateAlarm raises a new alarm through a database function with validation in PostgreSQL.
// @Summary Create alarm
// @Tags alarms
// @Accept json
// @Produce json
// @Param request body model.CreateAlarmRequest true "Alarm creation payload"
// @Success 201 {object} model.Alarm
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/alarms [post]
func (h *Handler) CreateAlarm(w http.ResponseWriter, r *http.Request) {
	var request model.CreateAlarmRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	if request.DeviceID == "" || request.Reason == "" {
		writeError(w, http.StatusBadRequest, "device_id and reason are required")
		return
	}

	var alarm model.Alarm
	err := h.db.QueryRowContext(r.Context(), `
		SELECT id, device_id, status, reason, created_at, acknowledged_at, resolved_at
		FROM create_alarm($1, $2)
	`, request.DeviceID, request.Reason).Scan(
		&alarm.ID, &alarm.DeviceID, &alarm.Status, &alarm.Reason, &alarm.CreatedAt, &alarm.AcknowledgedAt, &alarm.ResolvedAt,
	)
	if err != nil {
		if err.Error() == "pq: Device "+request.DeviceID+" was not found" || err.Error() == "Device "+request.DeviceID+" was not found" {
			writeError(w, http.StatusNotFound, "device not found")
			return
		}
		if err.Error() == "Device "+request.DeviceID+" already has an active alarm" || err.Error() == "pq: Device "+request.DeviceID+" already has an active alarm" {
			writeError(w, http.StatusConflict, "device already has an active alarm")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create alarm")
		return
	}

	writeJSON(w, http.StatusCreated, alarm)
}

// AcknowledgeAlarm marks an existing alarm as acknowledged using a stored function.
// @Summary Acknowledge alarm
// @Tags alarms
// @Produce json
// @Param id path string true "Alarm ID"
// @Success 200 {object} model.Alarm
// @Failure 404 {object} map[string]string
// @Router /api/v1/alarms/{id}/acknowledge [post]
func (h *Handler) AcknowledgeAlarm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "alarm ID is required")
		return
	}

	var alarm model.Alarm
	err := h.db.QueryRowContext(r.Context(), `
		SELECT id, device_id, status, reason, created_at, acknowledged_at, resolved_at
		FROM acknowledge_alarm($1)
	`, id).Scan(
		&alarm.ID, &alarm.DeviceID, &alarm.Status, &alarm.Reason, &alarm.CreatedAt, &alarm.AcknowledgedAt, &alarm.ResolvedAt,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to acknowledge alarm")
		return
	}

	writeJSON(w, http.StatusOK, alarm)
}

// ResolveAlarm marks an alarm as resolved through a database function.
// @Summary Resolve alarm
// @Tags alarms
// @Produce json
// @Param id path string true "Alarm ID"
// @Success 200 {object} model.Alarm
// @Failure 404 {object} map[string]string
// @Router /api/v1/alarms/{id}/resolve [post]
func (h *Handler) ResolveAlarm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "alarm ID is required")
		return
	}

	var alarm model.Alarm
	err := h.db.QueryRowContext(r.Context(), `
		SELECT id, device_id, status, reason, created_at, acknowledged_at, resolved_at
		FROM resolve_alarm($1)
	`, id).Scan(
		&alarm.ID, &alarm.DeviceID, &alarm.Status, &alarm.Reason, &alarm.CreatedAt, &alarm.AcknowledgedAt, &alarm.ResolvedAt,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to resolve alarm")
		return
	}

	writeJSON(w, http.StatusOK, alarm)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, "unable to encode response", http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func timeUTC(t time.Time) time.Time {
	return t.UTC()
}
