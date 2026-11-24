"use client";

import React, { useEffect, useRef, useState } from "react";
import * as THREE from "three";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls.js";
import type { StudioEvaluation } from "@/types/kilnscrub";

interface Viewport3DProps {
  evaluation: StudioEvaluation;
  cutaway: boolean;
  particles: boolean;
  showSystemsOverlay: boolean;
  focusStage: string | null;
  active: boolean;
  onWebGLUnavailable?: () => void;
}

const CAMERA_PRESETS: Record<
  string,
  { cam: [number, number, number]; target: [number, number, number] }
> = {
  overview: { cam: [1.5, 2.7, 6.4], target: [0.1, 1.35, 0] },
  inlet: { cam: [-2.6, 2.0, 2.8], target: [-1.85, 1.4, 0] },
  quench: { cam: [-1.2, 2.1, 3.0], target: [-0.95, 1.45, 0] },
  venturi: { cam: [0.0, 1.6, 2.6], target: [-0.05, 0.95, 0] },
  packed: { cam: [1.1, 2.2, 3.4], target: [0.95, 1.6, 0] },
  blower: { cam: [2.7, 2.0, 3.2], target: [2.05, 1.45, 0] },
};

export function Viewport3D({
  evaluation,
  cutaway,
  particles,
  showSystemsOverlay,
  focusStage,
  active,
  onWebGLUnavailable,
}: Viewport3DProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const activeRef = useRef(active);
  const particlesRef = useRef(particles);
  const camGoalRef = useRef<{ cam: THREE.Vector3; target: THREE.Vector3 } | null>(
    null
  );
  const [webglError, setWebglError] = useState<string | null>(null);

  activeRef.current = active;
  particlesRef.current = particles;

  useEffect(() => {
    const preset = CAMERA_PRESETS[focusStage || "overview"] || CAMERA_PRESETS.overview;
    camGoalRef.current = {
      cam: new THREE.Vector3(...preset.cam),
      target: new THREE.Vector3(...preset.target),
    };
  }, [focusStage]);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    let renderer: THREE.WebGLRenderer;
    try {
      renderer = new THREE.WebGLRenderer({
        antialias: true,
        powerPreference: "high-performance",
      });
    } catch {
      setWebglError("WebGL unavailable — switched to 2D Parametric Cutaway");
      onWebGLUnavailable?.();
      return;
    }

    renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2));
    renderer.toneMapping = THREE.ACESFilmicToneMapping;
    renderer.toneMappingExposure = 1.1;
    container.innerHTML = "";
    container.appendChild(renderer.domElement);

    const scene = new THREE.Scene();
    scene.background = new THREE.Color(0x0b1420);

    const camera = new THREE.PerspectiveCamera(40, 16 / 9, 0.05, 60);
    camera.position.set(1.5, 2.7, 6.4);

    const controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.08;
    controls.target.set(0.1, 1.35, 0);
    controls.minDistance = 1.2;
    controls.maxDistance = 16;

    // Studio Lighting + Floor Grid
    scene.add(new THREE.HemisphereLight(0xdbeafe, 0x1e293b, 0.95));
    const dirLight = new THREE.DirectionalLight(0xffffff, 1.25);
    dirLight.position.set(5, 8, 6);
    scene.add(dirLight);
    const rimLight = new THREE.DirectionalLight(0x38bdf8, 0.45);
    rimLight.position.set(-5, 4, -4);
    scene.add(rimLight);

    const grid = new THREE.GridHelper(12, 24, 0x264466, 0x15263a);
    grid.position.y = 0;
    scene.add(grid);

    // Build parametric scrubber geometry directly from StudioEvaluation
    const root = new THREE.Group();
    scene.add(root);

    const { project, design, frugal, architecture_options: arch } = evaluation;
    const isLocal = project.build_path === "local_ceramic_masonry";
    const isDeluxe = project.build_path === "custom_fab_deluxe";
    const isSteam = isLocal && !!frugal?.steam_enabled;
    const isSubsonic = isSteam && frugal?.draft_mode === "steam_ejector_subsonic";
    const isEductor = !isLocal && project.stage_alignment === "eductor_dry_fan";
    const isParallel = !isLocal && project.packed_bed_layout === "parallel_manifold";
    const isCompact = !isLocal && project.compact_footprint;

    const shellOpacity = cutaway ? 0.46 : 1.0;
    const matHotMetal = new THREE.MeshStandardMaterial({
      color: isLocal ? 0xb45309 : 0xcbd5e1,
      metalness: isLocal ? 0.1 : 0.72,
      roughness: isLocal ? 0.85 : 0.3,
      transparent: cutaway,
      opacity: shellOpacity,
    });
    const matColumn = new THREE.MeshStandardMaterial({
      color: isLocal ? 0x9a3412 : isDeluxe ? 0x0f766e : 0x334155,
      metalness: isLocal ? 0.08 : 0.2,
      roughness: isLocal ? 0.85 : 0.45,
      transparent: cutaway,
      opacity: shellOpacity,
    });
    const matPacking = new THREE.MeshStandardMaterial({
      color: isLocal ? 0xd97706 : 0x10b981,
      roughness: 0.7,
    });
    const matBlower = new THREE.MeshStandardMaterial({
      color: 0x0284c7,
      metalness: 0.35,
      roughness: 0.4,
    });
    const matSkid = new THREE.MeshStandardMaterial({
      color: 0x334155,
      metalness: 0.6,
      roughness: 0.5,
    });
    const matBypass = new THREE.MeshStandardMaterial({
      color: 0xef4444,
      metalness: 0.3,
      roughness: 0.4,
    });

    const addCyl = (
      rTop: number,
      rBot: number,
      h: number,
      mat: THREE.Material,
      x: number,
      y: number,
      z: number,
      rotZ = 0
    ) => {
      const mesh = new THREE.Mesh(
        new THREE.CylinderGeometry(rTop, rBot, h, 28),
        mat
      );
      mesh.position.set(x, y, z);
      mesh.rotation.z = rotZ;
      root.add(mesh);
      return mesh;
    };

    const qR = Math.max(0.18, Math.min(0.42, (design.quench.tower_diameter_m || 0.35) * 0.65));
    const qH = Math.max(1.2, Math.min(2.4, (design.quench.tower_height_m || 2.0) * 0.75));
    const colR = isLocal
      ? Math.max(0.32, Math.min(0.48, ((frugal?.bed_diameter_in || 24) * 0.0254) * 0.55))
      : Math.max(0.18, Math.min(0.45, (design.packed_bed.column_diameter_m || 0.32) * 0.75));
    const colH = Math.max(1.5, Math.min(2.7, (design.packed_bed.column_height_m || 2.6) * 0.72));

    const xQuench = -1.0;
    const xCol = isCompact ? -0.2 : 0.95;
    const zCol = isCompact ? -1.15 : 0;
    const xBlower = isCompact ? 1.15 : 2.1;
    const zBlower = isCompact ? -0.9 : 0;

    // Skid / Masonry Pad
    const pad = new THREE.Mesh(
      new THREE.BoxGeometry(isCompact ? 3.8 : 5.2, 0.12, isCompact ? 2.6 : 1.5),
      matSkid
    );
    pad.position.set(0.35, 0.06, isCompact ? -0.55 : 0);
    root.add(pad);

    // Inlet Flue Duct + Fail-Safe Bypass Tee
    addCyl(0.11, 0.11, 1.0, matHotMetal, xQuench - 0.65, qH - 0.15, 0, Math.PI / 2);
    if (!isLocal) {
      addCyl(0.1, 0.1, 0.65, matBypass, xQuench - 0.55, qH + 0.2, 0, 0);
    }

    // Stage 1: Quench Tower
    addCyl(qR, qR, qH, matHotMetal, xQuench, 0.12 + qH / 2, 0);
    // Quench spray cone inside
    addCyl(0.03, qR * 0.8, qH * 0.4, new THREE.MeshBasicMaterial({ color: 0x38bdf8, transparent: true, opacity: 0.35 }), xQuench, 0.12 + qH * 0.65, 0);

    // Stage 2: Venturi / Eductor / Steam-Jet
    const xMid = (xQuench + xCol) / 2;
    const zMid = zCol / 2;
    if (isSteam) {
      // Boiler / Catering Steamer + Steam Header S1 + optional 4.2m Standpipe S2
      const boiler = new THREE.Mesh(
        new THREE.CylinderGeometry(0.24, 0.24, 0.68, 24),
        new THREE.MeshStandardMaterial({ color: 0xeab308, metalness: 0.6, roughness: 0.3 })
      );
      boiler.position.set(xMid, 0.46, 0.85);
      root.add(boiler);
      addCyl(0.04, 0.04, 0.85, new THREE.MeshStandardMaterial({ color: 0xfde047 }), xMid, 0.72, 0.42, Math.PI / 2);
      if (isSubsonic) {
        // 4.2m Open Hydrostatic Standpipe S2
        addCyl(0.035, 0.035, 2.7, new THREE.MeshStandardMaterial({ color: 0x38bdf8 }), xMid - 0.32, 1.48, 0.85);
      }
      addCyl(0.11, 0.11, 1.0, matHotMetal, xMid, 0.65, zMid, Math.PI / 2);
    } else if (isEductor) {
      // Vertical downflow static eductor + dry motive fan on top
      addCyl(0.16, 0.07, 0.45, matHotMetal, xMid, 1.55, zMid);
      addCyl(0.07, 0.16, 0.65, matHotMetal, xMid, 1.0, zMid);
      addCyl(0.22, 0.22, 0.25, matBlower, xMid, 2.0, zMid);
    } else {
      // Horizontal converging-throat-diverging venturi
      addCyl(0.13, 0.06, 0.32, matHotMetal, xMid - 0.28, 0.65, zMid, Math.PI / 2);
      addCyl(0.06, 0.06, 0.18, matHotMetal, xMid, 0.65, zMid, Math.PI / 2);
      addCyl(0.06, 0.14, 0.45, matHotMetal, xMid + 0.32, 0.65, zMid, Math.PI / 2);
    }

    // Stage 3: Packed Column or Shell-Less Parallel Uniseal Bundle
    if (isParallel) {
      const count = Math.min(9, Math.max(4, arch.parallel_bed.pipe_count || 7));
      const cols = Math.ceil(Math.sqrt(count));
      const spacing = 0.17;
      for (let i = 0; i < count; i++) {
        const r = Math.floor(i / cols);
        const c = i % cols;
        const px = xCol + (c - (cols - 1) / 2) * spacing;
        const pz = zCol + (r - (cols - 1) / 2) * spacing;
        addCyl(0.065, 0.065, colH * 0.85, matColumn, px, 0.35 + (colH * 0.85) / 2, pz);
      }
      // Top & bottom collector plates
      const plateGeo = new THREE.BoxGeometry(cols * spacing + 0.18, 0.06, cols * spacing + 0.18);
      const pBot = new THREE.Mesh(plateGeo, matSkid);
      pBot.position.set(xCol, 0.35, zCol);
      const pTop = new THREE.Mesh(plateGeo, matSkid);
      pTop.position.set(xCol, 0.35 + colH * 0.85, zCol);
      root.add(pBot, pTop);
    } else {
      addCyl(colR, colR, colH, matColumn, xCol, 0.12 + colH / 2, zCol);
      // Visible internal packing bed + chevron demister
      addCyl(colR * 0.84, colR * 0.84, colH * 0.45, matPacking, xCol, 0.12 + colH * 0.46, zCol);
      addCyl(colR * 0.86, colR * 0.86, 0.18, new THREE.MeshStandardMaterial({ color: 0x38bdf8 }), xCol, 0.12 + colH * 0.82, zCol);
    }

    // Sump Tank (Atmospheric)
    const sump = new THREE.Mesh(
      new THREE.CylinderGeometry(0.32, 0.32, 0.62, 24),
      new THREE.MeshStandardMaterial({ color: 0x1d4ed8, roughness: 0.5 })
    );
    sump.position.set(xCol, 0.43, zCol + 0.78);
    root.add(sump);

    // Stage 4: Blower + Exhaust Stack OR Tall Local Chimney + 3.5m Gravity Tank
    let impellerGroup: THREE.Group | null = null;
    if (isLocal) {
      const chH = Math.min(3.8, Math.max(2.2, (frugal?.chimney_height_m || 8.5) * 0.36));
      addCyl(0.18, 0.22, chH, new THREE.MeshStandardMaterial({ color: 0x9a3412, roughness: 0.9 }), xBlower, 0.12 + chH / 2, zBlower);
      // Overhead 3.5m Braced Gravity Tank
      const tankStandH = 2.05;
      for (const dx of [-0.32, 0.32]) {
        for (const dz of [-0.32, 0.32]) {
          addCyl(0.03, 0.03, tankStandH, matSkid, xCol + 0.95 + dx, 0.12 + tankStandH / 2, zCol + 0.95 + dz);
        }
      }
      addCyl(0.36, 0.36, 0.65, new THREE.MeshStandardMaterial({ color: 0x0284c7, roughness: 0.4 }), xCol + 0.95, 0.12 + tankStandH + 0.33, zCol + 0.95);
    } else {
      // Blower Volute + Spinning Impeller + Exhaust Stack
      const scroll = new THREE.Mesh(
        new THREE.CylinderGeometry(0.34, 0.34, 0.28, 28),
        matBlower
      );
      scroll.rotation.x = Math.PI / 2;
      scroll.position.set(xBlower, 0.72, zBlower);
      root.add(scroll);

      impellerGroup = new THREE.Group();
      impellerGroup.position.set(xBlower, 0.72, zBlower + 0.16);
      const bladeMat = new THREE.MeshStandardMaterial({ color: 0xe2e8f0, metalness: 0.8, roughness: 0.2 });
      for (let i = 0; i < 6; i++) {
        const blade = new THREE.Mesh(new THREE.BoxGeometry(0.46, 0.045, 0.03), bladeMat);
        blade.rotation.z = (i * Math.PI) / 6;
        impellerGroup.add(blade);
      }
      root.add(impellerGroup);

      // Vertical discharge stack
      addCyl(0.13, 0.13, 2.35, new THREE.MeshStandardMaterial({ color: 0x64748b }), xBlower + 0.18, 1.9, zBlower);
    }

    // Optional Deluxe / Systems 3D Overlay (Acoustic Baffle Box + UL 508A Panel + K-Bracing)
    if ((showSystemsOverlay || isDeluxe) && !isLocal) {
      const baffleBox = new THREE.Mesh(
        new THREE.BoxGeometry(0.95, 1.05, 0.9),
        new THREE.MeshStandardMaterial({
          color: 0xa855f7,
          transparent: true,
          opacity: 0.25,
          roughness: 0.6,
        })
      );
      baffleBox.position.set(xBlower, 0.68, zBlower);
      root.add(baffleBox);

      const panel = new THREE.Mesh(
        new THREE.BoxGeometry(0.48, 0.72, 0.24),
        new THREE.MeshStandardMaterial({ color: 0x475569, metalness: 0.4 })
      );
      panel.position.set(xQuench - 0.45, 1.1, 0.65);
      root.add(panel);
    }

    // Animated Flow Particles along 4-stage centerline curve
    const flowCurve = new THREE.CatmullRomCurve3([
      new THREE.Vector3(xQuench - 1.1, qH - 0.15, 0),
      new THREE.Vector3(xQuench, qH - 0.15, 0),
      new THREE.Vector3(xQuench, 0.65, 0),
      new THREE.Vector3(xMid, 0.65, zMid),
      new THREE.Vector3(xCol, 0.65, zCol),
      new THREE.Vector3(xCol, colH, zCol),
      new THREE.Vector3(xBlower, 0.85, zBlower),
      new THREE.Vector3(xBlower + 0.18, 3.1, zBlower),
    ]);
    const pCount = 90;
    const pPositions = new Float32Array(pCount * 3);
    const pColors = new Float32Array(pCount * 3);
    const pPhases = new Float32Array(pCount);
    for (let i = 0; i < pCount; i++) {
      pPhases[i] = i / pCount;
    }
    const pGeo = new THREE.BufferGeometry();
    pGeo.setAttribute("position", new THREE.BufferAttribute(pPositions, 3));
    pGeo.setAttribute("color", new THREE.BufferAttribute(pColors, 3));
    const pMat = new THREE.PointsMaterial({
      size: 0.085,
      vertexColors: true,
      transparent: true,
      opacity: 0.9,
    });
    const particlePoints = new THREE.Points(pGeo, pMat);
    root.add(particlePoints);

    const resize = () => {
      const rect = container.getBoundingClientRect();
      const w = Math.max(280, rect.width);
      const h = Math.max(240, rect.height);
      camera.aspect = w / h;
      camera.updateProjectionMatrix();
      renderer.setSize(w, h, false);
    };
    resize();
    const ro = new ResizeObserver(resize);
    ro.observe(container);

    const clock = new THREE.Clock();
    const tmpPt = new THREE.Vector3();
    const hotColor = new THREE.Color(0xf97316);
    const coolColor = new THREE.Color(0x38bdf8);
    const cleanColor = new THREE.Color(0x10b981);
    const curColor = new THREE.Color();

    let rafId = 0;
    const animate = () => {
      rafId = requestAnimationFrame(animate);
      if (!activeRef.current || document.hidden) return;

      const dt = Math.min(0.05, clock.getDelta());

      if (camGoalRef.current) {
        camera.position.lerp(camGoalRef.current.cam, 0.09);
        controls.target.lerp(camGoalRef.current.target, 0.09);
        if (camera.position.distanceTo(camGoalRef.current.cam) < 0.04) {
          camGoalRef.current = null;
        }
      }

      if (impellerGroup) {
        impellerGroup.rotation.z -= dt * 5.0;
      }

      particlePoints.visible = particlesRef.current;
      if (particlePoints.visible) {
        const pos = pGeo.attributes.position as THREE.BufferAttribute;
        const col = pGeo.attributes.color as THREE.BufferAttribute;
        for (let i = 0; i < pCount; i++) {
          pPhases[i] = (pPhases[i] + dt * 0.16) % 1.0;
          const t = pPhases[i];
          flowCurve.getPointAt(t, tmpPt);
          pos.setXYZ(i, tmpPt.x, tmpPt.y, tmpPt.z);
          if (t < 0.35) {
            curColor.copy(hotColor).lerp(coolColor, t / 0.35);
          } else {
            curColor.copy(coolColor).lerp(cleanColor, (t - 0.35) / 0.65);
          }
          col.setXYZ(i, curColor.r, curColor.g, curColor.b);
        }
        pos.needsUpdate = true;
        col.needsUpdate = true;
      }

      controls.update();
      renderer.render(scene, camera);
    };
    animate();

    return () => {
      cancelAnimationFrame(rafId);
      ro.disconnect();
      controls.dispose();
      renderer.dispose();
    };
  }, [evaluation, cutaway, showSystemsOverlay, onWebGLUnavailable]);

  if (webglError) {
    return (
      <div style={{ padding: 20, color: "#fcd34d", fontFamily: "monospace" }}>
        {webglError}
      </div>
    );
  }

  return (
    <div
      id="viewport-3d"
      ref={containerRef}
      style={{ width: "100%", height: "100%" }}
    />
  );
}
