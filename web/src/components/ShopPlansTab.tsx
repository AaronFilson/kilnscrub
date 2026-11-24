"use client";

import React, { useRef, useState } from "react";
import type { StudioEvaluation } from "@/types/kilnscrub";
import { fmt, fmtUSD, triggerBlobDownload } from "@/lib/api";

interface ShopPlansTabProps {
  evaluation: StudioEvaluation;
  units: "in" | "mm";
}

type PlanLayer = "electrical" | "plumbing" | "bracing" | "clearances" | "acoustics";
type SchedTab = "electrical" | "plumbing" | "bracing" | "acoustics" | "parts" | "cuts";

export function ShopPlansTab({ evaluation, units }: ShopPlansTabProps) {
  const plan = evaluation.shop_plan;
  const svgRef = useRef<SVGSVGElement | null>(null);
  const [selectedPart, setSelectedPart] = useState<number | null>(null);
  const [schedTab, setSchedTab] = useState<SchedTab>("parts");
  const [layers, setLayers] = useState<Record<PlanLayer, boolean>>({
    electrical: true,
    plumbing: true,
    bracing: true,
    clearances: true,
    acoustics: true,
  });

  const toggleLayer = (l: PlanLayer) =>
    setLayers((prev) => ({ ...prev, [l]: !prev[l] }));

  const handleDownloadSVG = () => {
    if (!svgRef.current) return;
    const xml = new XMLSerializer().serializeToString(svgRef.current);
    triggerBlobDownload(
      new Blob([xml], { type: "image/svg+xml;charset=utf-8" }),
      `kilnscrub-${evaluation.project.preset_id || "studio"}-shop-plan.svg`
    );
  };

  const dim = (inches: number) =>
    units === "in" ? `${inches.toFixed(1)}"` : `${Math.round(inches * 25.4)} mm`;

  const isLocal = plan.build_path === "local_ceramic_masonry";
  const isSteam = !!plan.steam_ejector_enabled;
  const isSumpPoolOnly = isLocal && evaluation.frugal?.has_overhead_tank === false;

  return (
    <div className="page-container">
      {/* Toolbar & 5 Layer Toggles */}
      <div className="card no-print">
        <div className="card-header" style={{ marginBottom: 8 }}>
          <div>
            <h2 style={{ margin: 0, fontSize: 16 }}>{plan.title}</h2>
            <div style={{ fontSize: 12, color: "var(--text-muted)" }}>
              Overall Envelope: {dim(plan.overall_length_in)} L × {dim(plan.overall_width_in)} W ×{" "}
              {dim(plan.overall_height_in)} H · {plan.parts.length} Parts · {plan.cut_list.length} Cut Items
            </div>
          </div>
          <div style={{ display: "flex", gap: 6, flexWrap: "wrap" }}>
            <button type="button" className="btn btn-sm" onClick={handleDownloadSVG}>
              Download SVG
            </button>
            <button type="button" className="btn btn-sm btn-primary" onClick={() => window.print()}>
              Print Shop Plan
            </button>
          </div>
        </div>

        <div id="plan-overlay-toolbar" style={{ display: "flex", gap: 6, flexWrap: "wrap" }}>
          {(
            [
              ["electrical", "Electrical & Wiring"],
              ["plumbing", "Pumps & Hosing"],
              ["bracing", "Bracing & Mounting"],
              ["clearances", "Clearances & Codes"],
              ["acoustics", "Sound Baffling"],
            ] as const
          ).map(([key, label]) => (
            <button
              key={key}
              id={`btn-layer-${key}`}
              data-plan-layer={key}
              type="button"
              className={`btn btn-sm ${layers[key] ? "btn-accent" : ""}`}
              onClick={() => toggleLayer(key)}
            >
              {label}: {layers[key] ? "ON" : "OFF"}
            </button>
          ))}
        </div>
      </div>

      {/* Architectural Blueprint SVG (Elevation + System Overlays) */}
      <div className="card" style={{ padding: 10 }}>
        <svg
          ref={svgRef}
          viewBox="0 0 1100 420"
          width="100%"
          style={{ display: "block", background: "#091627", borderRadius: 6 }}
        >
          <defs>
            <pattern id="bpGrid" width="22" height="22" patternUnits="userSpaceOnUse">
              <path d="M 22 0 L 0 0 0 22" fill="none" stroke="#1d3c62" strokeWidth="0.5" opacity="0.45" />
            </pattern>
          </defs>
          <rect width="1100" height="420" fill="url(#bpGrid)" />
          <rect x="12" y="12" width="1076" height="396" fill="none" stroke="#38bdf8" strokeWidth="1.2" />

          {/* Title Stamp */}
          <text x="26" y="34" fill="#f8fafc" fontSize="13" fontWeight="700" fontFamily="monospace">
            {plan.title.toUpperCase()}
          </text>
          <text x="26" y="50" fill="#7dd3fc" fontSize="10" fontFamily="monospace">
            ENVELOPE: {dim(plan.overall_length_in)} L × {dim(plan.overall_width_in)} W × {dim(plan.overall_height_in)} H · BUILD PATH: {plan.build_path}
          </text>

          {/* Clearances Layer */}
          {layers.clearances && (
            <g opacity="0.9">
              <rect
                x="28"
                y="70"
                width="1044"
                height="28"
                rx="3"
                fill="rgba(245, 158, 11, 0.12)"
                stroke="#f59e0b"
                strokeDasharray="5 3"
              />
              <text x="42" y="88" fill="#fde68a" fontSize="10" fontFamily="monospace">
                CODE CLEARANCES: Flue→Combustibles ≥ {dim(plan.clearances.flue_to_combustibles_in)} · Property Line ≥{" "}
                {plan.clearances.property_line_min_ft} ft · Service Aisle ≥ {dim(plan.clearances.service_aisle_in)} · NEC Panel:{" "}
                {plan.clearances.nec_panel_depth_in > 0
                  ? `${dim(plan.clearances.nec_panel_depth_in)} × ${dim(plan.clearances.nec_panel_width_in)}`
                  : "0V OFF-GRID"}
              </text>
            </g>
          )}

          {/* Main Equipment Boxes with clickable [N] callouts */}
          {[
            { num: 1, x: 42, y: 155, w: 150, h: 120, color: "#f97316", tag: isLocal ? "REFRACTORY FLUE" : "INLET + 200°F BYPASS" },
            { num: 2, x: 226, y: 135, w: 165, h: 155, color: "#fb923c", tag: isLocal ? "MASONRY QUENCH" : "STAGE 1 QUENCH" },
            { num: 3, x: 424, y: 155, w: 165, h: 120, color: "#eab308", tag: isSumpPoolOnly ? "0.5\" S-VANE + WEIR" : isSteam ? "STEAM-JET EJECTOR" : "STAGE 2 VENTURI" },
            { num: 4, x: 622, y: 125, w: 185, h: 175, color: "#10b981", tag: isSumpPoolOnly ? "TROUGH + 2-STG SIEVE" : isLocal ? "WIDE STONEWARE BED" : "STAGE 3 ABSORBER" },
            { num: 6, x: 842, y: 130, w: 215, h: 165, color: "#38bdf8", tag: isSumpPoolOnly ? "SUMP POOLS + CHIMNEY" : isLocal ? "GRAVITY TANK + CHIMNEY" : "STAGE 4 BLOWER + STACK" },
          ].map((bx) => {
            const pt = plan.parts.find((p) => p.num === bx.num);
            const active = selectedPart === bx.num;
            return (
              <g
                key={bx.num}
                style={{ cursor: "pointer" }}
                onClick={() => setSelectedPart(active ? null : bx.num)}
              >
                <rect
                  x={bx.x}
                  y={bx.y}
                  width={bx.w}
                  height={bx.h}
                  rx="5"
                  fill={active ? "rgba(56, 189, 248, 0.22)" : "#10233d"}
                  stroke={active ? "#ffffff" : bx.color}
                  strokeWidth={active ? 2.6 : 1.8}
                />
                <circle cx={bx.x + 18} cy={bx.y + 18} r="11" fill={bx.color} />
                <text x={bx.x + 18} y={bx.y + 22} textAnchor="middle" fill="#081018" fontSize="10.5" fontWeight="800" fontFamily="monospace">
                  {bx.num}
                </text>
                <text x={bx.x + 36} y={bx.y + 22} fill={bx.color} fontSize="10" fontWeight="700" fontFamily="monospace">
                  {bx.tag}
                </text>
                <text x={bx.x + 10} y={bx.y + 52} fill="#f8fafc" fontSize="10.5" fontWeight="600">
                  {(pt?.label || "").slice(0, 24)}
                </text>
                <text x={bx.x + 10} y={bx.y + 72} fill="#93c5fd" fontSize="9.5" fontFamily="monospace">
                  {(units === "in" ? pt?.dimensions_in : pt?.dimensions_mm)?.slice(0, 26) || ""}
                </text>
                <text x={bx.x + 10} y={bx.y + 92} fill="#86efac" fontSize="9.5" fontFamily="monospace">
                  {pt ? fmtUSD(pt.cost_usd) : ""}
                </text>
              </g>
            );
          })}

          {/* Plumbing Overlay */}
          {layers.plumbing && (
            <g>
              <line x1="250" y1="316" x2="780" y2="316" stroke="#38bdf8" strokeWidth="2.5" strokeDasharray="6 3" />
              <text x="515" y="334" textAnchor="middle" fill="#7dd3fc" fontSize="10" fontFamily="monospace">
                PLUMBING: {plan.pump_station.pump_type} ({fmt(plan.pump_station.flow_gpm, 2)} GPM @ {fmt(plan.pump_station.head_ft, 1)} ft) · {plan.plumbing.length} Lines
              </text>
            </g>
          )}

          {/* Electrical Overlay */}
          {layers.electrical && (
            <g>
              <rect x="42" y="348" width="335" height="44" rx="4" fill="rgba(239, 68, 68, 0.12)" stroke="#f87171" />
              <text x="54" y="366" fill="#fca5a5" fontSize="10" fontWeight="700" fontFamily="monospace">
                ELECTRICAL ({plan.electrical.length} CIRCUITS)
              </text>
              <text x="54" y="382" fill="#e2e8f0" fontSize="9.2" fontFamily="monospace">
                {plan.electrical.length === 0
                  ? "0V / OFF-GRID — ZERO ELECTRICAL CIRCUITS"
                  : `${plan.electrical[0].label.slice(0, 44)} (${plan.electrical[0].voltage}V)`}
              </text>
            </g>
          )}

          {/* Bracing Overlay */}
          {layers.bracing && (
            <g>
              <rect x="392" y="348" width="330" height="44" rx="4" fill="rgba(16, 185, 129, 0.12)" stroke="#34d399" />
              <text x="404" y="366" fill="#6ee7b7" fontSize="10" fontWeight="700" fontFamily="monospace">
                STRUCTURAL BRACING ({plan.bracing.length} MEMBERS)
              </text>
              <text x="404" y="382" fill="#e2e8f0" fontSize="9.2" fontFamily="monospace">
                {plan.bracing[0]
                  ? `${plan.bracing[0].component.slice(0, 34)} (${Math.round(plan.bracing[0].load_lbs)} lbs)`
                  : "Skid / Pad Anchored"}
              </text>
            </g>
          )}

          {/* Acoustics Overlay */}
          {layers.acoustics && (
            <g>
              <rect x="738" y="348" width="320" height="44" rx="4" fill="rgba(168, 85, 247, 0.14)" stroke="#c084fc" />
              <text x="750" y="366" fill="#e9d5ff" fontSize="10" fontWeight="700" fontFamily="monospace">
                ACOUSTICS ({plan.acoustics.length} ZONES · -{fmt(plan.acoustic_summary.total_reduction_dba, 1)} dBA)
              </text>
              <text x="750" y="382" fill="#e2e8f0" fontSize="9.2" fontFamily="monospace">
                {fmt(plan.acoustic_summary.bare_dba_3ft, 1)} → {fmt(plan.acoustic_summary.treated_dba_3ft, 1)} dBA @ 3ft · {fmt(plan.acoustic_summary.property_line_dba, 1)} dBA @ {plan.acoustic_summary.property_line_ft}ft
              </text>
            </g>
          )}
        </svg>
      </div>

      {/* 6-Tab Interactive Engineering Schedules */}
      <div id="plan-schedules-panel" className="card">
        <div className="card-header">
          <div style={{ display: "flex", gap: 6, flexWrap: "wrap" }}>
            {(
              [
                ["parts", `Parts Legend (${plan.parts.length})`],
                ["cuts", `Cut Sheets (${plan.cut_list.length})`],
                ["electrical", `Electrical (${plan.electrical.length})`],
                ["plumbing", `Plumbing (${plan.plumbing.length})`],
                ["bracing", `Bracing (${plan.bracing.length})`],
                ["acoustics", `Acoustics (${plan.acoustics.length})`],
              ] as const
            ).map(([k, label]) => (
              <button
                key={k}
                type="button"
                className={`btn btn-sm ${schedTab === k ? "btn-primary" : ""}`}
                onClick={() => setSchedTab(k)}
              >
                {label}
              </button>
            ))}
          </div>
        </div>

        <div className="table-wrap">
          {schedTab === "parts" && (
            <table className="data-table">
              <thead>
                <tr>
                  <th>#</th>
                  <th>Part / Stage</th>
                  <th>Material</th>
                  <th>Dimensions ({units})</th>
                  <th>SKU</th>
                  <th className="num">Est. USD</th>
                </tr>
              </thead>
              <tbody>
                {plan.parts.map((pt) => (
                  <tr
                    key={pt.num}
                    className={selectedPart === pt.num ? "row-highlight" : ""}
                    onClick={() => setSelectedPart(pt.num)}
                    style={{ cursor: "pointer" }}
                  >
                    <td className="unit-literal">[{pt.num}]</td>
                    <td>
                      <div style={{ fontWeight: 600 }}>{pt.label}</div>
                      <div style={{ fontSize: 11, color: "var(--text-muted)" }}>{pt.notes}</div>
                    </td>
                    <td>{pt.material}</td>
                    <td className="unit-literal">
                      {units === "in" ? pt.dimensions_in : pt.dimensions_mm}
                    </td>
                    <td className="unit-literal">{pt.sku}</td>
                    <td className="num">{fmtUSD(pt.cost_usd)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}

          {schedTab === "cuts" && (
            <table className="data-table">
              <thead>
                <tr>
                  <th>Part #</th>
                  <th>Component</th>
                  <th>Shape</th>
                  <th>Flat Pattern ({units})</th>
                  <th className="num">Qty</th>
                  <th>Fabrication Notes</th>
                </tr>
              </thead>
              <tbody>
                {plan.cut_list.map((c, idx) => (
                  <tr key={idx}>
                    <td className="unit-literal">[{c.part_num}]</td>
                    <td style={{ fontWeight: 600 }}>{c.name}</td>
                    <td className="unit-literal">{c.shape}</td>
                    <td className="unit-literal">
                      {c.sweep_deg
                        ? `R_out=${dim(c.outer_radius_in || 0)}, R_in=${dim(c.inner_radius_in || 0)}, θ=${fmt(c.sweep_deg, 1)}°`
                        : `${dim(c.flat_width_in)} × ${dim(c.flat_height_in)}`}
                    </td>
                    <td className="num">{c.qty}</td>
                    <td style={{ fontSize: 11.5, color: "var(--text-muted)" }}>{c.notes}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}

          {schedTab === "electrical" && (
            <table className="data-table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Circuit</th>
                  <th>From → To</th>
                  <th>Rating</th>
                  <th>Wire / Conduit</th>
                  <th>Code Ref</th>
                </tr>
              </thead>
              <tbody>
                {plan.electrical.length === 0 ? (
                  <tr>
                    <td colSpan={6} style={{ color: "#6ee7b7", fontFamily: "monospace" }}>
                      0V / OFF-GRID — ZERO ELECTRICAL CIRCUITS REQUIRED
                    </td>
                  </tr>
                ) : (
                  plan.electrical.map((e) => (
                    <tr key={e.id}>
                      <td className="unit-literal">{e.id}</td>
                      <td>
                        <div style={{ fontWeight: 600 }}>{e.label}</div>
                        <div style={{ fontSize: 11, color: "var(--text-muted)" }}>{e.notes}</div>
                      </td>
                      <td>
                        {e.from} → {e.to}
                      </td>
                      <td className="unit-literal">
                        {e.voltage}V · {fmt(e.amps, 1)}A ({e.breaker_amps}A Bkr)
                      </td>
                      <td>
                        {e.wire_gauge} / {e.conduit}
                      </td>
                      <td className="unit-literal">{e.nec_ref}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          )}

          {schedTab === "plumbing" && (
            <table className="data-table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Line / Fluid</th>
                  <th>Size &amp; Material</th>
                  <th className="num">Flow</th>
                  <th className="num">PSI</th>
                  <th>Valves &amp; Notes</th>
                </tr>
              </thead>
              <tbody>
                {plan.plumbing.map((pl) => (
                  <tr key={pl.id}>
                    <td className="unit-literal">{pl.id}</td>
                    <td>
                      <div style={{ fontWeight: 600 }}>{pl.label}</div>
                      <div style={{ fontSize: 11, color: "var(--text-muted)" }}>{pl.fluid}</div>
                    </td>
                    <td>
                      {dim(pl.size_in)} · {pl.material}
                    </td>
                    <td className="num">
                      {pl.flow_gpm > 0 ? `${fmt(pl.flow_gpm, 2)} GPM` : "Static"}
                    </td>
                    <td className="num">{fmt(pl.pressure_psi, 1)}</td>
                    <td style={{ fontSize: 11.5, color: "var(--text-muted)" }}>
                      {pl.valves_fittings} — {pl.notes}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}

          {schedTab === "bracing" && (
            <table className="data-table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Component</th>
                  <th>Member Spec</th>
                  <th className="num">Load (lbs)</th>
                  <th>Anchors &amp; Notes</th>
                </tr>
              </thead>
              <tbody>
                {plan.bracing.map((br) => (
                  <tr key={br.id}>
                    <td className="unit-literal">{br.id}</td>
                    <td style={{ fontWeight: 600 }}>{br.component}</td>
                    <td>{br.member_spec}</td>
                    <td className="num">{Math.round(br.load_lbs)}</td>
                    <td style={{ fontSize: 11.5, color: "var(--text-muted)" }}>
                      {br.anchor_spec} — {br.notes}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}

          {schedTab === "acoustics" && (
            <table className="data-table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Location</th>
                  <th>Treatment &amp; Material</th>
                  <th className="num">Δ dBA</th>
                  <th>Notes</th>
                </tr>
              </thead>
              <tbody>
                {plan.acoustics.map((ac) => (
                  <tr key={ac.id}>
                    <td className="unit-literal">{ac.id}</td>
                    <td style={{ fontWeight: 600 }}>{ac.location}</td>
                    <td>
                      {ac.treatment} ({ac.material})
                    </td>
                    <td className="num">-{fmt(ac.attenuation_dba, 1)}</td>
                    <td style={{ fontSize: 11.5, color: "var(--text-muted)" }}>{ac.notes}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>
    </div>
  );
}
