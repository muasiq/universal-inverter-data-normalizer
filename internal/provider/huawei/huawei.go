package huawei

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/muasiq/universal-inverter-data-normalizer/internal/models"
	"github.com/muasiq/universal-inverter-data-normalizer/internal/provider"
	"github.com/rs/zerolog/log"
)

const (
	defaultBaseURL = "https://eu5.fusionsolar.huawei.com/thirdData"
	providerName   = "huawei"
)

func init() {
	provider.Register(providerName, func() provider.Provider {
		return &HuaweiProvider{}
	})
}

// HuaweiProvider implements the Provider interface for Huawei FusionSolar northbound API.
// Huawei uses a session-based auth: POST /login → returns XSRF-TOKEN cookie.
// Key endpoints:
//   - POST /getStationList — list plants
//   - POST /getStationRealKpi — plant real-time KPIs
//   - POST /getDevList — device list
//   - POST /getDevRealKpi — device real-time KPIs
//   - POST /getKpiStationHour/Day/Month/Year — historical data
//   - POST /getAlarmList — alarms
type HuaweiProvider struct {
	client    *provider.HTTPClient
	config    provider.ProviderConfig
	xsrfToken string
}

func (p *HuaweiProvider) Name() string { return providerName }

func (p *HuaweiProvider) Initialize(ctx context.Context, cfg provider.ProviderConfig) error {
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
	p.client.SetHeader("Content-Type", "application/json")

	return p.authenticate(ctx)
}

func (p *HuaweiProvider) authenticate(ctx context.Context) error {
	username := p.config.GetCredential("username")
	password := p.config.GetCredential("systemCode")

	body := map[string]string{
		"userName":   username,
		"systemCode": password,
	}

	var resp huaweiLoginResponse
	if err := p.client.Post(ctx, "/login", body, &resp); err != nil {
		return fmt.Errorf("Huawei auth: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("Huawei auth failed: failCode=%d message=%s", resp.FailCode, resp.Message)
	}

	// The XSRF token would come from the Set-Cookie header in a real implementation.
	// For the normalized client, we store it for subsequent requests.
	p.xsrfToken = "authenticated"
	log.Info().Str("provider", providerName).Msg("Authenticated successfully")
	return nil
}

// ── Plants ──

func (p *HuaweiProvider) GetPlants(ctx context.Context) ([]models.NormalizedPlant, error) {
	var allPlants []models.NormalizedPlant
	pageNo := 1

	for {
		body := map[string]interface{}{
			"pageNo": pageNo,
		}
		var resp huaweiStationListResponse
		if err := p.client.Post(ctx, "/getStationList", body, &resp); err != nil {
			return nil, fmt.Errorf("Huawei GetPlants: %w", err)
		}
		if !resp.Success {
			return nil, fmt.Errorf("Huawei GetPlants failed: %d %s", resp.FailCode, resp.Message)
		}

		for _, raw := range resp.Data.List {
			allPlants = append(allPlants, normalizeHuaweiStation(raw))
		}

		if pageNo*20 >= resp.Data.Total {
			break
		}
		pageNo++
	}

	return allPlants, nil
}

func (p *HuaweiProvider) GetPlantDetails(ctx context.Context, plantID string) (*models.NormalizedPlant, error) {
	body := map[string]interface{}{
		"stationCodes": plantID,
	}
	var resp huaweiStationRealKpiResponse
	if err := p.client.Post(ctx, "/getStationRealKpi", body, &resp); err != nil {
		return nil, err
	}

	if !resp.Success || len(resp.Data) == 0 {
		return nil, fmt.Errorf("Huawei GetPlantDetails: no data")
	}

	plant := normalizeHuaweiStationKpi(resp.Data[0], plantID)
	return &plant, nil
}

// ── Devices ──

func (p *HuaweiProvider) GetDevices(ctx context.Context, plantID string) ([]models.NormalizedDevice, error) {
	body := map[string]interface{}{
		"stationCodes": plantID,
	}
	var resp huaweiDevListResponse
	if err := p.client.Post(ctx, "/getDevList", body, &resp); err != nil {
		return nil, fmt.Errorf("Huawei GetDevices: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("Huawei GetDevices failed: %s", resp.Message)
	}

	var devices []models.NormalizedDevice
	for _, raw := range resp.Data {
		devices = append(devices, normalizeHuaweiDevice(raw, plantID))
	}
	return devices, nil
}

func (p *HuaweiProvider) GetDeviceDetails(ctx context.Context, deviceID string) (*models.NormalizedDevice, error) {
	return nil, fmt.Errorf("Huawei: use GetDevices with stationCode for device details")
}

// ── Real-Time Data ──

func (p *HuaweiProvider) GetRealTimeData(ctx context.Context, deviceID string) (*models.NormalizedRealtime, error) {
	// Huawei: POST /getDevRealKpi with devIds and devTypeId
	// We need to figure out the device type; default to inverter (typeId=1)
	body := map[string]interface{}{
		"devIds":    deviceID,
		"devTypeId": 1, // 1=inverter
	}

	var resp huaweiDevRealKpiResponse
	if err := p.client.Post(ctx, "/getDevRealKpi", body, &resp); err != nil {
		return nil, fmt.Errorf("Huawei GetRealTimeData: %w", err)
	}
	if !resp.Success || len(resp.Data) == 0 {
		return nil, fmt.Errorf("Huawei GetRealTimeData: no data")
	}

	rt := normalizeHuaweiDevRealKpi(resp.Data[0], deviceID)
	return &rt, nil
}

// ── Energy Stats ──

func (p *HuaweiProvider) GetEnergyStats(ctx context.Context, plantID string, period models.Period) (*models.NormalizedEnergy, error) {
	// Use getStationRealKpi for current stats
	body := map[string]interface{}{
		"stationCodes": plantID,
	}
	var resp huaweiStationRealKpiResponse
	if err := p.client.Post(ctx, "/getStationRealKpi", body, &resp); err != nil {
		return nil, fmt.Errorf("Huawei GetEnergyStats: %w", err)
	}
	if !resp.Success || len(resp.Data) == 0 {
		return nil, fmt.Errorf("Huawei GetEnergyStats: no data")
	}

	energy := normalizeHuaweiStationEnergy(resp.Data[0], plantID, period)
	return &energy, nil
}

// ── Historical Data ──

func (p *HuaweiProvider) GetHistoricalData(ctx context.Context, deviceID string, req models.HistoryRequest) (*models.HistoryResponse, error) {
	endpoint := huaweiKpiEndpoint(req.Granularity)

	collectTime := time.Now().UnixMilli()
	if t, err := time.Parse(time.RFC3339, req.StartTime); err == nil {
		collectTime = t.UnixMilli()
	}

	body := map[string]interface{}{
		"stationCodes": deviceID,
		"collectTime":  collectTime,
	}

	var resp huaweiKpiHistoryResponse
	if err := p.client.Post(ctx, endpoint, body, &resp); err != nil {
		return nil, fmt.Errorf("Huawei GetHistoricalData: %w", err)
	}

	history := normalizeHuaweiHistory(resp, deviceID, req)
	return &history, nil
}

// ── Alarms ──

func (p *HuaweiProvider) GetAlarms(ctx context.Context, deviceID string) ([]models.NormalizedAlarm, error) {
	now := time.Now()
	body := map[string]interface{}{
		"stationCodes": deviceID,
		"beginTime":    now.Add(-24 * time.Hour).UnixMilli(),
		"endTime":      now.UnixMilli(),
		"language":     "en_US",
	}

	var resp huaweiAlarmResponse
	if err := p.client.Post(ctx, "/getAlarmList", body, &resp); err != nil {
		return nil, fmt.Errorf("Huawei GetAlarms: %w", err)
	}

	var alarms []models.NormalizedAlarm
	for _, raw := range resp.Data {
		alarms = append(alarms, normalizeHuaweiAlarm(raw))
	}
	return alarms, nil
}

func (p *HuaweiProvider) GetAllAlarms(ctx context.Context) ([]models.NormalizedAlarm, error) {
	return p.GetAlarms(ctx, "")
}

func (p *HuaweiProvider) Healthy(ctx context.Context) bool {
	return p.xsrfToken != ""
}

func (p *HuaweiProvider) Close() error {
	return nil
}

// ══════════════════════════════════════════════════════════════════
// Huawei raw response types
// ══════════════════════════════════════════════════════════════════

type huaweiBaseResponse struct {
	Success  bool   `json:"success"`
	FailCode int    `json:"failCode"`
	Message  string `json:"message"`
}

type huaweiLoginResponse struct {
	huaweiBaseResponse
}

type huaweiStationListResponse struct {
	huaweiBaseResponse
	Data struct {
		Total int             `json:"total"`
		PageCount int         `json:"pageCount"`
		List  []huaweiStation `json:"list"`
	} `json:"data"`
}

type huaweiStation struct {
	StationCode  string  `json:"stationCode"`
	StationName  string  `json:"stationName"`
	StationAddr  string  `json:"stationAddr"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Capacity     float64 `json:"capacity"` // kWp
	ContactPerson string `json:"contactPerson"`
}

type huaweiStationRealKpiResponse struct {
	huaweiBaseResponse
	Data []huaweiStationKpi `json:"data"`
}

type huaweiStationKpi struct {
	StationCode string                 `json:"stationCode"`
	DataItemMap map[string]interface{} `json:"dataItemMap"`
}

type huaweiDevListResponse struct {
	huaweiBaseResponse
	Data []huaweiDevInfo `json:"data"`
}

type huaweiDevInfo struct {
	DevID      int64  `json:"id"`
	DevName    string `json:"devName"`
	DevTypeID  int    `json:"devTypeId"` // 1=inverter, 2=meter, 38=battery, 39=optimizer, 46=EMS, 47=dongle
	StationCode string `json:"stationCode"`
	InvType    string `json:"invType"`
	SoftwareVersion string `json:"softwareVersion"`
	ESN        string `json:"esnCode"`
}

type huaweiDevRealKpiResponse struct {
	huaweiBaseResponse
	Data []huaweiDevKpi `json:"data"`
}

type huaweiDevKpi struct {
	DevID      int64                  `json:"devId"`
	DataItemMap map[string]interface{} `json:"dataItemMap"`
}

type huaweiKpiHistoryResponse struct {
	huaweiBaseResponse
	Data []huaweiKpiHistoryEntry `json:"data"`
}

type huaweiKpiHistoryEntry struct {
	StationCode string                 `json:"stationCode"`
	CollectTime int64                  `json:"collectTime"`
	DataItemMap map[string]interface{} `json:"dataItemMap"`
}

type huaweiAlarmResponse struct {
	huaweiBaseResponse
	Data []huaweiAlarm `json:"data"`
}

type huaweiAlarm struct {
	AlarmID     int64  `json:"alarmId"`
	AlarmName   string `json:"alarmName"`
	DevName     string `json:"devName"`
	DeviceESN   string `json:"esnCode"`
	StationCode string `json:"stationCode"`
	StationName string `json:"stationName"`
	Severity    int    `json:"severity"` // 1=critical, 2=major, 3=minor, 4=warning
	Status      int    `json:"status"`   // 1=active, 2=recovered
	RaiseTime   int64  `json:"raiseTime"`
	ClearTime   int64  `json:"clearTime"`
}

// ══════════════════════════════════════════════════════════════════
// Normalization functions
// ══════════════════════════════════════════════════════════════════

func normalizeHuaweiStation(raw huaweiStation) models.NormalizedPlant {
	capacity := raw.Capacity
	return models.NormalizedPlant{
		ID:           fmt.Sprintf("%s_%s", providerName, raw.StationCode),
		Provider:     providerName,
		Name:         raw.StationName,
		Address:      raw.StationAddr,
		PeakPowerKWp: &capacity,
		Location: &models.LatLng{
			Latitude:  raw.Latitude,
			Longitude: raw.Longitude,
		},
		Meta: models.ProviderMeta{
			Provider:        providerName,
			ProviderPlantID: raw.StationCode,
			FetchedAt:       time.Now().UTC(),
		},
	}
}

func normalizeHuaweiStationKpi(raw huaweiStationKpi, plantID string) models.NormalizedPlant {
	return models.NormalizedPlant{
		ID:       fmt.Sprintf("%s_%s", providerName, plantID),
		Provider: providerName,
		Name:     plantID,
		Meta: models.ProviderMeta{
			Provider:        providerName,
			ProviderPlantID: plantID,
			RawDataAvailable: true,
			FetchedAt:       time.Now().UTC(),
		},
	}
}

func normalizeHuaweiDevice(raw huaweiDevInfo, plantID string) models.NormalizedDevice {
	devID := strconv.FormatInt(raw.DevID, 10)
	return models.NormalizedDevice{
		ID:           fmt.Sprintf("%s_%s", providerName, devID),
		Provider:     providerName,
		PlantID:      fmt.Sprintf("%s_%s", providerName, plantID),
		Name:         raw.DevName,
		SerialNumber: raw.ESN,
		Model:        raw.InvType,
		DeviceType:   huaweiDeviceType(raw.DevTypeID),
		Manufacturer: "Huawei",
		Status:       models.DeviceStatusUnknown,
		FirmwareInfo: &models.FirmwareInfo{
			MainVersion: raw.SoftwareVersion,
		},
		Meta: models.ProviderMeta{
			Provider:         providerName,
			ProviderDeviceID: devID,
			ProviderPlantID:  plantID,
			FetchedAt:        time.Now().UTC(),
			Extra: map[string]string{
				"devTypeId":   strconv.Itoa(raw.DevTypeID),
				"stationCode": raw.StationCode,
			},
		},
	}
}

func normalizeHuaweiDevRealKpi(raw huaweiDevKpi, deviceID string) models.NormalizedRealtime {
	now := time.Now().UTC()
	data := raw.DataItemMap

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

	// Huawei inverter real-time KPI keys:
	// "day_cap" (daily yield kWh), "mppt_power" (DC input power kW),
	// "active_power" (AC output power kW), "reactive_power",
	// "power_factor", "efficiency", "temperature",
	// "pv1_u", "pv1_i", "pv2_u", "pv2_i", ... (string voltages/currents)
	// "a_u", "b_u", "c_u" (AC phase voltages)
	// "a_i", "b_i", "c_i" (AC phase currents)
	// "ab_u", "bc_u", "ca_u" (line voltages)
	// "elec_freq" (grid frequency)

	// ── PV ──
	mpptPower := extractFloat(data, "mppt_power") * 1000 // kW → W
	dayEnergy := extractFloatP(data, "day_cap")
	totalEnergy := extractFloatP(data, "total_cap")

	var pvStrings []models.PVString
	for i := 1; i <= 24; i++ {
		vKey := fmt.Sprintf("pv%d_u", i)
		cKey := fmt.Sprintf("pv%d_i", i)
		v := extractFloatP(data, vKey)
		c := extractFloatP(data, cKey)
		if v != nil || c != nil {
			var pw *float64
			if v != nil && c != nil {
				p := *v * *c
				pw = &p
			}
			pvStrings = append(pvStrings, models.PVString{ID: i, VoltageV: v, CurrentA: c, PowerW: pw})
		}
	}

	rt.PV = &models.PVData{
		TotalPowerW:     mpptPower,
		TodayEnergyKWh:  dayEnergy,
		TotalEnergyKWh:  totalEnergy,
		Strings:         pvStrings,
	}

	// ── Grid ──
	activePower := extractFloat(data, "active_power") * 1000 // kW → W
	freq := extractFloatP(data, "elec_freq")
	pf := extractFloatP(data, "power_factor")

	var phases []models.PhaseData
	phaseNames := []struct{ volt, curr, phase string }{
		{"a_u", "a_i", "A"},
		{"b_u", "b_i", "B"},
		{"c_u", "c_i", "C"},
	}
	for _, ph := range phaseNames {
		v := extractFloatP(data, ph.volt)
		c := extractFloatP(data, ph.curr)
		if v != nil || c != nil {
			phases = append(phases, models.PhaseData{
				Phase: ph.phase, VoltageV: v, CurrentA: c, FrequencyHz: freq,
			})
		}
	}

	rt.Grid = &models.GridData{
		TotalPowerW: activePower,
		Direction:   models.GridDirectionIdle,
		FrequencyHz: freq,
		PowerFactor: pf,
		Phases:      phases,
	}

	// ── Environment ──
	temp := extractFloatP(data, "temperature")
	rt.Environment = &models.EnvironmentData{
		InverterTemperatureC: temp,
	}

	return rt
}

func normalizeHuaweiStationEnergy(raw huaweiStationKpi, plantID string, period models.Period) models.NormalizedEnergy {
	data := raw.DataItemMap

	// Huawei station KPI keys:
	// "day_power" (daily yield kWh), "month_power" (monthly kWh),
	// "total_power" (lifetime kWh), "day_income" (daily revenue),
	// "real_health_state" (plant health: 1=disconnected, 2=faulty, 3=healthy)
	// "total_income", "real_power" (current power kW)

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

	switch period {
	case models.PeriodDay:
		energy.PVGenerationKWh = extractFloatP(data, "day_power")
	case models.PeriodMonth:
		energy.PVGenerationKWh = extractFloatP(data, "month_power")
	case models.PeriodTotal:
		energy.PVGenerationKWh = extractFloatP(data, "total_power")
	}

	// Current power (kW → W)
	if rp := extractFloatP(data, "real_power"); rp != nil {
		pw := *rp * 1000
		energy.CurrentPowerW = &pw
	}

	return energy
}

func normalizeHuaweiHistory(resp huaweiKpiHistoryResponse, deviceID string, req models.HistoryRequest) models.HistoryResponse {
	result := models.HistoryResponse{
		DeviceID:    fmt.Sprintf("%s_%s", providerName, deviceID),
		Provider:    providerName,
		Granularity: req.Granularity,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
	}

	for _, entry := range resp.Data {
		ts := time.UnixMilli(entry.CollectTime)
		data := entry.DataItemMap

		dp := models.NormalizedTimeSeries{
			DeviceID:    result.DeviceID,
			Provider:    providerName,
			Timestamp:   ts,
			Granularity: req.Granularity,
			Meta: models.ProviderMeta{
				Provider: providerName,
				FetchedAt: time.Now().UTC(),
			},
		}

		// Map Huawei historical KPIs
		dp.PVEnergyKWh = extractFloatP(data, "inverter_power")
		dp.GridExportEnergyKWh = extractFloatP(data, "ongrid_power")
		dp.LoadEnergyKWh = extractFloatP(data, "use_power")

		result.DataPoints = append(result.DataPoints, dp)
	}

	result.TotalPoints = len(result.DataPoints)
	return result
}

func normalizeHuaweiAlarm(raw huaweiAlarm) models.NormalizedAlarm {
	alarmID := strconv.FormatInt(raw.AlarmID, 10)
	startTime := time.UnixMilli(raw.RaiseTime)

	alarm := models.NormalizedAlarm{
		ID:                 fmt.Sprintf("%s_alarm_%s", providerName, alarmID),
		Provider:           providerName,
		DeviceID:           fmt.Sprintf("%s_%s", providerName, raw.DeviceESN),
		PlantID:            fmt.Sprintf("%s_%s", providerName, raw.StationCode),
		Code:               alarmID,
		Name:               raw.AlarmName,
		Severity:           huaweiAlarmSeverity(raw.Severity),
		Status:             huaweiAlarmStatus(raw.Status),
		DeviceSerialNumber: raw.DeviceESN,
		PlantName:          raw.StationName,
		StartTime:          startTime,
		Meta: models.ProviderMeta{
			Provider:         providerName,
			ProviderDeviceID: raw.DeviceESN,
			ProviderPlantID:  raw.StationCode,
			FetchedAt:        time.Now().UTC(),
		},
	}

	if raw.ClearTime > 0 {
		t := time.UnixMilli(raw.ClearTime)
		alarm.EndTime = &t
	}

	return alarm
}

// ══════════════════════════════════════════════════════════════════
// Mapping helpers
// ══════════════════════════════════════════════════════════════════

func huaweiDeviceType(typeID int) models.DeviceType {
	switch typeID {
	case 1:
		return models.DeviceTypeStringInverter
	case 2:
		return models.DeviceTypeMeter
	case 38:
		return models.DeviceTypeBattery
	case 39:
		return models.DeviceTypeOptimizer
	case 46:
		return models.DeviceTypeEMS
	case 47:
		return models.DeviceTypeGateway
	case 62:
		return models.DeviceTypeHybridInverter
	default:
		return models.DeviceTypeUnknown
	}
}

func huaweiAlarmSeverity(severity int) models.AlarmSeverity {
	switch severity {
	case 1:
		return models.AlarmSeverityCritical
	case 2:
		return models.AlarmSeverityCritical
	case 3:
		return models.AlarmSeverityWarning
	case 4:
		return models.AlarmSeverityInfo
	default:
		return models.AlarmSeverityUnknown
	}
}

func huaweiAlarmStatus(status int) models.AlarmStatus {
	switch status {
	case 1:
		return models.AlarmStatusActive
	case 2:
		return models.AlarmStatusResolved
	default:
		return models.AlarmStatusUnknown
	}
}

func huaweiKpiEndpoint(g models.Granularity) string {
	switch g {
	case models.GranularityMinute, models.GranularityHour:
		return "/getKpiStationHour"
	case models.GranularityDay:
		return "/getKpiStationDay"
	case models.GranularityMonth:
		return "/getKpiStationMonth"
	case models.GranularityYear:
		return "/getKpiStationYear"
	default:
		return "/getKpiStationDay"
	}
}

// ── Extraction helpers ──

func extractFloat(m map[string]interface{}, key string) float64 {
	if m == nil {
		return 0
	}
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(val), 64)
		return f
	default:
		return 0
	}
}

func extractFloatP(m map[string]interface{}, key string) *float64 {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	var f float64
	switch val := v.(type) {
	case float64:
		f = val
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
		if err != nil {
			return nil
		}
		f = parsed
	default:
		return nil
	}
	return &f
}

// Unused import suppressors
var _ = url.Values{}
