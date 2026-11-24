"use client";

import React, { useCallback, useEffect, useRef, useState } from "react";
import type {
  StudioEvaluation,
  StudioPreset,
  StudioProject,
} from "@/types/kilnscrub";
import {
  downloadBuildPacket,
  downloadExportArtifact,
  evaluateStudioProject,
  fetchPresets,
  reconcileProject,
  triggerBlobDownload,
} from "@/lib/api";
import { Topbar, type ActiveTab } from "@/components/Topbar";
import { StudioSidebar } from "@/components/StudioSidebar";
import { DesignTab } from "@/components/DesignTab";
import { GuideTab } from "@/components/GuideTab";
import { ControlTab } from "@/components/ControlTab";
import { ShopPlansTab } from "@/components/ShopPlansTab";

export default function KilnScrubDashboard() {
  const [activeTab, setActiveTab] = useState<ActiveTab>("design");
  const [units, setUnits] = useState<"in" | "mm">("in");
  const [presets, setPresets] = useState<StudioPreset[]>([]);
  const [project, setProject] = useState<StudioProject | null>(null);
  const [evaluation, setEvaluation] = useState<StudioEvaluation | null>(null);
  const [evalError, setEvalError] = useState<string>("");
  const [busy, setBusy] = useState(false);
  const [toast, setToast] = useState<string>("");
  const [controlDot, setControlDot] = useState<"off" | "ok" | "warn" | "tripped">("off");

  const abortRef = useRef<AbortController | null>(null);
  const toastTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const showToast = useCallback((msg: string) => {
    setToast(msg);
    if (toastTimerRef.current) clearTimeout(toastTimerRef.current);
    toastTimerRef.current = setTimeout(() => setToast(""), 3200);
  }, []);

  // Load curated presets and initialize with reference wood_salt_24 preset
  useEffect(() => {
    let cancelled = false;
    fetchPresets()
      .then((list) => {
        if (cancelled || list.length === 0) return;
        setPresets(list);
        const ref = list.find((p) => p.id === "wood_salt_24") || list[0];
        setProject(reconcileProject(ref.project));
      })
      .catch((err) => {
        if (!cancelled) {
          setEvalError(
            `Failed to connect to kilnscrub backend API: ${
              err instanceof Error ? err.message : String(err)
            }`
          );
        }
      });
    return () => {
      cancelled = true;
    };
  }, []);

  // Debounced StudioEvaluation request whenever project changes
  useEffect(() => {
    if (!project) return;
    const timer = setTimeout(() => {
      abortRef.current?.abort();
      const ac = new AbortController();
      abortRef.current = ac;
      evaluateStudioProject(project, ac.signal)
        .then((res) => {
          setEvaluation(res);
          setEvalError("");
        })
        .catch((err) => {
          if (err instanceof DOMException && err.name === "AbortError") return;
          setEvalError(err instanceof Error ? err.message : String(err));
        });
    }, 90);
    return () => clearTimeout(timer);
  }, [project]);

  const handleSelectPreset = (preset: StudioPreset) => {
    setProject(reconcileProject(preset.project));
    showToast(`Loaded preset: ${preset.name}`);
  };

  const handleSaveJSON = () => {
    if (!project) return;
    const blob = new Blob([JSON.stringify(project, null, 2) + "\n"], {
      type: "application/json",
    });
    triggerBlobDownload(blob, `${project.preset_id || "studio"}-project.json`);
    showToast("Saved project.json");
  };

  const handleLoadJSON = async (file: File) => {
    try {
      const text = await file.text();
      const parsed = JSON.parse(text) as StudioProject;
      if (!parsed || typeof parsed !== "object" || !parsed.input) {
        throw new Error("Invalid project.json structure");
      }
      setProject(reconcileProject(parsed));
      showToast(`Loaded ${file.name}`);
    } catch (err) {
      showToast(
        `Could not load project JSON: ${err instanceof Error ? err.message : String(err)}`
      );
    }
  };

  const handleExport = async (format: "stl" | "dxf" | "bom_csv" | "bom_md") => {
    if (!project) return;
    setBusy(true);
    try {
      const fname = await downloadExportArtifact(format, project.input);
      showToast(`Downloaded ${fname}`);
    } catch (err) {
      showToast(`Export failed: ${err instanceof Error ? err.message : String(err)}`);
    } finally {
      setBusy(false);
    }
  };

  const handleDownloadPacket = async () => {
    if (!project) return;
    setBusy(true);
    try {
      const fname = await downloadBuildPacket(project);
      showToast(`Downloaded tailored Build Packet: ${fname}`);
    } catch (err) {
      showToast(`Packet failed: ${err instanceof Error ? err.message : String(err)}`);
    } finally {
      setBusy(false);
    }
  };

  const applyBudgetCombo = () => {
    if (!project) return;
    setProject(
      reconcileProject({
        ...project,
        build_path: "cots_no_weld",
        zero_electricity: false,
        stage_alignment: "inline_id_blower",
        packed_bed_layout: "parallel_manifold",
        parallel_pipe_size_in: 4,
        packing_media_option: "diy_ceramic_rings",
        blower_drive_option: "compressor_air_atomize",
        compact_footprint: true,
        shared_baffled_sump: true,
        control_tier: "electromechanical_only",
      })
    );
    showToast("Applied Budget Studio Combo (4\" Uniseal + DIY Rings + Air-Atomize + U-Turn)");
  };

  const applyDeluxe = () => {
    if (!project) return;
    setProject(
      reconcileProject({
        ...project,
        build_path: "custom_fab_deluxe",
        zero_electricity: false,
        packed_bed_layout: "single_column",
        include_silencer: true,
      })
    );
    showToast("Applied Custom Fab Deluxe (TIG / FRP / ANSI 150 / Skid / -22 dBA)");
  };

  const applyOffGrid = () => {
    if (!project) return;
    setProject(
      reconcileProject({
        ...project,
        build_path: "local_ceramic_masonry",
        zero_electricity: true,
        local_region: "nairobi",
        local_alkali: "soda_ash",
        local_draft_mode: "natural_chimney",
        available_voltage: 0,
      })
    );
    showToast("Applied Off-Grid Local Ceramic Mode (0 W · Nairobi)");
  };

  const applyOffGridSteam = () => {
    if (!project) return;
    setProject(
      reconcileProject({
        ...project,
        build_path: "local_ceramic_masonry",
        zero_electricity: true,
        local_region: "nairobi",
        local_alkali: "soda_ash",
        local_draft_mode: "steam_ejector_hp",
        available_voltage: 0,
      })
    );
    showToast("Applied Off-Grid 60 psig Certified Steam-Jet Mode (0 W)");
  };

  return (
    <div className="app-shell">
      <Topbar
        activeTab={activeTab}
        onSelectTab={setActiveTab}
        units={units}
        onToggleUnits={() => setUnits((u) => (u === "in" ? "mm" : "in"))}
        controlDotState={controlDot}
        busy={busy}
        isLocalCeramic={project?.build_path === "local_ceramic_masonry"}
        onExport={handleExport}
        onDownloadPacket={handleDownloadPacket}
      />

      {evalError && (
        <div
          style={{
            margin: "10px 16px 0",
            padding: "9px 12px",
            borderRadius: 8,
            background: "rgba(239, 68, 68, 0.16)",
            border: "1px solid rgba(239, 68, 68, 0.55)",
            color: "#fca5a5",
            fontSize: 12.5,
          }}
        >
          Validation / API Notice: {evalError}
        </div>
      )}

      {activeTab === "design" && (
        <main className="workspace-design">
          {project && (
            <StudioSidebar
              presets={presets}
              project={project}
              onChange={setProject}
              onSelectPreset={handleSelectPreset}
              onSaveJSON={handleSaveJSON}
              onLoadJSON={handleLoadJSON}
              onApplyBudgetCombo={applyBudgetCombo}
              onApplyDeluxe={applyDeluxe}
              onApplyOffGrid={applyOffGrid}
              onApplyOffGridSteam={applyOffGridSteam}
            />
          )}
          {evaluation ? (
            <DesignTab
              evaluation={evaluation}
              units={units}
              active={activeTab === "design"}
            />
          ) : (
            <div className="main-stage">
              <div className="card">Loading kilnscrub studio solver...</div>
            </div>
          )}
        </main>
      )}

      {activeTab === "guide" && evaluation && (
        <GuideTab
          evaluation={evaluation}
          busy={busy}
          onDownloadPacket={handleDownloadPacket}
          onSaveJSON={handleSaveJSON}
        />
      )}

      <ControlTab
        active={activeTab === "control"}
        onStatusDotChange={setControlDot}
        onToast={showToast}
      />

      {activeTab === "plans" && evaluation && (
        <ShopPlansTab evaluation={evaluation} units={units} />
      )}

      {toast && (
        <div className="toast" role="status" aria-live="polite">
          {toast}
        </div>
      )}
    </div>
  );
}
