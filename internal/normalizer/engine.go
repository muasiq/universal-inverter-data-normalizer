package normalizer

import (
	"context"
	"fmt"
	"sync"

	"github.com/muasiq/universal-inverter-data-normalizer/internal/models"
	"github.com/muasiq/universal-inverter-data-normalizer/internal/provider"
	"github.com/rs/zerolog/log"
)

// Engine is the central orchestrator that manages all configured providers
// and exposes a unified interface for querying normalized data.
type Engine struct {
	mu        sync.RWMutex
	providers map[string]provider.Provider
}

// NewEngine creates a new normalization engine.
func NewEngine() *Engine {
	return &Engine{
		providers: make(map[string]provider.Provider),
	}
}

// RegisterProvider adds an initialized provider to the engine.
func (e *Engine) RegisterProvider(p provider.Provider) {
	e.mu.Lock()
	defer e.mu.Unlock()
	name := p.Name()
	e.providers[name] = p
	log.Info().Str("provider", name).Msg("Provider registered with engine")
}

// GetProvider returns a specific provider by name.
func (e *Engine) GetProvider(name string) (provider.Provider, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	p, ok := e.providers[name]
	return p, ok
}

// ProviderNames returns the names of all registered providers.
func (e *Engine) ProviderNames() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	names := make([]string, 0, len(e.providers))
	for name := range e.providers {
		names = append(names, name)
	}
	return names
}

// ── Aggregated queries across all providers ──

// GetAllPlants returns plants from all configured providers, fetched concurrently.
func (e *Engine) GetAllPlants(ctx context.Context) ([]models.NormalizedPlant, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	type result struct {
		plants []models.NormalizedPlant
		err    error
		name   string
	}

	ch := make(chan result, len(e.providers))
	for name, p := range e.providers {
		go func(name string, p provider.Provider) {
			plants, err := p.GetPlants(ctx)
			ch <- result{plants: plants, err: err, name: name}
		}(name, p)
	}

	var allPlants []models.NormalizedPlant
	var errors []error
	for i := 0; i < len(e.providers); i++ {
		r := <-ch
		if r.err != nil {
			log.Error().Err(r.err).Str("provider", r.name).Msg("Failed to fetch plants")
			errors = append(errors, fmt.Errorf("%s: %w", r.name, r.err))
			continue
		}
		allPlants = append(allPlants, r.plants...)
	}

	if len(allPlants) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("all providers failed: %v", errors)
	}

	return allPlants, nil
}

// GetAllDevices returns all devices from all providers (from a specific plant or all).
func (e *Engine) GetAllDevices(ctx context.Context) ([]models.NormalizedDevice, error) {
	plants, err := e.GetAllPlants(ctx)
	if err != nil {
		return nil, err
	}

	type result struct {
		devices []models.NormalizedDevice
		err     error
	}

	ch := make(chan result, len(plants))
	for _, plant := range plants {
		go func(plant models.NormalizedPlant) {
			p, ok := e.providers[plant.Provider]
			if !ok {
				ch <- result{err: fmt.Errorf("provider %s not found", plant.Provider)}
				return
			}
			devices, err := p.GetDevices(ctx, plant.Meta.ProviderPlantID)
			ch <- result{devices: devices, err: err}
		}(plant)
	}

	var allDevices []models.NormalizedDevice
	for i := 0; i < len(plants); i++ {
		r := <-ch
		if r.err != nil {
			log.Warn().Err(r.err).Msg("Failed to fetch devices for a plant")
			continue
		}
		allDevices = append(allDevices, r.devices...)
	}

	return allDevices, nil
}

// GetRealTimeData fetches real-time data from the appropriate provider.
func (e *Engine) GetRealTimeData(ctx context.Context, providerName, deviceID string) (*models.NormalizedRealtime, error) {
	p, ok := e.GetProvider(providerName)
	if !ok {
		return nil, fmt.Errorf("provider %q not registered", providerName)
	}
	return p.GetRealTimeData(ctx, deviceID)
}

// GetPlantDetails fetches details for a specific plant from the given provider.
func (e *Engine) GetPlantDetails(ctx context.Context, providerName, plantID string) (*models.NormalizedPlant, error) {
	p, ok := e.GetProvider(providerName)
	if !ok {
		return nil, fmt.Errorf("provider %q not registered", providerName)
	}
	return p.GetPlantDetails(ctx, plantID)
}

// GetDevices fetches all devices from a specific plant/provider.
func (e *Engine) GetDevices(ctx context.Context, providerName, plantID string) ([]models.NormalizedDevice, error) {
	p, ok := e.GetProvider(providerName)
	if !ok {
		return nil, fmt.Errorf("provider %q not registered", providerName)
	}
	return p.GetDevices(ctx, plantID)
}

// GetEnergyStats fetches energy stats from the appropriate provider.
func (e *Engine) GetEnergyStats(ctx context.Context, providerName, plantID, period, date string) (*models.NormalizedEnergy, error) {
	p, ok := e.GetProvider(providerName)
	if !ok {
		return nil, fmt.Errorf("provider %q not registered", providerName)
	}
	return p.GetEnergyStats(ctx, plantID, models.Period(period))
}

// GetHistoricalData fetches historical data from the appropriate provider.
func (e *Engine) GetHistoricalData(ctx context.Context, providerName string, req models.HistoryRequest) (*models.HistoryResponse, error) {
	p, ok := e.GetProvider(providerName)
	if !ok {
		return nil, fmt.Errorf("provider %q not registered", providerName)
	}
	return p.GetHistoricalData(ctx, req.DeviceID, req)
}

// GetAllAlarms returns alarms from all providers concurrently.
func (e *Engine) GetAllAlarms(ctx context.Context) ([]models.NormalizedAlarm, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	type result struct {
		alarms []models.NormalizedAlarm
		err    error
		name   string
	}

	ch := make(chan result, len(e.providers))
	for name, p := range e.providers {
		go func(name string, p provider.Provider) {
			alarms, err := p.GetAllAlarms(ctx)
			ch <- result{alarms: alarms, err: err, name: name}
		}(name, p)
	}

	var allAlarms []models.NormalizedAlarm
	for i := 0; i < len(e.providers); i++ {
		r := <-ch
		if r.err != nil {
			log.Warn().Err(r.err).Str("provider", r.name).Msg("Failed to fetch alarms")
			continue
		}
		allAlarms = append(allAlarms, r.alarms...)
	}

	return allAlarms, nil
}

// GetDeviceAlarms fetches alarms for a specific device.
func (e *Engine) GetDeviceAlarms(ctx context.Context, providerName, deviceID string) ([]models.NormalizedAlarm, error) {
	p, ok := e.GetProvider(providerName)
	if !ok {
		return nil, fmt.Errorf("provider %q not registered", providerName)
	}
	return p.GetAlarms(ctx, deviceID)
}

// HealthCheck returns the health status of all providers.
func (e *Engine) HealthCheck(ctx context.Context) map[string]bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	status := make(map[string]bool, len(e.providers))
	for name, p := range e.providers {
		status[name] = p.Healthy(ctx)
	}
	return status
}

// Close shuts down all providers.
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for name, p := range e.providers {
		if err := p.Close(); err != nil {
			log.Error().Err(err).Str("provider", name).Msg("Error closing provider")
		}
	}
	return nil
}
