package models

import "time"

// NormalizedRealtime represents the unified real-time telemetry snapshot
// from any inverter brand. This is the most critical model — it must
// handle grid-tied, hybrid, off-grid, single-phase, and three-phase systems.
type NormalizedRealtime struct {
	DeviceID  string    `json:"deviceId"`
	Provider  string    `json:"provider"`
	Timestamp time.Time `json:"timestamp"`

	// Original timestamp info for traceability
	OriginalTimestamp string `json:"originalTimestamp,omitempty"`
	OriginalTimezone  string `json:"originalTimezone,omitempty"`

	// Device status
	Status        DeviceStatus  `json:"status"`
	OperatingMode OperatingMode `json:"operatingMode"`

	// ── PV (Solar) ──
	PV *PVData `json:"pv,omitempty"`

	// ── Battery ──
	Battery *BatteryData `json:"battery,omitempty"`

	// ── Grid ──
	Grid *GridData `json:"grid,omitempty"`

	// ── Load / Consumption ──
	Load *LoadData `json:"load,omitempty"`

	// ── Backup Output ──
	Backup *BackupData `json:"backup,omitempty"`

	// ── Meter ──
	Meters []MeterData `json:"meters,omitempty"`

	// ── Environmental ──
	Environment *EnvironmentData `json:"environment,omitempty"`

	// ── Provider metadata ──
	Meta ProviderMeta `json:"meta"`
}

// OperatingMode is the unified operating mode across all brands.
type OperatingMode string

const (
	OperatingModeInitializing  OperatingMode = "initializing"
	OperatingModeWaiting       OperatingMode = "waiting"
	OperatingModeGridConnected OperatingMode = "grid_connected"
	OperatingModeOffGrid       OperatingMode = "off_grid"
	OperatingModeFault         OperatingMode = "fault"
	OperatingModeUpgrading     OperatingMode = "upgrading"
	OperatingModeStandby       OperatingMode = "standby"
	OperatingModeShutdown      OperatingMode = "shutdown"
	OperatingModeUnknown       OperatingMode = "unknown"
)

// ── PV Data ──

type PVData struct {
	TotalPowerW     float64    `json:"totalPowerW"`
	TodayEnergyKWh  *float64   `json:"todayEnergyKWh,omitempty"`
	TotalEnergyKWh  *float64   `json:"totalEnergyKWh,omitempty"`
	MonthEnergyKWh  *float64   `json:"monthEnergyKWh,omitempty"`
	YearEnergyKWh   *float64   `json:"yearEnergyKWh,omitempty"`
	Strings         []PVString `json:"strings,omitempty"`
}

type PVString struct {
	ID       int      `json:"id"`
	VoltageV *float64 `json:"voltageV,omitempty"`
	CurrentA *float64 `json:"currentA,omitempty"`
	PowerW   *float64 `json:"powerW,omitempty"`
}

// ── Battery Data ──

// EnergyDirection is a unified direction enum.
type EnergyDirection string

const (
	DirectionCharging    EnergyDirection = "charging"
	DirectionDischarging EnergyDirection = "discharging"
	DirectionIdle        EnergyDirection = "idle"
	DirectionUnknown     EnergyDirection = "unknown"
)

type BatteryData struct {
	SOCPercent          *float64        `json:"socPercent,omitempty"`
	PowerW              float64         `json:"powerW"`
	Direction           EnergyDirection `json:"direction"`
	TemperatureC        *float64        `json:"temperatureC,omitempty"`
	TodayChargeKWh     *float64        `json:"todayChargeKWh,omitempty"`
	TodayDischargeKWh  *float64        `json:"todayDischargeKWh,omitempty"`
	TotalChargeKWh     *float64        `json:"totalChargeKWh,omitempty"`
	TotalDischargeKWh  *float64        `json:"totalDischargeKWh,omitempty"`
	MaxChargePowerW    *float64        `json:"maxChargePowerW,omitempty"`
	MaxDischargePowerW *float64        `json:"maxDischargePowerW,omitempty"`
	VoltageDC          *float64        `json:"voltageDC,omitempty"`
	CurrentDC          *float64        `json:"currentDC,omitempty"`
	Groups             []BatteryGroup  `json:"groups,omitempty"`
}

type BatteryGroup struct {
	ID         int      `json:"id"`
	SOCPercent *float64 `json:"socPercent,omitempty"`
	PowerW     *float64 `json:"powerW,omitempty"`
	VoltageV   *float64 `json:"voltageV,omitempty"`
	CurrentA   *float64 `json:"currentA,omitempty"`
	TempC      *float64 `json:"temperatureC,omitempty"`
}

// ── Grid Data ──

// GridDirection represents the direction of energy flow at the grid interconnection.
type GridDirection string

const (
	GridDirectionImporting GridDirection = "importing" // Buying from grid
	GridDirectionExporting GridDirection = "exporting" // Selling to grid
	GridDirectionIdle      GridDirection = "idle"
	GridDirectionUnknown   GridDirection = "unknown"
)

type GridData struct {
	TotalPowerW     float64       `json:"totalPowerW"`
	Direction       GridDirection  `json:"direction"`
	FrequencyHz     *float64      `json:"frequencyHz,omitempty"`
	PowerFactor     *float64      `json:"powerFactor,omitempty"`
	TodayImportKWh  *float64      `json:"todayImportKWh,omitempty"`
	TodayExportKWh  *float64      `json:"todayExportKWh,omitempty"`
	TotalImportKWh  *float64      `json:"totalImportKWh,omitempty"`
	TotalExportKWh  *float64      `json:"totalExportKWh,omitempty"`
	MonthImportKWh  *float64      `json:"monthImportKWh,omitempty"`
	MonthExportKWh  *float64      `json:"monthExportKWh,omitempty"`
	YearImportKWh   *float64      `json:"yearImportKWh,omitempty"`
	YearExportKWh   *float64      `json:"yearExportKWh,omitempty"`
	Phases          []PhaseData   `json:"phases,omitempty"`
}

type PhaseData struct {
	Phase       string   `json:"phase"` // "A", "B", "C" (or "R", "S", "T" → mapped to A/B/C)
	VoltageV    *float64 `json:"voltageV,omitempty"`
	CurrentA    *float64 `json:"currentA,omitempty"`
	PowerW      *float64 `json:"powerW,omitempty"`
	FrequencyHz *float64 `json:"frequencyHz,omitempty"`
	PowerFactor *float64 `json:"powerFactor,omitempty"`
	// Apparent power
	ApparentPowerVA *float64 `json:"apparentPowerVA,omitempty"`
	// Reactive power
	ReactivePowerVAR *float64 `json:"reactivePowerVAR,omitempty"`
}

// ── Load Data ──

type LoadData struct {
	TotalPowerW    float64  `json:"totalPowerW"`
	TodayEnergyKWh *float64 `json:"todayEnergyKWh,omitempty"`
	TotalEnergyKWh *float64 `json:"totalEnergyKWh,omitempty"`
	// Self-consumption metrics
	SelfConsumptionRate  *float64 `json:"selfConsumptionRate,omitempty"`
	SelfSufficiencyRate  *float64 `json:"selfSufficiencyRate,omitempty"`
}

// ── Backup Output Data ──

type BackupData struct {
	TotalPowerW float64     `json:"totalPowerW"`
	Phases      []PhaseData `json:"phases,omitempty"`
}

// ── Meter Data ──

type MeterData struct {
	ID         string    `json:"id"`
	MeterType  MeterType `json:"meterType"`
	TotalPowerW float64  `json:"totalPowerW"`
	TotalImportKWh *float64 `json:"totalImportKWh,omitempty"`
	TotalExportKWh *float64 `json:"totalExportKWh,omitempty"`
	Phases     []PhaseData  `json:"phases,omitempty"`
}

type MeterType string

const (
	MeterTypeGrid         MeterType = "grid"
	MeterTypePV           MeterType = "pv"
	MeterTypeExportLimiter MeterType = "export_limiter"
	MeterTypeStorage      MeterType = "storage"
	MeterTypeConsumption  MeterType = "consumption"
	MeterTypeUnknown      MeterType = "unknown"
)

// ── Environment Data ──

type EnvironmentData struct {
	InverterTemperatureC  *float64 `json:"inverterTemperatureC,omitempty"`
	AmbientTemperatureC   *float64 `json:"ambientTemperatureC,omitempty"`
	SinkTemperatureC      *float64 `json:"sinkTemperatureC,omitempty"`
	IrradianceWM2         *float64 `json:"irradianceWM2,omitempty"`
	WindSpeedMS           *float64 `json:"windSpeedMS,omitempty"`
	ModuleTemperatureC    *float64 `json:"moduleTemperatureC,omitempty"`
	SignalStrengthDBm     *int     `json:"signalStrengthDBm,omitempty"`
}
