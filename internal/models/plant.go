package models

import "time"

// NormalizedPlant represents a solar installation / power station
// in a brand-agnostic format.
type NormalizedPlant struct {
	// Normalized unique ID: "provider_originalID"
	ID       string `json:"id"`
	Provider string `json:"provider"`

	// Core info
	Name     string  `json:"name"`
	Timezone string  `json:"timezone"`
	Location *LatLng `json:"location,omitempty"`
	Address  string  `json:"address,omitempty"`
	Country  string  `json:"country,omitempty"`

	// System specs
	PeakPowerKWp *float64 `json:"peakPowerKWp,omitempty"`
	PlantType    PlantType `json:"plantType"`

	// Grid connection
	GridConnectionType *GridConnectionType `json:"gridConnectionType,omitempty"`
	ElectricityPrice   *float64            `json:"electricityPrice,omitempty"`
	Currency           string              `json:"currency,omitempty"`

	// Identifiers from the original provider
	Meta ProviderMeta `json:"meta"`
}

type LatLng struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// PlantType unifies the plant/system types across all brands.
type PlantType string

const (
	PlantTypeGridTied      PlantType = "grid_tied"
	PlantTypeHybrid        PlantType = "hybrid"       // Storage / hybrid
	PlantTypeOffGrid       PlantType = "off_grid"
	PlantTypeACCoupled     PlantType = "ac_coupled"
	PlantTypeCommercial    PlantType = "commercial"
	PlantTypeUtility       PlantType = "utility_scale"
	PlantTypeUnknown       PlantType = "unknown"
)

type GridConnectionType string

const (
	GridConnectionFullExport   GridConnectionType = "full_export"
	GridConnectionSelfConsume  GridConnectionType = "self_consumption"
	GridConnectionOffGrid      GridConnectionType = "off_grid"
	GridConnectionUnknown      GridConnectionType = "unknown"
)

// ProviderMeta preserves the original provider-specific identifiers
// so data can always be traced back to the source.
type ProviderMeta struct {
	Provider         string            `json:"provider"`
	ProviderPlantID  string            `json:"providerPlantId,omitempty"`
	ProviderDeviceID string            `json:"providerDeviceId,omitempty"`
	ProviderPlantUID string            `json:"providerPlantUid,omitempty"`
	RawDataAvailable bool              `json:"rawDataAvailable"`
	Extra            map[string]string `json:"extra,omitempty"`
	FetchedAt        time.Time         `json:"fetchedAt"`
}
