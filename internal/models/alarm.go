package models

import "time"

// NormalizedAlarm is the unified alarm/fault representation.
type NormalizedAlarm struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	DeviceID string `json:"deviceId"`
	PlantID  string `json:"plantId,omitempty"`

	// Alarm details
	Code     string       `json:"code"`
	Name     string       `json:"name"`
	Message  string       `json:"message,omitempty"`
	Severity AlarmSeverity `json:"severity"`
	Status   AlarmStatus   `json:"status"`

	// Device context
	DeviceSerialNumber string     `json:"deviceSerialNumber,omitempty"`
	DeviceType         DeviceType `json:"deviceType,omitempty"`
	PlantName          string     `json:"plantName,omitempty"`

	// Timestamps
	StartTime  time.Time  `json:"startTime"`
	EndTime    *time.Time `json:"endTime,omitempty"`
	UpdateTime *time.Time `json:"updateTime,omitempty"`
	Duration   string     `json:"duration,omitempty"`

	Meta ProviderMeta `json:"meta"`
}

// AlarmSeverity maps the various vendor severity levels to a unified scale.
type AlarmSeverity string

const (
	AlarmSeverityInfo     AlarmSeverity = "info"
	AlarmSeverityWarning  AlarmSeverity = "warning"
	AlarmSeverityCritical AlarmSeverity = "critical"
	AlarmSeverityUnknown  AlarmSeverity = "unknown"
)

// AlarmStatus tracks the lifecycle of an alarm.
type AlarmStatus string

const (
	AlarmStatusActive    AlarmStatus = "active"
	AlarmStatusResolved  AlarmStatus = "resolved"
	AlarmStatusAcknowledged AlarmStatus = "acknowledged"
	AlarmStatusUnknown   AlarmStatus = "unknown"
)
