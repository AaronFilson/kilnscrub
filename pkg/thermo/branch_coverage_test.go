package thermo

import (
	"math"
	"testing"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

func TestThermoBranchCoverageAndEdgeCases(t *testing.T) {
	// 1. LookupSpeciesConstants, selectShomateRange, SpeciesViscosity error & out-of-range branches.
	if _, err := LookupSpeciesConstants("Xe"); err == nil {
		t.Error("expected unknown species error")
	}
	if _, err := SpeciesViscosity("Xe", 400); err == nil {
		t.Error("expected unknown species viscosity error")
	}
	if _, _, err := SpeciesCpAndEnthalpy("Xe", 400); err == nil {
		t.Error("expected unknown species Shomate error")
	}
	if _, _, err := SpeciesCpAndEnthalpy(SpeciesN2, 0); err == nil {
		t.Error("expected T<=0 Shomate error")
	}
	if _, _, err := SpeciesCpAndEnthalpy(SpeciesN2, model.Kelvin(math.NaN())); err == nil {
		t.Error("expected NaN Shomate error")
	}
	// Below Tmin and above Tmax fallbacks in selectShomateRange.
	if _, _, err := SpeciesCpAndEnthalpy(SpeciesCO2, 200.0); err != nil {
		t.Errorf("below-Tmin Shomate fallback failed: %v", err)
	}
	if _, _, err := SpeciesCpAndEnthalpy(SpeciesN2, 6500.0); err != nil {
		t.Errorf("above-Tmax Shomate fallback failed: %v", err)
	}

	// 2. LucasGasViscosity: quantum gases (He, H2, D2), polarity tiers, invalid inputs.
	quantum := []SpeciesConstants{
		{Name: "He", CASRN: "7440-59-7", Tc: 5.195, Pc: 227460, Zc: 0.302, MW: 4.0026, Dipole: 0},
		{Name: "H2", CASRN: "1333-74-0", Tc: 33.145, Pc: 1296400, Zc: 0.305, MW: 2.01588, Dipole: 0},
		{Name: "D2", CASRN: "7782-39-0", Tc: 38.34, Pc: 1665000, Zc: 0.306, MW: 4.0282, Dipole: 0},
		// Moderate polarity mur in [0.022, 0.075).
		{Name: "ModPolar", CASRN: "0000-00-0", Tc: 300.0, Pc: 5000000, Zc: 0.27, MW: 40.0, Dipole: 0.45},
	}
	for _, sp := range quantum {
		// Use T = 2 * Tc so Tr = 2 < 12 triggers quantum F_Q^0 path.
		mu, err := LucasGasViscosity(sp.Tc*2.0, sp)
		if err != nil || mu <= 0 {
			t.Errorf("LucasGasViscosity(%s) = (%v, %v)", sp.Name, mu, err)
		}
	}
	if _, err := LucasGasViscosity(0, quantum[0]); err == nil {
		t.Error("expected T<=0 Lucas error")
	}
	if _, err := LucasGasViscosity(300, SpeciesConstants{Name: "Bad", Tc: 0, Pc: 100, MW: 10}); err == nil {
		t.Error("expected invalid critical constants error")
	}

	// 3. WaterPsat & WaterDewPoint edge cases.
	for _, badT := range []model.Kelvin{200, 700, model.Kelvin(math.NaN())} {
		if _, err := WaterPsat(badT); err == nil {
			t.Errorf("WaterPsat(%v) expected error", badT)
		}
	}
	// Taper windows (290..300 K and 373.15..385 K) and outside taper (>385 K).
	for _, validT := range []model.Kelvin{280.0, 295.0, 378.0, 450.0} {
		if ps, err := WaterPsat(validT); err != nil || ps <= 0 {
			t.Errorf("WaterPsat(%v) = (%v, %v)", validT, ps, err)
		}
	}
	if _, err := WaterDewPoint(0); err == nil {
		t.Error("WaterDewPoint(0) expected error")
	}
	if _, err := WaterDewPoint(waterPcPa + 1); err == nil {
		t.Error("WaterDewPoint(>Pc) expected error")
	}
	if dp, err := WaterDewPoint(100.0); err != nil || float64(dp) != waterTtK {
		t.Errorf("WaterDewPoint(100 Pa <= pMin) = (%v, %v)", dp, err)
	}

	// 4. Mixture properties & PopulateGasProperties error branches.
	dryGas := model.GasComposition{N2: 0.8, O2: 0.2, MolarFlowMolPerSec: 1.0, Temperature: 400, Pressure: 101325}
	if _, err := MixtureCp(dryGas, 0); err == nil {
		t.Error("MixtureCp(T=0) expected error")
	}
	if _, err := MixtureEnthalpy(dryGas, 0); err == nil {
		t.Error("MixtureEnthalpy(T=0) expected error")
	}
	if _, err := MixtureViscosity(dryGas, 0); err == nil {
		t.Error("MixtureViscosity(T=0) expected error")
	}
	if err := PopulateGasProperties(nil); err == nil {
		t.Error("PopulateGasProperties(nil) expected error")
	}
	badGas := dryGas
	badGas.Temperature = 0
	if err := PopulateGasProperties(&badGas); err == nil {
		t.Error("PopulateGasProperties(T=0) expected error")
	}
	badGas = dryGas
	badGas.Pressure = 0
	if err := PopulateGasProperties(&badGas); err == nil {
		t.Error("PopulateGasProperties(P=0) expected error")
	}
	badGas = dryGas
	badGas.MolarFlowMolPerSec = -1
	if err := PopulateGasProperties(&badGas); err == nil {
		t.Error("PopulateGasProperties(n<0) expected error")
	}

	// 5. BrentRoot edge cases: NaN bounds, default tol/maxIter, endpoint zeros, NaN endpoint, NaN mid-iter, bisection fallback, non-convergence.
	if _, err := BrentRoot(func(x float64) float64 { return x }, math.NaN(), 1, 1e-6, 10); err == nil {
		t.Error("BrentRoot(NaN) expected error")
	}
	if r, err := BrentRoot(func(x float64) float64 { return x - 1 }, 1, 2, 0, 0); err != nil || r != 1 {
		t.Errorf("BrentRoot(fa==0) = (%v, %v)", r, err)
	}
	if r, err := BrentRoot(func(x float64) float64 { return x - 2 }, 1, 2, 0, 0); err != nil || r != 2 {
		t.Errorf("BrentRoot(fb==0) = (%v, %v)", r, err)
	}
	if _, err := BrentRoot(func(x float64) float64 { return math.NaN() }, 0, 1, 1e-6, 10); err == nil {
		t.Error("BrentRoot(NaN endpoint) expected error")
	}
	if _, err := BrentRoot(func(x float64) float64 { return x*x + 1 }, 0, 1, 1e-6, 10); err == nil {
		t.Error("BrentRoot(unbracketed) expected error")
	}
	calls := 0
	if _, err := BrentRoot(func(x float64) float64 {
		calls++
		if calls > 2 {
			return math.NaN()
		}
		return x - 0.5
	}, 0, 1, 1e-6, 20); err == nil {
		t.Error("BrentRoot(mid-iter NaN) expected error")
	}
	// Step function forces repeated bisection fallback & non-convergence with tiny maxIter.
	stepFn := func(x float64) float64 {
		if x < 0.37 {
			return -1.0
		}
		return 1.0
	}
	if _, err := BrentRoot(stepFn, 0, 1, 1e-12, 3); err == nil {
		t.Error("BrentRoot(maxIter=3 on step) expected non-convergence error")
	}
	if r, err := BrentRoot(stepFn, 0, 1, 1e-6, 100); err != nil || math.Abs(r-0.37) > 1e-5 {
		t.Errorf("BrentRoot(step bisection) = (%v, %v)", r, err)
	}

	// 6. SolveAdiabaticQuench error branches + large tower wall thickness branch.
	if _, _, err := SolveAdiabaticQuench(model.GasComposition{MolarFlowMolPerSec: 0}, 293.15, 101325); err == nil {
		t.Error("SolveAdiabaticQuench(flow=0) expected error")
	}
	if _, _, err := SolveAdiabaticQuench(model.GasComposition{MolarFlowMolPerSec: 1, Temperature: 290}, 293.15, 101325); err == nil {
		t.Error("SolveAdiabaticQuench(T<=Tspray) expected error")
	}
	if _, _, err := SolveAdiabaticQuench(model.GasComposition{MolarFlowMolPerSec: 1, Temperature: 800}, 293.15, 0); err == nil {
		t.Error("SolveAdiabaticQuench(P<=0) expected error")
	}
	if _, _, err := SolveAdiabaticQuench(model.GasComposition{H2O: 1.0, MolarFlowMolPerSec: 1, Temperature: 800, Pressure: 101325}, 293.15, 101325); err == nil {
		t.Error("SolveAdiabaticQuench(nDry=0) expected error")
	}
	// Large molar flow triggering towerDiam > 0.60 thick wall & nozzleCount > 2.
	bigGas := model.GasComposition{N2: 0.72, CO2: 0.13, H2O: 0.12, O2: 0.03, MolarFlowMolPerSec: 15.0, Temperature: 1500, Pressure: 101325}
	if _, qr, err := SolveAdiabaticQuench(bigGas, 293.15, 101325); err != nil || qr.WallThickness != quenchWallThickM {
		t.Errorf("SolveAdiabaticQuench(big) = (wall=%v, err=%v)", qr.WallThickness, err)
	}
}
