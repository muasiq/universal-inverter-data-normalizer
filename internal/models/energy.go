package models

import "time"

// Period defines the aggregation granularity for energy data.
type Period string

const (
	PeriodDay   Period = "day"
	PeriodWeek  Period = "week"
	PeriodMonth Period = "month"
	PeriodYear  Period = "year"
	PeriodTotal Period = "total"
)

// NormalizedEnergy is the unified energy statistics summary for a plant or device.
type NormalizedEnergy struct {
	ID       string `json:"id"`       // plant or device ID
	Provider string `json:"provider"`
	Period   Period `json:"period"`

	Timestamp  time.Time `json:"timestamp"`
	PeriodStart time.Time `json:"periodStart,omitempty"`
	PeriodEnd   time.Time `json:"periodEnd,omitempty"`

	// ── Generation ──
	PVGenerationKWh *float64 `json:"pvGenerationKWh,omitempty"`

	// ── Consumption ──
	LoadConsumptionKWh *float64 `json:"loadConsumptionKWh,omitempty"`

	// ── Grid Exchange ──
	GridImportKWh *float64 `json:"gridImportKWh,omitempty"` // Bought from grid
	GridExportKWh *float64 `json:"gridExportKWh,omitempty"` // Sold to grid

	// ── Battery ──
	BatteryChargeKWh    *float64 `json:"batteryChargeKWh,omitempty"`
	BatteryDischargeKWh *float64 `json:"batteryDischargeKWh,omitempty"`

	// ── Self-Consumption ──
	SelfConsumptionKWh  *float64 `json:"selfConsumptionKWh,omitempty"`
	SelfConsumptionRate *float64 `json:"selfConsumptionRate,omitempty"` // 0.0 - 1.0
	SelfSufficiencyRate *float64 `json:"selfSufficiencyRate,omitempty"` // 0.0 - 1.0

	// ── Environmental Impact ──
	CO2SavedKg      *float64 `json:"co2SavedKg,omitempty"`
	TreesEquivalent *float64 `json:"treesEquivalent,omitempty"`

	// ── Financial ──
	Revenue  *float64 `json:"revenue,omitempty"`
	Savings  *float64 `json:"savings,omitempty"`
	Currency string   `json:"currency,omitempty"`

	// ── Current Power (snapshot at query time) ──
	CurrentPowerW *float64 `json:"currentPowerW,omitempty"`

	// ── Status ──
	DeviceStatus *DeviceStatus `json:"deviceStatus,omitempty"`
	BatterySOC   *float64      `json:"batterySOC,omitempty"`

	Meta ProviderMeta `json:"meta"`
}

// EnergyBreakdown is used for multi-period comparisons (e.g. daily bars in a month chart).
type EnergyBreakdown struct {
	Period    string  `json:"period"`     // "2025-01-15" or "2025-01" etc.
	GenKWh   float64 `json:"genKWh"`
	ConKWh   float64 `json:"conKWh"`
	ImpKWh   float64 `json:"impKWh"`
	ExpKWh   float64 `json:"expKWh"`
	ChgKWh   float64 `json:"chgKWh"`
	DischKWh float64 `json:"dischKWh"`
}
