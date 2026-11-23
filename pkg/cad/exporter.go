package cad

import (
	"fmt"
	"math"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

// Compile-time check that *Exporter implements model.CADExporter.
var _ model.CADExporter = (*Exporter)(nil)

// Exporter generates fabrication-ready 3D STL meshes, 2D AutoCAD R12 DXF
// sheet-metal flat patterns, and Bills of Materials from a ScrubberDesign.
type Exporter struct{}

// NewExporter returns a new CAD and BOM exporter implementing model.CADExporter.
func NewExporter() *Exporter {
	return &Exporter{}
}

// ExportSTL generates a binary STL mesh (units: millimeters) for the full 4-stage scrubber assembly.
func ExportSTL(design model.ScrubberDesign) ([]byte, error) {
	return NewExporter().ExportSTL(design)
}

// ExportDXF generates an AutoCAD R12 ASCII DXF (units: millimeters) containing unfolded
// sheet-metal cut templates, exact cone developments, and flange rings.
func ExportDXF(design model.ScrubberDesign) ([]byte, error) {
	return NewExporter().ExportDXF(design)
}

// GenerateBOM builds a realistic fabrication Bill of Materials from a ScrubberDesign.
func GenerateBOM(design model.ScrubberDesign) (model.BOM, error) {
	return NewExporter().GenerateBOM(design)
}

// FormatBOMCSV formats a BOM as RFC 4180 CSV bytes.
func FormatBOMCSV(bom model.BOM) ([]byte, error) {
	return NewExporter().FormatBOMCSV(bom)
}

// FormatBOMMarkdown formats a BOM as a GitHub-flavored Markdown report.
func FormatBOMMarkdown(bom model.BOM) ([]byte, error) {
	return NewExporter().FormatBOMMarkdown(bom)
}

func isFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

func checkPosFinite(name string, v float64) error {
	if !isFinite(v) || v <= 0 {
		return fmt.Errorf("cad: %s must be finite and > 0, got %v", name, v)
	}
	return nil
}

func checkNonNegFinite(name string, v float64) error {
	if !isFinite(v) || v < 0 {
		return fmt.Errorf("cad: %s must be finite and >= 0, got %v", name, v)
	}
	return nil
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// validateDesign verifies that all geometric and sizing fields required for CAD/BOM
// generation are finite and physically valid.
func validateDesign(d model.ScrubberDesign) error {
	if err := checkPosFinite("quench.tower_diameter_m", float64(d.Quench.TowerDiameter)); err != nil {
		return err
	}
	if err := checkPosFinite("quench.tower_height_m", float64(d.Quench.TowerHeight)); err != nil {
		return err
	}
	if err := checkNonNegFinite("quench.wall_thickness_m", float64(d.Quench.WallThickness)); err != nil {
		return err
	}
	if err := checkNonNegFinite("quench.total_spray_water_gpm", d.Quench.TotalSprayWaterGPM); err != nil {
		return err
	}

	if err := checkPosFinite("venturi.inlet_diameter_m", float64(d.Venturi.InletDiameter)); err != nil {
		return err
	}
	if err := checkPosFinite("venturi.throat_diameter_m", float64(d.Venturi.ThroatDiameter)); err != nil {
		return err
	}
	if err := checkPosFinite("venturi.outlet_diameter_m", float64(d.Venturi.OutletDiameter)); err != nil {
		return err
	}
	if err := checkPosFinite("venturi.converging_length_m", float64(d.Venturi.ConvergingLength)); err != nil {
		return err
	}
	if err := checkPosFinite("venturi.throat_length_m", float64(d.Venturi.ThroatLength)); err != nil {
		return err
	}
	if err := checkPosFinite("venturi.diverging_length_m", float64(d.Venturi.DivergingLength)); err != nil {
		return err
	}
	if err := checkNonNegFinite("venturi.wall_thickness_m", float64(d.Venturi.WallThickness)); err != nil {
		return err
	}
	if d.Venturi.ThroatDiameter > d.Venturi.InletDiameter || d.Venturi.ThroatDiameter > d.Venturi.OutletDiameter {
		return fmt.Errorf("cad: venturi.throat_diameter_m (%v) cannot exceed inlet (%v) or outlet (%v) diameter",
			d.Venturi.ThroatDiameter, d.Venturi.InletDiameter, d.Venturi.OutletDiameter)
	}

	if err := checkPosFinite("packed_bed.column_diameter_m", float64(d.PackedBed.ColumnDiameter)); err != nil {
		return err
	}
	if err := checkPosFinite("packed_bed.column_height_m", float64(d.PackedBed.ColumnHeight)); err != nil {
		return err
	}
	if err := checkNonNegFinite("packed_bed.wall_thickness_m", float64(d.PackedBed.WallThickness)); err != nil {
		return err
	}
	if err := checkPosFinite("packed_bed.packing_height_m", float64(d.PackedBed.PackingHeight)); err != nil {
		return err
	}
	if d.PackedBed.PackingHeight > d.PackedBed.ColumnHeight {
		return fmt.Errorf("cad: packed_bed.packing_height_m (%v) exceeds column_height_m (%v)",
			d.PackedBed.PackingHeight, d.PackedBed.ColumnHeight)
	}
	if err := checkNonNegFinite("packed_bed.demister_height_m", float64(d.PackedBed.DemisterHeight)); err != nil {
		return err
	}
	if err := checkNonNegFinite("packed_bed.packing_volume_m3", float64(d.PackedBed.PackingVolume)); err != nil {
		return err
	}
	if err := checkNonNegFinite("packed_bed.sump_volume_m3", float64(d.PackedBed.SumpVolume)); err != nil {
		return err
	}
	if err := checkNonNegFinite("packed_bed.recirculation_flow_gpm", d.PackedBed.RecirculationFlowGPM); err != nil {
		return err
	}
	if err := checkNonNegFinite("packed_bed.nahco3_consumption_kg_hr", d.PackedBed.NaHCO3ConsumptionKgPerHr); err != nil {
		return err
	}

	if err := checkPosFinite("blower.wheel_diameter_m", float64(d.Blower.WheelDiameter)); err != nil {
		return err
	}
	if err := checkNonNegFinite("blower.recommended_motor_hp", d.Blower.RecommendedMotorHP); err != nil {
		return err
	}
	if err := checkNonNegFinite("blower.flow_acfm", d.Blower.FlowACFM); err != nil {
		return err
	}
	if err := checkNonNegFinite("blower.total_static_pressure_in_wc", d.Blower.TotalStaticPressureInWC); err != nil {
		return err
	}

	if err := checkNonNegFinite("envelope.duct_diameter_m", float64(d.Envelope.DuctDiameter)); err != nil {
		return err
	}
	if err := checkNonNegFinite("input.duct_diameter_m", float64(d.Input.DuctDiameter)); err != nil {
		return err
	}
	if err := checkNonNegFinite("input.duct_length_m", float64(d.Input.DuctLength)); err != nil {
		return err
	}

	return nil
}

// flangeSpec captures fabrication dimensions for a bolted circular flange ring (in meters).
type flangeSpec struct {
	ID        float64 // inner bore diameter (m)
	OD        float64 // outer flange diameter (m)
	BCD       float64 // bolt circle diameter (m)
	Thickness float64 // flange plate thickness (m)
	BoltCount int
	BoltHoleD float64 // bolt hole diameter (m)
}

func makeFlangeSpec(boreDia, wallThickness float64) flangeSpec {
	lip := clamp(0.14*boreDia, 0.038, 0.080)
	od := boreDia + 2.0*lip
	bcd := boreDia + lip
	thick := math.Max(0.010, 3.0*wallThickness)
	bolts := 8
	holeD := 0.012
	if boreDia > 0.60 {
		bolts = 16
		holeD = 0.016
	} else if boreDia > 0.30 {
		bolts = 12
		holeD = 0.014
	}
	return flangeSpec{
		ID:        boreDia,
		OD:        od,
		BCD:       bcd,
		Thickness: thick,
		BoltCount: bolts,
		BoltHoleD: holeD,
	}
}

// assemblyGeometry holds unified derived dimensions (in meters) shared by STL, DXF, and BOM.
type assemblyGeometry struct {
	// Inlet flue duct
	ductDia      float64
	ductStubLen  float64
	ductTotalLen float64
	ductWall     float64

	// Stage 1: Quench tower (316L SS)
	quenchDia  float64
	quenchH    float64
	quenchWall float64
	quenchNozz int

	// Stage 2: Venturi (316L SS)
	ventInDia   float64
	ventThDia   float64
	ventOutDia  float64
	ventConvLen float64
	ventThLen   float64
	ventDivLen  float64
	ventWall    float64

	// Stage 3: Packed-bed column + sump + demister + stack (CPVC/PP)
	colDia      float64
	colH        float64
	colWall     float64
	packH       float64
	packVol     float64
	demisterH   float64
	sumpH       float64
	sumpVol     float64
	stackDia    float64
	reducerH    float64
	stackRiserH float64

	// Stage 4: ID centrifugal blower
	wheelDia    float64
	blowerWidth float64

	// Standardized flange rings
	flangeDuct    flangeSpec
	flangeQuench  flangeSpec
	flangeVentIn  flangeSpec
	flangeVentOut flangeSpec
	flangeColumn  flangeSpec
	flangeStack   flangeSpec
}

func deriveGeometry(d model.ScrubberDesign) assemblyGeometry {
	qWall := float64(d.Quench.WallThickness)
	if qWall <= 0 {
		qWall = 0.0020 // 14 ga 316L SS (~2.0 mm)
	}
	vWall := float64(d.Venturi.WallThickness)
	if vWall <= 0 {
		vWall = 0.0027 // 12 ga 316L SS (~2.7 mm)
	}
	cWall := float64(d.PackedBed.WallThickness)
	if cWall <= 0 {
		cWall = 0.00635 // 1/4" CPVC/PP
	}

	ductDia := float64(d.Envelope.DuctDiameter)
	if ductDia <= 0 {
		ductDia = float64(d.Input.DuctDiameter)
	}
	if ductDia <= 0 {
		ductDia = float64(d.Venturi.InletDiameter)
	}
	ductTotalLen := float64(d.Input.DuctLength)
	if ductTotalLen <= 0 {
		ductTotalLen = math.Max(1.0, 4.0*ductDia)
	}
	ductStubLen := clamp(ductTotalLen, 0.35, 1.0)

	qNozz := d.Quench.NozzleCount
	if qNozz <= 0 {
		qNozz = 3
	}

	colDia := float64(d.PackedBed.ColumnDiameter)
	colArea := 0.25 * math.Pi * colDia * colDia

	packVol := float64(d.PackedBed.PackingVolume)
	if packVol <= 0 {
		packVol = colArea * float64(d.PackedBed.PackingHeight)
	}

	sumpVol := float64(d.PackedBed.SumpVolume)
	var sumpH float64
	if sumpVol > 0 {
		sumpH = clamp(sumpVol/colArea, 0.30, 0.90)
	} else {
		sumpH = clamp(0.65*colDia, 0.30, 0.70)
		sumpVol = colArea * sumpH
	}

	demisterH := float64(d.PackedBed.DemisterHeight)
	if demisterH <= 0 {
		demisterH = 0.15
	}

	wheelDia := float64(d.Blower.WheelDiameter)
	stackDia := clamp(0.55*wheelDia, 0.35*colDia, 0.80*colDia)
	if stackDia < 0.10 {
		stackDia = math.Min(0.10, 0.80*colDia)
	}
	reducerH := math.Max(0.15, 0.50*(colDia-stackDia))
	stackRiserH := math.Max(0.25, 1.4*stackDia)
	blowerWidth := math.Max(0.14, 0.38*wheelDia)

	return assemblyGeometry{
		ductDia:      ductDia,
		ductStubLen:  ductStubLen,
		ductTotalLen: ductTotalLen,
		ductWall:     qWall,

		quenchDia:  float64(d.Quench.TowerDiameter),
		quenchH:    float64(d.Quench.TowerHeight),
		quenchWall: qWall,
		quenchNozz: qNozz,

		ventInDia:   float64(d.Venturi.InletDiameter),
		ventThDia:   float64(d.Venturi.ThroatDiameter),
		ventOutDia:  float64(d.Venturi.OutletDiameter),
		ventConvLen: float64(d.Venturi.ConvergingLength),
		ventThLen:   float64(d.Venturi.ThroatLength),
		ventDivLen:  float64(d.Venturi.DivergingLength),
		ventWall:    vWall,

		colDia:      colDia,
		colH:        float64(d.PackedBed.ColumnHeight),
		colWall:     cWall,
		packH:       float64(d.PackedBed.PackingHeight),
		packVol:     packVol,
		demisterH:   demisterH,
		sumpH:       sumpH,
		sumpVol:     sumpVol,
		stackDia:    stackDia,
		reducerH:    reducerH,
		stackRiserH: stackRiserH,

		wheelDia:    wheelDia,
		blowerWidth: blowerWidth,

		flangeDuct:    makeFlangeSpec(ductDia, qWall),
		flangeQuench:  makeFlangeSpec(float64(d.Quench.TowerDiameter), qWall),
		flangeVentIn:  makeFlangeSpec(float64(d.Venturi.InletDiameter), vWall),
		flangeVentOut: makeFlangeSpec(float64(d.Venturi.OutletDiameter), vWall),
		flangeColumn:  makeFlangeSpec(colDia, cWall),
		flangeStack:   makeFlangeSpec(stackDia, cWall),
	}
}
