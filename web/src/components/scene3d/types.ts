import * as THREE from "three";
import type { StudioEvaluation } from "@/types/kilnscrub";

export type StageKey =
  | "overview"
  | "inlet"
  | "quench"
  | "venturi"
  | "packed"
  | "blower";

export interface StageView {
  cam: THREE.Vector3;
  target: THREE.Vector3;
}

export type FlowStage =
  | "inlet"
  | "quench"
  | "venturi"
  | "packed"
  | "blower"
  | "plume"
  | "steam"
  | "motive"
  | "diluter";

export interface FlowSample {
  pt: THREE.Vector3;
  /** Maximum radial swirl offset in meters — strictly bounded inside vessel/pipe inner radius. */
  swirlRadius: number;
  /** Relative local velocity multiplier (e.g. 2.8x inside venturi throat, 0.55x inside wide bed). */
  speedMul: number;
  stage: FlowStage;
}

export interface SprayZone {
  topCenter: THREE.Vector3;
  topRadius: number;
  botRadius: number;
  height: number;
  color: number;
  count: number;
  speed: number;
}

export interface SceneBuildOptions {
  evaluation: StudioEvaluation;
  cutaway: boolean;
  showSystemsOverlay: boolean;
  showDimensions: boolean;
  units: "in" | "mm";
  cameraFovDeg: number;
  cameraAspect: number;
}

export interface SceneBuildResult {
  stageViews: Record<StageKey, StageView>;
  primaryFlow: FlowSample[];
  motiveFlow: FlowSample[];
  diluterFlow: FlowSample[];
  sprayZones: SprayZone[];
  impellerMesh: THREE.Object3D | null;
  impellerDir: number;
  diluterImpellerMesh: THREE.Object3D | null;
  architectureSignature: string;
  scaleSummary: string;
  disposableTextures: THREE.Texture[];
}
