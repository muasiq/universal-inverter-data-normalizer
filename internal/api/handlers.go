package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/muasiq/universal-inverter-data-normalizer/internal/models"
	"github.com/rs/zerolog/log"
)

// --- Response helpers ---

type apiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *apiMeta    `json:"meta,omitempty"`
}

type apiMeta struct {
	Total     int    `json:"total,omitempty"`
	Timestamp string `json:"timestamp"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error().Err(err).Msg("Failed to write JSON response")
	}
}

func writeSuccess(w http.ResponseWriter, data interface{}, total int) {
	writeJSON(w, http.StatusOK, apiResponse{
		Success: true,
		Data:    data,
		Meta: &apiMeta{
			Total:     total,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	})
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, apiResponse{
		Success: false,
		Error:   msg,
	})
}

// --- Handlers ---

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := s.engine.HealthCheck(r.Context())
	allHealthy := true
	for _, h := range health {
		if !h {
			allHealthy = false
			break
		}
	}
	status := http.StatusOK
	if !allHealthy {
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, apiResponse{
		Success: allHealthy,
		Data:    health,
		Meta: &apiMeta{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	})
}

// handleGetPlants returns all plants across all providers.
func (s *Server) handleGetPlants(w http.ResponseWriter, r *http.Request) {
	plants, err := s.engine.GetAllPlants(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get plants")
		writeError(w, http.StatusInternalServerError, "Failed to retrieve plants")
		return
	}
	writeSuccess(w, plants, len(plants))
}

// handleGetPlantDetails returns details for a specific plant.
func (s *Server) handleGetPlantDetails(w http.ResponseWriter, r *http.Request) {
	plantID := chi.URLParam(r, "plantId")
	if plantID == "" {
		writeError(w, http.StatusBadRequest, "Plant ID is required")
		return
	}

	provider := r.URL.Query().Get("provider")
	if provider == "" {
		writeError(w, http.StatusBadRequest, "Query parameter 'provider' is required (e.g. ?provider=saj)")
		return
	}

	plant, err := s.engine.GetPlantDetails(r.Context(), provider, plantID)
	if err != nil {
		log.Error().Err(err).Str("plant_id", plantID).Str("provider", provider).Msg("Failed to get plant details")
		writeError(w, http.StatusInternalServerError, "Failed to retrieve plant details")
		return
	}
	writeSuccess(w, plant, 1)
}

// handleGetPlantDevices returns all devices for a specific plant.
func (s *Server) handleGetPlantDevices(w http.ResponseWriter, r *http.Request) {
	plantID := chi.URLParam(r, "plantId")
	if plantID == "" {
		writeError(w, http.StatusBadRequest, "Plant ID is required")
		return
	}

	provider := r.URL.Query().Get("provider")
	if provider == "" {
		writeError(w, http.StatusBadRequest, "Query parameter 'provider' is required")
		return
	}

	devices, err := s.engine.GetDevices(r.Context(), provider, plantID)
	if err != nil {
		log.Error().Err(err).Str("plant_id", plantID).Str("provider", provider).Msg("Failed to get devices")
		writeError(w, http.StatusInternalServerError, "Failed to retrieve devices")
		return
	}
	writeSuccess(w, devices, len(devices))
}

// handleGetPlantEnergy returns energy statistics for a specific plant.
func (s *Server) handleGetPlantEnergy(w http.ResponseWriter, r *http.Request) {
	plantID := chi.URLParam(r, "plantId")
	if plantID == "" {
		writeError(w, http.StatusBadRequest, "Plant ID is required")
		return
	}

	provider := r.URL.Query().Get("provider")
	if provider == "" {
		writeError(w, http.StatusBadRequest, "Query parameter 'provider' is required")
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "day"
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}

	energy, err := s.engine.GetEnergyStats(r.Context(), provider, plantID, period, date)
	if err != nil {
		log.Error().Err(err).Str("plant_id", plantID).Str("provider", provider).Msg("Failed to get energy stats")
		writeError(w, http.StatusInternalServerError, "Failed to retrieve energy statistics")
		return
	}
	writeSuccess(w, energy, 1)
}

// handleGetDevices returns all devices across all providers.
func (s *Server) handleGetDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := s.engine.GetAllDevices(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all devices")
		writeError(w, http.StatusInternalServerError, "Failed to retrieve devices")
		return
	}
	writeSuccess(w, devices, len(devices))
}

// handleGetRealTimeData returns real-time data for a specific device.
func (s *Server) handleGetRealTimeData(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "deviceId")
	if deviceID == "" {
		writeError(w, http.StatusBadRequest, "Device ID is required")
		return
	}

	provider := r.URL.Query().Get("provider")
	if provider == "" {
		writeError(w, http.StatusBadRequest, "Query parameter 'provider' is required")
		return
	}

	data, err := s.engine.GetRealTimeData(r.Context(), provider, deviceID)
	if err != nil {
		log.Error().Err(err).Str("device_id", deviceID).Str("provider", provider).Msg("Failed to get real-time data")
		writeError(w, http.StatusInternalServerError, "Failed to retrieve real-time data")
		return
	}
	writeSuccess(w, data, 1)
}

// handleGetHistoricalData returns historical time-series data for a device.
func (s *Server) handleGetHistoricalData(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "deviceId")
	if deviceID == "" {
		writeError(w, http.StatusBadRequest, "Device ID is required")
		return
	}

	provider := r.URL.Query().Get("provider")
	if provider == "" {
		writeError(w, http.StatusBadRequest, "Query parameter 'provider' is required")
		return
	}

	granularity := r.URL.Query().Get("granularity")
	if granularity == "" {
		granularity = "hour"
	}

	startTime := r.URL.Query().Get("start")
	endTime := r.URL.Query().Get("end")
	if startTime == "" || endTime == "" {
		now := time.Now().UTC()
		if endTime == "" {
			endTime = now.Format(time.RFC3339)
		}
		if startTime == "" {
			startTime = now.Add(-24 * time.Hour).Format(time.RFC3339)
		}
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	pageSize := 100
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 1000 {
			pageSize = parsed
		}
	}

	req := models.HistoryRequest{
		DeviceID:    deviceID,
		Granularity: models.Granularity(granularity),
		StartTime:   startTime,
		EndTime:     endTime,
		Page:        page,
		PageSize:    pageSize,
	}

	// Collect requested metrics
	if metrics := r.URL.Query().Get("metrics"); metrics != "" {
		req.Metrics = splitCSV(metrics)
	}

	resp, err := s.engine.GetHistoricalData(r.Context(), provider, req)
	if err != nil {
		log.Error().Err(err).Str("device_id", deviceID).Str("provider", provider).Msg("Failed to get historical data")
		writeError(w, http.StatusInternalServerError, "Failed to retrieve historical data")
		return
	}
	writeSuccess(w, resp, resp.TotalPoints)
}

// handleGetDeviceAlarms returns alarms for a specific device.
func (s *Server) handleGetDeviceAlarms(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "deviceId")
	if deviceID == "" {
		writeError(w, http.StatusBadRequest, "Device ID is required")
		return
	}

	provider := r.URL.Query().Get("provider")
	if provider == "" {
		writeError(w, http.StatusBadRequest, "Query parameter 'provider' is required")
		return
	}

	alarms, err := s.engine.GetDeviceAlarms(r.Context(), provider, deviceID)
	if err != nil {
		log.Error().Err(err).Str("device_id", deviceID).Str("provider", provider).Msg("Failed to get device alarms")
		writeError(w, http.StatusInternalServerError, "Failed to retrieve device alarms")
		return
	}
	writeSuccess(w, alarms, len(alarms))
}

// handleGetAllAlarms returns all alarms across all providers.
func (s *Server) handleGetAllAlarms(w http.ResponseWriter, r *http.Request) {
	alarms, err := s.engine.GetAllAlarms(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all alarms")
		writeError(w, http.StatusInternalServerError, "Failed to retrieve alarms")
		return
	}
	writeSuccess(w, alarms, len(alarms))
}

// handleGetProviders returns list of registered providers and their health.
func (s *Server) handleGetProviders(w http.ResponseWriter, r *http.Request) {
	health := s.engine.HealthCheck(r.Context())
	type providerInfo struct {
		Name    string `json:"name"`
		Healthy bool   `json:"healthy"`
	}
	providers := make([]providerInfo, 0, len(health))
	for name, healthy := range health {
		providers = append(providers, providerInfo{
			Name:    name,
			Healthy: healthy,
		})
	}
	writeSuccess(w, providers, len(providers))
}

// --- Helpers ---

func splitCSV(s string) []string {
	var result []string
	current := ""
	for _, c := range s {
		if c == ',' {
			trimmed := trim(current)
			if trimmed != "" {
				result = append(result, trimmed)
			}
			current = ""
		} else {
			current += string(c)
		}
	}
	trimmed := trim(current)
	if trimmed != "" {
		result = append(result, trimmed)
	}
	return result
}

func trim(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
