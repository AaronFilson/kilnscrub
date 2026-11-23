package model

import (
	"context"
	"fmt"
	"math"
	"time"
)

// Allowed FuelType identifiers for DesignInput.FuelType.
const (
	FuelWoodSalt = "wood_salt"
	FuelWood     = "wood"
	FuelGasSalt  = "gas_salt"
	FuelGas      = "gas"
)

func isFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

// DesignInput specifies all kiln firing, aerodynamic, and scrubbing parameters
// for sizing a 4-stage studio kiln exhaust scrubber.
type DesignInput struct {
	KilnVolumeCuFt float64 `json:"kiln_volume_cu_ft"`
	PeakExitTempF  float64 `json:"peak_exit_temp_f"`
	SaltRateLbsPerHr float64 `json:"salt_rate_lbs_per_hr"`
	ExcessAirFraction float64 `json:"excess_air_fraction"`
	// FuelType selects combustion stoichiometry: "wood_salt" (default when empty),
	// "wood", "gas_salt", or "gas".
	FuelType string `json:"fuel_type"`
	// VenturiThroatVelocity is the target throat gas velocity (typical studio range 50-90 m/s).
	VenturiThroatVelocity MetersPerSec `json:"venturi_throat_velocity_m_s"`
	VenturiLGRatio        float64      `json:"venturi_lg_ratio"`
	// VenturiDropletDiameter is the liquid spray Sauter mean droplet diameter in meters.
	// When 0, the venturi solver auto-calculates droplet diameter (e.g. Nukiyama-Tanasawa or 50 um default).
	VenturiDropletDiameter Meters `json:"venturi_droplet_diameter_m"`
	// TargetHClRemoval is the fractional HCl removal efficiency in (0, 1) (target >= 0.98).
	TargetHClRemoval float64 `json:"target_hcl_removal"`
	PackingDiameter  Meters  `json:"packing_diameter_m"`
	PackingVoidage   float64 `json:"packing_voidage"`
	PackingName      string  `json:"packing_name"`
	// PackingFactor is the packing factor Fp in 1/m (e.g. 170 1/m for 1" PP Pall rings); must be > 0.
	PackingFactor float64 `json:"packing_factor_1_m"`
	// TargetKilnDraftInWC is the signed kiln draft setpoint in inWC (negative = suction relative
	// to ambient, e.g. -0.05 inWC; valid range [-2.0, 0]).
	TargetKilnDraftInWC float64 `json:"target_kiln_draft_in_wc"`
	AmbientTempK        Kelvin  `json:"ambient_temp_k"`
	AmbientPressurePa   Pascal  `json:"ambient_pressure_pa"`
	SprayWaterTempK     Kelvin  `json:"spray_water_temp_k"`
	SumpPHTarget        float64 `json:"sump_ph_target"`
	DuctLength          Meters  `json:"duct_length_m"`
	// DuctDiameter is the flue duct diameter in meters; when 0, it is auto-sized from flue velocity.
	DuctDiameter Meters `json:"duct_diameter_m"`
}

// DefaultSaltKilnInput returns the reference 24 cu.ft. wood/salt kiln
// operating at 2400 F peak exit temperature and 2 lbs NaCl/hr salt introduction.
func DefaultSaltKilnInput() DesignInput {
	return DesignInput{
		KilnVolumeCuFt:         24.0,
		PeakExitTempF:          2400.0,
		SaltRateLbsPerHr:       2.0,
		ExcessAirFraction:      0.30,
		FuelType:               FuelWoodSalt,
		VenturiThroatVelocity:  70.0,
		VenturiLGRatio:         0.0012,
		VenturiDropletDiameter: MicronsToMeters(50.0),
		TargetHClRemoval:       0.99,
		PackingDiameter:        InchesToMeters(1.0),
		PackingVoidage:         0.90,
		PackingName:            `1" PP Pall Rings`,
		PackingFactor:          170.0,
		TargetKilnDraftInWC:    -0.05,
		AmbientTempK:           CelsiusToKelvin(20.0),
		AmbientPressurePa:      StandardPressurePa,
		SprayWaterTempK:        CelsiusToKelvin(20.0),
		SumpPHTarget:           8.0,
		DuctLength:             3.0,
		DuctDiameter:           0.0, // Auto-sized when zero.
	}
}

// Validate checks that DesignInput parameters are finite and physically meaningful.
func (d DesignInput) Validate() error {
	if !isFinite(d.KilnVolumeCuFt) || d.KilnVolumeCuFt <= 0 {
		return fmt.Errorf("kiln_volume_cu_ft must be finite and > 0, got %v", d.KilnVolumeCuFt)
	}
	if !isFinite(float64(d.AmbientTempK)) || d.AmbientTempK <= 200.0 || d.AmbientTempK >= 400.0 {
		return fmt.Errorf("ambient_temp_k must be finite and in (200, 400), got %v", d.AmbientTempK)
	}
	if !isFinite(d.PeakExitTempF) {
		return fmt.Errorf("peak_exit_temp_f must be finite, got %v", d.PeakExitTempF)
	}
	peakK := FahrenheitToKelvin(d.PeakExitTempF)
	if peakK <= d.AmbientTempK || peakK > 2500.0 {
		return fmt.Errorf("peak_exit_temp_f (%v F / %.2f K) must be above ambient (%.2f K) and <= 2500 K", d.PeakExitTempF, peakK, d.AmbientTempK)
	}
	if !isFinite(d.SaltRateLbsPerHr) || d.SaltRateLbsPerHr < 0 {
		return fmt.Errorf("salt_rate_lbs_per_hr must be finite and >= 0, got %v", d.SaltRateLbsPerHr)
	}
	if !isFinite(d.ExcessAirFraction) || d.ExcessAirFraction < 0 || d.ExcessAirFraction > 5.0 {
		return fmt.Errorf("excess_air_fraction must be finite and in [0, 5.0], got %v", d.ExcessAirFraction)
	}
	switch d.FuelType {
	case "", FuelWoodSalt, FuelWood, FuelGasSalt, FuelGas:
	default:
		return fmt.Errorf("fuel_type must be one of %q, %q, %q, %q (or empty), got %q",
			FuelWoodSalt, FuelWood, FuelGasSalt, FuelGas, d.FuelType)
	}
	if !isFinite(float64(d.VenturiThroatVelocity)) || d.VenturiThroatVelocity <= 0 || d.VenturiThroatVelocity > 200.0 {
		return fmt.Errorf("venturi_throat_velocity_m_s must be finite and in (0, 200], got %v", d.VenturiThroatVelocity)
	}
	if !isFinite(d.VenturiLGRatio) || d.VenturiLGRatio <= 0 || d.VenturiLGRatio > 0.05 {
		return fmt.Errorf("venturi_lg_ratio must be finite and in (0, 0.05], got %v", d.VenturiLGRatio)
	}
	if !isFinite(float64(d.VenturiDropletDiameter)) || d.VenturiDropletDiameter < 0 {
		return fmt.Errorf("venturi_droplet_diameter_m must be finite and >= 0, got %v", d.VenturiDropletDiameter)
	}
	if !isFinite(d.TargetHClRemoval) || d.TargetHClRemoval <= 0 || d.TargetHClRemoval >= 1.0 {
		return fmt.Errorf("target_hcl_removal must be finite and in (0, 1), got %v", d.TargetHClRemoval)
	}
	if !isFinite(float64(d.PackingDiameter)) || d.PackingDiameter <= 0 {
		return fmt.Errorf("packing_diameter_m must be finite and > 0, got %v", d.PackingDiameter)
	}
	if !isFinite(d.PackingVoidage) || d.PackingVoidage <= 0 || d.PackingVoidage >= 1.0 {
		return fmt.Errorf("packing_voidage must be finite and in (0, 1), got %v", d.PackingVoidage)
	}
	if !isFinite(d.PackingFactor) || d.PackingFactor <= 0 {
		return fmt.Errorf("packing_factor_1_m must be finite and > 0, got %v", d.PackingFactor)
	}
	if !isFinite(d.TargetKilnDraftInWC) || d.TargetKilnDraftInWC > 0 || d.TargetKilnDraftInWC < -2.0 {
		return fmt.Errorf("target_kiln_draft_in_wc must be finite and in [-2.0, 0] (negative = suction draft), got %v", d.TargetKilnDraftInWC)
	}
	if !isFinite(float64(d.AmbientPressurePa)) || d.AmbientPressurePa <= 50000.0 || d.AmbientPressurePa >= 150000.0 {
		return fmt.Errorf("ambient_pressure_pa must be finite and in (50000, 150000), got %v", d.AmbientPressurePa)
	}
	if !isFinite(float64(d.SprayWaterTempK)) || d.SprayWaterTempK <= 273.15 || d.SprayWaterTempK >= 373.15 {
		return fmt.Errorf("spray_water_temp_k must be finite and in (273.15, 373.15), got %v", d.SprayWaterTempK)
	}
	if !isFinite(d.SumpPHTarget) || d.SumpPHTarget < 6.0 || d.SumpPHTarget > 11.0 {
		return fmt.Errorf("sump_ph_target must be finite and in [6.0, 11.0], got %v", d.SumpPHTarget)
	}
	if !isFinite(float64(d.DuctLength)) || d.DuctLength < 0 || !isFinite(float64(d.DuctDiameter)) || d.DuctDiameter < 0 {
		return fmt.Errorf("duct dimensions must be finite and non-negative")
	}
	return nil
}

// GasComposition holds flue gas species mole fractions and thermophysical state.
type GasComposition struct {
	N2                 float64           `json:"n2"`
	CO2                float64           `json:"co2"`
	H2O                float64           `json:"h2o"`
	O2                 float64           `json:"o2"`
	HCl                float64           `json:"hcl"`
	MolarFlowMolPerSec float64           `json:"molar_flow_mol_s"`
	MassFlow           KgPerSec          `json:"mass_flow_kg_s"`
	VolumetricFlow     CubicMetersPerSec `json:"volumetric_flow_m3_s"`
	VolumetricFlowACFM float64           `json:"volumetric_flow_acfm"`
	MolecularWeight    float64           `json:"molecular_weight_kg_mol"`
	Temperature        Kelvin            `json:"temperature_k"`
	Pressure           Pascal            `json:"pressure_pa"`
	Density            KgPerCubicMeter   `json:"density_kg_m3"`
	Viscosity          PaSec             `json:"viscosity_pa_s"`
	CpJPerMolK         float64           `json:"cp_j_mol_k"`
	EnthalpyJPerMol    float64           `json:"enthalpy_j_mol"`
}

// SumMoleFractions returns the sum of all tracked species mole fractions.
func (g GasComposition) SumMoleFractions() float64 {
	return g.N2 + g.CO2 + g.H2O + g.O2 + g.HCl
}

// QuenchResult captures Stage 1 (316L SS Adiabatic Spray Quench Tower) sizing and thermal performance.
type QuenchResult struct {
	InletTemp          Kelvin            `json:"inlet_temp_k"`
	AdiabaticSatTemp   Kelvin            `json:"adiabatic_sat_temp_k"`
	OutletTemp         Kelvin            `json:"outlet_temp_k"`
	DewPoint           Kelvin            `json:"dew_point_k"`
	EvaporatedWater    KgPerSec          `json:"evaporated_water_kg_s"`
	EvaporatedWaterGPM float64           `json:"evaporated_water_gpm"`
	TotalSprayWater    CubicMetersPerSec `json:"total_spray_water_m3_s"`
	TotalSprayWaterGPM float64           `json:"total_spray_water_gpm"`
	OversprayFactor    float64           `json:"overspray_factor"`
	TowerDiameter      Meters            `json:"tower_diameter_m"`
	TowerHeight        Meters            `json:"tower_height_m"`
	WallThickness      Meters            `json:"wall_thickness_m"`
	GasVelocity        MetersPerSec      `json:"gas_velocity_m_s"`
	ResidenceTimeSec   float64           `json:"residence_time_s"`
	NozzleCount        int               `json:"nozzle_count"`
	PressureDrop       Pascal            `json:"pressure_drop_pa"`
	PressureDropInWC   float64           `json:"pressure_drop_in_wc"`
	Material           string            `json:"material"`
}

// EfficiencyPoint stores fractional particulate collection efficiencies (on [0, 1])
// evaluated at a single physical particle diameter d_p for both wood ash (2400 kg/m^3)
// and condensed NaCl salt fume (2160 kg/m^3).
type EfficiencyPoint struct {
	ParticleDiameter        Meters  `json:"particle_diameter_m"`
	ParticleDiameterMicrons float64 `json:"particle_diameter_um"`
	// WoodAshEfficiency is fractional capture efficiency on [0, 1] at ParticleDiameter.
	WoodAshEfficiency float64 `json:"wood_ash_efficiency"`
	// NaClFumeEfficiency is fractional capture efficiency on [0, 1] at ParticleDiameter.
	NaClFumeEfficiency float64 `json:"nacl_fume_efficiency"`
}

// VenturiResult captures Stage 2 (316L SS Wet Venturi Scrubber) geometry and particulate capture.
type VenturiResult struct {
	InletDiameter    Meters            `json:"inlet_diameter_m"`
	ThroatDiameter   Meters            `json:"throat_diameter_m"`
	OutletDiameter   Meters            `json:"outlet_diameter_m"`
	ConvergingLength Meters            `json:"converging_length_m"`
	ThroatLength     Meters            `json:"throat_length_m"`
	DivergingLength  Meters            `json:"diverging_length_m"`
	WallThickness    Meters            `json:"wall_thickness_m"`
	ThroatVelocity   MetersPerSec      `json:"throat_velocity_m_s"`
	LGRatio          float64           `json:"lg_ratio"`
	LiquidFlow       CubicMetersPerSec `json:"liquid_flow_m3_s"`
	LiquidFlowGPM    float64           `json:"liquid_flow_gpm"`
	DropletDiameter  Meters            `json:"droplet_diameter_m"`
	// CutDiameterD50 is the Calvert cut diameter d_50 (50% collection) for condensed NaCl fume (2160 kg/m^3).
	CutDiameterD50        Meters  `json:"cut_diameter_d50_m"`
	CutDiameterD50Microns float64 `json:"cut_diameter_d50_um"`
	// WoodAshCutDiameterD50 is the Calvert cut diameter d_50 for wood fly ash (2400 kg/m^3).
	WoodAshCutDiameterD50        Meters            `json:"wood_ash_cut_diameter_d50_m"`
	WoodAshCutDiameterD50Microns float64           `json:"wood_ash_cut_diameter_d50_um"`
	EfficiencyCurve              []EfficiencyPoint `json:"efficiency_curve"`
	// OverallWoodAshEfficiency is fractional collection efficiency on [0, 1].
	OverallWoodAshEfficiency float64 `json:"overall_wood_ash_efficiency"`
	// OverallNaClFumeEfficiency is fractional collection efficiency on [0, 1].
	OverallNaClFumeEfficiency float64 `json:"overall_nacl_fume_efficiency"`
	PressureDrop              Pascal  `json:"pressure_drop_pa"`
	PressureDropInWC          float64 `json:"pressure_drop_in_wc"`
	Material                  string  `json:"material"`
}

// PackedBedResult captures Stage 3 (CPVC/PP Counter-Current Packed-Bed HCl Absorber & Chevron Demister).
type PackedBedResult struct {
	ColumnDiameter         Meters       `json:"column_diameter_m"`
	ColumnHeight           Meters       `json:"column_height_m"`
	WallThickness          Meters       `json:"wall_thickness_m"`
	FloodingVelocity       MetersPerSec `json:"flooding_velocity_m_s"`
	SuperficialGasVelocity MetersPerSec `json:"superficial_gas_velocity_m_s"`
	// PercentOfFlood is superficial velocity as a percentage of flooding velocity on 0..100 scale (target <= 75.0).
	PercentOfFlood  float64     `json:"percent_of_flood"`
	PackingHeight   Meters      `json:"packing_height_m"`
	NTU             float64     `json:"ntu"`
	HTU             Meters      `json:"htu_m"`
	PackingSpec     string      `json:"packing_spec"`
	PackingDiameter Meters      `json:"packing_diameter_m"`
	PackingVoidage  float64     `json:"packing_voidage"`
	PackingVolume   CubicMeters `json:"packing_volume_m3"`
	HClInletPPM     float64     `json:"hcl_inlet_ppm"`
	HClOutletPPM    float64     `json:"hcl_outlet_ppm"`
	// HClRemovalEfficiency is fractional HCl absorption efficiency on [0, 1] (target >= 0.98).
	HClRemovalEfficiency     float64           `json:"hcl_removal_efficiency"`
	NaHCO3ConsumptionKgPerHr float64           `json:"nahco3_consumption_kg_hr"`
	RecirculationFlow        CubicMetersPerSec `json:"recirculation_flow_m3_s"`
	RecirculationFlowGPM     float64           `json:"recirculation_flow_gpm"`
	SumpPH                   float64           `json:"sump_ph"`
	SumpVolume               CubicMeters       `json:"sump_volume_m3"`
	BedPressureDrop          Pascal            `json:"bed_pressure_drop_pa"`
	DemisterPressureDrop     Pascal            `json:"demister_pressure_drop_pa"`
	DemisterHeight           Meters            `json:"demister_height_m"`
	TotalPressureDrop        Pascal            `json:"total_pressure_drop_pa"`
	TotalPressureDropInWC    float64           `json:"total_pressure_drop_in_wc"`
	Material                 string            `json:"material"`
}

// CurvePoint represents a single (flow, static pressure) point on a fan or system resistance curve.
type CurvePoint struct {
	Flow         CubicMetersPerSec `json:"flow_m3_s"`
	FlowACFM     float64           `json:"flow_acfm"`
	Pressure     Pascal            `json:"pressure_pa"`
	PressureInWC float64           `json:"pressure_in_wc"`
}

// PressureDropBreakdown itemizes static pressure losses across the scrubber train.
// PackedBed / PackedBedInWC hold the Ergun packing-bed pressure drop ONLY (excluding Demister)
// so Demister is not double-counted when summing.
// KilnDraft / KilnDraftInWC store the positive suction magnitude abs(TargetKilnDraft) needed to
// maintain the negative kiln draft setpoint, so summing all terms equals TotalStaticPressure.
type PressureDropBreakdown struct {
	Duct          Pascal  `json:"duct_pa"`
	DuctInWC      float64 `json:"duct_in_wc"`
	Quench        Pascal  `json:"quench_pa"`
	QuenchInWC    float64 `json:"quench_in_wc"`
	Venturi       Pascal  `json:"venturi_pa"`
	VenturiInWC   float64 `json:"venturi_in_wc"`
	PackedBed     Pascal  `json:"packed_bed_pa"`
	PackedBedInWC float64 `json:"packed_bed_in_wc"`
	Demister      Pascal  `json:"demister_pa"`
	DemisterInWC  float64 `json:"demister_in_wc"`
	KilnDraft     Pascal  `json:"kiln_draft_pa"`
	KilnDraftInWC float64 `json:"kiln_draft_in_wc"`
}

// TotalPa returns the sum of all pressure drop components in Pascals.
func (b PressureDropBreakdown) TotalPa() Pascal {
	return b.Duct + b.Quench + b.Venturi + b.PackedBed + b.Demister + b.KilnDraft
}

// BlowerResult captures Stage 4 (Induced-Draft Centrifugal Blower) sizing and operating curves.
type BlowerResult struct {
	Breakdown PressureDropBreakdown `json:"breakdown"`
	// TargetKilnDraftInWC echoes the signed negative draft setpoint (e.g. -0.05 inWC).
	TargetKilnDraftInWC     float64           `json:"target_kiln_draft_in_wc"`
	TotalStaticPressure     Pascal            `json:"total_static_pressure_pa"`
	TotalStaticPressureInWC float64           `json:"total_static_pressure_in_wc"`
	Flow                    CubicMetersPerSec `json:"flow_m3_s"`
	FlowACFM                float64           `json:"flow_acfm"`
	// Efficiency is fractional blower static efficiency on (0, 1] (typically 0.60-0.65).
	Efficiency         float64      `json:"efficiency"`
	ShaftPower         Watts        `json:"shaft_power_w"`
	ShaftPowerHP       float64      `json:"shaft_power_hp"`
	RecommendedMotorHP float64      `json:"recommended_motor_hp"`
	WheelDiameter      Meters       `json:"wheel_diameter_m"`
	OperatingRPM       float64      `json:"operating_rpm"`
	OperatingPoint     CurvePoint   `json:"operating_point"`
	FanCurve           []CurvePoint `json:"fan_curve"`
	SystemCurve        []CurvePoint `json:"system_curve"`
	Material           string       `json:"material"`
}

// EnvelopeDimensions describes the overall bounding footprint and height of the assembled scrubber.
type EnvelopeDimensions struct {
	Length        Meters       `json:"length_m"`
	Width         Meters       `json:"width_m"`
	Height        Meters       `json:"height_m"`
	DuctDiameter  Meters       `json:"duct_diameter_m"`
	FootprintArea SquareMeters `json:"footprint_area_m2"`
}

// ScrubberDesign is the complete, unified output of the 4-stage scrubber engineering pipeline.
type ScrubberDesign struct {
	Input       DesignInput        `json:"input"`
	InletGas    GasComposition     `json:"inlet_gas"`
	QuenchedGas GasComposition     `json:"quenched_gas"`
	OutletGas   GasComposition     `json:"outlet_gas"`
	Quench      QuenchResult       `json:"quench"`
	Venturi     VenturiResult      `json:"venturi"`
	PackedBed   PackedBedResult    `json:"packed_bed"`
	Blower      BlowerResult       `json:"blower"`
	Envelope    EnvelopeDimensions `json:"envelope"`
	Warnings    []string           `json:"warnings,omitempty"`
	Notes       []string           `json:"notes,omitempty"`
}

// BOMItem represents a single line item in the fabrication Bill of Materials.
// UnitCostUSD is per-unit cost; TotalCostUSD is Quantity * UnitCostUSD (or an explicit
// lump sum when UnitCostUSD == 0). WeightKg is the line-item TOTAL weight in kg
// (already accounting for Quantity, not per-unit weight).
type BOMItem struct {
	Category     string  `json:"category"`
	Name         string  `json:"name"`
	Material     string  `json:"material"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
	Spec         string  `json:"spec"`
	UnitCostUSD  float64 `json:"unit_cost_usd"`
	TotalCostUSD float64 `json:"total_cost_usd"`
	// WeightKg is the total weight of this line item (for all Quantity units) in kg.
	WeightKg float64 `json:"weight_kg"`
}

// BOM aggregates fabrication materials, hardware, cost, and weight estimates.
type BOM struct {
	Title         string    `json:"title"`
	Items         []BOMItem `json:"items"`
	TotalCostUSD  float64   `json:"total_cost_usd"`
	TotalWeightKg float64   `json:"total_weight_kg"`
	Notes         []string  `json:"notes,omitempty"`
}

// RecalculateTotals recomputes each item's TotalCostUSD whenever UnitCostUSD > 0
// (preserving lump-sum TotalCostUSD when UnitCostUSD == 0) and sums TotalCostUSD
// and TotalWeightKg across all BOM items.
func (b *BOM) RecalculateTotals() {
	if b == nil {
		return
	}
	var totalCost, totalWeight float64
	for i := range b.Items {
		if b.Items[i].UnitCostUSD > 0 {
			b.Items[i].TotalCostUSD = b.Items[i].Quantity * b.Items[i].UnitCostUSD
		}
		totalCost += b.Items[i].TotalCostUSD
		totalWeight += b.Items[i].WeightKg
	}
	b.TotalCostUSD = totalCost
	b.TotalWeightKg = totalWeight
}

// --- Active Control & Hardware Abstraction Layer (HAL) types ---

// Standard fail-safe interlock trip reasons.
const (
	TripNone           = ""
	TripHighQuenchTemp = "HIGH_QUENCH_OUTLET_TEMP"
	TripWaterLoss      = "QUENCH_WATER_LOSS"
	TripSensorFault    = "SENSOR_FAULT"
	TripOverDraft      = "EXCESSIVE_DRAFT"
)

// SensorReadings holds a timestamped snapshot of scrubber field instrumentation.
type SensorReadings struct {
	DraftPa             Pascal            `json:"draft_pa"`
	DraftInWC           float64           `json:"draft_in_wc"`
	QuenchInletTemp     Kelvin            `json:"quench_inlet_temp_k"`
	QuenchOutletTemp    Kelvin            `json:"quench_outlet_temp_k"`
	WaterFlow           CubicMetersPerSec `json:"water_flow_m3_s"`
	WaterFlowGPM        float64           `json:"water_flow_gpm"`
	SumpPH              float64           `json:"sump_ph"`
	BlowerSpeedFraction float64           `json:"blower_speed_fraction"`
	Timestamp           time.Time         `json:"timestamp"`
}

// ActuatorCommands represents the control outputs driven by the active scrubber controller.
type ActuatorCommands struct {
	BlowerVFD        float64 `json:"blower_vfd"`
	QuenchSprayValve float64 `json:"quench_spray_valve"`
	BypassDamperOpen bool    `json:"bypass_damper_open"`
	AlarmLatched     bool    `json:"alarm_latched"`
	TripReason       string  `json:"trip_reason,omitempty"`
}

// HardwareHAL abstracts physical or simulated scrubber sensors and actuators.
type HardwareHAL interface {
	ReadSensors(ctx context.Context) (SensorReadings, error)
	ApplyCommands(ctx context.Context, cmd ActuatorCommands) error
}

// SafetyLimits defines hard fail-safe interlock thresholds protecting downstream CPVC/PP stages.
// Hard bypass-damper trips are triggered by MaxQuenchOutletTemp, MinWaterFlow (water loss),
// and SensorFaultTimeout / read errors. MinSumpPH and MaxSumpPH are warning/alarm thresholds
// only (they do not open the hot-gas bypass damper).
type SafetyLimits struct {
	// MaxQuenchOutletTemp trips the fail-safe bypass damper to protect downstream CPVC/PP
	// packing and blower. Default is 210 F (372.04 K).
	MaxQuenchOutletTemp Kelvin `json:"max_quench_outlet_temp_k"`
	// MinWaterFlow is the minimum quench spray water flow; dropping below triggers TripWaterLoss.
	MinWaterFlow CubicMetersPerSec `json:"min_water_flow_m3_s"`
	// MinSumpPH is the low-pH dosing warning bound (warning only, does not trip bypass).
	MinSumpPH float64 `json:"min_sump_ph"`
	// MaxSumpPH is the high-pH dosing warning bound (warning only, does not trip bypass).
	MaxSumpPH           float64       `json:"max_sump_ph"`
	SensorFaultTimeout  time.Duration `json:"sensor_fault_timeout_ns"`
	MaxDraftMagnitudePa Pascal        `json:"max_draft_magnitude_pa"`
}

// DefaultSafetyLimits returns conservative fail-safe limits for a 316L/CPVC kiln scrubber.
func DefaultSafetyLimits() SafetyLimits {
	return SafetyLimits{
		MaxQuenchOutletTemp: FahrenheitToKelvin(210.0), // 372.04 K CPVC protection threshold
		MinWaterFlow:        GPMToCubicMetersPerSec(0.25),
		MinSumpPH:           6.5,
		MaxSumpPH:           9.5,
		SensorFaultTimeout:  5 * time.Second,
		MaxDraftMagnitudePa: InWCToPascal(1.0),
	}
}

// Validate checks that SafetyLimits are finite and physically consistent.
func (s SafetyLimits) Validate() error {
	if !isFinite(float64(s.MaxQuenchOutletTemp)) || s.MaxQuenchOutletTemp <= 273.15 || s.MaxQuenchOutletTemp > 500.0 {
		return fmt.Errorf("max_quench_outlet_temp_k must be finite and in (273.15, 500], got %v", s.MaxQuenchOutletTemp)
	}
	if !isFinite(float64(s.MinWaterFlow)) || s.MinWaterFlow < 0 {
		return fmt.Errorf("min_water_flow_m3_s must be finite and >= 0, got %v", s.MinWaterFlow)
	}
	if !isFinite(s.MinSumpPH) || !isFinite(s.MaxSumpPH) || s.MinSumpPH < 0 || s.MaxSumpPH > 14.0 || s.MinSumpPH >= s.MaxSumpPH {
		return fmt.Errorf("sump pH bounds must be finite with 0 <= min < max <= 14, got [%v, %v]", s.MinSumpPH, s.MaxSumpPH)
	}
	if s.SensorFaultTimeout <= 0 {
		return fmt.Errorf("sensor_fault_timeout_ns must be > 0, got %v", s.SensorFaultTimeout)
	}
	if !isFinite(float64(s.MaxDraftMagnitudePa)) || s.MaxDraftMagnitudePa <= 0 {
		return fmt.Errorf("max_draft_magnitude_pa must be finite and > 0, got %v", s.MaxDraftMagnitudePa)
	}
	return nil
}

// PIDGains configures a single PID control loop (compatible with go.einride.tech/pid).
// Note on direct-acting control sign: go.einride.tech/pid computes e = ReferenceSignal - ActualSignal.
// Because both kiln draft (higher pressure -> increase blower VFD) and quench cooling
// (higher temp -> increase spray valve) are direct-acting, controllers using positive gains
// should supply (ReferenceSignal: actual, ActualSignal: setpoint) or negate error accordingly.
type PIDGains struct {
	Kp        float64 `json:"kp"`
	Ki        float64 `json:"ki"`
	Kd        float64 `json:"kd"`
	MinOutput float64 `json:"min_output"`
	MaxOutput float64 `json:"max_output"`
	// AntiWindUpGain is the back-calculation tracking gain for pid.AntiWindupController (>= 0).
	AntiWindUpGain float64 `json:"anti_windup_gain"`
	// DerivativeFilterTimeConst is the derivative low-pass time constant (> 0 to prevent
	// divide-by-zero in go.einride.tech/pid AntiWindupController).
	DerivativeFilterTimeConst time.Duration `json:"derivative_filter_time_const_ns"`
}

// Validate checks that PIDGains are finite and safe for go.einride.tech/pid controllers.
func (g PIDGains) Validate(name string) error {
	if !isFinite(g.Kp) || g.Kp < 0 {
		return fmt.Errorf("%s.kp must be finite and >= 0, got %v", name, g.Kp)
	}
	if !isFinite(g.Ki) || g.Ki < 0 {
		return fmt.Errorf("%s.ki must be finite and >= 0, got %v", name, g.Ki)
	}
	if !isFinite(g.Kd) || g.Kd < 0 {
		return fmt.Errorf("%s.kd must be finite and >= 0, got %v", name, g.Kd)
	}
	if !isFinite(g.AntiWindUpGain) || g.AntiWindUpGain < 0 {
		return fmt.Errorf("%s.anti_windup_gain must be finite and >= 0, got %v", name, g.AntiWindUpGain)
	}
	if g.DerivativeFilterTimeConst <= 0 {
		return fmt.Errorf("%s.derivative_filter_time_const_ns must be > 0, got %v", name, g.DerivativeFilterTimeConst)
	}
	if !isFinite(g.MinOutput) || !isFinite(g.MaxOutput) || g.MinOutput < 0 || g.MaxOutput > 1.0 || g.MinOutput >= g.MaxOutput {
		return fmt.Errorf("%s output bounds must satisfy 0 <= min < max <= 1, got [%v, %v]", name, g.MinOutput, g.MaxOutput)
	}
	return nil
}

// CutoffHzToTimeConst returns the first-order RC low-pass time constant tau = 1/(2*pi*fc).
func CutoffHzToTimeConst(fc float64) time.Duration {
	if !isFinite(fc) || fc <= 0 {
		return 0
	}
	return time.Duration(math.Round(float64(time.Second) / (2.0 * math.Pi * fc)))
}

// TimeConstToCutoffHz returns the first-order low-pass -3 dB cutoff frequency fc = 1/(2*pi*tau).
func TimeConstToCutoffHz(tau time.Duration) float64 {
	if tau <= 0 {
		return 0
	}
	return 1.0 / (2.0 * math.Pi * tau.Seconds())
}

// ControlConfig defines setpoints, PID tuning, low-pass draft filtering, and safety limits.
// DraftFilterCutoffHz is the canonical low-pass filter parameter (matching pctl.NewLPF);
// DraftFilterTimeConst tracks the equivalent tau = 1/(2*pi*fc). Call Sync() after mutating
// TargetDraftInWC or DraftFilterCutoffHz directly, or use SetTargetDraftInWC().
type ControlConfig struct {
	TargetDraftInWC      float64       `json:"target_draft_in_wc"`
	TargetDraftPa        Pascal        `json:"target_draft_pa"`
	DraftPID             PIDGains      `json:"draft_pid"`
	QuenchPID            PIDGains      `json:"quench_pid"`
	DraftFilterCutoffHz  float64       `json:"draft_filter_cutoff_hz"`
	DraftFilterTimeConst time.Duration `json:"draft_filter_time_const_ns"`
	QuenchSetpoint       Kelvin        `json:"quench_setpoint_k"`
	LoopInterval         time.Duration `json:"loop_interval_ns"`
	Safety               SafetyLimits  `json:"safety"`
}

// DefaultControlConfig returns baseline tuning for closed-loop kiln draft and quench control.
func DefaultControlConfig() ControlConfig {
	targetDraftInWC := -0.05
	cutoffHz := 0.5
	return ControlConfig{
		TargetDraftInWC: targetDraftInWC,
		TargetDraftPa:   InWCToPascal(targetDraftInWC),
		DraftPID: PIDGains{
			Kp:                        0.04,
			Ki:                        0.015,
			Kd:                        0.005,
			MinOutput:                 0.15,
			MaxOutput:                 1.0,
			AntiWindUpGain:            0.5,
			DerivativeFilterTimeConst: 200 * time.Millisecond,
		},
		QuenchPID: PIDGains{
			Kp:                        0.03,
			Ki:                        0.008,
			Kd:                        0.002,
			MinOutput:                 0.20,
			MaxOutput:                 1.0,
			AntiWindUpGain:            0.5,
			DerivativeFilterTimeConst: 500 * time.Millisecond,
		},
		DraftFilterCutoffHz:  cutoffHz,
		DraftFilterTimeConst: CutoffHzToTimeConst(cutoffHz), // ~318.31 ms for 0.5 Hz
		QuenchSetpoint:       FahrenheitToKelvin(165.0),
		LoopInterval:         100 * time.Millisecond,
		Safety:               DefaultSafetyLimits(),
	}
}

// SetTargetDraftInWC updates both TargetDraftInWC and TargetDraftPa consistently.
func (c *ControlConfig) SetTargetDraftInWC(inwc float64) {
	if c == nil {
		return
	}
	c.TargetDraftInWC = inwc
	c.TargetDraftPa = InWCToPascal(inwc)
}

// Sync reconciles derived fields in ControlConfig:
//   - TargetDraftPa is updated from TargetDraftInWC (or TargetDraftInWC from TargetDraftPa if InWC == 0 && Pa != 0).
//   - DraftFilterTimeConst is derived from canonical DraftFilterCutoffHz (or CutoffHz from TimeConst if CutoffHz == 0 && TimeConst > 0).
func (c *ControlConfig) Sync() {
	if c == nil {
		return
	}
	if c.TargetDraftInWC != 0 || c.TargetDraftPa == 0 {
		c.TargetDraftPa = InWCToPascal(c.TargetDraftInWC)
	} else {
		c.TargetDraftInWC = c.TargetDraftPa.InWC()
	}
	if c.DraftFilterCutoffHz > 0 {
		c.DraftFilterTimeConst = CutoffHzToTimeConst(c.DraftFilterCutoffHz)
	} else if c.DraftFilterTimeConst > 0 {
		c.DraftFilterCutoffHz = TimeConstToCutoffHz(c.DraftFilterTimeConst)
	}
}

// Validate checks that ControlConfig fields are finite, consistent, and safe.
func (c ControlConfig) Validate() error {
	if !isFinite(c.TargetDraftInWC) || c.TargetDraftInWC > 0 || c.TargetDraftInWC < -2.0 {
		return fmt.Errorf("target_draft_in_wc must be finite and in [-2.0, 0], got %v", c.TargetDraftInWC)
	}
	if !isFinite(float64(c.TargetDraftPa)) {
		return fmt.Errorf("target_draft_pa must be finite, got %v", c.TargetDraftPa)
	}
	expectedPa := InWCToPascal(c.TargetDraftInWC)
	if math.Abs(float64(c.TargetDraftPa-expectedPa)) > 1e-3 {
		return fmt.Errorf("target_draft_pa (%.4f Pa) disagrees with target_draft_in_wc (%.4f inWC -> %.4f Pa); call Sync() or SetTargetDraftInWC()",
			c.TargetDraftPa, c.TargetDraftInWC, expectedPa)
	}
	if !isFinite(c.DraftFilterCutoffHz) || c.DraftFilterCutoffHz <= 0 {
		return fmt.Errorf("draft_filter_cutoff_hz must be finite and > 0, got %v", c.DraftFilterCutoffHz)
	}
	if c.DraftFilterTimeConst <= 0 {
		return fmt.Errorf("draft_filter_time_const_ns must be > 0, got %v", c.DraftFilterTimeConst)
	}
	expectedTau := CutoffHzToTimeConst(c.DraftFilterCutoffHz)
	if expectedTau > 0 {
		relErr := math.Abs(c.DraftFilterTimeConst.Seconds()-expectedTau.Seconds()) / expectedTau.Seconds()
		if relErr > 0.01 {
			return fmt.Errorf("draft_filter_time_const (%v) disagrees with draft_filter_cutoff_hz (%.4f Hz -> %v); call Sync()",
				c.DraftFilterTimeConst, c.DraftFilterCutoffHz, expectedTau)
		}
	}
	if err := c.Safety.Validate(); err != nil {
		return fmt.Errorf("safety: %w", err)
	}
	if !isFinite(float64(c.QuenchSetpoint)) || c.QuenchSetpoint <= 273.15 || c.QuenchSetpoint >= c.Safety.MaxQuenchOutletTemp {
		return fmt.Errorf("quench_setpoint_k (%.2f K) must be finite and in (273.15, max_quench_outlet_temp_k=%.2f K)",
			c.QuenchSetpoint, c.Safety.MaxQuenchOutletTemp)
	}
	if c.LoopInterval <= 0 {
		return fmt.Errorf("loop_interval_ns must be > 0, got %v", c.LoopInterval)
	}
	if err := c.DraftPID.Validate("draft_pid"); err != nil {
		return err
	}
	if err := c.QuenchPID.Validate("quench_pid"); err != nil {
		return err
	}
	return nil
}

// ControlStatus captures live telemetry from an active controller tick.
type ControlStatus struct {
	State             string           `json:"state"`
	FilteredDraftPa   Pascal           `json:"filtered_draft_pa"`
	FilteredDraftInWC float64          `json:"filtered_draft_in_wc"`
	Readings          SensorReadings   `json:"readings"`
	Commands          ActuatorCommands `json:"commands"`
	InterlockActive   bool             `json:"interlock_active"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

// --- Service Interfaces for Cross-Package Decoupling ---

// Designer computes a complete 4-stage scrubber engineering design from input parameters.
type Designer interface {
	Design(input DesignInput) (ScrubberDesign, error)
}

// CADExporter generates fabrication geometry (.stl, .dxf) and Bills of Materials from a design.
type CADExporter interface {
	ExportSTL(design ScrubberDesign) ([]byte, error)
	ExportDXF(design ScrubberDesign) ([]byte, error)
	GenerateBOM(design ScrubberDesign) (BOM, error)
	FormatBOMCSV(bom BOM) ([]byte, error)
	FormatBOMMarkdown(bom BOM) ([]byte, error)
}
