package scrubber

import (
	"fmt"
	"math"

	"github.com/AaronFilson/kilnscrub/pkg/model"
	"github.com/AaronFilson/kilnscrub/pkg/thermo"
)

// Standard Calvert venturi physics constants.
const (
	DefaultCalvertFactor   = 0.5    // Empirical Calvert droplet-inertia factor f for hydrophobic/wetting balance
	DefaultMeanFreePathM   = 6.6e-8 // Gas mean free path [m] near atmospheric conditions (66 nm)
	DefaultVenturiRhoL     = 1000.0 // Reference water density [kg/m^3] matching oracle Calvert fixture
	venturiMaterialSpec    = "316L Stainless Steel"
	venturiConvergeHalfDeg = 12.5 // Standard 25 deg included converging angle
	venturiDivergeHalfDeg  = 3.5  // Standard 7 deg included diverging diffuser angle for pressure recovery
	venturiDiffuserLossK   = 0.25 // Gas-phase diffuser irreversible head loss coefficient
	venturiWallThinM       = 0.002
	venturiWallThickM      = 0.003
)

// CunninghamSlip computes the Davies (1945) / Allen-Raabe Cunningham slip correction factor
// for a spherical particle of physical diameter dParticle [m] in a gas with meanFreePath [m].
func CunninghamSlip(dParticle, meanFreePath float64) float64 {
	if dParticle <= 0 {
		return 1.0
	}
	kn := 2.0 * meanFreePath / dParticle
	return 1.0 + kn*(1.257+0.4*math.Exp(-1.1/kn))
}

// CalvertF evaluates the Calvert (1970, 1972) inertial impaction integral function:
//
//	F(Kp, f) = (1/Kp) * (-0.7 - x + 1.4*ln(1 + x/0.7) + 0.49/(0.7 + x)),  x = Kp * f
//
// Uses math.Log1p(x / 0.7) for moderate/large x and switches to the exact small-x Taylor series
// for u = x/0.7 <= 0.05 to avoid catastrophic floating-point cancellation for submicron particles.
func CalvertF(kp, fCalvert float64) float64 {
	if kp <= 0 || fCalvert <= 0 {
		return 0.0
	}
	x := kp * fCalvert
	u := x / 0.7
	if u <= 0.05 {
		// Exact expansion: F(Kp, f) = f * sum_{k=3}^{inf} (-1)^k * ((k-2)/k) * u^(k-1).
		// Truncating at k = 14 gives relative error < 3e-16 for u <= 0.05.
		uPow := u * u // u^(k-1) starting at k = 3
		sign := -1.0  // (-1)^3
		var sum float64
		for k := 3; k <= 14; k++ {
			sum += sign * (float64(k-2) / float64(k)) * uPow
			uPow *= u
			sign = -sign
		}
		return fCalvert * sum
	}
	return (-0.7 - x + 1.4*math.Log1p(u) + 0.49/(0.7+x)) / kp
}

// CalvertPenetration returns the fractional particle penetration Pt(d_p) on [0, 1]
// according to the Calvert (1970, 1972) wet venturi impaction model.
func CalvertPenetration(
	dParticle float64,
	muG float64,
	dDroplet float64,
	vThroat float64,
	rhoP float64,
	lgRatio float64,
	rhoL float64,
	fCalvert float64,
	meanFreePath float64,
) float64 {
	if dParticle <= 0 || muG <= 0 || dDroplet <= 0 || vThroat <= 0 || rhoP <= 0 || lgRatio <= 0 || rhoL <= 0 {
		return 1.0
	}
	cc := CunninghamSlip(dParticle, meanFreePath)
	kp := (cc * rhoP * dParticle * dParticle * vThroat) / (9.0 * muG * dDroplet)
	fVal := CalvertF(kp, fCalvert)
	exponent := (2.0 * lgRatio * rhoL * vThroat * dDroplet / (55.0 * muG)) * fVal
	if exponent >= 0 {
		return 1.0
	}
	pt := math.Exp(exponent)
	if pt < 0 {
		return 0
	}
	if pt > 1 {
		return 1
	}
	return pt
}

// CalvertEfficiency returns the fractional particulate collection efficiency 1 - Pt(d_p) on [0, 1].
func CalvertEfficiency(
	dParticle float64,
	muG float64,
	dDroplet float64,
	vThroat float64,
	rhoP float64,
	lgRatio float64,
	rhoL float64,
	fCalvert float64,
	meanFreePath float64,
) float64 {
	eff := 1.0 - CalvertPenetration(dParticle, muG, dDroplet, vThroat, rhoP, lgRatio, rhoL, fCalvert, meanFreePath)
	if eff < 0 {
		return 0
	}
	if eff > 1 {
		return 1
	}
	return eff
}

// CalvertCutDiameter solves for the physical particle diameter d_50 [m] at which
// Calvert fractional penetration Pt(d_50) = 0.5 (50% collection efficiency).
func CalvertCutDiameter(
	muG float64,
	dDroplet float64,
	vThroat float64,
	rhoP float64,
	lgRatio float64,
	rhoL float64,
	fCalvert float64,
	meanFreePath float64,
) (float64, error) {
	if muG <= 0 || dDroplet <= 0 || vThroat <= 0 || rhoP <= 0 || lgRatio <= 0 || rhoL <= 0 || fCalvert <= 0 {
		return 0, fmt.Errorf("calvert parameters must all be > 0")
	}
	residual := func(dp float64) float64 {
		return CalvertPenetration(dp, muG, dDroplet, vThroat, rhoP, lgRatio, rhoL, fCalvert, meanFreePath) - 0.5
	}

	low := 1e-9
	high := 1e-3
	// Widen bracket if needed for extreme operating points.
	for residual(low) < 0 && low > 1e-12 {
		low *= 0.1
	}
	for residual(high) > 0 && high < 1.0 {
		high *= 10.0
	}

	d50, err := thermo.BrentRoot(residual, low, high, 1e-18, 200)
	if err != nil {
		return 0, fmt.Errorf("calvert d50 root solve failed: %w", err)
	}
	return d50, nil
}

// NukiyamaTanasawaDropletDiameter estimates Sauter mean water droplet diameter [m]
// in a wet venturi throat when VenturiDropletDiameter is not explicitly specified.
func NukiyamaTanasawaDropletDiameter(vThroatMS, lgRatio float64) model.Meters {
	// Water at ~20 C: sigma = 72.8 dyn/cm, rho_L = 1.0 g/cm^3, mu_L = 0.01 poise.
	const (
		sigmaDynCm = 72.8
		rhoLGcm3   = 1.0
		muLPoise   = 0.01
	)
	v := math.Max(vThroatMS, 10.0)
	dMicrons := (585.0/v)*math.Sqrt(sigmaDynCm/rhoLGcm3) +
		597.0*math.Pow(muLPoise/math.Sqrt(sigmaDynCm*rhoLGcm3), 0.45)*math.Pow(1000.0*lgRatio, 1.5)
	if dMicrons < 20.0 {
		dMicrons = 20.0
	}
	if dMicrons > 250.0 {
		dMicrons = 250.0
	}
	return model.MicronsToMeters(dMicrons)
}

// logNormalOverallEfficiency integrates fractional capture efficiency over a log-normal
// mass-weighted particle size distribution with geometric mean diameter dg [m] and
// geometric standard deviation sigmaG.
func logNormalOverallEfficiency(
	dg, sigmaG float64,
	effFn func(dp float64) float64,
) float64 {
	lnDg := math.Log(dg)
	lnSig := math.Log(sigmaG)
	const (
		zMin  = -4.0
		zMax  = 4.0
		steps = 160
	)
	dz := (zMax - zMin) / float64(steps)
	invSqrt2Pi := 1.0 / math.Sqrt(2.0*math.Pi)

	var num, den float64
	for i := 0; i <= steps; i++ {
		z := zMin + float64(i)*dz
		weight := dz
		if i == 0 || i == steps {
			weight *= 0.5
		}
		pdf := invSqrt2Pi * math.Exp(-0.5*z*z)
		dp := math.Exp(lnDg + z*lnSig)
		w := weight * pdf
		num += w * effFn(dp)
		den += w
	}
	if den <= 0 {
		return 0
	}
	eff := num / den
	if eff < 0 {
		return 0
	}
	if eff > 1 {
		return 1
	}
	return eff
}

// SizeVenturi computes Stage 2 (316L SS Wet Venturi Scrubber) geometry, Calvert cut diameters,
// fractional efficiency curve (0.1-2.0 um), overall efficiencies, and throat pressure drop.
func SizeVenturi(
	quenchedGas model.GasComposition,
	input model.DesignInput,
	ductDiameter model.Meters,
) (model.VenturiResult, error) {
	if quenchedGas.VolumetricFlow <= 0 || quenchedGas.Viscosity <= 0 || quenchedGas.Density <= 0 {
		return model.VenturiResult{}, fmt.Errorf("quenched gas flow, density, and viscosity must be > 0")
	}

	vThroat := float64(input.VenturiThroatVelocity)
	lgRatio := input.VenturiLGRatio
	dDroplet := input.VenturiDropletDiameter
	if dDroplet <= 0 {
		dDroplet = NukiyamaTanasawaDropletDiameter(vThroat, lgRatio)
	}

	muG := float64(quenchedGas.Viscosity)
	rhoG := float64(quenchedGas.Density)
	rhoL := float64(model.WaterDensity)
	dDropM := float64(dDroplet)

	// Cut diameters d_50: NaCl fume (2160 kg/m^3) in CutDiameterD50, wood ash (2400 kg/m^3) in WoodAshCutDiameterD50.
	naclD50, err := CalvertCutDiameter(muG, dDropM, vThroat, float64(model.NaClParticleDensity), lgRatio, rhoL, DefaultCalvertFactor, DefaultMeanFreePathM)
	if err != nil {
		return model.VenturiResult{}, fmt.Errorf("NaCl d50: %w", err)
	}
	ashD50, err := CalvertCutDiameter(muG, dDropM, vThroat, float64(model.WoodAshParticleDensity), lgRatio, rhoL, DefaultCalvertFactor, DefaultMeanFreePathM)
	if err != nil {
		return model.VenturiResult{}, fmt.Errorf("wood ash d50: %w", err)
	}

	// Fractional collection efficiency curve across 0.1 - 2.0 um physical diameter.
	diametersUm := []float64{
		0.10, 0.15, 0.20, 0.25, 0.30, 0.35, 0.40, 0.50,
		0.60, 0.70, 0.80, 0.90, 1.00, 1.20, 1.40, 1.60, 1.80, 2.00,
	}
	curve := make([]model.EfficiencyPoint, len(diametersUm))
	for i, dUm := range diametersUm {
		dM := float64(model.MicronsToMeters(dUm))
		ashEff := CalvertEfficiency(dM, muG, dDropM, vThroat, float64(model.WoodAshParticleDensity), lgRatio, rhoL, DefaultCalvertFactor, DefaultMeanFreePathM)
		naclEff := CalvertEfficiency(dM, muG, dDropM, vThroat, float64(model.NaClParticleDensity), lgRatio, rhoL, DefaultCalvertFactor, DefaultMeanFreePathM)
		curve[i] = model.EfficiencyPoint{
			ParticleDiameter:        model.Meters(dM),
			ParticleDiameterMicrons: dUm,
			WoodAshEfficiency:       ashEff,
			NaClFumeEfficiency:      naclEff,
		}
	}

	// Overall mass-weighted collection efficiencies for typical kiln aerosol distributions:
	// - Wood fly ash: log-normal mass median ~1.5 um, sigma_g = 2.0
	// - Condensed NaCl salt fume: log-normal mass median ~0.55 um, sigma_g = 1.8
	overallAsh := logNormalOverallEfficiency(1.5e-6, 2.0, func(dp float64) float64 {
		return CalvertEfficiency(dp, muG, dDropM, vThroat, float64(model.WoodAshParticleDensity), lgRatio, rhoL, DefaultCalvertFactor, DefaultMeanFreePathM)
	})
	overallNaCl := logNormalOverallEfficiency(0.55e-6, 1.8, func(dp float64) float64 {
		return CalvertEfficiency(dp, muG, dDropM, vThroat, float64(model.NaClParticleDensity), lgRatio, rhoL, DefaultCalvertFactor, DefaultMeanFreePathM)
	})

	// Throat & cone geometry.
	qGas := float64(quenchedGas.VolumetricFlow)
	throatArea := qGas / vThroat
	throatDiam := math.Sqrt(4.0 * throatArea / math.Pi)

	inletDiam := float64(ductDiameter)
	if inletDiam < 2.2*throatDiam {
		inletDiam = math.Ceil((2.5*throatDiam)/0.01) * 0.01
	}
	outletDiam := inletDiam

	convRad := venturiConvergeHalfDeg * math.Pi / 180.0
	divRad := venturiDivergeHalfDeg * math.Pi / 180.0
	convLen := math.Max(0.10, (inletDiam-throatDiam)/(2.0*math.Tan(convRad)))
	throatLen := math.Max(0.08, 1.5*throatDiam)
	divLen := math.Max(0.20, (outletDiam-throatDiam)/(2.0*math.Tan(divRad)))

	wallThickness := venturiWallThinM
	if inletDiam > 0.35 {
		wallThickness = venturiWallThickM
	}

	liquidFlow := model.CubicMetersPerSec(lgRatio * qGas)

	// Calvert throat pressure drop: droplet acceleration (f * rho_L * v_t^2 * L/G) + gas diffuser loss.
	dpAccel := DefaultCalvertFactor * rhoL * vThroat * vThroat * lgRatio
	dpDiffuser := venturiDiffuserLossK * 0.5 * rhoG * vThroat * vThroat
	venturiDP := model.Pascal(dpAccel + dpDiffuser)

	return model.VenturiResult{
		InletDiameter:                model.Meters(inletDiam),
		ThroatDiameter:               model.Meters(throatDiam),
		OutletDiameter:               model.Meters(outletDiam),
		ConvergingLength:             model.Meters(convLen),
		ThroatLength:                 model.Meters(throatLen),
		DivergingLength:              model.Meters(divLen),
		WallThickness:                model.Meters(wallThickness),
		ThroatVelocity:               model.MetersPerSec(vThroat),
		LGRatio:                      lgRatio,
		LiquidFlow:                   liquidFlow,
		LiquidFlowGPM:                liquidFlow.GPM(),
		DropletDiameter:              dDroplet,
		CutDiameterD50:               model.Meters(naclD50),
		CutDiameterD50Microns:        model.Meters(naclD50).Microns(),
		WoodAshCutDiameterD50:        model.Meters(ashD50),
		WoodAshCutDiameterD50Microns: model.Meters(ashD50).Microns(),
		EfficiencyCurve:              curve,
		OverallWoodAshEfficiency:     overallAsh,
		OverallNaClFumeEfficiency:    overallNaCl,
		PressureDrop:                 venturiDP,
		PressureDropInWC:             venturiDP.InWC(),
		Material:                     venturiMaterialSpec,
	}, nil
}
