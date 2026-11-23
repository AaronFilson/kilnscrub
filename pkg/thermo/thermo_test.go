package thermo

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

type oracleFixtures struct {
	SpeciesConstants map[string]struct {
		CASRN   string  `json:"CASRN"`
		TcK     float64 `json:"Tc_K"`
		PcPa    float64 `json:"Pc_Pa"`
		Zc      float64 `json:"Zc"`
		MWGMol  float64 `json:"MW_g_mol"`
		DipoleD float64 `json:"dipole_D"`
	} `json:"species_constants"`
	GasProperties []struct {
		TK      float64 `json:"T_K"`
		Species map[string]struct {
			CpJMolK        float64 `json:"Cp_J_mol_K"`
			HMinusH298JMol float64 `json:"H_minus_H298_J_mol"`
			MuPaS          float64 `json:"mu_Pa_s"`
		} `json:"species"`
	} `json:"gas_properties"`
	WaterVaporPressure []struct {
		TK     float64 `json:"T_K"`
		PsatPa float64 `json:"Psat_Pa"`
	} `json:"water_vapor_pressure"`
}

func loadOracleFixtures(t *testing.T) oracleFixtures {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "oracle_fixtures.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read oracle fixtures %s: %v", path, err)
	}
	var fx oracleFixtures
	if err := json.Unmarshal(data, &fx); err != nil {
		t.Fatalf("unmarshal oracle fixtures: %v", err)
	}
	return fx
}

func relErr(got, want float64) float64 {
	if want == 0 {
		return math.Abs(got - want)
	}
	return math.Abs(got-want) / math.Abs(want)
}

func TestShomateAndViscosityAgainstOracle(t *testing.T) {
	fx := loadOracleFixtures(t)
	const maxRelTol = 1e-6

	// Verify embedded species constants match oracle fixture constants.
	for sp, oc := range fx.SpeciesConstants {
		c, err := LookupSpeciesConstants(sp)
		if err != nil {
			t.Fatalf("LookupSpeciesConstants(%s): %v", sp, err)
		}
		if c.CASRN != oc.CASRN {
			t.Errorf("%s CASRN = %s, want %s", sp, c.CASRN, oc.CASRN)
		}
		if relErr(float64(c.Tc), oc.TcK) > 1e-12 || relErr(float64(c.Pc), oc.PcPa) > 1e-12 {
			t.Errorf("%s critical constants mismatch: got (%v, %v), want (%v, %v)", sp, c.Tc, c.Pc, oc.TcK, oc.PcPa)
		}
	}

	// Table-driven oracle verification for Cp(T), H(T)-H(298.15), and Lucas gas viscosity mu(T).
	for _, entry := range fx.GasProperties {
		T := model.Kelvin(entry.TK)
		for sp, want := range entry.Species {
			t.Run(sp+"_"+formatTemp(entry.TK), func(t *testing.T) {
				cp, h, err := SpeciesCpAndEnthalpy(sp, T)
				if err != nil {
					t.Fatalf("SpeciesCpAndEnthalpy(%s, %.2f K): %v", sp, entry.TK, err)
				}
				if r := relErr(cp, want.CpJMolK); r > maxRelTol {
					t.Errorf("Cp(%s, %.2f K) = %.12g, want %.12g (relErr=%.3e > %.1e)", sp, entry.TK, cp, want.CpJMolK, r, maxRelTol)
				}
				if r := relErr(h, want.HMinusH298JMol); r > maxRelTol {
					t.Errorf("H(%s, %.2f K) = %.12g, want %.12g (relErr=%.3e > %.1e)", sp, entry.TK, h, want.HMinusH298JMol, r, maxRelTol)
				}

				mu, err := SpeciesViscosity(sp, T)
				if err != nil {
					t.Fatalf("SpeciesViscosity(%s, %.2f K): %v", sp, entry.TK, err)
				}
				if r := relErr(float64(mu), want.MuPaS); r > maxRelTol {
					t.Errorf("mu(%s, %.2f K) = %.12g, want %.12g (relErr=%.3e > %.1e)", sp, entry.TK, mu, want.MuPaS, r, maxRelTol)
				}
			})
		}
	}
}

func TestN2Shomate500KLowerRangeWins(t *testing.T) {
	// Critical boundary check: at T = 500.0 K, lower-range-wins (100, 500) yields
	// H = 5910.7544479... J/mol rather than 5911.523... J/mol from (500, 2000).
	h, err := SpeciesEnthalpy(SpeciesN2, 500.0)
	if err != nil {
		t.Fatalf("SpeciesEnthalpy(N2, 500 K): %v", err)
	}
	const wantLowerRange = 5910.754447916666
	if r := relErr(h, wantLowerRange); r > 1e-9 {
		t.Fatalf("N2 H(500 K) = %.9f J/mol, want %.9f J/mol (relErr=%.3e)", h, wantLowerRange, r)
	}
}

func TestWaterPsatAgainstOracle(t *testing.T) {
	fx := loadOracleFixtures(t)
	const maxRelTol = 1e-6

	var prev float64
	for _, entry := range fx.WaterVaporPressure {
		T := model.Kelvin(entry.TK)
		psat, err := WaterPsat(T)
		if err != nil {
			t.Fatalf("WaterPsat(%.2f K): %v", entry.TK, err)
		}
		if r := relErr(float64(psat), entry.PsatPa); r > maxRelTol {
			t.Errorf("WaterPsat(%.2f K) = %.12g Pa, want %.12g Pa (relErr=%.3e > %.1e)", entry.TK, psat, entry.PsatPa, r, maxRelTol)
		}
		if float64(psat) <= prev {
			t.Errorf("WaterPsat not strictly increasing at %.2f K: %v <= %v", entry.TK, psat, prev)
		}
		prev = float64(psat)

		// Verify WaterDewPoint round-trip inversion.
		dp, err := WaterDewPoint(psat)
		if err != nil {
			t.Fatalf("WaterDewPoint(%.2f Pa): %v", psat, err)
		}
		if math.Abs(float64(dp)-entry.TK) > 1e-6 {
			t.Errorf("WaterDewPoint(Psat(%.2f K)) = %.8f K, want %.2f K", entry.TK, dp, entry.TK)
		}
	}
}

func TestBrentRoot(t *testing.T) {
	// Solve cos(x) - x = 0 on [0, 1] (Dottie number ~ 0.7390851332151607).
	root, err := BrentRoot(func(x float64) float64 {
		return math.Cos(x) - x
	}, 0.0, 1.0, 1e-13, 100)
	if err != nil {
		t.Fatalf("BrentRoot: %v", err)
	}
	const want = 0.7390851332151607
	if math.Abs(root-want) > 1e-12 {
		t.Errorf("root = %.16f, want %.16f", root, want)
	}

	// Unbracketed root must return an error.
	if _, err := BrentRoot(func(x float64) float64 { return x*x + 1.0 }, -1, 1, 1e-9, 50); err == nil {
		t.Error("expected error for unbracketed root, got nil")
	}
}

func TestSolveAdiabaticQuench(t *testing.T) {
	inlet := model.GasComposition{
		N2:                 0.688,
		CO2:                0.137,
		H2O:                0.131,
		O2:                 0.042,
		HCl:                0.002,
		MolarFlowMolPerSec: 2.1,
		Temperature:        model.FahrenheitToKelvin(2400.0), // 1588.15 K
		Pressure:           model.StandardPressurePa,
	}
	sprayT := model.CelsiusToKelvin(20.0)
	quenched, res, err := SolveAdiabaticQuench(inlet, sprayT, model.StandardPressurePa)
	if err != nil {
		t.Fatalf("SolveAdiabaticQuench: %v", err)
	}

	// Adiabatic saturation for 2400 F kiln flue gas should fall in ~160-190 F (344-361 K),
	// safely below the 210 F (372.04 K) CPVC protection threshold.
	if res.AdiabaticSatTemp < 340.0 || res.AdiabaticSatTemp > 365.0 {
		t.Errorf("AdiabaticSatTemp = %.2f K (%.1f F), expected in [340, 365] K", res.AdiabaticSatTemp, res.AdiabaticSatTemp.Fahrenheit())
	}
	if res.OutletTemp >= model.DefaultSafetyLimits().MaxQuenchOutletTemp {
		t.Errorf("OutletTemp %.2f K >= CPVC safety limit %.2f K", res.OutletTemp, model.DefaultSafetyLimits().MaxQuenchOutletTemp)
	}
	if math.Abs(float64(res.DewPoint-res.AdiabaticSatTemp)) > 0.05 {
		t.Errorf("DewPoint (%.3f K) should match AdiabaticSatTemp (%.3f K)", res.DewPoint, res.AdiabaticSatTemp)
	}
	if res.EvaporatedWaterGPM <= 0 || res.TotalSprayWaterGPM <= res.EvaporatedWaterGPM {
		t.Errorf("invalid water rates: evap=%.4f GPM, total=%.4f GPM", res.EvaporatedWaterGPM, res.TotalSprayWaterGPM)
	}
	if math.Abs(quenched.SumMoleFractions()-1.0) > 1e-9 {
		t.Errorf("quenched mole fraction sum = %.12f, want 1.0", quenched.SumMoleFractions())
	}
	if res.Material != "316L Stainless Steel" {
		t.Errorf("Material = %q, want 316L Stainless Steel", res.Material)
	}

	// Independently verify energy balance closure: H_in + n_evap * h_liq == H_out.
	if err := PopulateGasProperties(&inlet); err != nil {
		t.Fatalf("PopulateGasProperties(inlet): %v", err)
	}
	nEvap := quenched.MolarFlowMolPerSec*quenched.H2O - inlet.MolarFlowMolPerSec*inlet.H2O
	lhs := inlet.MolarFlowMolPerSec*inlet.EnthalpyJPerMol + nEvap*LiquidWaterEnthalpyJPerMol(sprayT)
	rhs := quenched.MolarFlowMolPerSec * quenched.EnthalpyJPerMol
	if r := math.Abs(lhs-rhs) / math.Max(math.Abs(lhs), 1.0); r > 1e-6 {
		t.Errorf("energy balance residual %.3e > 1e-6 (LHS=%.2f W, RHS=%.2f W)", r, lhs, rhs)
	}
}

func formatTemp(t float64) string {
	return string([]byte{
		byte('0' + int(t)/1000%10),
		byte('0' + int(t)/100%10),
		byte('0' + int(t)/10%10),
		byte('0' + int(t)%10),
		'K',
	})
}
