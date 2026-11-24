"use client";

import React from "react";
import type { StudioEvaluation } from "@/types/kilnscrub";
import { fmt, kToF, mToIn } from "@/lib/api";

interface Schematic2DProps {
  evaluation: StudioEvaluation;
  units: "in" | "mm";
  highlightStage: string | null;
}

export function Schematic2D({
  evaluation,
  units,
  highlightStage,
}: Schematic2DProps) {
  const { project, design, frugal } = evaluation;
  const isLocal = project.build_path === "local_ceramic_masonry";
  const isSteam = isLocal && !!frugal?.steam_enabled;
  const isSubsonic = isSteam && frugal?.draft_mode === "steam_ejector_subsonic";
  const isSumpPoolOnly = isSteam && frugal?.has_overhead_tank === false;
  const isEductor = !isLocal && project.stage_alignment === "eductor_dry_fan";
  const isParallel = !isLocal && project.packed_bed_layout === "parallel_manifold";
  const isDeluxe = project.build_path === "custom_fab_deluxe";

  const dim = (inches: number) =>
    units === "in" ? `${inches.toFixed(1)}"` : `${Math.round(inches * 25.4)} mm`;

  const strokeFor = (stage: string, def: string) =>
    highlightStage === stage ? "#ffffff" : def;

  return (
    <svg
      id="viewport-2d"
      viewBox="0 0 960 410"
      width="100%"
      height="100%"
      style={{ display: "block", background: "#091320" }}
    >
      <defs>
        <pattern id="grid2d" width="24" height="24" patternUnits="userSpaceOnUse">
          <path
            d="M 24 0 L 0 0 0 24"
            fill="none"
            stroke="#1b324e"
            strokeWidth="0.6"
            opacity="0.45"
          />
        </pattern>
        <marker
          id="arr2d"
          viewBox="0 0 10 10"
          refX="8"
          refY="5"
          markerWidth="6"
          markerHeight="6"
          orient="auto"
        >
          <path d="M 0 1 L 10 5 L 0 9 z" fill="#38bdf8" />
        </marker>
      </defs>

      <rect width="960" height="410" fill="url(#grid2d)" />

      {/* Ground / Skid Baseline */}
      <line x1="28" y1="352" x2="932" y2="352" stroke="#475569" strokeWidth="2" />
      <text x="36" y="370" fill="#64748b" fontSize="10.5" fontFamily="monospace">
        {isLocal
          ? `LOCAL CERAMIC MASONRY PAD (${frugal?.region_label || "Off-Grid"}) · POWER: ${fmt(frugal?.power_watts, 0)} W`
          : `SKID ENVELOPE: ${fmt(evaluation.architecture_options.footprint_length_ft, 1)} ft L × ${fmt(evaluation.architecture_options.footprint_width_ft, 1)} ft W · ${project.build_path.toUpperCase()}`}
      </text>

      {/* Inlet Kiln Flue */}
      <rect
        x="36"
        y="148"
        width="104"
        height="46"
        rx="4"
        fill="#2c170f"
        stroke={strokeFor("inlet", "#f97316")}
        strokeWidth="2.2"
      />
      <text x="88" y="168" textAnchor="middle" fill="#fdba74" fontSize="11" fontWeight="700">
        KILN FLUE
      </text>
      <text x="88" y="184" textAnchor="middle" fill="#fed7aa" fontSize="9.5" fontFamily="monospace">
        {Math.round(design.input.peak_exit_temp_f)}°F · {dim(mToIn(design.envelope.duct_diameter_m))}
      </text>

      {!isLocal && (
        <>
          {/* Fail-Safe Bypass Damper Tee */}
          <rect
            x="66"
            y="84"
            width="52"
            height="64"
            rx="3"
            fill="#3b1116"
            stroke="#ef4444"
            strokeWidth="1.6"
            strokeDasharray="4 2"
          />
          <text x="92" y="76" textAnchor="middle" fill="#fca5a5" fontSize="9.5" fontFamily="monospace">
            200°F FAIL-OPEN BYPASS
          </text>
        </>
      )}

      <line
        x1="140"
        y1="171"
        x2="188"
        y2="171"
        stroke="#f97316"
        strokeWidth="3"
        markerEnd="url(#arr2d)"
      />

      {/* Stage 1: Quench Tower */}
      <rect
        x="194"
        y="108"
        width="126"
        height="242"
        rx="6"
        fill={isLocal ? "#352015" : "#1f1b18"}
        stroke={strokeFor("quench", "#f97316")}
        strokeWidth="2.2"
      />
      <text x="257" y="130" textAnchor="middle" fill="#fb923c" fontSize="11" fontWeight="700">
        STAGE 1: QUENCH
      </text>
      <text x="257" y="146" textAnchor="middle" fill="#cbd5e1" fontSize="9.5" fontFamily="monospace">
        {dim(evaluation.cots.quench_vessel.nominal_in)} × {dim(mToIn(design.quench.tower_height_m))}
      </text>
      <text x="257" y="164" textAnchor="middle" fill="#7dd3fc" fontSize="9.5" fontFamily="monospace">
        Out {fmt(kToF(design.quench.outlet_temp_k), 1)}°F
      </text>
      <text x="257" y="180" textAnchor="middle" fill="#93c5fd" fontSize="9.5" fontFamily="monospace">
        Spray {fmt(design.quench.total_spray_water_gpm, 2)} GPM
      </text>

      <line
        x1="320"
        y1="264"
        x2="366"
        y2="264"
        stroke="#38bdf8"
        strokeWidth="2.6"
        markerEnd="url(#arr2d)"
      />

      {/* Stage 2: Venturi / Eductor / Steam-Jet */}
      <g>
        {isSteam ? (
          <>
            <rect
              x="372"
              y="222"
              width="136"
              height="84"
              rx="6"
              fill="#1c2738"
              stroke={strokeFor("venturi", "#eab308")}
              strokeWidth="2.2"
            />
            <text x="440" y="244" textAnchor="middle" fill="#fde047" fontSize="10.5" fontWeight="700">
              {isSubsonic ? "SUBSONIC STEAM-JET" : "STEAM-JET EJECTOR"}
            </text>
            <text x="440" y="261" textAnchor="middle" fill="#e2e8f0" fontSize="9.5" fontFamily="monospace">
              {fmt(frugal?.steam_pressure_psig, 0)} psig · {fmt(frugal?.steam_nozzle_vel_m_s, 0)} m/s
            </text>
            <text x="440" y="276" textAnchor="middle" fill="#93c5fd" fontSize="9" fontFamily="monospace">
              M={fmt(frugal?.steam_nozzle_mach, 2)} · {fmt(frugal?.steam_flow_kg_hr, 1)} kg/hr
            </text>
            <text x="440" y="292" textAnchor="middle" fill="#86efac" fontSize="9" fontFamily="monospace">
              {isSubsonic ? "4.2m Standpipe S2" : "Certified Boiler + PRV"}
            </text>
          </>
        ) : (
          <>
            <polygon
              points="372,232 426,252 454,252 508,228 508,300 454,276 426,276 372,296"
              fill="#1e2418"
              stroke={strokeFor("venturi", "#eab308")}
              strokeWidth="2.2"
            />
            <text x="440" y="218" textAnchor="middle" fill="#facc15" fontSize="10.5" fontWeight="700">
              {isLocal
                ? "STAGE 2: TERRACOTTA BAFFLES"
                : isEductor
                  ? "STAGE 2: DRY-FAN EDUCTOR"
                  : "STAGE 2: WET VENTURI"}
            </text>
            <text x="440" y="318" textAnchor="middle" fill="#cbd5e1" fontSize="9.5" fontFamily="monospace">
              {isLocal
                ? "Gravity Spray Rose (0.03 inWC)"
                : `Throat ${dim(mToIn(design.venturi.throat_diameter_m))} · d₅₀ ${fmt(design.venturi.cut_diameter_d50_um, 2)} µm`}
            </text>
          </>
        )}
      </g>

      <line
        x1="508"
        y1="264"
        x2="554"
        y2="264"
        stroke="#38bdf8"
        strokeWidth="2.6"
        markerEnd="url(#arr2d)"
      />

      {/* Stage 3: Packed-Bed Absorber & Sump */}
      <rect
        x="560"
        y="76"
        width="162"
        height="274"
        rx="6"
        fill={isLocal ? "#2b1d18" : "#0e2620"}
        stroke={strokeFor("packed", "#10b981")}
        strokeWidth="2.2"
      />
      {/* Packing bed hatch */}
      <rect
        x="570"
        y="150"
        width="142"
        height="118"
        rx="3"
        fill="rgba(16, 185, 129, 0.16)"
        stroke="#10b981"
        strokeDasharray="5 3"
      />
      {isParallel && (
        <>
          <line x1="605" y1="150" x2="605" y2="268" stroke="#34d399" strokeWidth="1.2" />
          <line x1="641" y1="150" x2="641" y2="268" stroke="#34d399" strokeWidth="1.2" />
          <line x1="677" y1="150" x2="677" y2="268" stroke="#34d399" strokeWidth="1.2" />
        </>
      )}
      <text x="641" y="96" textAnchor="middle" fill="#34d399" fontSize="11" fontWeight="700">
        {isSumpPoolOnly
          ? "STAGE 3: TROUGH + SIEVE"
          : isLocal
            ? "STAGE 3: STONEWARE TOWER"
            : isParallel
              ? "STAGE 3: UNISEAL BUNDLE"
              : "STAGE 3: PACKED COLUMN"}
      </text>
      <text x="641" y="114" textAnchor="middle" fill="#cbd5e1" fontSize="9.5" fontFamily="monospace">
        {isSumpPoolOnly
          ? `1.4m Drenched · ${fmt(frugal?.hcl_removal_pct, 1)}% HCl`
          : isLocal
            ? `${dim(frugal?.bed_diameter_in || 24)} wide · ${fmt(frugal?.hcl_removal_pct, 1)}% HCl`
            : `${dim(evaluation.cots.packed_column.nominal_in)} · ${fmt(evaluation.cots.actual_percent_of_flood, 1)}% flood`}
      </text>
      <text x="641" y="212" textAnchor="middle" fill="#a7f3d0" fontSize="9.5" fontFamily="monospace">
        {isSumpPoolOnly
          ? "2-Stage 0.5\" Ceramic Sieve"
          : isLocal
            ? "1\" Stoneware Rings"
            : evaluation.architecture_options.packing_media.label.slice(0, 22)}
      </text>
      <text x="641" y="334" textAnchor="middle" fill="#7dd3fc" fontSize="9.5" fontFamily="monospace">
        {isSumpPoolOnly
          ? `DUAL SUMP POOLS (${fmt(evaluation.recirculation.sump_volume_gal, 0)} gal)`
          : `SUMP pH 7.5–8.5 (${fmt(evaluation.recirculation.sump_volume_gal, 0)} gal)`}
      </text>

      <line
        x1="722"
        y1="116"
        x2="778"
        y2="116"
        stroke="#38bdf8"
        strokeWidth="2.6"
        markerEnd="url(#arr2d)"
      />

      {/* Stage 4: Blower / Eductor Stack / Chimney */}
      <rect
        x="784"
        y="46"
        width="136"
        height="304"
        rx="6"
        fill="#102136"
        stroke={strokeFor("blower", "#38bdf8")}
        strokeWidth="2.2"
      />
      <text x="852" y="70" textAnchor="middle" fill="#7dd3fc" fontSize="11" fontWeight="700">
        {isLocal ? "STAGE 4: CHIMNEY" : isEductor ? "EXHAUST STACK" : "STAGE 4: BLOWER"}
      </text>
      <text x="852" y="90" textAnchor="middle" fill="#e2e8f0" fontSize="9.5" fontFamily="monospace">
        {isLocal
          ? `${fmt(frugal?.chimney_height_m, 1)} m (${fmt(frugal?.chimney_height_ft, 0)} ft)`
          : isEductor
            ? `No Wet Fan (+${fmt(evaluation.architecture_options.eductor.positive_pressure_inwc, 2)}")`
            : `${fmt(design.blower.flow_acfm, 0)} ACFM`}
      </text>
      <text x="852" y="108" textAnchor="middle" fill="#93c5fd" fontSize="9.5" fontFamily="monospace">
        {isLocal
          ? `Avail ${fmt(frugal?.available_draft_inwc, 3)}"`
          : isEductor
            ? `Dry Motive @ Stage 2`
            : `${fmt(design.blower.total_static_pressure_in_wc, 1)} inWC · ${design.blower.recommended_motor_hp} HP`}
      </text>
      <text x="852" y="128" textAnchor="middle" fill="#6ee7b7" fontSize="9.5" fontFamily="monospace">
        {fmt(evaluation.urban.noise.property_line_dba, 1)} dBA @ {evaluation.urban.noise.property_line_ft}ft
      </text>
      {isLocal && (
        <g>
          <rect
            x="798"
            y="164"
            width="108"
            height="72"
            rx="4"
            fill="#13344d"
            stroke={isSumpPoolOnly ? "#34d399" : "#38bdf8"}
          />
          <text x="852" y="192" textAnchor="middle" fill="#bae6fd" fontSize="9.3" fontWeight="700">
            {isSumpPoolOnly ? "0-PUMP POOLS" : `${fmt(frugal?.gravity_tank_liters, 0)} L TANK`}
          </text>
          <text x="852" y="209" textAnchor="middle" fill="#93c5fd" fontSize="8.6" fontFamily="monospace">
            {isSumpPoolOnly ? "0 Tower · 0 Pumps" : "3.5m Braced Stand"}
          </text>
          <text x="852" y="225" textAnchor="middle" fill="#86efac" fontSize="8.2" fontFamily="monospace">
            {isSteam ? (frugal?.boiler_source || "dedicated_hearth").slice(0, 16) : "Passive Reheat"}
          </text>
        </g>
      )}
      {isDeluxe && project.deluxe_condensing_reclaimer && (
        <g>
          <rect x="798" y="164" width="108" height="60" rx="4" fill="#10332b" stroke="#34d399" />
          <text x="852" y="188" textAnchor="middle" fill="#6ee7b7" fontSize="9" fontWeight="700">
            CERAMIC/SiC HX
          </text>
          <text x="852" y="206" textAnchor="middle" fill="#a7f3d0" fontSize="8.5" fontFamily="monospace">
            150°F + Dry-Cooler
          </text>
        </g>
      )}
    </svg>
  );
}
