package scrubber

import (
	"fmt"
	"math"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

// Stage 4 Induced-Draft (ID) centrifugal blower engineering constants.
const (
	blowerStaticEfficiency = 0.62 // Industrial corrosion-resistant centrifugal fan efficiency on (0, 1]
	blowerMotorSafetyFac   = 1.20 // 20% service factor margin for standard NEMA motor selection
	blowerPressureCoeffPsi = 1.00 // Narrow radial-blade pressure-blower static pressure coefficient (low-Ns high-head duty)
	blowerFlowCoeffPhi     = 0.14 // Typical centrifugal impeller flow coefficient
	blowerMinWheelDiamM    = 0.20 // Minimum wheel diameter [m] (~8 inches)
	blowerMaxRPM           = 3550.0
	blowerMaterialSpec     = "FRP / Polypropylene Housing, 316L SS Wheel"
	ductRoughnessM         = 0.00015 // Commercial 316L SS / galvanized flue duct roughness [m]
	ductMinorLossSumK      = 2.5     // Entrance + 2x 90-deg elbows + exit transition
	ductTargetVelMS        = 12.0    // Target hot flue duct velocity [m/s]
	ductMinDiamM           = 0.1524  // 6 inch minimum flue duct [m]
)

var standardMotorHPs = []float64{
	0.25, 0.33, 0.5, 0.75, 1.0, 1.5, 2.0, 3.0, 5.0, 7.5,
	10.0, 15.0, 20.0, 25.0, 30.0, 40.0, 50.0, 75.0, 100.0,
}

// selectStandardMotorHP returns the next standard NEMA motor rating >= shaftHP * safetyFactor.
func selectStandardMotorHP(shaftHP float64) float64 {
	req := shaftHP * blowerMotorSafetyFac
	for _, hp := range standardMotorHPs {
		if hp >= req {
			return hp
		}
	}
	return math.Ceil(req/25.0) * 25.0
}

// AutoSizeDuctDiameter selects a flue duct diameter [m] targeting ~12 m/s hot gas velocity,
// rounded up to the nearest 1-inch nominal size (minimum 6 inches).
func AutoSizeDuctDiameter(inletGas model.GasComposition) model.Meters {
	if inletGas.VolumetricFlow <= 0 {
		return model.Meters(ductMinDiamM)
	}
	raw := math.Sqrt(4.0 * float64(inletGas.VolumetricFlow) / (math.Pi * ductTargetVelMS))
	// Round up to nearest 1 inch (0.0254 m).
	nom := math.Ceil(raw/model.MetersPerInch) * model.MetersPerInch
	if nom < ductMinDiamM {
		nom = ductMinDiamM
	}
	return model.Meters(nom)
}

// DuctPressureDrop calculates Darcy-Weisbach friction + minor losses [Pa] for flue ducting
// using the Haaland explicit approximation to the Colebrook friction factor.
func DuctPressureDrop(gas model.GasComposition, ductLength, ductDiameter model.Meters) model.Pascal {
	if ductDiameter <= 0 || gas.VolumetricFlow <= 0 || gas.Density <= 0 {
		return 0
	}
	d := float64(ductDiameter)
	l := math.Max(0, float64(ductLength))
	area := math.Pi * d * d / 4.0
	v := float64(gas.VolumetricFlow) / area
	rho := float64(gas.Density)
	mu := float64(gas.Viscosity)

	var fD float64
	if mu > 0 && v > 0 {
		re := rho * v * d / mu
		if re < 2300.0 {
			fD = 64.0 / math.Max(re, 1.0)
		} else {
			// Haaland (1983) explicit Darcy friction factor.
			term := math.Pow((ductRoughnessM/d)/3.7, 1.11) + 6.9/re
			invSqrtF := -1.8 * math.Log10(term)
			fD = 1.0 / (invSqrtF * invSqrtF)
		}
	} else {
		fD = 0.02
	}

	kTotal := fD*(l/d) + ductMinorLossSumK
	return model.Pascal(kTotal * 0.5 * rho * v * v)
}

// SizeBlower computes Stage 4 (Induced-Draft Centrifugal Blower) total static pressure,
// breakdown (without double-counting demister, and with positive kiln draft draw magnitude),
// power, wheel diameter, RPM, and fan/system performance curves.
func SizeBlower(
	inletGas model.GasComposition,
	outletGas model.GasComposition,
	quench model.QuenchResult,
	venturi model.VenturiResult,
	packedBed model.PackedBedResult,
	input model.DesignInput,
	ductDiameter model.Meters,
) (model.BlowerResult, error) {
	if outletGas.VolumetricFlow <= 0 || outletGas.Density <= 0 {
		return model.BlowerResult{}, fmt.Errorf("outlet gas flow and density must be > 0")
	}

	ductDP := DuctPressureDrop(inletGas, input.DuctLength, ductDiameter)
	// KilnDraft in PressureDropBreakdown stores the positive draw magnitude abs(TargetKilnDraft)
	// so Breakdown.TotalPa() sums directly to TotalStaticPressure.
	kilnDraftDrawPa := model.InWCToPascal(math.Abs(input.TargetKilnDraftInWC))

	breakdown := model.PressureDropBreakdown{
		Duct:          ductDP,
		DuctInWC:      ductDP.InWC(),
		Quench:        quench.PressureDrop,
		QuenchInWC:    quench.PressureDrop.InWC(),
		Venturi:       venturi.PressureDrop,
		VenturiInWC:   venturi.PressureDrop.InWC(),
		PackedBed:     packedBed.BedPressureDrop, // Packing bed ONLY (excludes Demister!)
		PackedBedInWC: packedBed.BedPressureDrop.InWC(),
		Demister:      packedBed.DemisterPressureDrop,
		DemisterInWC:  packedBed.DemisterPressureDrop.InWC(),
		KilnDraft:     kilnDraftDrawPa,
		KilnDraftInWC: kilnDraftDrawPa.InWC(),
	}

	totalSP := breakdown.TotalPa()
	// Account for suction vacuum density expansion at the Stage 4 ID blower inlet
	// (P_in = P_amb - dP_stages_1_3) so blower ACFM and shaft power reflect real inlet density.
	pAmb := float64(outletGas.Pressure)
	if pAmb <= 0 {
		pAmb = 101325.0
	}
	suctionDrop := float64(ductDP + quench.PressureDrop + venturi.PressureDrop + packedBed.TotalPressureDrop + kilnDraftDrawPa)
	pIn := math.Max(pAmb-suctionDrop, 0.5*pAmb)
	densityRatio := pIn / pAmb
	rhoOp := float64(outletGas.Density) * densityRatio
	qOp := model.CubicMetersPerSec(float64(outletGas.VolumetricFlow) / densityRatio)

	eta := blowerStaticEfficiency
	shaftPowerW := model.Watts((float64(qOp) * float64(totalSP)) / eta)
	shaftHP := shaftPowerW.HP()
	motorHP := selectStandardMotorHP(shaftHP)

	// Impeller diameter and operating RPM from pressure/flow coefficients.
	uTip := math.Sqrt(2.0 * float64(totalSP) / (blowerPressureCoeffPsi * rhoOp))
	rawWheelDiam := math.Sqrt(4.0 * float64(qOp) / (math.Pi * blowerFlowCoeffPhi * math.Max(uTip, 1.0)))
	wheelDiam := math.Ceil(rawWheelDiam/0.01) * 0.01
	if wheelDiam < blowerMinWheelDiamM {
		wheelDiam = blowerMinWheelDiamM
	}
	rpm := (uTip / (math.Pi * wheelDiam)) * 60.0
	if rpm > blowerMaxRPM {
		wheelDiam = math.Ceil((uTip*60.0/(math.Pi*blowerMaxRPM))/0.01) * 0.01
		rpm = (uTip / (math.Pi * wheelDiam)) * 60.0
	}

	opPoint := model.CurvePoint{
		Flow:         qOp,
		FlowACFM:     qOp.CFM(),
		Pressure:     totalSP,
		PressureInWC: totalSP.InWC(),
	}

	// Generate 21-point system resistance curve and drooping centrifugal fan curve over [0, 1.6*Q_op],
	// intersecting cleanly at OperatingPoint (Q_op, totalSP).
	const numPts = 21
	qMax := 1.6 * float64(qOp)
	kSys := (float64(totalSP) - float64(kilnDraftDrawPa)) / (float64(qOp) * float64(qOp))
	if kSys < 0 {
		kSys = 0
	}

	fanCurve := make([]model.CurvePoint, numPts)
	sysCurve := make([]model.CurvePoint, numPts)
	for i := 0; i < numPts; i++ {
		q := qMax * float64(i) / float64(numPts-1)
		u := q / float64(qOp)

		// System curve: P_sys(Q) = P_draft + K_sys * Q^2
		pSys := model.Pascal(float64(kilnDraftDrawPa) + kSys*q*q)
		sysCurve[i] = model.CurvePoint{
			Flow:         model.CubicMetersPerSec(q),
			FlowACFM:     model.CubicMetersPerSec(q).CFM(),
			Pressure:     pSys,
			PressureInWC: pSys.InWC(),
		}

		// Drooping centrifugal fan curve: P_fan(u) = P_op * (1.30 - 0.10*u - 0.20*u^2), P_fan(1) == P_op.
		pFanVal := float64(totalSP) * (1.30 - 0.10*u - 0.20*u*u)
		if pFanVal < 0 {
			pFanVal = 0
		}
		pFan := model.Pascal(pFanVal)
		fanCurve[i] = model.CurvePoint{
			Flow:         model.CubicMetersPerSec(q),
			FlowACFM:     model.CubicMetersPerSec(q).CFM(),
			Pressure:     pFan,
			PressureInWC: pFan.InWC(),
		}
	}

	return model.BlowerResult{
		Breakdown:               breakdown,
		TargetKilnDraftInWC:     input.TargetKilnDraftInWC, // Signed negative draft echo
		TotalStaticPressure:     totalSP,
		TotalStaticPressureInWC: totalSP.InWC(),
		Flow:                    qOp,
		FlowACFM:                qOp.CFM(),
		Efficiency:              eta,
		ShaftPower:              shaftPowerW,
		ShaftPowerHP:            shaftHP,
		RecommendedMotorHP:      motorHP,
		WheelDiameter:           model.Meters(wheelDiam),
		OperatingRPM:            rpm,
		OperatingPoint:          opPoint,
		FanCurve:                fanCurve,
		SystemCurve:             sysCurve,
		Material:                blowerMaterialSpec,
	}, nil
}
