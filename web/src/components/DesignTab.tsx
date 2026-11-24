"use client";

import React, { useEffect, useRef, useState } from "react";
import type { StudioEvaluation } from "@/types/kilnscrub";
import { fmt, fmtUSD, kToF, mToIn, triggerBlobDownload } from "@/lib/api";
import { Viewport3D } from "./Viewport3D";
import type { StageKey } from "./scene3d/types";
import { Schematic2D } from "./Schematic2D";
import { EngineeringCharts } from "./EngineeringCharts";

interface DesignTabProps {
  evaluation: StudioEvaluation;
  units: "in" | "mm";
  active: boolean;
}

export function DesignTab({ evaluation, units, active }: DesignTabProps) {
  const [viewMode, setViewMode] = useState<"3d" | "2d">("3d");
  const [detail2DMode, setDetail2DMode] = useState<"detail" | "outline">("detail");
  const [show2DPlumbing, setShow2DPlumbing] = useState(true);
  const [show2DControls, setShow2DControls] = useState(true);
  const svg2DRef = useRef<SVGSVGElement | null>(null);
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
    // Auto-enable systems overlay when first entering Deluxe mode, while allowing users to toggle it OFF!
    if (isDeluxe && !prevDeluxeRef.current) {
      setShowSystems3D(true);
    }
    prevDeluxeRef.current = isDeluxe;
  }, [isDeluxe]);
  const quenchOutF = kToF(design.quench.outlet_temp_k);

  const dim = (inches: number) =>
    units === "in" ? `${inches.toFixed(1)}"` : `${Math.round(inches * 25.4)} mm`;

  return (
    <div className="main-stage">
      {/* 1. Urban Neighbor Readiness Bar */}
      <div className="readiness-bar" aria-label="Urban Neighbor Readiness">
        <span className={`pill ${urban.overall_status}`}>
          READINESS: {urban.overall_status.toUpperCase()}
        </span>
        <span className={`pill ${urban.electrical.status}`} title={urban.electrical.summary}>
          Elec: {urban.electrical.circuit_label}
        </span>
        <span className={`pill ${urban.water.status}`} title={urban.water.summary}>
          Water: {fmt(urban.water.required_gpm, 2)} GPM
        </span>
        <span className={`pill ${urban.noise.status}`} title={urban.noise.summary}>
          Noise: {fmt(urban.noise.property_line_dba, 1)} dBA @ {urban.noise.property_line_ft}ft
        </span>
        <span className={`pill ${urban.plume.status}`} title={urban.plume.summary}>
          Plume: ~{fmt(urban.plume.bare_plume_length_ft, 0)}ft
          {project.include_plume_diluter && !project.zero_electricity
            ? ` → ${fmt(urban.plume.diluted_plume_length_ft, 0)}ft`
            : ""}
        </span>
        <span className={`pill ${urban.drain.status}`} title={urban.drain.summary}>
          Alkali: {urban.drain.bulk_packaging_label.slice(0, 34)}
        </span>
        <span className={`pill ${urban.corrosion.status}`} title={urban.corrosion.material_summary}>
          {urban.corrosion.tier_label.slice(0, 26)}
        </span>
      </div>

      {/* 2. Conditional Deluxe or Local Ceramic / Off-Grid Banner */}
      {isDeluxe && plan.deluxe_spec && (
        <div id="deluxe-spec-banner" className="banner banner-deluxe">
          <div className="card-header" style={{ marginBottom: 4 }}>
            <strong>Custom Fabrication Deluxe Package Active</strong>
            <span className="pill info">
              +{fmtUSD(plan.deluxe_spec.cost_delta_usd)} vs No-Weld COTS
            </span>
          </div>
          <div style={{ fontSize: 12.5, color: "#e2e8f0" }}>
            {plan.deluxe_spec.summary}
          </div>
          {(plan.deluxe_spec.eco_cost_delta_usd ?? 0) > 0 && (
            <div className="grid-3" style={{ marginTop: 8 }}>
              <div className="card" style={{ padding: 8 }}>
                <div className="kpi-label">Evap Water Saved</div>
                <div className="kpi-value" style={{ fontSize: 15 }}>
                  {fmt(plan.deluxe_spec.water_saved_pct, 0)}% ({fmt(plan.deluxe_spec.water_saved_gal_per_firing, 0)} gal)
                </div>
              </div>
              <div className="card" style={{ padding: 8 }}>
                <div className="kpi-label">Blower Energy Saved</div>
                <div className="kpi-value" style={{ fontSize: 15 }}>
                  {fmt(plan.deluxe_spec.energy_saved_pct, 0)}% ({fmt(plan.deluxe_spec.energy_saved_kwh_per_firing, 1)} kWh)
                </div>
              </div>
              <div className="card" style={{ padding: 8 }}>
                <div className="kpi-label">Studio Waste Heat Reclaimed</div>
                <div className="kpi-value" style={{ fontSize: 15 }}>
                  {fmt(plan.deluxe_spec.recovered_heat_kw, 1)} kW
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {isLocal && frugal && (
        <div id="frugal-spec-banner" className="banner banner-frugal">
          <div className="card-header" style={{ marginBottom: 6 }}>
            <strong>
              Local Ceramic &amp; Off-Grid Mode ({frugal.region_label} · {frugal.draft_mode_label})
            </strong>
            <span className="pill ok">
              {frugal.local_currency_estimate} · {fmt(frugal.power_watts, 0)} W
            </span>
          </div>
          <div style={{ fontSize: 12.5, color: "#fef3c7", marginBottom: 6 }}>
            {frugal.summary}
          </div>
          {frugal.steam_enabled && (
            <div style={{ display: "flex", gap: 6, flexWrap: "wrap", marginBottom: 8 }}>
              <span className="pill info">
                Boiler: {frugal.boiler_source_label || "Dedicated Firebox"}
                {(frugal.boiler_water_saved_pct ?? 0) > 0
                  ? ` (-${fmt(frugal.boiler_water_saved_pct, 0)}% Quench Water)`
                  : ""}
              </span>
              <span className={`pill ${frugal.has_overhead_tank === false ? "ok" : "info"}`}>
                Water Drive: {frugal.pump_mode_label || frugal.manual_pump_type}
              </span>
              {frugal.has_overhead_tank === false && (
                <span className="pill ok">0 Pumps · 0 Water Tower (Ground Sump Pools Only)</span>
              )}
            </div>
          )}
          <div className="grid-3">
            <div className="card" style={{ padding: 8 }}>
              <div className="kpi-label">HCl Acid Gas Capture</div>
              <div className="kpi-value" style={{ fontSize: 16 }}>
                {fmt(frugal.hcl_removal_pct, 1)}%
              </div>
            </div>
            <div className="card" style={{ padding: 8 }}>
              <div className="kpi-label">Coarse Wood Fly Ash</div>
              <div className="kpi-value" style={{ fontSize: 16 }}>
                {fmt(frugal.coarse_ash_capture_pct, 1)}%
              </div>
            </div>
            <div className="card" style={{ padding: 8 }}>
              <div className="kpi-label">Submicron NaCl Fume</div>
              <div className="kpi-value" style={{ fontSize: 16 }}>
                {fmt(frugal.submicron_nacl_capture_pct, 0)}%
              </div>
            </div>
          </div>
        </div>
      )}

      {/* 3. KPI Strip */}
      <section className="kpi-grid" aria-label="Key Engineering KPIs">
        <div className="kpi-card" style={{ ["--kpi-accent" as string]: "var(--accent-quench)" }}>
          <div className="kpi-label">
            Quench Outlet Temp
            <span
              title="CPVC 200°F Cutoff: Schedule 80 CPVC is rated to 200°F max continuous. Stage 1 spray cools exhaust below 190°F, and both the hardwired snap-disc and software interlock trip at 200°F."
              style={{ cursor: "help" }}
            >
              ⓘ
            </span>
          </div>
          <div className="kpi-value">{fmt(quenchOutF, 1)} °F</div>
          <div className="kpi-sub">
            {isLocal && frugal ? (
              frugal.has_overhead_tank === false ? (
                <>Co-Current Flooded Ejector · 0 Tower</>
              ) : (
                <>
                  Gravity Rose: {fmt(frugal.spray_flow_gpm, 2)}{" "}
                  <span className="unit-literal">GPM</span> ({fmt(frugal.spray_flow_lpm, 1)} L/min)
                </>
              )
            ) : (
              <>
                Spray: {fmt(design.quench.total_spray_water_gpm, 2)}{" "}
                <span className="unit-literal">GPM</span> · Evap{" "}
                {fmt(design.quench.evaporated_water_gpm, 2)}
              </>
            )}
          </div>
          <span className={`pill ${quenchOutF < 200 ? "ok" : "danger"}`}>
            {isLocal ? "All-Ceramic >1800°F" : "CPVC 200°F Cutoff"}
          </span>
        </div>

        <div className="kpi-card" style={{ ["--kpi-accent" as string]: "var(--accent-venturi)" }}>
          {isLocal && frugal ? (
            <>
              <div className="kpi-label">
                Submicron <span className="unit-literal">NaCl</span> Capture
              </div>
              <div className="kpi-value">
                {fmt(frugal.submicron_nacl_capture_pct, 0)}%
              </div>
              <div className="kpi-sub">
                Ash {fmt(frugal.coarse_ash_capture_pct, 1)}% · Margin{" "}
                {frugal.draft_margin_inwc >= 0 ? "+" : ""}
                {fmt(frugal.draft_margin_inwc, 2)}{" "}
                <span className="unit-literal">inWC</span>
              </div>
              <span className={`pill ${frugal.draft_feasible ? "ok" : "danger"}`}>
                {frugal.steam_enabled
                  ? `${fmt(frugal.steam_pressure_psig, 0)} psig Steam-Jet`
                  : "Passive Buoyancy"}
              </span>
            </>
          ) : (
            <>
              <div className="kpi-label">
                Venturi Cut <span className="unit-literal">d₅₀</span> (<span className="unit-literal">NaCl</span>)
              </div>
              <div className="kpi-value">
                {fmt(design.venturi.cut_diameter_d50_um, 2)} <span className="unit-literal">µm</span>
              </div>
              <div className="kpi-sub">
                <span className="unit-literal">NaCl</span>{" "}
                {fmt(design.venturi.overall_nacl_fume_efficiency * 100, 1)}% · Ash{" "}
                {fmt(design.venturi.overall_wood_ash_efficiency * 100, 1)}%
              </div>
              <span className="pill info">
                ΔP {fmt(design.venturi.pressure_drop_in_wc, 1)} <span className="unit-literal">inWC</span>
              </span>
            </>
          )}
        </div>

        <div className="kpi-card" style={{ ["--kpi-accent" as string]: "var(--accent-packed)" }}>
          <div className="kpi-label">
            <span className="unit-literal">HCl</span> Removal
          </div>
          <div className="kpi-value">
            {fmt(
              isLocal && frugal
                ? frugal.hcl_removal_pct
                : design.packed_bed.hcl_removal_efficiency * 100,
              1
            )}
            %
          </div>
          <div className="kpi-sub">
            Flood: {fmt(cots.actual_percent_of_flood, 1)}% (≤75%) ·{" "}
            {fmt(design.packed_bed.hcl_outlet_ppm, 1)} ppm out
          </div>
          <span className={`pill ${cots.actual_percent_of_flood <= 75 ? "ok" : "danger"}`}>
            Sump <span className="unit-literal">pH</span> 7.5–8.5
          </span>
        </div>

        <div className="kpi-card" style={{ ["--kpi-accent" as string]: "var(--accent-blower)" }}>
          <div className="kpi-label">
            {isLocal ? "Chimney / Ejector Draft" : "Stage 4 Blower Duty"}
          </div>
          <div className="kpi-value">
            {isLocal && frugal
              ? `${fmt(frugal.available_draft_inwc, 3)}"`
              : `${fmt(design.blower.total_static_pressure_in_wc, 1)} inWC`}
          </div>
          <div className="kpi-sub">
            {isLocal && frugal
              ? `Req ${fmt(frugal.required_draft_inwc, 3)}" · Chimney ${fmt(frugal.chimney_height_m, 1)} m`
              : `${fmt(design.blower.flow_acfm, 0)} ACFM · ${design.blower.recommended_motor_hp} HP · ${dim(mToIn(design.blower.wheel_diameter_m))} wheel`}
          </div>
          <span className="pill info">
            Draft {fmt(design.input.target_kiln_draft_in_wc, 3)} <span className="unit-literal">inWC</span>
          </span>
        </div>

        <div className="kpi-card" style={{ ["--kpi-accent" as string]: "#a855f7" }}>
          <div className="kpi-label">Estimated Hardware Total</div>
          <div className="kpi-value">{fmtUSD(cots.estimated_total_usd)}</div>
          <div className="kpi-sub">
            {isLocal && frugal
              ? frugal.local_currency_estimate
              : `Delta ${evaluation.architecture_options.cost_delta_usd >= 0 ? "+" : ""}${fmtUSD(evaluation.architecture_options.cost_delta_usd)} vs baseline`}
          </div>
          <span className="pill ok">
            {fmt(evaluation.recirculation.water_saved_pct, 0)}% Water Saved
          </span>
        </div>
      </section>

      {/* 4. Interactive 3D WebGL (Offline-Bundled Three.js) / 2D Cutaway Viewport */}
      <section className="card">
        <div className="card-header">
          <div style={{ display: "flex", gap: 6, flexWrap: "wrap", alignItems: "center" }}>
            <button
              type="button"
              className={`btn btn-sm ${viewMode === "3d" ? "btn-primary" : ""}`}
              onClick={() => setViewMode("3d")}
            >
              3D Studio (Offline Three.js)
            </button>
            <button
              type="button"
              className={`btn btn-sm ${viewMode === "2d" ? "btn-primary" : ""}`}
              onClick={() => setViewMode("2d")}
            >
              2D Cutaway &amp; P&amp;ID
            </button>
            {viewMode === "2d" && (
              <>
                <button
                  id="btn-2d-detail-mode"
                  type="button"
                  className={`btn btn-sm ${detail2DMode === "detail" ? "btn-accent" : ""}`}
                  onClick={() =>
                    setDetail2DMode((m) => (m === "detail" ? "outline" : "detail"))
                  }
                >
                  {detail2DMode === "detail" ? "Detail P&ID" : "Block Outline"}
                </button>
                {detail2DMode === "detail" && (
                  <>
                    <button
                      type="button"
                      className={`btn btn-sm ${show2DPlumbing ? "btn-accent" : ""}`}
                      onClick={() => setShow2DPlumbing((v) => !v)}
                    >
                      Plumbing: {show2DPlumbing ? "ON" : "OFF"}
                    </button>
                    <button
                      type="button"
                      className={`btn btn-sm ${show2DControls ? "btn-accent" : ""}`}
                      onClick={() => setShow2DControls((v) => !v)}
                    >
                      Controls: {show2DControls ? "ON" : "OFF"}
                    </button>
                  </>
                )}
                <button
                  type="button"
                  className="btn btn-sm"
                  onClick={() => {
                    if (!svg2DRef.current) return;
                    const xml = new XMLSerializer().serializeToString(svg2DRef.current);
                    triggerBlobDownload(
                      new Blob([xml], { type: "image/svg+xml;charset=utf-8" }),
                      `kilnscrub-${project.preset_id || "studio"}-2d-${detail2DMode}.svg`
                    );
                  }}
                >
                  Save 2D SVG
                </button>
              </>
            )}
            {viewMode === "3d" && (
              <>
                <button
                  type="button"
                  className={`btn btn-sm ${cutaway ? "btn-accent" : ""}`}
                  onClick={() => setCutaway((v) => !v)}
                >
                  Cutaway: {cutaway ? "ON" : "OFF"}
                </button>
                <button
                  type="button"
                  className={`btn btn-sm ${particles ? "btn-accent" : ""}`}
                  onClick={() => setParticles((v) => !v)}
                >
                  Flow Particles: {particles ? "ON" : "OFF"}
                </button>
                <button
                  type="button"
                  className={`btn btn-sm ${showDimensions ? "btn-accent" : ""}`}
                  onClick={() => setShowDimensions((v) => !v)}
                >
                  Dimensions: {showDimensions ? "ON" : "OFF"}
                </button>
                {!isLocal && (
                  <button
                    id="btn-systems-3d"
                    type="button"
                    className={`btn btn-sm ${showSystems3D ? "btn-accent" : ""}`}
                    onClick={() => setShowSystems3D((v) => !v)}
                  >
                    Skid, Pump &amp; Baffling: {showSystems3D ? "ON" : "OFF"}
                  </button>
                )}
                <button
                  type="button"
                  className="btn btn-sm"
                  onClick={() => {
                    setFocusStage("overview");
                    setResetCounter((c) => c + 1);
                  }}
                >
                  Reset View
                </button>
              </>
            )}
          </div>

          <div style={{ display: "flex", gap: 4, flexWrap: "wrap" }}>
            {(
              [
                ["overview", "All Stages"],
                ["inlet", "Inlet"],
                ["quench", "1: Quench"],
                ["venturi", isLocal ? "2: Steam/Baffle" : "2: Venturi"],
                ["packed", "3: Absorber"],
                ["blower", isLocal ? "4: Chimney" : "4: Blower"],
              ] as Array<[StageKey, string]>
            ).map(([key, label]) => (
              <button
                key={key}
                type="button"
                className={`btn btn-sm ${focusStage === key ? "btn-primary" : ""}`}
                onClick={() => setFocusStage(key)}
              >
                {label}
              </button>
            ))}
          </div>
        </div>

        <div className="viewport-shell" style={viewMode === "2d" && detail2DMode === "detail" ? { height: "auto", minHeight: 540 } : undefined}>
          {viewMode === "3d" && (
            <div className="viewport-hud">
              <span className="hud-chip" id="hud-renderer">
                Three.js 0.170.0 WebGL · {scaleSummary}
              </span>
              <span className="hud-chip" id="hud-controls">
                Drag to orbit · Scroll to zoom · Right-drag to pan
              </span>
            </div>
          )}

          {viewMode === "3d" && particles && (
            <div className="viewport-legend" aria-label="Flow Particle Legend">
              <span><i className="legend-dot" style={{ background: "#fb923c" }} /> Hot Kiln Gas</span>
              <span><i className="legend-dot" style={{ background: "#38bdf8" }} /> Quenched Mist / Spray</span>
              <span><i className="legend-dot" style={{ background: "#f8fafc" }} /> Steam / Motive Jet</span>
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
      </section>

      {/* 5. Engineering Charts */}
      <EngineeringCharts design={design} />

      {/* 6. Warnings (if any) */}
      {(design.warnings?.length > 0 ||
        evaluation.architecture_options.warnings?.length > 0 ||
        (frugal?.warnings?.length ?? 0) > 0) && (
        <section className="card">
          <h3 className="card-title" style={{ marginBottom: 8, color: "#fcd34d" }}>
            Engineering &amp; Safety Notices
          </h3>
          <ul style={{ margin: 0, paddingLeft: 18, fontSize: 12.5, color: "#fde68a" }}>
            {[
              ...(design.warnings || []),
              ...(evaluation.architecture_options.warnings || []),
              ...(frugal?.warnings || []),
            ].map((w, i) => (
              <li key={i} style={{ marginBottom: 4 }}>
                {w}
              </li>
            ))}
          </ul>
        </section>
      )}
    </div>
  );
}
