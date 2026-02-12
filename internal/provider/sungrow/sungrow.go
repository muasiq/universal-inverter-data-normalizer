package sungrow

import (
	"context"
	"crypto/md5"
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
	defaultBaseURL = "https://gateway.isolarcloud.com"
	providerName   = "sungrow"
)

func init() {
	provider.Register(providerName, func() provider.Provider {
		return &SungrowProvider{}
	})
}

// SungrowProvider implements the Provider interface for Sungrow iSolarCloud API.
// Sungrow uses an appKey + userAccount + userPassword (MD5-hashed) auth.
// Key endpoints:
//   - /openapi/getPowerStationList — list plants
//   - /openapi/getPowerStationDetail — plant details
//   - /openapi/getDeviceList — device list
//   - /openapi/queryDeviceRealTimeData — real-time data
//   - /openapi/queryDeviceHistoryData — historical data
//   - /openapi/getAlarmList — alarms
type SungrowProvider struct {
	client  *provider.HTTPClient
	config  provider.ProviderConfig
	appKey  string
	token   string
	userID  string
}

func (p *SungrowProvider) Name() string { return providerName }

func (p *SungrowProvider) Initialize(ctx context.Context, cfg provider.ProviderConfig) error {
	p.config = cfg

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	rps := cfg.RateLimitRPS
	if rps <= 0 {
		rps = 10
	}

	p.client = provider.NewHTTPClient(baseURL, cfg.TimeoutSeconds, rps)
	p.client.SetHeader("Content-Type", "application/json")
	p.appKey = cfg.GetCredential("appKey")

	return p.authenticate(ctx)
}

func (p *SungrowProvider) authenticate(ctx context.Context) error {
	account := p.config.GetCredential("userAccount")
	password := p.config.GetCredential("userPassword")

	// Sungrow requires MD5-hashed password
	hashedPwd := fmt.Sprintf("%x", md5.Sum([]byte(password)))

	body := map[string]interface{}{
		"appkey":       p.appKey,
		"user_account": account,
		"user_password": hashedPwd,
	}

	var resp sungrowLoginResponse
	if err := p.client.Post(ctx, "/openapi/login", body, &resp); err != nil {
		return fmt.Errorf("Sungrow auth: %w", err)
	}
	if resp.ResultCode != "1" {
		return fmt.Errorf("Sungrow auth failed: code=%s msg=%s", resp.ResultCode, resp.ResultMsg)
	}

	p.token = resp.ResultData.Token
	p.userID = resp.ResultData.UserID
	p.client.SetHeader("token", p.token)

	log.Info().Str("provider", providerName).Msg("Authenticated successfully")
	return nil
}

// ── Plants ──

func (p *SungrowProvider) GetPlants(ctx context.Context) ([]models.NormalizedPlant, error) {
	body := map[string]interface{}{
		"appkey":  p.appKey,
		"token":   p.token,
		"user_id": p.userID,
	}

	var resp sungrowPlantListResponse
	if err := p.client.Post(ctx, "/openapi/getPowerStationList", body, &resp); err != nil {
		return nil, fmt.Errorf("Sungrow GetPlants: %w", err)
	}
	if resp.ResultCode != "1" {
		return nil, fmt.Errorf("Sungrow GetPlants failed: %s", resp.ResultMsg)
	}

	var plants []models.NormalizedPlant
	for _, raw := range resp.ResultData.PageList {
		plants = append(plants, normalizeSungrowPlant(raw))
	}
	return plants, nil
}

func (p *SungrowProvider) GetPlantDetails(ctx context.Context, plantID string) (*models.NormalizedPlant, error) {
	body := map[string]interface{}{
		"appkey": p.appKey,
		"token":  p.token,
		"ps_id":  plantID,
	}

	var resp sungrowPlantDetailResponse
	if err := p.client.Post(ctx, "/openapi/getPowerStationDetail", body, &resp); err != nil {
		return nil, fmt.Errorf("Sungrow GetPlantDetails: %w", err)
	}

	plant := normalizeSungrowPlantDetail(resp.ResultData)
	return &plant, nil
}

// ── Devices ──

func (p *SungrowProvider) GetDevices(ctx context.Context, plantID string) ([]models.NormalizedDevice, error) {
	body := map[string]interface{}{
		"appkey": p.appKey,
		"token":  p.token,
		"ps_id":  plantID,
	}

	var resp sungrowDeviceListResponse
	if err := p.client.Post(ctx, "/openapi/getDeviceList", body, &resp); err != nil {
		return nil, fmt.Errorf("Sungrow GetDevices: %w", err)
	}
	if resp.ResultCode != "1" {
		return nil, fmt.Errorf("Sungrow GetDevices failed: %s", resp.ResultMsg)
	}

	var devices []models.NormalizedDevice
	for _, raw := range resp.ResultData.PageList {
		devices = append(devices, normalizeSungrowDevice(raw, plantID))
	}
	return devices, nil
}

func (p *SungrowProvider) GetDeviceDetails(ctx context.Context, deviceID string) (*models.NormalizedDevice, error) {
	return nil, fmt.Errorf("Sungrow: use GetDevices with ps_id for device details")
}

// ── Real-Time Data ──

func (p *SungrowProvider) GetRealTimeData(ctx context.Context, deviceID string) (*models.NormalizedRealtime, error) {
	body := map[string]interface{}{
		"appkey":    p.appKey,
		"token":     p.token,
		"device_id": deviceID,
	}

	var resp sungrowRealtimeResponse
	if err := p.client.Post(ctx, "/openapi/queryDeviceRealTimeData", body, &resp); err != nil {
		return nil, fmt.Errorf("Sungrow GetRealTimeData: %w", err)
	}

	rt := normalizeSungrowRealtime(resp.ResultData, deviceID)
	return &rt, nil
}

// ── Energy Stats ──

func (p *SungrowProvider) GetEnergyStats(ctx context.Context, plantID string, period models.Period) (*models.NormalizedEnergy, error) {
	body := map[string]interface{}{
		"appkey": p.appKey,
		"token":  p.token,
		"ps_id":  plantID,
	}

	var resp sungrowPlantDetailResponse
	if err := p.client.Post(ctx, "/openapi/getPowerStationDetail", body, &resp); err != nil {
		return nil, fmt.Errorf("Sungrow GetEnergyStats: %w", err)
	}

	energy := normalizeSungrowEnergy(resp.ResultData, plantID, period)
	return &energy, nil
}

// ── Historical Data ──

func (p *SungrowProvider) GetHistoricalData(ctx context.Context, deviceID string, req models.HistoryRequest) (*models.HistoryResponse, error) {
	timeType := sungrowTimeType(req.Granularity)

	body := map[string]interface{}{
		"appkey":     p.appKey,
		"token":      p.token,
		"device_id":  deviceID,
		"start_time": req.StartTime,
		"end_time":   req.EndTime,
		"time_type":  timeType,
	}

	var resp sungrowHistoryResponse
	if err := p.client.Post(ctx, "/openapi/queryDeviceHistoryData", body, &resp); err != nil {
		return nil, fmt.Errorf("Sungrow GetHistoricalData: %w", err)
	}

	history := normalizeSungrowHistory(resp, deviceID, req)
	return &history, nil
}

// ── Alarms ──

func (p *SungrowProvider) GetAlarms(ctx context.Context, deviceID string) ([]models.NormalizedAlarm, error) {
	body := map[string]interface{}{
		"appkey":    p.appKey,
		"token":     p.token,
		"device_id": deviceID,
	}

	var resp sungrowAlarmResponse
	if err := p.client.Post(ctx, "/openapi/getAlarmList", body, &resp); err != nil {
		return nil, fmt.Errorf("Sungrow GetAlarms: %w", err)
	}

	var alarms []models.NormalizedAlarm
	for _, raw := range resp.ResultData.PageList {
		alarms = append(alarms, normalizeSungrowAlarm(raw))
	}
	return alarms, nil
}

func (p *SungrowProvider) GetAllAlarms(ctx context.Context) ([]models.NormalizedAlarm, error) {
	body := map[string]interface{}{
		"appkey":  p.appKey,
		"token":   p.token,
		"user_id": p.userID,
	}

	var resp sungrowAlarmResponse
	if err := p.client.Post(ctx, "/openapi/getAlarmList", body, &resp); err != nil {
		return nil, fmt.Errorf("Sungrow GetAllAlarms: %w", err)
	}

	var alarms []models.NormalizedAlarm
	for _, raw := range resp.ResultData.PageList {
		alarms = append(alarms, normalizeSungrowAlarm(raw))
	}
	return alarms, nil
}

func (p *SungrowProvider) Healthy(ctx context.Context) bool {
	return p.token != ""
}

func (p *SungrowProvider) Close() error {
	return nil
}

// ══════════════════════════════════════════════════════════════════
// Sungrow raw response types
// ══════════════════════════════════════════════════════════════════

type sungrowBaseResponse struct {
	ResultCode string `json:"result_code"` // "1" = success
	ResultMsg  string `json:"result_msg"`
}

type sungrowLoginResponse struct {
	sungrowBaseResponse
	ResultData struct {
		Token  string `json:"token"`
		UserID string `json:"user_id"`
	} `json:"result_data"`
}

type sungrowPlantListResponse struct {
	sungrowBaseResponse
	ResultData struct {
		PageList []sungrowPlant `json:"pageList"`
	} `json:"result_data"`
}

type sungrowPlant struct {
	PSID          string  `json:"ps_id"`
	PSName        string  `json:"ps_name"`
	PSLocation    string  `json:"ps_location"`
	Latitude      string  `json:"latitude"`
	Longitude     string  `json:"longitude"`
	DesignCapacity string `json:"design_capacity"` // kWp
	PSType        int     `json:"ps_type"`
	ConnectType   int     `json:"connect_type"`
}

type sungrowPlantDetailResponse struct {
	sungrowBaseResponse
	ResultData map[string]interface{} `json:"result_data"`
}

type sungrowDeviceListResponse struct {
	sungrowBaseResponse
	ResultData struct {
		PageList []sungrowDevice `json:"pageList"`
	} `json:"result_data"`
}

type sungrowDevice struct {
	DeviceID     string `json:"device_id"`
	DeviceName   string `json:"device_name"`
	DeviceType   int    `json:"device_type"` // 1=inverter, 2=combiner, 3=meter, etc.
	DeviceModel  string `json:"device_model_code"`
	DeviceSN     string `json:"device_sn"`
	DeviceStatus int    `json:"device_status"` // 0=offline, 1=online
	PSID         string `json:"ps_id"`
}

type sungrowRealtimeResponse struct {
	sungrowBaseResponse
	ResultData map[string]interface{} `json:"result_data"`
}

type sungrowHistoryResponse struct {
	sungrowBaseResponse
	ResultData struct {
		TimeType string                   `json:"time_type"`
		DataList []map[string]interface{} `json:"data_list"`
	} `json:"result_data"`
}

type sungrowAlarmResponse struct {
	sungrowBaseResponse
	ResultData struct {
		PageList []sungrowAlarm `json:"pageList"`
	} `json:"result_data"`
}

type sungrowAlarm struct {
	AlarmID     string `json:"alarm_id"`
	AlarmName   string `json:"alarm_name"`
	AlarmLevel  int    `json:"alarm_level"` // 1=low, 2=medium, 3=high
	AlarmStatus int    `json:"alarm_status"` // 1=active, 2=recovered
	DeviceSN    string `json:"device_sn"`
	DeviceName  string `json:"device_name"`
	PSName      string `json:"ps_name"`
	PSID        string `json:"ps_id"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
}

// ══════════════════════════════════════════════════════════════════
// Normalization functions
// ══════════════════════════════════════════════════════════════════

func normalizeSungrowPlant(raw sungrowPlant) models.NormalizedPlant {
	plant := models.NormalizedPlant{
		ID:       fmt.Sprintf("%s_%s", providerName, raw.PSID),
		Provider: providerName,
		Name:     raw.PSName,
		Address:  raw.PSLocation,
		Meta: models.ProviderMeta{
			Provider:        providerName,
			ProviderPlantID: raw.PSID,
			FetchedAt:       time.Now().UTC(),
		},
	}

	if lat, err := strconv.ParseFloat(raw.Latitude, 64); err == nil {
		if lng, err := strconv.ParseFloat(raw.Longitude, 64); err == nil {
			plant.Location = &models.LatLng{Latitude: lat, Longitude: lng}
		}
	}

	if cap, err := strconv.ParseFloat(raw.DesignCapacity, 64); err == nil {
		plant.PeakPowerKWp = &cap
	}

	return plant
}

func normalizeSungrowPlantDetail(data map[string]interface{}) models.NormalizedPlant {
	return models.NormalizedPlant{
		ID:       fmt.Sprintf("%s_%s", providerName, extractStr(data, "ps_id")),
		Provider: providerName,
		Name:     extractStr(data, "ps_name"),
		Address:  extractStr(data, "ps_location"),
		Meta: models.ProviderMeta{
			Provider:        providerName,
			ProviderPlantID: extractStr(data, "ps_id"),
			RawDataAvailable: true,
			FetchedAt:       time.Now().UTC(),
		},
	}
}

func normalizeSungrowDevice(raw sungrowDevice, plantID string) models.NormalizedDevice {
	return models.NormalizedDevice{
		ID:           fmt.Sprintf("%s_%s", providerName, raw.DeviceID),
		Provider:     providerName,
		PlantID:      fmt.Sprintf("%s_%s", providerName, plantID),
		Name:         raw.DeviceName,
		SerialNumber: raw.DeviceSN,
		Model:        raw.DeviceModel,
		DeviceType:   sungrowDeviceType(raw.DeviceType),
		Manufacturer: "Sungrow",
		Status:       sungrowDeviceStatus(raw.DeviceStatus),
		IsOnline:     raw.DeviceStatus == 1,
		Meta: models.ProviderMeta{
			Provider:         providerName,
			ProviderDeviceID: raw.DeviceID,
			ProviderPlantID:  plantID,
			FetchedAt:        time.Now().UTC(),
			Extra: map[string]string{
				"deviceType": strconv.Itoa(raw.DeviceType),
			},
		},
	}
}

func normalizeSungrowRealtime(data map[string]interface{}, deviceID string) models.NormalizedRealtime {
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

	// Sungrow real-time keys:
	// "p1", "p2", ... (MPPT power W), "e_today" (daily kWh), "e_total" (lifetime kWh)
	// "pac" (AC output power W), "reactive_power",
	// "ua", "ub", "uc" (phase voltages), "ia", "ib", "ic" (phase currents)
	// "fac" (frequency), "pf" (power factor)
	// "soc" (battery SOC %), "battery_power", "grid_power"

	// ── PV ──
	totalPVPower := extractFl(data, "total_dc_power")
	if totalPVPower == 0 {
		totalPVPower = extractFl(data, "pac")
	}
	todayEnergy := extractFlP(data, "e_today")
	totalEnergy := extractFlP(data, "e_total")

	var pvStrings []models.PVString
	for i := 1; i <= 12; i++ {
		vKey := fmt.Sprintf("mppt_%d_cap_u", i)
		cKey := fmt.Sprintf("mppt_%d_cap_i", i)
		v := extractFlP(data, vKey)
		c := extractFlP(data, cKey)
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
		TotalPowerW:     totalPVPower,
		TodayEnergyKWh:  todayEnergy,
		TotalEnergyKWh:  totalEnergy,
		Strings:         pvStrings,
	}

	// ── Grid ──
	gridPower := extractFl(data, "meter_power")
	freq := extractFlP(data, "fac")
	pf := extractFlP(data, "pf")

	var phases []models.PhaseData
	for _, ph := range []struct{ v, c, p string }{{"ua", "ia", "A"}, {"ub", "ib", "B"}, {"uc", "ic", "C"}} {
		voltage := extractFlP(data, ph.v)
		current := extractFlP(data, ph.c)
		if voltage != nil || current != nil {
			phases = append(phases, models.PhaseData{
				Phase: ph.p, VoltageV: voltage, CurrentA: current, FrequencyHz: freq,
			})
		}
	}

	rt.Grid = &models.GridData{
		TotalPowerW: gridPower,
		Direction:   sungrowGridDirection(gridPower),
		FrequencyHz: freq,
		PowerFactor: pf,
		Phases:      phases,
	}

	// ── Battery ──
	soc := extractFlP(data, "soc")
	batPower := extractFl(data, "battery_power")
	rt.Battery = &models.BatteryData{
		SOCPercent: soc,
		PowerW:     batPower,
		Direction:  sungrowBatteryDirection(batPower),
	}

	// ── Load ──
	loadPower := extractFl(data, "load_power")
	rt.Load = &models.LoadData{
		TotalPowerW: loadPower,
	}

	return rt
}

func normalizeSungrowEnergy(data map[string]interface{}, plantID string, period models.Period) models.NormalizedEnergy {
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
		energy.PVGenerationKWh = extractFlP(data, "today_energy")
	case models.PeriodMonth:
		energy.PVGenerationKWh = extractFlP(data, "month_energy")
	case models.PeriodYear:
		energy.PVGenerationKWh = extractFlP(data, "year_energy")
	case models.PeriodTotal:
		energy.PVGenerationKWh = extractFlP(data, "total_energy")
	}

	energy.CurrentPowerW = extractFlP(data, "curr_power")

	return energy
}

func normalizeSungrowHistory(resp sungrowHistoryResponse, deviceID string, req models.HistoryRequest) models.HistoryResponse {
	result := models.HistoryResponse{
		DeviceID:    fmt.Sprintf("%s_%s", providerName, deviceID),
		Provider:    providerName,
		Granularity: req.Granularity,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
	}

	for _, point := range resp.ResultData.DataList {
		ts := time.Now()
		if dt, ok := point["time_stamp"].(string); ok {
			if t, err := time.Parse("20060102150405", dt); err == nil {
				ts = t
			}
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

		dp.PVPowerW = extractFlP(point, "pac")
		dp.PVEnergyKWh = extractFlP(point, "e_day")
		dp.GridPowerW = extractFlP(point, "meter_power")

		result.DataPoints = append(result.DataPoints, dp)
	}

	result.TotalPoints = len(result.DataPoints)
	return result
}

func normalizeSungrowAlarm(raw sungrowAlarm) models.NormalizedAlarm {
	ts := time.Now()
	if t, err := time.Parse("2006-01-02 15:04:05", raw.StartTime); err == nil {
		ts = t
	}

	alarm := models.NormalizedAlarm{
		ID:                 fmt.Sprintf("%s_alarm_%s", providerName, raw.AlarmID),
		Provider:           providerName,
		DeviceID:           fmt.Sprintf("%s_%s", providerName, raw.DeviceSN),
		PlantID:            fmt.Sprintf("%s_%s", providerName, raw.PSID),
		Code:               raw.AlarmID,
		Name:               raw.AlarmName,
		Severity:           sungrowAlarmSeverity(raw.AlarmLevel),
		Status:             sungrowAlarmStatus(raw.AlarmStatus),
		DeviceSerialNumber: raw.DeviceSN,
		PlantName:          raw.PSName,
		StartTime:          ts,
		Meta: models.ProviderMeta{
			Provider:         providerName,
			ProviderDeviceID: raw.DeviceSN,
			ProviderPlantID:  raw.PSID,
			FetchedAt:        time.Now().UTC(),
		},
	}

	if raw.EndTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", raw.EndTime); err == nil {
			alarm.EndTime = &t
		}
	}

	return alarm
}

// ══════════════════════════════════════════════════════════════════
// Mapping helpers
// ══════════════════════════════════════════════════════════════════

func sungrowDeviceType(t int) models.DeviceType {
	switch t {
	case 1:
		return models.DeviceTypeStringInverter
	case 2:
		return models.DeviceTypeOptimizer // Combiner box → closest
	case 3:
		return models.DeviceTypeMeter
	case 4:
		return models.DeviceTypeWeatherStation
	case 5:
		return models.DeviceTypeGateway
	case 7:
		return models.DeviceTypeBattery
	case 11:
		return models.DeviceTypeHybridInverter
	case 14:
		return models.DeviceTypeEMS
	default:
		return models.DeviceTypeUnknown
	}
}

func sungrowDeviceStatus(s int) models.DeviceStatus {
	switch s {
	case 0:
		return models.DeviceStatusOffline
	case 1:
		return models.DeviceStatusOnline
	default:
		return models.DeviceStatusUnknown
	}
}

func sungrowGridDirection(power float64) models.GridDirection {
	if power > 0 {
		return models.GridDirectionImporting
	} else if power < 0 {
		return models.GridDirectionExporting
	}
	return models.GridDirectionIdle
}

func sungrowBatteryDirection(power float64) models.EnergyDirection {
	if power > 0 {
		return models.DirectionCharging
	} else if power < 0 {
		return models.DirectionDischarging
	}
	return models.DirectionIdle
}

func sungrowAlarmSeverity(level int) models.AlarmSeverity {
	switch level {
	case 1:
		return models.AlarmSeverityInfo
	case 2:
		return models.AlarmSeverityWarning
	case 3:
		return models.AlarmSeverityCritical
	default:
		return models.AlarmSeverityUnknown
	}
}

func sungrowAlarmStatus(status int) models.AlarmStatus {
	switch status {
	case 1:
		return models.AlarmStatusActive
	case 2:
		return models.AlarmStatusResolved
	default:
		return models.AlarmStatusUnknown
	}
}

func sungrowTimeType(g models.Granularity) int {
	switch g {
	case models.GranularityMinute:
		return 1
	case models.GranularityHour:
		return 1
	case models.GranularityDay:
		return 2
	case models.GranularityMonth:
		return 3
	case models.GranularityYear:
		return 4
	default:
		return 2
	}
}

// ── Extraction helpers ──

func extractStr(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func extractFl(m map[string]interface{}, key string) float64 {
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

func extractFlP(m map[string]interface{}, key string) *float64 {
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
