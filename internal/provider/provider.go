package provider

import (
	"context"

	"github.com/muasiq/universal-inverter-data-normalizer/internal/models"
)

// Provider is the interface every inverter brand adapter must implement.
// It acts as the contract between the normalization engine and brand-specific APIs.
//
// Each method returns already-normalized data — the adapter is responsible
// for translating vendor-specific schemas into the universal models.
type Provider interface {
	// Name returns the canonical provider identifier (e.g., "sma", "huawei", "sungrow", "saj").
	Name() string

	// Initialize sets up the provider with credentials and validates connectivity.
	Initialize(ctx context.Context, cfg ProviderConfig) error

	// ── Plant Operations ──

	// GetPlants returns all plants/sites accessible with the configured credentials.
	GetPlants(ctx context.Context) ([]models.NormalizedPlant, error)

	// GetPlantDetails returns detailed information about a specific plant.
	GetPlantDetails(ctx context.Context, plantID string) (*models.NormalizedPlant, error)

	// ── Device Operations ──

	// GetDevices returns all devices within a plant.
	GetDevices(ctx context.Context, plantID string) ([]models.NormalizedDevice, error)

	// GetDeviceDetails returns detailed information about a specific device.
	GetDeviceDetails(ctx context.Context, deviceID string) (*models.NormalizedDevice, error)

	// ── Data Operations ──

	// GetRealTimeData returns the latest telemetry snapshot for a device.
	GetRealTimeData(ctx context.Context, deviceID string) (*models.NormalizedRealtime, error)

	// GetEnergyStats returns energy statistics for a plant for a given period.
	GetEnergyStats(ctx context.Context, plantID string, period models.Period) (*models.NormalizedEnergy, error)

	// GetHistoricalData returns time-series data for a device within a time range.
	GetHistoricalData(ctx context.Context, deviceID string, req models.HistoryRequest) (*models.HistoryResponse, error)

	// ── Alarm Operations ──

	// GetAlarms returns active and recent alarms for a device.
	GetAlarms(ctx context.Context, deviceID string) ([]models.NormalizedAlarm, error)

	// GetAllAlarms returns alarms across all devices (if supported by the provider).
	GetAllAlarms(ctx context.Context) ([]models.NormalizedAlarm, error)

	// ── Lifecycle ──

	// Healthy returns true if the provider's API connection is working.
	Healthy(ctx context.Context) bool

	// Close releases any resources held by the provider.
	Close() error
}

// ProviderConfig holds the credentials and settings for a single provider.
type ProviderConfig struct {
	// Provider type identifier
	Type string `yaml:"type"`

	// Display name for this provider instance
	Name string `yaml:"name"`

	// Whether this provider is enabled
	Enabled bool `yaml:"enabled"`

	// Base URL override (for sandbox/staging environments)
	BaseURL string `yaml:"base_url"`

	// Credentials — varies by provider
	Credentials map[string]string `yaml:"credentials"`

	// Rate limit override (requests per second)
	RateLimitRPS int `yaml:"rate_limit_rps"`

	// Request timeout in seconds
	TimeoutSeconds int `yaml:"timeout_seconds"`

	// Timezone for this provider's data (e.g. "Asia/Shanghai")
	Timezone string `yaml:"timezone"`
}

// GetCredential retrieves a credential value or returns empty string.
func (c ProviderConfig) GetCredential(key string) string {
	if c.Credentials == nil {
		return ""
	}
	return c.Credentials[key]
}
