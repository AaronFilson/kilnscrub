package cad_test

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/AaronFilson/kilnscrub/pkg/cad"
	"github.com/AaronFilson/kilnscrub/pkg/model"
)

// sampleSaltKilnDesign builds a realistic self-contained ScrubberDesign fixture
// representing a 24 cu.ft. wood/salt kiln (2400 F peak, 2 lbs NaCl/hr) without
// importing pkg/scrubber.
func sampleSaltKilnDesign() model.ScrubberDesign {
	in := model.DefaultSaltKilnInput()
	return model.ScrubberDesign{
		Input: in,
		Quench: model.QuenchResult{
			InletTemp:          model.FahrenheitToKelvin(2400.0),
			AdiabaticSatTemp:   model.FahrenheitToKelvin(164.0),
			OutletTemp:         model.FahrenheitToKelvin(165.0),
			DewPoint:           model.FahrenheitToKelvin(163.5),
			EvaporatedWaterGPM: 0.42,
			TotalSprayWaterGPM: 1.25,
			OversprayFactor:    3.0,
			TowerDiameter:      0.45,
			TowerHeight:        1.60,
			WallThickness:      0.0020,
			GasVelocity:        1.8,
			ResidenceTimeSec:   0.89,
			NozzleCount:        3,
			PressureDrop:       model.InWCToPascal(0.35),
			PressureDropInWC:   0.35,
			Material:           "316L Stainless Steel",
		},
		Venturi: model.VenturiResult{
			InletDiameter:             0.22,
			ThroatDiameter:            0.085,
			OutletDiameter:            0.24,
			ConvergingLength:          0.26,
			ThroatLength:              0.12,
			DivergingLength:           0.62,
			WallThickness:             0.0027,
			ThroatVelocity:            70.0,
			LGRatio:                   0.0012,
			LiquidFlowGPM:             1.45,
			DropletDiameter:           model.MicronsToMeters(50.0),
			CutDiameterD50:            model.MicronsToMeters(0.38),
			CutDiameterD50Microns:     0.38,
			WoodAshCutDiameterD50:     model.MicronsToMeters(0.35),
			OverallWoodAshEfficiency:  0.94,
			OverallNaClFumeEfficiency: 0.89,
			PressureDrop:              model.InWCToPascal(14.5),
			PressureDropInWC:          14.5,
			Material:                  "316L Stainless Steel",
		},
		PackedBed: model.PackedBedResult{
			ColumnDiameter:           0.55,
			ColumnHeight:             2.10,
			WallThickness:            0.00635,
			FloodingVelocity:         2.25,
			SuperficialGasVelocity:   1.48,
			PercentOfFlood:           65.8,
			PackingHeight:            1.25,
			NTU:                      4.61,
			HTU:                      0.27,
			PackingSpec:              `1" PP Pall Rings`,
			PackingDiameter:          model.InchesToMeters(1.0),
			PackingVoidage:           0.90,
			PackingVolume:            0.297,
			HClInletPPM:              850.0,
			HClOutletPPM:             8.5,
			HClRemovalEfficiency:     0.99,
			NaHCO3ConsumptionKgPerHr: 1.31,
			RecirculationFlowGPM:     12.0,
			SumpPH:                   8.0,
			SumpVolume:               0.14,
			BedPressureDrop:          model.InWCToPascal(1.15),
			DemisterPressureDrop:     model.InWCToPascal(0.30),
			DemisterHeight:           0.15,
			TotalPressureDrop:        model.InWCToPascal(1.45),
			TotalPressureDropInWC:    1.45,
			Material:                 "CPVC / Polypropylene",
		},
		Blower: model.BlowerResult{
			TargetKilnDraftInWC:     -0.05,
			TotalStaticPressure:     model.InWCToPascal(16.85),
			TotalStaticPressureInWC: 16.85,
			Flow:                    model.CFMToCubicMetersPerSec(750.0),
			FlowACFM:                750.0,
			Efficiency:              0.62,
			ShaftPowerHP:            3.21,
			RecommendedMotorHP:      5.0,
			WheelDiameter:           0.46,
			OperatingRPM:            3450.0,
			Material:                "FRP / Coated Steel",
		},
		Envelope: model.EnvelopeDimensions{
			Length:       3.2,
			Width:        1.1,
			Height:       3.0,
			DuctDiameter: 0.25,
		},
	}
}

func TestExporterImplementsInterface(t *testing.T) {
	var exp model.CADExporter = cad.NewExporter()
	if exp == nil {
		t.Fatal("NewExporter returned nil")
	}
}

func TestExportSTL_BinaryIntegrityAndGeometry(t *testing.T) {
	design := sampleSaltKilnDesign()
	exp := cad.NewExporter()

	stlBytes, err := exp.ExportSTL(design)
	if err != nil {
		t.Fatalf("ExportSTL failed: %v", err)
	}
	if len(stlBytes) < 84 {
		t.Fatalf("STL output too short: %d bytes", len(stlBytes))
	}

	// Binary STL must NOT begin with "solid " to avoid ASCII misdetection.
	header := string(stlBytes[:80])
	if strings.HasPrefix(header, "solid ") {
		t.Errorf("binary STL header should not start with 'solid ': %q", header)
	}
	if !strings.Contains(header, "kilnscrub") {
		t.Errorf("expected kilnscrub identifier in STL header, got %q", header)
	}

	triCount := binary.LittleEndian.Uint32(stlBytes[80:84])
	if triCount < 400 {
		t.Fatalf("expected rich multi-stage mesh (>= 400 triangles), got %d", triCount)
	}

	expectedLen := 84 + 50*int(triCount)
	if len(stlBytes) != expectedLen {
		t.Fatalf("STL byte length mismatch: got %d, want %d (84 + 50*%d)", len(stlBytes), expectedLen, triCount)
	}

	minX, minY, minZ := math.Inf(1), math.Inf(1), math.Inf(1)
	maxX, maxY, maxZ := math.Inf(-1), math.Inf(-1), math.Inf(-1)
	var signedVol6 float64

	for i := 0; i < int(triCount); i++ {
		off := 84 + i*50
		readF32 := func(pos int) float64 {
			return float64(math.Float32frombits(binary.LittleEndian.Uint32(stlBytes[off+pos : off+pos+4])))
		}
		nx, ny, nz := readF32(0), readF32(4), readF32(8)
		x1, y1, z1 := readF32(12), readF32(16), readF32(20)
		x2, y2, z2 := readF32(24), readF32(28), readF32(32)
		x3, y3, z3 := readF32(36), readF32(40), readF32(44)
		attr := binary.LittleEndian.Uint16(stlBytes[off+48 : off+50])
		if attr != 0 {
			t.Fatalf("triangle %d: unexpected non-zero attribute byte count %d", i, attr)
		}

		vals := []float64{nx, ny, nz, x1, y1, z1, x2, y2, z2, x3, y3, z3}
		for _, v := range vals {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				t.Fatalf("triangle %d: non-finite float value %v", i, v)
			}
		}

		// Unit-length normal check.
		nLen := math.Sqrt(nx*nx + ny*ny + nz*nz)
		if math.Abs(nLen-1.0) > 1e-4 {
			t.Fatalf("triangle %d: normal length = %v, want 1.0", i, nLen)
		}

		// Non-degenerate triangle & right-hand rule normal consistency check.
		e1x, e1y, e1z := x2-x1, y2-y1, z2-z1
		e2x, e2y, e2z := x3-x1, y3-y1, z3-z1
		cx := e1y*e2z - e1z*e2y
		cy := e1z*e2x - e1x*e2z
		cz := e1x*e2y - e1y*e2x
		cLen := math.Sqrt(cx*cx + cy*cy + cz*cz)
		if cLen < 1e-6 {
			t.Fatalf("triangle %d: degenerate triangle (cross magnitude=%g)", i, cLen)
		}
		dot := (cx*nx + cy*ny + cz*nz) / cLen
		if dot < 0.999 {
			t.Fatalf("triangle %d: stored normal disagrees with vertex winding (dot=%v)", i, dot)
		}

		// Accumulate 6 * signed tetrahedron volume to confirm outward orientation.
		signedVol6 += x1*(y2*z3-z2*y3) - y1*(x2*z3-z2*x3) + z1*(x2*y3-y2*x3)

		for _, pt := range [][3]float64{{x1, y1, z1}, {x2, y2, z2}, {x3, y3, z3}} {
			minX, maxX = math.Min(minX, pt[0]), math.Max(maxX, pt[0])
			minY, maxY = math.Min(minY, pt[1]), math.Max(maxY, pt[1])
			minZ, maxZ = math.Min(minZ, pt[2]), math.Max(maxZ, pt[2])
		}
	}

	if signedVol6 <= 0 {
		t.Errorf("expected positive signed volume for outward-oriented closed solids, got %.3f", signedVol6/6.0)
	}

	// Bounding box sanity (units: millimeters).
	spanX := maxX - minX
	spanY := maxY - minY
	spanZ := maxZ - minZ

	if math.Abs(minX) > 1e-2 {
		t.Errorf("expected normalized minX ~ 0 mm, got %.3f", minX)
	}
	if minZ < -1e-2 || minZ > 1e-2 {
		t.Errorf("expected assembly resting on floor Z=0 mm, got minZ=%.3f", minZ)
	}
	if spanX < 1800.0 || spanX > 8000.0 {
		t.Errorf("unexpected assembly X span %.1f mm", spanX)
	}
	if spanY < 450.0 || spanY > 3000.0 {
		t.Errorf("unexpected assembly Y span %.1f mm", spanY)
	}
	// Total height must exceed packed-bed column + sump + reducer + stack.
	minExpectedH := (float64(design.PackedBed.ColumnHeight) + 0.30) * 1000.0
	if spanZ < minExpectedH || spanZ > 6000.0 {
		t.Errorf("unexpected assembly Z span %.1f mm (want >= %.1f mm)", spanZ, minExpectedH)
	}
}

func TestComputeFrustumDevelopment_ExactMath(t *testing.T) {
	// Venturi converging cone: D1=220mm (R1=110), D2=85mm (R2=42.5), L_axial=260mm.
	dev, err := cad.ComputeFrustumDevelopment(220.0, 85.0, 260.0)
	if err != nil {
		t.Fatalf("ComputeFrustumDevelopment failed: %v", err)
	}
	if dev.IsCylinder {
		t.Fatal("expected frustum development, got cylinder")
	}

	r1, r2, L := 42.5, 110.0, 260.0
	wantSlant := math.Hypot(L, r2-r1)
	wantInner := r1 * wantSlant / (r2 - r1)
	wantOuter := r2 * wantSlant / (r2 - r1)
	wantTheta := 2.0 * math.Pi * (r2 - r1) / wantSlant

	const tol = 1e-9
	if math.Abs(dev.SlantLength-wantSlant) > tol {
		t.Errorf("SlantLength = %.9f, want %.9f", dev.SlantLength, wantSlant)
	}
	if math.Abs(dev.InnerRadius-wantInner) > tol {
		t.Errorf("InnerRadius = %.9f, want %.9f", dev.InnerRadius, wantInner)
	}
	if math.Abs(dev.OuterRadius-wantOuter) > tol {
		t.Errorf("OuterRadius = %.9f, want %.9f", dev.OuterRadius, wantOuter)
	}
	if math.Abs(dev.ThetaRad-wantTheta) > tol {
		t.Errorf("ThetaRad = %.9f, want %.9f", dev.ThetaRad, wantTheta)
	}

	// Verify exact circumference preservation at both ends and radial width == L_slant.
	if math.Abs(dev.InnerRadius*dev.ThetaRad-2.0*math.Pi*r1) > 1e-8 {
		t.Errorf("inner arc length %.6f != small circumference %.6f", dev.InnerRadius*dev.ThetaRad, 2*math.Pi*r1)
	}
	if math.Abs(dev.OuterRadius*dev.ThetaRad-2.0*math.Pi*r2) > 1e-8 {
		t.Errorf("outer arc length %.6f != large circumference %.6f", dev.OuterRadius*dev.ThetaRad, 2*math.Pi*r2)
	}
	if math.Abs((dev.OuterRadius-dev.InnerRadius)-dev.SlantLength) > 1e-8 {
		t.Errorf("R_outer - R_inner = %.6f != L_slant %.6f", dev.OuterRadius-dev.InnerRadius, dev.SlantLength)
	}

	// Equal diameters should gracefully unfold as a cylinder rectangle.
	cyl, err := cad.ComputeFrustumDevelopment(200.0, 200.0, 300.0)
	if err != nil {
		t.Fatalf("equal-diameter development failed: %v", err)
	}
	if !cyl.IsCylinder {
		t.Fatal("expected IsCylinder=true for equal diameters")
	}
	if math.Abs(cyl.FlatWidth-math.Pi*200.0) > 1e-8 {
		t.Errorf("FlatWidth = %.6f, want %.6f", cyl.FlatWidth, math.Pi*200.0)
	}
}

type parsedArc struct {
	r, startDeg, endDeg float64
}

func parseDXFArcs(dxf string) []parsedArc {
	lines := strings.Split(strings.ReplaceAll(dxf, "\r\n", "\n"), "\n")
	var arcs []parsedArc
	for i := 0; i+1 < len(lines); i += 2 {
		code := strings.TrimSpace(lines[i])
		val := strings.TrimSpace(lines[i+1])
		if code == "0" && val == "ARC" {
			var a parsedArc
			for j := i + 2; j+1 < len(lines); j += 2 {
				c := strings.TrimSpace(lines[j])
				v := strings.TrimSpace(lines[j+1])
				if c == "0" {
					break
				}
				switch c {
				case "40":
					a.r, _ = strconv.ParseFloat(v, 64)
				case "50":
					a.startDeg, _ = strconv.ParseFloat(v, 64)
				case "51":
					a.endDeg, _ = strconv.ParseFloat(v, 64)
				}
			}
			arcs = append(arcs, a)
		}
	}
	return arcs
}

func TestExportDXF_StructureAndExactConeEntities(t *testing.T) {
	design := sampleSaltKilnDesign()
	dxfBytes, err := cad.ExportDXF(design)
	if err != nil {
		t.Fatalf("ExportDXF failed: %v", err)
	}
	dxf := string(dxfBytes)

	// Verify required R12 DXF structural tokens.
	for _, token := range []string{
		"SECTION", "HEADER", "AC1009", "TABLES", "LTYPE", "LAYER",
		"ENTITIES", "LINE", "ARC", "CIRCLE", "TEXT", "ENDSEC", "EOF",
		"STAGE 1: QUENCH TOWER SHELL",
		"STAGE 2: VENTURI THROAT SHELL",
		"STAGE 2: VENTURI CONVERGING CONE",
		"STAGE 2: VENTURI DIVERGING CONE",
		"STAGE 3: PACKED-BED COLUMN SHELL",
		"STAGE 3: COLUMN-TO-STACK REDUCER CONE",
	} {
		if !strings.Contains(dxf, token) {
			t.Errorf("DXF missing expected token %q", token)
		}
	}
	if !strings.HasSuffix(strings.TrimSpace(dxf), "EOF") {
		t.Errorf("DXF does not terminate with EOF")
	}
	if strings.Contains(dxf, "NaN") || strings.Contains(dxf, "Inf") {
		t.Errorf("DXF contains non-finite numeric values")
	}

	// Verify exact frustum ARC radii and sweep angles appear in the DXF ENTITIES.
	arcs := parseDXFArcs(dxf)
	// 3 frustums (venturi converging, venturi diverging, column reducer) * 2 arcs each = 6 arcs.
	if len(arcs) < 6 {
		t.Fatalf("expected at least 6 ARC entities for 3 cone developments, got %d", len(arcs))
	}

	wantConv, err := cad.ComputeFrustumDevelopment(
		float64(design.Venturi.InletDiameter)*1000.0,
		float64(design.Venturi.ThroatDiameter)*1000.0,
		float64(design.Venturi.ConvergingLength)*1000.0,
	)
	if err != nil {
		t.Fatal(err)
	}
	wantDiv, err := cad.ComputeFrustumDevelopment(
		float64(design.Venturi.ThroatDiameter)*1000.0,
		float64(design.Venturi.OutletDiameter)*1000.0,
		float64(design.Venturi.DivergingLength)*1000.0,
	)
	if err != nil {
		t.Fatal(err)
	}

	hasArc := func(wantR, wantSweepDeg float64) bool {
		for _, a := range arcs {
			sweep := a.endDeg - a.startDeg
			if math.Abs(a.r-wantR) < 1e-2 && math.Abs(sweep-wantSweepDeg) < 1e-2 {
				return true
			}
		}
		return false
	}

	if !hasArc(wantConv.InnerRadius, wantConv.ThetaDeg) || !hasArc(wantConv.OuterRadius, wantConv.ThetaDeg) {
		t.Errorf("DXF missing exact converging cone arcs (R_in=%.3f, R_out=%.3f, theta=%.3f deg); parsed=%+v",
			wantConv.InnerRadius, wantConv.OuterRadius, wantConv.ThetaDeg, arcs)
	}
	if !hasArc(wantDiv.InnerRadius, wantDiv.ThetaDeg) || !hasArc(wantDiv.OuterRadius, wantDiv.ThetaDeg) {
		t.Errorf("DXF missing exact diverging cone arcs (R_in=%.3f, R_out=%.3f, theta=%.3f deg); parsed=%+v",
			wantDiv.InnerRadius, wantDiv.OuterRadius, wantDiv.ThetaDeg, arcs)
	}
}

func TestGenerateBOM_TotalsCSVAndMarkdown(t *testing.T) {
	design := sampleSaltKilnDesign()
	exp := cad.NewExporter()

	bom, err := exp.GenerateBOM(design)
	if err != nil {
		t.Fatalf("GenerateBOM failed: %v", err)
	}
	if len(bom.Items) < 12 {
		t.Fatalf("expected comprehensive BOM (>=12 items), got %d", len(bom.Items))
	}

	// Verify RecalculateTotals consistency and total-weight contract.
	var sumCost, sumWeight float64
	foundPacking, foundDemister, foundBlower, foundDamper, foundNozzles, foundDosing := false, false, false, false, false, false
	for _, it := range bom.Items {
		if it.Quantity <= 0 {
			t.Errorf("item %q has non-positive quantity %v", it.Name, it.Quantity)
		}
		if it.WeightKg <= 0 {
			t.Errorf("item %q has non-positive WeightKg %v", it.Name, it.WeightKg)
		}
		if it.UnitCostUSD > 0 {
			wantTotal := it.Quantity * it.UnitCostUSD
			if math.Abs(it.TotalCostUSD-wantTotal) > 1e-6 {
				t.Errorf("item %q: TotalCostUSD=%.4f != Qty*UnitCost=%.4f", it.Name, it.TotalCostUSD, wantTotal)
			}
		}
		sumCost += it.TotalCostUSD
		sumWeight += it.WeightKg

		nameLower := strings.ToLower(it.Name)
		switch {
		case strings.Contains(nameLower, "tower packing"):
			foundPacking = true
			if math.Abs(it.Quantity-float64(design.PackedBed.PackingVolume)) > 0.002 {
				t.Errorf("packing quantity %.4f does not match PackingVolume %.4f", it.Quantity, design.PackedBed.PackingVolume)
			}
		case strings.Contains(nameLower, "demister") || strings.Contains(nameLower, "mist"):
			foundDemister = true
		case strings.Contains(nameLower, "blower"):
			foundBlower = true
		case strings.Contains(nameLower, "damper"):
			foundDamper = true
		case strings.Contains(nameLower, "nozzle"):
			foundNozzles = true
		case strings.Contains(nameLower, "nahco3") || strings.Contains(nameLower, "dosing"):
			foundDosing = true
		}
	}

	if !foundPacking || !foundDemister || !foundBlower || !foundDamper || !foundNozzles || !foundDosing {
		t.Errorf("BOM missing expected key hardware items: packing=%v demister=%v blower=%v damper=%v nozzles=%v dosing=%v",
			foundPacking, foundDemister, foundBlower, foundDamper, foundNozzles, foundDosing)
	}
	if math.Abs(bom.TotalCostUSD-sumCost) > 1e-6 {
		t.Errorf("BOM TotalCostUSD=%.4f != sum %.4f", bom.TotalCostUSD, sumCost)
	}
	if math.Abs(bom.TotalWeightKg-sumWeight) > 1e-6 {
		t.Errorf("BOM TotalWeightKg=%.4f != sum %.4f", bom.TotalWeightKg, sumWeight)
	}

	// Re-calling RecalculateTotals must be idempotent.
	origCost, origWeight := bom.TotalCostUSD, bom.TotalWeightKg
	bom.RecalculateTotals()
	if math.Abs(bom.TotalCostUSD-origCost) > 1e-9 || math.Abs(bom.TotalWeightKg-origWeight) > 1e-9 {
		t.Errorf("RecalculateTotals not idempotent: cost %.4f->%.4f, weight %.4f->%.4f",
			origCost, bom.TotalCostUSD, origWeight, bom.TotalWeightKg)
	}

	// Verify CSV formatting & parsing via encoding/csv.
	csvBytes, err := exp.FormatBOMCSV(bom)
	if err != nil {
		t.Fatalf("FormatBOMCSV failed: %v", err)
	}
	records, err := csv.NewReader(bytes.NewReader(csvBytes)).ReadAll()
	if err != nil {
		t.Fatalf("CSV parse failed: %v", err)
	}
	if len(records) != len(bom.Items)+2 {
		t.Fatalf("CSV row count = %d, want %d (header + %d items + TOTAL)", len(records), len(bom.Items)+2, len(bom.Items))
	}
	if records[0][0] != "Category" || records[0][8] != "WeightKg" {
		t.Errorf("unexpected CSV header: %v", records[0])
	}
	lastRow := records[len(records)-1]
	if lastRow[0] != "TOTAL" {
		t.Errorf("expected TOTAL summary row at end of CSV, got %v", lastRow)
	}
	csvCost, err := strconv.ParseFloat(lastRow[7], 64)
	if err != nil || math.Abs(csvCost-bom.TotalCostUSD) > 0.02 {
		t.Errorf("CSV total cost %q disagrees with BOM TotalCostUSD %.2f", lastRow[7], bom.TotalCostUSD)
	}
	csvWeight, err := strconv.ParseFloat(lastRow[8], 64)
	if err != nil || math.Abs(csvWeight-bom.TotalWeightKg) > 0.02 {
		t.Errorf("CSV total weight %q disagrees with BOM TotalWeightKg %.2f", lastRow[8], bom.TotalWeightKg)
	}

	// Verify Markdown formatting.
	mdBytes, err := exp.FormatBOMMarkdown(bom)
	if err != nil {
		t.Fatalf("FormatBOMMarkdown failed: %v", err)
	}
	md := string(mdBytes)
	for _, want := range []string{
		"# ",
		"| Category | Item | Material |",
		"| **TOTAL** |",
		`1" PP Pall Rings`,
		"316L Stainless Steel",
		"CPVC / Polypropylene",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("Markdown BOM missing expected text %q", want)
		}
	}
}

func TestInvalidDesignErrors(t *testing.T) {
	exp := cad.NewExporter()

	cases := []struct {
		name   string
		mutate func(*model.ScrubberDesign)
	}{
		{"empty_design", func(d *model.ScrubberDesign) { *d = model.ScrubberDesign{} }},
		{"zero_quench_diameter", func(d *model.ScrubberDesign) { d.Quench.TowerDiameter = 0 }},
		{"negative_quench_height", func(d *model.ScrubberDesign) { d.Quench.TowerHeight = -1.0 }},
		{"nan_venturi_throat", func(d *model.ScrubberDesign) { d.Venturi.ThroatDiameter = model.Meters(math.NaN()) }},
		{"inf_venturi_length", func(d *model.ScrubberDesign) { d.Venturi.DivergingLength = model.Meters(math.Inf(1)) }},
		{"throat_exceeds_inlet", func(d *model.ScrubberDesign) { d.Venturi.ThroatDiameter = d.Venturi.InletDiameter + 0.05 }},
		{"zero_column_diameter", func(d *model.ScrubberDesign) { d.PackedBed.ColumnDiameter = 0 }},
		{"packing_exceeds_column", func(d *model.ScrubberDesign) { d.PackedBed.PackingHeight = d.PackedBed.ColumnHeight + 0.5 }},
		{"negative_packing_volume", func(d *model.ScrubberDesign) { d.PackedBed.PackingVolume = -0.1 }},
		{"zero_blower_wheel", func(d *model.ScrubberDesign) { d.Blower.WheelDiameter = 0 }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := sampleSaltKilnDesign()
			tc.mutate(&d)

			if _, err := exp.ExportSTL(d); err == nil {
				t.Errorf("ExportSTL(%s): expected error, got nil", tc.name)
			}
			if _, err := exp.ExportDXF(d); err == nil {
				t.Errorf("ExportDXF(%s): expected error, got nil", tc.name)
			}
			if _, err := exp.GenerateBOM(d); err == nil {
				t.Errorf("GenerateBOM(%s): expected error, got nil", tc.name)
			}
		})
	}
}

func TestSTLVoluteEdgeManifoldAndR12DXFCompliance(t *testing.T) {
	exp := cad.NewExporter()
	d := sampleSaltKilnDesign()

	// Verify DXF header omits R2000-only $INSUNITS when claiming AC1009 (R12).
	dxfBytes, err := exp.ExportDXF(d)
	if err != nil {
		t.Fatalf("ExportDXF: %v", err)
	}
	if strings.Contains(string(dxfBytes), "$INSUNITS") {
		t.Errorf("R12 AC1009 DXF must not contain R2000-only $INSUNITS header variable")
	}

	// Verify the entire exported binary STL mesh is edge-manifold (zero unmatched half-edges / T-junctions).
	stlBytes, err := exp.ExportSTL(d)
	if err != nil {
		t.Fatalf("ExportSTL: %v", err)
	}
	if len(stlBytes) < 84 {
		t.Fatalf("STL too short: %d bytes", len(stlBytes))
	}
	triCount := int(binary.LittleEndian.Uint32(stlBytes[80:84]))
	type qpt struct{ x, y, z int64 }
	readPt := func(off int) qpt {
		fx := math.Float32frombits(binary.LittleEndian.Uint32(stlBytes[off : off+4]))
		fy := math.Float32frombits(binary.LittleEndian.Uint32(stlBytes[off+4 : off+8]))
		fz := math.Float32frombits(binary.LittleEndian.Uint32(stlBytes[off+8 : off+12]))
		return qpt{int64(math.Round(float64(fx) * 1e2)), int64(math.Round(float64(fy) * 1e2)), int64(math.Round(float64(fz) * 1e2))}
	}
	type edge struct{ a, b qpt }
	edges := make(map[edge]int, triCount*3)
	for i := 0; i < triCount; i++ {
		base := 84 + i*50 + 12
		p0 := readPt(base)
		p1 := readPt(base + 12)
		p2 := readPt(base + 24)
		edges[edge{p0, p1}]++
		edges[edge{p1, p2}]++
		edges[edge{p2, p0}]++
	}
	unmatched := 0
	for e, cnt := range edges {
		if cnt != edges[edge{e.b, e.a}] {
			unmatched++
		}
	}
	if unmatched != 0 {
		t.Errorf("exported binary STL has %d unbalanced/non-manifold half-edges, want 0", unmatched)
	}
}
