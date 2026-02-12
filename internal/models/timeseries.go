package models

import "time"

// HistoryRequest describes what historical data to fetch.
type HistoryRequest struct {
	DeviceID    string      `json:"deviceId"`
	StartTime   string      `json:"startTime"`
	EndTime     string      `json:"endTime"`
	Granularity Granularity `json:"granularity"`
	Metrics     []string    `json:"metrics,omitempty"`
	Page        int         `json:"page,omitempty"`
	PageSize    int         `json:"pageSize,omitempty"`
}

type Granularity string

const (
	GranularityMinute Granularity = "minute"
	GranularityHour   Granularity = "hour"
	GranularityDay    Granularity = "day"
	GranularityMonth  Granularity = "month"
	GranularityYear   Granularity = "year"
)

// NormalizedTimeSeries is a single data point in a historical time series.
// For minute/hour granularity it contains power snapshots;
// for day/month/year it contains energy aggregates.
type NormalizedTimeSeries struct {
	DeviceID  string    `json:"deviceId"`
	Provider  string    `json:"provider"`
	Timestamp time.Time `json:"timestamp"`

	Granularity Granularity `json:"granularity"`

	// ── Power snapshots (for minute/hour granularity) ──
	PVPowerW         *float64 `json:"pvPowerW,omitempty"`
	LoadPowerW       *float64 `json:"loadPowerW,omitempty"`
	GridPowerW       *float64 `json:"gridPowerW,omitempty"`
	BatteryPowerW    *float64 `json:"batteryPowerW,omitempty"`
	SelfUsePowerW    *float64 `json:"selfUsePowerW,omitempty"`
	GridImportPowerW *float64 `json:"gridImportPowerW,omitempty"`
	GridExportPowerW *float64 `json:"gridExportPowerW,omitempty"`

	// Battery state
	BatterySOC     *float64        `json:"batterySOC,omitempty"`
	BatteryDirection *EnergyDirection `json:"batteryDirection,omitempty"`

	// ── Energy totals (for day/month/year granularity) ──
	PVEnergyKWh          *float64 `json:"pvEnergyKWh,omitempty"`
	LoadEnergyKWh        *float64 `json:"loadEnergyKWh,omitempty"`
	GridImportEnergyKWh  *float64 `json:"gridImportEnergyKWh,omitempty"`
	GridExportEnergyKWh  *float64 `json:"gridExportEnergyKWh,omitempty"`
	BatteryChargeKWh     *float64 `json:"batteryChargeKWh,omitempty"`
	BatteryDischargeKWh  *float64 `json:"batteryDischargeKWh,omitempty"`
	SelfConsumptionKWh   *float64 `json:"selfConsumptionKWh,omitempty"`

	// ── PV String detail (minute data) ──
	PVStrings []PVString `json:"pvStrings,omitempty"`

	// ── Phase detail (minute data) ──
	GridPhases    []PhaseData `json:"gridPhases,omitempty"`
	InverterPhases []PhaseData `json:"inverterPhases,omitempty"`

	Meta ProviderMeta `json:"meta"`
}

// TimeSeriesAggregate contains totals computed over the full query window.
type TimeSeriesAggregate struct {
	TotalPVEnergyKWh          *float64 `json:"totalPvEnergyKWh,omitempty"`
	TotalLoadEnergyKWh        *float64 `json:"totalLoadEnergyKWh,omitempty"`
	TotalGridImportKWh        *float64 `json:"totalGridImportKWh,omitempty"`
	TotalGridExportKWh        *float64 `json:"totalGridExportKWh,omitempty"`
	TotalBatteryChargeKWh     *float64 `json:"totalBatteryChargeKWh,omitempty"`
	TotalBatteryDischargeKWh  *float64 `json:"totalBatteryDischargeKWh,omitempty"`
	SelfConsumptionRate       *float64 `json:"selfConsumptionRate,omitempty"`
	SelfSufficiencyRate       *float64 `json:"selfSufficiencyRate,omitempty"`
}

// HistoryResponse wraps time series with aggregate stats.
type HistoryResponse struct {
	DeviceID    string                 `json:"deviceId"`
	Provider    string                 `json:"provider"`
	Granularity Granularity            `json:"granularity"`
	StartTime   string                 `json:"startTime"`
	EndTime     string                 `json:"endTime"`
	TotalPoints int                    `json:"totalPoints"`
	DataPoints  []NormalizedTimeSeries  `json:"dataPoints"`
	Aggregate   *TimeSeriesAggregate   `json:"aggregate,omitempty"`
}
