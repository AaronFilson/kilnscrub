package cad

import (
	"math"
	"strings"
	"testing"

	"github.com/AaronFilson/kilnscrub/pkg/model"
	"github.com/AaronFilson/kilnscrub/pkg/scrubber"
)

func TestCADBranchCoverageAndFallbacks(t *testing.T) {
	base, err := scrubber.NewDesigner().Design(model.DefaultSaltKilnInput())
	if err != nil {
		t.Fatalf("Design: %v", err)
	}

	// 1. validateDesign invalid field paths.
	badDesigns := []func(*model.ScrubberDesign){
		func(d *model.ScrubberDesign) { d.Quench.TowerDiameter = 0 },
		func(d *model.ScrubberDesign) { d.Quench.TowerHeight = 0 },
		func(d *model.ScrubberDesign) { d.Quench.WallThickness = -1 },
		func(d *model.ScrubberDesign) { d.Quench.TotalSprayWaterGPM = -1 },
		func(d *model.ScrubberDesign) { d.Venturi.InletDiameter = 0 },
		func(d *model.ScrubberDesign) { d.Venturi.ThroatDiameter = 0 },
		func(d *model.ScrubberDesign) { d.Venturi.OutletDiameter = 0 },
		func(d *model.ScrubberDesign) { d.Venturi.ConvergingLength = 0 },
		func(d *model.ScrubberDesign) { d.Venturi.ThroatLength = 0 },
		func(d *model.ScrubberDesign) { d.Venturi.DivergingLength = 0 },
		func(d *model.ScrubberDesign) { d.Venturi.WallThickness = -1 },
		func(d *model.ScrubberDesign) { d.Venturi.ThroatDiameter = d.Venturi.InletDiameter + 0.1 },
		func(d *model.ScrubberDesign) { d.PackedBed.ColumnDiameter = 0 },
		func(d *model.ScrubberDesign) { d.PackedBed.ColumnHeight = 0 },
		func(d *model.ScrubberDesign) { d.PackedBed.WallThickness = -1 },
		func(d *model.ScrubberDesign) { d.PackedBed.PackingHeight = 0 },
		func(d *model.ScrubberDesign) { d.PackedBed.PackingHeight = d.PackedBed.ColumnHeight + 0.1 },
		func(d *model.ScrubberDesign) { d.PackedBed.DemisterHeight = -1 },
		func(d *model.ScrubberDesign) { d.PackedBed.PackingVolume = -1 },
		func(d *model.ScrubberDesign) { d.PackedBed.SumpVolume = -1 },
		func(d *model.ScrubberDesign) { d.PackedBed.RecirculationFlowGPM = -1 },
		func(d *model.ScrubberDesign) { d.PackedBed.NaHCO3ConsumptionKgPerHr = -1 },
		func(d *model.ScrubberDesign) { d.Blower.WheelDiameter = 0 },
		func(d *model.ScrubberDesign) { d.Blower.RecommendedMotorHP = -1 },
		func(d *model.ScrubberDesign) { d.Blower.FlowACFM = -1 },
		func(d *model.ScrubberDesign) { d.Blower.TotalStaticPressureInWC = -1 },
		func(d *model.ScrubberDesign) { d.Envelope.DuctDiameter = -1 },
		func(d *model.ScrubberDesign) { d.Input.DuctDiameter = -1 },
		func(d *model.ScrubberDesign) { d.Input.DuctLength = -1 },
	}
	for i, mutate := range badDesigns {
		d := base
		mutate(&d)
		if err := validateDesign(d); err == nil {
			t.Errorf("badDesigns[%d]: expected validateDesign error", i)
		}
	}

	// 2. deriveGeometry & GenerateBOM fallback branches (empty materials, zero optional fields, >0.60 m flange).
	fb := base
	fb.Quench.Material = ""
	fb.Quench.WallThickness = 0
	fb.Quench.NozzleCount = 0
	fb.Venturi.Material = ""
	fb.Venturi.WallThickness = 0
	fb.PackedBed.Material = ""
	fb.PackedBed.WallThickness = 0
	fb.PackedBed.PackingSpec = ""
	fb.Input.PackingName = ""
	fb.PackedBed.PackingVolume = 0
	fb.PackedBed.SumpVolume = 0
	fb.PackedBed.DemisterHeight = 0
	fb.PackedBed.RecirculationFlowGPM = 0
	fb.PackedBed.ColumnDiameter = 0.70 // >0.60 m -> 16 bolts flange
	fb.Blower.Material = ""
	fb.Blower.RecommendedMotorHP = 0
	fb.Blower.WheelDiameter = 0.12 // small stackDia clamp
	fb.Envelope.DuctDiameter = 0
	fb.Input.DuctDiameter = 0
	fb.Input.DuctLength = 0
	fb.Input.KilnVolumeCuFt = 0

	bom, err := GenerateBOM(fb)
	if err != nil {
		t.Fatalf("GenerateBOM fallbacks failed: %v", err)
	}
	if bom.TotalCostUSD <= 0 || len(bom.Items) == 0 {
		t.Fatalf("unexpected fallback BOM: %+v", bom)
	}
	// FormatBOMMarkdown with empty title.
	bomNoTitle := bom
	bomNoTitle.Title = ""
	mdBytes, err := FormatBOMMarkdown(bomNoTitle)
	if err != nil || !strings.Contains(string(mdBytes), "Kiln Exhaust Scrubber Bill of Materials") {
		t.Errorf("FormatBOMMarkdown empty title fallback failed: err=%v", err)
	}
	if _, err := ExportSTL(fb); err != nil {
		t.Errorf("ExportSTL fallback design failed: %v", err)
	}
	if _, err := ExportDXF(fb); err != nil {
		t.Errorf("ExportDXF fallback design failed: %v", err)
	}

	// 3. stlMesh primitive edge cases (cone tips r0>0/r1=0 & r0=0/r1>0, degenerate/NaN tris, invalid params).
	var m stlMesh
	m.addTri(vec3{math.NaN(), 0, 0}, vec3{1, 0, 0}, vec3{0, 1, 0})
	m.addTri(vec3{0, 0, 0}, vec3{1e-9, 0, 0}, vec3{0, 1e-9, 0})
	if _, _, _, ok := orthonormalBasis(vec3{0, 0, 0}); ok {
		t.Error("expected orthonormalBasis(zero) false")
	}
	m.addFrustum(vec3{0, 0, 0}, vec3{0, 0, 10}, 5, 0, 12, true, false) // cone tip at p1
	m.addFrustum(vec3{0, 0, 0}, vec3{0, 0, 10}, 0, 5, 12, false, true) // cone tip at p0
	m.addFrustum(vec3{0, 0, 0}, vec3{0, 0, 10}, 0, 0, 12, false, false)
	m.addFrustum(vec3{0, 0, 0}, vec3{0, 0, 0}, 5, 5, 12, true, true)
	m.addFlangeRing(vec3{}, vec3{}, 10, 2, 12)
	m.addBox(vec3{}, 0, 1, 1)
	m.addVoluteHousing(vec3{}, 10, 5, 4, 12)                         // rEnd <= rStart
	m.addPipeElbow(vec3{}, vec3{1, 0, 0}, vec3{1, 0, 0}, 2, 5, 4, 8) // parallel dirs
	m.addPipeElbow(vec3{}, vec3{1, 0, 0}, vec3{0, 1, 0}, 2, 1, 4, 8) // bendR <= pipeR
	if len(m.tris) == 0 {
		t.Error("expected cone-tip frustums to emit valid triangles")
	}
}
