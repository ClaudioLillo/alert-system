package model

import "time"

// Device represents a connected edge device.
type Device struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	IsActive  bool       `json:"is_active"`
	LastSeenAt *time.Time `json:"last_seen_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Alarm represents an active, acknowledged, or resolved alarm.
type Alarm struct {
	ID            string     `json:"id"`
	DeviceID      string     `json:"device_id"`
	Status        string     `json:"status"`
	Reason        string     `json:"reason"`
	CreatedAt     time.Time  `json:"created_at"`
	AcknowledgedAt *time.Time `json:"acknowledged_at"`
	ResolvedAt    *time.Time `json:"resolved_at"`
}

// HeartbeatRequest carries the device heartbeat payload.
type HeartbeatRequest struct {
	DeviceID string `json:"device_id,omitempty"`
}

// CreateDeviceRequest contains the payload used to register a new device.
type CreateDeviceRequest struct {
	Name string `json:"name"`
}

// DeviceStatusSyncResult reports the latest device status synchronization.
type DeviceStatusSyncResult struct {
	LastSyncAt          time.Time `json:"last_sync_at"`
	DevicesDeactivated int       `json:"devices_deactivated"`
}

// CreateAlarmRequest contains the payload used to raise a new alarm.
type CreateAlarmRequest struct {
	DeviceID string `json:"device_id"`
	Reason   string `json:"reason"`
}
