package scrubber

import (
	"fmt"
	"math"

	"github.com/AaronFilson/kilnscrub/pkg/model"
	"github.com/AaronFilson/kilnscrub/pkg/thermo"
)

// Stage 3 packed-bed HCl absorber & chevron mist eliminator engineering constants.
const (
	packedBedTargetFloodFrac     = 0.65   // Design target 65% of flooding velocity (safely <= 75% constraint)
	packedBedMaxFloodPct         = 75.0   // Hard upper bound on PercentOfFlood (0..100 scale)
	packedBedMinDiamM            = 0.25   // Minimum practical CPVC/PP column diameter [m]
	packedBedMinHeightM          = 0.60   // Minimum practical packing depth [m]
	packedBedLGRatioVol          = 0.0028 // Counter-current scrubbing liquor-to-gas ratio [m^3/m^3] (~21 gal/1000 ft^3)
	packedBedMinWettingFlux      = 0.0014 // Minimum wetting liquid flux [m^3/(m^2*s)] (~2.0 GPM/ft^2)
	packedBedSumpResTimeSec      = 240.0  // 4-minute recirculation sump holdup time [s]
	packedBedMinSumpVolM3        = 0.06   // Minimum sump volume [m^3] (~16 US gal)
	packedBedWallThinM           = 0.006  // 6 mm CPVC/PP shell thickness
	packedBedWallThickM          = 0.008  // 8 mm CPVC/PP shell thickness for larger columns
	packedBedDemisterHeightM     = 0.20   // 3-pass polypropylene chevron vane mist eliminator height [m]
	packedBedDemisterLossK       = 55.0   // 3-pass chevron mist eliminator velocity-head loss factor
	packedBedMaterialSpec        = "CPVC / Polypropylene"
	packedBedRefSpecificAreaM2M3 = 207.0     // Specific surface area a_p [m^2/m^3] for 1" (0.0254 m) PP Pall rings
	packedBedRefNominalDiamM     = 0.0254    // Reference nominal ring size [m] (1 inch)
	packedBedIrrigatedFactor     = 1.70      // Eckert GPDC irrigated-to-dry liquid holdup pressure drop multiplier
	nahco3MWKgPerMol             = 0.0840066 // NaHCO3 molar mass [kg/mol]
	nahco3ExcessFactor           = 1.15      // 15% excess alkalinity buffer for pH 7.5-8.5 control
	liquidWaterViscosityPaSec    = 0.001     // Reference scrubbing solution viscosity at ~20 C [Pa*s]
)

// ErgunPressureDrop computes total dry packed-bed pressure drop [Pa] using the Ergun equation,
// matching fluids.packed_bed.dP_packed_bed(..., Method="Ergun") to machine precision:
//
//	dP/L = 150 * ((1 - eps)^2 / eps^3) * (mu * vs / dp^2) + 1.75 * ((1 - eps) / eps^3) * (rho * vs^2 / dp)
func ErgunPressureDrop(dp, voidage, vs, rho, mu, length float64) (model.Pascal, error) {
	if dp <= 0 || voidage <= 0 || voidage >= 1.0 || vs < 0 || rho < 0 || mu < 0 || length < 0 {
		return 0, fmt.Errorf("invalid Ergun parameters: dp=%v, voidage=%v, vs=%v, rho=%v, mu=%v, L=%v", dp, voidage, vs, rho, mu, length)
	}
	eps3 := voidage * voidage * voidage
	oneMinusEps := 1.0 - voidage
	viscous := 150.0 * (oneMinusEps * oneMinusEps / eps3) * (mu * vs / (dp * dp))
	inertial := 1.75 * (oneMinusEps / eps3) * (rho * vs * vs / dp)
	return model.Pascal((viscous + inertial) * length), nil
}

// EckertFloodingVelocity estimates the column flooding superficial gas velocity [m/s]
// via the Souders-Brown / Eckert Generalized Pressure Drop Correlation (GPDC).
func EckertFloodingVelocity(
	liquidMassFlowKgS float64,
	gasMassFlowKgS float64,
	rhoGas float64,
	rhoLiquid float64,
	packingFactor1PerM float64,
	muLiquidPaS float64,
) (model.MetersPerSec, error) {
	if gasMassFlowKgS <= 0 || rhoGas <= 0 || rhoLiquid <= 0 || packingFactor1PerM <= 0 || muLiquidPaS <= 0 {
		return 0, fmt.Errorf("invalid Eckert flooding inputs")
	}

	// Eckert flow parameter X = (L/G)_mass * sqrt(rho_g / rho_L).
	x := (liquidMassFlowKgS / gasMassFlowKgS) * math.Sqrt(rhoGas/rhoLiquid)
	// Clamp X to the validated Eckert GPDC correlation range [0.01, 5.0].
	if x < 0.01 {
		x = 0.01
	}
	if x > 5.0 {
		x = 5.0
	}
	lnX := math.Log(x)
	// Polynomial representation of the Eckert / Sherwood-Lobo-Leva GPDC flooding curve:
	// ln(Y_flood) = -3.7121 - 1.0371*ln(X) - 0.1501*(ln(X))^2 - 0.007544*(ln(X))^3
	lnY := -3.7121 - 1.0371*lnX - 0.1501*lnX*lnX - 0.007544*lnX*lnX*lnX
	yFlood := math.Exp(lnY)

	// Y_flood = (v_flood^2 * F_p * (mu_L / mu_w)^0.2 * (rho_g / rho_L)) / g
	viscRatio := math.Pow(muLiquidPaS/liquidWaterViscosityPaSec, 0.2)
	vFloodSq := (yFlood * model.StandardGravity * (rhoLiquid / rhoGas)) / (packingFactor1PerM * viscRatio)
	if vFloodSq <= 0 || math.IsNaN(vFloodSq) || math.IsInf(vFloodSq, 0) {
		return 0, fmt.Errorf("non-positive Eckert flooding velocity squared: %v", vFloodSq)
	}
	return model.MetersPerSec(math.Sqrt(vFloodSq)), nil
}

// SizePackedBed designs Stage 3 (CPVC/PP counter-current packed-bed HCl absorber and
// chevron mist eliminator) and returns the treated outlet gas stream and PackedBedResult.
func SizePackedBed(
	quenchedGas model.GasComposition,
	input model.DesignInput,
) (model.GasComposition, model.PackedBedResult, error) {
	if quenchedGas.VolumetricFlow <= 0 || quenchedGas.MassFlow <= 0 || quenchedGas.Density <= 0 || quenchedGas.Viscosity <= 0 {
		return model.GasComposition{}, model.PackedBedResult{}, fmt.Errorf("quenched gas flow and properties must be > 0")
	}

	qGas := float64(quenchedGas.VolumetricFlow)
	gMass := float64(quenchedGas.MassFlow)
	rhoG := float64(quenchedGas.Density)
	muG := float64(quenchedGas.Viscosity)
	rhoL := float64(model.WaterDensity)
	fp := input.PackingFactor

	// Iteratively solve column diameter and recirculation liquid flow so both minimum wetting
	// and Eckert flooding constraint (PercentOfFlood <= 75%) are simultaneously satisfied.
	qRecirc := packedBedLGRatioVol * qGas
	var colDiam, colArea, vFlood, vSup, pctFlood float64

	for iter := 0; iter < 8; iter++ {
		lMass := qRecirc * rhoL
		vf, err := EckertFloodingVelocity(lMass, gMass, rhoG, rhoL, fp, liquidWaterViscosityPaSec)
		if err != nil {
			return model.GasComposition{}, model.PackedBedResult{}, err
		}
		vFlood = float64(vf)
		vDesign := packedBedTargetFloodFrac * vFlood
		reqArea := qGas / vDesign
		rawDiam := math.Sqrt(4.0 * reqArea / math.Pi)
		colDiam = math.Ceil(rawDiam/0.025) * 0.025
		if colDiam < packedBedMinDiamM {
			colDiam = packedBedMinDiamM
		}
		colArea = math.Pi * colDiam * colDiam / 4.0
		vSup = qGas / colArea
		pctFlood = (vSup / vFlood) * 100.0

		// Enforce minimum wetting liquid flux across the column cross-section.
		minRecirc := packedBedMinWettingFlux * colArea
		nextRecirc := math.Max(packedBedLGRatioVol*qGas, minRecirc)
		if math.Abs(nextRecirc-qRecirc)/math.Max(nextRecirc, 1e-12) < 1e-4 && pctFlood <= packedBedMaxFloodPct {
			qRecirc = nextRecirc
			break
		}
		qRecirc = nextRecirc
	}

	// Recompute final flooding velocity with converged recirculation rate and enforce <= 75% strictly.
	vfFinal, err := EckertFloodingVelocity(qRecirc*rhoL, gMass, rhoG, rhoL, fp, liquidWaterViscosityPaSec)
	if err != nil {
		return model.GasComposition{}, model.PackedBedResult{}, err
	}
	vFlood = float64(vfFinal)
	for step := 0; step < 200 && (qGas/(math.Pi*colDiam*colDiam/4.0))/vFlood*100.0 > packedBedMaxFloodPct; step++ {
		colDiam += 0.025
		colArea = math.Pi * colDiam * colDiam / 4.0
		qRecirc = math.Max(packedBedLGRatioVol*qGas, packedBedMinWettingFlux*colArea)
		vfFinal, err = EckertFloodingVelocity(qRecirc*rhoL, gMass, rhoG, rhoL, fp, liquidWaterViscosityPaSec)
		if err != nil {
			return model.GasComposition{}, model.PackedBedResult{}, err
		}
		vFlood = float64(vfFinal)
	}
	colArea = math.Pi * colDiam * colDiam / 4.0
	vSup = qGas / colArea
	pctFlood = (vSup / vFlood) * 100.0

	// NTU-HTU sizing for gas-film-controlled HCl absorption into buffered NaHCO3 (pH 7.5-8.5).
	ntuRequired := -math.Log(1.0 - input.TargetHClRemoval)
	lFlux := qRecirc / colArea
	// Cornell/Treybal-form gas-phase HTU correlation scaled to 1" Pall rings (HTU_ref = 0.24 m).
	dpRatio := float64(input.PackingDiameter) / 0.0254
	htu := 0.24 * math.Pow(dpRatio, 0.35) * math.Pow(math.Max(vSup, 0.2)/1.5, 0.25) * math.Pow(0.003/math.Max(lFlux, 0.0005), 0.20)
	if htu < 0.15 {
		htu = 0.15
	}
	if htu > 0.60 {
		htu = 0.60
	}

	rawPackingHeight := math.Max(ntuRequired*htu, packedBedMinHeightM)
	packingHeight := math.Ceil(rawPackingHeight/0.05) * 0.05
	ntuActual := packingHeight / htu
	hclEff := 1.0 - math.Exp(-ntuActual)
	if hclEff > 0.999999 {
		hclEff = 0.999999
	}

	// HCl & SO2 mass balance and combined NaHCO3 reagent consumption:
	//   HCl + NaHCO3 -> NaCl + H2O + CO2        (1.0 mol NaHCO3 / mol HCl)
	//   SO2 + 2 NaHCO3 -> Na2SO3 + H2O + 2 CO2  (2.0 mol NaHCO3 / mol SO2)
	// SO2 absorption has mixed gas/liquid-film resistance in buffered pH 7.5-8.5 liquor
	// (HTU_SO2 ~ 1.45 * HTU_HCl -> ~95.8% SO2 removal on a 99% HCl packed bed).
	hclInletPPM := quenchedGas.HCl * 1e6
	hclOutletPPM := hclInletPPM * (1.0 - hclEff)
	nHClIn := quenchedGas.MolarFlowMolPerSec * quenchedGas.HCl
	nHClRemoved := nHClIn * hclEff

	var so2InletPPM, so2OutletPPM, so2Eff, sulfurAcidDewF, nSO2Removed float64
	nSO2In := quenchedGas.MolarFlowMolPerSec * quenchedGas.SO2
	if quenchedGas.SO2 > 0 {
		so2InletPPM = quenchedGas.SO2 * 1e6
		ntuSO2 := ntuActual / 1.45
		so2Eff = 1.0 - math.Exp(-ntuSO2)
		if so2Eff > 0.995 {
			so2Eff = 0.995
		}
		so2OutletPPM = so2InletPPM * (1.0 - so2Eff)
		nSO2Removed = nSO2In * so2Eff

		// Verhoff-Banchero (1974) sulfuric acid dew point [K] assuming 1.5% SO2->SO3 oxidation.
		pTotalMmHg := float64(quenchedGas.Pressure) / 133.322368
		pH2OMmHg := math.Max(1.0, quenchedGas.H2O*pTotalMmHg)
		pSO3MmHg := math.Max(1e-5, 0.015*quenchedGas.SO2*pTotalMmHg)
		lnH2O := math.Log(pH2OMmHg)
		lnSO3 := math.Log(pSO3MmHg)
		invT := 2.276 - 0.0294*lnH2O - 0.0858*lnSO3 + 0.0062*lnH2O*lnSO3
		if invT > 1.5 && invT < 3.5 {
			sulfurAcidDewF = model.Kelvin(1000.0 / invT).Fahrenheit()
		} else {
			sulfurAcidDewF = 285.0
		}
	}

	nahco3KgPerHr := (nHClRemoved + 2.0*nSO2Removed) * nahco3ExcessFactor * nahco3MWKgPerMol * 3600.0

	// Irrigated packing bed pressure drop using Sauter equivalent sphere diameter
	// d_p,eq = 6*(1 - eps)/a_p (scaled from 207 m^2/m^3 for open PP Pall rings, or 185 m^2/m^3
	// for denser ceramic Raschig rings when voidage <= 0.78) multiplied by the Eckert GPDC
	// irrigated liquid-holdup factor, plus 3-pass chevron demister pressure drop.
	refSpecArea := packedBedRefSpecificAreaM2M3
	if input.PackingVoidage <= 0.78 {
		refSpecArea = 185.0
	}
	specificArea := refSpecArea * (packedBedRefNominalDiamM / float64(input.PackingDiameter))
	sauterDp := 6.0 * (1.0 - input.PackingVoidage) / specificArea
	dryBedDP, err := ErgunPressureDrop(sauterDp, input.PackingVoidage, vSup, rhoG, muG, packingHeight)
	if err != nil {
		return model.GasComposition{}, model.PackedBedResult{}, fmt.Errorf("packed-bed Ergun dP: %w", err)
	}
	bedDP := model.Pascal(float64(dryBedDP) * packedBedIrrigatedFactor)
	demisterDP := model.Pascal(math.Max(35.0, packedBedDemisterLossK*0.5*rhoG*vSup*vSup))
	totalDP := bedDP + demisterDP

	// Column sump & vertical envelope geometry (CPVC / Polypropylene).
	sumpVol := math.Max(qRecirc*packedBedSumpResTimeSec, packedBedMinSumpVolM3)
	sumpHeight := math.Max(0.45, math.Min(0.85, sumpVol/colArea))
	// Vertical stack: sump + inlet plenum (0.30 m) + packing + spray header (0.30 m) + demister + outlet plenum (0.25 m).
	colHeight := math.Ceil((sumpHeight+0.30+packingHeight+0.30+packedBedDemisterHeightM+0.25)/0.05) * 0.05

	wallThickness := packedBedWallThinM
	if colDiam > 0.60 {
		wallThickness = packedBedWallThickM
	}

	packingSpec := input.PackingName
	if packingSpec == "" {
		packingSpec = fmt.Sprintf(`%.2f" Random Packing (eps=%.2f)`, input.PackingDiameter.Inches(), input.PackingVoidage)
	}

	recircFlow := model.CubicMetersPerSec(qRecirc)
	res := model.PackedBedResult{
		ColumnDiameter:           model.Meters(colDiam),
		ColumnHeight:             model.Meters(colHeight),
		WallThickness:            model.Meters(wallThickness),
		FloodingVelocity:         model.MetersPerSec(vFlood),
		SuperficialGasVelocity:   model.MetersPerSec(vSup),
		PercentOfFlood:           pctFlood,
		PackingHeight:            model.Meters(packingHeight),
		NTU:                      ntuActual,
		HTU:                      model.Meters(htu),
		PackingSpec:              packingSpec,
		PackingDiameter:          input.PackingDiameter,
		PackingVoidage:           input.PackingVoidage,
		PackingVolume:            model.CubicMeters(colArea * packingHeight),
		HClInletPPM:              hclInletPPM,
		HClOutletPPM:             hclOutletPPM,
		HClRemovalEfficiency:     hclEff,
		SO2InletPPM:              so2InletPPM,
		SO2OutletPPM:             so2OutletPPM,
		SO2RemovalEfficiency:     so2Eff,
		SulfurAcidDewPointF:      sulfurAcidDewF,
		NaHCO3ConsumptionKgPerHr: nahco3KgPerHr,
		RecirculationFlow:        recircFlow,
		RecirculationFlowGPM:     recircFlow.GPM(),
		SumpPH:                   input.SumpPHTarget,
		SumpVolume:               model.CubicMeters(sumpVol),
		BedPressureDrop:          bedDP,
		DemisterPressureDrop:     demisterDP,
		DemisterHeight:           model.Meters(packedBedDemisterHeightM),
		TotalPressureDrop:        totalDP,
		TotalPressureDropInWC:    totalDP.InWC(),
		Material:                 packedBedMaterialSpec,
	}

	// Build outlet gas stream after HCl & SO2 absorption (re-normalized mole fractions).
	nOut := math.Max(quenchedGas.MolarFlowMolPerSec-nHClRemoved-nSO2Removed, 1e-12)
	nN2 := quenchedGas.MolarFlowMolPerSec * quenchedGas.N2
	nCO2 := quenchedGas.MolarFlowMolPerSec * quenchedGas.CO2
	nH2O := quenchedGas.MolarFlowMolPerSec * quenchedGas.H2O
	nO2 := quenchedGas.MolarFlowMolPerSec * quenchedGas.O2
	nHClOut := math.Max(0.0, nHClIn-nHClRemoved)
	nSO2Out := math.Max(0.0, nSO2In-nSO2Removed)

	outletGas := model.GasComposition{
		N2:                 nN2 / nOut,
		CO2:                nCO2 / nOut,
		H2O:                nH2O / nOut,
		O2:                 nO2 / nOut,
		HCl:                nHClOut / nOut,
		SO2:                nSO2Out / nOut,
		MolarFlowMolPerSec: nOut,
		Temperature:        quenchedGas.Temperature,
		Pressure:           quenchedGas.Pressure,
	}
	if err := thermo.PopulateGasProperties(&outletGas); err != nil {
		return model.GasComposition{}, model.PackedBedResult{}, fmt.Errorf("outlet gas properties: %w", err)
	}

	return outletGas, res, nil
}
