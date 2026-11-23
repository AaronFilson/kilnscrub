package cad

import (
	"bytes"
	"fmt"
	"math"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

const (
	layerCutOutline  = "CUT_OUTLINE"
	layerBendRoll    = "BEND_ROLL"
	layerCenterlines = "CENTERLINES"
	layerAnnotations = "ANNOTATIONS"
	layerFlanges     = "FLANGES"
)

// FrustumDevelopment holds the exact flat-pattern annular sector geometry for unfolding
// a right circular frustum (truncated cone) into 2D sheet metal.
//
// For radii R1 < R2 and axial length L_axial:
//   - L_slant = sqrt(L_axial^2 + (R2 - R1)^2)
//   - R_inner = R1 * L_slant / (R2 - R1)
//   - R_outer = R2 * L_slant / (R2 - R1) = R_inner + L_slant
//   - ThetaRad = 2 * pi * (R2 - R1) / L_slant
type FrustumDevelopment struct {
	SmallRadius float64
	LargeRadius float64
	AxialLength float64
	SlantLength float64
	InnerRadius float64 // apex-to-small-end slant radius
	OuterRadius float64 // apex-to-large-end slant radius
	ThetaRad    float64 // annular sector sweep angle in radians
	ThetaDeg    float64 // annular sector sweep angle in degrees
	IsCylinder  bool    // true when |R2 - R1| is negligible
	FlatWidth   float64 // circumference when IsCylinder == true
}

// ComputeFrustumDevelopment computes exact unfolded sheet-metal flat-pattern parameters
// for a right circular frustum with end diameters d1, d2 and axial length axialLength
// (all in consistent length units, e.g., millimeters).
func ComputeFrustumDevelopment(d1, d2, axialLength float64) (FrustumDevelopment, error) {
	if err := checkPosFinite("d1", d1); err != nil {
		return FrustumDevelopment{}, err
	}
	if err := checkPosFinite("d2", d2); err != nil {
		return FrustumDevelopment{}, err
	}
	if err := checkPosFinite("axialLength", axialLength); err != nil {
		return FrustumDevelopment{}, err
	}

	r1 := 0.5 * math.Min(d1, d2)
	r2 := 0.5 * math.Max(d1, d2)
	dr := r2 - r1

	// Degenerate frustum (equal end diameters) unfolds as a rectangular cylinder sheet.
	if dr <= 1e-9*math.Max(1.0, r2) {
		return FrustumDevelopment{
			SmallRadius: r1,
			LargeRadius: r2,
			AxialLength: axialLength,
			SlantLength: axialLength,
			IsCylinder:  true,
			FlatWidth:   2.0 * math.Pi * r1,
		}, nil
	}

	slant := math.Hypot(axialLength, dr)
	rInner := r1 * slant / dr
	rOuter := r2 * slant / dr
	thetaRad := 2.0 * math.Pi * dr / slant
	thetaDeg := thetaRad * 180.0 / math.Pi

	return FrustumDevelopment{
		SmallRadius: r1,
		LargeRadius: r2,
		AxialLength: axialLength,
		SlantLength: slant,
		InnerRadius: rInner,
		OuterRadius: rOuter,
		ThetaRad:    thetaRad,
		ThetaDeg:    thetaDeg,
		IsCylinder:  false,
	}, nil
}

type dxfBuilder struct {
	entities bytes.Buffer
	minX     float64
	minY     float64
	maxX     float64
	maxY     float64
	hasPts   bool
}

func newDXFBuilder() *dxfBuilder {
	return &dxfBuilder{
		minX: math.Inf(1),
		minY: math.Inf(1),
		maxX: math.Inf(-1),
		maxY: math.Inf(-1),
	}
}

func (b *dxfBuilder) trackPoint(x, y float64) {
	if !isFinite(x) || !isFinite(y) {
		return
	}
	if x < b.minX {
		b.minX = x
	}
	if y < b.minY {
		b.minY = y
	}
	if x > b.maxX {
		b.maxX = x
	}
	if y > b.maxY {
		b.maxY = y
	}
	b.hasPts = true
}

func (b *dxfBuilder) addLine(layer string, x1, y1, x2, y2 float64) {
	b.trackPoint(x1, y1)
	b.trackPoint(x2, y2)
	fmt.Fprintf(&b.entities,
		"  0\nLINE\n  8\n%s\n 10\n%.4f\n 20\n%.4f\n 30\n0.0000\n 11\n%.4f\n 21\n%.4f\n 31\n0.0000\n",
		layer, x1, y1, x2, y2)
}

func (b *dxfBuilder) addCircle(layer string, cx, cy, r float64) {
	b.trackPoint(cx-r, cy-r)
	b.trackPoint(cx+r, cy+r)
	fmt.Fprintf(&b.entities,
		"  0\nCIRCLE\n  8\n%s\n 10\n%.4f\n 20\n%.4f\n 30\n0.0000\n 40\n%.4f\n",
		layer, cx, cy, r)
}

func (b *dxfBuilder) addArc(layer string, cx, cy, r, startDeg, endDeg float64) {
	// Track arc endpoints and any cardinal angles covered by [startDeg, endDeg].
	angles := []float64{startDeg, endDeg}
	for _, card := range []float64{0, 90, 180, 270, 360} {
		if card >= startDeg && card <= endDeg {
			angles = append(angles, card)
		}
	}
	for _, deg := range angles {
		rad := deg * math.Pi / 180.0
		b.trackPoint(cx+r*math.Cos(rad), cy+r*math.Sin(rad))
	}
	fmt.Fprintf(&b.entities,
		"  0\nARC\n  8\n%s\n 10\n%.4f\n 20\n%.4f\n 30\n0.0000\n 40\n%.4f\n 50\n%.4f\n 51\n%.4f\n",
		layer, cx, cy, r, startDeg, endDeg)
}

func (b *dxfBuilder) addText(layer string, x, y, height float64, text string) {
	b.trackPoint(x, y)
	b.trackPoint(x+float64(len(text))*height*0.65, y+height)
	fmt.Fprintf(&b.entities,
		"  0\nTEXT\n  8\n%s\n 10\n%.4f\n 20\n%.4f\n 30\n0.0000\n 40\n%.4f\n  1\n%s\n",
		layer, x, y, height, text)
}

// addRect draws a rolled cylinder flat pattern rectangle with seam centerline and annotation.
func (b *dxfBuilder) addRolledCylinder(x, y, width, height float64, title, subtitle string, refLinesY []float64) {
	b.addLine(layerCutOutline, x, y, x+width, y)
	b.addLine(layerCutOutline, x+width, y, x+width, y+height)
	b.addLine(layerCutOutline, x+width, y+height, x, y+height)
	b.addLine(layerCutOutline, x, y+height, x, y)

	// Centerline along rolling midpoint.
	b.addLine(layerCenterlines, x+0.5*width, y-20.0, x+0.5*width, y+height+20.0)

	// Optional horizontal zone reference lines (e.g. packing / demister heights).
	for _, ry := range refLinesY {
		if ry > 5.0 && ry < height-5.0 {
			b.addLine(layerBendRoll, x, y+ry, x+width, y+ry)
		}
	}

	textH := clamp(math.Min(width, height)*0.045, 16.0, 42.0)
	b.addText(layerAnnotations, x, y+height+textH*1.6, textH, title)
	b.addText(layerAnnotations, x, y+height+textH*0.3, textH*0.75, subtitle)
}

// sectorBounds returns the tight bounding box (minX, minY, maxX, maxY) of an annular
// sector centered at (0,0) with radii [rIn, rOut] sweeping [0, thetaRad].
func sectorBounds(rIn, rOut, thetaRad float64) (minX, minY, maxX, maxY float64) {
	minX, minY = math.Inf(1), math.Inf(1)
	maxX, maxY = math.Inf(-1), math.Inf(-1)

	angles := []float64{0, thetaRad}
	for _, card := range []float64{0.5 * math.Pi, math.Pi, 1.5 * math.Pi} {
		if card > 0 && card < thetaRad {
			angles = append(angles, card)
		}
	}
	for _, ang := range angles {
		ca, sa := math.Cos(ang), math.Sin(ang)
		for _, r := range []float64{rIn, rOut} {
			px, py := r*ca, r*sa
			if px < minX {
				minX = px
			}
			if py < minY {
				minY = py
			}
			if px > maxX {
				maxX = px
			}
			if py > maxY {
				maxY = py
			}
		}
	}
	return minX, minY, maxX, maxY
}

// addFrustumPattern draws an exact unfolded cone/frustum annular sector positioned so
// its bounding box starts at (originX, originY), returning its placed width and height.
func (b *dxfBuilder) addFrustumPattern(originX, originY float64, dev FrustumDevelopment, title, material string) (w, h float64) {
	if dev.IsCylinder {
		sub := fmt.Sprintf("%s | CYLINDER FLAT %.1f x %.1f mm", material, dev.FlatWidth, dev.AxialLength)
		b.addRolledCylinder(originX, originY, dev.FlatWidth, dev.AxialLength, title, sub, nil)
		return dev.FlatWidth, dev.AxialLength
	}

	bMinX, bMinY, bMaxX, bMaxY := sectorBounds(dev.InnerRadius, dev.OuterRadius, dev.ThetaRad)
	w = bMaxX - bMinX
	h = bMaxY - bMinY

	// Place arc apex center (cx, cy) so the annular sector's lower-left corner sits at (originX, originY).
	cx := originX - bMinX
	cy := originY - bMinY

	// Concentric inner & outer cut arcs spanning 0..ThetaDeg.
	b.addArc(layerCutOutline, cx, cy, dev.InnerRadius, 0.0, dev.ThetaDeg)
	b.addArc(layerCutOutline, cx, cy, dev.OuterRadius, 0.0, dev.ThetaDeg)

	// Closing radial seam edges at theta=0 and theta=ThetaRad.
	b.addLine(layerCutOutline, cx+dev.InnerRadius, cy, cx+dev.OuterRadius, cy)
	cosT, sinT := math.Cos(dev.ThetaRad), math.Sin(dev.ThetaRad)
	b.addLine(layerCutOutline,
		cx+dev.InnerRadius*cosT, cy+dev.InnerRadius*sinT,
		cx+dev.OuterRadius*cosT, cy+dev.OuterRadius*sinT)

	// Mid-sector rolling centerline.
	midT := 0.5 * dev.ThetaRad
	b.addLine(layerCenterlines,
		cx+dev.InnerRadius*math.Cos(midT), cy+dev.InnerRadius*math.Sin(midT),
		cx+dev.OuterRadius*math.Cos(midT), cy+dev.OuterRadius*math.Sin(midT))

	textH := clamp(math.Min(w, h)*0.05, 16.0, 38.0)
	b.addText(layerAnnotations, originX, originY+h+textH*1.6, textH, title)
	sub := fmt.Sprintf("%s | R_in=%.2fmm R_out=%.2fmm L_slant=%.2fmm THETA=%.2fdeg",
		material, dev.InnerRadius, dev.OuterRadius, dev.SlantLength, dev.ThetaDeg)
	b.addText(layerAnnotations, originX, originY+h+textH*0.3, textH*0.75, sub)
	return w, h
}

// addFlangePattern draws a bolted flange ring with bore, OD, bolt circle, and bolt holes.
func (b *dxfBuilder) addFlangePattern(originX, originY float64, f flangeSpec, title, material string, qty int) (w, h float64) {
	const toMM = 1000.0
	od := f.OD * toMM
	id := f.ID * toMM
	bcd := f.BCD * toMM
	holeR := 0.5 * f.BoltHoleD * toMM

	cx := originX + 0.5*od
	cy := originY + 0.5*od

	b.addCircle(layerFlanges, cx, cy, 0.5*od)
	b.addCircle(layerFlanges, cx, cy, 0.5*id)
	b.addCircle(layerCenterlines, cx, cy, 0.5*bcd)

	// Crosshairs.
	b.addLine(layerCenterlines, cx-0.55*od, cy, cx+0.55*od, cy)
	b.addLine(layerCenterlines, cx, cy-0.55*od, cx, cy+0.55*od)

	// Bolt holes evenly distributed on bolt circle.
	for i := 0; i < f.BoltCount; i++ {
		ang := 2.0 * math.Pi * float64(i) / float64(f.BoltCount)
		hx := cx + 0.5*bcd*math.Cos(ang)
		hy := cy + 0.5*bcd*math.Sin(ang)
		b.addCircle(layerFlanges, hx, hy, holeR)
	}

	textH := clamp(od*0.045, 14.0, 32.0)
	b.addText(layerAnnotations, originX, originY+od+textH*1.6, textH,
		fmt.Sprintf("%s (QTY %d)", title, qty))
	b.addText(layerAnnotations, originX, originY+od+textH*0.3, textH*0.75,
		fmt.Sprintf("%s | OD=%.1f ID=%.1f BCD=%.1f %dxD%.1fmm",
			material, od, id, bcd, f.BoltCount, f.BoltHoleD*toMM))
	return od, od
}

func (b *dxfBuilder) build() []byte {
	var out bytes.Buffer
	minX, minY, maxX, maxY := b.minX, b.minY, b.maxX, b.maxY
	if !b.hasPts {
		minX, minY, maxX, maxY = 0, 0, 0, 0
	}

	// HEADER section (AutoCAD R12 AC1009, millimeter units).
	fmt.Fprintf(&out,
		"  0\nSECTION\n  2\nHEADER\n"+
			"  9\n$ACADVER\n  1\nAC1009\n"+
			"  9\n$INSUNITS\n 70\n4\n"+
			"  9\n$EXTMIN\n 10\n%.4f\n 20\n%.4f\n 30\n0.0000\n"+
			"  9\n$EXTMAX\n 10\n%.4f\n 20\n%.4f\n 30\n0.0000\n"+
			"  0\nENDSEC\n",
		minX, minY, maxX, maxY,
	)

	// TABLES section (LTYPE + LAYER tables).
	out.WriteString("  0\nSECTION\n  2\nTABLES\n")
	// LTYPE table
	out.WriteString(
		"  0\nTABLE\n  2\nLTYPE\n 70\n3\n" +
			"  0\nLTYPE\n  2\nCONTINUOUS\n 70\n0\n  3\nSolid line\n 72\n65\n 73\n0\n 40\n0.0\n" +
			"  0\nLTYPE\n  2\nDASHED\n 70\n0\n  3\nDashed line\n 72\n65\n 73\n2\n 40\n15.0\n 49\n10.0\n 49\n-5.0\n" +
			"  0\nLTYPE\n  2\nCENTER\n 70\n0\n  3\nCenter line\n 72\n65\n 73\n4\n 40\n35.0\n 49\n20.0\n 49\n-5.0\n 49\n5.0\n 49\n-5.0\n" +
			"  0\nENDTAB\n",
	)
	// LAYER table
	out.WriteString(
		"  0\nTABLE\n  2\nLAYER\n 70\n5\n" +
			"  0\nLAYER\n  2\nCUT_OUTLINE\n 70\n0\n 62\n7\n  6\nCONTINUOUS\n" +
			"  0\nLAYER\n  2\nBEND_ROLL\n 70\n0\n 62\n3\n  6\nDASHED\n" +
			"  0\nLAYER\n  2\nCENTERLINES\n 70\n0\n 62\n1\n  6\nCENTER\n" +
			"  0\nLAYER\n  2\nANNOTATIONS\n 70\n0\n 62\n5\n  6\nCONTINUOUS\n" +
			"  0\nLAYER\n  2\nFLANGES\n 70\n0\n 62\n6\n  6\nCONTINUOUS\n" +
			"  0\nENDTAB\n",
	)
	out.WriteString("  0\nENDSEC\n")

	// ENTITIES section.
	out.WriteString("  0\nSECTION\n  2\nENTITIES\n")
	out.Write(b.entities.Bytes())
	out.WriteString("  0\nENDSEC\n")
	out.WriteString("  0\nEOF\n")

	return out.Bytes()
}

// ExportDXF generates an AutoCAD R12 ASCII DXF containing unfolded 2D sheet-metal
// fabrication patterns (millimeters) laid out cleanly in three non-overlapping rows:
//   - Row 1: Rolled cylinder shells (quench tower, venturi throat, packed-bed column, inlet duct)
//   - Row 2: Exact frustum/cone flat-pattern developments (venturi converging, venturi diverging, top reducer)
//   - Row 3: Bolted flange rings with bolt-hole patterns
func (e *Exporter) ExportDXF(design model.ScrubberDesign) ([]byte, error) {
	if err := validateDesign(design); err != nil {
		return nil, err
	}
	g := deriveGeometry(design)

	const (
		toMM = 1000.0
		gapX = 250.0
		gapY = 350.0
	)

	dxf := newDXFBuilder()

	quenchMat := design.Quench.Material
	if quenchMat == "" {
		quenchMat = "316L SS"
	}
	venturiMat := design.Venturi.Material
	if venturiMat == "" {
		venturiMat = "316L SS"
	}
	colMat := design.PackedBed.Material
	if colMat == "" {
		colMat = "CPVC/PP"
	}

	// --- Row 1: Rolled Cylinder Shells ---
	curX := 0.0
	rowY := 0.0
	rowMaxH := 0.0

	// 1a. Quench Tower Shell
	qCirc := math.Pi * g.quenchDia * toMM
	qH := g.quenchH * toMM
	dxf.addRolledCylinder(
		curX, rowY, qCirc, qH,
		"STAGE 1: QUENCH TOWER SHELL",
		fmt.Sprintf("%s t=%.2fmm | D=%.1fmm CIRC=%.1fmm H=%.1fmm",
			quenchMat, g.quenchWall*toMM, g.quenchDia*toMM, qCirc, qH),
		nil,
	)
	curX += qCirc + gapX
	rowMaxH = math.Max(rowMaxH, qH)

	// 1b. Venturi Throat Cylinder
	vThCirc := math.Pi * g.ventThDia * toMM
	vThL := g.ventThLen * toMM
	dxf.addRolledCylinder(
		curX, rowY, vThCirc, vThL,
		"STAGE 2: VENTURI THROAT SHELL",
		fmt.Sprintf("%s t=%.2fmm | D=%.1fmm CIRC=%.1fmm L=%.1fmm",
			venturiMat, g.ventWall*toMM, g.ventThDia*toMM, vThCirc, vThL),
		nil,
	)
	curX += vThCirc + gapX
	rowMaxH = math.Max(rowMaxH, vThL)

	// 1c. Packed-Bed Column Shell (with packing & demister zone reference lines)
	colCirc := math.Pi * g.colDia * toMM
	colH := g.colH * toMM
	dxf.addRolledCylinder(
		curX, rowY, colCirc, colH,
		"STAGE 3: PACKED-BED COLUMN SHELL",
		fmt.Sprintf("%s t=%.2fmm | D=%.1fmm CIRC=%.1fmm H=%.1fmm",
			colMat, g.colWall*toMM, g.colDia*toMM, colCirc, colH),
		[]float64{g.packH * toMM, (g.colH - g.demisterH) * toMM},
	)
	curX += colCirc + gapX
	rowMaxH = math.Max(rowMaxH, colH)

	// 1d. Inlet Flue Duct Shell
	ductCirc := math.Pi * g.ductDia * toMM
	ductL := g.ductStubLen * toMM
	dxf.addRolledCylinder(
		curX, rowY, ductCirc, ductL,
		"INLET FLUE DUCT SHELL",
		fmt.Sprintf("%s t=%.2fmm | D=%.1fmm CIRC=%.1fmm L=%.1fmm",
			quenchMat, g.ductWall*toMM, g.ductDia*toMM, ductCirc, ductL),
		nil,
	)
	rowMaxH = math.Max(rowMaxH, ductL)

	// --- Row 2: Exact Frustum / Cone Developments ---
	rowY += rowMaxH + gapY
	curX = 0.0
	rowMaxH = 0.0

	// 2a. Venturi Converging Cone
	convDev, err := ComputeFrustumDevelopment(g.ventInDia*toMM, g.ventThDia*toMM, g.ventConvLen*toMM)
	if err != nil {
		return nil, fmt.Errorf("cad: venturi converging cone development: %w", err)
	}
	w, h := dxf.addFrustumPattern(curX, rowY, convDev, "STAGE 2: VENTURI CONVERGING CONE", venturiMat)
	curX += w + gapX
	rowMaxH = math.Max(rowMaxH, h)

	// 2b. Venturi Diverging Diffuser Cone
	divDev, err := ComputeFrustumDevelopment(g.ventThDia*toMM, g.ventOutDia*toMM, g.ventDivLen*toMM)
	if err != nil {
		return nil, fmt.Errorf("cad: venturi diverging cone development: %w", err)
	}
	w, h = dxf.addFrustumPattern(curX, rowY, divDev, "STAGE 2: VENTURI DIVERGING CONE", venturiMat)
	curX += w + gapX
	rowMaxH = math.Max(rowMaxH, h)

	// 2c. Stage 3 Top Column-to-Stack Reducer Cone
	redDev, err := ComputeFrustumDevelopment(g.colDia*toMM, g.stackDia*toMM, g.reducerH*toMM)
	if err != nil {
		return nil, fmt.Errorf("cad: column reducer cone development: %w", err)
	}
	w, h = dxf.addFrustumPattern(curX, rowY, redDev, "STAGE 3: COLUMN-TO-STACK REDUCER CONE", colMat)
	rowMaxH = math.Max(rowMaxH, h)

	// --- Row 3: Bolted Flange Rings ---
	rowY += rowMaxH + gapY
	curX = 0.0

	flanges := []struct {
		spec  flangeSpec
		title string
		mat   string
		qty   int
	}{
		{g.flangeQuench, "QUENCH TOWER FLANGE", quenchMat, 2},
		{g.flangeVentIn, "VENTURI INLET FLANGE", venturiMat, 2},
		{g.flangeVentOut, "VENTURI OUTLET FLANGE", venturiMat, 2},
		{g.flangeColumn, "PACKED-BED COLUMN FLANGE", colMat, 4},
		{g.flangeDuct, "INLET DUCT FLANGE", quenchMat, 2},
	}

	for _, fl := range flanges {
		fw, fh := dxf.addFlangePattern(curX, rowY, fl.spec, fl.title, fl.mat, fl.qty)
		curX += fw + gapX
		if fh > rowMaxH {
			rowMaxH = fh
		}
	}

	// Sheet title block annotation below Row 1.
	dxf.addText(
		layerAnnotations,
		0, -140.0, 40.0,
		fmt.Sprintf("KILNSCRUB 4-STAGE SCRUBBER FLAT PATTERNS (MM) | KILN %.1f CU.FT. | PEAK %.0f F",
			design.Input.KilnVolumeCuFt, design.Input.PeakExitTempF),
	)

	return dxf.build(), nil
}
