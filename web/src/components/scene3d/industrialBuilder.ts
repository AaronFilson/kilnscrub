import * as THREE from "three";
import type { MaterialBundle } from "./materials";
import {
  addBallValve,
  addBoltedFlange,
  addBypassDamperTee,
  addDimensionCallout3D,
  addPipeUnion,
  addTriClamp,
  buildImpellerGroup,
  buildRectToRoundGeometry,
  clamp,
  computeVoluteProfiles,
  fmtDimMeters,
  makeArc90Bezier,
} from "./primitives";
import type {
  FlowSample,
  FlowStage,
  SceneBuildOptions,
  SceneBuildResult,
  SprayZone,
  StageKey,
  StageView,
} from "./types";

const vUp = new THREE.Vector3(0, 1, 0);
const vDn = new THREE.Vector3(0, -1, 0);
const vRight = new THREE.Vector3(1, 0, 0);
const vLeft = new THREE.Vector3(-1, 0, 0);
const vFront = new THREE.Vector3(0, 0, 1);
const vBack = new THREE.Vector3(0, 0, -1);

function sampleCurve(
  curve: THREE.Curve<THREE.Vector3>,
  count: number,
  swirlRadius: number,
  speedMul: number,
  stage: FlowStage,
  ox: number,
  oz: number
): FlowSample[] {
  const pts = curve.getSpacedPoints(Math.max(2, count));
  return pts.map((p) => ({
    pt: new THREE.Vector3(p.x + ox, p.y, p.z + oz),
    swirlRadius,
    speedMul,
    stage,
  }));
}

interface SystemsOverlayConfig {
  qX: number;
  qZ: number;
  qR: number;
  qH: number;
  pX: number;
  pZ: number;
  pR: number;
  pH: number;
  vX: number;
  vY: number;
  vZ: number;
  bX: number;
  bY: number;
  bZ: number;
  bR: number;
  bDepth: number;
  stX: number;
  stBaseY: number;
  stH: number;
  stR: number;
  stZ: number;
  inletY: number;
  ductR: number;
}

export function buildIndustrialAssembly(
  assemblyGroup: THREE.Group,
  mats: MaterialBundle,
  opts: SceneBuildOptions
): SceneBuildResult {
  const {
    evaluation,
    cutaway: cut,
    showSystemsOverlay,
    showDimensions,
    units,
    cameraFovDeg,
    cameraAspect,
  } = opts;
  const { project, design, architecture_options: arch } = evaluation;
  const q = design.quench;
  const v = design.venturi;
  const p = design.packed_bed;
  const b = design.blower;
  const env = design.envelope;

  const isEductor = project.stage_alignment === "eductor_dry_fan";
  const isParallelBed = project.packed_bed_layout === "parallel_manifold";
  const isParallelVent = project.venturi_layout === "parallel_mini_venturis";
  const isCompact = !!project.compact_footprint;
  const hasDiluter = !!project.include_plume_diluter;
  const isDeluxe = project.build_path === "custom_fab_deluxe";
  const showSystems = showSystemsOverlay;
  const blowerDrive = String(project.blower_drive_option || "industrial_radial_vfd").toLowerCase();

  const thetaStart = cut ? Math.PI * 0.5 : 0;
  const thetaLen = cut ? Math.PI : Math.PI * 2;
  const disposableTextures: THREE.Texture[] = [];

  // True 1:1 meter engineering dimensions from Go physics engine
  const ductR = clamp((env.duct_diameter_m || 0.178) * 0.5, 0.052, 0.16);
  const qR = clamp((q.tower_diameter_m || 0.35) * 0.5, 0.14, 0.65);
  const qH = clamp(q.tower_height_m || 2.1, 1.0, 3.2);

  const vInR = clamp((v.inlet_diameter_m || 0.18) * 0.5, 0.06, 0.22);
  const vThR = clamp((v.throat_diameter_m || 0.05) * 0.5, 0.026, vInR * 0.72);
  const vOutR = clamp((v.outlet_diameter_m || 0.20) * 0.5, 0.065, 0.24);
  const vConvL = clamp(v.converging_length_m || 0.28, 0.18, 0.65);
  const vThL = clamp(v.throat_length_m || 0.16, 0.10, 0.36);
  const vDivL = clamp(v.diverging_length_m || 0.52, 0.28, 0.95);

  const basePR = clamp((p.column_diameter_m || 0.30) * 0.5, 0.16, 0.75);
  const pR = isEductor ? clamp(basePR * 1.18, 0.22, 0.85) : basePR;
  const pH = clamp(p.column_height_m || 2.75, 1.45, 3.8);
  const packH = clamp(p.packing_height_m || 1.05, 0.35, pH * 0.52);
  const demH = clamp(p.demister_height_m || 0.18, 0.08, 0.32);
  const sumpH = clamp(pH * 0.12, 0.16, 0.28);

  let impellerMesh: THREE.Object3D | null = null;
  let diluterImpellerMesh: THREE.Object3D | null = null;
  const sprayZones: SprayZone[] = [];

  // =========================================================================
  // HELPER: STAGE 1 QUENCH TOWER + NFPA 86 200°F FAIL-OPEN BYPASS TEE
  // =========================================================================
  const buildQuenchAt = (qx: number, qz: number, inletY: number) => {
    const qGroup = new THREE.Group();
    const qShell = new THREE.Mesh(
      new THREE.CylinderGeometry(qR, qR, qH, 36, 1, cut, thetaStart, thetaLen),
      mats.ss
    );
    qShell.position.set(qx, qH * 0.5, qz);
    qGroup.add(qShell);

    // Hot inlet flue + NFPA 86 200°F fail-open bypass damper tee
    const inletLen = 0.78;
    const inletDuct = new THREE.Mesh(
      new THREE.CylinderGeometry(ductR, ductR, inletLen + 0.016, 24, 1, true),
      mats.hotSolid
    );
    inletDuct.rotation.z = Math.PI / 2;
    inletDuct.position.set(qx - qR - inletLen * 0.5 + 0.008, inletY, qz);
    qGroup.add(inletDuct);
    addBoltedFlange(qGroup, qx - qR - inletLen + 0.015, inletY, qz, "x", ductR, mats.flange, mats.bolt, mats.gasket, 8);
    addBoltedFlange(qGroup, qx - qR - 0.012, inletY, qz, "x", ductR, mats.flange, mats.bolt, mats.gasket, 8);

    // Real NFPA 86 200°F fail-open bypass tee centered on inlet flue
    addBypassDamperTee(
      qGroup,
      qx - qR - inletLen * 0.50,
      inletY,
      qz,
      ductR,
      mats.hot,
      mats.flange,
      mats.bolt,
      mats.gasket,
      mats.safetyRed,
      mats.darkSteel,
      cut
    );

    // Top bolted blind flange + base ring + conical bottom drain boot
    addBoltedFlange(qGroup, qx, qH, qz, "y", qR, mats.flange, mats.bolt, mats.gasket, 12);
    const qTop = new THREE.Mesh(
      new THREE.CylinderGeometry(qR * 1.06, qR * 1.06, 0.032, 36),
      mats.flange
    );
    qTop.position.set(qx, qH + 0.028, qz);
    qGroup.add(qTop);

    const qBase = new THREE.Mesh(
      new THREE.CylinderGeometry(qR * 1.12, qR * 1.16, 0.06, 36),
      mats.darkSteel
    );
    qBase.position.set(qx, 0.03, qz);
    qGroup.add(qBase);

    // Top spray water manifold + multi-nozzle lances (driven by design.quench.nozzle_count)
    const nozCount = clamp(q.nozzle_count || 2, 1, 4);
    const lanceTopY = qH - 0.14;
    const lanceBotY = qH - 0.34;
    const headerBar = new THREE.Mesh(
      new THREE.CylinderGeometry(0.018, 0.018, 0.28, 12),
      mats.ssSolid
    );
    headerBar.position.set(qx, qH - 0.08, qz);
    qGroup.add(headerBar);
    addTriClamp(qGroup, qx, qH + 0.055, qz, "y", 0.022, mats.bolt, mats.gasket);

    for (let ni = 0; ni < nozCount; ni++) {
      const ang = (ni / nozCount) * Math.PI * 2;
      const offR = nozCount === 1 ? 0 : qR * 0.28;
      const nx = qx + Math.cos(ang) * offR;
      const nz = qz + Math.sin(ang) * offR;
      const tip = new THREE.Mesh(
        new THREE.CylinderGeometry(0.022, 0.015, 0.055, 12),
        mats.brass
      );
      tip.position.set(nx, lanceBotY, nz);
      qGroup.add(tip);
    }

    // Internal mist cone + quench sump pool + outlet RTD thermowell probe
    const mistCone = new THREE.Mesh(
      new THREE.ConeGeometry(qR * 0.80, Math.max(0.45, qH - 0.48), 24, 1, true),
      mats.sprayMist
    );
    mistCone.position.set(qx, (lanceBotY + 0.12) * 0.5, qz);
    qGroup.add(mistCone);

    const qPool = new THREE.Mesh(
      new THREE.CylinderGeometry(qR * 0.92, qR * 0.92, 0.16, 26),
      mats.sumpWater
    );
    qPool.position.set(qx, 0.14, qz);
    qGroup.add(qPool);

    // 200°F RTD & snap-disc probe protruding from lower quench wall
    const rtdBoss = new THREE.Mesh(
      new THREE.CylinderGeometry(0.016, 0.016, 0.10, 10),
      mats.brass
    );
    rtdBoss.rotation.x = Math.PI / 2;
    rtdBoss.position.set(qx + qR * 0.48, 0.42, qz + qR * 0.90);
    qGroup.add(rtdBoss);

    assemblyGroup.add(qGroup);
    return { inletLen, sprayTopY: lanceTopY - 0.20 };
  };

  // =========================================================================
  // HELPER: STAGE 3 PACKED BED (SINGLE COLUMN OR SHELL-LESS UNISEAL BUNDLE)
  // =========================================================================
  const buildPackedBedAt = (cx: number, cz: number, entryY: number) => {
    const g = new THREE.Group();
    g.position.set(cx, 0, cz);
    const bundleR = isParallelBed ? Math.max(pR * 1.24, 0.34) : pR;
    const mediaOpt = String(project.packing_media_option || "pp_pall_rings").toLowerCase();
    const matMediaFill =
      mediaOpt === "diy_ceramic_rings"
        ? mats.packingCeramic
        : mediaOpt === "aquaculture_bioballs"
          ? mats.packingBio
          : mats.packingPP;

    let distRingY = pH * 0.72;
    let sprayBedH = packH;

    if (isParallelBed) {
      const pipeNomIn = [2, 3, 4].includes(Math.round(arch.parallel_bed.pipe_nominal_in || project.parallel_pipe_size_in || 4))
        ? Math.round(arch.parallel_bed.pipe_nominal_in || project.parallel_pipe_size_in || 4)
        : 4;
      const rawCount = arch.parallel_bed.pipe_count || 7;

      const botPlenH = Math.max(sumpH + vOutR * 1.85 + 0.10, 0.46);
      const topPlenH = Math.max(demH + 0.18, 0.34);
      const plateT = 0.038;
      const pipeLen = Math.max(0.90, pH - botPlenH - topPlenH);

      const botPlen = new THREE.Mesh(
        new THREE.CylinderGeometry(bundleR, bundleR, botPlenH, 36, 1, cut, thetaStart, thetaLen),
        mats.cpvc
      );
      botPlen.position.y = botPlenH * 0.5;
      g.add(botPlen);

      const sumpM = new THREE.Mesh(
        new THREE.CylinderGeometry(bundleR * 0.94, bundleR * 0.94, sumpH, 28),
        mats.sumpWater
      );
      sumpM.position.y = sumpH * 0.5 + 0.04;
      g.add(sumpM);

      const botPlate = new THREE.Mesh(
        new THREE.CylinderGeometry(bundleR * 1.08, bundleR * 1.08, plateT, 36),
        mats.collectorPlate
      );
      botPlate.position.y = botPlenH + plateT * 0.5;
      g.add(botPlate);

      const topPlateY = pH - topPlenH - plateT * 0.5;
      const topPlate = new THREE.Mesh(
        new THREE.CylinderGeometry(bundleR * 1.08, bundleR * 1.08, plateT, 36),
        mats.collectorPlate
      );
      topPlate.position.y = topPlateY;
      g.add(topPlate);

      const topPlen = new THREE.Mesh(
        new THREE.CylinderGeometry(bundleR, bundleR, topPlenH, 36, 1, cut, thetaStart, thetaLen),
        mats.cpvc
      );
      topPlen.position.y = pH - topPlenH * 0.5;
      g.add(topPlen);

      const demM = new THREE.Mesh(
        new THREE.CylinderGeometry(bundleR * 0.93, bundleR * 0.93, demH, 28),
        mats.demister
      );
      demM.position.y = pH - demH * 0.5 - 0.06;
      g.add(demM);

      // 4 Stainless threaded tie-rods + hex nuts clamping flat collector plates
      const rodR = bundleR * 1.02;
      const rodSpan = pH - topPlenH - botPlenH;
      for (let rIdx = 0; rIdx < 4; rIdx++) {
        const rAng = (rIdx / 4) * Math.PI * 2 + Math.PI * 0.25;
        const rod = new THREE.Mesh(
          new THREE.CylinderGeometry(0.009, 0.009, rodSpan + 0.12, 8),
          mats.bolt
        );
        rod.position.set(Math.cos(rAng) * rodR, botPlenH + rodSpan * 0.5, Math.sin(rAng) * rodR);
        g.add(rod);
      }

      const pipeR =
        pipeNomIn === 2
          ? clamp(bundleR * 0.095, 0.024, 0.036)
          : pipeNomIn === 3
            ? clamp(bundleR * 0.135, 0.035, 0.049)
            : clamp(bundleR * 0.175, 0.046, 0.062);

      const positions: Array<[number, number]> = [[0, 0]];
      if (pipeNomIn === 2) {
        const innerN = 6;
        const outerN = Math.min(12, Math.max(8, rawCount - 7));
        for (let k = 0; k < innerN; k++) {
          const ang = (k / innerN) * Math.PI * 2;
          positions.push([Math.cos(ang) * bundleR * 0.36, Math.sin(ang) * bundleR * 0.36]);
        }
        for (let k = 0; k < outerN; k++) {
          const ang = (k / outerN) * Math.PI * 2 + 0.25;
          positions.push([Math.cos(ang) * bundleR * 0.76, Math.sin(ang) * bundleR * 0.76]);
        }
      } else if (pipeNomIn === 3) {
        const innerN = 5;
        const outerN = Math.min(10, Math.max(7, rawCount - 6));
        for (let k = 0; k < innerN; k++) {
          const ang = (k / innerN) * Math.PI * 2;
          positions.push([Math.cos(ang) * bundleR * 0.38, Math.sin(ang) * bundleR * 0.38]);
        }
        for (let k = 0; k < outerN; k++) {
          const ang = (k / outerN) * Math.PI * 2 + 0.3;
          positions.push([Math.cos(ang) * bundleR * 0.76, Math.sin(ang) * bundleR * 0.76]);
        }
      } else {
        const outerN = Math.max(6, Math.min(10, rawCount - 1));
        const ringR = bundleR * 0.66;
        for (let k = 0; k < outerN; k++) {
          const ang = (k / outerN) * Math.PI * 2;
          positions.push([Math.cos(ang) * ringR, Math.sin(ang) * ringR]);
        }
      }

      const pipeCenterY = botPlenH + pipeLen * 0.5;
      for (const [px, pz] of positions) {
        const cutThis = cut && pz > 0.02;
        const pTube = new THREE.Mesh(
          new THREE.CylinderGeometry(
            pipeR,
            pipeR,
            pipeLen + 0.04,
            18,
            1,
            cutThis,
            cutThis ? Math.PI * 0.5 : 0,
            cutThis ? Math.PI : Math.PI * 2
          ),
          mats.cpvc
        );
        pTube.position.set(px, pipeCenterY, pz);
        g.add(pTube);
        if (cutThis) {
          const pFill = new THREE.Mesh(
            new THREE.CylinderGeometry(pipeR * 0.84, pipeR * 0.84, pipeLen * 0.84, 14),
            matMediaFill
          );
          pFill.position.set(px, pipeCenterY, pz);
          g.add(pFill);
        }
        const sealBot = new THREE.Mesh(
          new THREE.CylinderGeometry(pipeR * 1.24, pipeR * 1.28, 0.042, 16),
          mats.rubberSeal
        );
        sealBot.position.set(px, botPlenH + plateT + 0.012, pz);
        g.add(sealBot);
        const sealTop = new THREE.Mesh(
          new THREE.CylinderGeometry(pipeR * 1.28, pipeR * 1.24, 0.042, 16),
          mats.rubberSeal
        );
        sealTop.position.set(px, pH - topPlenH - plateT - 0.012, pz);
        g.add(sealTop);
      }

      distRingY = pH - topPlenH + 0.04;
      sprayBedH = pipeLen * 0.88;
      const distH = new THREE.Mesh(
        new THREE.TorusGeometry(bundleR * 0.62, 0.016, 10, 28),
        mats.darkSteel
      );
      distH.rotation.x = Math.PI / 2;
      distH.position.y = distRingY;
      g.add(distH);
    } else {
      // Classic Single Large Column
      const pShell = new THREE.Mesh(
        new THREE.CylinderGeometry(bundleR, bundleR, pH, 36, 1, cut, thetaStart, thetaLen),
        mats.cpvc
      );
      pShell.position.y = pH * 0.5;
      g.add(pShell);

      // Mid-column bolted inspection flange ring
      addBoltedFlange(g, 0, pH * 0.55, 0, "y", bundleR, mats.cpvcFlange, mats.bolt, mats.gasket, 12);

      const sumpMesh = new THREE.Mesh(
        new THREE.CylinderGeometry(bundleR * 0.94, bundleR * 0.94, sumpH, 28),
        mats.sumpWater
      );
      sumpMesh.position.y = sumpH * 0.5 + 0.04;
      g.add(sumpMesh);

      const packBaseY = Math.max(entryY + vOutR + 0.12, pH * 0.28);
      const clampedPackH = Math.max(0.28, Math.min(packH, pH - packBaseY - demH - 0.30));
      const supportPlate = new THREE.Mesh(
        new THREE.CylinderGeometry(bundleR * 0.94, bundleR * 0.94, 0.025, 28),
        mats.darkSteel
      );
      supportPlate.position.y = packBaseY - 0.013;
      g.add(supportPlate);

      const packMesh = new THREE.Mesh(
        new THREE.CylinderGeometry(bundleR * 0.91, bundleR * 0.91, clampedPackH, 28),
        matMediaFill
      );
      packMesh.position.y = packBaseY + clampedPackH * 0.5;
      g.add(packMesh);

      distRingY = packBaseY + clampedPackH + 0.10;
      sprayBedH = clampedPackH + 0.10;
      const distRing = new THREE.Mesh(
        new THREE.TorusGeometry(bundleR * 0.64, 0.018, 10, 24),
        mats.darkSteel
      );
      distRing.rotation.x = Math.PI / 2;
      distRing.position.y = distRingY;
      g.add(distRing);

      const demY = pH - demH * 0.5 - 0.14;
      const demMesh = new THREE.Mesh(
        new THREE.CylinderGeometry(bundleR * 0.93, bundleR * 0.93, demH, 28),
        mats.demister
      );
      demMesh.position.y = demY;
      g.add(demMesh);
    }

    // Base plinth + top bolted flange + smooth conical reducer
    const pBase = new THREE.Mesh(
      new THREE.CylinderGeometry(bundleR * 1.10, bundleR * 1.14, 0.06, 36),
      mats.darkSteel
    );
    pBase.position.y = 0.03;
    g.add(pBase);
    addBoltedFlange(g, 0, pH, 0, "y", bundleR, mats.cpvcFlange, mats.bolt, mats.gasket, 12);

    const topConeH = clamp(bundleR * 0.90, 0.22, 0.38);
    const pReducer = new THREE.Mesh(
      new THREE.CylinderGeometry(ductR, bundleR, topConeH, 36, 1, cut, thetaStart, thetaLen),
      mats.cpvc
    );
    pReducer.position.y = pH + topConeH * 0.5;
    g.add(pReducer);

    const topNeckY = pH + topConeH;
    addBoltedFlange(g, 0, topNeckY, 0, "y", ductR, mats.cpvcFlange, mats.bolt, mats.gasket, 8);
    assemblyGroup.add(g);
    return { bundleR, topNeckY, distRingY, sprayBedH };
  };

  // =========================================================================
  // HELPER: EXHAUST STACK + OPTIONAL 120V SQUIRREL-CAGE PLUME DILUTER TEE
  // =========================================================================
  const addStackAndDiluter = (
    parent: THREE.Object3D,
    sx: number,
    baseY: number,
    sz: number,
    stackR: number,
    stackH: number
  ) => {
    const stackMesh = new THREE.Mesh(
      new THREE.CylinderGeometry(stackR, stackR, stackH, 28, 1, true),
      mats.blower
    );
    stackMesh.position.set(sx, baseY + stackH * 0.5, sz);
    parent.add(stackMesh);
    let exitY = baseY + stackH;
    let extraRight = 0;
    let diluterEntry: THREE.Vector3 | null = null;
    let diluterTeeY = baseY + stackH * 0.42;

    if (hasDiluter) {
      diluterTeeY = baseY + stackH * 0.42;
      const dilR = stackR * 0.72;
      const teeLen = stackR + 0.26;
      const teeStub = new THREE.Mesh(
        new THREE.CylinderGeometry(dilR, dilR, teeLen, 20, 1, true),
        mats.diluter
      );
      teeStub.rotation.z = Math.PI / 2;
      teeStub.position.set(sx + teeLen * 0.5, diluterTeeY, sz);
      parent.add(teeStub);
      addBoltedFlange(parent, sx + stackR + 0.04, diluterTeeY, sz, "x", dilR, mats.flange, mats.bolt, mats.gasket, 6);
      addBoltedFlange(parent, sx + teeLen - 0.02, diluterTeeY, sz, "x", dilR, mats.flange, mats.bolt, mats.gasket, 6);

      const sqR = Math.max(0.11, dilR * 1.55);
      const sqX = sx + teeLen + sqR * 0.75;
      const sqY = diluterTeeY - sqR * 0.15;
      const sqHousing = new THREE.Mesh(
        new THREE.CylinderGeometry(sqR, sqR, dilR * 1.7, 24),
        mats.diluter
      );
      sqHousing.rotation.x = Math.PI / 2;
      sqHousing.position.set(sqX, sqY, sz);
      parent.add(sqHousing);

      const sqIntake = new THREE.Mesh(
        new THREE.TorusGeometry(sqR * 0.56, 0.012, 8, 20),
        mats.darkSteel
      );
      sqIntake.position.set(sqX, sqY, sz + dilR * 0.88);
      parent.add(sqIntake);

      // Tiny spinning squirrel-cage wheel visible in intake
      const dWheel = new THREE.Group();
      dWheel.position.set(sqX, sqY, sz + dilR * 0.85);
      for (let i = 0; i < 6; i++) {
        const bMesh = new THREE.Mesh(
          new THREE.BoxGeometry(sqR * 0.9, 0.014, 0.014),
          mats.blade
        );
        bMesh.rotation.z = (i * Math.PI) / 6;
        dWheel.add(bMesh);
      }
      parent.add(dWheel);
      diluterImpellerMesh = dWheel;

      diluterEntry = new THREE.Vector3(sqX, sqY, sz + dilR * 1.05);
      extraRight = teeLen + sqR * 1.75;

      const hvH = 0.26;
      addBoltedFlange(parent, sx, baseY + stackH, sz, "y", stackR, mats.blowerFlange, mats.bolt, mats.gasket, 8);
      const hvCone = new THREE.Mesh(
        new THREE.CylinderGeometry(stackR * 0.64, stackR, hvH, 28, 1, true),
        mats.diluter
      );
      hvCone.position.set(sx, baseY + stackH + hvH * 0.5, sz);
      parent.add(hvCone);
      const hvLip = new THREE.Mesh(
        new THREE.TorusGeometry(stackR * 0.64, 0.010, 10, 28),
        mats.blowerFlange
      );
      hvLip.rotation.x = Math.PI / 2;
      hvLip.position.set(sx, baseY + stackH + hvH, sz);
      parent.add(hvLip);
      exitY += hvH;
    } else {
      const exitLip = new THREE.Mesh(
        new THREE.TorusGeometry(stackR, 0.012, 10, 28),
        mats.blowerFlange
      );
      exitLip.rotation.x = Math.PI / 2;
      exitLip.position.set(sx, baseY + stackH, sz);
      parent.add(exitLip);
    }
    return { exitY, extraRight, diluterEntry, diluterTeeY };
  };

  // =========================================================================
  // HELPER: PHASE 9 SYSTEMS & DELUXE OVERLAY (SKID, K-BRACING, PUMP, PANEL, BAFFLE)
  // =========================================================================
  const addSystemsOverlay3D = (cfg: SystemsOverlayConfig) => {
    if (!showSystems) return;
    const sysG = new THREE.Group();
    const {
      qX: sqX, qZ: sqZ, pX: spX, pZ: spZ, pR: spR, pH: spH,
      vX: svX, vY: svY, vZ: svZ,
      bX: sbX, bY: sbY, bZ: sbZ, bR: sbR, bDepth: sbDepth,
      stX: sstX, stBaseY: sstBaseY, stH: sstH, stR: sstR, stZ: sstZ,
      inletY: sInletY,
    } = cfg;

    // 1. Structural Steel Base Skid Frame + Anchor Pads
    const curBox = new THREE.Box3().setFromObject(assemblyGroup);
    const skidMinX = curBox.min.x - 0.18;
    const skidMaxX = curBox.max.x + 0.56;
    const skidMinZ = Math.min(curBox.min.z, -0.58) - 0.16;
    const skidMaxZ = Math.max(curBox.max.z, 0.72) + 0.28;
    const skidW = skidMaxX - skidMinX;
    const skidD = skidMaxZ - skidMinZ;
    const skidCX = (skidMinX + skidMaxX) * 0.5;
    const skidCZ = (skidMinZ + skidMaxZ) * 0.5;
    const beamH = 0.068;
    const beamW = 0.068;

    const railFront = new THREE.Mesh(new THREE.BoxGeometry(skidW, beamH, beamW), mats.skidSteel);
    railFront.position.set(skidCX, beamH * 0.5, skidMaxZ - beamW * 0.5);
    sysG.add(railFront);
    const railBack = new THREE.Mesh(new THREE.BoxGeometry(skidW, beamH, beamW), mats.skidSteel);
    railBack.position.set(skidCX, beamH * 0.5, skidMinZ + beamW * 0.5);
    sysG.add(railBack);
    const railLeft = new THREE.Mesh(new THREE.BoxGeometry(beamW, beamH, skidD), mats.skidSteel);
    railLeft.position.set(skidMinX + beamW * 0.5, beamH * 0.5, skidCZ);
    sysG.add(railLeft);
    const railRight = new THREE.Mesh(new THREE.BoxGeometry(beamW, beamH, skidD), mats.skidSteel);
    railRight.position.set(skidMaxX - beamW * 0.5, beamH * 0.5, skidCZ);
    sysG.add(railRight);

    for (const frac of [0.28, 0.54, 0.78]) {
      const cx = skidMinX + skidW * frac;
      const cross = new THREE.Mesh(
        new THREE.BoxGeometry(beamW * 0.9, beamH * 0.9, skidD - beamW),
        mats.skidSteel
      );
      cross.position.set(cx, beamH * 0.45, skidCZ);
      sysG.add(cross);
    }

    const padPts: Array<[number, number]> = [
      [skidMinX + 0.06, skidMinZ + 0.06],
      [skidCX, skidMinZ + 0.06],
      [skidMaxX - 0.06, skidMinZ + 0.06],
      [skidMinX + 0.06, skidMaxZ - 0.06],
      [skidCX, skidMaxZ - 0.06],
      [skidMaxX - 0.06, skidMaxZ - 0.06],
    ];
    for (const [ax, az] of padPts) {
      const pad = new THREE.Mesh(new THREE.BoxGeometry(0.14, 0.022, 0.14), mats.braceSteel);
      pad.position.set(ax, 0.011, az);
      sysG.add(pad);
      const bolt = new THREE.Mesh(new THREE.CylinderGeometry(0.012, 0.012, 0.06, 6), mats.bolt);
      bolt.position.set(ax, 0.032, az);
      sysG.add(bolt);
    }

    // 2. Stage 3 Support Frame & Diagonal K-Bracing
    const frameR = spR + 0.10;
    const postH = spH * 0.88;
    const postGeo = new THREE.CylinderGeometry(0.022, 0.022, postH, 10);
    for (const [dx, dz] of [
      [-frameR, -frameR],
      [frameR, -frameR],
      [-frameR, frameR],
      [frameR, frameR],
    ]) {
      const post = new THREE.Mesh(postGeo, mats.skidSteel);
      post.position.set(spX + dx, beamH + postH * 0.5, spZ + dz);
      sysG.add(post);
    }
    for (const hFrac of [0.46, 0.92]) {
      const ty = beamH + postH * hFrac;
      const ring = new THREE.Mesh(
        new THREE.TorusGeometry(frameR * 1.15, 0.018, 8, 24),
        mats.braceSteel
      );
      ring.rotation.x = Math.PI / 2;
      ring.position.set(spX, ty, spZ);
      sysG.add(ring);
    }
    const addStrut = (p1: THREE.Vector3, p2: THREE.Vector3) => {
      const dir = new THREE.Vector3().subVectors(p2, p1);
      const m = new THREE.Mesh(
        new THREE.CylinderGeometry(0.014, 0.014, dir.length(), 8),
        mats.braceSteel
      );
      m.position.copy(p1).addScaledVector(dir, 0.5);
      m.quaternion.setFromUnitVectors(vUp, dir.normalize());
      sysG.add(m);
    };
    const y0 = beamH + 0.06;
    const yMid = beamH + postH * 0.46;
    const yTop = beamH + postH * 0.92;
    addStrut(new THREE.Vector3(spX - frameR, y0, spZ - frameR), new THREE.Vector3(spX, yMid, spZ - frameR));
    addStrut(new THREE.Vector3(spX + frameR, y0, spZ - frameR), new THREE.Vector3(spX, yMid, spZ - frameR));
    addStrut(new THREE.Vector3(spX - frameR, yTop, spZ - frameR), new THREE.Vector3(spX, yMid, spZ - frameR));
    addStrut(new THREE.Vector3(spX + frameR, yTop, spZ - frameR), new THREE.Vector3(spX, yMid, spZ - frameR));
    addStrut(new THREE.Vector3(spX - frameR, y0, spZ - frameR), new THREE.Vector3(spX - frameR, yMid, spZ + frameR));
    addStrut(new THREE.Vector3(spX - frameR, yMid, spZ + frameR), new THREE.Vector3(spX - frameR, yTop, spZ - frameR));
    addStrut(new THREE.Vector3(spX + frameR, y0, spZ - frameR), new THREE.Vector3(spX + frameR, yMid, spZ + frameR));
    addStrut(new THREE.Vector3(spX + frameR, yMid, spZ + frameR), new THREE.Vector3(spX + frameR, yTop, spZ - frameR));

    // 3. Sump Tank + Magnetic-Drive Pump + Y-Strainer + CPVC Ball Valves + Routed Hoses
    const sumpPosX = (sqX + spX) * 0.5 - 0.12;
    const sumpPosZ = skidMaxZ - 0.30;
    const sumpTankW = 0.72;
    const sumpTankH = 0.44;
    const sumpTankD = 0.40;
    const sumpShell = new THREE.Mesh(
      new THREE.BoxGeometry(sumpTankW, sumpTankH, sumpTankD),
      mats.cpvcSolid
    );
    sumpShell.position.set(sumpPosX, beamH + sumpTankH * 0.5, sumpPosZ);
    sysG.add(sumpShell);
    const sumpLiquid = new THREE.Mesh(
      new THREE.BoxGeometry(sumpTankW - 0.04, sumpTankH * 0.65, sumpTankD - 0.04),
      mats.sumpWater
    );
    sumpLiquid.position.set(sumpPosX, beamH + sumpTankH * 0.36, sumpPosZ);
    sysG.add(sumpLiquid);

    const pumpPosX = sumpPosX + sumpTankW * 0.5 + 0.25;
    const pumpPosZ = sumpPosZ;
    const pumpPad = new THREE.Mesh(
      new THREE.BoxGeometry(0.32, 0.03, 0.22),
      mats.acousticTrim
    );
    pumpPad.position.set(pumpPosX, beamH + 0.015, pumpPosZ);
    sysG.add(pumpPad);
    const pumpMotor = new THREE.Mesh(
      new THREE.CylinderGeometry(0.068, 0.068, 0.20, 16),
      mats.motor
    );
    pumpMotor.rotation.z = Math.PI / 2;
    pumpMotor.position.set(pumpPosX + 0.04, beamH + 0.11, pumpPosZ);
    sysG.add(pumpMotor);
    const pumpHead = new THREE.Mesh(
      new THREE.CylinderGeometry(0.075, 0.075, 0.07, 16),
      mats.hoseSupply
    );
    pumpHead.rotation.z = Math.PI / 2;
    pumpHead.position.set(pumpPosX - 0.08, beamH + 0.11, pumpPosZ);
    sysG.add(pumpHead);
    const yStrain = new THREE.Mesh(
      new THREE.CylinderGeometry(0.028, 0.028, 0.12, 12),
      mats.flange
    );
    yStrain.rotation.z = Math.PI / 4;
    yStrain.position.set(pumpPosX - 0.16, beamH + 0.09, pumpPosZ);
    sysG.add(yStrain);

    const addHose = (pts: THREE.Vector3[], rad: number, mat: THREE.Material) => {
      sysG.add(
        new THREE.Mesh(
          new THREE.TubeGeometry(new THREE.CatmullRomCurve3(pts), 28, rad, 10, false),
          mat
        )
      );
    };
    addHose(
      [
        new THREE.Vector3(sumpPosX + sumpTankW * 0.48, beamH + 0.11, pumpPosZ),
        new THREE.Vector3(pumpPosX - 0.08, beamH + 0.11, pumpPosZ),
      ],
      0.022,
      mats.hoseSupply
    );
    addBallValve(sysG, sumpPosX + sumpTankW * 0.56, beamH + 0.11, pumpPosZ, "x", 0.022, mats.cpvcSolid, mats.safetyRed);
    addPipeUnion(sysG, pumpPosX - 0.08, beamH + 0.24, pumpPosZ, "y", 0.020, mats.cpvcSolid, mats.darkSteel);

    addHose(
      [
        new THREE.Vector3(pumpPosX - 0.08, beamH + 0.18, pumpPosZ),
        new THREE.Vector3(pumpPosX - 0.08, beamH + 0.34, pumpPosZ - 0.10),
        new THREE.Vector3(sqX + cfg.qR * 0.4, cfg.qH * 0.72, sqZ + cfg.qR + 0.10),
        new THREE.Vector3(sqX, cfg.qH + 0.04, sqZ),
      ],
      0.017,
      mats.hoseSupply
    );
    addHose(
      [
        new THREE.Vector3(pumpPosX - 0.05, beamH + 0.18, pumpPosZ),
        new THREE.Vector3(svX, svY - 0.10, svZ + 0.14),
        new THREE.Vector3(svX, svY, svZ),
      ],
      0.014,
      mats.hoseSupply
    );
    addHose(
      [
        new THREE.Vector3(pumpPosX - 0.02, beamH + 0.18, pumpPosZ),
        new THREE.Vector3(spX - spR - 0.08, spH * 0.45, spZ + spR * 0.5),
        new THREE.Vector3(spX - spR - 0.06, spH - 0.16, spZ),
        new THREE.Vector3(spX - spR * 0.55, spH - 0.18, spZ),
      ],
      0.017,
      mats.hoseSupply
    );
    addHose(
      [
        new THREE.Vector3(sqX, beamH + 0.10, sqZ + cfg.qR * 0.7),
        new THREE.Vector3(sumpPosX - sumpTankW * 0.32, beamH + 0.24, sumpPosZ - sumpTankD * 0.3),
        new THREE.Vector3(sumpPosX - sumpTankW * 0.32, beamH + sumpTankH * 0.88, sumpPosZ),
      ],
      0.024,
      mats.hoseReturn
    );
    addHose(
      [
        new THREE.Vector3(spX, beamH + 0.10, spZ + spR * 0.7),
        new THREE.Vector3(sumpPosX + sumpTankW * 0.22, beamH + 0.24, sumpPosZ - sumpTankD * 0.25),
        new THREE.Vector3(sumpPosX + sumpTankW * 0.22, beamH + sumpTankH * 0.88, sumpPosZ),
      ],
      0.024,
      mats.hoseReturn
    );

    // 4. Stand-Mounted UL 508A NEMA 4X Control/VFD Panel + E-Stop + Disconnect + NEC 110.26 Zone
    const panelX = skidMaxX - 0.24;
    const panelZ = skidCZ + 0.08;
    const standH = 1.05;
    const boxW = 0.44;
    const boxH = 0.56;
    const boxD = 0.24;
    for (const dx of [-0.14, 0.14]) {
      const standPost = new THREE.Mesh(
        new THREE.BoxGeometry(0.05, standH, 0.05),
        mats.skidSteel
      );
      standPost.position.set(panelX + dx, beamH + standH * 0.5, panelZ);
      sysG.add(standPost);
    }
    const panelBody = new THREE.Mesh(
      new THREE.BoxGeometry(boxW, boxH, boxD),
      mats.panelBox
    );
    const panelMidY = beamH + standH + boxH * 0.38;
    panelBody.position.set(panelX, panelMidY, panelZ);
    sysG.add(panelBody);

    const vfdScreen = new THREE.Mesh(
      new THREE.BoxGeometry(0.16, 0.11, 0.015),
      mats.blowerDark
    );
    vfdScreen.position.set(panelX - 0.08, panelMidY + 0.11, panelZ + boxD * 0.5 + 0.006);
    sysG.add(vfdScreen);

    const eStop = new THREE.Mesh(
      new THREE.CylinderGeometry(0.032, 0.028, 0.04, 16),
      mats.safetyRed
    );
    eStop.rotation.x = Math.PI / 2;
    eStop.position.set(panelX - 0.10, panelMidY - 0.11, panelZ + boxD * 0.5 + 0.02);
    sysG.add(eStop);

    const discHandle = new THREE.Mesh(
      new THREE.BoxGeometry(0.04, 0.12, 0.045),
      mats.conduit
    );
    discHandle.position.set(panelX + 0.12, panelMidY + 0.04, panelZ + boxD * 0.5 + 0.02);
    sysG.add(discHandle);

    addHose(
      [
        new THREE.Vector3(panelX, panelMidY - boxH * 0.5, panelZ),
        new THREE.Vector3(panelX, beamH + 0.14, panelZ),
        new THREE.Vector3(sbX, beamH + 0.14, sbZ - sbDepth * 0.4),
        new THREE.Vector3(sbX, sbY, sbZ - sbDepth * 0.4),
      ],
      0.016,
      mats.conduit
    );

    const necPad = new THREE.Mesh(new THREE.PlaneGeometry(0.76, 0.91), mats.necZone);
    necPad.rotation.x = -Math.PI / 2;
    necPad.position.set(panelX, 0.008, panelZ + boxD * 0.5 + 0.48);
    sysG.add(necPad);

    // 5. Cutaway Acoustic Blower Baffle Box + Inline Stack Silencer
    const boxBW = sbR * 3.1 + 0.18;
    const boxBH = sbR * 2.9 + 0.14;
    const boxBD = Math.max(sbDepth + 0.24, sbR * 2.4);
    const bBoxGroup = new THREE.Group();
    bBoxGroup.position.set(sbX, Math.max(beamH + boxBH * 0.5, sbY), sbZ - sbDepth * 0.12);
    const wallT2 = 0.045;
    const bTop = new THREE.Mesh(new THREE.BoxGeometry(boxBW, wallT2, boxBD), mats.acousticBox);
    bTop.position.y = boxBH * 0.5;
    bBoxGroup.add(bTop);
    const bBack = new THREE.Mesh(new THREE.BoxGeometry(boxBW, boxBH, wallT2), mats.acousticBox);
    bBack.position.z = -boxBD * 0.5;
    bBoxGroup.add(bBack);
    const bLeft = new THREE.Mesh(new THREE.BoxGeometry(wallT2, boxBH, boxBD), mats.acousticBox);
    bLeft.position.x = -boxBW * 0.5;
    bBoxGroup.add(bLeft);
    const bRight = new THREE.Mesh(new THREE.BoxGeometry(wallT2, boxBH, boxBD), mats.acousticBox);
    bRight.position.x = boxBW * 0.5;
    bBoxGroup.add(bRight);
    bBoxGroup.add(
      new THREE.LineSegments(
        new THREE.EdgesGeometry(new THREE.BoxGeometry(boxBW, boxBH, boxBD)),
        new THREE.LineBasicMaterial({ color: 0xc084fc, transparent: true, opacity: 0.78 })
      )
    );
    sysG.add(bBoxGroup);

    const silLen = Math.min(sstH * 0.52, 0.56);
    const silR = sstR * 1.62;
    const silY = sstBaseY + silLen * 0.5 + 0.12;
    const silMesh = new THREE.Mesh(
      new THREE.CylinderGeometry(silR, silR, silLen, 28, 1, cut, thetaStart, thetaLen),
      mats.silencerBody
    );
    silMesh.position.set(sstX, silY, sstZ);
    sysG.add(silMesh);
    addBoltedFlange(sysG, sstX, silY - silLen * 0.5, sstZ, "y", sstR, mats.acousticTrim, mats.bolt, mats.gasket, 8);
    addBoltedFlange(sysG, sstX, silY + silLen * 0.5, sstZ, "y", sstR, mats.acousticTrim, mats.bolt, mats.gasket, 8);

    // 6. Deluxe Mode Vessel Hardware (Sight Glasses, NFPA 91 Cleanouts, Hot-Flue Jacket)
    if (isDeluxe) {
      const addSightGlass = (sx: number, sy: number, sz: number, rad: number) => {
        const bezel = new THREE.Mesh(
          new THREE.TorusGeometry(rad, 0.015, 10, 20),
          mats.flange
        );
        bezel.position.set(sx, sy, sz);
        sysG.add(bezel);
        const lens = new THREE.Mesh(
          new THREE.CylinderGeometry(rad * 0.92, rad * 0.92, 0.02, 20),
          mats.sightGlass
        );
        lens.rotation.x = Math.PI / 2;
        lens.position.set(sx, sy, sz);
        sysG.add(lens);
      };
      addSightGlass(sqX + cfg.qR * 0.65, cfg.qH * 0.62, sqZ + cfg.qR * 0.76, 0.062);
      addSightGlass(spX + spR * 0.65, spH * 0.56, spZ + spR * 0.76, 0.070);

      for (const [hx, hy, hz] of [
        [sqX, 0.28, sqZ + cfg.qR + 0.015],
        [spX, 0.34, spZ + spR + 0.015],
      ]) {
        const plate = new THREE.Mesh(
          new THREE.BoxGeometry(0.20, 0.16, 0.03),
          mats.conduit
        );
        plate.position.set(hx, hy, hz);
        sysG.add(plate);
      }

      const flueJacket = new THREE.Mesh(
        new THREE.CylinderGeometry(cfg.ductR * 1.30, cfg.ductR * 1.30, 0.34, 24),
        mats.flange
      );
      flueJacket.rotation.z = Math.PI / 2;
      flueJacket.position.set(sqX - cfg.qR - 0.22, sInletY, sqZ);
      sysG.add(flueJacket);

      // Phase 14 Deluxe Eco Upgrade 3D Hardware (zero geometry collisions!)
      if (project.deluxe_stage1_heat_recovery) {
        // Mount Stage 1 water-jacket collar strictly BELOW hot inlet flue nozzle (sInletY - cfg.ductR - 0.06)
        const maxTopY = Math.max(0.55, sInletY - cfg.ductR - 0.06);
        const s1JktH = clamp(cfg.qH * 0.30, 0.34, Math.max(0.36, maxTopY - 0.24));
        const s1CenterY = maxTopY - s1JktH * 0.5;
        const s1Jkt = new THREE.Mesh(
          new THREE.CylinderGeometry(cfg.qR * 1.16, cfg.qR * 1.16, s1JktH, 32, 1, cut, thetaStart, thetaLen),
          mats.steamerSS
        );
        s1Jkt.position.set(sqX, s1CenterY, sqZ);
        sysG.add(s1Jkt);
        const s1Band = new THREE.Mesh(
          new THREE.TorusGeometry(cfg.qR * 1.18, 0.015, 8, 28),
          mats.steamHeader
        );
        s1Band.rotation.x = Math.PI / 2;
        s1Band.position.set(sqX, s1CenterY + s1JktH * 0.42, sqZ);
        sysG.add(s1Band);
        // Hydronic ASME PRV & E3 flow-switch boss on side of jacket
        const jPrv = new THREE.Mesh(new THREE.CylinderGeometry(0.018, 0.018, 0.08, 10), mats.safetyRed);
        jPrv.position.set(sqX + cfg.qR * 1.18, s1CenterY + s1JktH * 0.46, sqZ);
        sysG.add(jPrv);
      }
      if (project.deluxe_variable_throat_vfd && !isEductor && !isParallelVent) {
        // Motorized 24VDC variable-throat venturi actuator & linkage strictly on single-throat inline venturi
        const act = new THREE.Mesh(new THREE.BoxGeometry(0.14, 0.16, 0.14), mats.conduit);
        act.position.set(svX, svY + 0.24, svZ);
        sysG.add(act);
        const stem = new THREE.Mesh(new THREE.CylinderGeometry(0.014, 0.014, 0.18, 10), mats.bolt);
        stem.position.set(svX, svY + 0.12, svZ);
        sysG.add(stem);
      }
      if (project.deluxe_condensing_reclaimer) {
        // Ceramic-Coated / Sintered-SiC Condensing Plume Abater stacked cleanly ABOVE stack silencer (no collision!)
        const condY = silY + silLen * 0.5 + 0.34;
        const condBox = new THREE.Mesh(
          new THREE.BoxGeometry(sstR * 3.2, 0.48, sstR * 3.0),
          mats.cpvcSolid // ceramic-lined Derakane FRP shell
        );
        condBox.position.set(sstX, condY, sstZ);
        sysG.add(condBox);
        const ceramicTrim = new THREE.Mesh(
          new THREE.BoxGeometry(sstR * 3.32, 0.08, sstR * 3.12),
          mats.terracotta // vitreous ceramic-enamel / SiC core indicator collar
        );
        ceramicTrim.position.set(sstX, condY, sstZ);
        sysG.add(ceramicTrim);
        // Acidic condensate Sch 80 CPVC drain routed cleanly around exterior skid (zero vessel clipping!)
        addHose(
          [
            new THREE.Vector3(sstX, condY - 0.24, sstZ + sstR * 1.55),
            new THREE.Vector3(sstX, beamH + sumpTankH + 0.18, skidMaxZ - 0.14),
            new THREE.Vector3(sumpPosX + sumpTankW * 0.35, beamH + sumpTankH + 0.14, sumpPosZ),
            new THREE.Vector3(sumpPosX + sumpTankW * 0.35, beamH + sumpTankH * 0.85, sumpPosZ),
          ],
          0.016,
          mats.pipeCyan
        );
      }
    }

    assemblyGroup.add(sysG);
  };

  // =========================================================================
  // BRANCH A: DRY-FAN DOWNFLOW EDUCTOR (ZERO WET DOWNSTREAM BLOWER)
  // =========================================================================
  if (isEductor) {
    const rowZ = isCompact ? Math.max(qR + pR + 0.34, 0.88) * 0.5 : 0;
    const qZ = -rowZ;
    const pZ = rowZ;
    const qX = isCompact ? -0.58 : -1.42;
    const teeR = Math.max(vInR, ductR * 1.08);
    const sweepR = clamp(vOutR * 2.1, 0.18, 0.30);
    const sweepCenterY = sumpH + vOutR + 0.12;
    const vTotalH = clamp(vConvL + vThL + vDivL, 0.72, 1.38);
    const vBotY = sweepCenterY + sweepR;
    const teeY = Math.max(qH * 0.72, vBotY + vTotalH + teeR * 1.25);
    const inletY = Math.min(qH - ductR - 0.06, qH * 0.82);
    const qExitY = Math.max(sumpH + ductR + 0.18, qH * 0.38);
    const teeX = qX + qR + 0.60;
    const teeZ = pZ;
    const effBedR = isParallelBed ? Math.max(pR * 1.24, 0.34) : pR;
    const pX = isCompact ? qX - 0.02 : teeX + sweepR + 0.30 + effBedR;

    const { inletLen, sprayTopY } = buildQuenchAt(qX, qZ, inletY);

    if (isCompact) {
      const skid = new THREE.Mesh(
        new THREE.BoxGeometry(Math.abs(teeX - qX) + qR + effBedR + 0.9, 0.05, rowZ * 2 + qR + effBedR + 0.6),
        mats.darkSteel
      );
      skid.position.set((qX + teeX) * 0.5, 0.025, 0);
      assemblyGroup.add(skid);
    }

    // Smooth flanged duct from Stage 1 Quench exit -> flanged Tee side port
    const qToTeePath = new THREE.CurvePath<THREE.Vector3>();
    const qOutPt = new THREE.Vector3(qX + qR, qExitY, qZ);
    const midX = (qX + qR + teeX - teeR) * 0.5;
    const teeSidePt = new THREE.Vector3(teeX - teeR * 1.15, teeY, teeZ);
    qToTeePath.add(
      new THREE.CubicBezierCurve3(
        qOutPt,
        new THREE.Vector3(midX, qExitY, qZ),
        new THREE.Vector3(midX, teeY, teeZ),
        teeSidePt
      )
    );
    assemblyGroup.add(
      new THREE.Mesh(new THREE.TubeGeometry(qToTeePath, 36, vInR, 20, false), mats.ssSolid)
    );
    addBoltedFlange(assemblyGroup, qX + qR + 0.012, qExitY, qZ, "x", vInR, mats.flange, mats.bolt, mats.gasket, 8);
    addBoltedFlange(assemblyGroup, teeSidePt.x, teeY, teeZ, "x", vInR, mats.flange, mats.bolt, mats.gasket, 8);

    // Flanged Vertical Tee body
    const teeH = teeR * 2.4;
    const teeVert = new THREE.Mesh(
      new THREE.CylinderGeometry(teeR, teeR, teeH, 28, 1, cut, thetaStart, thetaLen),
      mats.ss
    );
    teeVert.position.set(teeX, teeY, teeZ);
    assemblyGroup.add(teeVert);
    const teeBranch = new THREE.Mesh(
      new THREE.CylinderGeometry(vInR, vInR, teeR * 1.4, 24, 1, true),
      mats.ssSolid
    );
    teeBranch.rotation.z = Math.PI / 2;
    teeBranch.position.set(teeX - teeR * 0.65, teeY, teeZ);
    assemblyGroup.add(teeBranch);
    addBoltedFlange(assemblyGroup, teeX, teeY + teeH * 0.5, teeZ, "y", teeR, mats.flange, mats.bolt, mats.gasket, 8);
    addBoltedFlange(assemblyGroup, teeX, teeY - teeH * 0.5, teeZ, "y", teeR, mats.flange, mats.bolt, mats.gasket, 8);

    // Vertical Downward-Pointing Static Venturi
    const vTopY = teeY - teeH * 0.5;
    const vSpan = Math.max(0.68, vTopY - vBotY);
    const cL = vSpan * 0.27;
    const tL = vSpan * 0.18;
    const dL = vSpan * 0.55;
    const vConv = new THREE.Mesh(
      new THREE.CylinderGeometry(teeR, vThR, cL, 28, 1, cut, thetaStart, thetaLen),
      mats.ss
    );
    vConv.position.set(teeX, vTopY - cL * 0.5, teeZ);
    assemblyGroup.add(vConv);

    const vThY = vTopY - cL - tL * 0.5;
    const vTh = new THREE.Mesh(
      new THREE.CylinderGeometry(vThR, vThR, tL, 28, 1, cut, thetaStart, thetaLen),
      mats.hot
    );
    vTh.position.set(teeX, vThY, teeZ);
    assemblyGroup.add(vTh);
    const thRing = new THREE.Mesh(
      new THREE.TorusGeometry(vThR + 0.024, 0.011, 10, 24),
      mats.brass
    );
    thRing.rotation.x = Math.PI / 2;
    thRing.position.set(teeX, vThY, teeZ);
    assemblyGroup.add(thRing);

    const vDiv = new THREE.Mesh(
      new THREE.CylinderGeometry(vThR, vOutR, dL, 28, 1, cut, thetaStart, thetaLen),
      mats.ss
    );
    vDiv.position.set(teeX, vTopY - cL - tL - dL * 0.5, teeZ);
    assemblyGroup.add(vDiv);
    addBoltedFlange(assemblyGroup, teeX, vBotY, teeZ, "y", vOutR, mats.cpvcFlange, mats.bolt, mats.gasket, 8);

    // 90° Bottom Cyclonic Sweep Elbow + Condensate Drain Leg + Ball Valve
    const sweepDir = isCompact ? vLeft : vRight;
    const sweepPath = new THREE.CurvePath<THREE.Vector3>();
    const sweepStart = new THREE.Vector3(teeX, vBotY, teeZ);
    const sweepArc = makeArc90Bezier(sweepStart, vDn, sweepDir, sweepR);
    sweepPath.add(sweepArc);
    const sweepArcEnd = sweepArc.getPoint(1);
    const bedEntryX = isCompact ? pX + effBedR : pX - effBedR;
    const sweepEndPt = new THREE.Vector3(bedEntryX, sweepCenterY, pZ);
    sweepPath.add(new THREE.LineCurve3(sweepArcEnd, sweepEndPt));
    assemblyGroup.add(
      new THREE.Mesh(new THREE.TubeGeometry(sweepPath, 40, vOutR, 24, false), mats.cpvcSolid)
    );
    addBoltedFlange(
      assemblyGroup,
      bedEntryX + (isCompact ? 0.012 : -0.012),
      sweepCenterY,
      pZ,
      "x",
      vOutR,
      mats.cpvcFlange,
      mats.bolt,
      mats.gasket,
      8
    );

    const drainX = teeX + (isCompact ? -sweepR * 0.45 : sweepR * 0.45);
    const drainStub = new THREE.Mesh(
      new THREE.CylinderGeometry(0.030, 0.030, sweepCenterY - 0.04, 16),
      mats.cpvcSolid
    );
    drainStub.position.set(drainX, (sweepCenterY - 0.04) * 0.5, teeZ);
    assemblyGroup.add(drainStub);
    addBallValve(assemblyGroup, drainX, sweepCenterY * 0.45, teeZ, "y", 0.030, mats.cpvcSolid, mats.safetyRed);

    // Elevated Dry Ambient-Air Radial Blower blowing STRAIGHT DOWN into top of Tee
    const dryWheelR = clamp(Math.max((b.wheel_diameter_m || 0.48) * 0.52, ductR * 2.1), 0.20, 0.34);
    const dryScrollW = Math.max(dryWheelR * 0.68, teeR * 1.65);
    const dryWallT = 0.016;
    const dryVol = computeVoluteProfiles(dryWheelR, teeR * 0.85, dryWallT);
    const dryStackLocalX = (dryVol.xIn + dryVol.rMax) * 0.5;
    const dryNozzleH = 0.24;
    const dryFanY = teeY + teeH * 0.5 + dryNozzleH + dryVol.yDis + 0.04;
    const dryFanX = teeX + dryStackLocalX;

    const dryFanGroup = new THREE.Group();
    dryFanGroup.position.set(dryFanX, dryFanY, teeZ);

    const scrollRotGroup = new THREE.Group();
    scrollRotGroup.rotation.z = Math.PI;
    const dWrapShape = new THREE.Shape();
    dWrapShape.moveTo(dryVol.outerPts[0].x, dryVol.outerPts[0].y);
    for (let i = 1; i < dryVol.outerPts.length; i++) dWrapShape.lineTo(dryVol.outerPts[i].x, dryVol.outerPts[i].y);
    for (let i = dryVol.innerPts.length - 1; i >= 0; i--) dWrapShape.lineTo(dryVol.innerPts[i].x, dryVol.innerPts[i].y);
    dWrapShape.closePath();
    const dWrapGeo = new THREE.ExtrudeGeometry(dWrapShape, { depth: dryScrollW, bevelEnabled: false });
    dWrapGeo.translate(0, 0, -dryScrollW * 0.5);
    scrollRotGroup.add(new THREE.Mesh(dWrapGeo, mats.blower));

    const dPlateShape = new THREE.Shape();
    dPlateShape.moveTo(dryVol.outerPts[0].x, dryVol.outerPts[0].y);
    for (let i = 1; i < dryVol.outerPts.length; i++) dPlateShape.lineTo(dryVol.outerPts[i].x, dryVol.outerPts[i].y);
    dPlateShape.closePath();
    const dBackGeo = new THREE.ExtrudeGeometry(dPlateShape, { depth: dryWallT, bevelEnabled: false });
    dBackGeo.translate(0, 0, -dryScrollW * 0.5 - dryWallT);
    scrollRotGroup.add(new THREE.Mesh(dBackGeo, mats.blower));

    const dFrontShape = new THREE.Shape();
    dFrontShape.moveTo(dryVol.outerPts[0].x, dryVol.outerPts[0].y);
    for (let i = 1; i < dryVol.outerPts.length; i++) dFrontShape.lineTo(dryVol.outerPts[i].x, dryVol.outerPts[i].y);
    dFrontShape.closePath();
    const dHole = new THREE.Path();
    dHole.absarc(0, 0, cut ? dryWheelR * 1.0 : teeR * 0.9, 0, Math.PI * 2, true);
    dFrontShape.holes.push(dHole);
    const dFrontGeo = new THREE.ExtrudeGeometry(dFrontShape, { depth: dryWallT, bevelEnabled: false, curveSegments: 32 });
    dFrontGeo.translate(0, 0, dryScrollW * 0.5);
    scrollRotGroup.add(new THREE.Mesh(dFrontGeo, mats.blower));

    const dSqToRnd = new THREE.Mesh(
      buildRectToRoundGeometry(dryVol.disW, dryScrollW, teeR, dryNozzleH, 32),
      mats.blower
    );
    dSqToRnd.position.set(dryStackLocalX, dryVol.yDis + 0.01, 0);
    scrollRotGroup.add(dSqToRnd);
    dryFanGroup.add(scrollRotGroup);

    const intakeRing = new THREE.Mesh(
      new THREE.TorusGeometry(dryWheelR * 0.55, 0.018, 10, 28),
      mats.diluter
    );
    intakeRing.position.z = dryScrollW * 0.5 + dryWallT + 0.02;
    dryFanGroup.add(intakeRing);

    const dryImp = buildImpellerGroup(dryWheelR, dryScrollW, mats.blowerDark, mats.blade, mats.blowerDark, mats.shroud);
    dryFanGroup.add(dryImp);
    impellerMesh = dryImp;

    const dmR = dryWheelR * 0.44;
    const dmLen = dryWheelR * 1.05;
    const dmZ = -dryScrollW * 0.5 - dryWallT - dmLen * 0.65;
    const dmBody = new THREE.Mesh(new THREE.CylinderGeometry(dmR, dmR, dmLen, 20), mats.motor);
    dmBody.rotation.x = Math.PI / 2;
    dmBody.position.z = dmZ;
    dryFanGroup.add(dmBody);
    const supportPost = new THREE.Mesh(
      new THREE.BoxGeometry(0.08, dryFanY - dryWheelR * 0.5, 0.08),
      mats.darkSteel
    );
    supportPost.position.set(0, -(dryFanY + dryWheelR * 0.5) * 0.5, dmZ);
    dryFanGroup.add(supportPost);
    assemblyGroup.add(dryFanGroup);

    // Stage 3 Packed Bed + Top Exhaust Stack
    const { bundleR: edBundleR, topNeckY, distRingY, sprayBedH } = buildPackedBedAt(pX, pZ, sweepCenterY);
    const stackR = ductR * 0.96;
    const stackH = clamp(1.05, 0.85, 1.32);
    const { exitY: stackExitWorldY, diluterEntry, diluterTeeY } = addStackAndDiluter(
      assemblyGroup,
      pX,
      topNeckY,
      pZ,
      stackR,
      stackH
    );

    addSystemsOverlay3D({
      qX, qZ, qR, qH,
      pX, pZ, pR: edBundleR, pH,
      vX: teeX, vY: vThY, vZ: teeZ,
      bX: dryFanX, bY: dryFanY, bZ: teeZ, bR: dryWheelR, bDepth: dryScrollW + dmLen,
      stX: pX, stBaseY: topNeckY, stH: stackH, stR: stackR, stZ: pZ,
      inletY, ductR,
    });

    if (showDimensions) {
      addDimensionCallout3D(
        assemblyGroup,
        new THREE.Vector3(qX - qR - 0.14, 0, qZ),
        new THREE.Vector3(qX - qR - 0.14, qH, qZ),
        new THREE.Vector3(-0.28, 0, 0),
        `Quench ${fmtDimMeters(qR * 2, units)} × ${fmtDimMeters(qH, units)}`,
        "#fb923c",
        disposableTextures
      );
      addDimensionCallout3D(
        assemblyGroup,
        new THREE.Vector3(pX + edBundleR + 0.14, 0, pZ),
        new THREE.Vector3(pX + edBundleR + 0.14, pH, pZ),
        new THREE.Vector3(0.32, 0, 0),
        `Bed ${fmtDimMeters(edBundleR * 2, units)} × ${fmtDimMeters(pH, units)}`,
        "#34d399",
        disposableTextures
      );
    }

    const rawBox = new THREE.Box3().setFromObject(assemblyGroup);
    const centerX = -(rawBox.min.x + rawBox.max.x) * 0.5;
    const centerZ = -(rawBox.min.z + rawBox.max.z) * 0.5;
    assemblyGroup.position.set(centerX, 0, centerZ);

    const totalW = rawBox.max.x - rawBox.min.x;
    const totalD = rawBox.max.z - rawBox.min.z;
    const totalH = Math.max(rawBox.max.y, stackExitWorldY) + 0.18;
    const tanHalfFov = Math.tan(THREE.MathUtils.degToRad(cameraFovDeg * 0.5));
    const allTargetY = totalH * 0.50;
    const allTargetX = Math.max(0.18, totalW * 0.06);
    const allDist = Math.max(
      5.8,
      Math.max(
        (totalH * 0.56) / tanHalfFov,
        (totalW * 0.48) / (tanHalfFov * Math.max(1.4, cameraAspect))
      ) *
        1.22 +
        totalD * 0.38
    );

    const stageViews: Record<StageKey, StageView> = {
      overview: {
        target: new THREE.Vector3(allTargetX, allTargetY, 0),
        cam: new THREE.Vector3(allTargetX + 0.85, allTargetY + totalH * 0.22, allDist),
      },
      inlet: {
        target: new THREE.Vector3(qX - qR - inletLen * 0.5 + centerX, inletY + 0.12, qZ + centerZ),
        cam: new THREE.Vector3(qX - qR + 0.15 + centerX, inletY + 0.45, qZ + centerZ + 2.3),
      },
      quench: {
        target: new THREE.Vector3(qX + centerX + 0.10, qH * 0.55, qZ + centerZ),
        cam: new THREE.Vector3(qX + centerX + 0.45, qH * 0.78, qZ + centerZ + Math.max(2.4, qH * 1.5)),
      },
      venturi: {
        target: new THREE.Vector3(teeX + centerX + 0.08, (teeY + sweepCenterY) * 0.5, teeZ + centerZ),
        cam: new THREE.Vector3(teeX + centerX + 0.45, (teeY + sweepCenterY) * 0.5 + 0.35, teeZ + centerZ + 2.65),
      },
      packed: {
        target: new THREE.Vector3(pX + centerX + 0.18, stackExitWorldY * 0.50, pZ + centerZ),
        cam: new THREE.Vector3(pX + centerX + 0.65, stackExitWorldY * 0.58, pZ + centerZ + Math.max(3.4, stackExitWorldY * 1.12)),
      },
      blower: {
        target: new THREE.Vector3(dryFanX + centerX, dryFanY - 0.15, teeZ + centerZ),
        cam: new THREE.Vector3(dryFanX + centerX + 0.55, dryFanY + 0.25, teeZ + centerZ + 2.55),
      },
    };

    const ox = centerX;
    const oz = centerZ;
    const primaryFlow: FlowSample[] = [
      ...sampleCurve(
        new THREE.LineCurve3(
          new THREE.Vector3(qX - qR - inletLen, inletY, qZ),
          new THREE.Vector3(qX, inletY, qZ)
        ),
        18,
        ductR * 0.42,
        1.25,
        "inlet",
        ox,
        oz
      ),
      ...sampleCurve(
        new THREE.LineCurve3(
          new THREE.Vector3(qX, inletY, qZ),
          new THREE.Vector3(qX, qExitY, qZ)
        ),
        24,
        qR * 0.55,
        0.72,
        "quench",
        ox,
        oz
      ),
      ...sampleCurve(qToTeePath, 26, vInR * 0.45, 1.2, "quench", ox, oz),
      ...sampleCurve(
        new THREE.LineCurve3(
          new THREE.Vector3(teeX, teeY, teeZ),
          new THREE.Vector3(teeX, vBotY, teeZ)
        ),
        22,
        vThR * 0.55,
        2.85,
        "venturi",
        ox,
        oz
      ),
      ...sampleCurve(sweepPath, 28, vOutR * 0.45, 1.35, "venturi", ox, oz),
      ...sampleCurve(
        new THREE.LineCurve3(
          new THREE.Vector3(bedEntryX, sweepCenterY, pZ),
          new THREE.Vector3(pX, sweepCenterY, pZ)
        ),
        8,
        vOutR * 0.45,
        1.0,
        "packed",
        ox,
        oz
      ),
      ...sampleCurve(
        new THREE.LineCurve3(
          new THREE.Vector3(pX, sweepCenterY, pZ),
          new THREE.Vector3(pX, topNeckY, pZ)
        ),
        46,
        edBundleR * (isParallelBed ? 0.62 : 0.52),
        0.58,
        "packed",
        ox,
        oz
      ),
      ...sampleCurve(
        new THREE.LineCurve3(
          new THREE.Vector3(pX, topNeckY, pZ),
          new THREE.Vector3(pX, stackExitWorldY, pZ)
        ),
        20,
        stackR * 0.48,
        1.45,
        "blower",
        ox,
        oz
      ),
      ...sampleCurve(
        new THREE.LineCurve3(
          new THREE.Vector3(pX, stackExitWorldY, pZ),
          new THREE.Vector3(pX, stackExitWorldY + 0.65, pZ)
        ),
        16,
        stackR * 1.15,
        1.15,
        "plume",
        ox,
        oz
      ),
    ];

    const motiveFlow: FlowSample[] = [
      ...sampleCurve(
        new THREE.LineCurve3(
          new THREE.Vector3(dryFanX, dryFanY, teeZ + 0.25),
          new THREE.Vector3(dryFanX, dryFanY, teeZ)
        ),
        10,
        0.025,
        1.4,
        "motive",
        ox,
        oz
      ),
      ...sampleCurve(
        new THREE.LineCurve3(
          new THREE.Vector3(dryFanX, dryFanY, teeZ),
          new THREE.Vector3(teeX, dryFanY - dryVol.yDis * 0.6, teeZ)
        ),
        12,
        0.028,
        1.8,
        "motive",
        ox,
        oz
      ),
      ...sampleCurve(
        new THREE.LineCurve3(
          new THREE.Vector3(teeX, dryFanY - dryVol.yDis * 0.6, teeZ),
          new THREE.Vector3(teeX, vBotY, teeZ)
        ),
        22,
        vThR * 0.50,
        3.0,
        "motive",
        ox,
        oz
      ),
      ...sampleCurve(sweepPath, 24, vOutR * 0.40, 1.35, "motive", ox, oz),
      ...sampleCurve(
        new THREE.LineCurve3(
          new THREE.Vector3(pX, sweepCenterY, pZ),
          new THREE.Vector3(pX, stackExitWorldY + 0.40, pZ)
        ),
        34,
        edBundleR * 0.45,
        0.95,
        "motive",
        ox,
        oz
      ),
    ];

    const diluterFlow: FlowSample[] = [];
    if (hasDiluter && diluterEntry) {
      diluterFlow.push(
        ...sampleCurve(
          new THREE.LineCurve3(diluterEntry, new THREE.Vector3(pX, diluterTeeY, pZ)),
          18,
          0.025,
          1.8,
          "diluter",
          ox,
          oz
        ),
        ...sampleCurve(
          new THREE.LineCurve3(
            new THREE.Vector3(pX, diluterTeeY, pZ),
            new THREE.Vector3(pX, stackExitWorldY + 0.65, pZ)
          ),
          22,
          stackR * 0.55,
          1.75,
          "diluter",
          ox,
          oz
        )
      );
    }

    sprayZones.push(
      {
        topCenter: new THREE.Vector3(qX + ox, sprayTopY, qZ + oz),
        topRadius: qR * 0.28,
        botRadius: qR * 0.72,
        height: Math.max(0.5, qH - 0.55),
        color: 0x38bdf8,
        count: 34,
        speed: 1.25,
      },
      {
        topCenter: new THREE.Vector3(pX + ox, distRingY, pZ + oz),
        topRadius: edBundleR * 0.58,
        botRadius: edBundleR * 0.76,
        height: sprayBedH,
        color: 0x22d3ee,
        count: 40,
        speed: 0.85,
      }
    );

    return {
      stageViews,
      primaryFlow,
      motiveFlow,
      diluterFlow,
      sprayZones,
      impellerMesh,
      impellerDir: 1,
      diluterImpellerMesh,
      architectureSignature: `eductor|${isParallelBed ? "par" : "col"}|${isCompact ? "u" : "lin"}|${showSystems ? "sys" : "std"}`,
      scaleSummary: `Scale 1:1 m · ${fmtDimMeters(env.length_m, units)} L × ${fmtDimMeters(env.width_m, units)} W × ${fmtDimMeters(env.height_m, units)} H`,
      disposableTextures,
    };
  }

  // =========================================================================
  // BRANCH B: INLINE ID BLOWER (LINEAR OR COMPACT U-TURN SKID)
  // =========================================================================
  const effPR = isParallelBed ? Math.max(pR * 1.24, 0.34) : pR;
  const vYMin = sumpH + vOutR + 0.08;
  const vY = clamp(Math.max(vYMin, 0.48), vYMin, Math.max(vYMin, qH - ductR * 2.2 - vInR - 0.10));
  const inletY = Math.min(
    qH - ductR - 0.06,
    Math.max(vY + vInR + ductR + 0.16, qH - ductR * 1.4 - 0.10)
  );

  const wheelR = clamp(Math.max((b.wheel_diameter_m || 0.60) * 0.50, ductR * 2.4), 0.22, 0.42);
  const scrollW = Math.max(wheelR * 0.72, ductR * 1.95);
  const wallT = 0.018;
  const vol = computeVoluteProfiles(wheelR, ductR, wallT);

  const rowZ = isCompact ? Math.max(qR + effPR + 0.34, 0.90) * 0.5 : 0;
  const qZ = -rowZ;
  const pZ = rowZ;

  const qX = isCompact ? -0.95 : -1.72;
  const spool1Len = 0.24;
  const vStartX = qX + qR + spool1Len;
  const vEndX = vStartX + vConvL + vThL + vDivL;
  const spool2Len = 0.24;
  const pX = isCompact ? (vStartX + vEndX) * 0.5 - 0.22 : vEndX + spool2Len + effPR;

  // Stage 1: Quench Tower + Bypass Tee
  const { inletLen, sprayTopY } = buildQuenchAt(qX, qZ, inletY);

  if (isCompact) {
    const skid = new THREE.Mesh(
      new THREE.BoxGeometry(Math.max(2.6, vEndX - qX + 1.15), 0.05, rowZ * 2 + qR + effPR + 0.65),
      mats.darkSteel
    );
    skid.position.set((qX + vEndX) * 0.5, 0.025, 0);
    assemblyGroup.add(skid);
  }

  // Stage 1 -> Stage 2 Flanged Spool
  const conn1 = new THREE.Mesh(
    new THREE.CylinderGeometry(vInR, vInR, spool1Len + 0.024, 24, 1, true),
    mats.ssSolid
  );
  conn1.rotation.z = Math.PI / 2;
  conn1.position.set(qX + qR + spool1Len * 0.5 - 0.008, vY, qZ);
  assemblyGroup.add(conn1);
  addBoltedFlange(assemblyGroup, qX + qR + 0.012, vY, qZ, "x", vInR, mats.flange, mats.bolt, mats.gasket, 8);
  addBoltedFlange(assemblyGroup, vStartX, vY, qZ, "x", vInR, mats.flange, mats.bolt, mats.gasket, 8);

  // Stage 2: Single Wet Venturi OR Parallel Sanitary Mini-Venturis
  const vThCenterX = vStartX + vConvL + vThL * 0.5;
  if (isParallelVent) {
    const nThroats = clamp(arch.parallel_venturi.throat_count || 3, 2, 4);
    const mInR = vInR * 0.42;
    const mThR = Math.max(0.022, vThR * 0.58);
    const mOutR = vOutR * 0.44;
    const head1 = new THREE.Mesh(
      new THREE.BoxGeometry(0.10, vInR * 2.0, vInR * 2.5),
      mats.ssSolid
    );
    head1.position.set(vStartX + 0.045, vY, qZ);
    assemblyGroup.add(head1);
    const head2 = new THREE.Mesh(
      new THREE.BoxGeometry(0.10, vOutR * 2.0, vOutR * 2.5),
      mats.ssSolid
    );
    head2.position.set(vEndX - 0.045, vY, qZ);
    assemblyGroup.add(head2);

    for (let ti = 0; ti < nThroats; ti++) {
      const dz =
        nThroats === 1
          ? 0
          : ((ti / (nThroats - 1)) - 0.5) * vInR * 1.65;
      const mc = new THREE.Mesh(new THREE.CylinderGeometry(mThR, mInR, vConvL, 20), mats.ssSolid);
      mc.rotation.z = -Math.PI / 2;
      mc.position.set(vStartX + vConvL * 0.5, vY, qZ + dz);
      assemblyGroup.add(mc);
      addTriClamp(assemblyGroup, vStartX + vConvL, vY, qZ + dz, "x", mThR, mats.bolt, mats.gasket);

      const mt = new THREE.Mesh(new THREE.CylinderGeometry(mThR, mThR, vThL, 20), mats.hotSolid);
      mt.rotation.z = -Math.PI / 2;
      mt.position.set(vThCenterX, vY, qZ + dz);
      assemblyGroup.add(mt);
      addTriClamp(assemblyGroup, vStartX + vConvL + vThL, vY, qZ + dz, "x", mThR, mats.bolt, mats.gasket);

      const md = new THREE.Mesh(new THREE.CylinderGeometry(mOutR, mThR, vDivL, 20), mats.ssSolid);
      md.rotation.z = -Math.PI / 2;
      md.position.set(vStartX + vConvL + vThL + vDivL * 0.5, vY, qZ + dz);
      assemblyGroup.add(md);
    }
  } else {
    const vConv = new THREE.Mesh(
      new THREE.CylinderGeometry(vThR, vInR, vConvL, 28, 1, cut, thetaStart, thetaLen),
      mats.ss
    );
    vConv.rotation.z = -Math.PI / 2;
    vConv.position.set(vStartX + vConvL * 0.5, vY, qZ);
    assemblyGroup.add(vConv);

    const vTh = new THREE.Mesh(
      new THREE.CylinderGeometry(vThR, vThR, vThL, 28, 1, cut, thetaStart, thetaLen),
      mats.hot
    );
    vTh.rotation.z = -Math.PI / 2;
    vTh.position.set(vThCenterX, vY, qZ);
    assemblyGroup.add(vTh);

    // Venturi throat spray injection collar + radial nozzles
    const throatRing = new THREE.Mesh(
      new THREE.TorusGeometry(vThR + 0.025, 0.011, 10, 24),
      mats.brass
    );
    throatRing.rotation.y = Math.PI / 2;
    throatRing.position.set(vThCenterX, vY, qZ);
    assemblyGroup.add(throatRing);

    for (const ang of [0, Math.PI / 2, Math.PI, (3 * Math.PI) / 2]) {
      const nStub = new THREE.Mesh(
        new THREE.CylinderGeometry(0.010, 0.010, 0.045, 8),
        mats.brass
      );
      nStub.position.set(
        vThCenterX,
        vY + Math.cos(ang) * (vThR + 0.02),
        qZ + Math.sin(ang) * (vThR + 0.02)
      );
      nStub.rotation.x = ang;
      assemblyGroup.add(nStub);
    }

    // Optional compressed air twin-fluid atomizing manifold + FRL regulator
    if (blowerDrive === "compressor_air_atomize") {
      const frl = new THREE.Mesh(
        new THREE.CylinderGeometry(0.032, 0.032, 0.14, 14),
        mats.conduit
      );
      frl.position.set(vThCenterX, vY + vInR + 0.16, qZ + 0.14);
      assemblyGroup.add(frl);
      const airLine = new THREE.Mesh(
        new THREE.CylinderGeometry(0.010, 0.010, 0.18, 10),
        mats.conduit
      );
      airLine.position.set(vThCenterX, vY + vInR + 0.04, qZ + 0.06);
      airLine.rotation.x = -0.35;
      assemblyGroup.add(airLine);
    }

    const vDiv = new THREE.Mesh(
      new THREE.CylinderGeometry(vOutR, vThR, vDivL, 28, 1, cut, thetaStart, thetaLen),
      mats.ss
    );
    vDiv.rotation.z = -Math.PI / 2;
    vDiv.position.set(vStartX + vConvL + vThL + vDivL * 0.5, vY, qZ);
    assemblyGroup.add(vDiv);
  }

  // Stage 2 -> Stage 3 Connection (Linear Flanged Spool OR Smooth 180° U-Turn)
  let uTurnPts: THREE.Vector3[] = [];
  if (isCompact) {
    const uPath = new THREE.CurvePath<THREE.Vector3>();
    const uR = Math.min(rowZ, 0.42);
    const u1 = makeArc90Bezier(new THREE.Vector3(vEndX, vY, qZ), vRight, vFront, uR);
    uPath.add(u1);
    const u1End = u1.getPoint(1);
    const u2Start = new THREE.Vector3(u1End.x, vY, pZ - uR);
    if (u2Start.z > u1End.z + 0.005) uPath.add(new THREE.LineCurve3(u1End, u2Start));
    const u2 = makeArc90Bezier(u2Start, vFront, vLeft, uR);
    uPath.add(u2);
    const u2End = u2.getPoint(1);
    const pEntryPt = new THREE.Vector3(pX + effPR, vY, pZ);
    if (u2End.x > pEntryPt.x + 0.005) uPath.add(new THREE.LineCurve3(u2End, pEntryPt));
    assemblyGroup.add(
      new THREE.Mesh(new THREE.TubeGeometry(uPath, 44, vOutR, 24, false), mats.cpvcSolid)
    );
    addBoltedFlange(assemblyGroup, vEndX, vY, qZ, "x", vOutR, mats.flange, mats.bolt, mats.gasket, 8);
    addBoltedFlange(assemblyGroup, pX + effPR + 0.012, vY, pZ, "x", vOutR, mats.cpvcFlange, mats.bolt, mats.gasket, 8);
    uTurnPts = uPath.getSpacedPoints(32);
  } else {
    const conn2 = new THREE.Mesh(
      new THREE.CylinderGeometry(vOutR, vOutR, spool2Len + 0.024, 24, 1, true),
      mats.cpvcSolid
    );
    conn2.rotation.z = Math.PI / 2;
    conn2.position.set(vEndX + spool2Len * 0.5 + 0.008, vY, 0);
    assemblyGroup.add(conn2);
    addBoltedFlange(assemblyGroup, vEndX, vY, 0, "x", vOutR, mats.flange, mats.bolt, mats.gasket, 8);
    addBoltedFlange(assemblyGroup, pX - effPR - 0.012, vY, 0, "x", vOutR, mats.cpvcFlange, mats.bolt, mats.gasket, 8);
  }

  // Stage 3: Packed Bed Absorber
  const { bundleR, topNeckY, distRingY, sprayBedH } = buildPackedBedAt(pX, pZ, vY);

  // Stage 4: Blower / Motive Drive Geometry (adapts to Phase 8 blower_drive_option!)
  const gapS3S4 = Math.max(ductR * 3.5, 0.42);
  const bDir = isCompact ? -1 : 1;
  const dropX = pX + bDir * (bundleR + gapS3S4 * 0.5);
  const bX = pX + bDir * (bundleR + gapS3S4 + vol.rLeft);
  const bY = vol.rBottom + 0.18;
  const eyeZ = scrollW * 0.5 + wallT + 0.028;

  const blowerGroup = new THREE.Group();
  blowerGroup.position.set(bX, bY, pZ);

  const isRegenRing = blowerDrive === "regen_side_channel";
  const isDualBypass = blowerDrive === "dual_bypass_blowers";
  const isCompEductor = blowerDrive === "compressor_eductor_only";

  const motorLen = wheelR * 1.12;
  const motorR = wheelR * 0.44;
  const shaftGap = wheelR * 0.42;
  const motorCenterZ = -scrollW * 0.5 - wallT - shaftGap - motorLen * 0.5;
  const skidDepth = isCompEductor ? scrollW + 0.24 : scrollW + shaftGap + motorLen + 0.18;
  const stackLocalX = isCompEductor ? 0 : (vol.xIn + vol.rMax) * 0.5;
  const stackX = bX + stackLocalX;
  const stackR = ductR * 0.96;
  const transH = 0.22;
  const stackBaseLocalY = isCompEductor ? 0.32 : vol.yDis + 0.012 + transH;

  if (isCompEductor) {
    // Compact Inline 316L Stainless Compressed-Air Amplifier / Eductor + FRL Regulator (NO scroll or electric motor!)
    const ampCollar = new THREE.Mesh(
      new THREE.CylinderGeometry(stackR * 1.48, stackR * 1.48, 0.22, 24),
      mats.steamerSS
    );
    ampCollar.position.set(0, 0.18, 0);
    blowerGroup.add(ampCollar);
    addBoltedFlange(blowerGroup, 0, 0.06, 0, "y", stackR, mats.flange, mats.bolt, mats.gasket, 8);
    addBoltedFlange(blowerGroup, 0, stackBaseLocalY, 0, "y", stackR, mats.flange, mats.bolt, mats.gasket, 8);
    const frlBox = new THREE.Mesh(new THREE.CylinderGeometry(0.036, 0.036, 0.16, 14), mats.conduit);
    frlBox.position.set(stackR * 1.9, 0.18, 0.12);
    blowerGroup.add(frlBox);
    const airPipe = new THREE.Mesh(new THREE.CylinderGeometry(0.014, 0.014, stackR * 0.9, 10), mats.brass);
    airPipe.rotation.z = Math.PI / 2;
    airPipe.position.set(stackR * 1.45, 0.18, 0.06);
    blowerGroup.add(airPipe);
    const ped = new THREE.Mesh(new THREE.BoxGeometry(0.32, Math.max(0.08, bY - 0.04), 0.32), mats.darkSteel);
    ped.position.set(0, -bY * 0.5, 0);
    blowerGroup.add(ped);
  } else if (isDualBypass) {
    // 2× Parallel Vertical 3-Stage Tangential Bypass Blowers (Ametek Lamb) on Y-Manifold
    const mani = new THREE.Mesh(new THREE.BoxGeometry(wheelR * 2.1, 0.18, scrollW * 1.15), mats.cpvcSolid);
    mani.position.set(stackLocalX * 0.5, 0.02, 0);
    blowerGroup.add(mani);
    for (const dx of [-wheelR * 0.52, wheelR * 0.52]) {
      const fanCan = new THREE.Mesh(new THREE.CylinderGeometry(0.11, 0.11, 0.20, 20), mats.steamerSS);
      fanCan.position.set(stackLocalX * 0.5 + dx, 0.22, 0);
      blowerGroup.add(fanCan);
      const motorCan = new THREE.Mesh(new THREE.CylinderGeometry(0.082, 0.082, 0.18, 18), mats.motor);
      motorCan.position.set(stackLocalX * 0.5 + dx, 0.41, 0);
      blowerGroup.add(motorCan);
    }
    const yMerge = new THREE.Mesh(new THREE.CylinderGeometry(stackR, stackR * 1.35, 0.24, 24), mats.blower);
    yMerge.position.set(stackLocalX, stackBaseLocalY - 0.12, 0);
    blowerGroup.add(yMerge);
    addBoltedFlange(blowerGroup, stackLocalX, stackBaseLocalY, 0, "y", stackR, mats.blowerFlange, mats.bolt, mats.gasket, 8);
    const ped = new THREE.Mesh(new THREE.BoxGeometry(wheelR * 2.2, Math.max(0.08, bY - 0.08), scrollW * 1.2), mats.darkSteel);
    ped.position.set(stackLocalX * 0.5, -bY * 0.5 - 0.04, 0);
    blowerGroup.add(ped);
  } else if (isRegenRing) {
    // Die-Cast Aluminum Side-Channel Regenerative Ring Blower (Toroidal Finned Ring + Rear TEFC Motor)
    const torusShell = new THREE.Mesh(new THREE.TorusGeometry(wheelR * 0.72, wheelR * 0.28, 18, 32), mats.steamerSS);
    blowerGroup.add(torusShell);
    const hubDisk = new THREE.Mesh(new THREE.CylinderGeometry(wheelR * 0.52, wheelR * 0.52, wheelR * 0.36, 24), mats.blowerDark);
    hubDisk.rotation.x = Math.PI / 2;
    blowerGroup.add(hubDisk);
    const impGroup = buildImpellerGroup(wheelR * 0.65, wheelR * 0.32, mats.blowerDark, mats.blade, mats.blowerDark, mats.shroud);
    blowerGroup.add(impGroup);
    impellerMesh = impGroup;
    const motorBody = new THREE.Mesh(new THREE.CylinderGeometry(motorR, motorR, motorLen, 24), mats.motor);
    motorBody.rotation.x = Math.PI / 2;
    motorBody.position.z = motorCenterZ * 0.78;
    blowerGroup.add(motorBody);
    const silOut = new THREE.Mesh(new THREE.CylinderGeometry(stackR, stackR * 1.12, stackBaseLocalY, 20), mats.steamerSS);
    silOut.position.set(stackLocalX, stackBaseLocalY * 0.5, 0);
    blowerGroup.add(silOut);
    addBoltedFlange(blowerGroup, stackLocalX, stackBaseLocalY, 0, "y", stackR, mats.blowerFlange, mats.bolt, mats.gasket, 8);
    const ped = new THREE.Mesh(new THREE.BoxGeometry(wheelR * 1.9, Math.max(0.08, bY - wheelR * 0.9), skidDepth * 0.8), mats.darkSteel);
    ped.position.set(0, -bY + Math.max(0.08, bY - wheelR * 0.9) * 0.5, -wheelR * 0.35);
    blowerGroup.add(ped);
  } else {
    // Standard Industrial Radial Scroll (Volute) Blower + 9-Blade Impeller + Rear TEFC Motor
    const wrapperShape = new THREE.Shape();
    wrapperShape.moveTo(vol.outerPts[0].x, vol.outerPts[0].y);
    for (let i = 1; i < vol.outerPts.length; i++) wrapperShape.lineTo(vol.outerPts[i].x, vol.outerPts[i].y);
    for (let i = vol.innerPts.length - 1; i >= 0; i--) wrapperShape.lineTo(vol.innerPts[i].x, vol.innerPts[i].y);
    wrapperShape.closePath();
    const wrapperGeo = new THREE.ExtrudeGeometry(wrapperShape, { depth: scrollW, bevelEnabled: false });
    wrapperGeo.translate(0, 0, -scrollW * 0.5);
    blowerGroup.add(new THREE.Mesh(wrapperGeo, mats.blower));

    const plateShape = new THREE.Shape();
    plateShape.moveTo(vol.outerPts[0].x, vol.outerPts[0].y);
    for (let i = 1; i < vol.outerPts.length; i++) plateShape.lineTo(vol.outerPts[i].x, vol.outerPts[i].y);
    plateShape.closePath();
    const backPlateGeo = new THREE.ExtrudeGeometry(plateShape, { depth: wallT, bevelEnabled: false });
    backPlateGeo.translate(0, 0, -scrollW * 0.5 - wallT);
    blowerGroup.add(new THREE.Mesh(backPlateGeo, mats.blower));

    const frontShape = new THREE.Shape();
    frontShape.moveTo(vol.outerPts[0].x, vol.outerPts[0].y);
    for (let i = 1; i < vol.outerPts.length; i++) frontShape.lineTo(vol.outerPts[i].x, vol.outerPts[i].y);
    frontShape.closePath();
    const frontHoleR = cut ? wheelR * 1.02 : ductR * 1.02;
    const frontHole = new THREE.Path();
    frontHole.absarc(0, 0, frontHoleR, 0, Math.PI * 2, true);
    frontShape.holes.push(frontHole);
    const frontPlateGeo = new THREE.ExtrudeGeometry(frontShape, { depth: wallT, bevelEnabled: false, curveSegments: 36 });
    frontPlateGeo.translate(0, 0, scrollW * 0.5);
    blowerGroup.add(new THREE.Mesh(frontPlateGeo, mats.blower));

    if (cut) {
      const bezel = new THREE.Mesh(
        new THREE.TorusGeometry(frontHoleR, 0.010, 10, 40),
        mats.blowerFlange
      );
      bezel.position.z = scrollW * 0.5 + wallT;
      blowerGroup.add(bezel);
    }

    addBoltedFlange(blowerGroup, 0, 0, eyeZ, "z", ductR, mats.cpvcFlange, mats.bolt, mats.gasket, 8);
    const bellLen = wallT + 0.032;
    const suctionBell = new THREE.Mesh(
      new THREE.CylinderGeometry(ductR, ductR * 1.16, bellLen, 28, 1, true),
      cut ? mats.cutGlass : mats.cpvcSolid
    );
    suctionBell.rotation.x = Math.PI / 2;
    suctionBell.position.z = scrollW * 0.5 + bellLen * 0.5 - 0.005;
    blowerGroup.add(suctionBell);

    const impGroup = buildImpellerGroup(wheelR, scrollW, mats.blowerDark, mats.blade, mats.blowerDark, mats.shroud);
    blowerGroup.add(impGroup);
    impellerMesh = impGroup;

    const shaftLen = scrollW * 0.5 + wallT + shaftGap + 0.08;
    const shaft = new THREE.Mesh(
      new THREE.CylinderGeometry(wheelR * 0.085, wheelR * 0.085, shaftLen, 16),
      mats.blade
    );
    shaft.rotation.x = Math.PI / 2;
    shaft.position.z = -shaftLen * 0.5;
    blowerGroup.add(shaft);

    const sealCollar = new THREE.Mesh(
      new THREE.CylinderGeometry(wheelR * 0.18, wheelR * 0.18, 0.04, 20),
      mats.darkSteel
    );
    sealCollar.rotation.x = Math.PI / 2;
    sealCollar.position.z = -scrollW * 0.5 - wallT - 0.02;
    blowerGroup.add(sealCollar);

    const bearingZ = -scrollW * 0.5 - wallT - shaftGap * 0.52;
    const pillowBlock = new THREE.Mesh(
      new THREE.BoxGeometry(wheelR * 0.44, wheelR * 0.42, wheelR * 0.24),
      mats.darkSteel
    );
    pillowBlock.position.set(0, -wheelR * 0.06, bearingZ);
    blowerGroup.add(pillowBlock);

    const motorBody = new THREE.Mesh(
      new THREE.CylinderGeometry(motorR, motorR, motorLen, 24),
      mats.motor
    );
    motorBody.rotation.x = Math.PI / 2;
    motorBody.position.z = motorCenterZ;
    blowerGroup.add(motorBody);
    const motorCap = new THREE.Mesh(
      new THREE.CylinderGeometry(motorR * 0.92, motorR * 0.98, motorLen * 0.20, 24),
      mats.motorFin
    );
    motorCap.rotation.x = Math.PI / 2;
    motorCap.position.z = motorCenterZ - motorLen * 0.55;
    blowerGroup.add(motorCap);
    for (let rIdx = -2; rIdx <= 2; rIdx++) {
      const fin = new THREE.Mesh(
        new THREE.TorusGeometry(motorR * 1.01, 0.009, 6, 24),
        mats.motorFin
      );
      fin.position.z = motorCenterZ + rIdx * (motorLen * 0.14);
      blowerGroup.add(fin);
    }
    const jBox = new THREE.Mesh(
      new THREE.BoxGeometry(motorR * 0.52, motorR * 0.36, motorR * 0.52),
      mats.motorFin
    );
    jBox.position.set(0, motorR + motorR * 0.14, motorCenterZ + motorLen * 0.08);
    blowerGroup.add(jBox);

    const basePlateY = -bY + 0.025;
    const skidCenterZ = (scrollW * 0.5 + (motorCenterZ - motorLen * 0.5)) * 0.5;
    const baseSkid = new THREE.Mesh(
      new THREE.BoxGeometry(wheelR * 2.15, 0.05, skidDepth),
      mats.darkSteel
    );
    baseSkid.position.set(wheelR * 0.15, basePlateY, skidCenterZ);
    blowerGroup.add(baseSkid);

    const cradleH = Math.max(0.08, bY - vol.rBottom - 0.03);
    const scrollCradle = new THREE.Mesh(
      new THREE.BoxGeometry(wheelR * 1.10, cradleH, scrollW + 0.04),
      mats.blowerDark
    );
    scrollCradle.position.set(0, -bY + 0.05 + cradleH * 0.5, 0);
    blowerGroup.add(scrollCradle);

    const riserTopY = -motorR + 0.02;
    const riserH = Math.max(0.08, bY + riserTopY - 0.05);
    const motorRiser = new THREE.Mesh(
      new THREE.BoxGeometry(motorR * 1.65, riserH, shaftGap + motorLen * 0.92),
      mats.blowerDark
    );
    motorRiser.position.set(0, -bY + 0.05 + riserH * 0.5, (bearingZ + motorCenterZ) * 0.5 - 0.04);
    blowerGroup.add(motorRiser);

    const rectFlange = new THREE.Mesh(
      new THREE.BoxGeometry(vol.disW + 0.075, 0.026, scrollW + 0.075),
      mats.blowerFlange
    );
    rectFlange.position.set(stackLocalX, vol.yDis, 0);
    blowerGroup.add(rectFlange);

    const rectBoltGeo = new THREE.CylinderGeometry(0.008, 0.008, 0.042, 6);
    const rbx = (vol.disW + 0.042) * 0.5;
    const rbz = (scrollW + 0.042) * 0.5;
    for (const sx of [-1, 0, 1]) {
      for (const sz of [-1, 1]) {
        const bt = new THREE.Mesh(rectBoltGeo, mats.bolt);
        bt.position.set(stackLocalX + sx * rbx, vol.yDis, sz * rbz);
        blowerGroup.add(bt);
      }
    }

    const sqToRound = new THREE.Mesh(
      buildRectToRoundGeometry(vol.disW, scrollW, stackR, transH, 36),
      mats.blower
    );
    sqToRound.position.set(stackLocalX, vol.yDis + 0.012, 0);
    blowerGroup.add(sqToRound);

    addBoltedFlange(blowerGroup, stackLocalX, stackBaseLocalY, 0, "y", stackR, mats.blowerFlange, mats.bolt, mats.gasket, 8);
  }

  const extraCondH = isDeluxe && project.deluxe_condensing_reclaimer ? 0.56 : 0;
  const stackH = clamp(pH - (bY + stackBaseLocalY) + 0.14, 0.85, 1.38) + extraCondH;
  const { exitY: stackExitLocalY, extraRight: diluterRightExtra, diluterEntry, diluterTeeY } =
    addStackAndDiluter(blowerGroup, stackLocalX, stackBaseLocalY, 0, stackR, stackH);
  const stackExitWorldY = bY + stackExitLocalY;
  assemblyGroup.add(blowerGroup);

  // Stage 3 -> Stage 4 Smooth Swept Curved Duct with Bolted Flanges
  const dirH = bDir > 0 ? vRight : vLeft;
  const R1 = Math.min(clamp(ductR * 2.0, 0.15, 0.26), Math.abs(dropX - pX) * 0.44);
  const R2 = R1;
  const R3 = Math.min(clamp(ductR * 1.85, 0.14, 0.24), Math.abs(bX - dropX) * 0.36);
  const R4 = R3;
  const frontZ = pZ + eyeZ + R4;
  const topDuctY = Math.max(topNeckY + R1 + 0.06, bY + R2 + R3 + 0.28);

  const mainDuctPath = new THREE.CurvePath<THREE.Vector3>();
  const pNeck = new THREE.Vector3(pX, topNeckY, pZ);
  const pE1Start = new THREE.Vector3(pX, topDuctY - R1, pZ);
  if (pE1Start.y > pNeck.y + 0.005) mainDuctPath.add(new THREE.LineCurve3(pNeck, pE1Start));
  const elbow1 = makeArc90Bezier(pE1Start, vUp, dirH, R1);
  mainDuctPath.add(elbow1);

  const pE1End = elbow1.getPoint(1);
  const pE2Start = new THREE.Vector3(dropX - bDir * R2, topDuctY, pZ);
  if (bDir * (pE2Start.x - pE1End.x) > 0.005) mainDuctPath.add(new THREE.LineCurve3(pE1End, pE2Start));
  const elbow2 = makeArc90Bezier(pE2Start, dirH, vDn, R2);
  mainDuctPath.add(elbow2);

  const pDropStart = elbow2.getPoint(1);
  const pDropEnd = new THREE.Vector3(dropX, bY + R3, frontZ);
  const dropSpan = Math.max(0.18, pDropStart.y - pDropEnd.y);
  mainDuctPath.add(
    new THREE.CubicBezierCurve3(
      pDropStart,
      new THREE.Vector3(dropX, pDropStart.y - dropSpan * 0.44, pZ),
      new THREE.Vector3(dropX, pDropEnd.y + dropSpan * 0.44, frontZ),
      pDropEnd
    )
  );

  const elbow3 = makeArc90Bezier(pDropEnd, vDn, dirH, R3);
  mainDuctPath.add(elbow3);

  const pE3End = elbow3.getPoint(1);
  const pE4Start = new THREE.Vector3(bX - bDir * R4, bY, frontZ);
  if (bDir * (pE4Start.x - pE3End.x) > 0.005) mainDuctPath.add(new THREE.LineCurve3(pE3End, pE4Start));

  assemblyGroup.add(
    new THREE.Mesh(new THREE.TubeGeometry(mainDuctPath, 88, ductR, 24, false), mats.cpvcSolid)
  );
  addBoltedFlange(assemblyGroup, (pE1End.x + pE2Start.x) * 0.5, topDuctY, pZ, "x", ductR, mats.cpvcFlange, mats.bolt, mats.gasket, 8);
  addBoltedFlange(assemblyGroup, pE4Start.x, bY, frontZ, "x", ductR, mats.cpvcFlange, mats.bolt, mats.gasket, 8);

  const elbow4 = makeArc90Bezier(pE4Start, dirH, vBack, R4);
  const eyePath = new THREE.CurvePath<THREE.Vector3>();
  eyePath.add(elbow4);
  assemblyGroup.add(
    new THREE.Mesh(
      new THREE.TubeGeometry(eyePath, 28, ductR, 24, false),
      cut ? mats.cutGlass : mats.cpvcSolid
    )
  );

  // Optional Phase 9 Systems / Deluxe Overlay
  addSystemsOverlay3D({
    qX, qZ, qR, qH,
    pX, pZ, pR: bundleR, pH,
    vX: vThCenterX, vY, vZ: qZ,
    bX, bY, bZ: pZ, bR: wheelR, bDepth: skidDepth,
    stX: stackX, stBaseY: bY + stackBaseLocalY, stH: stackH, stR: stackR, stZ: pZ,
    inletY, ductR,
  });

  // 3D Engineering Dimension Callouts (toggleable)
  if (showDimensions) {
    addDimensionCallout3D(
      assemblyGroup,
      new THREE.Vector3(qX - qR - 0.14, 0, qZ),
      new THREE.Vector3(qX - qR - 0.14, qH, qZ),
      new THREE.Vector3(-0.32, 0, 0),
      `Quench ${fmtDimMeters(q.tower_diameter_m, units)} × ${fmtDimMeters(q.tower_height_m, units)}`,
      "#fb923c",
      disposableTextures
    );
    addDimensionCallout3D(
      assemblyGroup,
      new THREE.Vector3(vStartX, vY - vInR - 0.12, qZ),
      new THREE.Vector3(vEndX, vY - vInR - 0.12, qZ),
      new THREE.Vector3(0, -0.22, 0.22),
      `Venturi Ø${fmtDimMeters(v.throat_diameter_m, units)} throat`,
      "#facc15",
      disposableTextures
    );
    addDimensionCallout3D(
      assemblyGroup,
      new THREE.Vector3(pX - bundleR, pH + 0.08, pZ),
      new THREE.Vector3(pX + bundleR, pH + 0.08, pZ),
      new THREE.Vector3(0, 0.36, 0),
      `Column ${fmtDimMeters(p.column_diameter_m, units)} × ${fmtDimMeters(p.column_height_m, units)}`,
      "#34d399",
      disposableTextures
    );
    addDimensionCallout3D(
      assemblyGroup,
      new THREE.Vector3(stackX + stackR + 0.14, 0, pZ),
      new THREE.Vector3(stackX + stackR + 0.14, stackExitWorldY, pZ),
      new THREE.Vector3(0.32, 0, 0),
      `Wheel ${fmtDimMeters(b.wheel_diameter_m, units)} · Stack ${fmtDimMeters(stackExitWorldY, units)}`,
      "#38bdf8",
      disposableTextures
    );
  }

  // Center & Dynamic Camera Framing
  const rawBox = new THREE.Box3().setFromObject(assemblyGroup);
  const centerX = -(rawBox.min.x + rawBox.max.x) * 0.5;
  const centerZ = -(rawBox.min.z + rawBox.max.z) * 0.5;
  assemblyGroup.position.set(centerX, 0, centerZ);

  const tanHalfFov = Math.tan(THREE.MathUtils.degToRad(cameraFovDeg * 0.5));
  const topY = Math.max(rawBox.max.y, topDuctY + ductR, stackExitWorldY) + 0.16;
  const totalW = rawBox.max.x - rawBox.min.x;
  const totalD = rawBox.max.z - rawBox.min.z;
  const allTargetY = topY * 0.48;
  const allTargetX = Math.max(0.18, totalW * 0.06);
  const allDistY = (topY * 0.56) / tanHalfFov;
  const allDistX = (totalW * 0.48) / (tanHalfFov * Math.max(1.4, cameraAspect));
  const allDist = Math.max(5.8, Math.max(allDistY, allDistX) * 1.22 + (showSystems ? totalD * 0.38 : 0));

  const blowerTopY = stackExitWorldY + 0.18;
  const blowerMidY = blowerTopY * 0.50;
  const blowerCenterX =
    bX + (hasDiluter ? stackLocalX + diluterRightExtra * 0.42 : stackLocalX * 0.55) + centerX;
  const blowerDist = Math.max(2.95, ((blowerTopY * 0.55) / tanHalfFov) * 1.18);
  const packedTopY = topDuctY + ductR + 0.12;
  const packedDist = Math.max(3.2, ((packedTopY * 0.54) / tanHalfFov) * 1.12);

  const stageViews: Record<StageKey, StageView> = {
    overview: {
      target: new THREE.Vector3(allTargetX, allTargetY, 0),
      cam: new THREE.Vector3(allTargetX + 0.72, allTargetY + topY * 0.22, allDist),
    },
    inlet: {
      target: new THREE.Vector3(qX - qR - inletLen * 0.5 + centerX, inletY + 0.14, qZ + centerZ),
      cam: new THREE.Vector3(qX - qR + 0.12 + centerX, inletY + 0.45, qZ + centerZ + 2.35),
    },
    quench: {
      target: new THREE.Vector3(qX + centerX + 0.12, qH * 0.52, qZ + centerZ),
      cam: new THREE.Vector3(qX + centerX + 0.45, qH * 0.72, qZ + centerZ + Math.max(2.4, qH * 1.55)),
    },
    venturi: {
      target: new THREE.Vector3((vStartX + vEndX) * 0.5 + centerX + 0.08, vY + 0.08, qZ + centerZ),
      cam: new THREE.Vector3((vStartX + vEndX) * 0.5 + centerX + 0.35, vY + 0.55, qZ + centerZ + 2.35),
    },
    packed: {
      target: new THREE.Vector3(pX + centerX + 0.22, packedTopY * 0.50, pZ + centerZ),
      cam: new THREE.Vector3(pX + centerX + 0.65, packedTopY * 0.62, pZ + centerZ + packedDist),
    },
    blower: {
      target: new THREE.Vector3(blowerCenterX + 0.22, blowerMidY, pZ + centerZ),
      cam: new THREE.Vector3(blowerCenterX + 0.75, blowerMidY + blowerTopY * 0.16, pZ + centerZ + blowerDist),
    },
  };

  // =========================================================================
  // EXACT ENCLOSED-DUCT & VOLUTE SPIRAL PARTICLE PATHS
  // =========================================================================
  const ox = centerX;
  const oz = centerZ;
  const primaryFlow: FlowSample[] = [];

  // Stage 1: Kiln flue -> Quench downward flow
  primaryFlow.push(
    ...sampleCurve(
      new THREE.LineCurve3(
        new THREE.Vector3(qX - qR - inletLen, inletY, qZ),
        new THREE.Vector3(qX - 0.10, inletY, qZ)
      ),
      18,
      ductR * 0.42,
      1.3,
      "inlet",
      ox,
      oz
    ),
    ...sampleCurve(
      new THREE.QuadraticBezierCurve3(
        new THREE.Vector3(qX - 0.10, inletY, qZ),
        new THREE.Vector3(qX, inletY, qZ),
        new THREE.Vector3(qX, inletY - 0.14, qZ)
      ),
      8,
      qR * 0.35,
      0.95,
      "quench",
      ox,
      oz
    ),
    ...sampleCurve(
      new THREE.LineCurve3(
        new THREE.Vector3(qX, inletY - 0.14, qZ),
        new THREE.Vector3(qX, vY + 0.12, qZ)
      ),
      26,
      qR * 0.55,
      0.72,
      "quench",
      ox,
      oz
    ),
    ...sampleCurve(
      new THREE.QuadraticBezierCurve3(
        new THREE.Vector3(qX, vY + 0.12, qZ),
        new THREE.Vector3(qX, vY, qZ),
        new THREE.Vector3(vStartX, vY, qZ)
      ),
      10,
      vInR * 0.45,
      1.15,
      "quench",
      ox,
      oz
    ),
    // Stage 2: High-velocity acceleration through Venturi throat!
    ...sampleCurve(
      new THREE.LineCurve3(
        new THREE.Vector3(vStartX, vY, qZ),
        new THREE.Vector3(vStartX + vConvL, vY, qZ)
      ),
      10,
      vThR * 0.65,
      1.8,
      "venturi",
      ox,
      oz
    ),
    ...sampleCurve(
      new THREE.LineCurve3(
        new THREE.Vector3(vStartX + vConvL, vY, qZ),
        new THREE.Vector3(vStartX + vConvL + vThL, vY, qZ)
      ),
      10,
      vThR * 0.45,
      3.4,
      "venturi",
      ox,
      oz
    ),
    ...sampleCurve(
      new THREE.LineCurve3(
        new THREE.Vector3(vStartX + vConvL + vThL, vY, qZ),
        new THREE.Vector3(vEndX, vY, qZ)
      ),
      14,
      vOutR * 0.45,
      1.75,
      "venturi",
      ox,
      oz
    )
  );

  // Stage 2 -> Stage 3 transfer
  if (isCompact && uTurnPts.length > 1) {
    for (const pt of uTurnPts) {
      primaryFlow.push({
        pt: new THREE.Vector3(pt.x + ox, pt.y, pt.z + oz),
        swirlRadius: vOutR * 0.42,
        speedMul: 1.25,
        stage: "venturi",
      });
    }
    primaryFlow.push(
      ...sampleCurve(
        new THREE.LineCurve3(
          new THREE.Vector3(pX + effPR, vY, pZ),
          new THREE.Vector3(pX, vY + 0.14, pZ)
        ),
        10,
        vOutR * 0.45,
        0.95,
        "packed",
        ox,
        oz
      )
    );
  } else {
    primaryFlow.push(
      ...sampleCurve(
        new THREE.LineCurve3(
          new THREE.Vector3(vEndX, vY, pZ),
          new THREE.Vector3(pX - 0.12, vY, pZ)
        ),
        10,
        vOutR * 0.45,
        1.15,
        "venturi",
        ox,
        oz
      ),
      ...sampleCurve(
        new THREE.QuadraticBezierCurve3(
          new THREE.Vector3(pX - 0.12, vY, pZ),
          new THREE.Vector3(pX, vY, pZ),
          new THREE.Vector3(pX, vY + 0.14, pZ)
        ),
        8,
        bundleR * 0.32,
        0.85,
        "packed",
        ox,
        oz
      )
    );
  }

  // Stage 3: Upward flow through packed bed & demister
  primaryFlow.push(
    ...sampleCurve(
      new THREE.LineCurve3(
        new THREE.Vector3(pX, vY + 0.14, pZ),
        new THREE.Vector3(pX, topNeckY, pZ)
      ),
      48,
      bundleR * (isParallelBed ? 0.60 : 0.50),
      0.56,
      "packed",
      ox,
      oz
    ),
    // Overhead curved CPVC duct + front eye elbow
    ...sampleCurve(mainDuctPath, 52, ductR * 0.42, 1.35, "blower", ox, oz),
    ...sampleCurve(elbow4, 16, ductR * 0.42, 1.45, "blower", ox, oz),
    ...sampleCurve(
      new THREE.LineCurve3(
        new THREE.Vector3(bX, bY, pZ + eyeZ),
        new THREE.Vector3(bX, bY, pZ)
      ),
      8,
      ductR * 0.38,
      1.6,
      "blower",
      ox,
      oz
    )
  );

  // Spiral volute path around spinning impeller -> tangential discharge -> exhaust stack -> plume
  const swirlN = 22;
  let prevSwirl = new THREE.Vector3(bX, bY, pZ);
  for (let k = 1; k <= swirlN; k++) {
    const u = k / swirlN;
    const ang = -1.15 * Math.PI + u * (1.15 * Math.PI);
    const rr = wheelR * 0.14 + u * (stackLocalX - wheelR * 0.14);
    const cur = new THREE.Vector3(bX + Math.cos(ang) * rr, bY + Math.sin(ang) * rr, pZ);
    primaryFlow.push({
      pt: new THREE.Vector3(cur.x + ox, cur.y, cur.z + oz),
      swirlRadius: 0.022,
      speedMul: 2.3,
      stage: "blower",
    });
    prevSwirl = cur;
  }

  primaryFlow.push(
    ...sampleCurve(
      new THREE.LineCurve3(prevSwirl, new THREE.Vector3(stackX, stackExitWorldY, pZ)),
      26,
      stackR * 0.45,
      1.65,
      "blower",
      ox,
      oz
    ),
    ...sampleCurve(
      new THREE.LineCurve3(
        new THREE.Vector3(stackX, stackExitWorldY, pZ),
        new THREE.Vector3(stackX, stackExitWorldY + 0.65, pZ)
      ),
      16,
      stackR * 1.15,
      1.2,
      "plume",
      ox,
      oz
    )
  );

  // Optional 120V Plume Diluter ambient mixing stream
  const diluterFlow: FlowSample[] = [];
  if (hasDiluter && diluterEntry) {
    const worldEntry = new THREE.Vector3(
      bX + diluterEntry.x,
      bY + diluterEntry.y,
      pZ + diluterEntry.z
    );
    const worldTee = new THREE.Vector3(stackX, bY + diluterTeeY, pZ);
    diluterFlow.push(
      ...sampleCurve(
        new THREE.LineCurve3(worldEntry, worldTee),
        18,
        0.025,
        1.8,
        "diluter",
        ox,
        oz
      ),
      ...sampleCurve(
        new THREE.LineCurve3(worldTee, new THREE.Vector3(stackX, stackExitWorldY + 0.65, pZ)),
        24,
        stackR * 0.58,
        1.85,
        "diluter",
        ox,
        oz
      )
    );
  }

  sprayZones.push(
    {
      topCenter: new THREE.Vector3(qX + ox, sprayTopY, qZ + oz),
      topRadius: qR * 0.28,
      botRadius: qR * 0.72,
      height: Math.max(0.55, qH - 0.55),
      color: 0x38bdf8,
      count: 34,
      speed: 1.25,
    },
    {
      topCenter: new THREE.Vector3(pX + ox, distRingY, pZ + oz),
      topRadius: bundleR * 0.56,
      botRadius: bundleR * 0.78,
      height: sprayBedH,
      color: 0x22d3ee,
      count: 42,
      speed: 0.85,
    }
  );

  return {
    stageViews,
    primaryFlow,
    motiveFlow: [],
    diluterFlow,
    sprayZones,
    impellerMesh,
    impellerDir: 1,
    diluterImpellerMesh,
    architectureSignature: `inline|${isParallelBed ? "par" : "col"}|${isParallelVent ? "pv" : "sv"}|${blowerDrive}|${isCompact ? "u" : "lin"}|${showSystems ? "sys" : "std"}|${hasDiluter ? "dil" : "nodil"}|${project.deluxe_condensing_reclaimer ? "cond" : "nocond"}|${project.deluxe_stage1_heat_recovery ? "s1h" : "nos1h"}|${project.deluxe_variable_throat_vfd ? "vt" : "novt"}`,
    scaleSummary: `Scale 1:1 m · ${fmtDimMeters(env.length_m, units)} L × ${fmtDimMeters(env.width_m, units)} W × ${fmtDimMeters(env.height_m, units)} H`,
    disposableTextures,
  };
}
