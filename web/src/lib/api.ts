import type {
  ControlConfig,
  ControlStatus,
  ControlStatusResponse,
  DesignInput,
  ScrubberDesign,
  StudioEvaluation,
  StudioPreset,
  StudioProject,
} from "@/types/kilnscrub";

export function kToF(k: number): number {
  return (k - 273.15) * (9 / 5) + 32;
}

export function mToIn(m: number): number {
  return m / 0.0254;
}

export function mToFt(m: number): number {
  return m / 0.3048;
}

export function fmt(v: number | undefined | null, digits = 1): string {
  if (typeof v !== "number" || !Number.isFinite(v)) return "—";
  return v.toFixed(digits);
}

export function fmtUSD(v: number | undefined | null): string {
  if (typeof v !== "number" || !Number.isFinite(v)) return "—";
  return `$${v.toLocaleString("en-US", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })}`;
}

export const CONE_OPTIONS = [
  { cone: "Cone 06", tempF: 1830 },
  { cone: "Cone 6", tempF: 2232 },
  { cone: "Cone 8", tempF: 2305 },
  { cone: "Cone 10", tempF: 2381 },
  { cone: "Cone 11", tempF: 2400 },
] as const;

export function inferConeFromTemp(tempF: number): string {
  let best: string = "Cone 11";
  let bestDiff = Infinity;
  for (const c of CONE_OPTIONS) {
    const d = Math.abs(tempF - c.tempF);
    if (d < bestDiff) {
      bestDiff = d;
      best = c.cone;
    }
  }
  return best;
}

// Reconciles A2 mutual-exclusivity rules client-side so invalid combinations
// are disabled and never fail validation when switching build paths or alignments.
export function reconcileProject(proj: StudioProject): StudioProject {
  const next: StudioProject = {
    ...proj,
    input: { ...proj.input },
  };

  if (next.build_path !== "local_ceramic_masonry") {
    next.zero_electricity = false;
    next.local_draft_mode = "natural_chimney";
    if (next.available_voltage === 0) {
      next.available_voltage = 240;
    }
  } else if (next.zero_electricity) {
    next.available_voltage = 0;
    next.include_plume_diluter = false;
  }

  if (next.build_path === "custom_fab_deluxe") {
    next.packed_bed_layout = "single_column";
  }

  if (next.stage_alignment === "eductor_dry_fan") {
    next.venturi_layout = "single_throat";
    if (
      next.blower_drive_option === "compressor_air_atomize" ||
      next.blower_drive_option === "compressor_eductor_only" ||
      next.blower_drive_option === "dual_bypass_blowers"
    ) {
      next.blower_drive_option = "industrial_radial_vfd";
    }
  }

  if (next.vapor_glaze_mode === "salt") {
    if (!(next.salting_hours > 0)) {
      next.salting_hours = Math.min(3.0, next.firing_duration_hours || 10);
    }
    if (next.salting_hours > next.firing_duration_hours) {
      next.salting_hours = next.firing_duration_hours;
    }
    if (!(next.input.salt_rate_lbs_per_hr > 0)) {
      next.input.salt_rate_lbs_per_hr = 2.0;
    }
    if (next.input.fuel_type === "wood") next.input.fuel_type = "wood_salt";
    if (next.input.fuel_type === "gas") next.input.fuel_type = "gas_salt";
  } else {
    next.salting_hours = 0;
    next.input.salt_rate_lbs_per_hr = 0;
    if (next.input.fuel_type === "wood_salt") next.input.fuel_type = "wood";
    if (next.input.fuel_type === "gas_salt") next.input.fuel_type = "gas";
  }

  return next;
}

async function parseErrorMessage(res: Response): Promise<string> {
  try {
    const data = (await res.json()) as { error?: string };
    if (data && data.error) return data.error;
  } catch {
    // ignore non-JSON body
  }
  return `HTTP ${res.status} ${res.statusText}`;
}

export async function fetchPresets(): Promise<StudioPreset[]> {
  const res = await fetch("/api/studio/presets", { cache: "no-store" });
  if (!res.ok) throw new Error(await parseErrorMessage(res));
  const data = (await res.json()) as { presets: StudioPreset[] };
  return data.presets || [];
}

export async function fetchDefaults(): Promise<DesignInput> {
  const res = await fetch("/api/defaults", { cache: "no-store" });
  if (!res.ok) throw new Error(await parseErrorMessage(res));
  return (await res.json()) as DesignInput;
}

export async function evaluateStudioProject(
  project: StudioProject,
  signal?: AbortSignal
): Promise<StudioEvaluation> {
  const res = await fetch("/api/studio/evaluate", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(project),
    signal,
  });
  if (!res.ok) throw new Error(await parseErrorMessage(res));
  return (await res.json()) as StudioEvaluation;
}

export async function sizeScrubber(
  input: DesignInput,
  signal?: AbortSignal
): Promise<ScrubberDesign> {
  const res = await fetch("/api/size", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
    signal,
  });
  if (!res.ok) throw new Error(await parseErrorMessage(res));
  return (await res.json()) as ScrubberDesign;
}

export async function downloadExportArtifact(
  format: "stl" | "dxf" | "bom_csv" | "bom_md",
  input: DesignInput
): Promise<string> {
  const res = await fetch(`/api/export?format=${encodeURIComponent(format)}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!res.ok) throw new Error(await parseErrorMessage(res));
  const blob = await res.blob();
  const fallbackNames: Record<typeof format, string> = {
    stl: "scrubber.stl",
    dxf: "scrubber.dxf",
    bom_csv: "bom.csv",
    bom_md: "bom.md",
  };
  const filename =
    parseFilenameFromDisposition(res.headers.get("Content-Disposition")) ||
    fallbackNames[format];
  triggerBlobDownload(blob, filename);
  return filename;
}

export async function downloadBuildPacket(
  project: StudioProject
): Promise<string> {
  const res = await fetch("/api/studio/packet", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(project),
  });
  if (!res.ok) throw new Error(await parseErrorMessage(res));
  const blob = await res.blob();
  const filename =
    parseFilenameFromDisposition(res.headers.get("Content-Disposition")) ||
    `kilnscrub-${project.preset_id || "studio"}-packet.zip`;
  triggerBlobDownload(blob, filename);
  return filename;
}

export async function fetchControlStatus(): Promise<ControlStatusResponse> {
  const res = await fetch("/api/control/status", { cache: "no-store" });
  if (!res.ok) return { enabled: false };
  return (await res.json()) as ControlStatusResponse;
}

export async function postControlReset(): Promise<ControlStatus> {
  const res = await fetch("/api/control/reset", { method: "POST" });
  if (!res.ok) throw new Error(await parseErrorMessage(res));
  return (await res.json()) as ControlStatus;
}

export async function postControlSetpoint(payload: {
  target_draft_in_wc?: number;
  quench_setpoint_f?: number;
}): Promise<ControlConfig> {
  const res = await fetch("/api/control/setpoint", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (!res.ok) throw new Error(await parseErrorMessage(res));
  return (await res.json()) as ControlConfig;
}

export async function postControlFault(
  action: string,
  value?: number
): Promise<ControlStatus> {
  const res = await fetch("/api/control/fault", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ action, value }),
  });
  if (!res.ok) throw new Error(await parseErrorMessage(res));
  return (await res.json()) as ControlStatus;
}

function parseFilenameFromDisposition(cd: string | null): string | null {
  if (!cd) return null;
  const match = /filename="?([^";]+)"?/i.exec(cd);
  return match ? match[1] : null;
}

export function triggerBlobDownload(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}
