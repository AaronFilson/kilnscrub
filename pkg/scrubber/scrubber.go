package scrubber

import (
	"fmt"
	"math"

	"github.com/AaronFilson/kilnscrub/pkg/model"
	"github.com/AaronFilson/kilnscrub/pkg/thermo"
)

// Compile-time check that *Designer satisfies model.Designer.
var _ model.Designer = (*Designer)(nil)

// Designer implements model.Designer to compute a complete 4-stage studio kiln wet scrubber design.
type Designer struct{}

// NewDesigner returns a new concrete scrubber engineering Designer.
func NewDesigner() *Designer {
	return &Designer{}
}

// Combustion & kiln firing stoichiometry constants.
const (
	// Reference studio ceramic kiln peak firing heat load (~18,400 BTU/hr per cu.ft. at 2400 F).
	refHeatLoadWattsPerCuFt = 5400.0

	// Wood combustion (seasoned hardwood, ~15% moisture as-fired, dry formula C1 H1.44 O0.66):
	// Dry unit MW = 24.022 g/mol C; as-fired (15% H2O) = 28.261 g/mol C; LHV_as_fired = 16.0 MJ/kg.
	woodEnergyJPerMolC  = 16.0e6 * 0.028261 // ~452,176 J per mol carbon burned
	woodStoichO2PerMolC = 1.03              // 1 + 1.44/4 - 0.66/2
	woodCombH2OPerMolC  = 0.72              // 1.44 / 2
	woodMoistH2OPerMolC = 0.235             // 15% moisture by mass

	// Natural gas combustion (methane CH4, LHV = 802.3 kJ/mol).
	methaneLHVJPerMol     = 802300.0
	methaneStoichO2PerMol = 2.0

	// Phase 16 solid/dirty urban pottery fuels:
	// 1. Bituminous kiln coal (Khurja/India, Bát Tràng/Vietnam, UK/EU: ~15% ash, 6% moisture, LHV ~25.5 MJ/kg,
	//    dry-ash-free formula C1 H0.78 O0.09, ~68% C by mass as-fired -> 0.01765 kg coal / mol C).
	coalKgPerMolC       = 0.01765
	coalEnergyJPerMolC  = 25.5e6 * coalKgPerMolC // ~450,075 J/mol C
	coalStoichO2PerMolC = 1.150                  // 1 + 0.78/4 - 0.09/2 (plus sulfur O2 added explicitly)
	coalCombH2OPerMolC  = 0.39
	coalMoistH2OPerMolC = 0.059

	// 2. Air-dried peat / turf / Andean turba (~30% moisture, 5% ash, LHV ~13.5 MJ/kg as-fired,
	//    ~36.6% C by mass as-fired -> 0.0328 kg peat / mol C; heavy phenolic tar & high H2O vapor).
	peatKgPerMolC       = 0.0328
	peatEnergyJPerMolC  = 13.5e6 * peatKgPerMolC // ~442,800 J/mol C
	peatStoichO2PerMolC = 1.0675                 // 1 + 1.35/4 - 0.54/2
	peatCombH2OPerMolC  = 0.675
	peatMoistH2OPerMolC = 0.546 // 30% as-fired moisture

	// 3. Agricultural waste biomass (Palm Kernel Shells / PKS, Olive Pomace / Jift, Rice Husks:
	//    ~12% moisture, 8% silica ash, LHV ~16.5 MJ/kg as-fired -> 0.0272 kg fuel / mol C).
	biomassAgKgPerMolC       = 0.0272
	biomassAgEnergyJPerMolC  = 16.5e6 * biomassAgKgPerMolC // ~448,800 J/mol C
	biomassAgStoichO2PerMolC = 1.05
	biomassAgCombH2OPerMolC  = 0.68
	biomassAgMoistH2OPerMolC = 0.181

	sulfurMWKgPerMol = 0.032065

	// Atmospheric air molar N2/O2 ratio (79% N2 / 21% O2).
	airN2PerO2 = 0.79 / 0.21

	// Salt volatilization stoichiometry: 2 NaCl + H2O + SiO2 -> Na2SiO3 + 2 HCl (1 mol HCl per mol NaCl).
	naclMWKgPerMol = 0.0584428
)

// Exact reference firing temperature rise from 20 C (68 F / 293.15 K) to 2400 F (1588.7056 K).
var refPeakDeltaTK = float64(model.FahrenheitToKelvin(2400.0)) - 293.15

// computeInletFlueGas derives kiln exit gas composition, molar flow, and thermophysical state
// from firing volume, temperature, fuel type, excess air, and salt introduction rate.
func computeInletFlueGas(input model.DesignInput) (model.GasComposition, []string, error) {
	var warnings []string

	fuel := input.FuelType
	if fuel == "" {
		fuel = model.FuelWoodSalt
	}

	peakK := model.FahrenheitToKelvin(input.PeakExitTempF)
	deltaT := float64(peakK - input.AmbientTempK)
	tempScale := deltaT / refPeakDeltaTK
	if tempScale < 0.35 {
		tempScale = 0.35
	}
	if tempScale > 1.80 {
		tempScale = 1.80
	}
	thermalWatts := input.KilnVolumeCuFt * refHeatLoadWattsPerCuFt * tempScale

	e := input.ExcessAirFraction
	var nN2, nCO2, nH2O, nO2, nSO2 float64

	sulfurPct := input.FuelSulfurPct
	if sulfurPct <= 0 {
		sulfurPct = model.DefaultFuelSulfurPct(fuel)
	}

	switch fuel {
	case model.FuelWoodSalt, model.FuelWood:
		nC := thermalWatts / woodEnergyJPerMolC
		nCO2 = nC
		nH2O = nC * (woodCombH2OPerMolC + woodMoistH2OPerMolC)
		nO2 = nC * woodStoichO2PerMolC * e
		nN2 = nC * woodStoichO2PerMolC * (1.0 + e) * airN2PerO2
	case model.FuelGasSalt, model.FuelGas:
		nCH4 := thermalWatts / methaneLHVJPerMol
		nCO2 = nCH4
		nH2O = 2.0 * nCH4
		nO2 = nCH4 * methaneStoichO2PerMol * e
		nN2 = nCH4 * methaneStoichO2PerMol * (1.0 + e) * airN2PerO2
	case model.FuelCoalSalt, model.FuelCoal:
		nC := thermalWatts / coalEnergyJPerMolC
		fuelMassKgS := nC * coalKgPerMolC
		nSO2 = (fuelMassKgS * (sulfurPct / 100.0)) / sulfurMWKgPerMol
		stoichO2 := nC*coalStoichO2PerMolC + nSO2
		nCO2 = nC
		nH2O = nC * (coalCombH2OPerMolC + coalMoistH2OPerMolC)
		nO2 = stoichO2 * e
		nN2 = stoichO2 * (1.0 + e) * airN2PerO2
	case model.FuelPeatSalt, model.FuelPeat:
		nC := thermalWatts / peatEnergyJPerMolC
		fuelMassKgS := nC * peatKgPerMolC
		nSO2 = (fuelMassKgS * (sulfurPct / 100.0)) / sulfurMWKgPerMol
		stoichO2 := nC*peatStoichO2PerMolC + nSO2
		nCO2 = nC
		nH2O = nC * (peatCombH2OPerMolC + peatMoistH2OPerMolC)
		nO2 = stoichO2 * e
		nN2 = stoichO2 * (1.0 + e) * airN2PerO2
	case model.FuelBiomassAgSalt, model.FuelBiomassAg:
		nC := thermalWatts / biomassAgEnergyJPerMolC
		fuelMassKgS := nC * biomassAgKgPerMolC
		nSO2 = (fuelMassKgS * (sulfurPct / 100.0)) / sulfurMWKgPerMol
		stoichO2 := nC*biomassAgStoichO2PerMolC + nSO2
		nCO2 = nC
		nH2O = nC * (biomassAgCombH2OPerMolC + biomassAgMoistH2OPerMolC)
		nO2 = stoichO2 * e
		nN2 = stoichO2 * (1.0 + e) * airN2PerO2
	default:
		return model.GasComposition{}, nil, fmt.Errorf("unsupported fuel type %q", fuel)
	}

	var nHCl float64
	if input.SaltRateLbsPerHr > 0 {
		nHCl = float64(model.LbsPerHrToKgPerSec(input.SaltRateLbsPerHr)) / naclMWKgPerMol
		if fuel == model.FuelWood || fuel == model.FuelGas || fuel == model.FuelCoal || fuel == model.FuelPeat || fuel == model.FuelBiomassAg {
			warnings = append(warnings, fmt.Sprintf("Non-salt fuel %q specified with salt_rate=%.2f lbs/hr; HCl loading is conservatively included in absorber sizing.", fuel, input.SaltRateLbsPerHr))
		}
	}

	nTotal := nN2 + nCO2 + nH2O + nO2 + nHCl + nSO2
	if nTotal <= 0 {
		return model.GasComposition{}, nil, fmt.Errorf("computed flue gas molar flow must be > 0")
	}

	gas := model.GasComposition{
		N2:                 nN2 / nTotal,
		CO2:                nCO2 / nTotal,
		H2O:                nH2O / nTotal,
		O2:                 nO2 / nTotal,
		HCl:                nHCl / nTotal,
		SO2:                nSO2 / nTotal,
		MolarFlowMolPerSec: nTotal,
		Temperature:        peakK,
		Pressure:           input.AmbientPressurePa,
	}
	if err := thermo.PopulateGasProperties(&gas); err != nil {
		return model.GasComposition{}, nil, fmt.Errorf("populate inlet flue gas: %w", err)
	}
	return gas, warnings, nil
}

// Design validates input and computes the complete 4-stage kiln wet scrubber engineering design:
// Stage 1 Adiabatic Quench -> Stage 2 Wet Venturi -> Stage 3 Packed-Bed Absorber -> Stage 4 ID Blower.
func (d *Designer) Design(input model.DesignInput) (model.ScrubberDesign, error) {
	if input.FuelType == "" {
		input.FuelType = model.FuelWoodSalt
	}
	if err := input.Validate(); err != nil {
		return model.ScrubberDesign{}, fmt.Errorf("invalid design input: %w", err)
	}

	inletGas, warnings, err := computeInletFlueGas(input)
	if err != nil {
		return model.ScrubberDesign{}, err
	}

	ductDiam := input.DuctDiameter
	if ductDiam <= 0 {
		ductDiam = AutoSizeDuctDiameter(inletGas)
	}

	// Stage 1: 316L SS Adiabatic Spray Quench Tower.
	quenchedGas, quenchRes, err := thermo.SolveAdiabaticQuench(inletGas, input.SprayWaterTempK, input.AmbientPressurePa)
	if err != nil {
		return model.ScrubberDesign{}, fmt.Errorf("stage 1 quench: %w", err)
	}

	// Stage 2: 316L SS Wet Venturi Particulate Scrubber.
	venturiRes, err := SizeVenturi(quenchedGas, input, ductDiam)
	if err != nil {
		return model.ScrubberDesign{}, fmt.Errorf("stage 2 venturi: %w", err)
	}

	// Stage 3: CPVC/PP Counter-Current Packed-Bed HCl Absorber & Chevron Demister.
	outletGas, packedRes, err := SizePackedBed(quenchedGas, input)
	if err != nil {
		return model.ScrubberDesign{}, fmt.Errorf("stage 3 packed bed: %w", err)
	}

	// Stage 4: Induced-Draft Centrifugal Blower.
	blowerRes, err := SizeBlower(inletGas, outletGas, quenchRes, venturiRes, packedRes, input, ductDiam)
	if err != nil {
		return model.ScrubberDesign{}, fmt.Errorf("stage 4 blower: %w", err)
	}

	// Assembled skid envelope dimensions.
	venturiTotalLen := float64(venturiRes.ConvergingLength + venturiRes.ThroatLength + venturiRes.DivergingLength)
	envLength := model.Meters(float64(quenchRes.TowerDiameter) + venturiTotalLen + float64(packedRes.ColumnDiameter) + 1.6*float64(blowerRes.WheelDiameter) + 0.50)
	envWidthVal := math.Max(float64(quenchRes.TowerDiameter), float64(packedRes.ColumnDiameter))
	envWidthVal = math.Max(envWidthVal, 1.5*float64(blowerRes.WheelDiameter)) + 0.40
	envWidth := model.Meters(envWidthVal)
	envHeight := model.Meters(math.Max(float64(quenchRes.TowerHeight), float64(packedRes.ColumnHeight)) + 0.30)

	envelope := model.EnvelopeDimensions{
		Length:        envLength,
		Width:         envWidth,
		Height:        envHeight,
		DuctDiameter:  ductDiam,
		FootprintArea: model.SquareMeters(float64(envLength) * float64(envWidth)),
	}

	// Engineering safety & operational warnings.
	if quenchRes.OutletTemp >= model.FahrenheitToKelvin(200.0) {
		warnings = append(warnings, fmt.Sprintf("Quench outlet temperature (%.1f F) is close to the 210 F CPVC safety interlock threshold.", quenchRes.OutletTemp.Fahrenheit()))
	}
	if venturiRes.ThroatVelocity < 50.0 || venturiRes.ThroatVelocity > 90.0 {
		warnings = append(warnings, fmt.Sprintf("Venturi throat velocity (%.1f m/s) is outside the recommended studio range of 50-90 m/s.", venturiRes.ThroatVelocity))
	}
	if packedRes.PercentOfFlood > 70.0 {
		warnings = append(warnings, fmt.Sprintf("Packed-bed superficial velocity is %.1f%% of flooding (approaching 75%% upper design limit).", packedRes.PercentOfFlood))
	}
	if quenchRes.TotalSprayWaterGPM < 0.50 {
		warnings = append(warnings, fmt.Sprintf("Total quench spray water (%.2f GPM) is near the 0.25 GPM minimum-flow interlock threshold.", quenchRes.TotalSprayWaterGPM))
	}
	if packedRes.SO2InletPPM > 10.0 {
		warnings = append(warnings, fmt.Sprintf("SULFUR / H2SO4 DEW POINT (%.0f ppm SO2 in -> %.0f ppm out, %.1f%% capture): ~1.5%% SO2->SO3 oxidation raises the sulfuric-acid dew point to ~%.0f F (vs 184 F water/HCl) — protect pre-quench flues with refractory masonry or ceramic coating and neutralize sulfite/bisulfite in an alkaline pH 7.5-8.5 sump.",
			packedRes.SO2InletPPM, packedRes.SO2OutletPPM, packedRes.SO2RemovalEfficiency*100.0, packedRes.SulfurAcidDewPointF))
	}
	if model.IsSolidDirtyFuel(input.FuelType) {
		warnings = append(warnings, fmt.Sprintf("TAR & HEAVY FLY-ASH FOULING GUARD (%s): Coal, peat/turf, and ag-biomass emit sticky phenolic/coal tars and 8-20%% fly ash that plug fine packing — install a hot (>750 C) secondary-air masonry smoke-burner/afterburner throat ahead of Stage 1 and skim tar/ash from Sump Pool #1.", input.FuelType))
	}

	notes := []string{
		fmt.Sprintf("Material boundary: Stage 1 Quench & Stage 2 Venturi fabricated in %s; Stage 3 Absorber in %s (protected by 210 F bypass interlock).", quenchRes.Material, packedRes.Material),
		fmt.Sprintf("Quench cools %.0f F flue gas to %.1f F adiabatic saturation (evaporating %.2f GPM, spraying %.2f GPM at %.1fx overspray).",
			input.PeakExitTempF, quenchRes.OutletTemp.Fahrenheit(), quenchRes.EvaporatedWaterGPM, quenchRes.TotalSprayWaterGPM, quenchRes.OversprayFactor),
		fmt.Sprintf("Venturi achieves d50 = %.2f um (NaCl fume) / %.2f um (wood ash) at %.1f inWC pressure drop.",
			venturiRes.CutDiameterD50Microns, venturiRes.WoodAshCutDiameterD50Microns, venturiRes.PressureDropInWC),
		fmt.Sprintf("Packed bed operates at %.1f%% of Eckert flooding with %.2f%% HCl removal (%.2f kg/hr NaHCO3 at pH %.1f).",
			packedRes.PercentOfFlood, packedRes.HClRemovalEfficiency*100.0, packedRes.NaHCO3ConsumptionKgPerHr, packedRes.SumpPH),
		fmt.Sprintf("ID Blower delivers %.0f ACFM at %.2f inWC total static pressure (%.2f shaft HP -> %.2f HP NEMA motor).",
			blowerRes.FlowACFM, blowerRes.TotalStaticPressureInWC, blowerRes.ShaftPowerHP, blowerRes.RecommendedMotorHP),
	}

	return model.ScrubberDesign{
		Input:       input,
		InletGas:    inletGas,
		QuenchedGas: quenchedGas,
		OutletGas:   outletGas,
		Quench:      quenchRes,
		Venturi:     venturiRes,
		PackedBed:   packedRes,
		Blower:      blowerRes,
		Envelope:    envelope,
		Warnings:    warnings,
		Notes:       notes,
	}, nil
}
