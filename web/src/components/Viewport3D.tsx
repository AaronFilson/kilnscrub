"use client";

import React, { useEffect, useRef, useState } from "react";
import * as THREE from "three";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls.js";
import type { StudioEvaluation } from "@/types/kilnscrub";
import {
  createFloorShadowTexture,
  createMaterialBundle,
  createStudioEnvironment,
  type MaterialBundle,
} from "./scene3d/materials";
import { buildLocalCeramicAssembly } from "./scene3d/localCeramicBuilder";
import { buildIndustrialAssembly } from "./scene3d/industrialBuilder";
import { FlowParticleSystem } from "./scene3d/particles";
import type { StageKey, StageView } from "./scene3d/types";

interface Viewport3DProps {
  evaluation: StudioEvaluation;
  cutaway: boolean;
  particles: boolean;
  showSystemsOverlay: boolean;
  showDimensions: boolean;
  units: "in" | "mm";
  focusStage: StageKey;
  resetCounter: number;
  active: boolean;
  onScaleSummary?: (summary: string) => void;
  onWebGLUnavailable?: () => void;
}

interface EngineState {
  renderer: THREE.WebGLRenderer;
  scene: THREE.Scene;
  camera: THREE.PerspectiveCamera;
  controls: OrbitControls;
  assemblyGroup: THREE.Group;
  particleSystem: FlowParticleSystem;
  envTexture: THREE.Texture | null;
  floorShadowTex: THREE.Texture | null;
  stageViews: Record<StageKey, StageView> | null;
  impellerMesh: THREE.Object3D | null;
  impellerDir: number;
  diluterImpellerMesh: THREE.Object3D | null;
  lastArchSig: string;
  hasFramedInitial: boolean;
  currentMatBundle: MaterialBundle | null;
  currentDisposableTextures: THREE.Texture[];
}

function disposeAssemblyGroup(
  group: THREE.Group,
  matBundle: MaterialBundle | null,
  extraTextures: THREE.Texture[]
): void {
  const geos = new Set<THREE.BufferGeometry>();
  const mats = new Set<THREE.Material>();
  while (group.children.length > 0) {
    const child = group.children[0];
    child.traverse((obj) => {
      if (obj instanceof THREE.Mesh || obj instanceof THREE.Line || obj instanceof THREE.LineSegments || obj instanceof THREE.Sprite) {
        if (obj.geometry) geos.add(obj.geometry);
        if (obj.material) {
          if (Array.isArray(obj.material)) {
            obj.material.forEach((m) => mats.add(m));
          } else {
            mats.add(obj.material);
          }
        }
      }
    });
    group.remove(child);
  }
  geos.forEach((g) => g.dispose());
  mats.forEach((m) => m.dispose());
  matBundle?.dispose();
  extraTextures.forEach((t) => t.dispose());
}

export function Viewport3D({
  evaluation,
  cutaway,
  particles,
  showSystemsOverlay,
  showDimensions,
  units,
  focusStage,
  resetCounter,
  active,
  onScaleSummary,
  onWebGLUnavailable,
}: Viewport3DProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const engineRef = useRef<EngineState | null>(null);
  const activeRef = useRef(active);
  const particlesRef = useRef(particles);
  const focusStageRef = useRef<StageKey>(focusStage);
  const onScaleSummaryRef = useRef(onScaleSummary);
  const onWebGLUnavailableRef = useRef(onWebGLUnavailable);
  const camGoalRef = useRef<{ cam: THREE.Vector3; target: THREE.Vector3 } | null>(null);
  const [webglError, setWebglError] = useState<string | null>(null);

  activeRef.current = active;
  particlesRef.current = particles;
  focusStageRef.current = focusStage;
  onScaleSummaryRef.current = onScaleSummary;
  onWebGLUnavailableRef.current = onWebGLUnavailable;

  // 1. Initialize persistent WebGLRenderer, Scene, Lighting, Studio PMREM Environment & Controls ONCE
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
      onWebGLUnavailableRef.current?.();
      return;
    }

    renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2));
    renderer.toneMapping = THREE.ACESFilmicToneMapping;
    renderer.toneMappingExposure = 1.14;
    container.innerHTML = "";
    container.appendChild(renderer.domElement);

    const scene = new THREE.Scene();
    scene.background = new THREE.Color(0x0a1320);
    scene.fog = new THREE.FogExp2(0x0a1320, 0.032);

    const envTexture = createStudioEnvironment(renderer);
    if (envTexture) {
      scene.environment = envTexture;
    }

    const camera = new THREE.PerspectiveCamera(38, 16 / 9, 0.05, 80);
    camera.position.set(1.2, 2.5, 6.6);

    const controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.07;
    controls.target.set(0.15, 1.45, 0);
    controls.minDistance = 1.0;
    controls.maxDistance = 18;
    controls.maxPolarAngle = Math.PI * 0.495; // keep camera above floor

    // Stop camera glide immediately when the user grabs the viewport
    const onControlStart = () => {
      camGoalRef.current = null;
    };
    controls.addEventListener("start", onControlStart);

    // 5-Point Studio Lighting Rig
    scene.add(new THREE.AmbientLight(0xc8dcf6, 0.72));
    scene.add(new THREE.HemisphereLight(0xe0f2fe, 0x263447, 0.78));

    const key = new THREE.DirectionalLight(0xffffff, 1.38);
    key.position.set(4.8, 7.8, 5.8);
    scene.add(key);

    const fill = new THREE.DirectionalLight(0x93c5fd, 0.50);
    fill.position.set(-4.8, 3.0, 4.6);
    scene.add(fill);

    const rim = new THREE.DirectionalLight(0x38bdf8, 0.64);
    rim.position.set(-5.2, 3.4, -4.6);
    scene.add(rim);

    const rearFill = new THREE.DirectionalLight(0xdbeafe, 0.85);
    rearFill.position.set(3.6, 4.6, -5.6);
    scene.add(rearFill);

    // Metric Scale Workshop Grid (12m x 12m, 0.5m cells) + Ground Contact AO Disc
    const grid = new THREE.GridHelper(12, 24, 0x2c476b, 0x162437);
    grid.position.y = 0;
    scene.add(grid);

    const floorShadowTex = createFloorShadowTexture();
    const shadowPlane = new THREE.Mesh(
      new THREE.PlaneGeometry(8.2, 4.8),
      new THREE.MeshBasicMaterial({
        map: floorShadowTex || undefined,
        color: floorShadowTex ? 0xffffff : 0x050a12,
        transparent: true,
        opacity: 0.85,
        depthWrite: false,
      })
    );
    shadowPlane.rotation.x = -Math.PI / 2;
    shadowPlane.position.y = 0.002;
    scene.add(shadowPlane);

    const assemblyGroup = new THREE.Group();
    scene.add(assemblyGroup);

    const particleSystem = new FlowParticleSystem();
    scene.add(particleSystem.group);

    const engine: EngineState = {
      renderer,
      scene,
      camera,
      controls,
      assemblyGroup,
      particleSystem,
      envTexture,
      floorShadowTex,
      stageViews: null,
      impellerMesh: null,
      impellerDir: 1,
      diluterImpellerMesh: null,
      lastArchSig: "",
      hasFramedInitial: false,
      currentMatBundle: null,
      currentDisposableTextures: [],
    };
    engineRef.current = engine;

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
    let rafId = 0;

    const animate = () => {
      rafId = requestAnimationFrame(animate);
      if (!activeRef.current || document.hidden) {
        clock.getDelta(); // prevent dt spike on resume
        return;
      }

      const dt = Math.min(0.05, clock.getDelta());

      if (camGoalRef.current) {
        camera.position.lerp(camGoalRef.current.cam, 0.09);
        controls.target.lerp(camGoalRef.current.target, 0.09);
        if (
          camera.position.distanceTo(camGoalRef.current.cam) < 0.03 &&
          controls.target.distanceTo(camGoalRef.current.target) < 0.03
        ) {
          camGoalRef.current = null;
        }
      }

      if (engine.impellerMesh) {
        engine.impellerMesh.rotation.z += dt * 4.4 * engine.impellerDir;
      }
      if (engine.diluterImpellerMesh) {
        engine.diluterImpellerMesh.rotation.z += dt * 7.5;
      }

      particleSystem.setVisible(particlesRef.current);
      if (particlesRef.current) {
        particleSystem.update(dt);
      }

      controls.update();
      renderer.render(scene, camera);
    };
    animate();

    return () => {
      cancelAnimationFrame(rafId);
      ro.disconnect();
      controls.removeEventListener("start", onControlStart);
      controls.dispose();
      disposeAssemblyGroup(
        engine.assemblyGroup,
        engine.currentMatBundle,
        engine.currentDisposableTextures
      );
      particleSystem.dispose();
      shadowPlane.geometry.dispose();
      (shadowPlane.material as THREE.Material).dispose();
      floorShadowTex?.dispose();
      envTexture?.dispose();
      renderer.dispose();
      engineRef.current = null;
    };
  }, []);

  // 2. Rebuild parametric 3D assembly & flow paths whenever design/cutaway/options change
  useEffect(() => {
    const engine = engineRef.current;
    if (!engine) return;

    disposeAssemblyGroup(
      engine.assemblyGroup,
      engine.currentMatBundle,
      engine.currentDisposableTextures
    );
    engine.assemblyGroup.position.set(0, 0, 0);

    const { project, frugal } = evaluation;
    const isLocal =
      project.build_path === "local_ceramic_masonry" || project.zero_electricity;
    const isHeavyAlloy = project.corrosion_tier === "heavy_c276_frp";
    const isSubsonicSteam =
      isLocal &&
      String(project.local_draft_mode || frugal?.draft_mode || "").toLowerCase() ===
        "steam_ejector_subsonic";

    const mats = createMaterialBundle(cutaway, isHeavyAlloy, isSubsonicSteam);
    engine.currentMatBundle = mats;

    const buildOpts = {
      evaluation,
      cutaway,
      showSystemsOverlay,
      showDimensions,
      units,
      cameraFovDeg: engine.camera.fov,
      cameraAspect: engine.camera.aspect || 16 / 9,
    };

    const result = isLocal
      ? buildLocalCeramicAssembly(engine.assemblyGroup, mats, buildOpts)
      : buildIndustrialAssembly(engine.assemblyGroup, mats, buildOpts);

    engine.currentDisposableTextures = result.disposableTextures;
    engine.stageViews = result.stageViews;
    engine.impellerMesh = result.impellerMesh;
    engine.impellerDir = result.impellerDir;
    engine.diluterImpellerMesh = result.diluterImpellerMesh;

    engine.particleSystem.configure(
      result.primaryFlow,
      result.motiveFlow,
      result.diluterFlow,
      result.sprayZones
    );

    onScaleSummaryRef.current?.(result.scaleSummary);

    const targetView =
      result.stageViews[focusStageRef.current] || result.stageViews.overview;

    if (!engine.hasFramedInitial) {
      engine.hasFramedInitial = true;
      engine.lastArchSig = result.architectureSignature;
      engine.camera.position.copy(targetView.cam);
      engine.controls.target.copy(targetView.target);
      engine.controls.update();
    } else if (engine.lastArchSig !== result.architectureSignature) {
      // Smoothly re-frame when user switches major architecture (e.g. Inline -> Local Ceramic -> Eductor)
      engine.lastArchSig = result.architectureSignature;
      camGoalRef.current = {
        cam: targetView.cam.clone(),
        target: targetView.target.clone(),
      };
    }
  }, [
    evaluation,
    cutaway,
    showSystemsOverlay,
    showDimensions,
    units,
  ]);

  // 3. Smoothly fly camera when user clicks a Stage Focus button or Reset View
  useEffect(() => {
    const engine = engineRef.current;
    if (!engine || !engine.stageViews) return;
    const view = engine.stageViews[focusStage] || engine.stageViews.overview;
    camGoalRef.current = {
      cam: view.cam.clone(),
      target: view.target.clone(),
    };
  }, [focusStage, resetCounter]);

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
