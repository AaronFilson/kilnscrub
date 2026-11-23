package scrubber

import (
	"testing"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

func TestScrubberBranchCoverageAndGuards(t *testing.T) {
	// 1. computeInletFlueGas: empty fuel default, unsupported fuel default, tempScale clamps, non-salt + salt warning.
	in := model.DefaultSaltKilnInput()
	in.FuelType = ""
	if gas, _, err := computeInletFlueGas(in); err != nil || gas.MolarFlowMolPerSec <= 0 {
		t.Fatalf("empty fuel computeInletFlueGas: %v", err)
	}
	in.FuelType = "coal"
	if _, _, err := computeInletFlueGas(in); err == nil {
		t.Error("expected unsupported fuel error")
	}
	inLow := model.DefaultSaltKilnInput()
	inLow.PeakExitTempF = 400.0 // tempScale < 0.35 clamp
	inLow.FuelType = model.FuelWood
	inLow.SaltRateLbsPerHr = 0.5 // warning: non-salt fuel with salt rate
	if _, warns, err := computeInletFlueGas(inLow); err != nil || len(warns) == 0 {
		t.Errorf("expected low-temp non-salt warning, warns=%v err=%v", warns, err)
	}
	inHigh := model.DefaultSaltKilnInput()
	inHigh.PeakExitTempF = 3900.0 // tempScale > 1.80 clamp (still < 2500 K)
	if _, _, err := computeInletFlueGas(inHigh); err != nil {
		t.Errorf("high tempScale clamp failed: %v", err)
	}

	// 2. Designer.Design: empty FuelType, invalid input, custom DuctDiameter, warnings (low/high throat vel & low spray).
	designer := NewDesigner()
	inBad := model.DefaultSaltKilnInput()
	inBad.KilnVolumeCuFt = -1
	if _, err := designer.Design(inBad); err == nil {
		t.Error("expected invalid input error")
	}
	inWarn := model.DefaultSaltKilnInput()
	inWarn.FuelType = ""
	inWarn.KilnVolumeCuFt = 3.0 // small spray < 0.50 GPM warning
	inWarn.VenturiThroatVelocity = 40.0
	inWarn.DuctDiameter = 0.20
	resWarn, err := designer.Design(inWarn)
	if err != nil {
		t.Fatalf("Design warning case failed: %v", err)
	}
	if len(resWarn.Warnings) < 2 {
		t.Errorf("expected >= 2 warnings for low throat vel & low spray, got %v", resWarn.Warnings)
	}

	// Large kiln -> thick quench/venturi/packed-bed walls, large motor HP.
	inBig := model.DefaultSaltKilnInput()
	inBig.KilnVolumeCuFt = 180.0
	inBig.PackingName = ""
	inBig.TargetHClRemoval = 0.9999999
	resBig, err := designer.Design(inBig)
	if err != nil {
		t.Fatalf("Design large kiln failed: %v", err)
	}
	if resBig.PackedBed.ColumnDiameter <= 0.60 {
		t.Errorf("expected large column diameter > 0.60 m, got %v", resBig.PackedBed.ColumnDiameter)
	}

	// 3. Blower helpers: selectStandardMotorHP > 100 HP, AutoSizeDuctDiameter guards, DuctPressureDrop laminar/zero-mu/invalid, SizeBlower guards.
	if hp := selectStandardMotorHP(120.0); hp < 125.0 {
		t.Errorf("selectStandardMotorHP(120) = %v, want >= 125", hp)
	}
	if d := AutoSizeDuctDiameter(model.GasComposition{}); float64(d) != ductMinDiamM {
		t.Errorf("AutoSizeDuctDiameter(0) = %v", d)
	}
	if d := AutoSizeDuctDiameter(model.GasComposition{VolumetricFlow: 0.001}); float64(d) != ductMinDiamM {
		t.Errorf("AutoSizeDuctDiameter(tiny) = %v", d)
	}
	if dp := DuctPressureDrop(model.GasComposition{}, 1.0, 0.2); dp != 0 {
		t.Errorf("DuctPressureDrop(zero) = %v", dp)
	}
	// Laminar Re < 2300.
	lamGas := model.GasComposition{VolumetricFlow: 0.001, Density: 1.0, Viscosity: 2e-5}
	if dp := DuctPressureDrop(lamGas, 3.0, 0.2); dp <= 0 {
		t.Errorf("DuctPressureDrop(laminar) = %v", dp)
	}
	// Zero viscosity fallback (fD = 0.02).
	noMuGas := model.GasComposition{VolumetricFlow: 0.2, Density: 1.0, Viscosity: 0}
	if dp := DuctPressureDrop(noMuGas, 3.0, 0.2); dp <= 0 {
		t.Errorf("DuctPressureDrop(mu=0) = %v", dp)
	}
	if _, err := SizeBlower(model.GasComposition{}, model.GasComposition{}, model.QuenchResult{}, model.VenturiResult{}, model.PackedBedResult{}, model.DefaultSaltKilnInput(), 0.2); err == nil {
		t.Error("SizeBlower zero outlet gas expected error")
	}

	// 4. Packed-bed helpers: Ergun invalid, Eckert invalid & X clamps, SizePackedBed invalid.
	if _, err := ErgunPressureDrop(0, 0.9, 1.0, 1.0, 1e-5, 1.0); err == nil {
		t.Error("ErgunPressureDrop(dp=0) expected error")
	}
	if _, err := EckertFloodingVelocity(1, 0, 1, 1000, 170, 0.001); err == nil {
		t.Error("EckertFloodingVelocity(gMass=0) expected error")
	}
	if vf, err := EckertFloodingVelocity(1e-6, 1.0, 0.8, 1000, 170, 0.001); err != nil || vf <= 0 {
		t.Errorf("Eckert low X clamp failed: vf=%v err=%v", vf, err)
	}
	if vf, err := EckertFloodingVelocity(500.0, 1.0, 0.8, 1000, 170, 0.001); err != nil || vf <= 0 {
		t.Errorf("Eckert high X clamp failed: vf=%v err=%v", vf, err)
	}
	if _, _, err := SizePackedBed(model.GasComposition{}, model.DefaultSaltKilnInput()); err == nil {
		t.Error("SizePackedBed zero gas expected error")
	}
	// Trigger HTU high clamp (>0.60) and low clamp (<0.15).
	inSmallPack := model.DefaultSaltKilnInput()
	inSmallPack.PackingDiameter = 0.002 // tiny packing -> HTU < 0.15 clamp
	if _, err := designer.Design(inSmallPack); err != nil {
		t.Errorf("Design tiny packing failed: %v", err)
	}
	inHugePack := model.DefaultSaltKilnInput()
	inHugePack.PackingDiameter = 0.50 // huge packing -> HTU > 0.60 clamp
	if _, err := designer.Design(inHugePack); err != nil {
		t.Errorf("Design huge packing failed: %v", err)
	}

	// 5. Venturi helpers: CunninghamSlip, CalvertF, CalvertPenetration, CalvertCutDiameter bracket widening,
	// NukiyamaTanasawa clamps, VenturiDropletAccelFactor, SizeVenturi guards.
	if cc := CunninghamSlip(0, 6.6e-8); cc != 1.0 {
		t.Errorf("CunninghamSlip(0) = %v, want 1.0", cc)
	}
	if f := CalvertF(0, 0.5); f != 0 {
		t.Errorf("CalvertF(0) = %v, want 0", f)
	}
	if pt := CalvertPenetration(0, 2e-5, 50e-6, 70, 2160, 0.0012, 1000, 0.5, 6.6e-8); pt != 1.0 {
		t.Errorf("CalvertPenetration(0) = %v, want 1.0", pt)
	}
	if _, err := CalvertCutDiameter(0, 50e-6, 70, 2160, 0.0012, 1000, 0.5, 6.6e-8); err == nil {
		t.Error("CalvertCutDiameter(mu=0) expected error")
	}
	// Trigger bracket widening high *= 10 (very weak venturi) and low *= 0.1 (ultra-high energy venturi),
	// exercising both the widening loops and the unbracketed Brent error guard.
	if _, err := CalvertCutDiameter(2e-5, 5e-3, 1.0, 500, 1e-6, 1000, 0.1, 6.6e-8); err == nil {
		t.Error("expected unbracketed error for ultra-weak venturi")
	}
	if _, err := CalvertCutDiameter(1e-6, 5e-6, 2000.0, 20000, 0.5, 1000, 0.5, 6.6e-8); err == nil {
		t.Error("expected unbracketed error for ultra-strong venturi")
	}
	if d := NukiyamaTanasawaDropletDiameter(500.0, 0.0001); d.Microns() < 19.999 {
		t.Errorf("NukiyamaTanasawa min clamp failed: %v um", d.Microns())
	}
	if d := NukiyamaTanasawaDropletDiameter(10.0, 0.05); d.Microns() > 250.001 {
		t.Errorf("NukiyamaTanasawa max clamp failed: %v um", d.Microns())
	}
	if b := VenturiDropletAccelFactor(0, 50e-6, 70, 0.8, 1000, 2e-5); b != 1.0 {
		t.Errorf("VenturiDropletAccelFactor(0) = %v", b)
	}
	// High Re0 -> cd0 < 0.44 clamp.
	if b := VenturiDropletAccelFactor(0.1, 5e-3, 200, 2.0, 1000, 1e-6); b <= 0 || b > 1.0 {
		t.Errorf("VenturiDropletAccelFactor(high Re) = %v", b)
	}
	if _, err := SizeVenturi(model.GasComposition{}, model.DefaultSaltKilnInput(), 0.2); err == nil {
		t.Error("SizeVenturi zero gas expected error")
	}
	// Auto droplet diameter (0) + small duct triggering inletDiam expansion.
	inAuto := model.DefaultSaltKilnInput()
	inAuto.VenturiDropletDiameter = 0
	if _, err := SizeVenturi(resBig.QuenchedGas, inAuto, 0.05); err != nil {
		t.Errorf("SizeVenturi auto droplet + small duct failed: %v", err)
	}
}
