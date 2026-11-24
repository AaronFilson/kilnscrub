package cad

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

const (
	stlHeaderSize   = 80
	stlTriangleSize = 50
	stlSegments     = 36
	stlDegenerateSq = 1e-12 // minimum squared cross-product area in mm^4
)

type vec3 struct {
	x, y, z float64
}

func (a vec3) add(b vec3) vec3 { return vec3{a.x + b.x, a.y + b.y, a.z + b.z} }
func (a vec3) sub(b vec3) vec3 { return vec3{a.x - b.x, a.y - b.y, a.z - b.z} }
func (a vec3) scale(s float64) vec3 {
	return vec3{a.x * s, a.y * s, a.z * s}
}
func (a vec3) dot(b vec3) float64 { return a.x*b.x + a.y*b.y + a.z*b.z }
func (a vec3) cross(b vec3) vec3 {
	return vec3{
		a.y*b.z - a.z*b.y,
		a.z*b.x - a.x*b.z,
		a.x*b.y - a.y*b.x,
	}
}
func (a vec3) lenSq() float64 { return a.dot(a) }
func (a vec3) len() float64   { return math.Sqrt(a.lenSq()) }
func (a vec3) finite() bool   { return isFinite(a.x) && isFinite(a.y) && isFinite(a.z) }

type stlTriangle struct {
	n          vec3
	v1, v2, v3 vec3
}

type stlMesh struct {
	tris []stlTriangle
}

func (m *stlMesh) addTri(v1, v2, v3 vec3) {
	if !v1.finite() || !v2.finite() || !v3.finite() {
		return
	}
	e1 := v2.sub(v1)
	e2 := v3.sub(v1)
	cr := e1.cross(e2)
	lsq := cr.lenSq()
	if !isFinite(lsq) || lsq < stlDegenerateSq {
		return
	}
	invLen := 1.0 / math.Sqrt(lsq)
	n := cr.scale(invLen)
	if !n.finite() {
		return
	}
	// Verify triangle remains non-degenerate in float32 coordinates.
	f1 := vec3{float64(float32(v1.x)), float64(float32(v1.y)), float64(float32(v1.z))}
	f2 := vec3{float64(float32(v2.x)), float64(float32(v2.y)), float64(float32(v2.z))}
	f3 := vec3{float64(float32(v3.x)), float64(float32(v3.y)), float64(float32(v3.z))}
	if f2.sub(f1).cross(f3.sub(f1)).lenSq() < stlDegenerateSq {
		return
	}
	m.tris = append(m.tris, stlTriangle{n: n, v1: v1, v2: v2, v3: v3})
}

func (m *stlMesh) addQuad(a, b, c, d vec3) {
	m.addTri(a, b, c)
	m.addTri(a, c, d)
}

// orthonormalBasis constructs a right-handed orthonormal frame (u, v, w) with w = normalize(axis).
func orthonormalBasis(axis vec3) (u, v, w vec3, ok bool) {
	l := axis.len()
	if !isFinite(l) || l < 1e-9 {
		return vec3{}, vec3{}, vec3{}, false
	}
	w = axis.scale(1.0 / l)
	ref := vec3{0, 0, 1}
	if math.Abs(w.z) > 0.9 {
		ref = vec3{1, 0, 0}
	}
	u = ref.cross(w)
	ul := u.len()
	if ul < 1e-9 {
		return vec3{}, vec3{}, vec3{}, false
	}
	u = u.scale(1.0 / ul)
	v = w.cross(u)
	return u, v, w, true
}

// addFrustum appends a closed (or optionally capped) frustum/cylinder from p0 (radius r0)
// to p1 (radius r1) with strictly outward-pointing unit normals.
func (m *stlMesh) addFrustum(p0, p1 vec3, r0, r1 float64, segs int, cap0, cap1 bool) {
	if segs < 3 || (r0 <= 1e-6 && r1 <= 1e-6) {
		return
	}
	u, v, _, ok := orthonormalBasis(p1.sub(p0))
	if !ok {
		return
	}

	ring0 := make([]vec3, segs)
	ring1 := make([]vec3, segs)
	for i := 0; i < segs; i++ {
		ang := 2.0 * math.Pi * float64(i) / float64(segs)
		ca, sa := math.Cos(ang), math.Sin(ang)
		dir := u.scale(ca).add(v.scale(sa))
		ring0[i] = p0.add(dir.scale(r0))
		ring1[i] = p1.add(dir.scale(r1))
	}

	for i := 0; i < segs; i++ {
		j := (i + 1) % segs
		// Outward lateral surface: (ring0[j]-ring0[i]) x (ring1[i]-ring0[i]) points radially outward.
		if r0 > 1e-6 && r1 > 1e-6 {
			m.addQuad(ring0[i], ring0[j], ring1[j], ring1[i])
		} else if r0 > 1e-6 {
			m.addTri(ring0[i], ring0[j], p1)
		} else {
			m.addTri(p0, ring1[j], ring1[i])
		}

		if cap0 && r0 > 1e-6 {
			// Outward normal at p0 is -w.
			m.addTri(p0, ring0[j], ring0[i])
		}
		if cap1 && r1 > 1e-6 {
			// Outward normal at p1 is +w.
			m.addTri(p1, ring1[i], ring1[j])
		}
	}
}

// addCylinder appends a watertight closed right circular cylinder from p0 to p1.
func (m *stlMesh) addCylinder(p0, p1 vec3, r float64, segs int) {
	m.addFrustum(p0, p1, r, r, segs, true, true)
}

// addFlangeRing appends a raised bolted flange collar centered at center along axis.
func (m *stlMesh) addFlangeRing(center, axis vec3, od, thickness float64, segs int) {
	l := axis.len()
	if l < 1e-9 || od <= 0 || thickness <= 0 {
		return
	}
	dir := axis.scale(1.0 / l)
	half := dir.scale(0.5 * thickness)
	m.addCylinder(center.sub(half), center.add(half), 0.5*od, segs)
}

// addBox appends an axis-aligned watertight box centered at c with full dimensions (sx, sy, sz).
func (m *stlMesh) addBox(c vec3, sx, sy, sz float64) {
	if sx <= 1e-6 || sy <= 1e-6 || sz <= 1e-6 {
		return
	}
	hx, hy, hz := 0.5*sx, 0.5*sy, 0.5*sz
	p000 := vec3{c.x - hx, c.y - hy, c.z - hz}
	p100 := vec3{c.x + hx, c.y - hy, c.z - hz}
	p110 := vec3{c.x + hx, c.y + hy, c.z - hz}
	p010 := vec3{c.x - hx, c.y + hy, c.z - hz}
	p001 := vec3{c.x - hx, c.y - hy, c.z + hz}
	p101 := vec3{c.x + hx, c.y - hy, c.z + hz}
	p111 := vec3{c.x + hx, c.y + hy, c.z + hz}
	p011 := vec3{c.x - hx, c.y + hy, c.z + hz}

	// +Z top
	m.addQuad(p001, p101, p111, p011)
	// -Z bottom
	m.addQuad(p000, p010, p110, p100)
	// +X right
	m.addQuad(p100, p110, p111, p101)
	// -X left
	m.addQuad(p000, p001, p011, p010)
	// +Y back
	m.addQuad(p010, p011, p111, p110)
	// -Y front
	m.addQuad(p000, p100, p101, p001)
}

// addVoluteHousing appends a centrifugal blower spiral scroll housing in the XZ plane
// centered at center, extruded along Y by [-width/2, +width/2], with radius growing linearly
// from rStart at theta=0 (+X) around 2*pi to rEnd at theta=2*pi (+X) plus cutoff step wall.
func (m *stlMesh) addVoluteHousing(center vec3, width, rStart, rEnd float64, segs int) {
	if segs < 8 || width <= 1e-6 || rStart <= 1e-6 || rEnd <= rStart {
		return
	}
	hy := 0.5 * width
	frontCenter := vec3{center.x, center.y - hy, center.z}
	backCenter := vec3{center.x, center.y + hy, center.z}

	frontPts := make([]vec3, segs+1)
	backPts := make([]vec3, segs+1)
	for i := 0; i <= segs; i++ {
		t := float64(i) / float64(segs)
		ang := 2.0 * math.Pi * t
		r := rStart + (rEnd-rStart)*t
		px := center.x + r*math.Cos(ang)
		pz := center.z + r*math.Sin(ang)
		frontPts[i] = vec3{px, center.y - hy, pz}
		backPts[i] = vec3{px, center.y + hy, pz}
	}

	for i := 0; i < segs; i++ {
		// Outer spiral perimeter: as ang increases in XZ from +X toward +Z,
		// tangent T = (-sin, 0, cos), back-front = (0, +w, 0).
		// (front[i+1] - front[i]) x (back[i] - front[i]) = (-sin,0,cos) x (0,1,0) = (-cos, 0, -sin) (inward!).
		// Therefore outward winding is (front[i], back[i], back[i+1], front[i+1]).
		m.addQuad(frontPts[i], backPts[i], backPts[i+1], frontPts[i+1])

		if i < segs-1 {
			// Front face (-Y outward normal):
			m.addTri(frontCenter, frontPts[i], frontPts[i+1])
			// Back face (+Y outward normal): reverse winding.
			m.addTri(backCenter, backPts[i+1], backPts[i])
		} else {
			// Split the final wedge through frontPts[0]/backPts[0] (collinear with center and pts[segs])
			// so the radial cutoff step edge (pts[0], pts[segs]) shares exact vertices with zero T-junctions.
			m.addTri(frontCenter, frontPts[i], frontPts[0])
			m.addTri(frontPts[0], frontPts[i], frontPts[segs])

			m.addTri(backCenter, backPts[0], backPts[i])
			m.addTri(backPts[0], backPts[segs], backPts[i])
		}
	}

	// Close spiral radial step (cutoff tongue) from rStart to rEnd at ang=0 (+X):
	// The spiral interior lies below z=cz (angle -> 2*pi, radius -> rEnd) and exterior
	// lies above z=cz (angle -> 0+, radius -> rStart), so the outward normal points +Z.
	// (+X) x (+Y) = +Z -> (frontStart, frontEnd, backEnd, backStart).
	m.addQuad(frontPts[0], frontPts[segs], backPts[segs], backPts[0])
}

// addPipeElbow appends a watertight segmented 90-degree circular bend connecting pStart
// (traveling along dirIn) to pEnd (exiting along dirOut) with bend radius bendR.
func (m *stlMesh) addPipeElbow(pStart, dirIn, dirOut vec3, pipeR, bendR float64, bendSegs, ringSegs int) {
	if bendSegs < 2 || ringSegs < 6 || pipeR <= 1e-6 || bendR <= pipeR*1.02 {
		return
	}
	dIn := dirIn.scale(1.0 / dirIn.len())
	dOut := dirOut.scale(1.0 / dirOut.len())
	if dIn.cross(dOut).len() < 1e-6 {
		return
	}
	// Center of curvature is at pStart + bendR * dOut.
	bendCenter := pStart.add(dOut.scale(bendR))

	rings := make([][]vec3, bendSegs+1)
	for s := 0; s <= bendSegs; s++ {
		phi := 0.5 * math.Pi * float64(s) / float64(bendSegs)
		// Radial vector from bendCenter toward pipe centerline:
		// At phi=0: -dOut; at phi=pi/2: +dIn.
		radial := dOut.scale(-math.Cos(phi)).add(dIn.scale(math.Sin(phi)))
		center := bendCenter.add(radial.scale(bendR))
		// Tangent along pipe progression: d(radial)/dphi = dOut*sin(phi) + dIn*cos(phi).
		tangent := dOut.scale(math.Sin(phi)).add(dIn.scale(math.Cos(phi)))
		// Right-handed frame (u=radial, v=tangent x radial, w=tangent):
		// Verify: u x v = radial x (tangent x radial) = tangent (since orthogonal unit vectors)!
		u := radial
		v := tangent.cross(u)
		ring := make([]vec3, ringSegs)
		for i := 0; i < ringSegs; i++ {
			ang := 2.0 * math.Pi * float64(i) / float64(ringSegs)
			dir := u.scale(math.Cos(ang)).add(v.scale(math.Sin(ang)))
			ring[i] = center.add(dir.scale(pipeR))
		}
		rings[s] = ring
	}

	for s := 0; s < bendSegs; s++ {
		r0, r1 := rings[s], rings[s+1]
		for i := 0; i < ringSegs; i++ {
			j := (i + 1) % ringSegs
			m.addQuad(r0[i], r0[j], r1[j], r1[i])
		}
	}
	// Close both ends so the elbow is watertight.
	p0 := bendCenter.add(dOut.scale(-bendR))
	p1 := bendCenter.add(dIn.scale(bendR))
	for i := 0; i < ringSegs; i++ {
		j := (i + 1) % ringSegs
		m.addTri(p0, rings[0][j], rings[0][i])
		m.addTri(p1, rings[bendSegs][i], rings[bendSegs][j])
	}
}

// shiftX translates all mesh vertices along X so that minX equals targetMinX.
func (m *stlMesh) shiftX(dx float64) {
	for i := range m.tris {
		m.tris[i].v1.x += dx
		m.tris[i].v2.x += dx
		m.tris[i].v3.x += dx
	}
}

func (m *stlMesh) minX() float64 {
	if len(m.tris) == 0 {
		return 0
	}
	mn := math.Inf(1)
	for _, t := range m.tris {
		mn = math.Min(mn, math.Min(t.v1.x, math.Min(t.v2.x, t.v3.x)))
	}
	return mn
}

// encodeBinarySTL serializes the mesh to standard binary STL format (84 + 50*N bytes).
func (m *stlMesh) encodeBinarySTL(headerText string) ([]byte, error) {
	if len(m.tris) == 0 {
		return nil, fmt.Errorf("cad: cannot encode empty STL mesh")
	}

	totalBytes := stlHeaderSize + 4 + len(m.tris)*stlTriangleSize
	buf := bytes.NewBuffer(make([]byte, 0, totalBytes))

	var header [stlHeaderSize]byte
	copy(header[:], headerText)
	buf.Write(header[:])

	if err := binary.Write(buf, binary.LittleEndian, uint32(len(m.tris))); err != nil {
		return nil, fmt.Errorf("cad: write STL triangle count: %w", err)
	}

	var rec [stlTriangleSize]byte
	for _, t := range m.tris {
		binary.LittleEndian.PutUint32(rec[0:4], math.Float32bits(float32(t.n.x)))
		binary.LittleEndian.PutUint32(rec[4:8], math.Float32bits(float32(t.n.y)))
		binary.LittleEndian.PutUint32(rec[8:12], math.Float32bits(float32(t.n.z)))

		binary.LittleEndian.PutUint32(rec[12:16], math.Float32bits(float32(t.v1.x)))
		binary.LittleEndian.PutUint32(rec[16:20], math.Float32bits(float32(t.v1.y)))
		binary.LittleEndian.PutUint32(rec[20:24], math.Float32bits(float32(t.v1.z)))

		binary.LittleEndian.PutUint32(rec[24:28], math.Float32bits(float32(t.v2.x)))
		binary.LittleEndian.PutUint32(rec[28:32], math.Float32bits(float32(t.v2.y)))
		binary.LittleEndian.PutUint32(rec[32:36], math.Float32bits(float32(t.v2.z)))

		binary.LittleEndian.PutUint32(rec[36:40], math.Float32bits(float32(t.v3.x)))
		binary.LittleEndian.PutUint32(rec[40:44], math.Float32bits(float32(t.v3.y)))
		binary.LittleEndian.PutUint32(rec[44:48], math.Float32bits(float32(t.v3.z)))

		rec[48] = 0
		rec[49] = 0
		buf.Write(rec[:])
	}

	return buf.Bytes(), nil
}

// ExportSTL generates a binary STL (millimeters) of the complete 4-stage scrubber assembly:
// inlet flue duct + bypass damper tee, Stage 1 316L SS quench tower + end caps/flanges,
// Stage 2 converging-throat-diverging venturi, Stage 3 CPVC/PP packed-bed column + sump +
// demister section + top reducer & stack, and Stage 4 ID centrifugal blower housing +
// outlet transition.
func (e *Exporter) ExportSTL(design model.ScrubberDesign) ([]byte, error) {
	if err := validateDesign(design); err != nil {
		return nil, err
	}
	g := deriveGeometry(design)

	// Convert all assembly dimensions to millimeters (Z-up, flow along +X).
	const toMM = 1000.0
	ductR := 0.5 * g.ductDia * toMM
	ductStub := g.ductStubLen * toMM

	qR := 0.5 * g.quenchDia * toMM
	qH := g.quenchH * toMM

	vInR := 0.5 * g.ventInDia * toMM
	vThR := 0.5 * g.ventThDia * toMM
	vOutR := 0.5 * g.ventOutDia * toMM
	vConvL := g.ventConvLen * toMM
	vThL := g.ventThLen * toMM
	vDivL := g.ventDivLen * toMM

	colR := 0.5 * g.colDia * toMM
	colH := g.colH * toMM
	sumpH := g.sumpH * toMM
	packH := g.packH * toMM
	demisterH := g.demisterH * toMM
	stackR := 0.5 * g.stackDia * toMM
	reducerH := g.reducerH * toMM
	stackRiserH := g.stackRiserH * toMM

	wheelD := g.wheelDia * toMM
	blowerW := g.blowerWidth * toMM

	// Vertical datum (Z=0 is floor).
	// Venturi centerline enters Stage 3 just above the liquid sump.
	zVenturi := sumpH + 0.65*(2.0*vOutR)
	zQuenchBase := math.Max(0.0, zVenturi-0.65*(2.0*vInR))
	zQuenchTop := zQuenchBase + qH

	// Horizontal layout along +X.
	stubGap := 50.0
	xQuench := 0.0
	xVentStart := xQuench + qR + stubGap
	xThroatStart := xVentStart + vConvL
	xThroatEnd := xThroatStart + vThL
	xVentEnd := xThroatEnd + vDivL
	xPacked := xVentEnd + stubGap + colR

	mesh := &stlMesh{}

	// --- 0. Inlet Flue Duct + Fail-Safe Bypass Damper Tee ---
	zInlet := zQuenchTop - math.Max(1.3*ductR, 0.16*qH)
	if zInlet < zQuenchBase+ductR {
		zInlet = zQuenchBase + 0.5*qH
	}
	xInletStart := xQuench - qR - ductStub
	mesh.addCylinder(vec3{xInletStart, 0, zInlet}, vec3{xQuench, 0, zInlet}, ductR, stlSegments)
	mesh.addFlangeRing(
		vec3{xInletStart + 0.5*g.flangeDuct.Thickness*toMM, 0, zInlet},
		vec3{1, 0, 0},
		g.flangeDuct.OD*toMM,
		g.flangeDuct.Thickness*toMM,
		stlSegments,
	)
	// Bypass damper vertical tee stub on inlet duct.
	xBypass := xInletStart + 0.5*ductStub
	bypassH := math.Max(180.0, 2.2*ductR)
	mesh.addCylinder(vec3{xBypass, 0, zInlet}, vec3{xBypass, 0, zInlet + bypassH}, 0.85*ductR, stlSegments)
	mesh.addFlangeRing(
		vec3{xBypass, 0, zInlet + bypassH},
		vec3{0, 0, 1},
		g.flangeDuct.OD*0.90*toMM,
		g.flangeDuct.Thickness*toMM,
		stlSegments,
	)

	// --- 1. Stage 1: 316L SS Quench Tower ---
	// Support skirt to floor if quench base is elevated above Z=0.
	if zQuenchBase > 5.0 {
		mesh.addCylinder(vec3{xQuench, 0, 0}, vec3{xQuench, 0, zQuenchBase}, 0.82*qR, stlSegments)
		mesh.addFlangeRing(
			vec3{xQuench, 0, 0.5 * g.flangeQuench.Thickness * toMM},
			vec3{0, 0, 1},
			g.flangeQuench.OD*toMM,
			g.flangeQuench.Thickness*toMM,
			stlSegments,
		)
	}
	// Main quench shell cylinder + top/bottom caps & bolted flanges.
	mesh.addCylinder(vec3{xQuench, 0, zQuenchBase}, vec3{xQuench, 0, zQuenchTop}, qR, stlSegments)
	qFlangeT := g.flangeQuench.Thickness * toMM
	mesh.addFlangeRing(vec3{xQuench, 0, zQuenchBase + 0.5*qFlangeT}, vec3{0, 0, 1}, g.flangeQuench.OD*toMM, qFlangeT, stlSegments)
	mesh.addFlangeRing(vec3{xQuench, 0, zQuenchTop - 0.5*qFlangeT}, vec3{0, 0, 1}, g.flangeQuench.OD*toMM, qFlangeT, stlSegments)
	// Slightly domed / beveled top closure cap.
	mesh.addFrustum(vec3{xQuench, 0, zQuenchTop}, vec3{xQuench, 0, zQuenchTop + 0.12*qR}, qR, 0.72*qR, stlSegments, true, true)

	// Quench spray lance ports around upper quench zone.
	nozzCount := g.quenchNozz
	if nozzCount > 6 {
		nozzCount = 6
	}
	zSpray := zQuenchTop - 0.25*qH
	for i := 0; i < nozzCount; i++ {
		ang := 2.0*math.Pi*float64(i)/float64(nozzCount) + math.Pi/4.0
		pInner := vec3{xQuench + 0.85*qR*math.Cos(ang), 0.85 * qR * math.Sin(ang), zSpray}
		pOuter := vec3{xQuench + (qR+45.0)*math.Cos(ang), (qR + 45.0) * math.Sin(ang), zSpray}
		mesh.addCylinder(pInner, pOuter, 16.0, 16)
	}

	// --- 2. Stage 2: 316L SS Wet Venturi Scrubber ---
	// Quench outlet stub to venturi inlet.
	mesh.addCylinder(vec3{xQuench, 0, zVenturi}, vec3{xVentStart, 0, zVenturi}, vInR, stlSegments)
	mesh.addFlangeRing(vec3{xVentStart, 0, zVenturi}, vec3{1, 0, 0}, g.flangeVentIn.OD*toMM, g.flangeVentIn.Thickness*toMM, stlSegments)

	// Converging cone -> throat cylinder -> diverging diffuser cone.
	mesh.addFrustum(vec3{xVentStart, 0, zVenturi}, vec3{xThroatStart, 0, zVenturi}, vInR, vThR, stlSegments, true, true)
	mesh.addCylinder(vec3{xThroatStart, 0, zVenturi}, vec3{xThroatEnd, 0, zVenturi}, vThR, stlSegments)
	// Venturi throat liquid injection manifold collar.
	xThMid := 0.5 * (xThroatStart + xThroatEnd)
	mesh.addFlangeRing(vec3{xThMid, 0, zVenturi}, vec3{1, 0, 0}, 2.0*vThR+45.0, math.Min(35.0, 0.6*vThL), stlSegments)
	mesh.addFrustum(vec3{xThroatEnd, 0, zVenturi}, vec3{xVentEnd, 0, zVenturi}, vThR, vOutR, stlSegments, true, true)

	// Venturi outlet flange & transition stub into Stage 3 packed-bed column.
	mesh.addFlangeRing(vec3{xVentEnd, 0, zVenturi}, vec3{1, 0, 0}, g.flangeVentOut.OD*toMM, g.flangeVentOut.Thickness*toMM, stlSegments)
	mesh.addCylinder(vec3{xVentEnd, 0, zVenturi}, vec3{xPacked, 0, zVenturi}, vOutR, stlSegments)

	// --- 3. Stage 3: CPVC/PP Packed-Bed Column + Sump + Demister + Stack ---
	colFlangeT := g.flangeColumn.Thickness * toMM
	colFlangeOD := g.flangeColumn.OD * toMM

	// Liquid sump tank (Z = 0 to sumpH) with base flange.
	mesh.addCylinder(vec3{xPacked, 0, 0}, vec3{xPacked, 0, sumpH}, colR, stlSegments)
	mesh.addFlangeRing(vec3{xPacked, 0, 0.5 * colFlangeT}, vec3{0, 0, 1}, colFlangeOD, colFlangeT, stlSegments)
	mesh.addFlangeRing(vec3{xPacked, 0, sumpH}, vec3{0, 0, 1}, colFlangeOD, colFlangeT, stlSegments)

	// Main packed-bed & demister column shell. PackedBed.ColumnHeight (colH) already includes
	// the integral liquid sump height (Z = 0 to sumpH), so the upper shell spans Z = sumpH..colH.
	zColTop := colH
	if zColTop <= sumpH+2.0*colFlangeT {
		zColTop = sumpH + colH
	}
	mesh.addCylinder(vec3{xPacked, 0, sumpH}, vec3{xPacked, 0, zColTop}, colR, stlSegments)

	// Mid-column bolted service flange between packing zone and demister section.
	zDemisterBase := clamp(zColTop-demisterH-0.15*colH, sumpH+0.4*packH, zColTop-2.0*colFlangeT)
	if zDemisterBase > sumpH+2.0*colFlangeT && zDemisterBase < zColTop-2.0*colFlangeT {
		mesh.addFlangeRing(vec3{xPacked, 0, zDemisterBase}, vec3{0, 0, 1}, colFlangeOD, colFlangeT, stlSegments)
	}
	mesh.addFlangeRing(vec3{xPacked, 0, zColTop - 0.5*colFlangeT}, vec3{0, 0, 1}, colFlangeOD, colFlangeT, stlSegments)

	// Top reducer cone (colDia -> stackDia) + exhaust riser stack.
	zStackBase := zColTop + reducerH
	zStackTop := zStackBase + stackRiserH
	mesh.addFrustum(vec3{xPacked, 0, zColTop}, vec3{xPacked, 0, zStackBase}, colR, stackR, stlSegments, true, true)
	mesh.addCylinder(vec3{xPacked, 0, zStackBase}, vec3{xPacked, 0, zStackTop}, stackR, stlSegments)
	mesh.addFlangeRing(vec3{xPacked, 0, zStackBase + 0.5*g.flangeStack.Thickness*toMM}, vec3{0, 0, 1}, g.flangeStack.OD*toMM, g.flangeStack.Thickness*toMM, stlSegments)

	// --- 4. Stage 4: ID Centrifugal Blower Housing + Suction Duct + Outlet Transition ---
	scrollRStart := 0.55 * wheelD
	scrollREnd := 0.78 * wheelD
	zBlowerCenter := math.Max(scrollREnd+160.0, zVenturi)
	bendR := math.Max(1.4*stackR, 120.0)
	xBlower := xPacked + colR + scrollREnd + 2.0*bendR + 120.0

	// Overhead suction duct from Stage 3 stack top to ID blower suction eye.
	// 1) 90-deg elbow turning +Z -> +X at top of Stage 3 stack.
	mesh.addPipeElbow(vec3{xPacked, 0, zStackTop}, vec3{0, 0, 1}, vec3{1, 0, 0}, stackR, bendR, 10, stlSegments)
	zOverhead := zStackTop + bendR
	xHorizStart := xPacked + bendR
	xHorizEnd := xBlower - bendR
	if xHorizEnd > xHorizStart+5.0 {
		mesh.addCylinder(vec3{xHorizStart, 0, zOverhead}, vec3{xHorizEnd, 0, zOverhead}, stackR, stlSegments)
	}
	// 2) 90-deg elbow turning +X -> -Z down toward blower.
	mesh.addPipeElbow(vec3{xHorizEnd, 0, zOverhead}, vec3{1, 0, 0}, vec3{0, 0, -1}, stackR, bendR, 10, stlSegments)
	zDropTop := zOverhead - bendR
	ySuctionOffset := -(0.5*blowerW + stackR + 60.0)
	// Offset drop into suction plenum aligned with blower -Y suction eye.
	mesh.addCylinder(vec3{xBlower, 0, zDropTop}, vec3{xBlower, ySuctionOffset, zBlowerCenter}, stackR, stlSegments)
	// Suction inlet collar entering blower housing along +Y.
	mesh.addCylinder(vec3{xBlower, ySuctionOffset, zBlowerCenter}, vec3{xBlower, -0.5 * blowerW, zBlowerCenter}, stackR, stlSegments)
	mesh.addFlangeRing(vec3{xBlower, -0.5*blowerW - 0.35*stackR, zBlowerCenter}, vec3{0, 1, 0}, g.flangeStack.OD*toMM, g.flangeStack.Thickness*toMM, stlSegments)

	// Blower spiral scroll housing (XZ plane, axis along Y).
	mesh.addVoluteHousing(vec3{xBlower, 0, zBlowerCenter}, blowerW, scrollRStart, scrollREnd, stlSegments)

	// Blower motor housing on +Y side.
	motorLen := math.Max(200.0, 0.65*wheelD)
	motorR := math.Max(70.0, 0.28*wheelD)
	mesh.addCylinder(
		vec3{xBlower, 0.5 * blowerW, zBlowerCenter},
		vec3{xBlower, 0.5*blowerW + motorLen, zBlowerCenter},
		motorR,
		stlSegments,
	)

	// Blower structural pedestal resting on Z=0.
	pedH := math.Max(80.0, zBlowerCenter-0.60*scrollREnd)
	mesh.addBox(
		vec3{xBlower, 0.20 * motorLen, 0.5 * pedH},
		1.25*scrollREnd,
		blowerW+0.9*motorLen,
		pedH,
	)

	// Tangential rectangular outlet transition rising vertically (+Z) from outer scroll +X edge.
	outWidthX := math.Max(100.0, scrollREnd-scrollRStart+0.35*wheelD)
	outDepthY := 0.90 * blowerW
	outHeightZ := math.Max(220.0, 0.85*wheelD)
	xOutlet := xBlower + scrollREnd - 0.5*outWidthX
	zOutletBase := zBlowerCenter
	mesh.addBox(vec3{xOutlet, 0, zOutletBase + 0.5*outHeightZ}, outWidthX, outDepthY, outHeightZ)
	// Rectangular discharge flange at top of outlet transition.
	mesh.addBox(vec3{xOutlet, 0, zOutletBase + outHeightZ + 12.0}, outWidthX+60.0, outDepthY+60.0, 24.0)
	// Short round discharge stack transition above outlet flange.
	outStackR := math.Max(0.85*stackR, 0.38*math.Min(outWidthX, outDepthY))
	mesh.addFrustum(
		vec3{xOutlet, 0, zOutletBase + outHeightZ + 24.0},
		vec3{xOutlet, 0, zOutletBase + outHeightZ + 24.0 + 0.5*outHeightZ},
		1.15*outStackR,
		outStackR,
		stlSegments,
		true,
		true,
	)

	// Normalize X origin so minX = 0 mm for clean CAD placement.
	mesh.shiftX(-mesh.minX())

	header := fmt.Sprintf("kilnscrub 4-stage wet scrubber (mm) Q:%.0f V:%.0f P:%.0f",
		g.quenchDia*toMM, g.ventThDia*toMM, g.colDia*toMM)
	return mesh.encodeBinarySTL(header)
}
