import * as THREE from "three";
import { createParticleSpriteTexture } from "./materials";
import type { FlowSample, FlowStage, SprayZone } from "./types";

const STAGE_COLORS: Record<FlowStage, [number, number, number]> = {
  inlet: [0.99, 0.48, 0.14],   // Glowing hot kiln exhaust orange
  quench: [0.36, 0.78, 0.98],  // Adiabatically quenched cool mist cyan-blue
  venturi: [0.58, 0.90, 1.00], // High-velocity atomized throat mist
  packed: [0.16, 0.88, 0.62],  // Scrubbed HCl-free emerald vapor
  blower: [0.22, 0.92, 0.68],  // Clean exhaust through blower/stack
  plume: [0.82, 0.96, 1.00],   // Dissipating clean water-vapor plume
  steam: [0.96, 0.99, 1.00],   // High-speed bright white/cyan steam jet
  motive: [0.46, 0.84, 1.00],  // Dry ambient motive air stream
  diluter: [0.56, 0.88, 1.00], // Ambient squirrel-cage plume dilution air
};

interface StreamParticle {
  stream: "primary" | "motive" | "diluter";
  t: number;
  radialFrac: number;
  angleOffset: number;
  swirlRate: number;
}

interface DropletState {
  zoneIdx: number;
  t: number;
  radialFrac: number;
  angle: number;
}

/**
 * Multi-stream physics-aware 3D particle & water-spray system:
 * - Soft radial circular particle sprites (never blocky square points)
 * - Local velocity modulation (visibly accelerates through venturi/steam nozzle & slows in beds)
 * - Strictly confined inside each duct/vessel's true inner radius
 * - Dedicated falling spray droplets inside Quench & Packed Bed liquid distributors
 */
export class FlowParticleSystem {
  readonly group: THREE.Group;
  private spriteTex: THREE.CanvasTexture | null;

  private gasGeo: THREE.BufferGeometry;
  private gasMat: THREE.PointsMaterial;
  private gasPoints: THREE.Points;
  private gasParticles: StreamParticle[] = [];

  private sprayGeo: THREE.BufferGeometry;
  private sprayMat: THREE.PointsMaterial;
  private sprayPoints: THREE.Points;
  private droplets: DropletState[] = [];

  private primaryFlow: FlowSample[] = [];
  private motiveFlow: FlowSample[] = [];
  private diluterFlow: FlowSample[] = [];
  private sprayZones: SprayZone[] = [];

  private readonly gasCount = 240;
  private readonly maxDroplets = 96;

  constructor() {
    this.group = new THREE.Group();
    this.spriteTex = createParticleSpriteTexture();

    // Gas / steam / motive / plume stream points
    this.gasGeo = new THREE.BufferGeometry();
    this.gasGeo.setAttribute(
      "position",
      new THREE.BufferAttribute(new Float32Array(this.gasCount * 3), 3)
    );
    this.gasGeo.setAttribute(
      "color",
      new THREE.BufferAttribute(new Float32Array(this.gasCount * 3), 3)
    );
    this.gasMat = new THREE.PointsMaterial({
      size: 0.088,
      map: this.spriteTex || undefined,
      vertexColors: true,
      transparent: true,
      opacity: 0.92,
      alphaTest: 0.02,
      depthWrite: false,
      blending: THREE.AdditiveBlending,
    });
    this.gasPoints = new THREE.Points(this.gasGeo, this.gasMat);
    this.group.add(this.gasPoints);

    // Falling water spray droplets
    this.sprayGeo = new THREE.BufferGeometry();
    this.sprayGeo.setAttribute(
      "position",
      new THREE.BufferAttribute(new Float32Array(this.maxDroplets * 3), 3)
    );
    this.sprayGeo.setAttribute(
      "color",
      new THREE.BufferAttribute(new Float32Array(this.maxDroplets * 3), 3)
    );
    this.sprayMat = new THREE.PointsMaterial({
      size: 0.056,
      map: this.spriteTex || undefined,
      vertexColors: true,
      transparent: true,
      opacity: 0.85,
      alphaTest: 0.02,
      depthWrite: false,
    });
    this.sprayPoints = new THREE.Points(this.sprayGeo, this.sprayMat);
    this.group.add(this.sprayPoints);
  }

  setVisible(visible: boolean): void {
    this.group.visible = visible;
  }

  configure(
    primaryFlow: FlowSample[],
    motiveFlow: FlowSample[],
    diluterFlow: FlowSample[],
    sprayZones: SprayZone[]
  ): void {
    this.primaryFlow = primaryFlow;
    this.motiveFlow = motiveFlow;
    this.diluterFlow = diluterFlow;
    this.sprayZones = sprayZones;

    const hasMotive = motiveFlow.length > 1;
    const hasDiluter = diluterFlow.length > 1;
    const motiveQuota = hasMotive ? 64 : 0;
    const diluterQuota = hasDiluter ? 34 : 0;
    const primaryQuota = this.gasCount - motiveQuota - diluterQuota;

    this.gasParticles = [];
    for (let i = 0; i < this.gasCount; i++) {
      let stream: "primary" | "motive" | "diluter" = "primary";
      let localIdx = i;
      let totalInStream = primaryQuota;
      if (i >= primaryQuota && i < primaryQuota + motiveQuota) {
        stream = "motive";
        localIdx = i - primaryQuota;
        totalInStream = motiveQuota;
      } else if (i >= primaryQuota + motiveQuota) {
        stream = "diluter";
        localIdx = i - primaryQuota - motiveQuota;
        totalInStream = diluterQuota;
      }

      // Deterministic quasi-random radial/angular spread
      const h1 = ((i * 2654435761) >>> 0) / 4294967295;
      const h2 = (((i + 37) * 2246822519) >>> 0) / 4294967295;
      this.gasParticles.push({
        stream,
        t: localIdx / Math.max(1, totalInStream),
        radialFrac: 0.15 + Math.sqrt(h1) * 0.82,
        angleOffset: h2 * Math.PI * 2,
        swirlRate: 18 + (i % 7) * 2.5,
      });
    }

    this.droplets = [];
    const colAttr = this.sprayGeo.attributes.color as THREE.BufferAttribute;
    let dIdx = 0;
    const tmpCol = new THREE.Color();
    for (let z = 0; z < sprayZones.length && dIdx < this.maxDroplets; z++) {
      const zone = sprayZones[z];
      tmpCol.setHex(zone.color);
      const count = Math.min(zone.count, this.maxDroplets - dIdx);
      for (let k = 0; k < count; k++) {
        const h1 = (((dIdx + 11) * 2654435761) >>> 0) / 4294967295;
        const h2 = (((dIdx + 53) * 2246822519) >>> 0) / 4294967295;
        this.droplets.push({
          zoneIdx: z,
          t: k / Math.max(1, count),
          radialFrac: Math.sqrt(h1),
          angle: h2 * Math.PI * 2,
        });
        colAttr.setXYZ(dIdx, tmpCol.r, tmpCol.g, tmpCol.b);
        dIdx++;
      }
    }
    this.sprayGeo.setDrawRange(0, this.droplets.length);
    colAttr.needsUpdate = true;
  }

  update(dt: number): void {
    if (!this.group.visible) return;

    const posAttr = this.gasGeo.attributes.position as THREE.BufferAttribute;
    const colAttr = this.gasGeo.attributes.color as THREE.BufferAttribute;

    const baseRatePrimary = 0.135;
    const baseRateMotive = 0.24;
    const baseRateDiluter = 0.26;

    for (let i = 0; i < this.gasParticles.length; i++) {
      const p = this.gasParticles[i];
      const samples =
        p.stream === "motive"
          ? this.motiveFlow
          : p.stream === "diluter"
            ? this.diluterFlow
            : this.primaryFlow;
      if (samples.length < 2) {
        posAttr.setXYZ(i, 0, -999, 0);
        continue;
      }

      const maxIdx = samples.length - 1;
      const curIdx = Math.min(maxIdx - 1, Math.floor(p.t * maxIdx));
      const localSpeed = samples[curIdx].speedMul;
      const baseRate =
        p.stream === "motive"
          ? baseRateMotive
          : p.stream === "diluter"
            ? baseRateDiluter
            : baseRatePrimary;

      p.t = (p.t + dt * baseRate * localSpeed) % 1.0;

      const fIdx = p.t * maxIdx;
      const i0 = Math.min(maxIdx - 1, Math.floor(fIdx));
      const i1 = i0 + 1;
      const frac = fIdx - i0;
      const s0 = samples[i0];
      const s1 = samples[i1];

      // Tangent vector to build perpendicular swirl disk (keeps swirl strictly inside duct walls!)
      const tx = s1.pt.x - s0.pt.x;
      const ty = s1.pt.y - s0.pt.y;
      const tz = s1.pt.z - s0.pt.z;
      const tLen = Math.hypot(tx, ty, tz) || 1;
      const ntx = tx / tLen;
      const nty = ty / tLen;
      const ntz = tz / tLen;

      // Orthogonal basis (u, v) perpendicular to tangent
      let ux: number, uy: number, uz: number;
      if (Math.abs(nty) < 0.9) {
        // cross(t, up(0,1,0)) = (-ntz, 0, ntx)
        const l = Math.hypot(-ntz, ntx) || 1;
        ux = -ntz / l;
        uy = 0;
        uz = ntx / l;
      } else {
        // cross(t, right(1,0,0)) = (0, ntz, -nty)
        const l = Math.hypot(ntz, -nty) || 1;
        ux = 0;
        uy = ntz / l;
        uz = -nty / l;
      }
      // v = cross(t, u)
      const vx = nty * uz - ntz * uy;
      const vy = ntz * ux - ntx * uz;
      const vz = ntx * uy - nty * ux;

      const r = (s0.swirlRadius + (s1.swirlRadius - s0.swirlRadius) * frac) * p.radialFrac;
      const ang = p.t * p.swirlRate + p.angleOffset;
      const ca = Math.cos(ang) * r;
      const sa = Math.sin(ang) * r;

      const cx = s0.pt.x + tx * frac + ux * ca + vx * sa;
      const cy = s0.pt.y + ty * frac + uy * ca + vy * sa;
      const cz = s0.pt.z + tz * frac + uz * ca + vz * sa;
      posAttr.setXYZ(i, cx, cy, cz);

      const c0 = STAGE_COLORS[s0.stage];
      const c1 = STAGE_COLORS[s1.stage];
      colAttr.setXYZ(
        i,
        c0[0] + (c1[0] - c0[0]) * frac,
        c0[1] + (c1[1] - c0[1]) * frac,
        c0[2] + (c1[2] - c0[2]) * frac
      );
    }
    posAttr.needsUpdate = true;
    colAttr.needsUpdate = true;

    // Update falling water spray droplets
    if (this.droplets.length > 0) {
      const sPos = this.sprayGeo.attributes.position as THREE.BufferAttribute;
      for (let i = 0; i < this.droplets.length; i++) {
        const d = this.droplets[i];
        const zone = this.sprayZones[d.zoneIdx];
        if (!zone) continue;
        d.t = (d.t + dt * zone.speed) % 1.0;
        const r =
          (zone.topRadius + (zone.botRadius - zone.topRadius) * d.t) *
          d.radialFrac;
        const y = zone.topCenter.y - d.t * zone.height;
        const x = zone.topCenter.x + Math.cos(d.angle) * r;
        const z = zone.topCenter.z + Math.sin(d.angle) * r;
        sPos.setXYZ(i, x, y, z);
      }
      sPos.needsUpdate = true;
    }
  }

  dispose(): void {
    this.gasGeo.dispose();
    this.gasMat.dispose();
    this.sprayGeo.dispose();
    this.sprayMat.dispose();
    this.spriteTex?.dispose();
  }
}
