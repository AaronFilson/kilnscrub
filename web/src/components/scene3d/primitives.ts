import * as THREE from "three";

export function clamp(val: number | undefined | null, lo: number, hi: number): number {
  const n = typeof val === "number" && Number.isFinite(val) ? val : lo;
  return Math.max(lo, Math.min(hi, n));
}

export function fmtDimMeters(meters: number, units: "in" | "mm"): string {
  if (units === "mm") {
    if (meters >= 1.5) return `${meters.toFixed(2)} m`;
    return `${Math.round(meters * 1000)} mm`;
  }
  const inches = meters / 0.0254;
  if (inches >= 48) {
    const ft = Math.floor(inches / 12);
    const remIn = Math.round(inches - ft * 12);
    return remIn === 12 ? `${ft + 1}'-0"` : `${ft}'-${remIn}"`;
  }
  return `${inches.toFixed(1)}"`;
}

/** Smooth 90-degree circular-arc cubic Bezier elbow (kappa = 0.55228475). */
export function makeArc90Bezier(
  start: THREE.Vector3,
  dirIn: THREE.Vector3,
  dirOut: THREE.Vector3,
  radius: number
): THREE.CubicBezierCurve3 {
  const k = 0.55228475 * radius;
  const p0 = start.clone();
  const p1 = start.clone().addScaledVector(dirIn, k);
  const p3 = start
    .clone()
    .addScaledVector(dirIn, radius)
    .addScaledVector(dirOut, radius);
  const p2 = p3.clone().addScaledVector(dirOut, -k);
  return new THREE.CubicBezierCurve3(p0, p1, p2, p3);
}

/**
 * ANSI/ASME-style bolted pipe flange pair with PTFE/EPDM gasket ring and hex bolts/nuts
 * aligned along "x", "y", or "z".
 */
export function addBoltedFlange(
  parent: THREE.Object3D,
  x: number,
  y: number,
  z: number,
  axis: "x" | "y" | "z",
  pipeR: number,
  matFlange: THREE.Material,
  matBolt: THREE.Material,
  matGasket: THREE.Material,
  boltCount = 8
): THREE.Group {
  const flangeR = pipeR + Math.max(0.026, pipeR * 0.28);
  const thick = 0.026;
  const g = new THREE.Group();
  g.position.set(x, y, z);

  const ringShape = new THREE.Shape();
  ringShape.absarc(0, 0, flangeR, 0, Math.PI * 2, false);
  const hole = new THREE.Path();
  hole.absarc(0, 0, Math.max(0.01, pipeR * 0.96), 0, Math.PI * 2, true);
  ringShape.holes.push(hole);

  // Paired flange faces with dark PTFE/EPDM gasket sandwiched at center
  const faceThick = thick * 0.42;
  const faceGeo = new THREE.ExtrudeGeometry(ringShape, {
    depth: faceThick,
    bevelEnabled: false,
    curveSegments: 24,
  });
  const frontFace = new THREE.Mesh(faceGeo, matFlange);
  frontFace.position.z = thick * 0.08;
  g.add(frontFace);
  const backFace = new THREE.Mesh(faceGeo, matFlange);
  backFace.position.z = -thick * 0.08 - faceThick;
  g.add(backFace);

  const gasketGeo = new THREE.ExtrudeGeometry(ringShape, {
    depth: thick * 0.18,
    bevelEnabled: false,
    curveSegments: 24,
  });
  gasketGeo.translate(0, 0, -thick * 0.09);
  g.add(new THREE.Mesh(gasketGeo, matGasket));

  const boltCircleR = pipeR + (flangeR - pipeR) * 0.58;
  const boltR = Math.max(0.0055, (flangeR - pipeR) * 0.19);
  const boltGeo = new THREE.CylinderGeometry(boltR, boltR, thick + 0.020, 6);
  boltGeo.rotateX(Math.PI / 2);
  for (let i = 0; i < boltCount; i++) {
    const a = (i / boltCount) * Math.PI * 2;
    const b = new THREE.Mesh(boltGeo, matBolt);
    b.position.set(Math.cos(a) * boltCircleR, Math.sin(a) * boltCircleR, 0);
    g.add(b);
  }

  if (axis === "x") g.rotation.y = Math.PI / 2;
  else if (axis === "y") g.rotation.x = Math.PI / 2;
  parent.add(g);
  return g;
}

/** 316L Stainless Sanitary Tri-Clamp ferrule + hinged band + wingnut. */
export function addTriClamp(
  parent: THREE.Object3D,
  x: number,
  y: number,
  z: number,
  axis: "x" | "y" | "z",
  pipeR: number,
  matMetal: THREE.Material,
  matGasket: THREE.Material
): THREE.Group {
  const g = new THREE.Group();
  g.position.set(x, y, z);

  const band = new THREE.Mesh(
    new THREE.TorusGeometry(pipeR + 0.014, 0.011, 8, 22),
    matMetal
  );
  g.add(band);

  const gasket = new THREE.Mesh(
    new THREE.TorusGeometry(pipeR + 0.005, 0.005, 6, 20),
    matGasket
  );
  g.add(gasket);

  const wingPin = new THREE.Mesh(
    new THREE.CylinderGeometry(0.005, 0.005, 0.048, 8),
    matMetal
  );
  wingPin.position.set(pipeR + 0.028, 0, 0);
  g.add(wingPin);

  if (axis === "x") g.rotation.y = Math.PI / 2;
  else if (axis === "y") g.rotation.x = Math.PI / 2;
  parent.add(g);
  return g;
}

/** Schedule 80 CPVC / brass octagonal socket pipe union. */
export function addPipeUnion(
  parent: THREE.Object3D,
  x: number,
  y: number,
  z: number,
  axis: "x" | "y" | "z",
  pipeR: number,
  matBody: THREE.Material,
  matNut: THREE.Material
): THREE.Group {
  const g = new THREE.Group();
  g.position.set(x, y, z);

  const sleeve = new THREE.Mesh(
    new THREE.CylinderGeometry(pipeR * 1.32, pipeR * 1.32, 0.06, 14),
    matBody
  );
  sleeve.rotation.x = Math.PI / 2;
  g.add(sleeve);

  const nut = new THREE.Mesh(
    new THREE.CylinderGeometry(pipeR * 1.62, pipeR * 1.62, 0.028, 8),
    matNut
  );
  nut.rotation.x = Math.PI / 2;
  g.add(nut);

  if (axis === "x") g.rotation.y = Math.PI / 2;
  else if (axis === "y") g.rotation.x = Math.PI / 2;
  parent.add(g);
  return g;
}

/** Quarter-turn ball valve with socket ends + lever handle. */
export function addBallValve(
  parent: THREE.Object3D,
  x: number,
  y: number,
  z: number,
  axis: "x" | "y" | "z",
  pipeR: number,
  matBody: THREE.Material,
  matHandle: THREE.Material
): THREE.Group {
  const g = new THREE.Group();
  g.position.set(x, y, z);

  const body = new THREE.Mesh(
    new THREE.CylinderGeometry(pipeR * 1.55, pipeR * 1.55, Math.max(0.065, pipeR * 3.4), 14),
    matBody
  );
  body.rotation.x = Math.PI / 2;
  g.add(body);

  const stem = new THREE.Mesh(
    new THREE.CylinderGeometry(pipeR * 0.42, pipeR * 0.42, pipeR * 1.9, 10),
    matBody
  );
  stem.position.set(0, pipeR * 1.45, 0);
  g.add(stem);

  const lever = new THREE.Mesh(
    new THREE.BoxGeometry(pipeR * 0.55, pipeR * 0.35, Math.max(0.075, pipeR * 4.0)),
    matHandle
  );
  lever.position.set(0, pipeR * 2.4, pipeR * 0.9);
  g.add(lever);

  if (axis === "x") g.rotation.y = Math.PI / 2;
  else if (axis === "y") g.rotation.x = Math.PI / 2;
  parent.add(g);
  return g;
}

/** IS 651 Salt-Glazed Stoneware (SWG) bell-and-spigot flared mortar joint collar. */
export function addBellSpigotCollar(
  parent: THREE.Object3D,
  x: number,
  y: number,
  z: number,
  axis: "x" | "y" | "z",
  pipeR: number,
  matClay: THREE.Material,
  matMortar: THREE.Material
): THREE.Group {
  const g = new THREE.Group();
  g.position.set(x, y, z);

  const bell = new THREE.Mesh(
    new THREE.CylinderGeometry(pipeR * 1.18, pipeR * 1.24, 0.085, 20),
    matClay
  );
  bell.rotation.x = Math.PI / 2;
  g.add(bell);

  const fillet = new THREE.Mesh(
    new THREE.TorusGeometry(pipeR * 1.13, 0.015, 8, 20),
    matMortar
  );
  fillet.position.z = 0.044;
  g.add(fillet);

  if (axis === "x") g.rotation.y = Math.PI / 2;
  else if (axis === "y") g.rotation.x = Math.PI / 2;
  parent.add(g);
  return g;
}

/** Brass Bourdon pressure gauge on siphon stem facing +Z. */
export function addPressureGauge(
  parent: THREE.Object3D,
  x: number,
  y: number,
  z: number,
  needleAngleRad: number,
  matBrass: THREE.Material,
  matDial: THREE.Material,
  matNeedle: THREE.Material,
  scale = 1.0
): THREE.Group {
  const g = new THREE.Group();
  g.position.set(x, y, z);

  const stem = new THREE.Mesh(
    new THREE.CylinderGeometry(0.013 * scale, 0.013 * scale, 0.07 * scale, 10),
    matBrass
  );
  stem.rotation.x = Math.PI / 2;
  stem.position.z = -0.025 * scale;
  g.add(stem);

  const bezel = new THREE.Mesh(
    new THREE.CylinderGeometry(0.068 * scale, 0.068 * scale, 0.032 * scale, 22),
    matBrass
  );
  bezel.rotation.x = Math.PI / 2;
  g.add(bezel);

  const face = new THREE.Mesh(
    new THREE.CylinderGeometry(0.056 * scale, 0.056 * scale, 0.036 * scale, 22),
    matDial
  );
  face.rotation.x = Math.PI / 2;
  face.position.z = 0.003 * scale;
  g.add(face);

  const needle = new THREE.Mesh(
    new THREE.BoxGeometry(0.008 * scale, 0.044 * scale, 0.010 * scale),
    matNeedle
  );
  needle.position.set(0.008 * scale, 0.008 * scale, 0.020 * scale);
  needle.rotation.z = needleAngleRad;
  g.add(needle);

  parent.add(g);
  return g;
}

/**
 * Real NFPA 86 200°F Fail-Open Bypass Damper Tee on hot kiln inlet flue:
 * vertical relief tee + rain cap + external spring-return actuator arm + internal blade.
 */
export function addBypassDamperTee(
  parent: THREE.Object3D,
  x: number,
  y: number,
  z: number,
  ductR: number,
  matHot: THREE.Material,
  matFlange: THREE.Material,
  matBolt: THREE.Material,
  matGasket: THREE.Material,
  matSafetyRed: THREE.Material,
  matDark: THREE.Material,
  cut: boolean
): THREE.Group {
  const g = new THREE.Group();
  g.position.set(x, y, z);

  const riserH = Math.max(0.42, ductR * 4.2);
  const riserR = ductR * 0.92;
  const riser = new THREE.Mesh(
    new THREE.CylinderGeometry(
      riserR,
      riserR,
      riserH,
      24,
      1,
      cut,
      cut ? Math.PI * 0.5 : 0,
      cut ? Math.PI : Math.PI * 2
    ),
    matHot
  );
  riser.position.y = riserH * 0.5;
  g.add(riser);

  // Bolted tee collar + upper flange
  addBoltedFlange(g, 0, ductR * 1.05, 0, "y", riserR, matFlange, matBolt, matGasket, 8);
  addBoltedFlange(g, 0, riserH, 0, "y", riserR, matFlange, matBolt, matGasket, 8);

  // Conical weather rain cap on standoffs
  const cap = new THREE.Mesh(
    new THREE.ConeGeometry(riserR * 1.48, 0.09, 20),
    matFlange
  );
  cap.position.y = riserH + 0.14;
  g.add(cap);
  for (const ang of [0, (2 * Math.PI) / 3, (4 * Math.PI) / 3]) {
    const leg = new THREE.Mesh(
      new THREE.CylinderGeometry(0.006, 0.006, 0.11, 6),
      matBolt
    );
    leg.position.set(Math.cos(ang) * riserR * 0.9, riserH + 0.055, Math.sin(ang) * riserR * 0.9);
    g.add(leg);
  }

  // Internal butterfly damper blade (closed in normal scrub mode, fails OPEN on 200°F trip)
  const blade = new THREE.Mesh(
    new THREE.CylinderGeometry(riserR * 0.88, riserR * 0.88, 0.012, 20),
    matSafetyRed
  );
  blade.position.y = ductR * 1.65;
  blade.rotation.z = 0.16;
  g.add(blade);

  // External 24VDC spring-return fail-open actuator box + red counterweight lever
  const actBox = new THREE.Mesh(
    new THREE.BoxGeometry(0.09, 0.11, 0.08),
    matSafetyRed
  );
  actBox.position.set(0, ductR * 1.65, riserR + 0.06);
  g.add(actBox);

  const lever = new THREE.Mesh(
    new THREE.CylinderGeometry(0.008, 0.008, 0.18, 8),
    matDark
  );
  lever.rotation.z = Math.PI / 3;
  lever.position.set(0.07, ductR * 1.65 + 0.03, riserR + 0.105);
  g.add(lever);

  const weight = new THREE.Mesh(
    new THREE.CylinderGeometry(0.024, 0.024, 0.032, 12),
    matSafetyRed
  );
  weight.rotation.x = Math.PI / 2;
  weight.position.set(0.14, ductR * 1.65 + 0.065, riserR + 0.105);
  g.add(weight);

  parent.add(g);
  return g;
}

/** True lofted rectangular-to-round sheet-metal transition (y = 0 rect -> y = h circle). */
export function buildRectToRoundGeometry(
  w: number,
  d: number,
  r: number,
  h: number,
  segs = 36
): THREE.BufferGeometry {
  const positions: number[] = [];
  const indices: number[] = [];
  const hw = w * 0.5;
  const hd = d * 0.5;
  for (let i = 0; i <= segs; i++) {
    const a = (i / segs) * Math.PI * 2;
    const ca = Math.cos(a);
    const sa = Math.sin(a);
    const s = 1.0 / Math.max(Math.abs(ca) / hw, Math.abs(sa) / hd);
    positions.push(ca * s, 0, sa * s);
    positions.push(ca * r, h, sa * r);
  }
  for (let i = 0; i < segs; i++) {
    const b0 = i * 2;
    const t0 = b0 + 1;
    const b1 = b0 + 2;
    const t1 = b0 + 3;
    indices.push(b0, t0, b1);
    indices.push(b1, t0, t1);
  }
  const geo = new THREE.BufferGeometry();
  geo.setAttribute("position", new THREE.Float32BufferAttribute(positions, 3));
  geo.setIndex(indices);
  geo.computeVertexNormals();
  return geo;
}

export interface VoluteProfile {
  innerPts: THREE.Vector2[];
  outerPts: THREE.Vector2[];
  xIn: number;
  rMax: number;
  yDis: number;
  rLeft: number;
  rBottom: number;
  disW: number;
}

/** Spiral volute 2D profiles (CCW around origin, vertical tangential upward discharge on +X). */
export function computeVoluteProfiles(
  wheelR: number,
  ductR: number,
  wallT: number
): VoluteProfile {
  const phi0 = 0.36 * Math.PI;
  const r0 = wheelR * 1.12;
  const xIn = wheelR * 0.88;
  const yTongue = wheelR * 0.82;
  const yDis = Math.max(wheelR * 1.32, yTongue + 0.14);
  const rMax = Math.max(wheelR * 1.66, xIn + ductR * 2.04);
  const spiralSteps = 56;

  const rAt = (frac: number) =>
    r0 + (rMax - r0) * (1.0 - Math.pow(1.0 - frac, 1.35));

  const innerPts: THREE.Vector2[] = [];
  innerPts.push(new THREE.Vector2(xIn, yDis));
  innerPts.push(new THREE.Vector2(xIn, yTongue));
  const tongueCurveIn = new THREE.QuadraticBezierCurve(
    new THREE.Vector2(xIn, yTongue),
    new THREE.Vector2(wheelR * 0.72, wheelR * 0.96),
    new THREE.Vector2(r0 * Math.cos(phi0), r0 * Math.sin(phi0))
  );
  for (let i = 1; i <= 8; i++) innerPts.push(tongueCurveIn.getPoint(i / 8));
  for (let s = 1; s <= spiralSteps; s++) {
    const frac = s / spiralSteps;
    const phi = phi0 + frac * (Math.PI * 2 - phi0);
    const rr = rAt(frac);
    innerPts.push(new THREE.Vector2(rr * Math.cos(phi), rr * Math.sin(phi)));
  }
  innerPts.push(new THREE.Vector2(rMax, yDis));

  const outerPts: THREE.Vector2[] = [];
  outerPts.push(new THREE.Vector2(xIn - wallT, yDis));
  outerPts.push(new THREE.Vector2(xIn - wallT, yTongue + wallT * 1.1));
  const r0Out = r0 + wallT;
  const tongueCurveOut = new THREE.QuadraticBezierCurve(
    new THREE.Vector2(xIn - wallT, yTongue + wallT * 1.1),
    new THREE.Vector2(wheelR * 0.72 - wallT * 0.7, wheelR * 0.96 + wallT * 0.9),
    new THREE.Vector2(r0Out * Math.cos(phi0), r0Out * Math.sin(phi0))
  );
  for (let i = 1; i <= 8; i++) outerPts.push(tongueCurveOut.getPoint(i / 8));
  for (let s = 1; s <= spiralSteps; s++) {
    const frac = s / spiralSteps;
    const phi = phi0 + frac * (Math.PI * 2 - phi0);
    const rr = rAt(frac) + wallT;
    outerPts.push(new THREE.Vector2(rr * Math.cos(phi), rr * Math.sin(phi)));
  }
  outerPts.push(new THREE.Vector2(rMax + wallT, yDis));

  const rLeft = rAt((Math.PI - phi0) / (Math.PI * 2 - phi0)) + wallT;
  const rBottom = rAt((1.5 * Math.PI - phi0) / (Math.PI * 2 - phi0)) + wallT;

  return { innerPts, outerPts, xIn, rMax, yDis, rLeft, rBottom, disW: rMax - xIn };
}

/** Backward-curved 9-blade centrifugal impeller rotating cleanly around +Z (CCW). */
export function buildImpellerGroup(
  wheelR: number,
  scrollW: number,
  matHub: THREE.Material,
  matBlade: THREE.Material,
  matBackplate: THREE.Material,
  matShroud: THREE.Material
): THREE.Group {
  const impGroup = new THREE.Group();
  const bladeDepth = scrollW * 0.62;
  const backZ = -scrollW * 0.34;

  const backplate = new THREE.Mesh(
    new THREE.CylinderGeometry(wheelR * 0.94, wheelR * 0.94, 0.016, 36),
    matBackplate
  );
  backplate.rotation.x = Math.PI / 2;
  backplate.position.z = backZ - 0.008;
  impGroup.add(backplate);

  const hub = new THREE.Mesh(
    new THREE.CylinderGeometry(wheelR * 0.16, wheelR * 0.23, scrollW * 0.56, 20),
    matHub
  );
  hub.rotation.x = Math.PI / 2;
  hub.position.z = -scrollW * 0.06;
  impGroup.add(hub);

  const noseNut = new THREE.Mesh(
    new THREE.CylinderGeometry(wheelR * 0.09, wheelR * 0.09, 0.032, 6),
    matBlade
  );
  noseNut.rotation.x = Math.PI / 2;
  noseNut.position.z = scrollW * 0.24;
  impGroup.add(noseNut);

  const rIn = wheelR * 0.30;
  const rOut = wheelR * 0.92;
  const bladeCount = 9;
  const halfT = Math.max(0.006, wheelR * 0.024);

  const bladeShape = new THREE.Shape();
  const ptsFwd: THREE.Vector2[] = [];
  const ptsRev: THREE.Vector2[] = [];
  const segs = 10;
  for (let i = 0; i <= segs; i++) {
    const u = i / segs;
    const r = rIn + (rOut - rIn) * u;
    const ang = -0.46 * Math.pow(u, 1.25);
    const cx = r * Math.cos(ang);
    const cy = r * Math.sin(ang);
    const rNext = rIn + (rOut - rIn) * Math.min(1, u + 0.02);
    const angNext = -0.46 * Math.pow(Math.min(1, u + 0.02), 1.25);
    const tx = rNext * Math.cos(angNext) - cx;
    const ty = rNext * Math.sin(angNext) - cy;
    const tl = Math.hypot(tx, ty) || 1;
    const nx = -ty / tl;
    const ny = tx / tl;
    ptsFwd.push(new THREE.Vector2(cx + nx * halfT, cy + ny * halfT));
    ptsRev.push(new THREE.Vector2(cx - nx * halfT, cy - ny * halfT));
  }
  bladeShape.moveTo(ptsFwd[0].x, ptsFwd[0].y);
  for (let i = 1; i < ptsFwd.length; i++) bladeShape.lineTo(ptsFwd[i].x, ptsFwd[i].y);
  for (let i = ptsRev.length - 1; i >= 0; i--) bladeShape.lineTo(ptsRev[i].x, ptsRev[i].y);
  bladeShape.closePath();

  const bladeGeo = new THREE.ExtrudeGeometry(bladeShape, {
    depth: bladeDepth,
    bevelEnabled: false,
  });
  bladeGeo.translate(0, 0, backZ);

  for (let k = 0; k < bladeCount; k++) {
    const vane = new THREE.Mesh(bladeGeo, matBlade);
    vane.rotation.z = (k * Math.PI * 2) / bladeCount;
    impGroup.add(vane);
  }

  const shroudShape = new THREE.Shape();
  shroudShape.absarc(0, 0, wheelR * 0.94, 0, Math.PI * 2, false);
  const shroudHole = new THREE.Path();
  shroudHole.absarc(0, 0, wheelR * 0.80, 0, Math.PI * 2, true);
  shroudShape.holes.push(shroudHole);
  const shroudGeo = new THREE.ExtrudeGeometry(shroudShape, {
    depth: 0.012,
    bevelEnabled: false,
    curveSegments: 36,
  });
  shroudGeo.translate(0, 0, backZ + bladeDepth);
  impGroup.add(new THREE.Mesh(shroudGeo, matShroud));

  return impGroup;
}

/** High-DPI architectural CAD dimension line + sprite label in 3D space. */
export function addDimensionCallout3D(
  parent: THREE.Object3D,
  p1: THREE.Vector3,
  p2: THREE.Vector3,
  offset: THREE.Vector3,
  label: string,
  accentHex: string,
  disposableTextures: THREE.Texture[]
): void {
  const a = p1.clone().add(offset);
  const b = p2.clone().add(offset);
  const dir = new THREE.Vector3().subVectors(b, a);
  const len = dir.length();
  if (len < 0.01) return;

  const lineColor = new THREE.Color(accentHex);
  const lineMat = new THREE.LineBasicMaterial({
    color: lineColor,
    transparent: true,
    opacity: 0.85,
  });
  const extMat = new THREE.LineBasicMaterial({
    color: lineColor,
    transparent: true,
    opacity: 0.45,
  });

  // Main dimension line
  const mainGeo = new THREE.BufferGeometry().setFromPoints([a, b]);
  parent.add(new THREE.Line(mainGeo, lineMat));

  // Extension lines back to geometry
  const ext1 = new THREE.BufferGeometry().setFromPoints([p1, a]);
  const ext2 = new THREE.BufferGeometry().setFromPoints([p2, b]);
  parent.add(new THREE.Line(ext1, extMat));
  parent.add(new THREE.Line(ext2, extMat));

  // Small end ticks perpendicular to offset
  const tickDir = offset.lengthSq() > 1e-6
    ? offset.clone().normalize().multiplyScalar(0.045)
    : new THREE.Vector3(0, 0.045, 0);
  parent.add(
    new THREE.Line(
      new THREE.BufferGeometry().setFromPoints([
        a.clone().sub(tickDir),
        a.clone().add(tickDir),
      ]),
      lineMat
    )
  );
  parent.add(
    new THREE.Line(
      new THREE.BufferGeometry().setFromPoints([
        b.clone().sub(tickDir),
        b.clone().add(tickDir),
      ]),
      lineMat
    )
  );

  // Canvas sprite pill centered on the dimension line
  if (typeof document === "undefined") return;
  const canvas = document.createElement("canvas");
  canvas.width = 384;
  canvas.height = 72;
  const ctx = canvas.getContext("2d");
  if (!ctx) return;

  ctx.clearRect(0, 0, canvas.width, canvas.height);
  ctx.font = "700 28px ui-monospace, SFMono-Regular, Menlo, Consolas, monospace";
  const textW = Math.min(350, ctx.measureText(label).width);
  const boxW = textW + 34;
  const boxH = 48;
  const bx = (canvas.width - boxW) * 0.5;
  const by = (canvas.height - boxH) * 0.5;

  ctx.fillStyle = "rgba(8, 16, 28, 0.90)";
  ctx.strokeStyle = accentHex;
  ctx.lineWidth = 3;
  ctx.beginPath();
  ctx.roundRect(bx, by, boxW, boxH, 10);
  ctx.fill();
  ctx.stroke();

  ctx.fillStyle = "#f1f5f9";
  ctx.textAlign = "center";
  ctx.textBaseline = "middle";
  ctx.fillText(label, canvas.width * 0.5, canvas.height * 0.5 + 1);

  const tex = new THREE.CanvasTexture(canvas);
  tex.colorSpace = THREE.SRGBColorSpace;
  disposableTextures.push(tex);

  const spriteMat = new THREE.SpriteMaterial({
    map: tex,
    transparent: true,
    depthTest: false,
    depthWrite: false,
  });
  const sprite = new THREE.Sprite(spriteMat);
  const mid = a.clone().lerp(b, 0.5);
  sprite.position.copy(mid);
  sprite.scale.set(0.92, 0.17, 1);
  sprite.renderOrder = 20;
  parent.add(sprite);
}
