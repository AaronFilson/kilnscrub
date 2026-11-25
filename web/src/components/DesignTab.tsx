"use client";

import React, { useEffect, useRef, useState } from "react";
import type { StudioEvaluation } from "@/types/kilnscrub";
import { fmt, fmtUSD, kToF, triggerBlobDownload } from "@/lib/api";
import { Viewport3D } from "./Viewport3D";
import type { StageKey } from "./scene3d/types";
import { Schematic2D } from "./Schematic2D";
import { EngineeringCharts } from "./EngineeringCharts";

interface DesignTabProps {
  evaluation: StudioEvaluation;
  units: "in" | "mm";
  active: boolean;
}

const FUEL_LABELS: Record<string, string> = {
  wood_salt: "Wood + Salt",
  wood: "Wood",
  gas_salt: "Gas + Salt",
  gas: "Gas / Soda",
  coal_salt: "Coal + Salt",
  coal: "Bituminous Coal",
  peat_salt: "Peat + Salt",
  peat: "Peat / Turf",
  biomass_ag_salt: "Ag-Biomass + Salt",
  biomass_ag: "Ag-Biomass (PKS/Pomace)",
};

const BUILD_LABELS: Record<string, string> = {
  cots_no_weld: "No-Weld DIY",
  welded_dxf: "Welded Metal",
  custom_fab_deluxe: "Pro Deluxe",
  local_ceramic_masonry: "Local Ceramic / Off-Grid",
};

export function DesignTab({ evaluation, units, active }: DesignTabProps) {
  const [viewMode, setViewMode] = useState<"3d" | "2d">("3d");
  const [detail2DMode, setDetail2DMode] = useState<"detail" | "outline">("detail");
  const [show2DPlumbing, setShow2DPlumbing] = useState(true);
  const [show2DControls, setShow2DControls] = useState(true);
  const [viewOptsOpen, setViewOptsOpen] = useState(false);
  const [showTour, setShowTour] = useState(false);
  const [detailTab, setDetailTab] = useState<"overview" | "charts" | "notices">("overview");
  const svg2DRef = useRef<SVGSVGElement | null>(null);
  const viewOptsRef = useRef<HTMLDivElement | null>(null);
  const detailsRef = useRef<HTMLElement | null>(null);

  const [cutaway, setCutaway] = useState(true);
  const [particles, setParticles] = useState(true);
  const [showSystems3D, setShowSystems3D] = useState(false);
  const [showDimensions, setShowDimensions] = useState(true);
  const [focusStage, setFocusStage] = useState<StageKey>("overview");
  const [resetCounter, setResetCounter] = useState(0);
  const [scaleSummary, setScaleSummary] = useState("Scale 1:1 m");

  const { project, design, cots, urban, frugal, shop_plan: plan } = evaluation;
  const isLocal = project.build_path === "local_ceramic_masonry";
  const isDeluxe = project.build_path === "custom_fab_deluxe";
  const prevDeluxeRef = useRef(isDeluxe);

  useEffect(() => {
    try {
      if (!localStorage.getItem("kilnscrub:tourDismissed")) {
        setShowTour(true);
      }
    } catch {
      // ignore storage errors
    }
  }, []);

  useEffect(() => {
    if (isDeluxe && !prevDeluxeRef.current) {
      setShowSystems3D(true);
    }
    prevDeluxeRef.current = isDeluxe;
  }, [isDeluxe]);

  useEffect(() => {
    if (!viewOptsOpen) return;
    const onDocClick = (e: MouseEvent) => {
      if (viewOptsRef.current && !viewOptsRef.current.contains(e.target as Node)) {
        setViewOptsOpen(false);
      }
    };
    document.addEventListener("mousedown", onDocClick);
    return () => document.removeEventListener("mousedown", onDocClick);
  }, [viewOptsOpen]);

  const dismissTour = () => {
    setShowTour(false);
    try {
      localStorage.setItem("kilnscrub:tourDismissed", "1");
    } catch {
      // ignore
    }
  };

  const quenchOutF = kToF(design.quench.outlet_temp_k);
  const allNotices = [
    ...(design.warnings || []),
    ...(evaluation.architecture_options.warnings || []),
    ...(frugal?.warnings || []),
  ];

  const stageExplanation = (() => {
    switch (focusStage) {
      case "inlet":
        return isLocal && frugal?.afterburner_enabled
          ? `Kiln Inlet & Stage 0 Tar Afterburner: ${Math.round(design.input.peak_exit_temp_f)}°F flue gas passes through a hot (>750°C) secondary-air firebrick checkerwork throat that burns off 70–85% of heavy coal/peat/PKS tars before quenching.`
          : `Kiln Exhaust Inlet: Draws ${Math.round(design.input.peak_exit_temp_f)}°F exhaust gently at ${fmt(design.input.target_kiln_draft_in_wc, 3)} inWC suction${!isLocal ? " with an automatic 200°F fail-open safety bypass damper." : " through an all-brick refractory flue."}`;
      case "quench":
        return isLocal
          ? `Stage 1 Brick Quench: Water spray cools hot kiln exhaust down to ~${fmt(quenchOutF, 0)}°F inside an all-ceramic chamber (>1,800°F dry-run safe — never melts if water stops).`
          : `Stage 1 Spray Quench: Cools ${Math.round(design.input.peak_exit_temp_f)}°F exhaust to ${fmt(quenchOutF, 1)}°F using ${fmt(design.quench.total_spray_water_gpm, 2)} GPM spray so downstream CPVC stays safely below the 200°F limit.`;
      case "venturi":
        return isLocal && frugal?.steam_enabled
          ? `Stage 2 Steam-Jet Ejector (${fmt(frugal.steam_pressure_psig, 0)} psig): A high-speed steam jet pulls draft (${fmt(frugal.ejector_draft_inwc, 2)} inWC) at 0 W electricity and grows fine salt fume into catchable droplets (~${fmt(frugal.submicron_nacl_capture_pct, 0)}% capture).`
          : isLocal
            ? `Stage 2 Terracotta Baffles: Drops coarse ash and sparks into the brick settling pool with minimal draft resistance (${fmt(frugal?.coarse_ash_capture_pct ?? 90, 0)}% coarse ash caught).`
            : `Stage 2 Wet Venturi: Accelerates gas to ${fmt(design.venturi.throat_velocity_m_s, 0)} m/s to collide fine smoke & salt fume into water droplets (${fmt(design.venturi.overall_nacl_fume_efficiency * 100, 1)}% NaCl · ${fmt(design.venturi.overall_wood_ash_efficiency * 100, 1)}% ash).`;
      case "packed":
        return isLocal && frugal
          ? `Stage 3 Stoneware Absorber (${fmt(frugal.bed_diameter_in, 0)}" wide): Alkaline ${frugal.alkali_name} trickles down potter-made ceramic rings, neutralizing ${fmt(frugal.hcl_removal_pct, 1)}% HCl${(frugal.so2_inlet_ppm ?? 0) > 0 ? ` and ${fmt(frugal.so2_removal_pct, 1)}% SO₂` : ""}.`
          : `Stage 3 Packed-Bed Absorber: Alkaline baking-soda solution (pH 7.5–8.5) neutralizes ${fmt(design.packed_bed.hcl_removal_efficiency * 100, 1)}% of HCl acid gas${(design.packed_bed.so2_inlet_ppm ?? 0) > 0 ? ` & ${fmt((design.packed_bed.so2_removal_efficiency ?? 0) * 100, 1)}% SO₂` : ""} before the mist eliminator.`;
      case "blower":
        return isLocal && frugal
          ? `Stage 4 Exhaust Chimney (${fmt(frugal.chimney_height_m, 1)} m / ${fmt(frugal.chimney_height_ft, 0)} ft): Discharges clean scrubbed vapor outdoors above roof eaves (${fmt(frugal.available_draft_inwc, 2)} inWC available vs ${fmt(frugal.required_draft_inwc, 2)} inWC needed).`
          : `Stage 4 Induced-Draft Blower (${design.blower.recommended_motor_hp} HP): Pulls ${fmt(design.blower.flow_acfm, 0)} ACFM across all stages (${fmt(design.blower.total_static_pressure_in_wc, 1)} inWC) to keep steady kiln draft.`;
      default:
        return "Click any stage button above to zoom in and see what that section does, or drag the 3D model to explore.";
    }
  })();

  const readinessLabel =
    urban.overall_status === "ok"
      ? "Site Ready"
      : urban.overall_status === "caution"
        ? "Site: Caution"
        : "Site: Action Needed";

  const scrollToDetails = (tab: "overview" | "charts" | "notices") => {
    setDetailTab(tab);
    detailsRef.current?.scrollIntoView({ behavior: "smooth", block: "nearest" });
  };

  const siteChecks = [
    {
      title: "Power & Electrical",
      status: urban.electrical.status,
      headline: isLocal && frugal?.power_watts === 0
        ? "0 W · Off-Grid"
        : `${urban.electrical.required_voltage}V · ~${fmt(urban.electrical.required_amps, 1)} A`,
      detail: urban.electrical.summary,
    },
    {
      title: "Water Supply",
      status: urban.water.status,
      headline: `${fmt(urban.water.required_gpm, 2)} GPM (~${fmt(urban.water.total_firing_gallons, 0)} gal/firing)`,
      detail: urban.water.summary,
    },
    {
      title: "Neighbor Noise",
      status: urban.noise.status,
      headline: `${fmt(urban.noise.property_line_dba, 1)} dBA @ ${urban.noise.property_line_ft} ft`,
      detail: urban.noise.summary,
    },
    {
      title: "Winter Steam Plume",
      status: urban.plume.status,
      headline: `~${fmt(urban.plume.bare_plume_length_ft, 0)} ft${
        project.include_plume_diluter && !project.zero_electricity
          ? ` → ${fmt(urban.plume.diluted_plume_length_ft, 0)} ft`
          : ""
      }`,
      detail: urban.plume.summary,
    },
    {
      title: "Alkali & Drain",
      status: urban.drain.status,
      headline: isLocal && frugal
        ? `${fmt(frugal.alkali_kg_per_firing, 1)} kg ${frugal.alkali_name}`
        : urban.drain.bulk_packaging_label.slice(0, 38),
      detail: urban.drain.summary,
    },
    {
      title: "Materials & Care",
      status: urban.corrosion.status,
      headline: isLocal ? "All-Ceramic >1,800°F" : urban.corrosion.tier_label,
      detail: urban.corrosion.material_summary,
    },
  ];

  return (
    <div className="main-stage">
      {/* 1. Calm Studio Identity & Status Bar */}
      <div className="studio-header-bar">
        <div>
          <h1 className="studio-heading">{project.studio_name}</h1>
          <div className="studio-subtitle">
            {project.input.kiln_volume_cu_ft.toFixed(0)} cu.ft. ·{" "}
            {FUEL_LABELS[project.input.fuel_type] || project.input.fuel_type} ·{" "}
            {BUILD_LABELS[project.build_path] || project.build_path}
            {isLocal && frugal ? ` · ${frugal.region_label}` : ""}
          </div>
        </div>
        <div style={{ display: "flex", alignItems: "center", gap: 8, flexWrap: "wrap" }}>
          <button
            type="button"
            className={`pill pill-btn ${urban.overall_status}`}
            onClick={() => scrollToDetails("overview")}
            title="View all 6 site & neighbor readiness checks below"
          >
            {readinessLabel}
          </button>
          {allNotices.length > 0 && (
            <button
              type="button"
              className="pill pill-btn warn"
              onClick={() => scrollToDetails("notices")}
              title="View safety & engineering notices"
            >
              {allNotices.length} {allNotices.length === 1 ? "Notice" : "Notices"}
            </button>
          )}
          <button
            type="button"
            className="btn btn-sm btn-ghost"
            onClick={() => setShowTour((v) => !v)}
          >
            {showTour ? "Hide Guide" : "? How It Works"}
          </button>
        </div>
      </div>

      {/* Optional First-Time Friendly 4-Step Explainer */}
      {showTour && (
        <section className="welcome-strip" aria-label="How a Kiln Scrubber Works">
          <div className="card-header" style={{ marginBottom: 0 }}>
            <strong style={{ fontSize: 13.5, color: "#f8fafc" }}>
              New to kiln scrubbers? Here is how the 4 stages clean your exhaust:
            </strong>
            <button type="button" className="btn btn-sm btn-ghost" onClick={dismissTour}>
              Got it ✕
            </button>
          </div>
          <div className="welcome-steps">
            <div className="welcome-step">
              <strong>1. Cool (Quench)</strong>
              Water spray drops 2,400°F kiln heat down to ~184°F safely.
            </div>
            <div className="welcome-step">
              <strong>2. Catch Smoke &amp; Salt</strong>
              High-speed water mist or steam traps ash and fine salt haze.
            </div>
            <div className="welcome-step">
              <strong>3. Neutralize Acid</strong>
              Alkaline water (soda ash, lime, or baking soda) neutralizes HCl &amp; SO₂.
            </div>
            <div className="welcome-step">
              <strong>4. Vent Clean Steam</strong>
              A blower or chimney vents clean vapor outdoors while keeping gentle kiln draft.
            </div>
          </div>
        </section>
      )}

      {/* 2. Four Clean Hero Metric Tiles */}
      <section className="kpi-grid" aria-label="Key Engineering KPIs">
        <div className="kpi-card" style={{ ["--kpi-accent" as string]: "#a855f7" }}>
          <div className="kpi-label">
            <span>Estimated Build Cost</span>
            <span className="unit-literal">
              {isLocal && frugal ? `${fmt(frugal.power_watts, 0)} W` : `${design.blower.recommended_motor_hp} HP`}
            </span>
          </div>
          <div className="kpi-value">{fmtUSD(cots.estimated_total_usd)}</div>
          <div className="kpi-sub">
            {isLocal && frugal
              ? frugal.local_currency_estimate
              : `${fmt(evaluation.recirculation.water_saved_pct, 0)}% water saved via recirculation`}
          </div>
        </div>

        <div className="kpi-card" style={{ ["--kpi-accent" as string]: "var(--accent-venturi)" }}>
          <div className="kpi-label">
            <span>Smoke &amp; Salt Capture</span>
            {!isLocal && (
              <span className="unit-literal">
                d₅₀ {fmt(design.venturi.cut_diameter_d50_um, 2)} µm
              </span>
            )}
          </div>
          <div className="kpi-value">
            {isLocal && frugal
              ? `${fmt(frugal.submicron_nacl_capture_pct, 0)}%`
              : `${fmt(design.venturi.overall_nacl_fume_efficiency * 100, 1)}%`}
          </div>
          <div className="kpi-sub">
            Fine salt fume ·{" "}
            {isLocal && frugal
              ? `${fmt(frugal.coarse_ash_capture_pct, 0)}% coarse ash`
              : `${fmt(design.venturi.overall_wood_ash_efficiency * 100, 1)}% wood ash`}
          </div>
        </div>

        <div className="kpi-card" style={{ ["--kpi-accent" as string]: "var(--accent-packed)" }}>
          <div className="kpi-label">
            <span>Acid Neutralization</span>
            <span className="unit-literal">pH 7.5–8.5</span>
          </div>
          <div className="kpi-value">
            {fmt(
              isLocal && frugal
                ? frugal.hcl_removal_pct
                : design.packed_bed.hcl_removal_efficiency * 100,
              1
            )}
            % <span style={{ fontSize: 12, fontWeight: 600, color: "var(--text-muted)" }}>HCl</span>
          </div>
          <div className="kpi-sub">
            {(isLocal && (frugal?.so2_inlet_ppm ?? 0) > 0)
              ? `SO₂ ${fmt(frugal?.so2_removal_pct, 1)}% (${fmt(frugal?.so2_inlet_ppm, 0)} ppm in)`
              : (design.packed_bed.so2_inlet_ppm ?? 0) > 0
                ? `SO₂ ${fmt((design.packed_bed.so2_removal_efficiency ?? 0) * 100, 1)}% (${fmt(design.packed_bed.so2_inlet_ppm, 0)} ppm in)`
                : `${fmt(design.packed_bed.hcl_outlet_ppm, 1)} ppm out · ${fmt(cots.actual_percent_of_flood, 0)}% flood`}
          </div>
        </div>

        <div className="kpi-card" style={{ ["--kpi-accent" as string]: "var(--accent-quench)" }}>
          <div className="kpi-label">
            <span>Cooling &amp; Safety</span>
            <span
              style={{
                color: quenchOutF < 200 ? "#6ee7b7" : "#fca5a5",
                fontWeight: 700,
              }}
            >
              {isLocal ? "All-Ceramic >1800°F" : "CPVC 200°F Cutoff"}
            </span>
          </div>
          <div className="kpi-value">{fmt(quenchOutF, 1)} °F</div>
          <div className="kpi-sub">
            {isLocal && frugal
              ? `Draft ${frugal.draft_feasible ? "OK" : "Short"} (${frugal.draft_margin_inwc >= 0 ? "+" : ""}${fmt(frugal.draft_margin_inwc, 2)} inWC)`
              : `Spray ${fmt(design.quench.total_spray_water_gpm, 2)} GPM · ${fmt(design.blower.total_static_pressure_in_wc, 1)} inWC`}
          </div>
        </div>
      </section>

      {/* 3. Hero Interactive 3D / 2D Scrubber Viewport */}
      <section className="card" style={{ padding: 12 }}>
        <div className="card-header" style={{ marginBottom: 10 }}>
          {/* Left: 3D vs 2D Toggle */}
          <div style={{ display: "flex", gap: 6, alignItems: "center", flexWrap: "wrap" }}>
            <div className="tab-nav" style={{ padding: 2 }}>
              <button
                type="button"
                className={`tab-btn ${viewMode === "3d" ? "active" : ""}`}
                style={{ padding: "5px 11px", fontSize: 12 }}
                onClick={() => setViewMode("3d")}
              >
                3D Model
              </button>
              <button
                type="button"
                className={`tab-btn ${viewMode === "2d" ? "active" : ""}`}
                style={{ padding: "5px 11px", fontSize: 12 }}
                onClick={() => setViewMode("2d")}
              >
                2D Blueprint
              </button>
            </div>
            {viewMode === "2d" && (
              <button
                id="btn-2d-detail-mode"
                type="button"
                className="btn btn-sm btn-ghost"
                onClick={() =>
                  setDetail2DMode((m) => (m === "detail" ? "outline" : "detail"))
                }
              >
                {detail2DMode === "detail" ? "Detailed P&ID" : "Simple Outline"}
              </button>
            )}
          </div>

          {/* Center: Stage Focus Chips */}
          <div style={{ display: "flex", gap: 4, flexWrap: "wrap" }}>
            {(
              [
                ["overview", "All"],
                ["inlet", "Inlet"],
                ["quench", "1 Quench"],
                ["venturi", isLocal ? "2 Steam/Baffle" : "2 Venturi"],
                ["packed", "3 Absorber"],
                ["blower", isLocal ? "4 Chimney" : "4 Blower"],
              ] as Array<[StageKey, string]>
            ).map(([key, label]) => (
              <button
                key={key}
                type="button"
                className={`btn btn-sm ${focusStage === key ? "btn-primary" : "btn-ghost"}`}
                onClick={() => setFocusStage(key)}
              >
                {label}
              </button>
            ))}
          </div>

          {/* Right: Compact Display Options Popover */}
          <div className="popover-wrap" ref={viewOptsRef}>
            <button
              type="button"
              className="btn btn-sm"
              onClick={() => setViewOptsOpen((v) => !v)}
              aria-expanded={viewOptsOpen}
            >
              Display ▾
            </button>
            {viewOptsOpen && (
              <div className="popover-menu">
                {viewMode === "3d" ? (
                  <>
                    <div className="popover-label">3D View Layers</div>
                    <button
                      type="button"
                      className="popover-item"
                      onClick={() => setCutaway((v) => !v)}
                    >
                      <span>Cutaway Cross-Section</span>
                      <span className={`pill ${cutaway ? "ok" : ""}`}>{cutaway ? "ON" : "OFF"}</span>
                    </button>
                    <button
                      type="button"
                      className="popover-item"
                      onClick={() => setParticles((v) => !v)}
                    >
                      <span>Animated Gas Flow</span>
                      <span className={`pill ${particles ? "ok" : ""}`}>{particles ? "ON" : "OFF"}</span>
                    </button>
                    <button
                      type="button"
                      className="popover-item"
                      onClick={() => setShowDimensions((v) => !v)}
                    >
                      <span>Dimension Callouts</span>
                      <span className={`pill ${showDimensions ? "ok" : ""}`}>{showDimensions ? "ON" : "OFF"}</span>
                    </button>
                    {!isLocal && (
                      <button
                        id="btn-systems-3d"
                        type="button"
                        className="popover-item"
                        onClick={() => setShowSystems3D((v) => !v)}
                      >
                        <span>Skid, Pump &amp; Baffling</span>
                        <span className={`pill ${showSystems3D ? "ok" : ""}`}>{showSystems3D ? "ON" : "OFF"}</span>
                      </button>
                    )}
                    <div className="popover-sep" />
                    <button
                      type="button"
                      className="popover-item"
                      onClick={() => {
                        setFocusStage("overview");
                        setResetCounter((c) => c + 1);
                        setViewOptsOpen(false);
                      }}
                    >
                      <span>Reset Camera View</span>
                    </button>
                  </>
                ) : (
                  <>
                    <div className="popover-label">2D Blueprint Layers</div>
                    <button
                      type="button"
                      className="popover-item"
                      onClick={() => setShow2DPlumbing((v) => !v)}
                    >
                      <span>Plumbing &amp; Water Lines</span>
                      <span className={`pill ${show2DPlumbing ? "ok" : ""}`}>{show2DPlumbing ? "ON" : "OFF"}</span>
                    </button>
                    <button
                      type="button"
                      className="popover-item"
                      onClick={() => setShow2DControls((v) => !v)}
                    >
                      <span>Sensors &amp; Safety Interlocks</span>
                      <span className={`pill ${show2DControls ? "ok" : ""}`}>{show2DControls ? "ON" : "OFF"}</span>
                    </button>
                    <button
                      type="button"
                      className="popover-item"
                      onClick={() => setShowDimensions((v) => !v)}
                    >
                      <span>Dimensions</span>
                      <span className={`pill ${showDimensions ? "ok" : ""}`}>{showDimensions ? "ON" : "OFF"}</span>
                    </button>
                    <div className="popover-sep" />
                    <button
                      type="button"
                      className="popover-item"
                      onClick={() => {
                        setViewOptsOpen(false);
                        if (!svg2DRef.current) return;
                        const xml = new XMLSerializer().serializeToString(svg2DRef.current);
                        triggerBlobDownload(
                          new Blob([xml], { type: "image/svg+xml;charset=utf-8" }),
                          `kilnscrub-${project.preset_id || "studio"}-2d-${detail2DMode}.svg`
                        );
                      }}
                    >
                      <span>Save 2D Blueprint (.svg)</span>
                    </button>
                  </>
                )}
              </div>
            )}
          </div>
        </div>

        <div
          className="viewport-shell"
          style={viewMode === "2d" && detail2DMode === "detail" ? { height: "auto", minHeight: 520 } : undefined}
        >
          {viewMode === "3d" && (
            <div className="viewport-hud">
              <span className="hud-chip" id="hud-renderer">
                {scaleSummary} · Drag to orbit · Scroll to zoom
              </span>
            </div>
          )}

          {viewMode === "3d" && particles && (
            <div className="viewport-legend" aria-label="Flow Particle Legend">
              <span><i className="legend-dot" style={{ background: "#fb923c" }} /> Hot Flue</span>
              <span><i className="legend-dot" style={{ background: "#38bdf8" }} /> Quenched Mist</span>
              <span><i className="legend-dot" style={{ background: "#f8fafc" }} /> Steam / Jet</span>
              <span><i className="legend-dot" style={{ background: "#34d399" }} /> Clean Exhaust</span>
            </div>
          )}

          {viewMode === "3d" ? (
            <Viewport3D
              evaluation={evaluation}
              cutaway={cutaway}
              particles={particles}
              showSystemsOverlay={showSystems3D}
              showDimensions={showDimensions}
              units={units}
              focusStage={focusStage}
              resetCounter={resetCounter}
              active={active && viewMode === "3d"}
              onScaleSummary={setScaleSummary}
              onWebGLUnavailable={() => setViewMode("2d")}
            />
          ) : (
            <Schematic2D
              evaluation={evaluation}
              units={units}
              highlightStage={focusStage}
              detailMode={detail2DMode}
              showPlumbing={show2DPlumbing}
              showControls={show2DControls}
              showDimensions={showDimensions}
              svgRef={svg2DRef}
            />
          )}
        </div>

        {/* Plain-English Contextual Stage Explainer */}
        <div className="stage-caption">
          <span>{stageExplanation}</span>
          {focusStage !== "overview" && (
            <button
              type="button"
              className="btn btn-sm btn-ghost"
              onClick={() => setFocusStage("overview")}
            >
              Show All Stages
            </button>
          )}
        </div>
      </section>

      {/* 4. Organized Deep-Dive Tabs (Site Readiness, Charts, Safety Notices) */}
      <section className="card" ref={detailsRef}>
        <div className="subtab-bar" role="tablist" aria-label="Engineering Details">
          <button
            type="button"
            role="tab"
            aria-selected={detailTab === "overview"}
            className={`subtab-btn ${detailTab === "overview" ? "active" : ""}`}
            onClick={() => setDetailTab("overview")}
          >
            Site Checks &amp; Mode Summary
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={detailTab === "charts"}
            className={`subtab-btn ${detailTab === "charts" ? "active" : ""}`}
            onClick={() => setDetailTab("charts")}
          >
            Performance Charts
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={detailTab === "notices"}
            className={`subtab-btn ${detailTab === "notices" ? "active" : ""}`}
            onClick={() => setDetailTab("notices")}
          >
            Safety &amp; Engineering Notices
            {allNotices.length > 0 && (
              <span className="pill warn" style={{ padding: "1px 7px" }}>
                {allNotices.length}
              </span>
            )}
          </button>
        </div>

        {detailTab === "overview" && (
          <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
            {isDeluxe && plan.deluxe_spec && (
              <div id="deluxe-spec-banner" className="banner banner-deluxe">
                <div className="card-header" style={{ marginBottom: 4 }}>
                  <strong>Pro Deluxe Package</strong>
                  <span className="pill info">
                    +{fmtUSD(plan.deluxe_spec.cost_delta_usd)} vs No-Weld DIY
                  </span>
                </div>
                <div style={{ fontSize: 12.5, color: "#e2e8f0" }}>
                  {plan.deluxe_spec.summary}
                </div>
                {(plan.deluxe_spec.eco_cost_delta_usd ?? 0) > 0 && (
                  <div style={{ display: "flex", gap: 8, flexWrap: "wrap", marginTop: 8 }}>
                    <span className="pill ok">
                      Water Saved: {fmt(plan.deluxe_spec.water_saved_pct, 0)}% ({fmt(plan.deluxe_spec.water_saved_gal_per_firing, 0)} gal)
                    </span>
                    <span className="pill ok">
                      Energy Saved: {fmt(plan.deluxe_spec.energy_saved_pct, 0)}% ({fmt(plan.deluxe_spec.energy_saved_kwh_per_firing, 1)} kWh)
                    </span>
                    <span className="pill info">
                      Waste Heat Reclaimed: {fmt(plan.deluxe_spec.recovered_heat_kw, 1)} kW
                    </span>
                  </div>
                )}
              </div>
            )}

            {isLocal && frugal && (
              <div id="frugal-spec-banner" className="banner banner-frugal">
                <div className="card-header" style={{ marginBottom: 6 }}>
                  <strong>
                    {frugal.region_label} · {frugal.draft_mode_label}
                  </strong>
                  <span className="pill ok">
                    {frugal.local_currency_estimate} · {fmt(frugal.power_watts, 0)} W
                  </span>
                </div>
                <div style={{ fontSize: 12.5, color: "#fef3c7", marginBottom: 8 }}>
                  {frugal.summary}
                </div>
                <div style={{ display: "flex", gap: 6, flexWrap: "wrap" }}>
                  {frugal.elevation_note && (
                    <span className="pill info">Altitude: {frugal.elevation_note}</span>
                  )}
                  {frugal.afterburner_enabled && (
                    <span className="pill ok">Stage 0 Tar Afterburner Active (0 W)</span>
                  )}
                  {(frugal.so2_inlet_ppm ?? 0) > 0 && (
                    <span className="pill warn">
                      SO₂ {fmt(frugal.so2_inlet_ppm, 0)} ppm → {fmt(frugal.so2_removal_pct, 1)}% scrubbed (H₂SO₄ dew pt ~{fmt(frugal.sulfur_acid_dew_point_f, 0)}°F)
                    </span>
                  )}
                  {frugal.steam_enabled && (
                    <span className="pill info">
                      {frugal.boiler_source_label || "Dedicated Firebox"} · {frugal.pump_mode_label || frugal.manual_pump_type}
                    </span>
                  )}
                </div>
              </div>
            )}

            <div className="grid-3">
              {siteChecks.map((chk) => (
                <div key={chk.title} className="site-check-item">
                  <div className="control-row">
                    <strong style={{ fontSize: 12.5 }}>{chk.title}</strong>
                    <span className={`pill ${chk.status}`}>{chk.status.replace("_", " ")}</span>
                  </div>
                  <div style={{ fontSize: 13, fontWeight: 700, color: "#f8fafc" }}>
                    {chk.headline}
                  </div>
                  <div style={{ fontSize: 11.5, color: "var(--text-muted)", lineHeight: 1.42 }}>
                    {chk.detail}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {detailTab === "charts" && <EngineeringCharts design={design} />}

        {detailTab === "notices" && (
          <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
            {allNotices.length === 0 ? (
              <div className="banner banner-ok" style={{ fontSize: 13 }}>
                No engineering or safety warnings for this configuration.
              </div>
            ) : (
              allNotices.map((w, i) => (
                <div key={i} className="notice-card">
                  {w}
                </div>
              ))
            )}
          </div>
        )}
      </section>
    </div>
  );
}
