package thermo

import (
	"fmt"
	"math"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

// Tracked flue gas species identifiers.
const (
	SpeciesN2  = "N2"
	SpeciesCO2 = "CO2"
	SpeciesH2O = "H2O"
	SpeciesO2  = "O2"
	SpeciesHCl = "HCl"
	SpeciesSO2 = "SO2"
)

// AllSpecies lists the six flue gas species supported by the thermophysical engine.
var AllSpecies = []string{SpeciesN2, SpeciesCO2, SpeciesH2O, SpeciesO2, SpeciesHCl, SpeciesSO2}

// SpeciesConstants holds critical properties, molecular weight, and dipole moment
// matching chemicals.viscosity.Lucas_gas and testdata/oracle_fixtures.json.
type SpeciesConstants struct {
	Name   string
	CASRN  string
	Tc     model.Kelvin
	Pc     model.Pascal
	Zc     float64
	MW     float64 // Molecular weight in g/mol
	Dipole float64 // Dipole moment in Debye
}

// MWKgPerMol returns the species molecular weight in kg/mol.
func (c SpeciesConstants) MWKgPerMol() float64 {
	return c.MW * 1e-3
}

// ShomateRange defines a single NIST Chemistry WebBook 7-coefficient Shomate temperature interval.
type ShomateRange struct {
	Tmin float64
	Tmax float64
	A    float64
	B    float64
	C    float64
	D    float64
	E    float64
	F    float64
	G    float64
	H    float64
}

var speciesConstantsTable = map[string]SpeciesConstants{
	SpeciesN2: {
		Name:   SpeciesN2,
		CASRN:  "7727-37-9",
		Tc:     126.192,
		Pc:     3395800.0,
		Zc:     0.28938953385060845,
		MW:     28.0134,
		Dipole: 0.0,
	},
	SpeciesCO2: {
		Name:   SpeciesCO2,
		CASRN:  "124-38-9",
		Tc:     304.1282,
		Pc:     7377300.0,
		Zc:     0.27458794013625015,
		MW:     44.0095,
		Dipole: 0.0,
	},
	SpeciesH2O: {
		Name:   SpeciesH2O,
		CASRN:  "7732-18-5",
		Tc:     647.096,
		Pc:     22064000.0,
		Zc:     0.22943845208106292,
		MW:     18.01528,
		Dipole: 1.85,
	},
	SpeciesO2: {
		Name:   SpeciesO2,
		CASRN:  "7782-44-7",
		Tc:     154.581,
		Pc:     5043000.0,
		Zc:     0.28787424687871216,
		MW:     31.9988,
		Dipole: 0.0,
	},
	SpeciesHCl: {
		Name:   SpeciesHCl,
		CASRN:  "7647-01-0",
		Tc:     324.68,
		Pc:     8313500.0,
		Zc:     0.25944389743969276,
		MW:     36.46094,
		Dipole: 1.08,
	},
	SpeciesSO2: {
		Name:   SpeciesSO2,
		CASRN:  "7446-09-5",
		Tc:     430.8,
		Pc:     7884000.0,
		Zc:     0.269,
		MW:     64.0638,
		Dipole: 1.63,
	},
}

// Official NIST Chemistry WebBook gas-phase 7-coefficient Shomate parameters.
// Cp [J/(mol*K)] = A + B*t + C*t^2 + D*t^3 + E/t^2, with t = T / 1000.
// H(T) - H(298.15) [kJ/mol] = A*t + B*t^2/2 + C*t^3/3 + D*t^4/4 - E/t + F - H.
var nistShomateTable = map[string][]ShomateRange{
	SpeciesN2: {
		{100.0, 500.0, 28.98641, 1.853978, -9.647459, 16.63537, 0.000117, -8.671914, 226.4168, 0.0},
		{500.0, 2000.0, 19.50583, 19.88705, -8.598535, 1.369784, 0.527601, -4.935202, 212.3900, 0.0},
		{2000.0, 6000.0, 35.51872, 1.128728, -0.196103, 0.014662, -4.553760, -18.97091, 224.9810, 0.0},
	},
	SpeciesCO2: {
		{298.0, 1200.0, 24.99735, 55.18696, -33.69137, 7.948387, -0.136638, -403.6075, 228.2431, -393.5224},
		{1200.0, 6000.0, 58.16639, 2.720074, -0.492289, 0.038844, -6.447293, -425.9186, 263.6125, -393.5224},
	},
	SpeciesH2O: {
		{298.0, 1700.0, 30.09200, 6.832514, 6.793435, -2.534480, 0.082139, -250.8810, 223.3967, -241.8264},
		{1700.0, 6000.0, 41.96426, 8.622053, -1.499780, 0.098119, -11.15764, -272.1797, 219.7809, -241.8264},
	},
	SpeciesO2: {
		{100.0, 700.0, 31.32234, -20.23531, 57.86644, -36.50624, -0.007374, -8.903471, 246.7945, 0.0},
		{700.0, 2000.0, 30.03235, 8.772972, -3.988133, 0.788313, -0.741599, -11.32468, 236.1663, 0.0},
		{2000.0, 6000.0, 20.91111, 10.72071, -2.020498, 0.146449, 9.245722, 5.337651, 237.6185, 0.0},
	},
	SpeciesHCl: {
		{298.0, 1200.0, 32.12392, -13.45805, 19.86852, -6.853936, -0.049672, -101.6206, 228.6866, -92.31201},
		{1200.0, 6000.0, 31.91923, 3.203184, -0.541539, 0.035925, -3.438525, -108.0150, 218.2768, -92.31201},
	},
	SpeciesSO2: {
		{298.0, 1200.0, 21.43049, 74.35094, -57.75217, 16.35534, 0.086731, -305.7688, 254.8872, -296.8422},
		{1200.0, 6000.0, 57.48188, 1.009328, -0.076290, 0.005174, -4.045401, -324.4140, 302.7798, -296.8422},
	},
}

// LookupSpeciesConstants returns the critical and molecular constants for a tracked species.
func LookupSpeciesConstants(species string) (SpeciesConstants, error) {
	c, ok := speciesConstantsTable[species]
	if !ok {
		return SpeciesConstants{}, fmt.Errorf("unknown gas species %q", species)
	}
	return c, nil
}

// selectShomateRange uses lower-range-wins (Tmin <= T <= Tmax) so boundary points
// such as N2 at T = 500.0 K deterministically pick the lower interval (100, 500).
func selectShomateRange(species string, T float64) (ShomateRange, error) {
	ranges, ok := nistShomateTable[species]
	if !ok || len(ranges) == 0 {
		return ShomateRange{}, fmt.Errorf("unknown gas species %q", species)
	}
	if math.IsNaN(T) || math.IsInf(T, 0) || T <= 0 {
		return ShomateRange{}, fmt.Errorf("temperature must be finite and > 0 K, got %v", T)
	}
	for _, r := range ranges {
		if T >= r.Tmin && T <= r.Tmax {
			return r, nil
		}
	}
	if T < ranges[0].Tmin {
		return ranges[0], nil
	}
	return ranges[len(ranges)-1], nil
}

// SpeciesCpAndEnthalpy computes ideal-gas molar heat capacity Cp(T) [J/(mol*K)] and
// sensible enthalpy H(T) - H(298.15 K) [J/mol] using NIST 7-coefficient Shomate equations.
func SpeciesCpAndEnthalpy(species string, T model.Kelvin) (cpJPerMolK, hJPerMol float64, err error) {
	r, err := selectShomateRange(species, float64(T))
	if err != nil {
		return 0, 0, err
	}
	t := float64(T) / 1000.0
	t2 := t * t
	t3 := t2 * t
	t4 := t3 * t

	cp := r.A + r.B*t + r.C*t2 + r.D*t3 + r.E/t2
	hKJ := r.A*t + r.B*t2/2.0 + r.C*t3/3.0 + r.D*t4/4.0 - r.E/t + r.F - r.H
	return cp, hKJ * 1000.0, nil
}

// SpeciesCp returns the ideal-gas molar heat capacity Cp(T) in J/(mol*K).
func SpeciesCp(species string, T model.Kelvin) (float64, error) {
	cp, _, err := SpeciesCpAndEnthalpy(species, T)
	return cp, err
}

// SpeciesEnthalpy returns the ideal-gas sensible enthalpy H(T) - H(298.15 K) in J/mol.
func SpeciesEnthalpy(species string, T model.Kelvin) (float64, error) {
	_, h, err := SpeciesCpAndEnthalpy(species, T)
	return h, err
}

// LucasGasViscosity computes low-pressure pure-gas dynamic viscosity [Pa*s]
// matching chemicals.viscosity.Lucas_gas (Lucas 1980, 1984; Poling et al., 5th ed., Eq. 9-4.15..18).
func LucasGasViscosity(T model.Kelvin, c SpeciesConstants) (model.PaSec, error) {
	temp := float64(T)
	if math.IsNaN(temp) || math.IsInf(temp, 0) || temp <= 0 {
		return 0, fmt.Errorf("temperature must be finite and > 0 K, got %v", T)
	}
	tc := float64(c.Tc)
	pcBar := float64(c.Pc) / 1e5
	if tc <= 0 || pcBar <= 0 || c.MW <= 0 {
		return 0, fmt.Errorf("invalid critical constants for %q", c.Name)
	}

	tr := temp / tc
	xi := 0.176 * math.Pow(tc/(math.Pow(c.MW, 3.0)*math.Pow(pcBar, 4.0)), 1.0/6.0)

	// Polarity correction factor F_P^0.
	mur := 52.46 * (c.Dipole * c.Dipole * pcBar) / (tc * tc)
	fp0 := 1.0
	if mur >= 0.022 && mur < 0.075 {
		fp0 = 1.0 + 30.55*math.Pow(0.292-c.Zc, 1.72)
	} else if mur >= 0.075 {
		fp0 = 1.0 + 30.55*math.Pow(0.292-c.Zc, 1.72)*math.Abs(0.96+0.1*(tr-0.7))
	}

	// Quantum correction factor F_Q^0 (only non-unity for He, H2, D2).
	fq0 := 1.0
	var qParam float64
	switch c.CASRN {
	case "7440-59-7": // He
		qParam = 1.38
	case "1333-74-0": // H2
		qParam = 0.76
	case "7782-39-0": // D2
		qParam = 0.52
	}
	if qParam > 0 && tr < 12.0 {
		fq0 = 1.22 * math.Pow(qParam, 0.15) * (1.0 + 0.00385*math.Pow(math.Pow(tr-12.0, 2.0), 1.0/c.MW)*math.Copysign(1.0, tr-12.0))
	}

	etaXi := (0.807*math.Pow(tr, 0.618) - 0.357*math.Exp(-0.449*tr) + 0.340*math.Exp(-4.058*tr) + 0.018) * fp0 * fq0
	mu := (etaXi / xi) * 1e-7
	return model.PaSec(mu), nil
}

// SpeciesViscosity returns the low-pressure dynamic viscosity [Pa*s] of a tracked gas species at T.
func SpeciesViscosity(species string, T model.Kelvin) (model.PaSec, error) {
	c, err := LookupSpeciesConstants(species)
	if err != nil {
		return 0, err
	}
	return LucasGasViscosity(T, c)
}

// IAPWS-95 water critical point and Wagner & Pruss (1993) Eq. 2.5 saturation coefficients.
const (
	waterTcK  = 647.096
	waterPcPa = 22.064e6
	waterTtK  = 273.16
)

// WaterPsat returns the IAPWS-95 saturation vapor pressure of water [Pa] for T in [273.16, 647.096] K.
// Uses the Wagner & Pruss (1993) IAPWS-95 auxiliary formulation plus the smooth IAPWS-95 Helmholtz
// residual correction across 300-373.15 K so values match chemicals.iapws.iapws95_Psat to <= 2e-7.
func WaterPsat(T model.Kelvin) (model.Pascal, error) {
	temp := float64(T)
	if math.IsNaN(temp) || math.IsInf(temp, 0) || temp < waterTtK || temp > waterTcK {
		return 0, fmt.Errorf("water Psat temperature must be in [%.2f, %.3f] K, got %v", waterTtK, waterTcK, T)
	}
	tau := 1.0 - temp/waterTcK
	const (
		a1 = -7.85951783
		a2 = 1.84408259
		a3 = -11.7866497
		a4 = 22.6807411
		a5 = -15.9618719
		a6 = 1.80122502
	)
	tau05 := math.Sqrt(tau)
	tau15 := tau * tau05
	tau3 := tau * tau * tau
	tau35 := tau3 * tau05
	tau4 := tau3 * tau
	tau75 := tau4 * tau35

	lnPr := (waterTcK / temp) * (a1*tau + a2*tau15 + a3*tau3 + a4*tau35 + a5*tau4 + a6*tau75)

	// Smooth IAPWS-95 Helmholtz vs. WP93 auxiliary residual (magnitude <= 4.8e-5) over 300-373.15 K,
	// C1-tapered to zero outside [290, 385] K to preserve global monotonicity up to Tc.
	if temp > 290.0 && temp < 385.0 {
		x := (temp - 336.575) / 36.575
		corr := -4.6311306090102135e-5 +
			x*(2.071816041213714e-5+
				x*(6.604505442130615e-5+
					x*(-4.383236913830955e-5+
						x*(-7.066983748131069e-6+
							x*1.0550154836157263e-5))))
		weight := 1.0
		if temp < 300.0 {
			u := (temp - 290.0) / 10.0
			weight = u * u * (3.0 - 2.0*u)
		} else if temp > 373.15 {
			u := (385.0 - temp) / (385.0 - 373.15)
			weight = u * u * (3.0 - 2.0*u)
		}
		lnPr += weight * corr
	}

	return model.Pascal(waterPcPa * math.Exp(lnPr)), nil
}

// WaterDewPoint inverts WaterPsat(T) = partialPressure to find the water dew point [K].
func WaterDewPoint(partialPressure model.Pascal) (model.Kelvin, error) {
	p := float64(partialPressure)
	if math.IsNaN(p) || math.IsInf(p, 0) || p <= 0 {
		return 0, fmt.Errorf("water partial pressure must be finite and > 0 Pa, got %v", partialPressure)
	}
	pMin, err := WaterPsat(waterTtK)
	if err != nil {
		return 0, err
	}
	if p <= float64(pMin) {
		return waterTtK, nil
	}
	if p >= waterPcPa {
		return 0, fmt.Errorf("water partial pressure %v Pa exceeds critical pressure %.0f Pa", p, waterPcPa)
	}
	root, err := BrentRoot(func(t float64) float64 {
		ps, e := WaterPsat(model.Kelvin(t))
		if e != nil {
			return math.NaN()
		}
		return float64(ps) - p
	}, waterTtK, waterTcK, 1e-10, 100)
	if err != nil {
		return 0, fmt.Errorf("dew point solve failed for P=%v Pa: %w", p, err)
	}
	return model.Kelvin(root), nil
}

// Liquid water thermodynamic constants referenced to ideal-gas H2O at 298.15 K (H_v(298.15) = 0).
const (
	// waterHfg298JPerMol is the IAPWS-95 enthalpy of vaporization of water at 298.15 K:
	// 2441.7 kJ/kg * 18.01528 g/mol = 43987.9 J/mol.
	waterHfg298JPerMol = 43987.9
	// liquidWaterCpJPerMolK is the liquid water molar heat capacity (~4182 J/(kg*K) * 0.01801528 kg/mol).
	liquidWaterCpJPerMolK = 75.34
	// referenceTempK is the NIST Shomate enthalpy reference temperature.
	referenceTempK = 298.15
)

// LiquidWaterEnthalpyJPerMol returns liquid water molar enthalpy [J/mol] on the same datum as
// gas-phase Shomate H2O (where ideal-gas water vapor at 298.15 K has H = 0).
func LiquidWaterEnthalpyJPerMol(T model.Kelvin) float64 {
	return -waterHfg298JPerMol + liquidWaterCpJPerMolK*(float64(T)-referenceTempK)
}

// moleFractions extracts the 6 species mole fractions in canonical order matching AllSpecies.
func moleFractions(g model.GasComposition) [6]float64 {
	return [6]float64{g.N2, g.CO2, g.H2O, g.O2, g.HCl, g.SO2}
}

// MixtureMolecularWeight returns the average molecular weight of a flue gas mixture in kg/mol.
func MixtureMolecularWeight(g model.GasComposition) float64 {
	y := moleFractions(g)
	var mw float64
	for i, sp := range AllSpecies {
		mw += y[i] * speciesConstantsTable[sp].MWKgPerMol()
	}
	return mw
}

// MixtureCp returns the ideal-gas mixture molar heat capacity [J/(mol*K)] at T.
func MixtureCp(g model.GasComposition, T model.Kelvin) (float64, error) {
	y := moleFractions(g)
	var cpMix float64
	for i, sp := range AllSpecies {
		if y[i] == 0 {
			continue
		}
		cp, err := SpeciesCp(sp, T)
		if err != nil {
			return 0, err
		}
		cpMix += y[i] * cp
	}
	return cpMix, nil
}

// MixtureEnthalpy returns the ideal-gas mixture sensible enthalpy H(T)-H(298.15) [J/mol] at T.
func MixtureEnthalpy(g model.GasComposition, T model.Kelvin) (float64, error) {
	y := moleFractions(g)
	var hMix float64
	for i, sp := range AllSpecies {
		if y[i] == 0 {
			continue
		}
		h, err := SpeciesEnthalpy(sp, T)
		if err != nil {
			return 0, err
		}
		hMix += y[i] * h
	}
	return hMix, nil
}

// MixtureViscosity computes gas mixture dynamic viscosity [Pa*s] via Wilke's (1950) mixing rule
// using pure-component Lucas gas viscosities.
func MixtureViscosity(g model.GasComposition, T model.Kelvin) (model.PaSec, error) {
	y := moleFractions(g)
	var mu [6]float64
	var mw [6]float64
	for i, sp := range AllSpecies {
		c := speciesConstantsTable[sp]
		mw[i] = c.MW
		if y[i] <= 0 {
			continue
		}
		m, err := LucasGasViscosity(T, c)
		if err != nil {
			return 0, err
		}
		mu[i] = float64(m)
	}

	var muMix float64
	for i := 0; i < len(AllSpecies); i++ {
		if y[i] <= 0 {
			continue
		}
		var denom float64
		for j := 0; j < len(AllSpecies); j++ {
			if y[j] <= 0 {
				continue
			}
			var phi float64
			if i == j {
				phi = 1.0
			} else {
				num := 1.0 + math.Sqrt(mu[i]/mu[j])*math.Pow(mw[j]/mw[i], 0.25)
				phi = (num * num) / math.Sqrt(8.0*(1.0+mw[i]/mw[j]))
			}
			denom += y[j] * phi
		}
		if denom > 0 {
			muMix += (y[i] * mu[i]) / denom
		}
	}
	return model.PaSec(muMix), nil
}

// PopulateGasProperties fills derived thermophysical and flow properties on g from its
// mole fractions, MolarFlowMolPerSec, Temperature, and Pressure.
func PopulateGasProperties(g *model.GasComposition) error {
	if g == nil {
		return fmt.Errorf("gas composition pointer is nil")
	}
	if g.Temperature <= 0 {
		return fmt.Errorf("gas temperature must be > 0 K, got %v", g.Temperature)
	}
	if g.Pressure <= 0 {
		return fmt.Errorf("gas pressure must be > 0 Pa, got %v", g.Pressure)
	}
	if g.MolarFlowMolPerSec < 0 {
		return fmt.Errorf("gas molar flow must be >= 0, got %v", g.MolarFlowMolPerSec)
	}

	mw := MixtureMolecularWeight(*g)
	cp, err := MixtureCp(*g, g.Temperature)
	if err != nil {
		return err
	}
	h, err := MixtureEnthalpy(*g, g.Temperature)
	if err != nil {
		return err
	}
	mu, err := MixtureViscosity(*g, g.Temperature)
	if err != nil {
		return err
	}

	density := (float64(g.Pressure) * mw) / (model.UniversalGasConst * float64(g.Temperature))
	massFlow := g.MolarFlowMolPerSec * mw
	volFlow := (g.MolarFlowMolPerSec * model.UniversalGasConst * float64(g.Temperature)) / float64(g.Pressure)

	g.MolecularWeight = mw
	g.Density = model.KgPerCubicMeter(density)
	g.Viscosity = mu
	g.CpJPerMolK = cp
	g.EnthalpyJPerMol = h
	g.MassFlow = model.KgPerSec(massFlow)
	g.VolumetricFlow = model.CubicMetersPerSec(volFlow)
	g.VolumetricFlowACFM = g.VolumetricFlow.CFM()
	return nil
}
