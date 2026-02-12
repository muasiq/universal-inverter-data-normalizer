package sma

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/muasiq/universal-inverter-data-normalizer/internal/models"
	"github.com/muasiq/universal-inverter-data-normalizer/internal/provider"
	"github.com/rs/zerolog/log"
)

const (
	defaultBaseURL = "https://monitoring.smaapis.de/monitoring/v1"
	sandboxBaseURL = "https://sandbox.smaapis.de/monitoring/v1"
	providerName   = "sma"
)

func init() {
	provider.Register(providerName, func() provider.Provider {
		return &SMAProvider{}
	})
}

// SMAProvider implements the Provider interface for SMA Monitoring API.
// SMA uses OAuth2 Bearer tokens. Their API provides:
//   - /plants — list of solar systems
//   - /plants/{plantId}/devices — devices in a plant
//   - /plants/{plantId}/measurements/sets/{setName}/{period} — energy data
//   - /devices/{deviceId}/measurements/sets/{setName}/{period} — device measurements
//   - /plants/{plantId}/logs — plant log events
//   - /devices/{deviceId}/logs — device log events
type SMAProvider struct {
	client *provider.HTTPClient
	config provider.ProviderConfig
	token  string
}

func (p *SMAProvider) Name() string { return providerName }

func (p *SMAProvider) Initialize(ctx context.Context, cfg provider.ProviderConfig) error {
	p.config = cfg

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	rps := cfg.RateLimitRPS
	if rps <= 0 {
		rps = 5
	}

	p.client = provider.NewHTTPClient(baseURL, cfg.TimeoutSeconds, rps)

	// SMA uses OAuth2 Bearer token
	p.token = cfg.GetCredential("bearerToken")
	if p.token == "" {
		return fmt.Errorf("SMA provider requires 'bearerToken' credential")
	}
	p.client.SetHeader("Authorization", "Bearer "+p.token)

	log.Info().Str("provider", providerName).Msg("Initialized")
	return nil
}

// ── Plants ──

func (p *SMAProvider) GetPlants(ctx context.Context) ([]models.NormalizedPlant, error) {
	var resp smaPlantListResponse
	if err := p.client.Get(ctx, "/plants", nil, &resp); err != nil {
		return nil, fmt.Errorf("SMA GetPlants: %w", err)
	}

	var plants []models.NormalizedPlant
	for _, raw := range resp.Plants {
		plants = append(plants, normalizeSMAPlant(raw))
	}
	return plants, nil
}

func (p *SMAProvider) GetPlantDetails(ctx context.Context, plantID string) (*models.NormalizedPlant, error) {
	// SMA: /plants/{plantId}/installation for detailed info
	var resp smaPlantInstallation
	if err := p.client.Get(ctx, fmt.Sprintf("/plants/%s/installation", plantID), nil, &resp); err != nil {
		return nil, fmt.Errorf("SMA GetPlantDetails: %w", err)
	}

	plant := normalizeSMAPlantInstallation(resp, plantID)
	return &plant, nil
}

// ── Devices ──

func (p *SMAProvider) GetDevices(ctx context.Context, plantID string) ([]models.NormalizedDevice, error) {
	var resp smaDeviceListResponse
	if err := p.client.Get(ctx, fmt.Sprintf("/plants/%s/devices", plantID), nil, &resp); err != nil {
		return nil, fmt.Errorf("SMA GetDevices: %w", err)
	}

	var devices []models.NormalizedDevice
	for _, raw := range resp.Devices {
		devices = append(devices, normalizeSMADevice(raw, plantID))
	}
	return devices, nil
}

func (p *SMAProvider) GetDeviceDetails(ctx context.Context, deviceID string) (*models.NormalizedDevice, error) {
	// SMA doesn't have a dedicated single-device endpoint; info is within plant devices
	return nil, fmt.Errorf("SMA: use GetDevices with plantID to retrieve device details")
}

// ── Real-Time Data ──

func (p *SMAProvider) GetRealTimeData(ctx context.Context, deviceID string) (*models.NormalizedRealtime, error) {
	// SMA: Use PowerAc and PowerDc measurement sets for near-realtime data
	// GET /devices/{deviceId}/measurements/sets/PowerAc/Recent
	// GET /devices/{deviceId}/measurements/sets/PowerDc/Recent
	var acResp smaMeasurementSetResponse
	err := p.client.Get(ctx, fmt.Sprintf("/devices/%s/measurements/sets/PowerAc/Recent", deviceID), nil, &acResp)
	if err != nil {
		return nil, fmt.Errorf("SMA GetRealTimeData (AC): %w", err)
	}

	var dcResp smaMeasurementSetResponse
	_ = p.client.Get(ctx, fmt.Sprintf("/devices/%s/measurements/sets/PowerDc/Recent", deviceID), nil, &dcResp)

	rt := normalizeSMARealtime(deviceID, acResp, dcResp)
	return &rt, nil
}

// ── Energy Stats ──

func (p *SMAProvider) GetEnergyStats(ctx context.Context, plantID string, period models.Period) (*models.NormalizedEnergy, error) {
	// SMA: /plants/{plantId}/measurements/sets/EnergyBalance/{period}
	smaPeriod := periodToSMAPeriod(period)

	var resp smaMeasurementSetResponse
	path := fmt.Sprintf("/plants/%s/measurements/sets/EnergyBalance/%s", plantID, smaPeriod)
	if err := p.client.Get(ctx, path, nil, &resp); err != nil {
		return nil, fmt.Errorf("SMA GetEnergyStats: %w", err)
	}

	energy := normalizeSMAEnergy(resp, plantID, period)
	return &energy, nil
}

// ── Historical Data ──

func (p *SMAProvider) GetHistoricalData(ctx context.Context, deviceID string, req models.HistoryRequest) (*models.HistoryResponse, error) {
	// SMA: /devices/{deviceId}/measurements/sets/{setName}/{period}
	smaPeriod := granularityToSMAPeriod(req.Granularity)

	params := url.Values{
		"Date": {req.StartTime},
	}

	// Fetch both AC power and DC power
	var acResp smaMeasurementSetResponse
	err := p.client.Get(ctx, fmt.Sprintf("/devices/%s/measurements/sets/PowerAc/%s", deviceID, smaPeriod), params, &acResp)
	if err != nil {
		return nil, fmt.Errorf("SMA GetHistoricalData: %w", err)
	}

	history := normalizeSMAHistory(acResp, deviceID, req)
	return &history, nil
}

// ── Alarms ──

func (p *SMAProvider) GetAlarms(ctx context.Context, deviceID string) ([]models.NormalizedAlarm, error) {
	// SMA: /devices/{deviceId}/logs
	var resp smaLogResponse
	if err := p.client.Get(ctx, fmt.Sprintf("/devices/%s/logs", deviceID), nil, &resp); err != nil {
		return nil, fmt.Errorf("SMA GetAlarms: %w", err)
	}

	var alarms []models.NormalizedAlarm
	for _, entry := range resp.Logs {
		if entry.Severity == "alarm" || entry.Severity == "error" || entry.Severity == "warning" {
			alarms = append(alarms, normalizeSMALogEntry(entry, deviceID))
		}
	}
	return alarms, nil
}

func (p *SMAProvider) GetAllAlarms(ctx context.Context) ([]models.NormalizedAlarm, error) {
	// Get plants first, then iterate
	plants, err := p.GetPlants(ctx)
	if err != nil {
		return nil, err
	}

	var allAlarms []models.NormalizedAlarm
	for _, plant := range plants {
		var resp smaLogResponse
		plantID := plant.Meta.ProviderPlantID
		if err := p.client.Get(ctx, fmt.Sprintf("/plants/%s/logs", plantID), nil, &resp); err != nil {
			log.Warn().Err(err).Str("plantId", plantID).Msg("Failed to fetch SMA plant logs")
			continue
		}
		for _, entry := range resp.Logs {
			if entry.Severity == "alarm" || entry.Severity == "error" || entry.Severity == "warning" {
				allAlarms = append(allAlarms, normalizeSMALogEntry(entry, ""))
			}
		}
	}
	return allAlarms, nil
}

func (p *SMAProvider) Healthy(ctx context.Context) bool {
	return p.token != ""
}

func (p *SMAProvider) Close() error {
	return nil
}

// ══════════════════════════════════════════════════════════════════
// SMA raw API response types
// ══════════════════════════════════════════════════════════════════

type smaPlantListResponse struct {
	Plants []smaPlant `json:"plants"`
}

type smaPlant struct {
	PlantID string `json:"plantId"`
	Name    string `json:"name"`
}

type smaPlantInstallation struct {
	PlantID              string   `json:"plantId"`
	Name                 string   `json:"name"`
	Timezone             string   `json:"timezone"`
	Latitude             *float64 `json:"latitude"`
	Longitude            *float64 `json:"longitude"`
	PeakPower            *float64 `json:"peakPower"` // Wp
	CalculatedAcNomPower *float64 `json:"calculatedAcNominalPower"` // W
	CO2SavingsFactor     *float64 `json:"co2SavingsFactor"` // g/kWh
	Country              string   `json:"country"`
	City                 string   `json:"city"`
	Street               string   `json:"street"`
	PostalCode           string   `json:"postalCode"`
}

type smaDeviceListResponse struct {
	Devices []smaDevice `json:"devices"`
}

type smaDevice struct {
	DeviceID     string `json:"deviceId"`
	Name         string `json:"name"`
	Serial       string `json:"serial"`
	Product      string `json:"product"`
	DeviceType   string `json:"type"`
	IsActive     bool   `json:"isActive"`
	Status       string `json:"status"`
	Manufacturer string `json:"manufacturer"`
}

type smaMeasurementSetResponse struct {
	SetType string                 `json:"setType"`
	Sets    []smaMeasurementEntry  `json:"set"`
}

type smaMeasurementEntry struct {
	Time   string                  `json:"time"`
	Values map[string]*float64     `json:"values"`
}

type smaLogResponse struct {
	Logs []smaLogEntry `json:"logs"`
}

type smaLogEntry struct {
	LogID    string `json:"logId"`
	Time     string `json:"time"`
	Message  string `json:"message"`
	Severity string `json:"severity"` // "info", "warning", "alarm", "error"
	DeviceID string `json:"deviceId"`
}

// ══════════════════════════════════════════════════════════════════
// Normalization functions
// ══════════════════════════════════════════════════════════════════

func normalizeSMAPlant(raw smaPlant) models.NormalizedPlant {
	return models.NormalizedPlant{
		ID:       fmt.Sprintf("%s_%s", providerName, raw.PlantID),
		Provider: providerName,
		Name:     raw.Name,
		Meta: models.ProviderMeta{
			Provider:        providerName,
			ProviderPlantID: raw.PlantID,
			FetchedAt:       time.Now().UTC(),
		},
	}
}

func normalizeSMAPlantInstallation(raw smaPlantInstallation, plantID string) models.NormalizedPlant {
	plant := models.NormalizedPlant{
		ID:       fmt.Sprintf("%s_%s", providerName, plantID),
		Provider: providerName,
		Name:     raw.Name,
		Timezone: raw.Timezone,
		Country:  raw.Country,
		Address:  fmt.Sprintf("%s, %s %s", raw.Street, raw.PostalCode, raw.City),
		Meta: models.ProviderMeta{
			Provider:        providerName,
			ProviderPlantID: plantID,
			FetchedAt:       time.Now().UTC(),
		},
	}

	if raw.Latitude != nil && raw.Longitude != nil {
		plant.Location = &models.LatLng{
			Latitude:  *raw.Latitude,
			Longitude: *raw.Longitude,
		}
	}

	// SMA peakPower is in Wp, convert to kWp
	if raw.PeakPower != nil {
		kWp := *raw.PeakPower / 1000.0
		plant.PeakPowerKWp = &kWp
	}

	return plant
}

func normalizeSMADevice(raw smaDevice, plantID string) models.NormalizedDevice {
	return models.NormalizedDevice{
		ID:           fmt.Sprintf("%s_%s", providerName, raw.DeviceID),
		Provider:     providerName,
		PlantID:      fmt.Sprintf("%s_%s", providerName, plantID),
		Name:         raw.Name,
		SerialNumber: raw.Serial,
		Model:        raw.Product,
		DeviceType:   smaDeviceType(raw.DeviceType),
		Manufacturer: defaultIfEmpty(raw.Manufacturer, "SMA"),
		Status:       smaDeviceStatus(raw.Status, raw.IsActive),
		IsOnline:     raw.IsActive,
		Meta: models.ProviderMeta{
			Provider:         providerName,
			ProviderDeviceID: raw.DeviceID,
			ProviderPlantID:  plantID,
			FetchedAt:        time.Now().UTC(),
		},
	}
}

func normalizeSMARealtime(deviceID string, acResp, dcResp smaMeasurementSetResponse) models.NormalizedRealtime {
	now := time.Now().UTC()

	rt := models.NormalizedRealtime{
		DeviceID:      fmt.Sprintf("%s_%s", providerName, deviceID),
		Provider:      providerName,
		Timestamp:     now,
		Status:        models.DeviceStatusOnline,
		OperatingMode: models.OperatingModeGridConnected,
		Meta: models.ProviderMeta{
			Provider:         providerName,
			ProviderDeviceID: deviceID,
			RawDataAvailable: true,
			FetchedAt:        now,
		},
	}

	// Parse latest AC measurement
	if len(acResp.Sets) > 0 {
		latest := acResp.Sets[len(acResp.Sets)-1]
		if t, err := time.Parse(time.RFC3339, latest.Time); err == nil {
			rt.Timestamp = t
		}

		var phases []models.PhaseData
		totalPower := 0.0

		// SMA uses keys like "PowerActive", "VoltagePhaseA", "CurrentPhaseA"
		for _, phase := range []string{"A", "B", "C"} {
			v := latest.Values["VoltagePhase"+phase]
			c := latest.Values["CurrentPhase"+phase]
			pw := latest.Values["PowerActivePhase"+phase]
			f := latest.Values["FrequencyPhase"+phase]

			if v != nil || c != nil || pw != nil {
				pd := models.PhaseData{Phase: phase, VoltageV: v, CurrentA: c, PowerW: pw, FrequencyHz: f}
				phases = append(phases, pd)
				if pw != nil {
					totalPower += *pw
				}
			}
		}

		// Total active power
		if pa := latest.Values["PowerActive"]; pa != nil {
			totalPower = *pa
		}

		rt.Grid = &models.GridData{
			TotalPowerW: totalPower,
			Direction:   models.GridDirectionIdle,
			Phases:      phases,
		}
	}

	// Parse DC (PV) measurements
	if len(dcResp.Sets) > 0 {
		latest := dcResp.Sets[len(dcResp.Sets)-1]
		totalDC := 0.0
		var pvStrings []models.PVString

		for i := 1; i <= 12; i++ {
			vKey := fmt.Sprintf("Voltage%d", i)
			cKey := fmt.Sprintf("Current%d", i)
			pKey := fmt.Sprintf("Power%d", i)
			v := latest.Values[vKey]
			c := latest.Values[cKey]
			pw := latest.Values[pKey]
			if v != nil || c != nil || pw != nil {
				pvStrings = append(pvStrings, models.PVString{ID: i, VoltageV: v, CurrentA: c, PowerW: pw})
				if pw != nil {
					totalDC += *pw
				}
			}
		}

		if dc := latest.Values["PowerDc"]; dc != nil {
			totalDC = *dc
		}

		rt.PV = &models.PVData{
			TotalPowerW: totalDC,
			Strings:     pvStrings,
		}
	}

	return rt
}

func normalizeSMAEnergy(resp smaMeasurementSetResponse, plantID string, period models.Period) models.NormalizedEnergy {
	energy := models.NormalizedEnergy{
		ID:        fmt.Sprintf("%s_%s", providerName, plantID),
		Provider:  providerName,
		Period:    period,
		Timestamp: time.Now().UTC(),
		Meta: models.ProviderMeta{
			Provider:        providerName,
			ProviderPlantID: plantID,
			FetchedAt:       time.Now().UTC(),
		},
	}

	// SMA EnergyBalance keys: "PvGeneration", "TotalConsumption",
	// "GridFeedIn", "GridPurchase", "BatteryCharge", "BatteryDischarge",
	// "DirectConsumption", "SelfConsumption"

	// Aggregate across all measurement entries for the period
	var pvGen, consumption, gridFeedIn, gridPurchase, batChg, batDis float64
	for _, entry := range resp.Sets {
		if v := entry.Values["PvGeneration"]; v != nil {
			pvGen += *v / 1000.0 // Wh → kWh
		}
		if v := entry.Values["TotalConsumption"]; v != nil {
			consumption += *v / 1000.0
		}
		if v := entry.Values["GridFeedIn"]; v != nil {
			gridFeedIn += *v / 1000.0
		}
		if v := entry.Values["GridPurchase"]; v != nil {
			gridPurchase += *v / 1000.0
		}
		if v := entry.Values["BatteryCharge"]; v != nil {
			batChg += *v / 1000.0
		}
		if v := entry.Values["BatteryDischarge"]; v != nil {
			batDis += *v / 1000.0
		}
	}

	energy.PVGenerationKWh = &pvGen
	energy.LoadConsumptionKWh = &consumption
	energy.GridExportKWh = &gridFeedIn
	energy.GridImportKWh = &gridPurchase
	energy.BatteryChargeKWh = &batChg
	energy.BatteryDischargeKWh = &batDis

	// Self-consumption rate
	if pvGen > 0 {
		rate := (pvGen - gridFeedIn) / pvGen
		energy.SelfConsumptionRate = &rate
	}
	if consumption > 0 {
		rate := (consumption - gridPurchase) / consumption
		energy.SelfSufficiencyRate = &rate
	}

	return energy
}

func normalizeSMAHistory(resp smaMeasurementSetResponse, deviceID string, req models.HistoryRequest) models.HistoryResponse {
	result := models.HistoryResponse{
		DeviceID:    fmt.Sprintf("%s_%s", providerName, deviceID),
		Provider:    providerName,
		Granularity: req.Granularity,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
	}

	for _, entry := range resp.Sets {
		ts := time.Now()
		if t, err := time.Parse(time.RFC3339, entry.Time); err == nil {
			ts = t
		}

		dp := models.NormalizedTimeSeries{
			DeviceID:    result.DeviceID,
			Provider:    providerName,
			Timestamp:   ts,
			Granularity: req.Granularity,
			Meta: models.ProviderMeta{
				Provider:         providerName,
				ProviderDeviceID: deviceID,
				FetchedAt:        time.Now().UTC(),
			},
		}

		// Map SMA measurement values to normalized fields
		if v := entry.Values["PowerActive"]; v != nil {
			dp.PVPowerW = v
		}

		var phases []models.PhaseData
		for _, phase := range []string{"A", "B", "C"} {
			v := entry.Values["VoltagePhase"+phase]
			c := entry.Values["CurrentPhase"+phase]
			pw := entry.Values["PowerActivePhase"+phase]
			if v != nil || c != nil || pw != nil {
				phases = append(phases, models.PhaseData{
					Phase: phase, VoltageV: v, CurrentA: c, PowerW: pw,
				})
			}
		}
		dp.GridPhases = phases

		result.DataPoints = append(result.DataPoints, dp)
	}

	result.TotalPoints = len(result.DataPoints)
	return result
}

func normalizeSMALogEntry(entry smaLogEntry, deviceID string) models.NormalizedAlarm {
	ts := time.Now()
	if t, err := time.Parse(time.RFC3339, entry.Time); err == nil {
		ts = t
	}

	return models.NormalizedAlarm{
		ID:       fmt.Sprintf("%s_log_%s", providerName, entry.LogID),
		Provider: providerName,
		DeviceID: fmt.Sprintf("%s_%s", providerName, entry.DeviceID),
		Code:     entry.LogID,
		Name:     entry.Message,
		Message:  entry.Message,
		Severity: smaLogSeverity(entry.Severity),
		Status:   models.AlarmStatusActive,
		StartTime: ts,
		Meta: models.ProviderMeta{
			Provider:         providerName,
			ProviderDeviceID: entry.DeviceID,
			FetchedAt:        time.Now().UTC(),
		},
	}
}

// ══════════════════════════════════════════════════════════════════
// Mapping helpers
// ══════════════════════════════════════════════════════════════════

func smaDeviceType(t string) models.DeviceType {
	switch t {
	case "SolarInverter", "PV Inverter":
		return models.DeviceTypeStringInverter
	case "HybridInverter":
		return models.DeviceTypeHybridInverter
	case "Battery", "BatteryInverter":
		return models.DeviceTypeBattery
	case "EnergyMeter", "Meter":
		return models.DeviceTypeMeter
	case "Gateway", "CommunicationProduct":
		return models.DeviceTypeGateway
	case "Sensor", "SatelliteSensor":
		return models.DeviceTypeWeatherStation
	case "EVCharger":
		return models.DeviceTypeEVCharger
	default:
		return models.DeviceTypeUnknown
	}
}

func smaDeviceStatus(status string, isActive bool) models.DeviceStatus {
	if !isActive {
		return models.DeviceStatusOffline
	}
	switch status {
	case "Ok", "ok":
		return models.DeviceStatusNormal
	case "Warning", "warning":
		return models.DeviceStatusWarning
	case "Error", "error", "Alarm", "alarm":
		return models.DeviceStatusFault
	default:
		return models.DeviceStatusOnline
	}
}

func smaLogSeverity(severity string) models.AlarmSeverity {
	switch severity {
	case "info":
		return models.AlarmSeverityInfo
	case "warning":
		return models.AlarmSeverityWarning
	case "alarm", "error":
		return models.AlarmSeverityCritical
	default:
		return models.AlarmSeverityUnknown
	}
}

func periodToSMAPeriod(p models.Period) string {
	switch p {
	case models.PeriodDay:
		return "Day"
	case models.PeriodWeek:
		return "Week"
	case models.PeriodMonth:
		return "Month"
	case models.PeriodYear:
		return "Year"
	case models.PeriodTotal:
		return "Total"
	default:
		return "Day"
	}
}

func granularityToSMAPeriod(g models.Granularity) string {
	switch g {
	case models.GranularityMinute:
		return "FiveMinutes"
	case models.GranularityHour:
		return "QuarterOfAnHour"
	case models.GranularityDay:
		return "Day"
	case models.GranularityMonth:
		return "Month"
	case models.GranularityYear:
		return "Year"
	default:
		return "Day"
	}
}

func defaultIfEmpty(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

// Unused import suppressor
var _ = strconv.Itoa
