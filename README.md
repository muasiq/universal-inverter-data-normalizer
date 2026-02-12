# Universal Inverter Data Normalizer

A production-grade Go service that normalizes solar inverter data from multiple brands (SMA, Huawei FusionSolar, Sungrow iSolarCloud, SAJ Elekeeper) into a **single unified schema** — regardless of manufacturer, API version, or data format quirks.

## Architecture

```
┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│   SMA API    │  │ Huawei API   │  │ Sungrow API  │  │   SAJ API    │
│  (Monitoring)│  │ (FusionSolar)│  │(iSolarCloud) │  │ (Elekeeper)  │
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       │                 │                 │                 │
       ▼                 ▼                 ▼                 ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  SMA Adapter │  │Huawei Adapter│  │Sungrow Adapt.│  │  SAJ Adapter │
│  (provider)  │  │  (provider)  │  │  (provider)  │  │  (provider)  │
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       │                 │                 │                 │
       └────────────┬────┴────────┬────────┘                 │
                    │             │                           │
                    ▼             ▼                           ▼
              ┌─────────────────────────────────────────────────┐
              │              Normalizer Engine                   │
              │  ┌──────────┐ ┌───────────┐ ┌───────────────┐  │
              │  │  Unit     │ │  Status   │ │  Direction    │  │
              │  │  Converter│ │  Mapper   │ │  Normalizer   │  │
              │  └──────────┘ └───────────┘ └───────────────┘  │
              └─────────────────────┬───────────────────────────┘
                                    │
                                    ▼
              ┌─────────────────────────────────────────────────┐
              │           Unified Data Models                    │
              │  • NormalizedPlant      • NormalizedDevice       │
              │  • NormalizedRealtime   • NormalizedEnergy       │
              │  • NormalizedAlarm      • NormalizedBattery      │
              │  • NormalizedTimeSeries • NormalizedMeter        │
              └─────────────────────┬───────────────────────────┘
                                    │
                         ┌──────────┼──────────┐
                         ▼          ▼          ▼
                    ┌────────┐ ┌────────┐ ┌────────┐
                    │REST API│ │Webhook │ │  MQTT  │
                    │(chi)   │ │Forward │ │Publish │
                    └────────┘ └────────┘ └────────┘
```

## Why Go?

| Requirement | Why Go Excels |
|---|---|
| **Multi-API concurrency** | Goroutines + channels for parallel polling of 4+ APIs |
| **Single binary deployment** | `go build` → one binary, no runtime deps, Docker-friendly |
| **Strong typing** | Compile-time safety for complex data normalization |
| **HTTP performance** | net/http is production-grade out of the box |
| **JSON marshaling** | Excellent struct tags, custom unmarshalers for quirky APIs |
| **Error handling** | Explicit error handling ideal for unreliable third-party APIs |
| **Cross-compilation** | Build for Linux/ARM (edge devices) from macOS |

## Project Structure

```
├── cmd/
│   └── normalizer/          # Main entry point
│       └── main.go
├── internal/
│   ├── config/              # YAML configuration loader
│   │   └── config.go
│   ├── models/              # Unified normalized data models
│   │   ├── plant.go         # Plant/site model
│   │   ├── device.go        # Device/inverter model
│   │   ├── realtime.go      # Real-time telemetry
│   │   ├── energy.go        # Energy statistics
│   │   ├── alarm.go         # Alarms and faults
│   │   └── timeseries.go    # Historical time-series
│   ├── provider/            # Brand-specific API adapters
│   │   ├── provider.go      # Provider interface
│   │   ├── registry.go      # Provider registry
│   │   ├── sma/             # SMA Monitoring API adapter
│   │   │   └── sma.go
│   │   ├── huawei/          # Huawei FusionSolar adapter
│   │   │   └── huawei.go
│   │   ├── sungrow/         # Sungrow iSolarCloud adapter
│   │   │   └── sungrow.go
│   │   └── saj/             # SAJ Elekeeper adapter
│   │       └── saj.go
│   ├── normalizer/          # Normalization engine
│   │   ├── engine.go        # Core normalization orchestrator
│   │   └── units.go         # Unit conversion utilities
│   └── api/                 # REST API server
│       ├── server.go        # HTTP server setup
│       ├── handlers.go      # API route handlers
│       └── middleware.go     # Logging, auth, CORS middleware
├── config.example.yaml      # Example configuration
├── go.mod
├── go.sum
├── Dockerfile
├── Makefile
└── README.md
```

## Quick Start

### 1. Configure

```bash
cp config.example.yaml config.yaml
# Edit config.yaml with your API credentials
```

### 2. Build & Run

```bash
make build
./bin/normalizer --config config.yaml

# Or with Docker
docker build -t inverter-normalizer .
docker run -p 8080:8080 -v $(pwd)/config.yaml:/app/config.yaml inverter-normalizer
```

### 3. Query Unified Data

```bash
# List all plants across all providers
curl http://localhost:8080/api/v1/plants

# Get real-time data for a specific device (brand-agnostic)
curl http://localhost:8080/api/v1/devices/{deviceId}/realtime

# Get energy statistics for a plant
curl http://localhost:8080/api/v1/plants/{plantId}/energy?period=day

# Get alarms across all providers
curl http://localhost:8080/api/v1/alarms?severity=critical

# Get historical time-series
curl http://localhost:8080/api/v1/devices/{deviceId}/history?start=2025-01-01&end=2025-01-31&granularity=day
```

## Unified API Reference

### Plants

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/plants` | List all plants from all configured providers |
| GET | `/api/v1/plants/{plantId}` | Get normalized plant details |
| GET | `/api/v1/plants/{plantId}/devices` | List all devices in a plant |
| GET | `/api/v1/plants/{plantId}/energy` | Plant energy statistics |

### Devices

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/devices` | List all devices across all providers |
| GET | `/api/v1/devices/{deviceId}` | Get normalized device details |
| GET | `/api/v1/devices/{deviceId}/realtime` | Real-time telemetry data |
| GET | `/api/v1/devices/{deviceId}/history` | Historical time-series data |
| GET | `/api/v1/devices/{deviceId}/alarms` | Device alarms |

### Alarms

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/alarms` | List all alarms across providers |

## Normalized Data Models

### Core Principles

1. **All power in Watts (W)** — no kW/MW ambiguity
2. **All energy in kWh** — consistent across all brands
3. **All timestamps in UTC ISO 8601** — with original timezone preserved
4. **All voltages in V, currents in A, frequencies in Hz**
5. **Unified status enums** — mapped from each brand's proprietary codes
6. **Unified direction enums** — charging/discharging/importing/exporting
7. **Provider metadata preserved** — original brand-specific IDs always accessible
8. **Null-safe** — pointer fields for optional data, zero values never confused with "no data"

### Example: Normalized Real-Time Response

```json
{
  "deviceId": "norm_saj_HSS2502J2351E34643",
  "provider": "saj",
  "timestamp": "2025-02-12T08:40:00Z",
  "originalTimestamp": "2025-02-12 16:40:00",
  "originalTimezone": "Asia/Shanghai",
  "status": "online",
  "operatingMode": "grid_connected",
  "pv": {
    "totalPowerW": 3500.0,
    "todayEnergyKWh": 17.67,
    "totalEnergyKWh": 2153.85,
    "strings": [
      {"id": 1, "voltageV": 350.0, "currentA": 10.0, "powerW": 3500.0}
    ]
  },
  "battery": {
    "socPercent": 100.0,
    "powerW": 171.0,
    "direction": "discharging",
    "todayChargeKWh": 5.2,
    "todayDischargeKWh": 3.1,
    "temperatureC": 25.5
  },
  "grid": {
    "totalPowerW": 283.0,
    "direction": "idle",
    "todayImportKWh": 0.0,
    "todayExportKWh": 3.75,
    "totalExportKWh": 473.89,
    "phases": [
      {"phase": "A", "voltageV": 244.1, "currentA": 1.64, "powerW": 242.0, "frequencyHz": 50.01}
    ]
  },
  "load": {
    "totalPowerW": 443.0,
    "todayEnergyKWh": 12.5
  },
  "environment": {
    "inverterTemperatureC": 36.9,
    "ambientTemperatureC": 45.5,
    "signalStrengthDBm": -48
  },
  "meta": {
    "provider": "saj",
    "providerDeviceId": "HSS2502J2351E34643",
    "providerPlantId": "17198068xx",
    "rawDataAvailable": true
  }
}
```

## Supported Providers

| Provider | Auth Method | Data Coverage | Status |
|---|---|---|---|
| **SMA** (Sunny Portal / ennexOS) | OAuth2 Bearer Token | Plants, Devices, Measurements, Logs | ✅ Implemented |
| **Huawei** (FusionSolar) | Login + XSRF Token | Plants, Devices, Real-time KPI, Alarms | ✅ Implemented |
| **Sungrow** (iSolarCloud) | API Key + App Secret | Plants, Devices, Real-time, History | ✅ Implemented |
| **SAJ** (Elekeeper) | App ID + App Secret Token | Plants, Devices, Real-time, History, EMS, Alarms | ✅ Implemented |

## Adding a New Provider

1. Create a new package under `internal/provider/yourbrand/`
2. Implement the `provider.Provider` interface
3. Register in `internal/provider/registry.go`
4. Add config section in `config.yaml`

```go
type Provider interface {
    Name() string
    Initialize(ctx context.Context, cfg ProviderConfig) error
    GetPlants(ctx context.Context) ([]models.NormalizedPlant, error)
    GetDevices(ctx context.Context, plantID string) ([]models.NormalizedDevice, error)
    GetRealTimeData(ctx context.Context, deviceID string) (*models.NormalizedRealtime, error)
    GetEnergyStats(ctx context.Context, plantID string, period models.Period) (*models.NormalizedEnergy, error)
    GetHistoricalData(ctx context.Context, deviceID string, req models.HistoryRequest) ([]models.NormalizedTimeSeries, error)
    GetAlarms(ctx context.Context, deviceID string) ([]models.NormalizedAlarm, error)
    Close() error
}
```

## License

MIT
