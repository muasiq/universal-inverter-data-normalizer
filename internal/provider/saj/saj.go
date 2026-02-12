package saj

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
	defaultBaseURL = "https://intl-developer.saj-electric.com/prod-api"
	providerName   = "saj"
)

func init() {
	provider.Register(providerName, func() provider.Provider {
		return &SAJProvider{}
	})
}

// SAJProvider implements the Provider interface for SAJ Elekeeper Open Platform.
type SAJProvider struct {
	client   *provider.HTTPClient
	config   provider.ProviderConfig
	appID    string
	token    string
	tokenExp time.Time
}

func (p *SAJProvider) Name() string { return providerName }

func (p *SAJProvider) Initialize(ctx context.Context, cfg provider.ProviderConfig) error {
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
	p.client.SetHeader("content-language", "en_US")
	p.appID = cfg.GetCredential("appId")

	return p.authenticate(ctx)
}

func (p *SAJProvider) authenticate(ctx context.Context) error {
	appID := p.config.GetCredential("appId")
	appSecret := p.config.GetCredential("appSecret")

	params := url.Values{
		"appId":     {appID},
		"appSecret": {appSecret},
	}

	var resp sajTokenResponse
	if err := p.client.Get(ctx, "/open/api/access_token", params, &resp); err != nil {
		return fmt.Errorf("SAJ auth: %w", err)
	}
	if resp.Code != 200 {
		return fmt.Errorf("SAJ auth failed: code=%d msg=%s", resp.Code, resp.Msg)
	}

	p.token = resp.Data.AccessToken
	p.tokenExp = time.Now().Add(time.Duration(resp.Data.Expires) * time.Second)
	p.client.SetHeader("accessToken", p.token)

	log.Info().Str("provider", providerName).Msg("Authenticated successfully")
	return nil
}

func (p *SAJProvider) ensureToken(ctx context.Context) error {
	if time.Now().After(p.tokenExp.Add(-5 * time.Minute)) {
		return p.authenticate(ctx)
	}
	return nil
}

// ── Plants ──

func (p *SAJProvider) GetPlants(ctx context.Context) ([]models.NormalizedPlant, error) {
	if err := p.ensureToken(ctx); err != nil {
		return nil, err
	}

	var allPlants []models.NormalizedPlant
	page := 1
	for {
		params := url.Values{
			"appId":    {p.appID},
			"pageSize": {"100"},
			"pageNum":  {strconv.Itoa(page)},
		}

		var resp sajPlantListResponse
		if err := p.client.Get(ctx, "/open/api/developer/plant/page", params, &resp); err != nil {
			return nil, err
		}
		if resp.Code != 200 {
			return nil, fmt.Errorf("SAJ GetPlants: code=%d msg=%s", resp.Code, resp.Msg)
		}

		for _, raw := range resp.Rows {
			allPlants = append(allPlants, normalizeSAJPlant(raw))
		}

		if page >= int(resp.TotalPage) || len(resp.Rows) == 0 {
			break
		}
		page++
	}

	return allPlants, nil
}

func (p *SAJProvider) GetPlantDetails(ctx context.Context, plantID string) (*models.NormalizedPlant, error) {
	if err := p.ensureToken(ctx); err != nil {
		return nil, err
	}

	params := url.Values{"plantId": {plantID}}
	var resp sajPlantDetailsResponse
	if err := p.client.Get(ctx, "/open/api/plant/details", params, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, fmt.Errorf("SAJ GetPlantDetails: code=%d msg=%s", resp.Code, resp.Msg)
	}

	plant := normalizeSAJPlantDetails(resp.Data)
	return &plant, nil
}

// ── Devices ──

func (p *SAJProvider) GetDevices(ctx context.Context, plantID string) ([]models.NormalizedDevice, error) {
	if err := p.ensureToken(ctx); err != nil {
		return nil, err
	}

	params := url.Values{
		"plantId": {plantID},
		"userId":  {""},
	}

	var resp sajDeviceListResponse
	if err := p.client.Get(ctx, "/open/api/plant/getPlantAllDeviceList", params, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, fmt.Errorf("SAJ GetDevices: code=%d msg=%s", resp.Code, resp.Msg)
	}

	var devices []models.NormalizedDevice
	for _, raw := range resp.Data {
		dev := normalizeSAJDevice(raw, plantID)
		devices = append(devices, dev)
	}
	return devices, nil
}

func (p *SAJProvider) GetDeviceDetails(ctx context.Context, deviceID string) (*models.NormalizedDevice, error) {
	if err := p.ensureToken(ctx); err != nil {
		return nil, err
	}

	params := url.Values{"deviceSn": {deviceID}}
	var resp sajBaseInfoResponse
	if err := p.client.Get(ctx, "/open/api/device/baseinfo", params, &resp); err != nil {
		return nil, err
	}

	dev := normalizeSAJBaseInfo(resp.Data, deviceID)
	return &dev, nil
}

// ── Real-Time Data ──

func (p *SAJProvider) GetRealTimeData(ctx context.Context, deviceID string) (*models.NormalizedRealtime, error) {
	if err := p.ensureToken(ctx); err != nil {
		return nil, err
	}

	params := url.Values{"deviceSn": {deviceID}}
	var resp sajRealtimeResponse
	if err := p.client.Get(ctx, "/open/api/device/realtimeDataCommon", params, &resp); err != nil {
		return nil, err
	}

	rt := normalizeSAJRealtime(resp.Data)
	return &rt, nil
}

// ── Energy Stats ──

func (p *SAJProvider) GetEnergyStats(ctx context.Context, plantID string, period models.Period) (*models.NormalizedEnergy, error) {
	if err := p.ensureToken(ctx); err != nil {
		return nil, err
	}

	params := url.Values{
		"plantId":    {plantID},
		"clientDate": {time.Now().Format("2006-01-02 15:04:05")},
	}

	var resp sajPlantStatsResponse
	if err := p.client.Get(ctx, "/open/api/plant/getPlantStatisticsData", params, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, fmt.Errorf("SAJ GetEnergyStats: code=%d msg=%s", resp.Code, resp.Msg)
	}

	energy := normalizeSAJPlantStats(resp.Data, plantID, period)
	return &energy, nil
}

// ── Historical Data ──

func (p *SAJProvider) GetHistoricalData(ctx context.Context, deviceID string, req models.HistoryRequest) (*models.HistoryResponse, error) {
	if err := p.ensureToken(ctx); err != nil {
		return nil, err
	}

	timeUnit := granularityToSAJTimeUnit(req.Granularity)

	params := url.Values{
		"deviceSn":  {deviceID},
		"startTime": {req.StartTime},
		"endTime":   {req.EndTime},
		"timeUnit":  {strconv.Itoa(timeUnit)},
	}

	var resp sajUploadDataResponse
	if err := p.client.Get(ctx, "/open/api/device/uploadData", params, &resp); err != nil {
		return nil, err
	}

	history := normalizeSAJHistory(resp, deviceID, req)
	return &history, nil
}

// ── Alarms ──

func (p *SAJProvider) GetAlarms(ctx context.Context, deviceID string) ([]models.NormalizedAlarm, error) {
	if err := p.ensureToken(ctx); err != nil {
		return nil, err
	}

	params := url.Values{
		"deviceSn": {deviceID},
		"status":   {"1,4"},
	}

	var resp sajAlarmResponse
	if err := p.client.Get(ctx, "/open/api/device/alarmList", params, &resp); err != nil {
		return nil, err
	}

	var alarms []models.NormalizedAlarm
	for _, raw := range resp.Data {
		alarms = append(alarms, normalizeSAJAlarm(raw))
	}
	return alarms, nil
}

func (p *SAJProvider) GetAllAlarms(ctx context.Context) ([]models.NormalizedAlarm, error) {
	if err := p.ensureToken(ctx); err != nil {
		return nil, err
	}

	params := url.Values{
		"appId":  {p.appID},
		"status": {"1,4"},
	}

	var resp sajAlarmResponse
	if err := p.client.Get(ctx, "/open/api/device/alarmList", params, &resp); err != nil {
		return nil, err
	}

	var alarms []models.NormalizedAlarm
	for _, raw := range resp.Data {
		alarms = append(alarms, normalizeSAJAlarm(raw))
	}
	return alarms, nil
}

func (p *SAJProvider) Healthy(ctx context.Context) bool {
	return p.token != "" && time.Now().Before(p.tokenExp)
}

func (p *SAJProvider) Close() error {
	return nil
}

// ══════════════════════════════════════════════════════════════════
// Raw API response types (SAJ-specific)
// ══════════════════════════════════════════════════════════════════

type sajTokenResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		AccessToken string `json:"access_token"`
		Expires     int    `json:"expires"`
	} `json:"data"`
}

type sajPlantListResponse struct {
	Code      int              `json:"code"`
	Msg       string           `json:"msg"`
	Total     int64            `json:"total"`
	TotalPage int64            `json:"totalPage"`
	Rows      []sajPlantRow    `json:"rows"`
}

type sajPlantRow struct {
	PlantID   string `json:"plantId"`
	PlantName string `json:"plantName"`
	NMI       string `json:"NMI"`
	PlantNo   string `json:"plantNo"`
}

type sajPlantDetailsResponse struct {
	Code int              `json:"code"`
	Msg  string           `json:"msg"`
	Data sajPlantDetails  `json:"data"`
}

type sajPlantDetails struct {
	PlantUid       string  `json:"plantUid"`
	PlantName      string  `json:"plantName"`
	PlantNo        string  `json:"plantNo"`
	Type           int     `json:"type"`
	SystemPower    float64 `json:"systemPower"`
	GridPrice      float64 `json:"gridPrice"`
	CurrencyName   string  `json:"currencyName"`
	Currency       string  `json:"currency"`
	TimeZone       string  `json:"timeZone"`
	Country        string  `json:"country"`
	CountryCode    string  `json:"countryCode"`
	Province       string  `json:"province"`
	City           string  `json:"city"`
	FullAddress    string  `json:"fullAddress"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	GridNetType    string  `json:"gridNetType"`
	DeviceSnList   []string `json:"deviceSnList"`
	ModuleSnList   []string `json:"moduleSnList"`
}

type sajDeviceListResponse struct {
	Code int                `json:"code"`
	Msg  string             `json:"msg"`
	Data []sajDeviceWrapper `json:"data"`
}

type sajDeviceWrapper struct {
	DeviceType   int                    `json:"deviceType"`
	SN           string                 `json:"sn"`
	InverterData map[string]interface{} `json:"inverterData,omitempty"`
	EmsData      map[string]interface{} `json:"emsModuleData,omitempty"`
	MeterData    map[string]interface{} `json:"electricMeterData,omitempty"`
	ChargerData  map[string]interface{} `json:"chargerData,omitempty"`
}

type sajBaseInfoResponse struct {
	Code int                    `json:"code"`
	Msg  string                 `json:"msg"`
	Data map[string]interface{} `json:"data"`
}

type sajRealtimeResponse struct {
	ErrCode string                 `json:"errCode"`
	ErrMsg  string                 `json:"errMsg"`
	Data    map[string]interface{} `json:"data"`
}

type sajPlantStatsResponse struct {
	Code int                    `json:"code"`
	Msg  string                 `json:"msg"`
	Data map[string]interface{} `json:"data"`
}

type sajUploadDataResponse struct {
	ErrCode int    `json:"errCode"`
	ErrMsg  string `json:"errMsg"`
	Data    struct {
		DeviceType int                      `json:"deviceType"`
		TimeUnit   int                      `json:"timeUnit"`
		Total      map[string]interface{}   `json:"total"`
		Data       []map[string]interface{} `json:"data"`
	} `json:"data"`
}

type sajAlarmResponse struct {
	Code int           `json:"code"`
	Msg  string        `json:"msg"`
	Data []sajAlarmRaw `json:"data"`
}

type sajAlarmRaw struct {
	DeviceSn   string `json:"deviceSn"`
	AlarmCode  int    `json:"alarmCode"`
	AlarmName  string `json:"alarmName"`
	AlarmLevel int    `json:"alarmLevel"`
	Status     int    `json:"status"`
	AlarmTime  string `json:"alarmTime"`
}

// ══════════════════════════════════════════════════════════════════
// Normalization functions
// ══════════════════════════════════════════════════════════════════

func normalizeSAJPlant(raw sajPlantRow) models.NormalizedPlant {
	return models.NormalizedPlant{
		ID:       fmt.Sprintf("%s_%s", providerName, raw.PlantID),
		Provider: providerName,
		Name:     raw.PlantName,
		Meta: models.ProviderMeta{
			Provider:        providerName,
			ProviderPlantID: raw.PlantID,
			FetchedAt:       time.Now().UTC(),
			Extra: map[string]string{
				"plantNo": raw.PlantNo,
				"nmi":     raw.NMI,
			},
		},
	}
}

func normalizeSAJPlantDetails(raw sajPlantDetails) models.NormalizedPlant {
	peakPower := raw.SystemPower
	plantType := sajPlantType(raw.Type)
	gridConn := sajGridConnectionType(raw.GridNetType)

	return models.NormalizedPlant{
		ID:           fmt.Sprintf("%s_%s", providerName, raw.PlantUid),
		Provider:     providerName,
		Name:         raw.PlantName,
		Timezone:     raw.TimeZone,
		Country:      raw.Country,
		Address:      raw.FullAddress,
		PeakPowerKWp: &peakPower,
		PlantType:    plantType,
		GridConnectionType: &gridConn,
		ElectricityPrice:   &raw.GridPrice,
		Currency:     raw.Currency,
		Location: &models.LatLng{
			Latitude:  raw.Latitude,
			Longitude: raw.Longitude,
		},
		Meta: models.ProviderMeta{
			Provider:         providerName,
			ProviderPlantID:  raw.PlantUid,
			ProviderPlantUID: raw.PlantUid,
			FetchedAt:        time.Now().UTC(),
			Extra: map[string]string{
				"plantNo":     raw.PlantNo,
				"countryCode": raw.CountryCode,
			},
		},
	}
}

func normalizeSAJDevice(raw sajDeviceWrapper, plantID string) models.NormalizedDevice {
	deviceType := sajDeviceType(raw.DeviceType)
	sn := extractStringSafe(raw.InverterData, "deviceSn")
	if sn == "" {
		sn = extractStringSafe(raw.EmsData, "deviceSn")
	}
	if sn == "" {
		sn = extractStringSafe(raw.MeterData, "meterSn")
	}
	if sn == "" {
		sn = raw.SN
	}

	name := extractStringSafe(raw.InverterData, "aliases")
	if name == "" {
		name = extractStringSafe(raw.EmsData, "emsModuleName")
	}
	if name == "" {
		name = sn
	}

	model := extractStringSafe(raw.InverterData, "deviceModel")
	if model == "" {
		model = extractStringSafe(raw.EmsData, "emsModel")
	}

	return models.NormalizedDevice{
		ID:           fmt.Sprintf("%s_%s", providerName, sn),
		Provider:     providerName,
		PlantID:      fmt.Sprintf("%s_%s", providerName, plantID),
		Name:         name,
		SerialNumber: sn,
		Model:        model,
		DeviceType:   deviceType,
		Manufacturer: "SAJ",
		Status:       models.DeviceStatusUnknown,
		Meta: models.ProviderMeta{
			Provider:         providerName,
			ProviderDeviceID: sn,
			ProviderPlantID:  plantID,
			FetchedAt:        time.Now().UTC(),
		},
	}
}

func normalizeSAJBaseInfo(raw map[string]interface{}, deviceID string) models.NormalizedDevice {
	return models.NormalizedDevice{
		ID:           fmt.Sprintf("%s_%s", providerName, deviceID),
		Provider:     providerName,
		SerialNumber: extractStringSafe(raw, "invSN"),
		Model:        extractStringSafe(raw, "invType"),
		DeviceType:   models.DeviceTypeInverter,
		Manufacturer: "SAJ",
		FirmwareInfo: &models.FirmwareInfo{
			MainVersion:    extractStringSafe(raw, "invMFW"),
			SlaveVersion:   extractStringSafe(raw, "invSFW"),
			DisplayVersion: extractStringSafe(raw, "invDFW"),
			ModuleModel:    extractStringSafe(raw, "moduleModel"),
			ModuleSN:       extractStringSafe(raw, "moduleSN"),
			ModuleFirmware: extractStringSafe(raw, "moduleFW"),
		},
		Meta: models.ProviderMeta{
			Provider:         providerName,
			ProviderDeviceID: deviceID,
			RawDataAvailable: true,
			FetchedAt:        time.Now().UTC(),
		},
	}
}

func normalizeSAJRealtime(raw map[string]interface{}) models.NormalizedRealtime {
	now := time.Now().UTC()

	// Parse timestamp
	dataTime := extractStringSafe(raw, "dataTime")
	ts := now
	if t, err := time.Parse("2006-01-02 15:04:05", dataTime); err == nil {
		ts = t
	}

	deviceSn := extractStringSafe(raw, "deviceSn")

	// Operating mode
	mpvMode := extractIntSafe(raw, "mpvMode")
	opMode := sajOperatingMode(mpvMode)

	// Online status
	isOnline := extractStringSafe(raw, "isOnline") == "1"
	status := models.DeviceStatusOffline
	if isOnline {
		status = models.DeviceStatusOnline
	}

	rt := models.NormalizedRealtime{
		DeviceID:          fmt.Sprintf("%s_%s", providerName, deviceSn),
		Provider:          providerName,
		Timestamp:         ts,
		OriginalTimestamp:  dataTime,
		Status:            status,
		OperatingMode:     opMode,
		Meta: models.ProviderMeta{
			Provider:         providerName,
			ProviderDeviceID: deviceSn,
			RawDataAvailable: true,
			FetchedAt:        now,
		},
	}

	// ── PV ──
	totalPV := extractFloatSafe(raw, "totalPVPower")
	todayPV := extractFloatPtr(raw, "todayPvEnergy")
	totalPVEnergy := extractFloatPtr(raw, "totalPvEnergy")

	var pvStrings []models.PVString
	for i := 1; i <= 16; i++ {
		vKey := fmt.Sprintf("pv%dvolt", i)
		cKey := fmt.Sprintf("pv%dcurr", i)
		pKey := fmt.Sprintf("pv%dpower", i)
		v := extractFloatPtr(raw, vKey)
		c := extractFloatPtr(raw, cKey)
		pw := extractFloatPtr(raw, pKey)
		if v != nil || c != nil || pw != nil {
			pvStrings = append(pvStrings, models.PVString{
				ID: i, VoltageV: v, CurrentA: c, PowerW: pw,
			})
		}
	}

	rt.PV = &models.PVData{
		TotalPowerW:    totalPV,
		TodayEnergyKWh: todayPV,
		TotalEnergyKWh: totalPVEnergy,
		Strings:        pvStrings,
	}

	// ── Battery ──
	batSOC := extractFloatPtr(raw, "batEnergyPercent")
	batPower := extractFloatSafe(raw, "totalBatteryPower")
	batDir := sajBatteryDirection(extractIntSafe(raw, "batteryDirection"))
	batTemp := extractFloatPtr(raw, "batTempC")
	todayChg := extractFloatPtr(raw, "todayBatChgEnergy")
	todayDis := extractFloatPtr(raw, "todayBatDisEnergy")
	totalChg := extractFloatPtr(raw, "totalBatChgEnergy")
	totalDis := extractFloatPtr(raw, "totalBatDisEnergy")

	rt.Battery = &models.BatteryData{
		SOCPercent:         batSOC,
		PowerW:             batPower,
		Direction:          batDir,
		TemperatureC:       batTemp,
		TodayChargeKWh:    todayChg,
		TodayDischargeKWh: todayDis,
		TotalChargeKWh:    totalChg,
		TotalDischargeKWh: totalDis,
	}

	// ── Grid ──
	gridPower := extractFloatSafe(raw, "totalGridPowerWatt")
	if gridPower == 0 {
		gridPower = extractFloatSafe(raw, "sysGridPowerWatt")
	}
	gridDir := sajGridDirection(extractIntSafe(raw, "gridDirection"))
	todayExport := extractFloatPtr(raw, "todaySellEnergy")
	todayImport := extractFloatPtr(raw, "todayFeedInEnergy")
	totalExport := extractFloatPtr(raw, "totalSellEnergy")
	totalImport := extractFloatPtr(raw, "totalFeedInEnergy")

	var gridPhases []models.PhaseData
	phaseNames := []string{"r", "s", "t"}
	normalizedNames := []string{"A", "B", "C"}
	for i, prefix := range phaseNames {
		v := extractFloatPtr(raw, prefix+"GridVolt")
		c := extractFloatPtr(raw, prefix+"GridCurr")
		pw := extractFloatPtr(raw, prefix+"GridPowerWatt")
		f := extractFloatPtr(raw, prefix+"GridFreq")
		if v != nil || c != nil || pw != nil {
			gridPhases = append(gridPhases, models.PhaseData{
				Phase: normalizedNames[i], VoltageV: v, CurrentA: c, PowerW: pw, FrequencyHz: f,
			})
		}
	}

	rt.Grid = &models.GridData{
		TotalPowerW:    gridPower,
		Direction:      gridDir,
		TodayImportKWh: todayImport,
		TodayExportKWh: todayExport,
		TotalImportKWh: totalImport,
		TotalExportKWh: totalExport,
		Phases:         gridPhases,
	}

	// ── Load ──
	loadPower := extractFloatSafe(raw, "totalLoadPowerWatt")
	if loadPower == 0 {
		loadPower = extractFloatSafe(raw, "sysTotalLoadWatt")
	}
	todayLoad := extractFloatPtr(raw, "todayLoadEnergy")
	totalLoad := extractFloatPtr(raw, "totalTotalLoadEnergy")

	rt.Load = &models.LoadData{
		TotalPowerW:    loadPower,
		TodayEnergyKWh: todayLoad,
		TotalEnergyKWh: totalLoad,
	}

	// ── Environment ──
	sinkTemp := extractFloatPtr(raw, "sinkTempC")
	ambTemp := extractFloatPtr(raw, "ambTempC")
	invTemp := extractFloatPtr(raw, "invTempC")
	signal := extractIntPtr(raw, "linkSignal")

	rt.Environment = &models.EnvironmentData{
		SinkTemperatureC:     sinkTemp,
		AmbientTemperatureC:  ambTemp,
		InverterTemperatureC: invTemp,
		SignalStrengthDBm:    signal,
	}

	return rt
}

func normalizeSAJPlantStats(raw map[string]interface{}, plantID string, period models.Period) models.NormalizedEnergy {
	now := time.Now().UTC()

	energy := models.NormalizedEnergy{
		ID:        fmt.Sprintf("%s_%s", providerName, plantID),
		Provider:  providerName,
		Period:    period,
		Timestamp: now,
		Meta: models.ProviderMeta{
			Provider:        providerName,
			ProviderPlantID: plantID,
			FetchedAt:       now,
		},
	}

	// Map based on period
	switch period {
	case models.PeriodDay:
		energy.PVGenerationKWh = extractFloatPtr(raw, "todayPvEnergy")
		energy.LoadConsumptionKWh = extractFloatPtr(raw, "todayLoadEnergy")
		energy.GridImportKWh = extractFloatPtr(raw, "todayBuyEnergy")
		energy.GridExportKWh = extractFloatPtr(raw, "todaySellEnergy")
		energy.BatteryChargeKWh = extractFloatPtr(raw, "todayChargeEnergy")
		energy.BatteryDischargeKWh = extractFloatPtr(raw, "todayDisChargeEnergy")
	case models.PeriodMonth:
		energy.PVGenerationKWh = extractFloatPtr(raw, "monthPvEnergy")
		energy.LoadConsumptionKWh = extractFloatPtr(raw, "monthLoadEnergy")
		energy.GridImportKWh = extractFloatPtr(raw, "monthBuyEnergy")
		energy.GridExportKWh = extractFloatPtr(raw, "monthSellEnergy")
	case models.PeriodYear:
		energy.PVGenerationKWh = extractFloatPtr(raw, "yearPvEnergy")
		energy.LoadConsumptionKWh = extractFloatPtr(raw, "yearLoadEnergy")
		energy.GridImportKWh = extractFloatPtr(raw, "yearBuyEnergy")
		energy.GridExportKWh = extractFloatPtr(raw, "yearSellEnergy")
	case models.PeriodTotal:
		energy.PVGenerationKWh = extractFloatPtr(raw, "totalPvEnergy")
		energy.LoadConsumptionKWh = extractFloatPtr(raw, "totalLoadEnergy")
		energy.GridImportKWh = extractFloatPtr(raw, "totalBuyEnergy")
		energy.GridExportKWh = extractFloatPtr(raw, "totalSellEnergy")
		energy.BatteryChargeKWh = extractFloatPtr(raw, "totalChargeEnergy")
		energy.BatteryDischargeKWh = extractFloatPtr(raw, "totalDisChargeEnergy")
	}

	energy.CurrentPowerW = extractFloatPtr(raw, "powerNow")
	energy.BatterySOC = extractFloatPtr(raw, "batEnergyPercent")
	energy.CO2SavedKg = extractFloatPtr(raw, "totalReduceCo2")
	energy.TreesEquivalent = extractFloatPtr(raw, "totalPlantTreeNum")

	return energy
}

func normalizeSAJHistory(resp sajUploadDataResponse, deviceID string, req models.HistoryRequest) models.HistoryResponse {
	result := models.HistoryResponse{
		DeviceID:    fmt.Sprintf("%s_%s", providerName, deviceID),
		Provider:    providerName,
		Granularity: req.Granularity,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
	}

	for _, point := range resp.Data.Data {
		ts := time.Now()
		if dt, ok := point["dataTime"].(string); ok {
			if t, err := time.Parse("2006-01-02 15:04:05", dt); err == nil {
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

		// Minute data → power snapshots
		if req.Granularity == models.GranularityMinute {
			dp.PVPowerW = extractFloatPtr(point, "pvPower")
			dp.LoadPowerW = extractFloatPtr(point, "loadPower")
			dp.GridImportPowerW = extractFloatPtr(point, "buyPower")
			dp.GridExportPowerW = extractFloatPtr(point, "sellPower")
			dp.SelfUsePowerW = extractFloatPtr(point, "selfUsePower")
			dp.BatterySOC = extractFloatPtr(point, "batterySOC")

			chg := extractFloatPtr(point, "chargePower")
			dis := extractFloatPtr(point, "dischargePower")
			if chg != nil {
				dp.BatteryPowerW = chg
			} else if dis != nil {
				dp.BatteryPowerW = dis
			}
		} else {
			// Day/month/year → energy aggregates
			dp.PVEnergyKWh = extractFloatPtr(point, "pVEnergy")
			if dp.PVEnergyKWh == nil {
				dp.PVEnergyKWh = extractFloatPtr(point, "pvEnergy")
			}
			dp.LoadEnergyKWh = extractFloatPtr(point, "loadEnergy")
			dp.GridImportEnergyKWh = extractFloatPtr(point, "buyEnergy")
			dp.GridExportEnergyKWh = extractFloatPtr(point, "sellEnergy")
			dp.SelfConsumptionKWh = extractFloatPtr(point, "selfConsumption")
		}

		result.DataPoints = append(result.DataPoints, dp)
	}

	// Aggregates from total
	if resp.Data.Total != nil {
		result.Aggregate = &models.TimeSeriesAggregate{
			TotalPVEnergyKWh:     extractFloatPtr(resp.Data.Total, "totalPVEnergy"),
			TotalLoadEnergyKWh:   extractFloatPtr(resp.Data.Total, "totalLoad"),
			TotalGridImportKWh:   extractFloatPtr(resp.Data.Total, "totalBuyEnergy"),
			TotalGridExportKWh:   extractFloatPtr(resp.Data.Total, "totalSellEnergy"),
			SelfConsumptionRate:  extractFloatPtr(resp.Data.Total, "pvSelfConsumedRate"),
		}
	}

	result.TotalPoints = len(result.DataPoints)
	return result
}

func normalizeSAJAlarm(raw sajAlarmRaw) models.NormalizedAlarm {
	ts := time.Now()
	if t, err := time.Parse("2006-01-02 15:04:05", raw.AlarmTime); err == nil {
		ts = t
	}

	return models.NormalizedAlarm{
		ID:                 fmt.Sprintf("%s_alarm_%s_%d", providerName, raw.DeviceSn, raw.AlarmCode),
		Provider:           providerName,
		DeviceID:           fmt.Sprintf("%s_%s", providerName, raw.DeviceSn),
		Code:               strconv.Itoa(raw.AlarmCode),
		Name:               raw.AlarmName,
		Severity:           sajAlarmSeverity(raw.AlarmLevel),
		Status:             sajAlarmStatus(raw.Status),
		DeviceSerialNumber: raw.DeviceSn,
		StartTime:          ts,
		Meta: models.ProviderMeta{
			Provider:         providerName,
			ProviderDeviceID: raw.DeviceSn,
			FetchedAt:        time.Now().UTC(),
		},
	}
}

// ══════════════════════════════════════════════════════════════════
// Mapping helpers
// ══════════════════════════════════════════════════════════════════

func sajPlantType(t int) models.PlantType {
	switch t {
	case 0:
		return models.PlantTypeGridTied
	case 1:
		return models.PlantTypeHybrid
	case 3:
		return models.PlantTypeACCoupled
	default:
		return models.PlantTypeUnknown
	}
}

func sajGridConnectionType(t string) models.GridConnectionType {
	switch t {
	case "1":
		return models.GridConnectionFullExport
	case "2":
		return models.GridConnectionSelfConsume
	case "3":
		return models.GridConnectionOffGrid
	default:
		return models.GridConnectionUnknown
	}
}

func sajDeviceType(t int) models.DeviceType {
	switch t {
	case 0:
		return models.DeviceTypeStringInverter
	case 1:
		return models.DeviceTypeHybridInverter
	case 2:
		return models.DeviceTypeLoadMonitor
	case 3:
		return models.DeviceTypeEVCharger
	case 6:
		return models.DeviceTypeEMS
	case 7:
		return models.DeviceTypeMeter
	case 10:
		return models.DeviceTypeDieselGenerator
	case 16, 18:
		return models.DeviceTypeOptimizer
	case 17:
		return models.DeviceTypeHeatPump
	case 19:
		return models.DeviceTypeWeatherStation
	default:
		return models.DeviceTypeUnknown
	}
}

func sajOperatingMode(mode int) models.OperatingMode {
	switch mode {
	case 0:
		return models.OperatingModeInitializing
	case 1:
		return models.OperatingModeWaiting
	case 2:
		return models.OperatingModeGridConnected
	case 3:
		return models.OperatingModeOffGrid
	case 5:
		return models.OperatingModeFault
	case 6:
		return models.OperatingModeUpgrading
	default:
		return models.OperatingModeUnknown
	}
}

func sajBatteryDirection(dir int) models.EnergyDirection {
	switch dir {
	case -1:
		return models.DirectionCharging
	case 1:
		return models.DirectionDischarging
	case 0:
		return models.DirectionIdle
	default:
		return models.DirectionUnknown
	}
}

func sajGridDirection(dir int) models.GridDirection {
	switch dir {
	case -1:
		return models.GridDirectionImporting
	case 1:
		return models.GridDirectionExporting
	case 0:
		return models.GridDirectionIdle
	default:
		return models.GridDirectionUnknown
	}
}

func sajAlarmSeverity(level int) models.AlarmSeverity {
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

func sajAlarmStatus(status int) models.AlarmStatus {
	switch status {
	case 1:
		return models.AlarmStatusActive
	case 4:
		return models.AlarmStatusResolved
	default:
		return models.AlarmStatusUnknown
	}
}

func granularityToSAJTimeUnit(g models.Granularity) int {
	switch g {
	case models.GranularityMinute:
		return 0
	case models.GranularityDay:
		return 1
	case models.GranularityMonth:
		return 2
	case models.GranularityYear:
		return 3
	default:
		return 1
	}
}

// ══════════════════════════════════════════════════════════════════
// Extraction helpers for map[string]interface{} → typed values
// ══════════════════════════════════════════════════════════════════

func extractStringSafe(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func extractFloatSafe(m map[string]interface{}, key string) float64 {
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
	case int:
		return float64(val)
	default:
		return 0
	}
}

func extractFloatPtr(m map[string]interface{}, key string) *float64 {
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
	case int:
		f = float64(val)
	default:
		return nil
	}
	return &f
}

func extractIntSafe(m map[string]interface{}, key string) int {
	if m == nil {
		return 0
	}
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case string:
		i, _ := strconv.Atoi(val)
		return i
	default:
		return 0
	}
}

func extractIntPtr(m map[string]interface{}, key string) *int {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	var i int
	switch val := v.(type) {
	case float64:
		i = int(val)
	case int:
		i = val
	case string:
		parsed, err := strconv.Atoi(val)
		if err != nil {
			return nil
		}
		i = parsed
	default:
		return nil
	}
	return &i
}
