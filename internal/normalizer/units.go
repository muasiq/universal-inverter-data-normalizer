package normalizer

// UnitConversion provides utilities for converting between different
// unit systems used by various inverter brands.

// WattsToKilowatts converts power from W to kW.
func WattsToKilowatts(w float64) float64 {
	return w / 1000.0
}

// KilowattsToWatts converts power from kW to W.
func KilowattsToWatts(kw float64) float64 {
	return kw * 1000.0
}

// WhToKWh converts energy from Wh to kWh.
func WhToKWh(wh float64) float64 {
	return wh / 1000.0
}

// KWhToWh converts energy from kWh to Wh.
func KWhToWh(kwh float64) float64 {
	return kwh * 1000.0
}

// MWhToKWh converts energy from MWh to kWh.
func MWhToKWh(mwh float64) float64 {
	return mwh * 1000.0
}

// WpToKWp converts peak power from Wp to kWp.
func WpToKWp(wp float64) float64 {
	return wp / 1000.0
}

// KWpToWp converts peak power from kWp to Wp.
func KWpToWp(kwp float64) float64 {
	return kwp * 1000.0
}

// CelsiusToFahrenheit converts temperature.
func CelsiusToFahrenheit(c float64) float64 {
	return c*9.0/5.0 + 32.0
}

// FahrenheitToCelsius converts temperature.
func FahrenheitToCelsius(f float64) float64 {
	return (f - 32.0) * 5.0 / 9.0
}

// CO2KgPerKWh is the default CO2 emissions factor.
// Average global grid emissions factor is ~0.475 kg CO2/kWh.
const CO2KgPerKWh = 0.475

// CalculateCO2Savings returns CO2 saved in kg for given kWh of solar generation.
func CalculateCO2Savings(kWh float64) float64 {
	return kWh * CO2KgPerKWh
}

// CalculateTreesEquivalent estimates the number of trees equivalent
// for the given CO2 savings in kg.
// Average tree absorbs ~22 kg CO2/year.
func CalculateTreesEquivalent(co2SavedKg float64) float64 {
	return co2SavedKg / 22.0
}

// CalculateSelfConsumptionRate returns the self-consumption rate (0.0 - 1.0).
// selfConsumedKWh = pvGenerationKWh - gridExportKWh
func CalculateSelfConsumptionRate(pvGenerationKWh, gridExportKWh float64) float64 {
	if pvGenerationKWh <= 0 {
		return 0
	}
	rate := (pvGenerationKWh - gridExportKWh) / pvGenerationKWh
	if rate < 0 {
		return 0
	}
	if rate > 1 {
		return 1
	}
	return rate
}

// CalculateSelfSufficiencyRate returns the self-sufficiency rate (0.0 - 1.0).
// selfSufficiency = (totalConsumption - gridImport) / totalConsumption
func CalculateSelfSufficiencyRate(totalConsumptionKWh, gridImportKWh float64) float64 {
	if totalConsumptionKWh <= 0 {
		return 0
	}
	rate := (totalConsumptionKWh - gridImportKWh) / totalConsumptionKWh
	if rate < 0 {
		return 0
	}
	if rate > 1 {
		return 1
	}
	return rate
}

// SafeFloat returns the value of a float64 pointer, or 0 if nil.
func SafeFloat(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

// FloatPtr returns a pointer to a float64 value.
func FloatPtr(v float64) *float64 {
	return &v
}

// IntPtr returns a pointer to an int value.
func IntPtr(v int) *int {
	return &v
}
