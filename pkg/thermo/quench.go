package thermo

import (
	"fmt"
	"math"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

// BrentRoot finds a root of f(x) = 0 inside the bracket [a, b] using Brent's method
// (inverse quadratic interpolation, secant, and bisection fallback).
// Note: gonum v0.17.0 does not include a 1D Brent root finder in optimize.
func BrentRoot(f func(float64) float64, a, b, tol float64, maxIter int) (float64, error) {
	if math.IsNaN(a) || math.IsNaN(b) || math.IsInf(a, 0) || math.IsInf(b, 0) {
		return 0, fmt.Errorf("brent bracket bounds must be finite: [%v, %v]", a, b)
	}
	if tol <= 0 {
		tol = 1e-12
	}
	if maxIter <= 0 {
		maxIter = 200
	}

	fa := f(a)
	fb := f(b)
	if math.IsNaN(fa) || math.IsNaN(fb) {
		return 0, fmt.Errorf("brent objective returned NaN at bracket endpoints [%v, %v]", a, b)
	}
	if fa == 0 {
		return a, nil
	}
	if fb == 0 {
		return b, nil
	}
	if (fa > 0 && fb > 0) || (fa < 0 && fb < 0) {
		return 0, fmt.Errorf("brent root not bracketed on [%g, %g]: f(a)=%g, f(b)=%g", a, b, fa, fb)
	}

	c, fc := a, fa
	d := b - a
	e := d
	const eps = 2.220446049250313e-16

	for iter := 0; iter < maxIter; iter++ {
		if (fb > 0 && fc > 0) || (fb < 0 && fc < 0) {
			c, fc = a, fa
			d = b - a
			e = d
		}
		if math.Abs(fc) < math.Abs(fb) {
			a, b, c = b, c, b
			fa, fb, fc = fb, fc, fb
		}

		tol1 := 2.0*eps*math.Abs(b) + 0.5*tol
		xm := 0.5 * (c - b)
		if math.Abs(xm) <= tol1 || fb == 0 {
			return b, nil
		}

		if math.Abs(e) >= tol1 && math.Abs(fa) > math.Abs(fb) {
			s := fb / fa
			var p, q float64
			if a == c {
				p = 2.0 * xm * s
				q = 1.0 - s
			} else {
				q = fa / fc
				r := fb / fc
				p = s * (2.0*xm*q*(q-r) - (b-a)*(r-1.0))
				q = (q - 1.0) * (r - 1.0) * (s - 1.0)
			}
			if p > 0 {
				q = -q
			}
			p = math.Abs(p)
			min1 := 3.0*xm*q - math.Abs(tol1*q)
			min2 := math.Abs(e * q)
			if 2.0*p < math.Min(min1, min2) {
				e = d
				d = p / q
			} else {
				d = xm
				e = d
			}
		} else {
			d = xm
			e = d
		}

		a, fa = b, fb
		if math.Abs(d) > tol1 {
			b += d
		} else {
			b += math.Copysign(tol1, xm)
		}
		fb = f(b)
		if math.IsNaN(fb) {
			return 0, fmt.Errorf("brent objective returned NaN at x=%g", b)
		}
	}

	return b, fmt.Errorf("brent failed to converge within %d iterations (last b=%g, f(b)=%g)", maxIter, b, fb)
}

// Quench sizing engineering constants for a 316L stainless steel adiabatic spray quench tower.
const (
	defaultOversprayFactor   = 2.0   // 2x stoichiometric spray water for reliable saturation & wet walls
	targetQuenchVelocityMS   = 2.0   // Target superficial gas velocity in open spray quench tower [m/s]
	targetQuenchResidenceSec = 1.25  // Minimum vapor-droplet contact residence time [s]
	minQuenchTowerDiameterM  = 0.25  // Minimum practical studio fabrication diameter [m] (~10 in)
	minQuenchTowerHeightM    = 1.20  // Minimum height for spray cone development [m] (~4 ft)
	quenchLossCoeffK         = 2.2   // Combined inlet/outlet transition & spray-curtain velocity head loss
	quenchNozzleCapacityGPM  = 0.75  // Nominal flow per 316L full-cone atomizing nozzle [GPM]
	minQuenchNozzleCount     = 2     // Minimum redundant atomizing nozzles
	quenchWallThinM          = 0.002 // 14-gauge 316L SS (~2.0 mm)
	quenchWallThickM         = 0.003 // 11-gauge 316L SS (~3.0 mm) for larger diameters
	quenchMaterialSpec       = "316L Stainless Steel"
)

// SolveAdiabaticQuench solves the adiabatic saturation energy balance
//
//	H_gas(T_in) + m_w * h_liquid(T_spray) = H_gas_sat(T_sat)
//
// using Brent's method, computes evaporated and total spray water rates, stack dew point,
// quenched gas state, and sizes the Stage 1 316L SS quench tower.
func SolveAdiabaticQuench(
	inletGas model.GasComposition,
	sprayWaterTemp model.Kelvin,
	ambientPressure model.Pascal,
) (model.GasComposition, model.QuenchResult, error) {
	if inletGas.MolarFlowMolPerSec <= 0 {
		return model.GasComposition{}, model.QuenchResult{}, fmt.Errorf("inlet gas molar flow must be > 0, got %v", inletGas.MolarFlowMolPerSec)
	}
	if inletGas.Temperature <= sprayWaterTemp {
		return model.GasComposition{}, model.QuenchResult{}, fmt.Errorf("inlet gas temperature (%.2f K) must exceed spray water temperature (%.2f K)", inletGas.Temperature, sprayWaterTemp)
	}
	if ambientPressure <= 0 {
		return model.GasComposition{}, model.QuenchResult{}, fmt.Errorf("ambient pressure must be > 0 Pa, got %v", ambientPressure)
	}

	// Ensure inlet gas derived properties are populated.
	if err := PopulateGasProperties(&inletGas); err != nil {
		return model.GasComposition{}, model.QuenchResult{}, fmt.Errorf("inlet gas properties: %w", err)
	}

	nIn := inletGas.MolarFlowMolPerSec
	nN2 := nIn * inletGas.N2
	nCO2 := nIn * inletGas.CO2
	nO2 := nIn * inletGas.O2
	nHCl := nIn * inletGas.HCl
	nWIn := nIn * inletGas.H2O
	nDry := nN2 + nCO2 + nO2 + nHCl
	if nDry <= 0 {
		return model.GasComposition{}, model.QuenchResult{}, fmt.Errorf("inlet gas dry molar flow must be > 0")
	}

	hInTotal := nIn * inletGas.EnthalpyJPerMol
	hLiq := LiquidWaterEnthalpyJPerMol(sprayWaterTemp)
	pTotal := float64(ambientPressure)

	// Determine lower bound (inlet dew point, clamped >= 273.2 K) and upper bound (just below boiling at P).
	tLow := 273.25
	if inletGas.H2O > 0 {
		pH2OIn := model.Pascal(inletGas.H2O * pTotal)
		if dp, err := WaterDewPoint(pH2OIn); err == nil && float64(dp) > tLow {
			tLow = float64(dp)
		}
	}
	tBoil, err := WaterDewPoint(ambientPressure)
	if err != nil {
		return model.GasComposition{}, model.QuenchResult{}, fmt.Errorf("boiling temperature lookup: %w", err)
	}
	// Keep upper bound slightly below T_boil so 1 - Psat/P > 0.
	tHigh := float64(tBoil) - 0.05
	if tLow >= tHigh {
		return model.GasComposition{}, model.QuenchResult{}, fmt.Errorf("inlet gas dew point (%.2f K) >= water boiling point (%.2f K)", tLow, tBoil)
	}

	residual := func(tSat float64) float64 {
		psat, e := WaterPsat(model.Kelvin(tSat))
		if e != nil {
			return math.NaN()
		}
		ywSat := float64(psat) / pTotal
		if ywSat >= 0.9999 {
			return -1e12
		}
		nWSat := nDry * ywSat / (1.0 - ywSat)
		nEvap := nWSat - nWIn

		tK := model.Kelvin(tSat)
		hN2, e1 := SpeciesEnthalpy(SpeciesN2, tK)
		hCO2, e2 := SpeciesEnthalpy(SpeciesCO2, tK)
		hH2O, e3 := SpeciesEnthalpy(SpeciesH2O, tK)
		hO2, e4 := SpeciesEnthalpy(SpeciesO2, tK)
		hHCl, e5 := SpeciesEnthalpy(SpeciesHCl, tK)
		if e1 != nil || e2 != nil || e3 != nil || e4 != nil || e5 != nil {
			return math.NaN()
		}
		hSatTotal := nN2*hN2 + nCO2*hCO2 + nO2*hO2 + nHCl*hHCl + nWSat*hH2O
		return (hInTotal + nEvap*hLiq) - hSatTotal
	}

	tSatVal, err := BrentRoot(residual, tLow, tHigh, 1e-8, 120)
	if err != nil {
		return model.GasComposition{}, model.QuenchResult{}, fmt.Errorf("adiabatic saturation solver failed: %w", err)
	}

	tSat := model.Kelvin(tSatVal)
	psat, err := WaterPsat(tSat)
	if err != nil {
		return model.GasComposition{}, model.QuenchResult{}, err
	}
	ywSat := float64(psat) / pTotal
	nSat := nDry / (1.0 - ywSat)
	nWSat := nSat * ywSat
	nEvap := math.Max(0.0, nWSat-nWIn)

	quenchedGas := model.GasComposition{
		N2:                 nN2 / nSat,
		CO2:                nCO2 / nSat,
		H2O:                nWSat / nSat,
		O2:                 nO2 / nSat,
		HCl:                nHCl / nSat,
		MolarFlowMolPerSec: nSat,
		Temperature:        tSat,
		Pressure:           ambientPressure,
	}
	if err := PopulateGasProperties(&quenchedGas); err != nil {
		return model.GasComposition{}, model.QuenchResult{}, fmt.Errorf("quenched gas properties: %w", err)
	}

	dewPoint, err := WaterDewPoint(model.Pascal(quenchedGas.H2O * pTotal))
	if err != nil {
		return model.GasComposition{}, model.QuenchResult{}, fmt.Errorf("stack dew point: %w", err)
	}

	mwH2O := speciesConstantsTable[SpeciesH2O].MWKgPerMol()
	evapKgPerSec := model.KgPerSec(nEvap * mwH2O)
	evapVolM3S := model.CubicMetersPerSec(float64(evapKgPerSec) / float64(model.WaterDensity))
	evapGPM := evapVolM3S.GPM()

	overspray := defaultOversprayFactor
	totalSprayM3S := model.CubicMetersPerSec(float64(evapVolM3S) * overspray)
	totalSprayGPM := totalSprayM3S.GPM()

	// Size 316L SS quench tower using geometric mean volumetric flow between hot inlet and saturated exit.
	qMean := math.Sqrt(float64(inletGas.VolumetricFlow) * float64(quenchedGas.VolumetricFlow))
	areaReq := qMean / targetQuenchVelocityMS
	rawDiam := math.Sqrt(4.0 * areaReq / math.Pi)
	// Round up to nearest 25 mm (~1 inch) for clean sheet-metal rolling.
	towerDiam := math.Ceil(rawDiam/0.025) * 0.025
	if towerDiam < minQuenchTowerDiameterM {
		towerDiam = minQuenchTowerDiameterM
	}
	towerArea := math.Pi * towerDiam * towerDiam / 4.0
	gasVel := float64(quenchedGas.VolumetricFlow) / towerArea
	meanGasVel := qMean / towerArea

	// Tower height provides >= targetQuenchResidenceSec at mean gas velocity and >= 3.5*D for full spray evaporation.
	rawHeight := math.Max(meanGasVel*targetQuenchResidenceSec, 3.5*towerDiam)
	rawHeight = math.Max(rawHeight, minQuenchTowerHeightM)
	towerHeight := math.Ceil(rawHeight/0.05) * 0.05
	residenceTime := towerHeight / math.Max(meanGasVel, 1e-6)

	wallThickness := quenchWallThinM
	if towerDiam > 0.60 {
		wallThickness = quenchWallThickM
	}

	nozzleCount := int(math.Ceil(totalSprayGPM / quenchNozzleCapacityGPM))
	if nozzleCount < minQuenchNozzleCount {
		nozzleCount = minQuenchNozzleCount
	}

	// Quench tower pressure drop: inlet/outlet dynamic losses + spray curtain drag.
	rhoMean := 0.5 * (float64(inletGas.Density) + float64(quenchedGas.Density))
	sprayDragPa := 60.0 + 35.0*totalSprayGPM
	quenchDP := model.Pascal(quenchLossCoeffK*0.5*rhoMean*meanGasVel*meanGasVel + sprayDragPa)

	res := model.QuenchResult{
		InletTemp:          inletGas.Temperature,
		AdiabaticSatTemp:   tSat,
		OutletTemp:         tSat,
		DewPoint:           dewPoint,
		EvaporatedWater:    evapKgPerSec,
		EvaporatedWaterGPM: evapGPM,
		TotalSprayWater:    totalSprayM3S,
		TotalSprayWaterGPM: totalSprayGPM,
		OversprayFactor:    overspray,
		TowerDiameter:      model.Meters(towerDiam),
		TowerHeight:        model.Meters(towerHeight),
		WallThickness:      model.Meters(wallThickness),
		GasVelocity:        model.MetersPerSec(gasVel),
		ResidenceTimeSec:   residenceTime,
		NozzleCount:        nozzleCount,
		PressureDrop:       quenchDP,
		PressureDropInWC:   quenchDP.InWC(),
		Material:           quenchMaterialSpec,
	}
	return quenchedGas, res, nil
}
