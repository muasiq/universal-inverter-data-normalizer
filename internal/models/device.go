package models

// NormalizedDevice represents a single device (inverter, meter, battery, EV charger, etc.)
// in a brand-agnostic format.
type NormalizedDevice struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	PlantID  string `json:"plantId"`

	// Core info
	Name         string     `json:"name"`
	SerialNumber string     `json:"serialNumber"`
	Model        string     `json:"model,omitempty"`
	DeviceType   DeviceType `json:"deviceType"`
	Manufacturer string     `json:"manufacturer"`

	// Status
	Status       DeviceStatus `json:"status"`
	IsOnline     bool         `json:"isOnline"`
	HasAlarm     bool         `json:"hasAlarm"`

	// Specs
	RatedPowerW  *float64 `json:"ratedPowerW,omitempty"`
	FirmwareInfo *FirmwareInfo `json:"firmwareInfo,omitempty"`

	// Battery info (if device has battery)
	BatteryInfo *DeviceBatteryInfo `json:"batteryInfo,omitempty"`

	// Provider metadata
	Meta ProviderMeta `json:"meta"`
}

// DeviceType enumerates the categories of solar equipment.
type DeviceType string

const (
	DeviceTypeInverter        DeviceType = "inverter"
	DeviceTypeHybridInverter  DeviceType = "hybrid_inverter"
	DeviceTypeMicroInverter   DeviceType = "micro_inverter"
	DeviceTypeStringInverter  DeviceType = "string_inverter"
	DeviceTypeBattery         DeviceType = "battery"
	DeviceTypeMeter           DeviceType = "meter"
	DeviceTypeEMS             DeviceType = "ems"
	DeviceTypeEVCharger       DeviceType = "ev_charger"
	DeviceTypeLoadMonitor     DeviceType = "load_monitor"
	DeviceTypeWeatherStation  DeviceType = "weather_station"
	DeviceTypeDieselGenerator DeviceType = "diesel_generator"
	DeviceTypeHeatPump        DeviceType = "heat_pump"
	DeviceTypeSmartPlug       DeviceType = "smart_plug"
	DeviceTypeOptimizer       DeviceType = "optimizer"
	DeviceTypeGateway         DeviceType = "gateway"
	DeviceTypeUnknown         DeviceType = "unknown"
)

// DeviceStatus is the unified status across all brands.
type DeviceStatus string

const (
	DeviceStatusOnline   DeviceStatus = "online"
	DeviceStatusOffline  DeviceStatus = "offline"
	DeviceStatusStandby  DeviceStatus = "standby"
	DeviceStatusNormal   DeviceStatus = "normal"
	DeviceStatusWarning  DeviceStatus = "warning"
	DeviceStatusFault    DeviceStatus = "fault"
	DeviceStatusUpgrade  DeviceStatus = "upgrading"
	DeviceStatusUnknown  DeviceStatus = "unknown"
)

type FirmwareInfo struct {
	MainVersion    string `json:"mainVersion,omitempty"`
	SlaveVersion   string `json:"slaveVersion,omitempty"`
	DisplayVersion string `json:"displayVersion,omitempty"`
	ModuleModel    string `json:"moduleModel,omitempty"`
	ModuleSN       string `json:"moduleSn,omitempty"`
	ModuleFirmware string `json:"moduleFirmware,omitempty"`
}

type DeviceBatteryInfo struct {
	Count         int      `json:"count"`
	BatteryType   string   `json:"batteryType,omitempty"`
	CapacityUnit  string   `json:"capacityUnit,omitempty"` // "Ah" or "kWh"
	SerialNumbers []string `json:"serialNumbers,omitempty"`
}
