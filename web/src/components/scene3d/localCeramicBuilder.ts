import * as THREE from "three";
import type { MaterialBundle } from "./materials";
import {
  addBallValve,
  addBellSpigotCollar,
  addDimensionCallout3D,
  addPipeUnion,
  addPressureGauge,
  clamp,
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

function samplePathWithMeta(
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

export function buildLocalCeramicAssembly(
  assemblyGroup: THREE.Group,
  mats: MaterialBundle,
  opts: SceneBuildOptions
): SceneBuildResult {
  const { evaluation, cutaway: cut, showDimensions, units, cameraFovDeg, cameraAspect } = opts;
  const { project, frugal: fr } = evaluation;
  const disposableTextures: THREE.Texture[] = [];

  const uiDraft = String(
    project.local_draft_mode || fr?.draft_mode || "natural_chimney"
  ).toLowerCase();
  const steamEnabled =
    typeof fr?.steam_enabled === "boolean" &&
    String(fr.draft_mode || "").toLowerCase() === uiDraft
      ? fr.steam_enabled
      : uiDraft === "steam_ejector_subsonic" ||
        uiDraft === "steam_ejector_lp" ||
        uiDraft === "steam_ejector_hp";
  const isSubsonic =
    steamEnabled &&
    (uiDraft === "steam_ejector_subsonic" ||
      (typeof fr?.steam_nozzle_mach === "number" &&
        fr.steam_nozzle_mach > 0 &&
        fr.steam_nozzle_mach < 1.0));
  const isHP =
    steamEnabled &&
    !isSubsonic &&
    (uiDraft === "steam_ejector_hp" ||
      (typeof fr?.steam_pressure_psig === "number" && fr.steam_pressure_psig >= 40));
  const isLP = steamEnabled && !isSubsonic && !isHP;
  const isZeroElec = project.zero_electricity;
  const boilerSource = String(
    project.local_boiler_source || fr?.boiler_source || "dedicated_hearth"
  ).toLowerCase();
  const pumpMode = String(
    project.local_pump_mode || fr?.pump_mode || "manual_treadle"
  ).toLowerCase();
  const isSumpPoolOnly = steamEnabled && pumpMode === "sump_pool_no_pump";
  const isSteamPPCP = steamEnabled && pumpMode === "steam_ppcp";
  const isSteamSyphon = steamEnabled && pumpMode === "steam_jet_syphon";
  const hasOverheadTank = !isSumpPoolOnly;

  const defaultChimM = steamEnabled ? (isHP ? 4.5 : isSubsonic ? 5.3 : 5.0) : isZeroElec ? 9.5 : 4.5;
  const realChimM =
    typeof fr?.chimney_height_m === "number" &&
    Number.isFinite(fr.chimney_height_m) &&
    fr.chimney_height_m > 0
      ? fr.chimney_height_m
      : defaultChimM;
  const bedDiamIn =
    typeof fr?.bed_diameter_in === "number" && Number.isFinite(fr.bed_diameter_in)
      ? fr.bed_diameter_in
      : 26;
  const bedDiamM = clamp(bedDiamIn * 0.0254, 0.56, 0.96);
  const tankLiters = fr?.gravity_tank_liters || 250;
  const tankElevM = (fr?.tank_elevation_ft || 11.5) * 0.3048;

  const thetaStart = cut ? Math.PI * 0.5 : 0;
  const thetaLen = cut ? Math.PI : Math.PI * 2;

  // Concrete foundation plinth (compact when no overhead water tower!)
  const plinthW = isSumpPoolOnly ? 4.9 : steamEnabled ? 6.4 : 5.8;
  const plinthD = steamEnabled ? 3.0 : 2.6;
  const plinthY = 0.10;
  const plinth = new THREE.Mesh(
    new THREE.BoxGeometry(plinthW, plinthY, plinthD),
    mats.plinth
  );
  plinth.position.set(isSumpPoolOnly ? -0.20 : 0.55, plinthY * 0.5, 0.05);
  assemblyGroup.add(plinth);

  // =========================================================================
  // 1. REFRACTORY BRICK KILN INLET FLUE + QUENCH & ASH-SETTLING CHAMBER
  // =========================================================================
  const qX = steamEnabled ? -1.90 : -1.68;
  const qZ = 0;
  const qW = 0.90;
  const qH = 1.56;
  const qD = 0.90;
  const qInY = 1.24;
  const wallT = 0.10;

  // Hot refractory brick kiln flue duct + glowing core + mortar entry collar
  const flueLen = 0.72;
  const flueIn = new THREE.Mesh(
    new THREE.BoxGeometry(flueLen, 0.34, 0.34),
    mats.brickSolid
  );
  flueIn.position.set(qX - qW * 0.5 - flueLen * 0.5 + 0.02, qInY, qZ);
  assemblyGroup.add(flueIn);

  const flueGlow = new THREE.Mesh(
    new THREE.BoxGeometry(flueLen + 0.02, 0.21, 0.21),
    mats.hotCore
  );
  flueGlow.position.set(qX - qW * 0.5 - flueLen * 0.5 + 0.02, qInY, qZ);
  assemblyGroup.add(flueGlow);

  const flueCollar = new THREE.Mesh(
    new THREE.BoxGeometry(0.10, 0.42, 0.42),
    mats.mortar
  );
  flueCollar.position.set(qX - qW * 0.5 - 0.03, qInY, qZ);
  assemblyGroup.add(flueCollar);

  // Brick quench walls (front wall opens when cutaway=true)
  const qBack = new THREE.Mesh(new THREE.BoxGeometry(qW, qH, wallT), mats.brickSolid);
  qBack.position.set(qX, plinthY + qH * 0.5, qZ - qD * 0.5 + wallT * 0.5);
  assemblyGroup.add(qBack);
  const qLeft = new THREE.Mesh(new THREE.BoxGeometry(wallT, qH, qD), mats.brickSolid);
  qLeft.position.set(qX - qW * 0.5 + wallT * 0.5, plinthY + qH * 0.5, qZ);
  assemblyGroup.add(qLeft);
  const qRight = new THREE.Mesh(new THREE.BoxGeometry(wallT, qH, qD), mats.brickSolid);
  qRight.position.set(qX + qW * 0.5 - wallT * 0.5, plinthY + qH * 0.5, qZ);
  assemblyGroup.add(qRight);
  const qCap = new THREE.Mesh(new THREE.BoxGeometry(qW + 0.10, 0.09, qD + 0.10), mats.mortar);
  qCap.position.set(qX, plinthY + qH + 0.045, qZ);
  assemblyGroup.add(qCap);
  if (!cut) {
    const qFront = new THREE.Mesh(new THREE.BoxGeometry(qW, qH, wallT), mats.brickSolid);
    qFront.position.set(qX, plinthY + qH * 0.5, qZ + qD * 0.5 - wallT * 0.5);
    assemblyGroup.add(qFront);
  }

  // Internal hanging terracotta baffle curtain + brass gravity spray rose (OR submerged S-baffle in sump-pool mode!)
  const qBaffle = new THREE.Mesh(
    new THREE.BoxGeometry(0.055, isSumpPoolOnly ? qH * 0.72 : qH * 0.46, qD - wallT * 2.1),
    mats.terracotta
  );
  qBaffle.position.set(qX + 0.12, plinthY + (isSumpPoolOnly ? qH * 0.46 : qH * 0.72), qZ);
  assemblyGroup.add(qBaffle);

  if (hasOverheadTank) {
    const qRose = new THREE.Mesh(
      new THREE.CylinderGeometry(0.032, 0.082, 0.11, 18),
      mats.brass
    );
    qRose.position.set(qX - 0.08, plinthY + qH - 0.12, qZ);
    assemblyGroup.add(qRose);

    const qSpray = new THREE.Mesh(
      new THREE.ConeGeometry(qW * 0.34, qH * 0.70, 22, 1, true),
      mats.sprayMist
    );
    qSpray.position.set(qX - 0.08, plinthY + qH * 0.44, qZ);
    assemblyGroup.add(qSpray);
  }

  const qPoolH = isSumpPoolOnly ? 0.38 : 0.20;
  const qPool = new THREE.Mesh(
    new THREE.BoxGeometry(qW - wallT * 2, qPoolH, qD - wallT * 2),
    mats.sumpWater
  );
  qPool.position.set(qX, plinthY + qPoolH * 0.5, qZ);
  assemblyGroup.add(qPool);

  // Phase 13: Optional Stage 1 Water-Jacket Collar OR Kiln Exhaust Steam Coil (SiC Muffled)
  if (steamEnabled && boilerSource === "stage1_water_jacket") {
    const jH = 0.56;
    const jacket = new THREE.Mesh(
      new THREE.BoxGeometry(qW + 0.14, jH, qD + 0.14),
      mats.steamerSS
    );
    jacket.position.set(qX, plinthY + qH - jH * 0.48, qZ);
    assemblyGroup.add(jacket);
    const jBand = new THREE.Mesh(
      new THREE.BoxGeometry(qW + 0.17, 0.04, qD + 0.17),
      mats.steamHeader
    );
    jBand.position.set(qX, plinthY + qH - 0.04, qZ);
    assemblyGroup.add(jBand);
  } else if (steamEnabled && boilerSource === "kiln_exhaust_coil") {
    const muffleX = qX - qW * 0.5 - flueLen * 0.52;
    const muffleBox = new THREE.Mesh(
      new THREE.BoxGeometry(0.46, 0.50, 0.50),
      mats.mortar
    );
    muffleBox.position.set(muffleX, qInY + 0.04, qZ);
    assemblyGroup.add(muffleBox);
    const coilPts: THREE.Vector3[] = [];
    for (let ci = 0; ci <= 48; ci++) {
      const u = ci / 48;
      const a = u * 4.5 * Math.PI * 2;
      coilPts.push(
        new THREE.Vector3(
          muffleX - 0.16 + u * 0.32,
          qInY + Math.cos(a) * 0.18,
          qZ + Math.sin(a) * 0.18
        )
      );
    }
    assemblyGroup.add(
      new THREE.Mesh(
        new THREE.TubeGeometry(new THREE.CatmullRomCurve3(coilPts), 64, 0.015, 8, false),
        mats.steamerSS
      )
    );
  }

  // =========================================================================
  // 2. STAGE 2: STEAM-JET EJECTOR (SUBSONIC / LP MONOTUBE / HP BOILER) OR BRICK BAFFLE FLUE
  // =========================================================================
  const pX = steamEnabled ? -0.05 : -0.22;
  const pZ = 0;
  const pR = clamp(bedDiamM * 0.5, 0.32, 0.52);
  const pH = 2.42;
  const ejY = 0.60;
  const ejStart = qX + qW * 0.5 - 0.02;
  const ejEnd = pX - pR + 0.06;
  const ejLen = Math.max(0.62, ejEnd - ejStart);
  const ejMidX = (ejStart + ejEnd) * 0.5;

  let steamHdrCurve: THREE.CatmullRomCurve3 | null = null;
  let nozzleTipX = ejStart + ejLen * 0.28;
  let standpipeTopPt: THREE.Vector3 | null = null;
  let standpipeBasePt: THREE.Vector3 | null = null;

  if (steamEnabled) {
    const boX = qX - 0.28;
    const boZ = -0.94;
    const boR = isSubsonic ? 0.30 : isHP ? 0.28 : 0.25;
    const boH = isSubsonic ? 0.66 : isHP ? 1.18 : 0.96;

    // Refractory firebox plinth + waste-heat tap from kiln flue
    const boBase = new THREE.Mesh(
      new THREE.CylinderGeometry(boR * 1.16, boR * 1.24, 0.18, 24),
      mats.brickSolid
    );
    boBase.position.set(boX, plinthY + 0.09, boZ);
    assemblyGroup.add(boBase);

    const wasteTap = new THREE.Mesh(
      new THREE.BoxGeometry(0.24, 0.26, Math.abs(boZ - qZ) - 0.10),
      mats.brickSolid
    );
    wasteTap.position.set(qX - qW * 0.5 - 0.12, 0.72, (boZ + qZ) * 0.5);
    assemblyGroup.add(wasteTap);

    const vesselBaseY = plinthY + 0.18;

    if (isSubsonic) {
      // 6 psig Commercial Brushed Stainless Catering Steamer / Kettle
      const stCyl = new THREE.Mesh(
        new THREE.CylinderGeometry(boR, boR, boH, 28),
        mats.steamerSS
      );
      stCyl.position.set(boX, vesselBaseY + boH * 0.5, boZ);
      assemblyGroup.add(stCyl);

      const lidRim = new THREE.Mesh(
        new THREE.TorusGeometry(boR + 0.015, 0.018, 10, 28),
        mats.steelTrim
      );
      lidRim.rotation.x = Math.PI / 2;
      lidRim.position.set(boX, vesselBaseY + boH, boZ);
      assemblyGroup.add(lidRim);

      const lidCap = new THREE.Mesh(
        new THREE.CylinderGeometry(boR * 0.84, boR * 1.02, 0.11, 24),
        mats.steamerSS
      );
      lidCap.position.set(boX, vesselBaseY + boH + 0.055, boZ);
      assemblyGroup.add(lidCap);

      for (let li = 0; li < 6; li++) {
        const ang = (li / 6) * Math.PI * 2;
        const latch = new THREE.Mesh(
          new THREE.BoxGeometry(0.028, 0.075, 0.024),
          mats.iron
        );
        latch.position.set(
          boX + Math.cos(ang) * (boR + 0.018),
          vesselBaseY + boH - 0.01,
          boZ + Math.sin(ang) * (boR + 0.018)
        );
        latch.rotation.y = -ang;
        assemblyGroup.add(latch);
      }

      // Tall 4.2 m Open-Top Hydrostatic Water-Column Relief Standpipe (S2)
      const spX = boX - boR - 0.20;
      const spZ = boZ + 0.04;
      const spBaseY = vesselBaseY + 0.10;
      const spVisualH = 2.95; // scaled cleanly alongside 5.3m chimney; exact 4.2 m shown in callout
      const spCopperH = 0.84;
      const spPvcH = spVisualH - spCopperH;
      standpipeBasePt = new THREE.Vector3(spX, spBaseY, spZ);
      standpipeTopPt = new THREE.Vector3(spX, spBaseY + spVisualH + 0.16, spZ);

      const spTee = new THREE.Mesh(
        new THREE.CylinderGeometry(0.025, 0.025, 0.24, 12),
        mats.steamHeader
      );
      spTee.rotation.z = Math.PI / 2;
      spTee.position.set((boX - boR + spX) * 0.5, spBaseY, spZ);
      assemblyGroup.add(spTee);
      addPipeUnion(assemblyGroup, spX + 0.07, spBaseY, spZ, "x", 0.025, mats.brass, mats.iron);

      const spCopper = new THREE.Mesh(
        new THREE.CylinderGeometry(0.030, 0.030, spCopperH, 16),
        mats.steamHeader
      );
      spCopper.position.set(spX, spBaseY + spCopperH * 0.5, spZ);
      assemblyGroup.add(spCopper);
      addPipeUnion(assemblyGroup, spX, spBaseY + spCopperH, spZ, "y", 0.033, mats.brass, mats.iron);

      const spPvc = new THREE.Mesh(
        new THREE.CylinderGeometry(0.036, 0.036, spPvcH, 18, 1, true),
        mats.standpipePVC
      );
      spPvc.position.set(spX, spBaseY + spCopperH + spPvcH * 0.5, spZ);
      assemblyGroup.add(spPvc);

      const spWater = new THREE.Mesh(
        new THREE.CylinderGeometry(0.025, 0.025, spPvcH * 0.88, 14),
        mats.sumpWater
      );
      spWater.position.set(spX, spBaseY + spCopperH + spPvcH * 0.44, spZ);
      assemblyGroup.add(spWater);

      const spFunnel = new THREE.Mesh(
        new THREE.CylinderGeometry(0.11, 0.038, 0.16, 20, 1, true),
        mats.steamHeader
      );
      spFunnel.position.set(spX, spBaseY + spVisualH + 0.08, spZ);
      assemblyGroup.add(spFunnel);

      const spLip = new THREE.Mesh(
        new THREE.TorusGeometry(0.11, 0.010, 8, 20),
        mats.brass
      );
      spLip.rotation.x = Math.PI / 2;
      spLip.position.set(spX, spBaseY + spVisualH + 0.16, spZ);
      assemblyGroup.add(spLip);

      // Backup brass 10-15 psig PRV on steamer lid
      const bPrv = new THREE.Mesh(
        new THREE.CylinderGeometry(0.020, 0.024, 0.10, 12),
        mats.brass
      );
      bPrv.position.set(boX + 0.12, vesselBaseY + boH + 0.14, boZ + 0.05);
      assemblyGroup.add(bPrv);

      for (const bh of [0.95, 2.05]) {
        const brk = new THREE.Mesh(new THREE.BoxGeometry(0.24, 0.035, 0.035), mats.iron);
        brk.position.set(spX + 0.10, spBaseY + bh, spZ);
        assemblyGroup.add(brk);
      }
    } else if (isLP) {
      // 20 psig Code-Rated Seamless 316L Helical Monotube Flash Coil (< 2 L stored water)
      const firebox = new THREE.Mesh(
        new THREE.CylinderGeometry(boR, boR, boH, 24, 1, cut, thetaStart, thetaLen),
        mats.brick
      );
      firebox.position.set(boX, vesselBaseY + boH * 0.5, boZ);
      assemblyGroup.add(firebox);

      // Visible helical 316L seamless monotube coil inside the firebox
      const coilPts: THREE.Vector3[] = [];
      const turns = 6.5;
      const coilR = boR * 0.64;
      for (let i = 0; i <= 90; i++) {
        const u = i / 90;
        const a = u * turns * Math.PI * 2;
        coilPts.push(
          new THREE.Vector3(
            boX + Math.cos(a) * coilR,
            vesselBaseY + 0.10 + u * (boH - 0.18),
            boZ + Math.sin(a) * coilR
          )
        );
      }
      const coilCurve = new THREE.CatmullRomCurve3(coilPts);
      assemblyGroup.add(
        new THREE.Mesh(new THREE.TubeGeometry(coilCurve, 120, 0.017, 10, false), mats.steamerSS)
      );

      const fbCap = new THREE.Mesh(
        new THREE.CylinderGeometry(boR * 1.06, boR * 1.06, 0.06, 24),
        mats.mortar
      );
      fbCap.position.set(boX, vesselBaseY + boH + 0.03, boZ);
      assemblyGroup.add(fbCap);
    } else {
      // 60 psig Certified Non-IBR / KEBS / ASME Commercial Baby Steam Boiler (<= 25 L)
      const boCyl = new THREE.Mesh(
        new THREE.CylinderGeometry(boR, boR, boH, 28),
        mats.boilerSteel
      );
      boCyl.position.set(boX, vesselBaseY + boH * 0.5, boZ);
      assemblyGroup.add(boCyl);

      const boDome = new THREE.Mesh(
        new THREE.SphereGeometry(boR, 24, 14, 0, Math.PI * 2, 0, Math.PI * 0.5),
        mats.boilerSteel
      );
      boDome.position.set(boX, vesselBaseY + boH, boZ);
      assemblyGroup.add(boDome);

      for (const bf of [0.22, 0.52, 0.82]) {
        const strap = new THREE.Mesh(
          new THREE.TorusGeometry(boR + 0.008, 0.010, 8, 28),
          mats.steelTrim
        );
        strap.rotation.x = Math.PI / 2;
        strap.position.set(boX, vesselBaseY + boH * bf, boZ);
        assemblyGroup.add(strap);
      }

      const boStack = new THREE.Mesh(
        new THREE.CylinderGeometry(0.062, 0.070, 0.38, 16),
        mats.iron
      );
      boStack.position.set(boX - 0.10, vesselBaseY + boH + boR * 0.72 + 0.16, boZ - 0.06);
      assemblyGroup.add(boStack);
    }

    // Bourdon pressure gauge facing viewer
    addPressureGauge(
      assemblyGroup,
      boX + 0.09,
      vesselBaseY + boH * (isSubsonic ? 0.72 : 0.80),
      boZ + boR + 0.05,
      isHP ? -0.68 : isSubsonic ? 0.36 : -0.22,
      mats.brass,
      mats.dialWhite,
      isSubsonic ? mats.valveGreen : mats.safetyRed
    );

    if (!isSubsonic) {
      // Stamped ASME/IBR spring-loaded PRV + outdoor vertical vent discharge pipe
      const prvX = boX + 0.10;
      const prvY = vesselBaseY + boH + (isHP ? boR * 0.78 : 0.06);
      const prvZ = boZ + 0.06;
      const prvBody = new THREE.Mesh(
        new THREE.CylinderGeometry(0.025, 0.030, 0.11, 12),
        mats.brass
      );
      prvBody.position.set(prvX, prvY + 0.055, prvZ);
      assemblyGroup.add(prvBody);
      const prvCap = new THREE.Mesh(
        new THREE.CylinderGeometry(0.032, 0.032, 0.055, 14),
        mats.safetyRed
      );
      prvCap.position.set(prvX, prvY + 0.13, prvZ);
      assemblyGroup.add(prvCap);
      const prvVent = new THREE.Mesh(
        new THREE.CylinderGeometry(0.016, 0.016, 0.52, 12),
        mats.safetyRed
      );
      prvVent.position.set(prvX + 0.045, prvY + 0.30, prvZ);
      assemblyGroup.add(prvVent);
    }

    // Water-level Sight Glass column on boiler/steamer
    if (!isLP) {
      const sgX = boX - 0.12;
      const sgZ = boZ + boR + 0.036;
      const sgH = isSubsonic ? 0.30 : 0.44;
      const sgTube = new THREE.Mesh(
        new THREE.CylinderGeometry(0.014, 0.014, sgH, 12),
        mats.sightGlass
      );
      sgTube.position.set(sgX, vesselBaseY + boH * 0.50, sgZ);
      assemblyGroup.add(sgTube);
      for (const sy of [vesselBaseY + boH * 0.28, vesselBaseY + boH * 0.72]) {
        const cock = new THREE.Mesh(
          new THREE.CylinderGeometry(0.017, 0.017, 0.06, 10),
          mats.brass
        );
        cock.rotation.x = Math.PI / 2;
        cock.position.set(sgX, sy, boZ + boR + 0.015);
        assemblyGroup.add(cock);
      }
    }

    // Insulated Copper Steam Header (S1) with Brass Isolation Handwheel + Unions -> Ejector Axis
    const hdrStartY = isSubsonic
      ? vesselBaseY + boH + 0.10
      : isHP
        ? vesselBaseY + boH + boR * 0.92
        : vesselBaseY + boH + 0.05;
    const hdrPeakY = vesselBaseY + boH + 0.56;
    const steamDropX = ejStart + 0.08;
    const steamPts = [
      new THREE.Vector3(boX + 0.02, hdrStartY, boZ),
      new THREE.Vector3(boX + 0.02, hdrPeakY, boZ),
      new THREE.Vector3(steamDropX, hdrPeakY, boZ * 0.45),
      new THREE.Vector3(steamDropX, ejY + 0.26, qZ),
      new THREE.Vector3(steamDropX + 0.08, ejY, qZ),
    ];
    steamHdrCurve = new THREE.CatmullRomCurve3(steamPts);
    assemblyGroup.add(
      new THREE.Mesh(
        new THREE.TubeGeometry(steamHdrCurve, 44, 0.028, 12, false),
        mats.steamHeader
      )
    );
    const isoValve = new THREE.Mesh(
      new THREE.CylinderGeometry(0.036, 0.036, 0.075, 14),
      mats.brass
    );
    isoValve.position.set(steamDropX, ejY + 0.44, qZ + 0.02);
    assemblyGroup.add(isoValve);
    const isoWheel = new THREE.Mesh(
      new THREE.TorusGeometry(0.060, 0.011, 8, 18),
      isSubsonic ? mats.valveGreen : mats.safetyRed
    );
    isoWheel.position.set(steamDropX, ejY + 0.44, qZ + 0.08);
    assemblyGroup.add(isoWheel);

    // Venturi Ejector Mixing Tube (Terracotta Converging -> Steel Throat -> Terracotta Diffuser)
    const convL = ejLen * 0.34;
    const thL = ejLen * 0.24;
    const divL = ejLen * 0.42;

    addBellSpigotCollar(assemblyGroup, ejStart + 0.02, ejY, qZ, "x", 0.19, mats.terracotta, mats.mortar);
    const convMesh = new THREE.Mesh(
      new THREE.CylinderGeometry(0.105, 0.195, convL, 22, 1, cut, thetaStart, thetaLen),
      mats.terracotta
    );
    convMesh.rotation.z = -Math.PI / 2;
    convMesh.position.set(ejStart + convL * 0.5, ejY, qZ);
    assemblyGroup.add(convMesh);

    const throatMesh = new THREE.Mesh(
      new THREE.CylinderGeometry(0.10, 0.10, thL, 22, 1, cut, thetaStart, thetaLen),
      mats.steelTrim
    );
    throatMesh.rotation.z = -Math.PI / 2;
    throatMesh.position.set(ejStart + convL + thL * 0.5, ejY, qZ);
    assemblyGroup.add(throatMesh);

    const divMesh = new THREE.Mesh(
      new THREE.CylinderGeometry(0.205, 0.105, divL, 22, 1, cut, thetaStart, thetaLen),
      mats.terracotta
    );
    divMesh.rotation.z = -Math.PI / 2;
    divMesh.position.set(ejStart + convL + thL + divL * 0.5, ejY, qZ);
    assemblyGroup.add(divMesh);
    addBellSpigotCollar(assemblyGroup, ejEnd - 0.04, ejY, qZ, "x", 0.20, mats.terracotta, mats.mortar);

    nozzleTipX = ejStart + convL * 0.65;
    if (isSubsonic) {
      const capHex = new THREE.Mesh(
        new THREE.CylinderGeometry(0.050, 0.050, 0.055, 6),
        mats.brass
      );
      capHex.rotation.z = -Math.PI / 2;
      capHex.position.set(ejStart + convL * 0.48, ejY, qZ);
      assemblyGroup.add(capHex);
      const capTip = new THREE.Mesh(
        new THREE.CylinderGeometry(0.022, 0.046, 0.065, 16),
        mats.brass
      );
      capTip.rotation.z = -Math.PI / 2;
      capTip.position.set(ejStart + convL * 0.48 + 0.06, ejY, qZ);
      assemblyGroup.add(capTip);
    } else {
      const lavalNozzle = new THREE.Mesh(
        new THREE.ConeGeometry(0.044, 0.19, 16),
        mats.brass
      );
      lavalNozzle.rotation.z = -Math.PI / 2;
      lavalNozzle.position.set(nozzleTipX, ejY, qZ);
      assemblyGroup.add(lavalNozzle);
    }

    // Visual steam plume cone inside ejector throat & tower base
    const plumeLen = thL + divL + pR * 0.78;
    const plumeCone = new THREE.Mesh(
      new THREE.ConeGeometry(isSubsonic ? 0.14 : 0.17, plumeLen, 20, 1, true),
      mats.steamPlume
    );
    plumeCone.rotation.z = Math.PI / 2;
    plumeCone.position.set(ejStart + convL + plumeLen * 0.45, ejY, qZ);
    assemblyGroup.add(plumeCone);
  } else {
    // Passive masonry transfer flue with mortar joints & internal baffle
    const xferLen = ejEnd - ejStart + 0.08;
    const xferFlue = new THREE.Mesh(
      new THREE.BoxGeometry(xferLen, 0.40, 0.44),
      mats.brickSolid
    );
    xferFlue.position.set(ejMidX, ejY, qZ);
    assemblyGroup.add(xferFlue);
    const xferCollar = new THREE.Mesh(
      new THREE.BoxGeometry(0.09, 0.48, 0.52),
      mats.mortar
    );
    xferCollar.position.set(ejMidX, ejY, qZ);
    assemblyGroup.add(xferCollar);
  }

  // =========================================================================
  // 3. STAGE 3: WIDE MASONRY / STONEWARE PACKED ABSORBER TOWER
  // =========================================================================
  const pShell = new THREE.Mesh(
    new THREE.CylinderGeometry(pR, pR * 1.03, pH, 32, 1, cut, thetaStart, thetaLen),
    mats.brick
  );
  pShell.position.set(pX, plinthY + pH * 0.5, pZ);
  assemblyGroup.add(pShell);

  // Iron tension hoop bands with bolted turnbuckle lugs
  for (const hf of [0.22, 0.50, 0.78, 0.96]) {
    const hoop = new THREE.Mesh(
      new THREE.TorusGeometry(pR * 1.035, 0.014, 8, 32, cut ? Math.PI : Math.PI * 2),
      mats.iron
    );
    hoop.rotation.x = Math.PI / 2;
    if (cut) hoop.rotation.z = Math.PI * 0.5;
    hoop.position.set(pX, plinthY + pH * hf, pZ);
    assemblyGroup.add(hoop);
  }

  // Bottom sump pool & perforated clay brick support arch
  const pPool = new THREE.Mesh(
    new THREE.CylinderGeometry(pR * 0.92, pR * 0.92, 0.22, 24),
    mats.sumpWater
  );
  pPool.position.set(pX, plinthY + 0.11, pZ);
  assemblyGroup.add(pPool);

  const pGrate = new THREE.Mesh(
    new THREE.CylinderGeometry(pR * 0.92, pR * 0.92, 0.05, 24),
    mats.mortar
  );
  pGrate.position.set(pX, 0.92, pZ);
  assemblyGroup.add(pGrate);

  // Textured stoneware clay Raschig ring bed + visible individual rings in cutaway
  const bedBaseY = 0.95;
  const bedH = 1.05;
  const bedCore = new THREE.Mesh(
    new THREE.CylinderGeometry(pR * 0.87, pR * 0.87, bedH, 26),
    mats.packingCeramic
  );
  bedCore.position.set(pX, bedBaseY + bedH * 0.5, pZ);
  assemblyGroup.add(bedCore);

  if (cut) {
    const ringGeo = new THREE.TorusGeometry(0.034, 0.011, 6, 12);
    for (let ri = 0; ri < 42; ri++) {
      const row = Math.floor(ri / 7);
      const col = ri % 7;
      const ang = -0.40 + (col / 6) * 0.80;
      const ry = bedBaseY + 0.08 + row * 0.16 + (col % 2) * 0.05;
      const rr = pR * 0.88;
      const ringMesh = new THREE.Mesh(ringGeo, mats.stoneware);
      ringMesh.position.set(pX + Math.sin(ang) * rr, ry, pZ + Math.cos(ang) * rr);
      ringMesh.rotation.set(ri * 0.7, ri * 1.1, ri * 0.4);
      assemblyGroup.add(ringMesh);
    }
  }

  // Brass gravity trickle distribution ring + spray cone
  const trickleY = bedBaseY + bedH + 0.09;
  const trickleRing = new THREE.Mesh(
    new THREE.TorusGeometry(pR * 0.62, 0.016, 8, 24),
    mats.brass
  );
  trickleRing.rotation.x = Math.PI / 2;
  trickleRing.position.set(pX, trickleY, pZ);
  assemblyGroup.add(trickleRing);

  const trickleCone = new THREE.Mesh(
    new THREE.CylinderGeometry(pR * 0.65, pR * 0.84, 0.18, 20, 1, true),
    mats.sprayMist
  );
  trickleCone.position.set(pX, bedBaseY + bedH + 0.01, pZ);
  assemblyGroup.add(trickleCone);

  // Stacked Terracotta Chevron Mist-Eliminator Baffles
  for (let ci = 0; ci < 3; ci++) {
    const cy = bedBaseY + bedH + 0.21 + ci * 0.08;
    const chev = new THREE.Mesh(
      new THREE.ConeGeometry(pR * 0.76, 0.07, 18, 1, true),
      mats.terracotta
    );
    chev.position.set(pX, cy, pZ);
    assemblyGroup.add(chev);
  }

  const pTopCap = new THREE.Mesh(
    new THREE.CylinderGeometry(pR * 1.07, pR * 1.07, 0.075, 32),
    mats.mortar
  );
  pTopCap.position.set(pX, plinthY + pH + 0.037, pZ);
  assemblyGroup.add(pTopCap);

  // =========================================================================
  // 4. STAGE 4: SWG CERAMIC OVERHEAD DUCT (SMOOTH CURVED ELBOWS) + CHIMNEY
  // =========================================================================
  const chX = steamEnabled ? 1.32 : 1.18;
  const chZ = -0.05;
  const chBaseW = 0.68;
  const chH = steamEnabled
    ? clamp(1.15 + realChimM * 0.36, 2.45, 3.35)
    : clamp(2.45 + realChimM * 0.28, 3.85, 5.40);
  const swgR = 0.135;
  const ductTopY = plinthY + pH - 0.18;
  const ductBotY = 0.74;
  const elbowR = 0.22;
  const dropX = chX - 0.48;

  // Smooth curved SWG ceramic duct from tower top -> chimney base with bell-and-spigot collars
  const chDuctPath = new THREE.CurvePath<THREE.Vector3>();
  const dStart = new THREE.Vector3(pX + pR * 0.84, ductTopY, pZ);
  const dE1Start = new THREE.Vector3(dropX - elbowR, ductTopY, chZ);
  if (dE1Start.x > dStart.x + 0.01) {
    chDuctPath.add(new THREE.LineCurve3(dStart, dE1Start));
  }
  const dE1 = makeArc90Bezier(dE1Start, vRight, vDn, elbowR);
  chDuctPath.add(dE1);
  const dE1End = dE1.getPoint(1);
  const dE2Start = new THREE.Vector3(dropX, ductBotY + elbowR, chZ);
  chDuctPath.add(new THREE.LineCurve3(dE1End, dE2Start));
  const dE2 = makeArc90Bezier(dE2Start, vDn, vRight, elbowR);
  chDuctPath.add(dE2);
  const dE2End = dE2.getPoint(1);
  const dEnd = new THREE.Vector3(chX, ductBotY, chZ);
  chDuctPath.add(new THREE.LineCurve3(dE2End, dEnd));

  assemblyGroup.add(
    new THREE.Mesh(
      new THREE.TubeGeometry(chDuctPath, 52, swgR, 18, false),
      mats.terracotta
    )
  );
  addBellSpigotCollar(assemblyGroup, pX + pR * 0.96, ductTopY, pZ, "x", swgR, mats.terracotta, mats.mortar);
  addBellSpigotCollar(assemblyGroup, dropX, (dE1End.y + dE2Start.y) * 0.5, chZ, "y", swgR, mats.terracotta, mats.mortar);
  addBellSpigotCollar(assemblyGroup, chX - chBaseW * 0.5, ductBotY, chZ, "x", swgR, mats.terracotta, mats.mortar);

  // Brick plinth / thermal-reheat recuperator jacket at chimney base
  const reheatH = steamEnabled ? 0.92 : 1.38;
  const reheatOuter = new THREE.Mesh(
    new THREE.BoxGeometry(
      chBaseW + (steamEnabled ? 0.06 : 0.18),
      reheatH,
      chBaseW + (steamEnabled ? 0.06 : 0.18)
    ),
    mats.brickSolid
  );
  reheatOuter.position.set(chX, plinthY + reheatH * 0.5, chZ);
  assemblyGroup.add(reheatOuter);

  if (!steamEnabled) {
    const reheatBand = new THREE.Mesh(
      new THREE.BoxGeometry(chBaseW + 0.21, 0.24, chBaseW + 0.21),
      mats.hotCore
    );
    reheatBand.position.set(chX, plinthY + reheatH * 0.46, chZ);
    assemblyGroup.add(reheatBand);
  }

  // Tapered IS 651 Salt-Glazed Stoneware Chimney Shaft + Bell-and-Spigot Mortar Collars
  const chShaftH = Math.max(1.3, chH - reheatH);
  const chShaft = new THREE.Mesh(
    new THREE.CylinderGeometry(0.19, 0.25, chShaftH, 26),
    mats.terracotta
  );
  chShaft.position.set(chX, plinthY + reheatH + chShaftH * 0.5, chZ);
  assemblyGroup.add(chShaft);

  const collarN = steamEnabled ? 2 : 4;
  for (let j = 1; j <= collarN; j++) {
    const jy = plinthY + reheatH + (chShaftH * j) / (collarN + 1);
    const jr = 0.25 - (0.06 * j) / (collarN + 1);
    addBellSpigotCollar(assemblyGroup, chX, jy, chZ, "y", jr, mats.terracotta, mats.mortar);
  }
  const chExitY = plinthY + chH;
  const chLip = new THREE.Mesh(new THREE.TorusGeometry(0.20, 0.022, 8, 24), mats.mortar);
  chLip.rotation.x = Math.PI / 2;
  chLip.position.set(chX, chExitY, chZ);
  assemblyGroup.add(chLip);

  // =========================================================================
  // 5. ELEVATED KENTANK / SINTEX GRAVITY TANK (OMITTED IN ZERO-PUMP SUMP-POOL MODE!)
  // =========================================================================
  const tkX = steamEnabled ? 2.62 : 2.52;
  const tkZ = 0.02;
  const standH = clamp(tankElevM * 0.78, 2.25, 2.85);
  const stHalf = 0.44;
  const tkR = clamp(0.36 + (tankLiters - 200) * 0.00035, 0.36, 0.50);
  const tkH = clamp(0.86 + (tankLiters - 200) * 0.00045, 0.86, 1.06);
  const tkBaseY = plinthY + standH + 0.085;

  if (hasOverheadTank) {
    const postGeo = new THREE.BoxGeometry(0.088, standH, 0.088);
    for (const [dx, dz] of [
      [-stHalf, -stHalf],
      [stHalf, -stHalf],
      [-stHalf, stHalf],
      [stHalf, stHalf],
    ]) {
      const post = new THREE.Mesh(postGeo, mats.timber);
      post.position.set(tkX + dx, plinthY + standH * 0.5, tkZ + dz);
      assemblyGroup.add(post);
    }
    const deck = new THREE.Mesh(
      new THREE.BoxGeometry(stHalf * 2 + 0.24, 0.085, stHalf * 2 + 0.24),
      mats.timber
    );
    deck.position.set(tkX, plinthY + standH + 0.042, tkZ);
    assemblyGroup.add(deck);

    const addTimberBrace = (p1: THREE.Vector3, p2: THREE.Vector3) => {
      const dir = new THREE.Vector3().subVectors(p2, p1);
      const m = new THREE.Mesh(new THREE.BoxGeometry(0.045, dir.length(), 0.045), mats.timber);
      m.position.copy(p1).addScaledVector(dir, 0.5);
      m.quaternion.setFromUnitVectors(vUp, dir.normalize());
      assemblyGroup.add(m);
    };
    const yBot = plinthY + 0.06;
    const yMid = plinthY + standH * 0.5;
    const yTop = plinthY + standH - 0.05;
    const faces: Array<[[number, number], [number, number]]> = [
      [[-stHalf, -stHalf], [stHalf, -stHalf]],
      [[-stHalf, stHalf], [stHalf, stHalf]],
      [[-stHalf, -stHalf], [-stHalf, stHalf]],
      [[stHalf, -stHalf], [stHalf, stHalf]],
    ];
    for (const [a, bPt] of faces) {
      addTimberBrace(new THREE.Vector3(tkX + a[0], yBot, tkZ + a[1]), new THREE.Vector3(tkX + bPt[0], yMid, tkZ + bPt[1]));
      addTimberBrace(new THREE.Vector3(tkX + bPt[0], yBot, tkZ + bPt[1]), new THREE.Vector3(tkX + a[0], yMid, tkZ + a[1]));
      addTimberBrace(new THREE.Vector3(tkX + a[0], yMid, tkZ + a[1]), new THREE.Vector3(tkX + bPt[0], yTop, tkZ + bPt[1]));
      addTimberBrace(new THREE.Vector3(tkX + bPt[0], yMid, tkZ + bPt[1]), new THREE.Vector3(tkX + a[0], yTop, tkZ + a[1]));
    }

    const tankMesh = new THREE.Mesh(
      new THREE.CylinderGeometry(tkR, tkR, tkH, 28),
      mats.kentank
    );
    tankMesh.position.set(tkX, tkBaseY + tkH * 0.5, tkZ);
    assemblyGroup.add(tankMesh);
    for (const rf of [0.22, 0.48, 0.74]) {
      const rib = new THREE.Mesh(
        new THREE.TorusGeometry(tkR + 0.012, 0.018, 8, 28),
        mats.kentankRib
      );
      rib.rotation.x = Math.PI / 2;
      rib.position.set(tkX, tkBaseY + tkH * rf, tkZ);
      assemblyGroup.add(rib);
    }
    const tkLid = new THREE.Mesh(
      new THREE.CylinderGeometry(0.18, 0.19, 0.07, 18),
      mats.iron
    );
    tkLid.position.set(tkX, tkBaseY + tkH + 0.035, tkZ);
    assemblyGroup.add(tkLid);
  }

  // =========================================================================
  // 6. BRICK BAFFLED SUMP + PUMP (TREADLE / PPCP / SYPHON / ZERO-PUMP CARBURETOR DIP-TUBE)
  // =========================================================================
  const sumpX = -0.96;
  const sumpZ = 0.94;
  const sumpW = isSumpPoolOnly ? 1.65 : 1.22;
  const sumpOuter = new THREE.Mesh(
    new THREE.BoxGeometry(sumpW, 0.46, 0.62),
    mats.brickSolid
  );
  sumpOuter.position.set(sumpX, plinthY + 0.23, sumpZ);
  assemblyGroup.add(sumpOuter);
  const sumpWater = new THREE.Mesh(
    new THREE.BoxGeometry(sumpW - 0.16, 0.32, 0.48),
    mats.sumpWater
  );
  sumpWater.position.set(sumpX, plinthY + 0.29, sumpZ);
  assemblyGroup.add(sumpWater);
  const sumpWeir = new THREE.Mesh(
    new THREE.BoxGeometry(0.06, 0.28, 0.48),
    mats.mortar
  );
  sumpWeir.position.set(sumpX, plinthY + 0.25, sumpZ);
  assemblyGroup.add(sumpWeir);

  const addCPipe = (pts: THREE.Vector3[], rad: number, mat: THREE.Material) => {
    assemblyGroup.add(
      new THREE.Mesh(
        new THREE.TubeGeometry(new THREE.CatmullRomCurve3(pts), 30, rad, 10, false),
        mat
      )
    );
  };

  // Gravity drains from Quench & Stoneware Tower -> Baffled Brick Sump
  addCPipe(
    [
      new THREE.Vector3(qX + 0.10, plinthY + 0.12, qZ + qD * 0.48),
      new THREE.Vector3(sumpX - 0.38, plinthY + 0.22, sumpZ - 0.28),
    ],
    0.026,
    mats.hoseReturn
  );
  addCPipe(
    [
      new THREE.Vector3(pX - 0.12, plinthY + 0.14, pZ + pR * 0.92),
      new THREE.Vector3(sumpX + 0.42, plinthY + 0.24, sumpZ - 0.28),
    ],
    0.026,
    mats.hoseReturn
  );

  const pumpX = 0.36;
  const pumpZ = 0.94;
  if (isSumpPoolOnly) {
    // ZERO PUMP & ZERO WATER TOWER: 1/2" Brass/316L Carburetor Aspiration Dip-Tube rising from Sump Pool #2 into Steam-Jet Ejector Throat!
    addCPipe(
      [
        new THREE.Vector3(sumpX + 0.22, plinthY + 0.14, sumpZ),
        new THREE.Vector3(sumpX + 0.22, plinthY + 0.56, sumpZ - 0.22),
        new THREE.Vector3(ejMidX, ejY - 0.16, qZ + 0.22),
        new THREE.Vector3(ejMidX, ejY - 0.06, qZ),
      ],
      0.018,
      mats.brass
    );
    addBallValve(assemblyGroup, ejMidX, ejY - 0.18, qZ + 0.22, "y", 0.020, mats.brass, mats.valveGreen);
  } else {
    if (isSteamPPCP) {
      // Forbes Marshall / Spirax Sarco Mechanical Pressure-Powered Condensate Pump (PPCP, 0 W!)
      const ppcpBody = new THREE.Mesh(
        new THREE.CylinderGeometry(0.17, 0.17, 0.46, 24),
        mats.iron
      );
      ppcpBody.position.set(pumpX, plinthY + 0.26, pumpZ);
      assemblyGroup.add(ppcpBody);
      const ppcpCap = new THREE.Mesh(
        new THREE.CylinderGeometry(0.19, 0.19, 0.05, 24),
        mats.steamerSS
      );
      ppcpCap.position.set(pumpX, plinthY + 0.50, pumpZ);
      assemblyGroup.add(ppcpCap);
      // Copper motive steam supply tap from header S1 -> PPCP top head
      addCPipe(
        [
          new THREE.Vector3(ejStart + 0.08, ejY + 0.44, qZ),
          new THREE.Vector3(pumpX, ejY + 0.44, pumpZ - 0.18),
          new THREE.Vector3(pumpX, plinthY + 0.52, pumpZ),
        ],
        0.015,
        mats.steamHeader
      );
    } else if (isSteamSyphon) {
      // Gunmetal/Brass Steam-Jet Cold Makeup Syphon + Bronze Boiler Feed Injector
      const syphBody = new THREE.Mesh(
        new THREE.CylinderGeometry(0.055, 0.075, 0.28, 16),
        mats.brass
      );
      syphBody.rotation.z = Math.PI / 2;
      syphBody.position.set(pumpX, plinthY + 0.28, pumpZ);
      assemblyGroup.add(syphBody);
      // Steam motive branch to syphon
      addCPipe(
        [
          new THREE.Vector3(ejStart + 0.08, ejY + 0.36, qZ),
          new THREE.Vector3(pumpX - 0.12, plinthY + 0.36, pumpZ),
        ],
        0.014,
        mats.steamHeader
      );
    } else if (isZeroElec) {
      // MoneyMaker Twin-Cylinder Manual Treadle Pump (0 W)
      const pBase = new THREE.Mesh(new THREE.BoxGeometry(0.48, 0.08, 0.38), mats.timber);
      pBase.position.set(pumpX, plinthY + 0.04, pumpZ);
      assemblyGroup.add(pBase);
      for (const dx of [-0.09, 0.09]) {
        const cyl = new THREE.Mesh(
          new THREE.CylinderGeometry(0.055, 0.055, 0.38, 16),
          mats.iron
        );
        cyl.position.set(pumpX + dx, plinthY + 0.27, pumpZ - 0.04);
        assemblyGroup.add(cyl);
        const treadle = new THREE.Mesh(new THREE.BoxGeometry(0.10, 0.04, 0.52), mats.timber);
        treadle.position.set(pumpX + dx, plinthY + 0.14 + (dx > 0 ? 0.08 : 0), pumpZ + 0.14);
        treadle.rotation.x = dx > 0 ? -0.18 : 0.12;
        assemblyGroup.add(treadle);
      }
      const handleBar = new THREE.Mesh(
        new THREE.CylinderGeometry(0.018, 0.018, 0.72, 10),
        mats.iron
      );
      handleBar.position.set(pumpX + 0.22, plinthY + 0.42, pumpZ - 0.06);
      assemblyGroup.add(handleBar);
    } else {
      // 0.5 HP 230V 50Hz Domestic Monoblock Pump (Kirloskar / Davis & Shirtliff)
      const monoblock = new THREE.Mesh(
        new THREE.CylinderGeometry(0.085, 0.085, 0.28, 18),
        mats.kentank
      );
      monoblock.rotation.z = Math.PI / 2;
      monoblock.position.set(pumpX, plinthY + 0.12, pumpZ);
      assemblyGroup.add(monoblock);
    }

    // Gravity downcomer piping + brass shutoff ball valves + pump refill riser
    const headerY = tkBaseY - 0.28;
    addCPipe(
      [
        new THREE.Vector3(tkX - 0.16, tkBaseY, tkZ),
        new THREE.Vector3(tkX - 0.16, headerY, tkZ + 0.36),
        new THREE.Vector3(pX, headerY, pZ + 0.36),
        new THREE.Vector3(qX - 0.08, plinthY + qH + 0.24, qZ),
        new THREE.Vector3(qX - 0.08, plinthY + qH - 0.06, qZ),
      ],
      0.022,
      mats.pipeCyan
    );
    addBallValve(assemblyGroup, tkX - 0.16, tkBaseY - 0.14, tkZ + 0.12, "y", 0.024, mats.brass, mats.safetyRed);
    addCPipe(
      [
        new THREE.Vector3(pX, headerY, pZ + 0.36),
        new THREE.Vector3(pX, trickleY + 0.10, pZ),
      ],
      0.020,
      mats.pipeCyan
    );
    addBallValve(assemblyGroup, pX, headerY - 0.16, pZ + 0.22, "y", 0.021, mats.brass, mats.valveGreen);

    // Sump suction -> Pump -> Refill riser to top of Kentank
    addCPipe(
      [
        new THREE.Vector3(sumpX + 0.58, plinthY + 0.18, sumpZ),
        new THREE.Vector3(pumpX - 0.10, plinthY + 0.18, pumpZ),
      ],
      0.024,
      mats.pipeCyan
    );
    addCPipe(
      [
        new THREE.Vector3(pumpX + 0.10, plinthY + 0.28, pumpZ),
        new THREE.Vector3(tkX - 0.36, plinthY + 0.28, sumpZ),
        new THREE.Vector3(tkX - 0.36, tkBaseY + tkH + 0.08, tkZ + 0.18),
        new THREE.Vector3(tkX - 0.10, tkBaseY + tkH - 0.02, tkZ),
      ],
      0.020,
      mats.pipeCyan
    );
  }

  // =========================================================================
  // 7. 3D ENGINEERING DIMENSION CALLOUTS (TOGGLEABLE)
  // =========================================================================
  if (showDimensions) {
    addDimensionCallout3D(
      assemblyGroup,
      new THREE.Vector3(pX - pR, plinthY + pH + 0.10, pZ),
      new THREE.Vector3(pX + pR, plinthY + pH + 0.10, pZ),
      new THREE.Vector3(0, 0.26, 0),
      `Bed Ø ${fmtDimMeters(bedDiamM, units)}`,
      "#34d399",
      disposableTextures
    );
    addDimensionCallout3D(
      assemblyGroup,
      new THREE.Vector3(chX + 0.28, plinthY, chZ),
      new THREE.Vector3(chX + 0.28, chExitY, chZ),
      new THREE.Vector3(0.28, 0, 0),
      `Chimney ${fmtDimMeters(realChimM, units)}`,
      "#fb923c",
      disposableTextures
    );
    if (hasOverheadTank) {
      addDimensionCallout3D(
        assemblyGroup,
        new THREE.Vector3(tkX + stHalf + 0.06, plinthY, tkZ),
        new THREE.Vector3(tkX + stHalf + 0.06, tkBaseY, tkZ),
        new THREE.Vector3(0.32, 0, 0),
        `Gravity ${fmtDimMeters(tankElevM, units)} (${Math.round(tankLiters)} L)`,
        "#38bdf8",
        disposableTextures
      );
    } else {
      addDimensionCallout3D(
        assemblyGroup,
        new THREE.Vector3(sumpX - sumpW * 0.5, plinthY + 0.46, sumpZ),
        new THREE.Vector3(sumpX + sumpW * 0.5, plinthY + 0.46, sumpZ),
        new THREE.Vector3(0, 0.24, 0.34),
        `Zero-Pump Sump Pools (0 Tower)`,
        "#38bdf8",
        disposableTextures
      );
    }
    if (isSubsonic && standpipeBasePt && standpipeTopPt) {
      addDimensionCallout3D(
        assemblyGroup,
        standpipeBasePt,
        standpipeTopPt,
        new THREE.Vector3(-0.34, 0, 0),
        `S2 Relief 4.2 m (6 psig cap)`,
        "#38bdf8",
        disposableTextures
      );
    }
  }

  // =========================================================================
  // 8. CENTER ASSEMBLY & COMPUTE DYNAMIC BOUNDING-BOX CAMERA VIEWS
  // =========================================================================
  const rawBox = new THREE.Box3().setFromObject(assemblyGroup);
  const centerX = -(rawBox.min.x + rawBox.max.x) * 0.5;
  const centerZ = -(rawBox.min.z + rawBox.max.z) * 0.5;
  assemblyGroup.position.set(centerX, 0, centerZ);

  const totalW = rawBox.max.x - rawBox.min.x;
  const totalD = rawBox.max.z - rawBox.min.z;
  const totalH = Math.max(rawBox.max.y, chExitY) + 0.24;
  const tanHalfFov = Math.tan(THREE.MathUtils.degToRad(cameraFovDeg * 0.5));
  const allTargetY = totalH * 0.45;
  const allDist = Math.max(
    6.4,
    Math.max(
      (totalH * 0.58) / tanHalfFov,
      (totalW * 0.54) / (tanHalfFov * Math.max(1.4, cameraAspect))
    ) *
      1.18 +
      totalD * 0.38
  );

  const stageViews: Record<StageKey, StageView> = {
    overview: {
      target: new THREE.Vector3(0.10, allTargetY, 0),
      cam: new THREE.Vector3(0.85, allTargetY + totalH * 0.16, allDist),
    },
    inlet: {
      target: new THREE.Vector3(qX - qW * 0.5 - 0.18 + centerX, qInY, qZ + centerZ),
      cam: new THREE.Vector3(qX - qW * 0.5 + 0.25 + centerX, qInY + 0.45, qZ + centerZ + 2.4),
    },
    quench: {
      target: new THREE.Vector3(qX + centerX + 0.08, 0.92, qZ + centerZ),
      cam: new THREE.Vector3(qX + centerX + 0.45, 1.32, qZ + centerZ + 2.85),
    },
    venturi: steamEnabled
      ? {
          target: new THREE.Vector3(ejMidX + centerX - 0.18, 0.86, qZ + centerZ - 0.22),
          cam: new THREE.Vector3(ejMidX + centerX + 0.55, 1.42, qZ + centerZ + 2.95),
        }
      : {
          target: new THREE.Vector3(ejMidX + centerX, 0.72, qZ + centerZ),
          cam: new THREE.Vector3(ejMidX + centerX + 0.45, 1.15, qZ + centerZ + 2.6),
        },
    packed: {
      target: new THREE.Vector3(pX + centerX + 0.10, 1.38, pZ + centerZ),
      cam: new THREE.Vector3(pX + centerX + 0.55, 1.72, pZ + centerZ + 3.45),
    },
    blower: {
      target: new THREE.Vector3(chX + centerX + 0.22, totalH * 0.48, chZ + centerZ),
      cam: new THREE.Vector3(
        chX + centerX + 1.05,
        totalH * 0.56,
        chZ + centerZ + Math.max(4.6, totalH * 1.12)
      ),
    },
  };

  // =========================================================================
  // 9. EXACT ENCLOSED-DUCT FLOW PARTICLE PATHS (ZERO WALL CLIPPING!)
  // =========================================================================
  const ox = centerX;
  const oz = centerZ;
  const primaryFlow: FlowSample[] = [];

  // Hot kiln inlet flue -> quench turn -> downward through spray -> ejector/baffle throat
  primaryFlow.push(
    ...samplePathWithMeta(
      new THREE.LineCurve3(
        new THREE.Vector3(qX - qW * 0.5 - flueLen, qInY, qZ),
        new THREE.Vector3(qX - 0.12, qInY, qZ)
      ),
      18,
      0.032,
      1.25,
      "inlet",
      ox,
      oz
    ),
    ...samplePathWithMeta(
      new THREE.QuadraticBezierCurve3(
        new THREE.Vector3(qX - 0.12, qInY, qZ),
        new THREE.Vector3(qX - 0.04, qInY, qZ),
        new THREE.Vector3(qX - 0.04, qInY - 0.16, qZ)
      ),
      8,
      0.045,
      0.95,
      "quench",
      ox,
      oz
    ),
    ...samplePathWithMeta(
      new THREE.LineCurve3(
        new THREE.Vector3(qX - 0.04, qInY - 0.16, qZ),
        new THREE.Vector3(qX - 0.04, ejY + 0.12, qZ)
      ),
      20,
      0.09,
      0.72,
      "quench",
      ox,
      oz
    ),
    ...samplePathWithMeta(
      new THREE.QuadraticBezierCurve3(
        new THREE.Vector3(qX - 0.04, ejY + 0.12, qZ),
        new THREE.Vector3(qX - 0.04, ejY, qZ),
        new THREE.Vector3(ejStart, ejY, qZ)
      ),
      8,
      0.04,
      1.0,
      "quench",
      ox,
      oz
    ),
    // Stage 2: Accelerated jet through steam ejector throat (or gentle pass through masonry flue)
    ...samplePathWithMeta(
      new THREE.LineCurve3(
        new THREE.Vector3(ejStart, ejY, qZ),
        new THREE.Vector3(pX - 0.12, ejY, pZ)
      ),
      24,
      steamEnabled ? 0.038 : 0.055,
      steamEnabled ? 2.6 : 1.05,
      "venturi",
      ox,
      oz
    ),
    ...samplePathWithMeta(
      new THREE.QuadraticBezierCurve3(
        new THREE.Vector3(pX - 0.12, ejY, pZ),
        new THREE.Vector3(pX, ejY, pZ),
        new THREE.Vector3(pX, ejY + 0.18, pZ)
      ),
      8,
      0.06,
      0.85,
      "packed",
      ox,
      oz
    ),
    // Stage 3: Slow upward percolation through wide stoneware Raschig ring bed
    ...samplePathWithMeta(
      new THREE.LineCurve3(
        new THREE.Vector3(pX, ejY + 0.18, pZ),
        new THREE.Vector3(pX, ductTopY - 0.12, pZ)
      ),
      44,
      pR * 0.56,
      0.55,
      "packed",
      ox,
      oz
    ),
    ...samplePathWithMeta(
      new THREE.QuadraticBezierCurve3(
        new THREE.Vector3(pX, ductTopY - 0.12, pZ),
        new THREE.Vector3(pX, ductTopY, pZ),
        new THREE.Vector3(pX + pR * 0.84, ductTopY, pZ)
      ),
      8,
      0.045,
      0.95,
      "packed",
      ox,
      oz
    ),
    // Exact curved SWG ceramic overhead duct centerline into chimney base
    ...samplePathWithMeta(chDuctPath, 46, swgR * 0.42, 1.15, "blower", ox, oz),
    // Up inside chimney shaft -> rising clean plume above lip
    ...samplePathWithMeta(
      new THREE.LineCurve3(
        new THREE.Vector3(chX, ductBotY, chZ),
        new THREE.Vector3(chX, chExitY, chZ)
      ),
      40,
      0.075,
      1.35,
      "blower",
      ox,
      oz
    ),
    ...samplePathWithMeta(
      new THREE.LineCurve3(
        new THREE.Vector3(chX, chExitY, chZ),
        new THREE.Vector3(chX, chExitY + 0.72, chZ)
      ),
      18,
      0.14,
      1.15,
      "plume",
      ox,
      oz
    )
  );

  // Motive Steam-Jet Stream through insulated copper header S1 -> brass orifice -> ejector throat -> tower
  const motiveFlow: FlowSample[] = [];
  if (steamEnabled && steamHdrCurve) {
    motiveFlow.push(
      ...samplePathWithMeta(steamHdrCurve, 34, 0.010, 1.65, "steam", ox, oz),
      ...samplePathWithMeta(
        new THREE.LineCurve3(
          new THREE.Vector3(nozzleTipX, ejY, qZ),
          new THREE.Vector3(pX, ejY, pZ)
        ),
        22,
        0.045,
        3.2,
        "steam",
        ox,
        oz
      ),
      ...samplePathWithMeta(
        new THREE.LineCurve3(
          new THREE.Vector3(pX, ejY, pZ),
          new THREE.Vector3(pX, bedBaseY + bedH * 0.75, pZ)
        ),
        20,
        pR * 0.42,
        0.95,
        "steam",
        ox,
        oz
      )
    );
  }

  // Falling gravity water spray droplets (or self-induced sump curtain mist in zero-pump sump-pool mode!)
  const sprayZones: SprayZone[] = hasOverheadTank
    ? [
        {
          topCenter: new THREE.Vector3(qX - 0.08 + ox, plinthY + qH - 0.18, qZ + oz),
          topRadius: 0.05,
          botRadius: qW * 0.30,
          height: qH * 0.68,
          color: 0x38bdf8,
          count: 34,
          speed: 1.15,
        },
        {
          topCenter: new THREE.Vector3(pX + ox, trickleY - 0.02, pZ + oz),
          topRadius: pR * 0.55,
          botRadius: pR * 0.78,
          height: bedH + 0.14,
          color: 0x22d3ee,
          count: 42,
          speed: 0.85,
        },
      ]
    : [
        {
          topCenter: new THREE.Vector3(qX + 0.08 + ox, plinthY + 0.78, qZ + oz),
          topRadius: qW * 0.24,
          botRadius: qW * 0.34,
          height: 0.44,
          color: 0x38bdf8,
          count: 36,
          speed: 1.35,
        },
        {
          topCenter: new THREE.Vector3(pX + ox, bedBaseY + bedH * 0.78, pZ + oz),
          topRadius: pR * 0.48,
          botRadius: pR * 0.72,
          height: bedH * 0.78,
          color: 0x22d3ee,
          count: 32,
          speed: 0.75,
        },
      ];

  const archSig = `local|${uiDraft}|${boilerSource}|${pumpMode}|${isZeroElec ? "0w" : "grid"}|${Math.round(bedDiamIn)}`;
  const scaleSummary = isSumpPoolOnly
    ? `Scale 1:1 m · Bed ${fmtDimMeters(bedDiamM, units)} · Chimney ${fmtDimMeters(realChimM, units)} · 0-Pump Sump Pools`
    : `Scale 1:1 m · Bed ${fmtDimMeters(bedDiamM, units)} · Chimney ${fmtDimMeters(realChimM, units)}`;

  return {
    stageViews,
    primaryFlow,
    motiveFlow,
    diluterFlow: [],
    sprayZones,
    impellerMesh: null,
    impellerDir: 1,
    diluterImpellerMesh: null,
    architectureSignature: archSig,
    scaleSummary,
    disposableTextures,
  };
}
