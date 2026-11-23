package model

// Strict SI unit types used across all kilnscrub packages to prevent unit mix-ups.
type (
	// Kelvin represents absolute temperature in K.
	Kelvin float64
	// Pascal represents pressure or pressure drop in Pa (N/m^2).
	Pascal float64
	// CubicMetersPerSec represents volumetric flow rate in m^3/s.
	CubicMetersPerSec float64
	// Meters represents linear length or diameter in m.
	Meters float64
	// MetersPerSec represents linear or superficial velocity in m/s.
	MetersPerSec float64
	// KgPerSec represents mass flow rate in kg/s.
	KgPerSec float64
	// KgPerCubicMeter represents density in kg/m^3.
	KgPerCubicMeter float64
	// PaSec represents dynamic viscosity in Pa*s (kg/(m*s)).
	PaSec float64
	// Watts represents power or heat rate in W (J/s).
	Watts float64
	// SquareMeters represents area in m^2.
	SquareMeters float64
	// CubicMeters represents volume in m^3.
	CubicMeters float64
)

// Exact and NIST-conventional unit conversion constants and physical constants.
const (
	// PaPerInWC is the NIST conventional inch of water column (at 4 C): 0.0254 m * 9806.65 Pa/m.
	PaPerInWC = 249.08891
	// MetersPerInch is the exact international inch in meters.
	MetersPerInch = 0.0254
	// MetersPerFoot is the exact international foot in meters.
	MetersPerFoot = 0.3048
	// CubicMetersPerCuFt is 0.3048^3 m^3/ft^3.
	CubicMetersPerCuFt = 0.028316846592
	// SquareMetersPerSqFt is 0.3048^2 m^2/ft^2.
	SquareMetersPerSqFt = 0.09290304
	// SquareMetersPerSqIn is 0.0254^2 m^2/in^2.
	SquareMetersPerSqIn = 0.00064516
	// CubicMetersPerSecPerCFM converts cubic feet per minute to m^3/s.
	CubicMetersPerSecPerCFM = CubicMetersPerCuFt / 60.0
	// CubicMetersPerUSGallon is 231 in^3 in m^3.
	CubicMetersPerUSGallon = 0.003785411784
	// CubicMetersPerSecPerGPM converts US gallons per minute to m^3/s.
	CubicMetersPerSecPerGPM = CubicMetersPerUSGallon / 60.0
	// KgPerLb is the exact international avoirdupois pound in kg.
	KgPerLb = 0.45359237
	// KgPerSecPerLbPerHr converts lbs/hr to kg/s.
	KgPerSecPerLbPerHr = KgPerLb / 3600.0
	// WattsPerHP is mechanical horsepower (550 ft*lbf/s) in Watts.
	WattsPerHP = 745.6998715822702
	// MetersPerMicron converts micrometers to meters.
	MetersPerMicron = 1e-6
	// MetersPerMillimeter converts millimeters to meters.
	MetersPerMillimeter = 1e-3

	// StandardPressurePa is 1 standard atmosphere in Pa.
	StandardPressurePa Pascal = 101325.0
	// StandardGravity is standard gravitational acceleration in m/s^2.
	StandardGravity = 9.80665
	// UniversalGasConst is the molar gas constant R in J/(mol*K).
	UniversalGasConst = 8.314462618
	// WaterDensity is liquid water density at ~20 C in kg/m^3.
	WaterDensity KgPerCubicMeter = 998.2
	// NaClParticleDensity is solid/condensed NaCl fume density in kg/m^3.
	NaClParticleDensity KgPerCubicMeter = 2160.0
	// WoodAshParticleDensity is typical wood fly ash particulate density in kg/m^3.
	WoodAshParticleDensity KgPerCubicMeter = 2400.0
)

// Temperature conversions.

// FahrenheitToKelvin converts degrees Fahrenheit to Kelvin.
func FahrenheitToKelvin(f float64) Kelvin {
	return Kelvin((f-32.0)*5.0/9.0 + 273.15)
}

// KelvinToFahrenheit converts Kelvin to degrees Fahrenheit.
func KelvinToFahrenheit(k Kelvin) float64 {
	return (float64(k)-273.15)*9.0/5.0 + 32.0
}

// CelsiusToKelvin converts degrees Celsius to Kelvin.
func CelsiusToKelvin(c float64) Kelvin {
	return Kelvin(c + 273.15)
}

// KelvinToCelsius converts Kelvin to degrees Celsius.
func KelvinToCelsius(k Kelvin) float64 {
	return float64(k) - 273.15
}

// Fahrenheit returns temperature in degrees Fahrenheit.
func (k Kelvin) Fahrenheit() float64 { return KelvinToFahrenheit(k) }

// Celsius returns temperature in degrees Celsius.
func (k Kelvin) Celsius() float64 { return KelvinToCelsius(k) }

// Pressure conversions.

// InWCToPascal converts inches of water column (inWC) to Pascals.
func InWCToPascal(inwc float64) Pascal {
	return Pascal(inwc * PaPerInWC)
}

// PascalToInWC converts Pascals to inches of water column (inWC).
func PascalToInWC(p Pascal) float64 {
	return float64(p) / PaPerInWC
}

// InWC returns pressure in inches of water column.
func (p Pascal) InWC() float64 { return PascalToInWC(p) }

// Volumetric flow conversions.

// CFMToCubicMetersPerSec converts cubic feet per minute (CFM) to m^3/s.
func CFMToCubicMetersPerSec(cfm float64) CubicMetersPerSec {
	return CubicMetersPerSec(cfm * CubicMetersPerSecPerCFM)
}

// CubicMetersPerSecToCFM converts m^3/s to cubic feet per minute (CFM).
func CubicMetersPerSecToCFM(q CubicMetersPerSec) float64 {
	return float64(q) / CubicMetersPerSecPerCFM
}

// GPMToCubicMetersPerSec converts US gallons per minute (GPM) to m^3/s.
func GPMToCubicMetersPerSec(gpm float64) CubicMetersPerSec {
	return CubicMetersPerSec(gpm * CubicMetersPerSecPerGPM)
}

// CubicMetersPerSecToGPM converts m^3/s to US gallons per minute (GPM).
func CubicMetersPerSecToGPM(q CubicMetersPerSec) float64 {
	return float64(q) / CubicMetersPerSecPerGPM
}

// CFM returns volumetric flow in cubic feet per minute.
func (q CubicMetersPerSec) CFM() float64 { return CubicMetersPerSecToCFM(q) }

// GPM returns volumetric flow in US gallons per minute.
func (q CubicMetersPerSec) GPM() float64 { return CubicMetersPerSecToGPM(q) }

// Mass flow conversions.

// LbsPerHrToKgPerSec converts pounds per hour (lbs/hr) to kg/s.
func LbsPerHrToKgPerSec(lbsPerHr float64) KgPerSec {
	return KgPerSec(lbsPerHr * KgPerSecPerLbPerHr)
}

// KgPerSecToLbsPerHr converts kg/s to pounds per hour (lbs/hr).
func KgPerSecToLbsPerHr(m KgPerSec) float64 {
	return float64(m) / KgPerSecPerLbPerHr
}

// LbsPerHr returns mass flow in pounds per hour.
func (m KgPerSec) LbsPerHr() float64 { return KgPerSecToLbsPerHr(m) }

// Length conversions.

// InchesToMeters converts inches to meters.
func InchesToMeters(in float64) Meters {
	return Meters(in * MetersPerInch)
}

// MetersToInches converts meters to inches.
func MetersToInches(m Meters) float64 {
	return float64(m) / MetersPerInch
}

// FeetToMeters converts feet to meters.
func FeetToMeters(ft float64) Meters {
	return Meters(ft * MetersPerFoot)
}

// MetersToFeet converts meters to feet.
func MetersToFeet(m Meters) float64 {
	return float64(m) / MetersPerFoot
}

// MicronsToMeters converts micrometers (um) to meters.
func MicronsToMeters(um float64) Meters {
	return Meters(um * MetersPerMicron)
}

// MetersToMicrons converts meters to micrometers (um).
func MetersToMicrons(m Meters) float64 {
	return float64(m) / MetersPerMicron
}

// MillimetersToMeters converts millimeters to meters.
func MillimetersToMeters(mm float64) Meters {
	return Meters(mm * MetersPerMillimeter)
}

// MetersToMillimeters converts meters to millimeters.
func MetersToMillimeters(m Meters) float64 {
	return float64(m) / MetersPerMillimeter
}

// Inches returns length in inches.
func (m Meters) Inches() float64 { return MetersToInches(m) }

// Feet returns length in feet.
func (m Meters) Feet() float64 { return MetersToFeet(m) }

// Microns returns length in micrometers.
func (m Meters) Microns() float64 { return MetersToMicrons(m) }

// Millimeters returns length in millimeters.
func (m Meters) Millimeters() float64 { return MetersToMillimeters(m) }

// Velocity conversions.

// FeetPerSecToMetersPerSec converts ft/s to m/s.
func FeetPerSecToMetersPerSec(fps float64) MetersPerSec {
	return MetersPerSec(fps * MetersPerFoot)
}

// MetersPerSecToFeetPerSec converts m/s to ft/s.
func MetersPerSecToFeetPerSec(v MetersPerSec) float64 {
	return float64(v) / MetersPerFoot
}

// FeetPerSec returns velocity in feet per second.
func (v MetersPerSec) FeetPerSec() float64 { return MetersPerSecToFeetPerSec(v) }

// Area and Volume conversions.

// CuFtToCubicMeters converts cubic feet (cu.ft.) to m^3.
func CuFtToCubicMeters(cuft float64) CubicMeters {
	return CubicMeters(cuft * CubicMetersPerCuFt)
}

// CubicMetersToCuFt converts m^3 to cubic feet (cu.ft.).
func CubicMetersToCuFt(v CubicMeters) float64 {
	return float64(v) / CubicMetersPerCuFt
}

// CuFt returns volume in cubic feet.
func (v CubicMeters) CuFt() float64 { return CubicMetersToCuFt(v) }

// SqFtToSquareMeters converts square feet to m^2.
func SqFtToSquareMeters(sqft float64) SquareMeters {
	return SquareMeters(sqft * SquareMetersPerSqFt)
}

// SquareMetersToSqFt converts m^2 to square feet.
func SquareMetersToSqFt(a SquareMeters) float64 {
	return float64(a) / SquareMetersPerSqFt
}

// SqFt returns area in square feet.
func (a SquareMeters) SqFt() float64 { return SquareMetersToSqFt(a) }

// Power conversions.

// HPToWatts converts mechanical horsepower (HP) to Watts.
func HPToWatts(hp float64) Watts {
	return Watts(hp * WattsPerHP)
}

// WattsToHP converts Watts to mechanical horsepower (HP).
func WattsToHP(w Watts) float64 {
	return float64(w) / WattsPerHP
}

// HP returns power in mechanical horsepower.
func (w Watts) HP() float64 { return WattsToHP(w) }
