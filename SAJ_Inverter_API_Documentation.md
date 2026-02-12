# SAJ Inverter — Elekeeper Open Platform API Documentation

> **Base URL (Production):** `https://intl-developer.saj-electric.com/prod-api`
>
> **Rate Limit:** 5 QPS (unless otherwise noted per endpoint)
>
> **Platform:** Elekeeper Open Platform

---

## Table of Contents

1. [Overview](#1-overview)
2. [Authentication & Authorization](#2-authentication--authorization)
   - 2.1 [Get Developer's Access Token](#21-get-developers-access-token)
   - 2.2 [Signature Rule (clientSign)](#22-signature-rule-clientsign)
   - 2.3 [Auth Plant to Developer by Token](#23-auth-plant-to-developer-by-token)
3. [Common Headers](#3-common-headers)
4. [Common Response Codes](#4-common-response-codes)
5. [System API](#5-system-api)
   - 5.1 [Get Developer's Authorized Device By Page](#51-get-developers-authorized-device-by-page)
   - 5.2 [Get Developer's Authorized Plant By Page](#52-get-developers-authorized-plant-by-page)
6. [Information API](#6-information-api)
   - 6.1 [Get All Devices from Plant](#61-get-all-devices-from-plant)
   - 6.2 [Get Basic Information](#62-get-basic-information)
   - 6.3 [Get Device Details Information](#63-get-device-details-information)
7. [Data API](#7-data-api)
   - 7.1 [Realtime Data (Common)](#71-realtime-data-common)
   - 7.2 [EMS Real-Time Data](#72-ems-real-time-data)
   - 7.3 [Get Device Upload Data](#73-get-device-upload-data)
   - 7.4 [Get History Data (Common)](#74-get-history-data-common)
   - 7.5 [Get EMS History Data](#75-get-ems-history-data)
   - 7.6 [EMS Meter Historical Data](#76-ems-meter-historical-data)
   - 7.7 [Query Energy Flow Data of Device in Plant](#77-query-energy-flow-data-of-device-in-plant)
   - 7.8 [Query Details of Plant](#78-query-details-of-plant)
   - 7.9 [Query Energy of Plant](#79-query-energy-of-plant)
   - 7.10 [Query Plant Statistics Data](#710-query-plant-statistics-data)
   - 7.11 [Query Load Monitoring Data](#711-query-load-monitoring-data)
8. [Alarm & Fault API](#8-alarm--fault-api)
   - 8.1 [Get Fault Events of Device (New)](#81-get-fault-events-of-device-new)
   - 8.2 [Get Alarms of Device](#82-get-alarms-of-device)

---

## 1. Overview

The SAJ Elekeeper Open Platform provides a RESTful API that enables third-party developers to access solar inverter data, plant (power station) information, device telemetry, energy statistics, and alarm/fault events for SAJ inverter systems.

**Key Concepts:**

| Concept | Description |
|---------|-------------|
| **Plant** | A solar power station installation containing one or more devices |
| **Device** | An inverter, EMS module, battery, meter, or other connected equipment |
| **deviceSn** | The unique serial number identifying a specific device |
| **plantId** | The unique identifier for a plant/power station |
| **plantUid** | The unique UID for a plant |
| **moduleSn** | The serial number of a communication/monitoring module |

**Supported Device Types:**

| deviceType | Description |
|------------|-------------|
| 0 | Grid-tied inverter |
| 1 | Inverter (storage/hybrid) |
| 2 | Load monitor |
| 3 | EV Charger |
| 4 | Mobile storage |
| 6 | EMS/SEP module |
| 7 | Electric meter |
| 8 | Air conditioner |
| 9 | Fire safety |
| 10 | Diesel generator |
| 11 | Dry contact |
| 14 | eddi water heater |
| 16 | Longi PV optimizer |
| 17 | Heat pump (Phnix) |
| 18 | JA PV optimizer |
| 19 | Weather station |

---

## 2. Authentication & Authorization

### 2.1 Get Developer's Access Token

Obtain a time-limited access token using your `appId` and `appSecret`.

| Property | Value |
|----------|-------|
| **Method** | `GET` |
| **Endpoint** | `/open/api/access_token` |
| **Full URL** | `https://intl-developer.saj-electric.com/prod-api/open/api/access_token` |

#### Request Headers

| Name | Value | Required | Description |
|------|-------|----------|-------------|
| `content-language` | `zh_CN` / `en_US` | YES | Response language: Chinese or English |

#### Request Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `appId` | String | YES | Your application ID |
| `appSecret` | String | YES | Your application secret |

#### Request Example

```
GET /prod-api/open/api/access_token?appId=VH_GiQz6CMd&appSecret=A9VMLcFbt14iczKVh9i4RDrz6CMduedpTaT3Q1noQaEaGiQzl0U9GWftUSTbAfie

Headers:
  content-language: en_US
```

#### Response

| Name | Type | Description |
|------|------|-------------|
| `code` | integer | Response code |
| `msg` | string | Response message |
| `data` | object | Response data |
| `data.access_token` | string | The access token string |
| `data.expires` | integer | Token expiration time in seconds (default: 28800s = 8 hours) |

#### Response Example

```json
{
  "code": 200,
  "msg": "",
  "data": {
    "access_token": "MFwwDQYJKoZIhvcNAQEBBQAD...",
    "expires": 7200
  }
}
```

#### Endpoint-Specific Response Codes

| Code | Description |
|------|-------------|
| `200` | Request success |
| `100009` | appId or appSecret does not exist |
| `200014` | {xxx} app is not released |

---

### 2.2 Signature Rule (clientSign)

For endpoints that accept `clientSign`, generate a signature as per the SAJ Signature Rule documentation. The `clientSign` and `clientSecret` are **not required** for most endpoints (only `accessToken` is mandatory).

**Common authentication headers used across all protected endpoints:**

| Header | Required | Description |
|--------|----------|-------------|
| `accessToken` | YES | Obtained from the access token endpoint |
| `clientSecret` | NO | Optional client secret |
| `clientSign` | NO | Optional signature (see SAJ Signature Rule docs) |
| `content-language` | YES | `zh_CN` (Chinese) or `en_US` (English) |

---

### 2.3 Auth Plant to Developer by Token

Authorize one or more plants to a developer account using an OpenAPI key and credentials.

| Property | Value |
|----------|-------|
| **Method** | `POST` |
| **Endpoint** | `/open/api/developer/auth/plantAuthByOpenapiKey` |
| **Full URL** | `https://intl-developer.saj-electric.com/prod-api/open/api/developer/auth/plantAuthByOpenapiKey` |

#### Request Headers

| Name | Value | Required | Description |
|------|-------|----------|-------------|
| `content-language` | `zh_CN` / `en_US` | YES | Response language |

#### Request Parameters (JSON Body)

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `appKey` | String | YES | App key — use the Client Secret of the token |
| `credentials` | String | YES | App secret — use the Grant Type of the token |

#### Request Example

```json
{
  "appKey": "BD624*********************C4E9D",
  "credentials": "12E784F******************358F8DE"
}
```

#### Response

| Name | Type | Description |
|------|------|-------------|
| `code` | integer | Response code |
| `msg` | string | Response message |
| `data.authSuccessPlantId` | array(string) | List of successfully authorized plant IDs |

#### Response Example

```json
{
  "code": 200,
  "msg": "",
  "data": {
    "authSuccessPlantId": ["plant001", "plant002"]
  }
}
```

#### Endpoint-Specific Response Codes

| Code | Description |
|------|-------------|
| `200` | Request success |
| `10001` | System error, please try again later |
| `10006` | Parameter error |
| `10016` | Device connect timeout |
| `10021` | Data exception |

---

## 3. Common Headers

All authenticated API requests must include the following headers:

| Header | Value | Required | Description |
|--------|-------|----------|-------------|
| `accessToken` | `<token>` | YES | Developer access token (see §2.1) |
| `clientSecret` | `<secret>` | NO | Client secret (optional for most endpoints) |
| `clientSign` | `<signature>` | NO | Request signature (optional; see §2.2) |
| `content-language` | `zh_CN` or `en_US` | YES | Language preference |
| `Content-Type` | `application/json` | YES (POST only) | Required for POST requests |

---

## 4. Common Response Codes

These codes may be returned by any endpoint:

| Code | Description |
|------|-------------|
| `200` | Request success |
| `10001` | System error, please try again later |
| `10002` | Not logged in |
| `10003` | Parameter validation failed |
| `10004` | Authentication failed |
| `10005` | No permission |
| `10006` | Parameter error / Too frequent access |
| `10007` | User logged out |
| `10016` | Device connect timeout |
| `10021` | Data exception |
| `200008` | App {xxx} does not exist |
| `200010` | accessToken does not exist |
| `200015` | App {xxx} do not match. Unauthorized access to data |

---

## 5. System API

### 5.1 Get Developer's Authorized Device By Page

Retrieve a paginated list of devices authorized to the developer.

| Property | Value |
|----------|-------|
| **Method** | `GET` |
| **Endpoint** | `/open/api/developer/device/page` |
| **Full URL** | `https://intl-developer.saj-electric.com/prod-api/open/api/developer/device/page` |

#### Request Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `appId` | String | YES | App ID |
| `pageSize` | Integer | NO | Page size (default: 100) |
| `pageNum` | Integer | NO | Current page number (default: 1) |

#### Request Example

```
GET /prod-api/open/api/developer/device/page?appId=VH_GiQz6CMd&pageSize=100&pageNum=1

Headers:
  accessToken: <your_token>
  content-language: en_US
```

#### Response

| Name | Type | Description |
|------|------|-------------|
| `code` | integer | Response code |
| `msg` | string | Response message |
| `total` | integer | Total number of devices |
| `totalPage` | integer | Total number of pages |
| `rows` | array | List of device objects |
| `rows[].deviceSn` | string | Device serial number |
| `rows[].deviceType` | string | Device type code |
| `rows[].plantId` | string | Plant ID |
| `rows[].plantName` | string | Plant name |
| `rows[].isOnline` | integer | Online status (0: No, 1: Yes) |
| `rows[].isAlarm` | integer | Alarm status (0: No, 1: Yes) |
| `rows[].country` | string | Country of the device |
| `rows[].userName` | string | Plant user name |
| `rows[].modelType` | string | Model type |

#### Response Example

```json
{
  "total": 39,
  "rows": [
    {
      "deviceSn": "R5T2203J202xxx",
      "deviceType": "R5",
      "plantId": "17198068xx",
      "plantName": "Siddhirgaxxx",
      "isOnline": 0,
      "isAlarm": 0,
      "country": "Bangladesh"
    }
  ],
  "code": 200,
  "msg": "Query success",
  "totalPage": 4
}
```

#### Endpoint-Specific Response Codes

| Code | Description |
|------|-------------|
| `200` | Request success |
| `200008` | App {xxx} does not exist |
| `200010` | accessToken does not exist |
| `200015` | App {xxx} do not match. Unauthorized access to data |

---

### 5.2 Get Developer's Authorized Plant By Page

Retrieve a paginated list of plants authorized to the developer.

| Property | Value |
|----------|-------|
| **Method** | `GET` |
| **Endpoint** | `/open/api/developer/plant/page` |
| **Full URL** | `https://intl-developer.saj-electric.com/prod-api/open/api/developer/plant/page` |

#### Request Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `appId` | String | YES | App ID |
| `pageSize` | Integer | NO | Page size (default: 100) |
| `pageNum` | Integer | NO | Current page number (default: 1) |

#### Request Example

```
GET /prod-api/open/api/developer/plant/page?appId=VH_GiQz6CMd&pageSize=100&pageNum=1

Headers:
  accessToken: <your_token>
  content-language: en_US
```

#### Response

| Name | Type | Description |
|------|------|-------------|
| `code` | integer | Response code |
| `msg` | string | Response message |
| `total` | long | Total number of plants |
| `totalPage` | long | Total number of pages |
| `rows` | array | List of plant objects |
| `rows[].plantId` | string | Plant ID |
| `rows[].plantName` | string | Plant name |
| `rows[].NMI` | string | Electricity meter NMI |
| `rows[].plantNo` | string | Plant number (eleoHome's plantId) |

#### Response Example

```json
{
  "code": 200,
  "msg": "",
  "total": 1,
  "totalPage": 1,
  "rows": [
    {
      "plantId": "10543662677",
      "plantName": "My Solar Plant",
      "NMI": "NMI12345",
      "plantNo": "47TFYB"
    }
  ]
}
```

#### Endpoint-Specific Response Codes

| Code | Description |
|------|-------------|
| `200` | Request success |
| `10001` | System error, please try again later |
| `10006` | Parameter error |

---

## 6. Information API

### 6.1 Get All Devices from Plant

Returns a comprehensive list of all devices in a plant, including inverters, batteries, EMS modules, meters, air conditioners, fire safety, EV chargers, mobile storage, diesel generators, heat pumps, weather stations, smart sockets, and more.

| Property | Value |
|----------|-------|
| **Method** | `GET` |
| **Endpoint** | `/open/api/plant/getPlantAllDeviceList` |
| **Full URL** | `https://intl-developer.saj-electric.com/prod-api/open/api/plant/getPlantAllDeviceList` |

#### Request Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `plantId` | String | YES | Plant ID |
| `userId` | String | YES | User ID (can be blank during third-party authorization) |

#### Response Structure

The top-level response returns `data` as an array of device wrapper objects, each containing a `deviceType` field and a type-specific data object.

| Name | Type | Description |
|------|------|-------------|
| `code` | integer | Response code |
| `msg` | string | Response message |
| `data` | array | Array of device wrapper objects |
| `data[].deviceType` | integer | Device type (see Device Types table in §1) |
| `data[].sn` | string | Device serial number |
| `data[].sort` | integer | Sort order |

**Device-Type-Specific Data Fields:**

Each item in `data[]` contains exactly ONE of the following, based on `deviceType`:

| Field | deviceType | Description |
|-------|-----------|-------------|
| `inverterData` | 1 | Inverter data |
| `monitorData` | 2 | Load monitor data |
| `chargerData` | 3 | EV charger data |
| `mobileStorageBean` | 4 | Mobile storage data |
| `emsModuleData` | 6 | EMS/SEP module data |
| `electricMeterData` | 7 | Electric meter data |
| `airConditioningListBean` | 8 | Air conditioner data |
| `fireFightingListBean` | 9 | Fire safety data |
| `dieselBean` | 10 | Diesel generator data |
| `dryContactData` | 11 | Dry contact data |
| `eddiBean` | 14 | eddi water heater data |
| `collectorBean` | 16/18 | Data logger/monitor (PV optimizer) |
| `heatPumpBean` | 17 | Heat pump data |
| `pvMeteorologicalInstrumentBean` | 19 | Weather station data |
| `smartSocketListBean` | — | Smart socket data |
| `fudaModuleData` | — | FUDA module data |
| `cm2LiquidCoolingBean` | — | CM2 liquid cooling data |

---

#### InverterListBean (`inverterData`)

| Name | Type | Description |
|------|------|-------------|
| `aliases` | string | Bound inverter alias |
| `batEnergyPercent` | string | Battery SOC (%) |
| `batteryDirection` | integer | Battery energy flow: 0=standby, 1=discharge, -1=charge |
| `createDate` | string | Creation time |
| `deviceModel` | string | Device model |
| `devicePic` | string | Device image URL |
| `deviceSn` | string | Inverter serial number |
| `deviceStatus` | string | Status (Grid-tied: 1=standby/2=normal/3=error/4=warning; Storage: 1=standby/2=On-grid/3=off-grid/4=bypass/5=fault/6=upgrade) |
| `enableDelete` | integer | Show delete button (1=yes, 0=no) |
| `enableEdit` | integer | Show edit button (1=yes, 0=no) |
| `enableRemote` | integer | Allow remote config (1=yes, 0=no) |
| `enableShowBatteryRealDataBtn` | integer | Show battery real-time data (1=yes, 0=no) |
| `enableShowDiagnosis` | integer | Show installation diagnosis (1=yes, 0=no) |
| `enableShowSingleVoltageBtn` | integer | Show cell voltage data (1=yes, 0=no) |
| `firstOnlineTime` | string | First data time |
| `hasBattery` | integer | Battery bound (1=yes, 0=no) |
| `ifCMPDevice` | integer | Is CMP device (1=yes, 0=no) |
| `ifInstallPv` | integer | PV installed (0=no, 1=yes) |
| `ifShowAFCI` | integer | Show AFCI data (1=yes, 0=no) |
| `ifShowFCAS` | integer | Show FCAS data (1=yes, 0=no) |
| `ifShowRemoteCloudLink` | integer | Show remote cloud link (1=yes, 0=no) |
| `isDeviceUnderWarranty` | integer | Warranty status (0=expired, 1=under, -1=unknown) |
| `isFlowExpired` | integer | Cellular data expired (0=no, 1=yes) |
| `isMasterFlag` | integer | Show master in parallel (1=yes, 0=no) |
| `moduleSn` | string | Module serial number |
| `monthEnergy` | number | Monthly energy (kWh) |
| `plantName` | string | Plant name |
| `plantUid` | string | Plant UID |
| `powerNow` | number | Current power (W) |
| `runningState` | integer | 1=normal, 2=alarm, 3=offline, 4=historical, 5=not monitored |
| `solarPower` | number | User-defined PV panel power |
| `tableType` | integer | Chart type (0=grid-tied, 1=other hybrid, 2=single-phase H2, 3=three-phase H2) |
| `todayBatChgEnergy` | string | Battery charge today (kWh) |
| `todayBatDisChgEnergy` | string | Battery discharge today (kWh) |
| `todayEnergy` | number | Energy today (kWh) |
| `todayEquivalentHours` | string | Daily equivalent hours |
| `totalEnergy` | number | Total energy (kWh) |
| `totalEquivalentHours` | string | Cumulative equivalent hours |
| `type` | integer | Inverter type (0=grid-tied, 1=storage, 2=AC-coupled) |
| `unitOfCapacity` | string | Capacity unit (Ah or kWh) |
| `yearEnergy` | number | Yearly energy (kWh) |

---

#### PlantEmsModuleBean (`emsModuleData`)

| Name | Type | Description |
|------|------|-------------|
| `availableCapacity` | integer | Available capacity |
| `bluePassword` | string | Module Bluetooth passcode |
| `currentStrategy` | integer | Current control strategy |
| `currentStrategyI18n` | string | Current strategy (i18n) |
| `devicePic` | string | Device image URL |
| `deviceSn` | string | Inverter serial number |
| `emsModel` | string | EMS model |
| `emsModuleName` | string | Device name |
| `emsModulePc` | string | EMS module PC code |
| `emsModuleSn` | string | EMS module serial number |
| `enableBluePassword` | integer | Show Bluetooth passcode (1=yes, 0=no) |
| `enableEditAliases` | integer | Can edit alias (1=yes, 0=no) |
| `enableShowDiagnosis` | integer | Show installation diagnosis (1=yes, 0=no) |
| `firmwareVersion` | string | Firmware version |
| `hardwareVersion` | string | Hardware version |
| `hasRealData` | integer | Has real-time data (1=yes, 0=no) |
| `plantName` | string | Plant name |
| `plantUid` | string | Plant UID |
| `runStatus` | integer | 0=offline, 1=online, 2=not monitored |
| `runTime` | string | System runtime (days) |
| `softwareVersion` | string | Software version |
| `startTime` | string | First power-on time |
| `totalRunTime` | string | Cumulative runtime (days) |

---

#### EmsElectricMeterBean (`electricMeterData`)

| Name | Type | Description |
|------|------|-------------|
| `devicePic` | string | Meter image URL |
| `emsModuleSn` | string | EMS module serial number |
| `meterModel` | string | Meter model |
| `meterName` | string | Meter name |
| `meterSn` | string | Meter serial number |
| `meterType` | integer | 0=grid, 1=export-limiter, 2=storage metering, 3=PV, 4=ATS, 5=export-limiter direct, 6=PV direct, 7=dual meters, 8=SEC |
| `newMeterFlag` | integer | New logic flag (0=old, 1=new) |
| `totalFeedInEnergy` | number | Total import energy |
| `totalFeedInEnergySecondWay` | number | Total import (2nd channel) |
| `totalGridPower` | number | Current power |
| `totalGridPowerSecondWay` | number | Current power (2nd channel) |
| `totalSellEnergy` | number | Total export energy |
| `totalSellEnergySecondWay` | number | Total export (2nd channel) |

---

#### Other Device Beans (Summary)

<details>
<summary><b>ChargerListBean (EV Charger)</b></summary>

| Name | Type | Description |
|------|------|-------------|
| `carAlias` | string | Car/charger alias |
| `chargeEnergy` | number | Charge energy this session (kWh) |
| `chargePower` | number | Charge power (kW) |
| `chargeStatus` | integer | 1=offline, 2=idle, 3=ready, 4=charging, 5=paused (no power), 6=paused (suspended), 7=completed, 8=fault |
| `chargeStatusName` | string | Status name |
| `chargerDeviceSn` | string | Charger SN |
| `chargerType` | integer | 1=gen1, 2=gen2, 3=gen3 |
| `workMode` | integer | 0=off, 1=standard, 2=time-of-use, 3=PV |
| `workModeName` | string | Work mode name |

</details>

<details>
<summary><b>MobileStorageListBean</b></summary>

| Name | Type | Description |
|------|------|-------------|
| `batEnergyPercent` | number | Battery SOC |
| `chargeStatus` | integer | 1=offline, 2=charging, 3=discharging, 4=standby |
| `connectStatus` | integer | 1=Bluetooth online, 2=network online, 3=offline |
| `mobileStorageModel` | string | Model (e.g., PS3600) |
| `mobileStorageSn` | string | Serial number |

</details>

<details>
<summary><b>PlantMeterModuleListBean (Load Monitor)</b></summary>

| Name | Type | Description |
|------|------|-------------|
| `devicePic` | string | Image URL |
| `deviceSnList` | array(string) | Reported inverter SN list |
| `gridDirection` | integer | Grid energy flow (1=export, 0=none, -1=import) |
| `gridPower` | number | Grid power (W) |
| `isOnline` | integer | Online status |
| `moduleSn` | string | Module SN |
| `totalLoadPower` | number | Load power (W) |
| `updateDate` | string | Update time |

</details>

<details>
<summary><b>PlantAirConditioningListBean (Air Conditioner / Fire Safety)</b></summary>

| Name | Type | Description |
|------|------|-------------|
| `acHumidity` | number | AC humidity |
| `acTemp` | number | AC temperature |
| `alarmCount` | integer | Alarm count |
| `coConc` | number | CO concentration |
| `deviceSn` | string | PCS SN |
| `existAc` | integer | AC exists (1=yes, 0=no) |
| `onlineStatus` | integer | 1=online, 0=offline |
| `smokeSensor` | integer | 1=running, 0=stopped |
| `workStatus` | integer | 1=running, 0=stopped |

</details>

<details>
<summary><b>DieselBean (Diesel Generator)</b></summary>

| Name | Type | Description |
|------|------|-------------|
| `dieseName` | string | Diesel generator name |
| `genType` | integer | Bound device type (0=ems, 1=ch2, 2=NA high-voltage, 3=NA low-voltage) |
| `powerWatt` | string | Current power (W) |
| `status` | integer | 0=off, 1=started |
| `todayEnergy` | string | Today's energy |
| `totalEnergy` | string | Total energy |
| `workMode` | string | 0=auto, 1=manual |

</details>

<details>
<summary><b>PhnixHeatPumpBean (Heat Pump)</b></summary>

| Name | Type | Description |
|------|------|-------------|
| `deviceCode` | string | Heat pump device code |
| `deviceName` | string | Device name |
| `deviceSn` | string | Heat pump SN |
| `deviceStatus` | string | ONLINE / OFFLINE / UNACTIVE |
| `heatMode` | string | Heating mode |
| `hpPower` | string | Heat pump power |
| `indoorTemp` | string | Indoor temperature |
| `outdoorTemp` | string | Outdoor temperature |
| `powerStatus` | string | 0=off, 1=on |

</details>

<details>
<summary><b>SmartSocketListBean</b></summary>

| Name | Type | Description |
|------|------|-------------|
| `deviceSn` | string | SN |
| `deviceState` | integer | 1=on, 2=off |
| `power` | number | Current power |
| `status` | integer | 3=online, 4=offline |
| `todayEnergy` | number | Energy today |
| `totalEnergy` | number | Total energy |
| `smartDeviceType` | integer | 1=smart socket, 5=shelly |

</details>

#### Response Example

```json
{
  "code": 200,
  "msg": "request success",
  "data": [
    {
      "deviceType": 6,
      "emsModuleData": {
        "deviceSn": "H2T2103Y2311E00002",
        "emsModel": "eSolar AIO3",
        "emsModuleSn": "M5380Y2302006707",
        "firmwareVersion": "V1.212.11",
        "runStatus": 0
      },
      "sn": "",
      "sort": 0
    },
    {
      "deviceType": 1,
      "inverterData": {
        "aliases": "H2T2103Y2311E00002",
        "batEnergyPercent": "100",
        "deviceSn": "H2T2103Y2311E00002",
        "deviceStatus": "On-grid",
        "powerNow": 0,
        "todayEnergy": 0,
        "totalEnergy": 314.93,
        "monthEnergy": 271.69,
        "yearEnergy": 314.93,
        "type": 1,
        "unitOfCapacity": "kWh"
      },
      "sn": "",
      "sort": 0
    }
  ]
}
```

#### Endpoint-Specific Response Codes

| Code | Description |
|------|-------------|
| `200` | Request successful |
| `10001` | Server exception |
| `10006` | API parameter error |
| `10016` | Device connection timeout |
| `10021` | Data packet exception |

---

### 6.2 Get Basic Information

Retrieve the basic hardware information of a device, including inverter details, module info, and battery BMS data (up to 5 batteries).

| Property | Value |
|----------|-------|
| **Method** | `GET` |
| **Endpoint** | `/open/api/device/baseinfo` |
| **Full URL** | `https://intl-developer.saj-electric.com/prod-api/open/api/device/baseinfo` |

#### Request Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `deviceSn` | String | YES | Device serial number |

#### Request Example

```
GET /prod-api/open/api/device/baseinfo?deviceSn=HSS2602Y2223123456

Headers:
  accessToken: <your_token>
  content-language: en_US
```

#### Response

| Name | Type | Description |
|------|------|-------------|
| `code` | integer | Response code |
| `msg` | string | Response message |
| `data` | object | Device basic info |
| `data.invType` | string | Inverter type |
| `data.invPower` | string | Power of inverter |
| `data.invSN` | string | Device serial number |
| `data.invPC` | string | Inverter PC code |
| `data.invDFW` | string | Display board software version |
| `data.invMFW` | string | Master software version |
| `data.invSFW` | string | Slave software version |
| `data.dispHWVersion` | string | Display board hardware version |
| `data.ctrlHWVersion` | string | Control board hardware version |
| `data.powerHWVersion` | string | Power board hardware version |
| `data.ratedPower` | string | Rated power (kW) |
| `data.moduleModel` | string | Module model name |
| `data.moduleSN` | string | Module SN code |
| `data.modulePC` | string | Module PC code |
| `data.moduleFW` | string | Module firmware version |
| `data.moduleCCID` | string | SIM card CCID |
| `data.moduleIMEI` | string | Module IMEI code |
| `data.batNum` | string | Number of batteries |
| `data.bat1SN` – `data.bat5SN` | string | Battery 1–5 serial numbers |
| `data.bat1Type` – `data.bat5Type` | string | Battery 1–5 types |
| `data.bms1SN` – `data.bms5SN` | string | BMS 1–5 serial numbers |
| `data.bms1Type` – `data.bms5Type` | string | BMS 1–5 types |
| `data.bms1SoftwareVersion` – `data.bms5SoftwareVersion` | string | BMS 1–5 software versions |
| `data.bms1HardwareVersion` – `data.bms5HardwareVersion` | string | BMS 1–5 hardware versions |

#### Response Example

```json
{
  "code": 0,
  "msg": "",
  "data": {
    "invType": "ASP 10KW-3P-X",
    "invPower": "-1.00",
    "invSN": "HSS2602Y2223123456",
    "invPC": "0123456789",
    "invDFW": "v1.042",
    "invMFW": "v0.213",
    "invSFW": "v65.535",
    "dispHWVersion": "v1.100",
    "ctrlHWVersion": "v1.100",
    "powerHWVersion": "v1.100",
    "moduleModel": "eSolar AIO3",
    "moduleSN": "M5380J2312050850",
    "modulePC": "123456789",
    "moduleFW": "v1.500",
    "moduleCCID": "",
    "moduleIMEI": "",
    "batNum": "1",
    "bat1SN": "BAT001",
    "bat1Type": "LFP",
    "bms1SN": "BMS001",
    "bms1SoftwareVersion": "v1.20",
    "bms1HardwareVersion": "v1.04"
  }
}
```

---

### 6.3 Get Device Details Information

Retrieve detailed device and battery information for an inverter.

| Property | Value |
|----------|-------|
| **Method** | `GET` |
| **Endpoint** | `/open/api/device/batInfo` |
| **Full URL** | `https://intl-developer.saj-electric.com/prod-api/open/api/device/batInfo` |

#### Request Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `deviceSn` | String | YES | Device serial number |

#### Response

| Name | Type | Description |
|------|------|-------------|
| `data.deviceInfo` | object | Device information |
| `data.deviceInfo.invType` | string | Type of device |
| `data.deviceInfo.invPower` | string | Power of device (W) |
| `data.deviceInfo.invSN` | string | Device serial number |
| `data.deviceInfo.invPC` | string | Inverter PC code |
| `data.deviceInfo.invDFW` | string | Display board software version |
| `data.deviceInfo.invMFW` | string | Master software version |
| `data.deviceInfo.invSFW` | string | Slave software version |
| `data.deviceInfo.dispHWVersion` | string | Display board hardware version |
| `data.deviceInfo.powerHWVersion` | string | Power board hardware version |
| `data.deviceInfo.moduleModel` | string | Module model name |
| `data.deviceInfo.moduleSN` | string | Module SN code |
| `data.deviceInfo.modulePC` | string | Module PC code |
| `data.deviceInfo.moduleFW` | string | Module firmware version |
| `data.deviceInfo.moduleCCID` | string | SIM card CCID |
| `data.deviceInfo.moduleIMEI` | string | Module IMEI code |
| `data.batteryInfoList` | array | Battery information list |
| `data.batteryInfoList[].batModel` | string | Battery model |
| `data.batteryInfoList[].batModule` | string | Battery module |
| `data.batteryInfoList[].bmsSN` | string | BMS serial number |
| `data.batteryInfoList[].batPC` | string | Battery PC number |
| `data.batteryInfoList[].batType` | string | Battery type |
| `data.batteryInfoList[].bmsSoftwareVersion` | string | BMS software version |
| `data.batteryInfoList[].bmsHardwareVersion` | string | BMS hardware version |
| `data.batteryInfoList[].batSN` | string | Battery serial number |
| `data.batteryInfoList[].ifBatBox` | boolean | Whether it's a battery box |
| `data.batteryInfoList[].numberIndex` | integer | Battery box/battery index number |

#### Response Example

```json
{
  "code": 200,
  "msg": "request success",
  "data": {
    "deviceInfo": {
      "invType": "ASP 10KW-3P-X",
      "invPower": "-1.00",
      "invSN": "Hxxxx03Y22xxxxxxx",
      "invPC": "0xxxxxxxx4",
      "invDFW": "v1.042",
      "invMFW": "v0.213",
      "invSFW": "v65.535",
      "dispHWVersion": "v1.100",
      "ctrlHWVersion": "v1.100",
      "powerHWVersion": "v1.100",
      "moduleModel": "eSolar AIO3",
      "moduleSN": "Mxxxx22xxxxxxx3",
      "modulePC": "1xxxxxxx0",
      "moduleFW": "v1.500",
      "moduleCCID": "",
      "moduleIMEI": ""
    },
    "batteryInfoList": [
      {
        "batModel": "",
        "batModule": "",
        "bmsSN": "",
        "batPC": "",
        "batType": "",
        "bmsSoftwareVersion": "v1.20",
        "bmsHardwareVersion": "v1.04",
        "batSN": "",
        "ifBatBox": true,
        "numberIndex": 1
      }
    ]
  }
}
```

---

## 7. Data API

### 7.1 Realtime Data (Common)

Retrieve the latest real-time telemetry data for a device. Returns comprehensive PV, grid, battery, inverter, and load measurements.

| Property | Value |
|----------|-------|
| **Method** | `GET` |
| **Endpoint** | `/open/api/device/realtimeDataCommon` |
| **Full URL** | `https://intl-developer.saj-electric.com/prod-api/open/api/device/realtimeDataCommon` |
| **Rate Limit** | 150 QPS |

#### Request Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `deviceSn` | String | YES | Device serial number |

#### Response Fields

**Core Fields:**

| Name | Type | Description |
|------|------|-------------|
| `deviceSn` | string | Device serial number |
| `dataTime` | string | Device update time |
| `invTime` | string | Device time (UTC) |
| `moduleSn` | string | Module serial number |
| `isOnline` | string | Online status |
| `mpvMode` | integer | Work mode: 0=Init, 1=Waiting, 2=Grid-connected, 3=Off-grid, 4=Grid load, 5=Fault, 6=Upgrade, 7=Debugging, 8=Self-inspection, 9=Reset |

**PV String Data (PV1–PV16):**

| Name | Type | Description |
|------|------|-------------|
| `pv1volt` – `pv16volt` | string | PV string voltage (V) |
| `pv1curr` – `pv16curr` | string | PV string current (A) |
| `pv1power` – `pv16power` | string | PV string power (W) |

**Battery Data:**

| Name | Type | Description |
|------|------|-------------|
| `batTempC` | string | Battery temperature (°C) |
| `batEnergyPercent` | string | Battery SOC (%) |
| `batPower` | string | Battery power (W) |
| `totalBatteryPower` | string | Current battery power (W) |
| `batteryDirection` | integer | Battery direction: 1=discharge, -1=charge |
| `batteryGroupDataList` | array | Battery pack information |
| `currentMaxBatPoweLimitSupport` | boolean | Supports max charge/discharge power limit calculation |
| `currentMaxChargePowerLimit` | string | Max allowed charge power limit (-1=invalid) |
| `currentMaxDisChargePowerLimit` | string | Max allowed discharge power limit (-1=invalid) |

**Grid Data (Three-Phase R/S/T):**

| Name | Type | Description |
|------|------|-------------|
| `rGridVolt` / `sGridVolt` / `tGridVolt` | string | Grid voltage per phase (V) |
| `rGridCurr` / `sGridCurr` / `tGridCurr` | string | Grid current per phase (A) |
| `rGridFreq` / `sGridFreq` / `tGridFreq` | string | Grid frequency per phase (Hz) |
| `rGridPowerWatt` / `sGridPowerWatt` / `tGridPowerWatt` | string | Grid active power per phase (W) |
| `rGridPowerPF` | number | Grid power factor (R phase) |
| `gridDirection` | integer | 1=sell, 0=no flow, -1=feed in |

**Inverter Output (Three-Phase R/S/T):**

| Name | Type | Description |
|------|------|-------------|
| `rInvVolt` / `sInvVolt` / `tInvVolt` | string | Inverter voltage per phase (V) |
| `rInvCurr` / `sInvCurr` / `tInvCurr` | string | Inverter current per phase (A) |
| `rInvFreq` / `sInvFreq` / `tInvFreq` | string | Inverter frequency per phase (Hz) |
| `rInvPowerWatt` / `sInvPowerWatt` / `tInvPowerWatt` | string | Inverter active power per phase (W) |

**Backup Output (Three-Phase R/S/T):**

| Name | Type | Description |
|------|------|-------------|
| `rOutVolt` / `sOutVolt` / `tOutVolt` | string | Output voltage per phase (V) |
| `rOutCurr` / `sOutCurr` / `tOutCurr` | string | Output current per phase (A) |
| `rOutFreq` / `sOutFreq` / `tOutFreq` | string | Output frequency per phase (Hz) |
| `rOutPowerVA` / `sOutPowerVA` / `tOutPowerVA` | string | Output apparent power per phase (VA) |
| `rOutPowerWatt` / `sOutPowerWatt` / `tOutPowerWatt` | string | Output active power per phase (W) |

**Meter Data:**

| Name | Type | Description |
|------|------|-------------|
| `meterAStatus` | string | Meter A status (1=has meter, 0=no meter) |
| `meterAVolt1` – `meterAVolt3` | string | Meter A voltage per phase (V) |
| `meterACurr1` – `meterACurr3` | string | Meter A current per phase (A) |
| `meterAPowerWatt1` – `meterAPowerWatt3` | string | Meter A active power per phase (W) |
| `meterAPowerVA1` – `meterAPowerVA3` | string | Meter A apparent power per phase (VA) |
| `meterAFreq1` – `meterAFreq3` | string | Meter A frequency per phase (Hz) |

**Energy Totals:**

| Name | Type | Description |
|------|------|-------------|
| `totalPVPower` | string | Current PV power (W) — for H series hybrid |
| `totalLoadPowerWatt` | string | Total load power (W) |
| `totalGridPowerWatt` | string | Total grid power (W) — for R series on-grid |
| `backupTotalLoadPowerWatt` | string | Backup total load power (W) |
| `sysGridPowerWatt` | string | System grid power (W) |
| `sysTotalLoadWatt` | string | System total load (W) |
| `todayPvEnergy` | string | Today's PV energy (kWh) |
| `totalPvEnergy` | string | Total PV energy (kWh) |
| `todayLoadEnergy` | string | Today's load energy (kWh) |
| `todaySellEnergy` | string | Today's sell energy (kWh) |
| `todayFeedInEnergy` | string | Today's feed-in energy (kWh) |
| `todayBatChgEnergy` | string | Today's battery charge energy (kWh) |
| `todayBatDisEnergy` | string | Today's battery discharge energy (kWh) |
| `totalFeedInEnergy` | string | Total feed-in energy (kWh) |
| `totalTotalLoadEnergy` | string | Total load energy (kWh) |
| `totalBatChgEnergy` | string | Total battery charge energy (kWh) |
| `totalBatDisEnergy` | string | Total battery discharge energy (kWh) |
| `totalSellEnergy` | string | Total sell energy (kWh) |
| `parallTotalPVMeterEnergy` | string | Total PV meter energy (kWh) |

**System & Environmental:**

| Name | Type | Description |
|------|------|-------------|
| `sinkTempC` | number | Radiator temperature (°C) |
| `ambTempC` | number | Ambient temperature (°C) |
| `invTempC` | number | Inverter temperature (°C) |
| `linkSignal` | integer | Link signal strength (dBm) |
| `parallelEnable` | integer | Parallel function enabled (0=No, 1=Yes) |
| `parallelMaster` | integer | Current device is parallel master |

**DNSP Fields:**

| Name | Type | Description |
|------|------|-------------|
| `registerDnsp` | boolean | Registered with DNSP |
| `dnspExportLimitPower` | string | DNSP current grid export power limit |

#### Response Example (Abbreviated)

```json
{
  "errCode": "0",
  "errMsg": "",
  "data": {
    "deviceSn": "HSS2502J2351E34643",
    "dataTime": "2024-11-08 16:40:00",
    "invTime": "2024-11-08T05:40Z",
    "isOnline": "1",
    "mpvMode": 2,
    "batEnergyPercent": "100",
    "totalBatteryPower": 171,
    "batteryDirection": 1,
    "totalPVPower": "271",
    "totalLoadPowerWatt": "443",
    "totalGridPowerWatt": "283",
    "gridDirection": 0,
    "rGridVolt": "244.10",
    "rGridCurr": "1.64",
    "rGridFreq": "50.01",
    "rGridPowerWatt": "242",
    "pv1volt": "0.00",
    "pv1curr": "0.00",
    "todayPvEnergy": 17.67,
    "totalPvEnergy": 2153.85,
    "todaySellEnergy": 3.75,
    "totalSellEnergy": 473.89,
    "sinkTempC": 36.9,
    "ambTempC": 45.5,
    "linkSignal": -48
  }
}
```

---

### 7.2 EMS Real-Time Data

Retrieve the latest real-time data from an EMS (Energy Management System) device with parallel system data.

| Property | Value |
|----------|-------|
| **Method** | `GET` |
| **Endpoint** | `/open/api/device/emsRealtimeDataCommon` |
| **Full URL** | `https://intl-developer.saj-electric.com/prod-api/open/api/device/emsRealtimeDataCommon` |
| **Rate Limit** | 10 QPS |

#### Request Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `emsSn` | String | YES | EMS serial number |

#### Response Fields

**Basic Information:**

| Name | Type | Description |
|------|------|-------------|
| `deviceSn` | string | Device SN |
| `plantUid` | string | Plant UID |
| `invTime` | string | Device time |
| `dataTime` | string | Data time |
| `updateDate` | string | Data update time |
| `invType` | integer | Model code (59=emanager) |

**Real-Time Parallel Data:**

| Name | Type | Description |
|------|------|-------------|
| `parallSOC` | BigDecimal | Parallel SOC (%) |
| `parallPVPower` | BigDecimal | Parallel PV total power (W) |
| `parallGridPower` | BigDecimal | Parallel grid total power (W) |
| `parallLoadPower` | BigDecimal | Parallel total load (W) |
| `parallBatPower` | BigDecimal | Parallel battery total power (W) |
| `parallMeterPower` | BigDecimal | PV inverter meter power (W) |
| `parallBackupPower` | BigDecimal | Parallel backup total load (W) |
| `parallOngridPower` | BigDecimal | Parallel grid-side total load (W) |
| `parallTotalPvPower` | BigDecimal | Parallel total PV power (W) |
| `parallTotalCap` | BigDecimal | Parallel total battery capacity |
| `parallUserCap` | BigDecimal | Parallel total available capacity |

**Today's Parallel Energy:**

| Name | Type | Description |
|------|------|-------------|
| `parallTodayPVEnergy` | BigDecimal | Today's parallel PV energy (kWh) |
| `parallTodaySellEnergy` | BigDecimal | Today's parallel sell energy (kWh) |
| `parallTodayFeedInEnergy` | BigDecimal | Today's parallel feed-in energy (kWh) |
| `parallTodayBatChgEnergy` | BigDecimal | Today's parallel battery charge energy (kWh) |
| `parallTodayBatDisEnergy` | BigDecimal | Today's parallel battery discharge energy (kWh) |
| `parallTodayTotalLoadEnergy` | BigDecimal | Today's parallel total load energy (kWh) |

**Monthly Parallel Energy:**

| Name | Type | Description |
|------|------|-------------|
| `parallMonthBatChgEnergy` | BigDecimal | Monthly battery charge energy (kWh) |
| `parallMonthBatDisEnergy` | BigDecimal | Monthly battery discharge energy (kWh) |
| `parallMonthPVEnergy` | BigDecimal | Monthly PV energy (kWh) |
| `parallMonthTotalLoadEnergy` | BigDecimal | Monthly total load energy (kWh) |
| `parall_Month_FeedInEnergy` | BigDecimal | Monthly feed-in energy (kWh) |
| `parall_Month_SellEnergy` | BigDecimal | Monthly sell energy (kWh) |

**Annual Parallel Energy:**

| Name | Type | Description |
|------|------|-------------|
| `parallYearFeedInEnergy` | BigDecimal | Year feed-in energy (kWh) |
| `parallYearSellEnergy` | BigDecimal | Year sell energy (kWh) |
| `parallYearBatChgEnergy` | BigDecimal | Year battery charge energy (kWh) |
| `parallYearBatDisEnergy` | BigDecimal | Year battery discharge energy (kWh) |
| `parallYearPVEnergy` | BigDecimal | Year PV generation energy (kWh) |
| `parallYearTotalLoadEnergy` | BigDecimal | Year total load energy (kWh) |

**Lifetime Totals:**

| Name | Type | Description |
|------|------|-------------|
| `parallTotalPVEnergy` | BigDecimal | Total PV generation energy (kWh) |
| `parallTotalSellEnergy` | BigDecimal | Total sell energy (kWh) |
| `parallTotalFeedInEnergy` | BigDecimal | Total feed-in energy (kWh) |
| `parallTotalTotalLoadEnergy` | BigDecimal | Total load energy (kWh) |
| `parallTotalBatChgEnergy` | BigDecimal | Total battery charge energy (kWh) |
| `parallTotalBatDisEnergy` | BigDecimal | Total battery discharge energy (kWh) |

**Status & Direction:**

| Name | Type | Description |
|------|------|-------------|
| `batteryDirection` | integer | Battery energy flow direction |
| `gridDirection` | integer | Grid direction: 1=selling, 0=no flow, -1=buying |

**DNSP Fields:**

| Name | Type | Description |
|------|------|-------------|
| `registerDnsp` | boolean | Registered to DNSP |
| `dnspExportLimitPower` | integer | DNSP current grid export power limit |
| `currentMaxBatPoweLimitSupport` | boolean | Supports max charge/discharge limit |
| `currentMaxChargePowerLimit` | BigDecimal | Max allowed charge power limit (-1=invalid) |
| `currentMaxDisChargePowerLimit` | BigDecimal | Max allowed discharge power limit (-1=invalid) |

**EMS Status Fields:**

| Name | Type | Description |
|------|------|-------------|
| `di0IoValue` | integer | DI0 interface IO status value |
| `di1IoValue` | integer | DI1 interface IO status value |
| `di2IoValue` | integer | DI2 interface IO status value |
| `rcrIoValue` | integer | RCR interface IO status value |
| `drmIoValue` | integer | DRM interface IO status value |
| `signal4G` | string | 4G signal strength |
| `wifiSignal` | string | WiFi signal strength |
| `emsAlarmInfo` | long | eManager alarm information |
| `emsAlarmStatus` | long | C&I EMS module alarm status |
| `emsAlarmStatus1` | long | C&I EMS module alarm status 1 |

#### Response Example

```json
{
  "data": {
    "deviceSn": "M5560G2405001314",
    "plantUid": "9E6D85DE021447FF9F5C6DCB6960AAE1",
    "dataTime": "2025-09-30 10:28:00",
    "invTime": "2025-09-30T05:58Z",
    "updateDate": "2025-09-30 10:28:02",
    "invType": 59,
    "parallSOC": 3.1,
    "parallPVPower": 0,
    "parallGridPower": -12,
    "parallLoadPower": 24,
    "parallBatPower": 12,
    "parallMeterPower": -4,
    "parallBackupPower": 0,
    "parallOngridPower": 26,
    "parallTotalCap": 10,
    "batteryDirection": 1,
    "gridDirection": -1,
    "parallTodayPVEnergy": 0,
    "parallTodaySellEnergy": 0,
    "parallTodayFeedInEnergy": 0.456,
    "parallTodayBatChgEnergy": 0.27,
    "parallTodayBatDisEnergy": 0.13,
    "parallTodayTotalLoadEnergy": 0.316,
    "parallMonthBatChgEnergy": 28.59,
    "parallMonthBatDisEnergy": 24.56,
    "parallTotalPVEnergy": 0,
    "parallTotalSellEnergy": 12.224,
    "parallTotalFeedInEnergy": 26.252,
    "parallTotalBatChgEnergy": 28.59,
    "parallTotalBatDisEnergy": 24.56
  },
  "errCode": 0,
  "errMsg": ""
}
```

---

### 7.3 Get Device Upload Data

Retrieve historical upload data reported by a device at minute, day, month, or year granularity. Response format varies based on `deviceType` and `timeUnit`.

| Property | Value |
|----------|-------|
| **Method** | `GET` |
| **Endpoint** | `/open/api/device/uploadData` |
| **Full URL** | `https://intl-developer.saj-electric.com/prod-api/open/api/device/uploadData` |
| **Rate Limit** | 50 QPS |

#### Request Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `deviceSn` | String | YES | Device serial number |
| `startTime` | String | YES | Start time `yyyy-MM-dd HH:mm:ss` |
| `endTime` | String | YES | End time `yyyy-MM-dd HH:mm:ss` |
| `timeUnit` | Integer | YES | Granularity: 0=minute, 1=day, 2=month, 3=year |

#### Request Example

```
GET /prod-api/open/api/device/uploadData?deviceSn=HSS2602Y2223123456&startTime=2021-01-04 00:00:00&endTime=2022-07-11 20:00:00&timeUnit=3

Headers:
  accessToken: <your_token>
  content-language: en_US
```

#### Response Structure

| Name | Type | Description |
|------|------|-------------|
| `code` | integer | Response code |
| `msg` | string | Response message |
| `data.deviceType` | integer | Device type: 0=grid, 1=storage, 2=AC coupling |
| `data.timeUnit` | integer | Time unit: 0=minute, 1=day, 2=month, 3=year |
| `data.total` | object | Data aggregation totals |
| `data.data` | array | Detailed time-series data |

#### Response Variants by deviceType and timeUnit

**timeUnit=0, deviceType=0 (Grid-Tied, Minute Data):**

```json
{
  "data": {
    "deviceType": 0,
    "timeUnit": 0,
    "total": null,
    "data": [
      {
        "dataTime": "2022-08-11 13:30:00",
        "l1Volt": 229.6,
        "l2Volt": 231.3,
        "l3Volt": 231.3,
        "power": 0,
        "pv1Curr": 0, "pv1Power": 0, "pv1Volt": 770,
        "pv2Curr": 0, "pv2Power": 0, "pv2Volt": 769.3,
        "pv3Curr": 0, "pv3Power": 0, "pv3Volt": 771.5,
        "todayEnergy": 0
      }
    ]
  },
  "errCode": 0,
  "errMsg": ""
}
```

**timeUnit=1/2/3, deviceType=0 (Grid-Tied, Day/Month/Year Data):**

```json
{
  "data": {
    "deviceType": 0,
    "timeUnit": 1,
    "total": null,
    "data": [
      {
        "dataTime": "2021-08-11 00:00:00",
        "pVEnergy": 11.89
      }
    ]
  },
  "errCode": 0,
  "errMsg": ""
}
```

**timeUnit=0, deviceType=1/2 (Storage/AC-Coupled, Minute Data):**

```json
{
  "data": {
    "deviceType": 1,
    "timeUnit": 0,
    "total": {
      "buyRate": 0,
      "loadSelfConsumedRate": 1,
      "pvSelfConsumedRate": 1,
      "sellRate": 0,
      "totalBuyEnergy": 0,
      "totalLoad": 185589.36,
      "totalPVEnergy": 185589,
      "totalPVSelfConsumption": 185589,
      "totalSelfConsumption": 185589.36,
      "totalSellEnergy": 0
    },
    "data": [
      {
        "batterySOC": 100,
        "buyPower": 15,
        "chargePower": 69,
        "dataTime": "2022-08-21 09:55:00",
        "dischargePower": 0,
        "loadPower": 119,
        "pvPower": 173,
        "selfUsePower": 104,
        "sellPower": 0
      }
    ]
  },
  "errCode": 0,
  "errMsg": ""
}
```

**timeUnit=1/2/3, deviceType=1/2 (Storage/AC-Coupled, Day/Month/Year Data):**

```json
{
  "data": {
    "deviceType": 1,
    "timeUnit": 1,
    "total": {
      "buyRate": 0,
      "loadSelfConsumedRate": 1,
      "pvSelfConsumedRate": 1,
      "sellRate": 0,
      "totalBuyEnergy": 0,
      "totalLoad": 129.27,
      "totalPVEnergy": 119.02,
      "totalPVSelfConsumption": 119.02,
      "totalSelfConsumption": 129.27,
      "totalSellEnergy": 0
    },
    "data": [
      {
        "buyEnergy": 0,
        "dataTime": "2022-08-21 00:00:00",
        "loadEnergy": 185589.36,
        "pVEnergy": 185589,
        "selfConsumption": 185589,
        "sellEnergy": 0
      }
    ]
  },
  "errCode": 0,
  "errMsg": ""
}
```

---

### 7.4 Get History Data (Common)

Retrieve history data for a device (inverter) with full telemetry across all phases.

| Property | Value |
|----------|-------|
| **Method** | `GET` |
| **Endpoint** | `/open/api/device/historyDataCommon` |
| **Full URL** | `https://intl-developer.saj-electric.com/prod-api/open/api/device/historyDataCommon` |
| **Rate Limit** | 5 QPS |

#### Request Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `deviceSn` | String | YES | Device serial number |
| `startTime` | String | YES | Start time `yyyy-MM-dd HH:mm:ss` (max interval: 24 hours) |
| `endTime` | String | YES | End time `yyyy-MM-dd HH:mm:ss` (max interval: 24 hours) |

#### Response Fields

Returns an array of data points. Each point contains extensive three-phase electrical measurements (R/S/T for grid, inverter, and output), PV string data (PV1–PV16), battery data, energy totals, and system status.

The response fields are identical to the [Realtime Data](#71-realtime-data-common) fields, with data returned as time-series arrays.

Key fields include all of: `pv1volt`–`pv16volt`, `pv1curr`–`pv16curr`, `pv1power`–`pv16power`, `rGridVolt`/`sGridVolt`/`tGridVolt`, `rInvVolt`/`sInvVolt`/`tInvVolt`, `rOutVolt`/`sOutVolt`/`tOutVolt`, battery data, meter data, energy totals, `batteryGroupDataList`, `faultMsgList`, etc.

#### Response Example

```json
{
  "code": 0,
  "msg": "",
  "data": [
    {
      "deviceSn": "HSS2602Y2223123456",
      "moduleSn": "M5380J2312050850",
      "dataTime": "2022-07-04 10:00:00",
      "batEnergyPercent": "85",
      "batPower": "500",
      "totalPVPower": "3500",
      "totalLoadPowerWatt": "2000",
      "rGridVolt": "230.5",
      "rGridCurr": "5.2",
      "rGridPowerWatt": "1200",
      "pv1volt": "350.0",
      "pv1curr": "10.0",
      "pv1power": "3500",
      "gridDirection": 1,
      "batteryDirection": -1,
      "todayPvEnergy": "12.5",
      "totalPvEnergy": "5000.0",
      "mpvMode": 2
    }
  ]
}
```

---

### 7.5 Get EMS History Data

Retrieve historical data for EMS devices with parallel system energy metrics.

| Property | Value |
|----------|-------|
| **Method** | `GET` |
| **Endpoint** | `/open/api/device/emsHistoryData` |
| **Full URL** | `https://intl-developer.saj-electric.com/prod-api/open/api/device/emsHistoryData` |

#### Request Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `startTime` | String | YES | Start time `yyyy-MM-dd HH:mm:ss` (max interval: 24 hours) |
| `endTime` | String | YES | End time `yyyy-MM-dd HH:mm:ss` (max interval: 24 hours) |
| `emsSn` | String | YES | EMS serial number |
| `plantId` | String | YES | Plant ID |

#### Response Fields

Returns the same parallel energy fields as the [EMS Real-Time Data](#72-ems-real-time-data) endpoint, plus annual breakdown fields:

| Name | Type | Description |
|------|------|-------------|
| `yearBackupLoadEnergy` | BigDecimal | Annual backup load consumption |
| `yearBatChgEnergy` | BigDecimal | Annual battery charging energy |
| `yearBatDisEnergy` | BigDecimal | Annual battery discharging energy |
| `yearGridConsumpEnergy` | BigDecimal | Annual grid consumption energy |
| `yearFeedInEnergy` | BigDecimal | Annual system feed-in energy |
| `yearGridFeedInEnergy` | BigDecimal | Annual grid feed-in energy |
| `yearGridFeedInBatEnergy` | BigDecimal | Annual battery feed-in to grid energy |
| `yearGridFeedInPVEnergy` | BigDecimal | Annual PV feed-in to grid energy |
| `yearPVEnergy` | BigDecimal | Annual PV generation energy |
| `yearSellEnergy` | BigDecimal | Annual system sell energy |
| `yearTotalLoadEnergy` | BigDecimal | Annual total load consumption |

#### Response Example

```json
{
  "code": 200,
  "msg": "request success",
  "data": [
    {
      "deviceSn": "M5560G2405000005",
      "plantUid": "2E9800A878C644CF8B6E8EDD5275C29D",
      "dataTime": "2025-09-03 10:23:00",
      "invTime": "2025-09-03T05:53Z",
      "updateDate": "2025-09-03 10:23:04",
      "invType": 59,
      "parallSOC": "0",
      "parallPVPower": "0",
      "parallGridPower": "2622",
      "parallLoadPower": "356",
      "parallBatPower": "0",
      "parallTodayPVEnergy": "4.49",
      "parallTodaySellEnergy": "4.16",
      "parallTodayFeedInEnergy": "0.24",
      "parallTotalPVEnergy": "4239.55",
      "parallTotalSellEnergy": "4079.92",
      "parallTotalFeedInEnergy": "1157.72",
      "batteryDirection": 0,
      "gridDirection": 1,
      "yearPVEnergy": "0",
      "yearSellEnergy": "0",
      "yearFeedInEnergy": "0"
    }
  ]
}
```

---

### 7.6 EMS Meter Historical Data

Retrieve detailed historical data from EMS-connected meters with dual-channel support.

| Property | Value |
|----------|-------|
| **Method** | `GET` |
| **Endpoint** | `/open/api/device/emsHistoryData4Meter` |
| **Full URL** | `https://intl-developer.saj-electric.com/prod-api/open/api/device/emsHistoryData4Meter` |

#### Request Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `startTime` | String | YES | Start time `yyyy-MM-dd HH:mm:ss` (max interval: 2 hours) |
| `endTime` | String | YES | End time `yyyy-MM-dd HH:mm:ss` (max interval: 2 hours) |
| `emsSn` | String | YES | EMS serial number |
| `plantId` | String | YES | Plant ID |

#### Response Fields (First Channel)

| Name | Type | Description |
|------|------|-------------|
| `sn` | string | Meter SN |
| `dataTime` | date | Data timestamp |
| `meterType` | string | Meter model |
| `currRatio` | BigDecimal | Current transformer ratio |
| `wiringMode` | integer | 0=three-phase four-wire, 1=three-phase three-wire |
| `phase1Volt` / `phase2Volt` / `phase3Volt` | BigDecimal | Phase voltage (V) |
| `phase1Curr` / `phase2Curr` / `phase3Curr` | BigDecimal | Phase current (A) |
| `freq` | BigDecimal | Total frequency (Hz) |
| `freqA` / `freqB` / `freqC` | BigDecimal | Per-phase frequency (Hz) |
| `totalGridPower` | BigDecimal | Total active power (three-phase) (W) |
| `phase1Power` / `phase2Power` / `phase3Power` | BigDecimal | Phase active power (W) |
| `totalQpower` | BigDecimal | Total reactive power (VAR) |
| `phase1Qpower` / `phase2Qpower` / `phase3Qpower` | BigDecimal | Phase reactive power (VAR) |
| `totalPowerfactor` | BigDecimal | Total power factor |
| `phase1Powerfactor` / `phase2Powerfactor` / `phase3Powerfactor` | BigDecimal | Phase power factor |
| `impEp` | BigDecimal | Total imported energy (forward active energy) |
| `expEp` | BigDecimal | Total exported energy (reverse active energy) |
| `todayImpEp` | string | Today's imported energy |
| `todayExpEp` | string | Today's exported energy |
| `todayImpEpA` / `todayImpEpB` / `todayImpEpC` | string | Today's per-phase imported energy |
| `todayExpEpA` / `todayExpEpB` / `todayExpEpC` | string | Today's per-phase exported energy |
| `firstWayEnable` | integer | First channel has data (null=invalid, 0=no, 1=yes) — Grid meter |
| `secondWayEnable` | integer | Second channel has data (null=invalid, 0=no, 1=yes) — PV meter |

**Second Channel Fields:**

All first-channel fields are duplicated with a `SecondWay` suffix (e.g., `phase1VoltSecondWay`, `totalGridPowerSecondWay`, `impEpSecondWay`, etc.).

#### Response Example

```json
[
  {
    "sn": "METER123456",
    "dataTime": "2022-07-04T10:00:00",
    "meterType": "SAJ-EM100",
    "currRatio": 100,
    "wiringMode": 0,
    "phase1Volt": 230.5,
    "phase2Volt": 229.8,
    "phase3Volt": 231.0,
    "phase1Curr": 5.2,
    "phase2Curr": 4.9,
    "phase3Curr": 5.1,
    "freq": 50.02,
    "totalGridPower": 3.6,
    "impEp": 12345.67,
    "expEp": 890.12,
    "todayImpEp": "23.45",
    "todayExpEp": "12.34",
    "firstWayEnable": 1,
    "secondWayEnable": 1
  }
]
```

---

### 7.7 Query Energy Flow Data of Device in Plant

> **Note:** This endpoint is listed in the API menu. Please refer to SAJ platform documentation for specific parameters and response details if not fully covered in the source material.

---

### 7.8 Query Details of Plant

Retrieve comprehensive information about a plant, including location, configuration, and device lists.

| Property | Value |
|----------|-------|
| **Method** | `GET` |
| **Endpoint** | `/open/api/plant/details` |
| **Full URL** | `https://intl-developer.saj-electric.com/prod-api/open/api/plant/details` |

#### Request Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `plantId` | String | YES | Plant ID |

#### Request Example

```
GET /prod-api/open/api/plant/details?plantId=10543662677

Headers:
  accessToken: <your_token>
  content-language: en_US
```

#### Response

| Name | Type | Description |
|------|------|-------------|
| `plantUid` | string | Plant UID |
| `plantName` | string | Plant name |
| `plantNo` | string | Plant code |
| `type` | integer | Plant type: 0=grid, 1=energy storage, 3=AC coupling |
| `useType` | integer | Use type: 0=household, 1=factory roof, 2=ground station, 3=poverty alleviation |
| `systemPower` | BigDecimal | Peak power of plant (kWp) |
| `gridPrice` | BigDecimal | Electricity price |
| `currencyName` | string | Currency unit |
| `currency` | string | Currency symbol |
| `timeZone` | string | Time zone ID |
| `timeZoneName` | string | Time zone name |
| `countryCode` | string | Country code |
| `country` | string | Country name |
| `provinceCode` | string | Province code |
| `province` | string | Province name |
| `cityCode` | string | City code |
| `city` | string | City name |
| `countyCode` | string | County code |
| `county` | string | County name |
| `street` | string | Street name |
| `streetType` | string | Street type |
| `address` | string | Address |
| `fullAddress` | string | Full address |
| `latitude` | BigDecimal | Latitude |
| `longitude` | BigDecimal | Longitude |
| `zipCode` | string | Postal code |
| `meterId` | string | Meter ID (NMI) |
| `moduleNum` | integer | Number of modules |
| `gridNetType` | string | Grid connection type: 1=full internet, 2=self-use balance, 3=off-internet |
| `payMode` | string | Investment mode: 1=full payment, 2=loan, 3=self-invested, 4=joint venture |
| `pvPanelAzimuth` | string | PV panel azimuth (°) |
| `pvPanelAngle` | string | PV panel inclination (°) |
| `phone` | string | Owner's phone number |
| `email` | string | Owner's email |
| `projectPic` | string | Plant picture URL |
| `isInstallMeter` | integer | Load monitoring installed (0=no, 1=yes) |
| `secModuleIfNewVersion` | integer | New load monitor (0=no, 1=yes) |
| `isShared` | integer | Shared plant (0=no, 1=yes) |
| `deviceSnList` | array | Device serial numbers in plant |
| `moduleSnList` | array | Module serial numbers in plant |

#### Response Example

```json
{
  "code": 200,
  "msg": "request success",
  "data": {
    "plantUid": "85DE398EDFC545FC99351F72C3381CF2",
    "plantName": "My Solar Plant",
    "plantNo": "47TFYB",
    "type": 0,
    "systemPower": 189,
    "country": "China",
    "countryCode": "CN",
    "province": "guangdong",
    "city": "guangzhou",
    "fullAddress": "guangdong guangzhou huangpu",
    "latitude": 12,
    "longitude": 21,
    "timeZone": "PRC",
    "timeZoneName": "(UTC+08:00) Beijing, Chongqing, Hong Kong, Urumqi",
    "gridPrice": 6,
    "deviceSnList": [
      "ZP033K0011670079",
      "ZP033K0011670139"
    ],
    "moduleSnList": [],
    "isInstallMeter": 0,
    "isShared": 0,
    "projectPic": "https://esolar-images.jtaiyang.com/plant/defaultPlantCover.png"
  }
}
```

---

### 7.9 Query Energy of Plant

Retrieve current energy generation summary for a plant.

| Property | Value |
|----------|-------|
| **Method** | `GET` |
| **Endpoint** | `/open/api/plant/energy` |
| **Full URL** | `https://intl-developer.saj-electric.com/prod-api/open/api/plant/energy` |

#### Request Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `plantId` | String | YES | Plant ID |
| `deviceSns` | String | YES | Comma-separated device serial numbers |
| `clientDate` | String | YES | Client datetime `yyyy-MM-dd HH:mm:ss` |

#### Request Example

```
GET /prod-api/open/api/plant/energy?deviceSns=ZP033K0011670079,ZP033K0011670139&plantId=10543662677&clientDate=2022-10-12 11:27:17

Headers:
  accessToken: <your_token>
  content-language: en_US
```

#### Response

| Name | Type | Description |
|------|------|-------------|
| `updateDate` | string | Update time of reported data |
| `powerNow` | BigDecimal | Current power (W) |
| `todayPvEnergy` | BigDecimal | PV energy today (kWh) |
| `monthPvEnergy` | BigDecimal | PV energy this month (kWh) |
| `yearPvEnergy` | BigDecimal | PV energy this year (kWh) |
| `totalPvEnergy` | BigDecimal | Total PV energy (kWh) |
| `batEnergyPercent` | BigDecimal | Battery SOC (%) |
| `deviceStatus` | integer | Device status |

#### Response Example

```json
{
  "code": 200,
  "msg": "request success",
  "data": {
    "batEnergyPercent": 0,
    "deviceStatus": 0,
    "monthPvEnergy": 0,
    "powerNow": 0,
    "todayPvEnergy": 0,
    "totalPvEnergy": 0,
    "updateDate": null,
    "yearPvEnergy": 0
  }
}
```

---

### 7.10 Query Plant Statistics Data

Retrieve comprehensive statistics for a plant including generation, consumption, battery, and grid exchange data.

| Property | Value |
|----------|-------|
| **Method** | `GET` |
| **Endpoint** | `/open/api/plant/getPlantStatisticsData` |
| **Full URL** | `https://intl-developer.saj-electric.com/prod-api/open/api/plant/getPlantStatisticsData` |

#### Request Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `plantId` | String | YES | Plant (power station) ID |
| `clientDate` | String | YES | Client date `yyyy-MM-dd HH:mm:ss` |

#### Response

| Name | Type | Description |
|------|------|-------------|
| `plantUid` | string | Plant UID |
| `deviceSn` | string | Device SN |
| `type` | integer | Plant type: 0=Grid-connected, 1=Energy Storage, 3=AC Coupled |
| `plantName` | string | Plant name |
| `powerNow` | BigDecimal | Current power (W) |
| `deviceStatus` | integer | Inverter status |
| `batEnergyPercent` | BigDecimal | Battery remaining energy (SOC %) |
| `todayPvEnergy` | BigDecimal | PV generation today (kWh) |
| `monthPvEnergy` | BigDecimal | PV generation monthly (kWh) |
| `yearPvEnergy` | BigDecimal | PV generation yearly (kWh) |
| `totalPvEnergy` | BigDecimal | PV generation total (kWh) |
| `todayLoadEnergy` | BigDecimal | Load consumption today (kWh) |
| `monthLoadEnergy` | BigDecimal | Load consumption monthly (kWh) |
| `yearLoadEnergy` | BigDecimal | Load consumption yearly (kWh) |
| `totalLoadEnergy` | BigDecimal | Load consumption total (kWh) |
| `monthBuyEnergy` | BigDecimal | Grid purchase monthly (kWh) |
| `yearBuyEnergy` | BigDecimal | Grid purchase yearly (kWh) |
| `dataTime` | string | Data reporting time |
| `updateDate` | string | Update time |
| `totalPlantTreeNum` | BigDecimal | Total equivalent trees planted |
| `totalReduceCo2` | BigDecimal | Total CO2 reduction |
| `co2UnitOfWeight` | string | CO2 weight unit |
| `deviceSnList` | array | List of device SNs |
| `moduleSnList` | array | List of module SNs |
| `unitOfCapacity` | string | Capacity unit (Ah/kWh) |
| `usableBatCapacity` | BigDecimal | Usable battery capacity |
| `todayChargeEnergy` | BigDecimal | Charge energy today (kWh) |
| `todayDisChargeEnergy` | BigDecimal | Discharge energy today (kWh) |
| `totalChargeEnergy` | BigDecimal | Total charge energy (kWh) |
| `totalDisChargeEnergy` | BigDecimal | Total discharge energy (kWh) |
| `todayBuyEnergy` | BigDecimal | Grid purchase today (kWh) |
| `todaySellEnergy` | BigDecimal | Grid sell today (kWh) |
| `totalBuyEnergy` | BigDecimal | Total grid purchase (kWh) |
| `totalSellEnergy` | BigDecimal | Total grid sell (kWh) |
| `monthSellEnergy` | BigDecimal | Grid sell monthly (kWh) |
| `yearSellEnergy` | BigDecimal | Grid sell yearly (kWh) |
| `todayEquivalentHours` | string | Equivalent hours today |

#### Response Example

```json
{
  "code": 200,
  "msg": "request success",
  "data": {
    "plantUid": "F0387466-8E1D-4E51-959F28A89",
    "plantName": "My Plant",
    "type": 0,
    "powerNow": "128.0",
    "deviceStatus": 2,
    "batEnergyPercent": "0",
    "todayPvEnergy": "42.82",
    "monthPvEnergy": "805.78",
    "yearPvEnergy": "15438.09",
    "totalPvEnergy": "30591.25",
    "todaySellEnergy": "42.82",
    "monthSellEnergy": "805.78",
    "yearSellEnergy": "15438.09",
    "totalSellEnergy": "30591.25",
    "totalPlantTreeNum": "54.24",
    "totalReduceCo2": "30.5",
    "co2UnitOfWeight": "t",
    "deviceSnList": ["R6I3333J2346C31446"],
    "todayEquivalentHours": "1.53"
  }
}
```

---

### 7.11 Query Load Monitoring Data

Retrieve load monitoring data at minute, day, month, or year granularity. Supports self-consumption rate and grid exchange analysis.

| Property | Value |
|----------|-------|
| **Method** | `GET` |
| **Endpoint** | `/open/api/device/secData` |
| **Full URL** | `https://intl-developer.saj-electric.com/prod-api/open/api/device/secData` |

#### Request Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `plantId` | String | YES | Plant ID |
| `startTime` | String | YES | Start time `yyyy-MM-dd HH:mm:ss` |
| `endTime` | String | YES | End time `yyyy-MM-dd HH:mm:ss` |
| `timeUnit` | Integer | YES | Granularity: 0=minute, 1=day, 2=month, 3=year |

#### Request Example

```
GET /prod-api/open/api/device/secData?plantId=10407466169&startTime=2022-05-05 00:00:00&endTime=2022-05-17 00:00:00&timeUnit=1

Headers:
  accessToken: <your_token>
  content-language: en_US
```

#### Response Structure

| Name | Type | Description |
|------|------|-------------|
| `data.dataList` | array | List of module data |
| `data.dataList[].moduleSn` | string | Load monitoring serial number |
| `data.dataList[].timeUnit` | integer | Time unit |
| `data.dataList[].total` | object | Aggregated totals |
| `data.dataList[].data` | array | Time-series data points |

**Aggregated Totals (total):**

| Name | Type | Description |
|------|------|-------------|
| `pvEnergy` | number | Total PV energy (kWh) |
| `loadEnergy` | number | Total load energy (kWh) |
| `buyEnergy` | number | Total grid purchase energy (kWh) |
| `sellEnergy` | number | Total grid sell energy (kWh) |
| `chargeEnergy` | number | Total battery charge energy (kWh) |
| `dischargeEnergy` | number | Total battery discharge energy (kWh) |
| `energyUnit` | string | Energy unit (kWh) |
| `pvSelfConsumedEnergy` | number | PV self-consumed energy (kWh) |
| `pvSelfConsumedRate` | number | PV self-consumption rate (0–1) |
| `pvSellRate` | number | PV sell rate (0–1) |
| `loadSelfConsumedEnergy` | number | Load self-consumed energy (kWh) |
| `loadSelfConsumedRate` | number | Load self-consumption rate (0–1) |
| `loadBuyRate` | number | Load buy rate (0–1) |

**timeUnit=0 (Minute) Data Points:**

| Name | Type | Description |
|------|------|-------------|
| `dataTime` | string | Timestamp |
| `pvPower` | number | PV power (W) |
| `loadPower` | number | Load power (W) |
| `selfUsePower` | number | Self-use power (W) |
| `buyPower` | number | Grid buy power (W) |
| `sellPower` | number | Grid sell power (W) |

**timeUnit=1/2/3 (Day/Month/Year) Data Points:**

| Name | Type | Description |
|------|------|-------------|
| `dataTime` | string | Timestamp |
| `pvEnergy` | number | PV energy (kWh) |
| `loadEnergy` | number | Load energy (kWh) |
| `sellEnergy` | number | Grid sell energy (kWh) |
| `buyEnergy` | number | Grid buy energy (kWh) |
| `selfUseEnergy` | number | Self-use energy (kWh) |

#### Response Example (timeUnit=0, Minute Data)

```json
{
  "code": 200,
  "msg": "request success",
  "data": {
    "dataList": [
      {
        "moduleSn": "M5370G2212015150",
        "timeUnit": 0,
        "total": {
          "buyEnergy": 2.49,
          "chargeEnergy": 0,
          "dischargeEnergy": 0,
          "energyUnit": "kWh",
          "loadBuyRate": 0.049,
          "loadEnergy": 50.5,
          "loadSelfConsumedEnergy": 48.02,
          "loadSelfConsumedRate": 0.951,
          "pvEnergy": 48.05,
          "pvSelfConsumedEnergy": 48.02,
          "pvSelfConsumedRate": 0.999,
          "pvSellRate": 0.001,
          "sellEnergy": 0.04
        },
        "data": [
          {
            "dataTime": "2022-05-05 00:00:00",
            "pvPower": 0,
            "loadPower": 0,
            "selfUsePower": 0,
            "buyPower": 0,
            "sellPower": 0
          }
        ]
      }
    ]
  }
}
```

---

## 8. Alarm & Fault API

### 8.1 Get Fault Events of Device (New)

Retrieve fault/alarm events for a device with filtering options.

| Property | Value |
|----------|-------|
| **Method** | `GET` |
| **Endpoint** | `/open/api/device/alarmList` |
| **Full URL** | `https://intl-developer.saj-electric.com/prod-api/open/api/device/alarmList` |

#### Request Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `deviceSn` | String | NO | Inverter serial number |
| `alarmLevel` | String | NO | Alarm level (comma-separated): 1=Normal, 2=Important, 3=Urgent |
| `startTime` | String | NO | Start time `yyyy-MM-dd HH:mm:ss` (defaults to status=1 if omitted) |
| `endTime` | String | NO | End time `yyyy-MM-dd HH:mm:ss` (defaults to status=1 if omitted) |
| `status` | String | NO | Alarm status (comma-separated): 1=pending, 4=closed |
| `appId` | String | NO | Required if deviceSn is not provided |

#### Response

| Name | Type | Description |
|------|------|-------------|
| `code` | integer | Response code |
| `msg` | string | Response message |
| `data` | array | Array of alarm objects |
| `data[].deviceSn` | string | Inverter serial number |
| `data[].alarmCode` | integer | Alarm code |
| `data[].alarmName` | string | Alarm name |
| `data[].alarmLevel` | integer | Alarm level: 1=Normal, 2=Important, 3=Urgent |
| `data[].status` | integer | Alarm status: 1=pending, 4=closed |
| `data[].alarmTime` | string | Alarm time |

#### Response Example

```json
{
  "code": 0,
  "msg": "",
  "data": [
    {
      "deviceSn": "HSS2602Y2223123456",
      "alarmCode": 1001,
      "alarmName": "Grid Over-Voltage",
      "alarmLevel": 1,
      "status": 1,
      "alarmTime": "2022-07-04 10:30:00"
    }
  ]
}
```

---

### 8.2 Get Alarms of Device

Retrieve device alarms with detailed type filtering and sorting options. Supports filtering by device type, sub-type, and alarm state.

| Property | Value |
|----------|-------|
| **Method** | `GET` |
| **Endpoint** | `/open/api/device/getDeviceAlarmList` |
| **Full URL** | `https://intl-developer.saj-electric.com/prod-api/open/api/device/getDeviceAlarmList` |

#### Request Parameters

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `deviceSn` | String | YES | — | Inverter SN |
| `alarmCommonState` | Integer | YES | — | Alarm state: 1=Pending, 2=In Progress, 3=Closed |
| `alarmDeviceType` | Integer | YES | — | Device type: 1=Inverter, 2=Charging Pile, 3=Battery, 4=Air Conditioning, 5=Fire Protection, 6=Diesel Generator |
| `alarmDeviceSubType` | Integer | NO | — | Sub-type (when alarmDeviceType=3): 1=Battery, 2=Battery Group, 3=Battery Cluster, 4=Battery Pack, 5=Battery Cell |
| `orderByIndex` | Integer | NO | 1 | Sort: 1=desc start time, 2=asc start time, 3=desc update time, 4=asc update time |
| `userId` | String | NO | — | User ID |

#### Response

| Name | Type | Description |
|------|------|-------------|
| `data` | array | Array of alarm records |
| `data[].serialNumber` | integer | Serial number |
| `data[].alarmID` | string | Alarm record ID |
| `data[].alarmName` | string | Alarm name |
| `data[].deviceSn` | string | Associated device SN |
| `data[].optSn` | string | Device SN |
| `data[].deviceType` | integer | Device type: 1=Inverter, 2=Charging Pile, 3=Battery, 4=AC, 5=Fire, 6=Diesel, 7=Module, 8=Liquid Cooling |
| `data[].deviceTypeName` | string | Device type name |
| `data[].plantUid` | string | Plant UID |
| `data[].plantName` | string | Plant name |
| `data[].plantCountry` | string | Plant country |
| `data[].alarmLevel` | integer | Level: 1=General, 2=Important, 3=Urgent |
| `data[].alarmLevelName` | string | Alarm level name |
| `data[].alarmState` | integer | Status: 1=Happening, 4=Recovered |
| `data[].alarmStateName` | string | Alarm status name |
| `data[].alarmStartTime` | string | Alarm start time |
| `data[].alarmUpdateTime` | string | Alarm update/recovery time |
| `data[].alarmDuration` | string | Alarm duration |

#### Response Example

```json
{
  "code": 0,
  "msg": "",
  "data": [
    {
      "serialNumber": 1,
      "alarmID": "ALM123456789",
      "alarmName": "Inverter Over-Temperature Alarm",
      "deviceSn": "SN123456789",
      "optSn": "OPTSN987654321",
      "deviceType": 1,
      "deviceTypeName": "Inverter",
      "plantUid": "PLANTUID001",
      "plantName": "Some Power Plant Name",
      "plantCountry": "China",
      "alarmLevel": 3,
      "alarmLevelName": "Urgent",
      "alarmState": 1,
      "alarmStateName": "Happening",
      "alarmStartTime": "2023-10-01T10:00:00Z",
      "alarmUpdateTime": "2023-10-01T10:30:00Z",
      "alarmDuration": "30 minutes"
    }
  ]
}
```

#### Endpoint-Specific Response Codes

| Code | Description |
|------|-------------|
| `200` | Request successful |
| `10001` | Server exception |
| `10006` | API parameter error |

---

## Appendix A: Quick Reference — All Endpoints

| # | Category | Endpoint | Method | Rate Limit | Description |
|---|----------|----------|--------|------------|-------------|
| 1 | Auth | `/open/api/access_token` | GET | — | Get developer access token |
| 2 | Auth | `/open/api/developer/auth/plantAuthByOpenapiKey` | POST | — | Auth plant to developer |
| 3 | System | `/open/api/developer/device/page` | GET | 5 QPS | Get authorized devices (paginated) |
| 4 | System | `/open/api/developer/plant/page` | GET | 5 QPS | Get authorized plants (paginated) |
| 5 | Info | `/open/api/plant/getPlantAllDeviceList` | GET | 5 QPS | Get all devices from plant |
| 6 | Info | `/open/api/device/baseinfo` | GET | 5 QPS | Get device basic information |
| 7 | Info | `/open/api/device/batInfo` | GET | 5 QPS | Get device & battery details |
| 8 | Data | `/open/api/device/realtimeDataCommon` | GET | **150 QPS** | Realtime data (common) |
| 9 | Data | `/open/api/device/emsRealtimeDataCommon` | GET | **10 QPS** | EMS real-time data |
| 10 | Data | `/open/api/device/uploadData` | GET | **50 QPS** | Device upload data (historical) |
| 11 | Data | `/open/api/device/historyDataCommon` | GET | 5 QPS | History data (common) |
| 12 | Data | `/open/api/device/emsHistoryData` | GET | 5 QPS | EMS history data |
| 13 | Data | `/open/api/device/emsHistoryData4Meter` | GET | 5 QPS | EMS meter historical data |
| 14 | Data | `/open/api/plant/details` | GET | 5 QPS | Plant details |
| 15 | Data | `/open/api/plant/energy` | GET | 5 QPS | Plant energy summary |
| 16 | Data | `/open/api/plant/getPlantStatisticsData` | GET | 5 QPS | Plant statistics data |
| 17 | Data | `/open/api/device/secData` | GET | 5 QPS | Load monitoring data |
| 18 | Alarm | `/open/api/device/alarmList` | GET | 5 QPS | Fault events (new) |
| 19 | Alarm | `/open/api/device/getDeviceAlarmList` | GET | 5 QPS | Device alarms (detailed) |

---

## Appendix B: Authentication Flow

```
┌─────────────┐     1. GET /access_token         ┌──────────────────┐
│  Developer   │  ──────(appId + appSecret)──────▶ │  SAJ API Server  │
│  Application │  ◀──────(access_token)──────────  │                  │
│              │                                   │                  │
│              │     2. API calls with headers:     │                  │
│              │  ──────accessToken: <token>──────▶ │                  │
│              │        content-language: en_US     │                  │
│              │  ◀──────JSON Response────────────  │                  │
└─────────────┘                                   └──────────────────┘
```

**Token Lifecycle:**
- Default expiration: **28,800 seconds (8 hours)**
- Refresh by calling the access token endpoint again before expiry

---

## Appendix C: Work Mode (mpvMode) Reference

| Value | Mode | Description |
|-------|------|-------------|
| 0 | Initialization | System starting up |
| 1 | Waiting | Waiting for conditions |
| 2 | Grid-Connected | Normal grid-connected operation |
| 3 | Off-Grid | Off-grid mode (energy storage) |
| 4 | Grid Load | Grid load mode (energy storage) |
| 5 | Fault | System fault |
| 6 | Upgrade | Firmware upgrade in progress |
| 7 | Debugging | Debug mode |
| 8 | Self-Inspection | Self-inspection mode |
| 9 | Reset | System reset |

---

## Appendix D: Direction & Status Enums

**Battery Direction (`batteryDirection`):**

| Value | Meaning |
|-------|---------|
| -1 | Charging |
| 0 | Standby |
| 1 | Discharging |

**Grid Direction (`gridDirection`):**

| Value | Meaning |
|-------|---------|
| -1 | Buying from grid (feed-in) |
| 0 | No flow |
| 1 | Selling to grid (export) |

**Device Running State (`runningState`):**

| Value | Meaning |
|-------|---------|
| 1 | Normal |
| 2 | Alarm |
| 3 | Offline |
| 4 | Historical |
| 5 | Not monitored |

**Inverter Device Status (`deviceStatus`):**

| Category | Value | Status |
|----------|-------|--------|
| Grid-tied | 1 | Standby |
| Grid-tied | 2 | Normal |
| Grid-tied | 3 | Error |
| Grid-tied | 4 | Warning |
| Storage | 1 | Standby |
| Storage | 2 | On-grid |
| Storage | 3 | Off-grid |
| Storage | 4 | Bypass |
| Storage | 5 | Fault |
| Storage | 6 | Upgrade |
| AC-coupled | 1 | Standby |
| AC-coupled | 2 | On-grid |
| AC-coupled | 3 | Off-grid |
| AC-coupled | 4 | Bypass |
| AC-coupled | 5 | Faults |
| AC-coupled | 6 | Upgrade |

---

*Document generated from SAJ Elekeeper Open Platform API source material. For the latest updates, refer to [https://intl-developer.saj-electric.com](https://intl-developer.saj-electric.com).*
