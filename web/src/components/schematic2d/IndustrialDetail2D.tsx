"use client";

import React from "react";
import type { StudioEvaluation } from "@/types/kilnscrub";
import { fmt, kToF, mToIn } from "@/lib/api";

export interface Detail2DProps {
  evaluation: StudioEvaluation;
  units: "in" | "mm";
  highlightStage: string | null;
  showPlumbing: boolean;
  showControls: boolean;
  showDimensions: boolean;
  svgRef?: React.RefObject<SVGSVGElement | null>;
}

/**
 * High-detail 2D Engineering Cutaway & P&ID for Industrial / Studio builds
 * (No-Weld COTS, Welded DXF, Custom Fab Deluxe) showing exact stage internals,
 * bolted flanges, P1-P9 plumbing interconnections, barometric water-seal legs,
 * baffled sump & mag-drive pump, NFPA 86 200°F hardwired safety loop (E3),
 * and every Phase 7-14 option (Dry-Fan Eductor, Parallel Uniseal Bundle,
 * Parallel Mini-Venturis, Blower Drives, Air-Atomize, Plume Diluter,
 * Stage 1 Jacket, Variable-Throat VFD, and Ceramic/SiC Condensing Reclaimer).
 */
export function IndustrialDetail2D({
  evaluation,
  units,
  highlightStage,
  showPlumbing,
  showControls,
  showDimensions,
  svgRef,
}: Detail2DProps) {
  const { project, design, cots, urban, architecture_options: arch, recirculation: recirc } =
    evaluation;

  const isDeluxe = project.build_path === "custom_fab_deluxe";
  const isEductor = project.stage_alignment === "eductor_dry_fan";
  const isParallelBed = project.packed_bed_layout === "parallel_manifold";
  const isParallelVent = project.venturi_layout === "parallel_mini_venturis";
  const hasSilencer = project.include_silencer || isDeluxe;
  const hasDiluter = !!project.include_plume_diluter;
  const hasS1Heat = isDeluxe && !!project.deluxe_stage1_heat_recovery;
  const hasCondHX = isDeluxe && !!project.deluxe_condensing_reclaimer;
  const hasVarThroat = isDeluxe && !!project.deluxe_variable_throat_vfd && !isEductor && !isParallelVent;
  const blowerDrive = project.blower_drive_option || "industrial_radial_vfd";
  const isAirAtomize = blowerDrive === "compressor_air_atomize";
  const isCompEductor = blowerDrive === "compressor_eductor_only";
  const isDualBypass = blowerDrive === "dual_bypass_blowers";
  const isRegenRing = blowerDrive === "regen_side_channel";
  const isHeavyAlloy = project.corrosion_tier === "heavy_c276_frp";
  const isSalt = project.vapor_glaze_mode === "salt";

  const dim = (inches: number) =>
    units === "in" ? `${inches.toFixed(1)}"` : `${Math.round(inches * 25.4)} mm`;

  const hi = (stage: string, def: string) =>
    highlightStage === stage ? "#ffffff" : def;
  const hiW = (stage: string, def = 2) => (highlightStage === stage ? def + 1.2 : def);

  const quenchOutF = kToF(design.quench.outlet_temp_k);
  const sealLegIn = isEductor
    ? Math.max(8, Math.ceil(arch.eductor.positive_pressure_inwc + 4))
    : Math.max(26, Math.ceil(design.blower.total_static_pressure_in_wc + 6));
  const totalPumpGPM =
    design.quench.total_spray_water_gpm +
    design.venturi.liquid_flow_gpm +
    design.packed_bed.recirculation_flow_gpm;

  // Reusable P&ID valve bow-tie symbol
  const Valve = ({
    x,
    y,
    vertical = false,
    color = "#38bdf8",
    label,
  }: {
    x: number;
    y: number;
    vertical?: boolean;
    color?: string;
    label?: string;
  }) => (
    <g transform={`translate(${x},${y})${vertical ? " rotate(90)" : ""}`}>
      <polygon points="-7,-5 0,0 -7,5" fill="#091320" stroke={color} strokeWidth="1.4" />
      <polygon points="7,-5 0,0 7,5" fill="#091320" stroke={color} strokeWidth="1.4" />
      <line x1="0" y1="0" x2="0" y2="-8" stroke={color} strokeWidth="1.2" />
      <line x1="-4" y1="-8" x2="4" y2="-8" stroke={color} strokeWidth="1.4" />
      {label && (
        <text
          x="0"
          y={vertical ? -12 : 14}
          transform={vertical ? "rotate(-90)" : undefined}
          textAnchor="middle"
          fill="#94a3b8"
          fontSize="7.5"
          fontFamily="monospace"
        >
          {label}
        </text>
      )}
    </g>
  );

  // Reusable P&ID flanged joint marker
  const Flange = ({
    x,
    y,
    horizontal = false,
    color = "#94a3b8",
  }: {
    x: number;
    y: number;
    horizontal?: boolean;
    color?: string;
  }) =>
    horizontal ? (
      <g>
        <line x1={x - 16} y1={y - 2} x2={x + 16} y2={y - 2} stroke={color} strokeWidth="2.2" />
        <line x1={x - 16} y1={y + 2} x2={x + 16} y2={y + 2} stroke={color} strokeWidth="2.2" />
      </g>
    ) : (
      <g>
        <line x1={x - 2} y1={y - 14} x2={x - 2} y2={y + 14} stroke={color} strokeWidth="2.2" />
        <line x1={x + 2} y1={y - 14} x2={x + 2} y2={y + 14} stroke={color} strokeWidth="2.2" />
      </g>
    );

  // Reusable P&ID Instrument Bubble (ISA-5.1 style)
  const InstBubble = ({
    x,
    y,
    tag,
    sub,
    color = "#f87171",
    tx,
    ty,
  }: {
    x: number;
    y: number;
    tag: string;
    sub?: string;
    color?: string;
    tx?: number;
    ty?: number;
  }) => (
    <g>
      {tx !== undefined && ty !== undefined && (
        <line x1={x} y1={y} x2={tx} y2={ty} stroke={color} strokeWidth="1" strokeDasharray="3 2" />
      )}
      <circle cx={x} cy={y} r="14" fill="#0b1626" stroke={color} strokeWidth="1.5" />
      <line x1={x - 12} y1={y} x2={x + 12} y2={y} stroke={color} strokeWidth="0.8" opacity="0.6" />
      <text x={x} y={y - 2} textAnchor="middle" fill={color} fontSize="7.5" fontWeight="700" fontFamily="monospace">
        {tag}
      </text>
      {sub && (
        <text x={x} y={y + 9} textAnchor="middle" fill="#e2e8f0" fontSize="6.8" fontFamily="monospace">
          {sub}
        </text>
      )}
    </g>
  );

  return (
    <svg
      ref={svgRef}
      id="viewport-2d"
      viewBox="0 0 1160 630"
      width="100%"
      height="100%"
      style={{ display: "block", background: "#07111e" }}
    >
      <defs>
        <pattern id="pidGrid" width="20" height="20" patternUnits="userSpaceOnUse">
          <path d="M 20 0 L 0 0 0 20" fill="none" stroke="#162a42" strokeWidth="0.6" opacity="0.55" />
        </pattern>
        <pattern id="packHatch" width="12" height="12" patternUnits="userSpaceOnUse">
          <circle cx="4" cy="4" r="2.6" fill="none" stroke="#34d399" strokeWidth="0.9" opacity="0.65" />
          <circle cx="10" cy="10" r="2.6" fill="none" stroke="#10b981" strokeWidth="0.9" opacity="0.55" />
        </pattern>
        <pattern id="rockwoolHatch" width="8" height="8" patternUnits="userSpaceOnUse" patternTransform="rotate(45)">
          <line x1="0" y1="0" x2="0" y2="8" stroke="#a855f7" strokeWidth="1.4" opacity="0.35" />
        </pattern>
        <linearGradient id="hotGasGrad" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stopColor="#ef4444" stopOpacity="0.45" />
          <stop offset="100%" stopColor="#f97316" stopOpacity="0.20" />
        </linearGradient>
        <linearGradient id="quenchSprayGrad" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="#38bdf8" stopOpacity="0.55" />
          <stop offset="100%" stopColor="#0ea5e9" stopOpacity="0.06" />
        </linearGradient>
        <linearGradient id="sumpLiquidGrad" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="#0284c7" stopOpacity="0.42" />
          <stop offset="100%" stopColor="#0c4a6e" stopOpacity="0.72" />
        </linearGradient>
        <marker id="arrHot" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto">
          <path d="M 0 1 L 10 5 L 0 9 z" fill="#fb923c" />
        </marker>
        <marker id="arrGas" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto">
          <path d="M 0 1 L 10 5 L 0 9 z" fill="#38bdf8" />
        </marker>
        <marker id="arrClean" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto">
          <path d="M 0 1 L 10 5 L 0 9 z" fill="#34d399" />
        </marker>
        <marker id="arrLiq" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="5" markerHeight="5" orient="auto">
          <path d="M 0 1 L 10 5 L 0 9 z" fill="#22d3ee" />
        </marker>
        <marker id="arrRet" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="5" markerHeight="5" orient="auto">
          <path d="M 0 1 L 10 5 L 0 9 z" fill="#f59e0b" />
        </marker>
        <marker id="arrSafe" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="5" markerHeight="5" orient="auto">
          <path d="M 0 1 L 10 5 L 0 9 z" fill="#f87171" />
        </marker>
      </defs>

      {/* Grid & Drawing Border */}
      <rect width="1160" height="630" fill="url(#pidGrid)" />
      <rect x="8" y="8" width="1144" height="614" rx="4" fill="none" stroke="#1e3a5f" strokeWidth="1.2" />

      {/* Top Engineering Title Strip */}
      <rect x="16" y="14" width="1128" height="28" rx="3" fill="#0d1d31" stroke="#234266" />
      <text x="26" y="32" fill="#f8fafc" fontSize="11.5" fontWeight="700" fontFamily="monospace">
        2D P&amp;ID &amp; CUTAWAY — {project.studio_name.toUpperCase()} ({project.build_path.toUpperCase()} · {isHeavyAlloy ? "C-276 / FRP" : "316L / CPVC"})
      </text>
      <text x="1134" y="32" textAnchor="end" fill="#7dd3fc" fontSize="10" fontFamily="monospace">
        ENVELOPE: {fmt(arch.footprint_length_ft, 1)} ft L × {fmt(arch.footprint_width_ft, 1)} ft W × {fmt(design.envelope.height_m / 0.3048, 1)} ft H · {isEductor ? "POSITIVE +inWC EDUCTOR" : "NEGATIVE -inWC ID SUCTION"}
      </text>

      {/* Structural Steel Skid Deck Line (y = 398) */}
      <rect x="148" y="394" width="958" height="10" rx="2" fill="#1e293b" stroke="#64748b" strokeWidth="1.3" />
      {[175, 420, 660, 880, 1080].map((ax) => (
        <g key={ax}>
          <rect x={ax - 10} y="404" width="20" height="5" fill="#475569" />
          <line x1={ax} y1="392" x2={ax} y2="411" stroke="#94a3b8" strokeWidth="1.4" />
        </g>
      ))}
      <text x="160" y="402" fill="#cbd5e1" fontSize="7.5" fontFamily="monospace" fontWeight="700">
        {isDeluxe ? "WELDED 2\"×2\" FORKLIFT STEEL SKID (6× 316 SS ANCHORS)" : "BASE SKID / CONCRETE PAD (LEVEL ±1/8\"/FT)"}
      </text>

      {/* ================================================================= */}
      {/* INLET: KILN HOT FLUE + NFPA 86 200°F FAIL-OPEN BYPASS DAMPER TEE  */}
      {/* ================================================================= */}
      <g>
        {/* Kiln wall stub */}
        <rect x="20" y="136" width="22" height="100" fill="#2c170f" stroke="#9a3412" strokeWidth="1.6" />
        <text x="31" y="190" textAnchor="middle" fill="#fdba74" fontSize="8.5" fontWeight="700" fontFamily="monospace" transform="rotate(-90 31 190)">
          KILN FLUE
        </text>

        {/* Horizontal hot flue duct into Stage 1 */}
        <rect
          x="42"
          y="168"
          width="150"
          height="32"
          fill="url(#hotGasGrad)"
          stroke={hi("inlet", "#f97316")}
          strokeWidth={hiW("inlet", 2.2)}
        />
        {/* Flow arrow inside flue */}
        <line x1="52" y1="184" x2="184" y2="184" stroke="#fb923c" strokeWidth="2.4" markerEnd="url(#arrHot)" />
        <Flange x={56} y={184} color="#fb923c" />
        <Flange x={188} y={184} color="#fb923c" />

        {/* Vertical NFPA 86 Fail-Open Auxiliary Bypass Tee & Chimney */}
        <rect
          x="96"
          y="64"
          width="34"
          height="104"
          fill="#2d1016"
          stroke="#ef4444"
          strokeWidth="1.8"
          strokeDasharray="5 3"
        />
        {/* Spring-return bypass damper blade (angled closed during normal run, springs OPEN on fault) */}
        <line x1="100" y1="156" x2="126" y2="140" stroke="#fca5a5" strokeWidth="3" />
        <circle cx="113" cy="148" r="3.5" fill="#ef4444" />
        {/* Belimo spring-return actuator box */}
        <rect x="134" y="138" width="22" height="18" rx="2" fill="#450a0a" stroke="#f87171" strokeWidth="1.3" />
        <text x="145" y="150" textAnchor="middle" fill="#fecaca" fontSize="7" fontWeight="700" fontFamily="monospace">M</text>
        <line x1="126" y1="148" x2="134" y2="148" stroke="#f87171" strokeWidth="1.5" />

        <text x="113" y="56" textAnchor="middle" fill="#fca5a5" fontSize="8.5" fontWeight="700" fontFamily="monospace">
          NFPA 211 AUX BYPASS
        </text>
        <text x="113" y="79" textAnchor="middle" fill="#fecaca" fontSize="7.5" fontFamily="monospace">
          FAIL-OPEN
        </text>
        <text x="113" y="90" textAnchor="middle" fill="#fecaca" fontSize="7.5" fontFamily="monospace">
          200°F TRIP
        </text>
        <text x="114" y="216" textAnchor="middle" fill="#fdba74" fontSize="9" fontWeight="700" fontFamily="monospace">
          {Math.round(design.input.peak_exit_temp_f)}°F · Ø{dim(mToIn(design.envelope.duct_diameter_m))}
        </text>
        <text x="114" y="228" textAnchor="middle" fill="#94a3b8" fontSize="7.8" fontFamily="monospace">
          {fmt(design.inlet_gas.volumetric_flow_acfm, 0)} ACFM · ≥36&quot; clr
        </text>
      </g>

      {/* ================================================================= */}
      {/* STAGE 1: ADIABATIC SPRAY QUENCH TOWER (CROSS-SECTION CUTAWAY)     */}
      {/* ================================================================= */}
      <g>
        {/* Outer Deluxe Stage 1 Water-Jacket Pre-Cooler (strictly below inlet flue!) */}
        {hasS1Heat && (
          <g>
            <rect
              x="184"
              y="206"
              width="126"
              height="92"
              rx="4"
              fill="rgba(56, 189, 248, 0.14)"
              stroke="#38bdf8"
              strokeWidth="2"
            />
            <text x="247" y="220" textAnchor="middle" fill="#7dd3fc" fontSize="7.5" fontWeight="700" fontFamily="monospace">
              S1 WATER JACKET (P8)
            </text>
            {/* ASME PRV symbol on left of jacket */}
            <polygon points="184,218 175,213 175,223" fill="#ef4444" />
            <text x="169" y="221" textAnchor="end" fill="#fca5a5" fontSize="7" fontFamily="monospace">PRV</text>
          </g>
        )}

        {/* Quench vessel shell */}
        <path
          d="M 194 128 L 300 128 L 300 356 L 266 384 L 228 384 L 194 356 Z"
          fill="#1b1512"
          stroke={hi("quench", "#f97316")}
          strokeWidth={hiW("quench", 2.2)}
        />
        {/* Top blind flange */}
        <rect x="188" y="122" width="118" height="6" rx="1" fill="#64748b" stroke="#cbd5e1" strokeWidth="0.8" />

        {/* Top Quench Spray Header & Full-Cone Nozzles */}
        <line x1="216" y1="142" x2="278" y2="142" stroke="#22d3ee" strokeWidth="2.5" />
        {[226, 268].map((nx) => (
          <g key={nx}>
            <line x1={nx} y1="142" x2={nx} y2="152" stroke="#22d3ee" strokeWidth="2" />
            <polygon points={`${nx - 4},152 ${nx + 4},152 ${nx},158`} fill="#eab308" />
            <polygon points={`${nx},158 ${nx - 26},285 ${nx + 26},285`} fill="url(#quenchSprayGrad)" />
          </g>
        ))}

        {/* Downward cooling gas flow centerline */}
        <line x1="247" y1="184" x2="247" y2="306" stroke="#38bdf8" strokeWidth="2" strokeDasharray="6 4" markerEnd="url(#arrGas)" />

        {/* Bottom quench liquid pool */}
        <path d="M 196 346 L 298 346 L 266 382 L 228 382 Z" fill="url(#sumpLiquidGrad)" />

        {/* Deluxe sight glass & cleanout hatch */}
        {isDeluxe && (
          <g>
            <circle cx="282" cy="252" r="9" fill="#0c4a6e" stroke="#e2e8f0" strokeWidth="1.6" />
            <text x="282" y="255" textAnchor="middle" fill="#bae6fd" fontSize="6.5" fontFamily="monospace">SG</text>
            <rect x="202" y="314" width="22" height="16" rx="2" fill="#334155" stroke="#94a3b8" strokeWidth="1.2" />
            <text x="213" y="325" textAnchor="middle" fill="#e2e8f0" fontSize="6.5" fontFamily="monospace">CO</text>
          </g>
        )}

        {/* Stage 1 Labels */}
        <text x="247" y="106" textAnchor="middle" fill="#fb923c" fontSize="10.5" fontWeight="700" fontFamily="monospace">
          STAGE 1: QUENCH
        </text>
        <text x="247" y="118" textAnchor="middle" fill="#cbd5e1" fontSize="8.5" fontFamily="monospace">
          Ø{dim(cots.quench_vessel.nominal_in)} × {dim(mToIn(design.quench.tower_height_m))}
        </text>
        <text x="247" y="298" textAnchor="middle" fill="#7dd3fc" fontSize="8.8" fontWeight="700" fontFamily="monospace">
          Out {fmt(quenchOutF, 1)}°F
        </text>
        <text x="247" y="310" textAnchor="middle" fill="#93c5fd" fontSize="8" fontFamily="monospace">
          {design.quench.nozzle_count || 2}× Noz · {fmt(design.quench.total_spray_water_gpm, 2)} GPM
        </text>
      </g>

      {/* ================================================================= */}
      {/* STAGE 2: WET VENTURI / PARALLEL MINI-VENTURIS / DRY-FAN EDUCTOR   */}
      {/* ================================================================= */}
      {isEductor ? (
        <g>
          {/* Duct from Stage 1 exit -> Vertical Downflow Eductor Tee */}
          <path d="M 300 318 L 352 318 L 352 178 L 402 178" fill="none" stroke="#38bdf8" strokeWidth="18" opacity="0.22" />
          <path d="M 300 318 L 352 318 L 352 178 L 402 178" fill="none" stroke={hi("venturi", "#eab308")} strokeWidth="2" />
          <line x1="306" y1="318" x2="348" y2="318" stroke="#38bdf8" strokeWidth="2" markerEnd="url(#arrGas)" />
          <Flange x={306} y={318} />

          {/* Elevated Clean Dry-Air Motive Fan blowing straight DOWN */}
          <circle cx="436" cy="96" r="26" fill="#102841" stroke="#38bdf8" strokeWidth="2.2" />
          <circle cx="436" cy="96" r="13" fill="none" stroke="#7dd3fc" strokeWidth="1.4" strokeDasharray="4 2" />
          <text x="436" y="64" textAnchor="middle" fill="#7dd3fc" fontSize="9" fontWeight="700" fontFamily="monospace">
            DRY MOTIVE FAN ({arch.eductor.dry_fan_motor_hp} HP)
          </text>
          <text x="436" y="75" textAnchor="middle" fill="#cbd5e1" fontSize="7.8" fontFamily="monospace">
            {fmt(arch.eductor.motive_air_cfm, 0)} CFM Clean Air (0 Acid)
          </text>
          <line x1="436" y1="122" x2="436" y2="162" stroke="#f8fafc" strokeWidth="2.4" markerEnd="url(#arrGas)" />

          {/* Vertical Flanged Mixing Tee + Downflow Venturi */}
          <polygon
            points="418,164 454,164 454,196 444,234 444,256 456,312 416,312 428,256 428,234 418,196"
            fill="#1f2316"
            stroke={hi("venturi", "#eab308")}
            strokeWidth={hiW("venturi", 2.2)}
          />
          <Flange x={436} y={164} horizontal />
          <Flange x={436} y={312} horizontal />
          <line x1="436" y1="184" x2="436" y2="306" stroke="#facc15" strokeWidth="2.2" markerEnd="url(#arrGas)" />

          {/* 90° Cyclonic Sweep Elbow -> Stage 3 Bottom Plenum */}
          <path d="M 436 312 Q 436 346 474 346 L 566 346" fill="none" stroke="#38bdf8" strokeWidth="22" opacity="0.22" />
          <path d="M 436 312 Q 436 346 474 346 L 566 346" fill="none" stroke="#10b981" strokeWidth="2" />
          <line x1="464" y1="346" x2="558" y2="346" stroke="#38bdf8" strokeWidth="2.2" markerEnd="url(#arrGas)" />
          <Flange x={562} y={346} />

          <text x="498" y="216" fill="#fde047" fontSize="9.5" fontWeight="700" fontFamily="monospace">
            STAGE 2: DOWNFLOW
          </text>
          <text x="498" y="228" fill="#fde047" fontSize="9.5" fontWeight="700" fontFamily="monospace">
            STATIC EDUCTOR
          </text>
          <text x="498" y="242" fill="#cbd5e1" fontSize="8" fontFamily="monospace">
            Mixed {fmt(arch.eductor.mixed_temp_f, 1)}°F
          </text>
          <text x="498" y="254" fill="#86efac" fontSize="8" fontFamily="monospace">
            +{fmt(arch.eductor.positive_pressure_inwc, 2)} inWC (Outdoor)
          </text>
        </g>
      ) : (
        <g>
          {/* Stage 1 -> Stage 2 flanged spool */}
          <rect x="300" y="302" width="44" height="28" fill="rgba(56,189,248,0.16)" stroke="#94a3b8" strokeWidth="1.6" />
          <line x1="306" y1="316" x2="340" y2="316" stroke="#38bdf8" strokeWidth="2.2" markerEnd="url(#arrGas)" />
          <Flange x={306} y={316} />
          <Flange x={342} y={316} />

          {isParallelVent ? (
            <g>
              {/* Parallel Sanitary Mini-Venturi Manifold */}
              <rect x="344" y="276" width="16" height="80" rx="2" fill="#1e293b" stroke={hi("venturi", "#eab308")} strokeWidth="1.8" />
              <rect x="506" y="276" width="16" height="80" rx="2" fill="#1e293b" stroke={hi("venturi", "#eab308")} strokeWidth="1.8" />
              {[290, 316, 342].map((my) => (
                <g key={my}>
                  <polygon
                    points={`360,${my - 9} 404,${my - 4} 426,${my - 4} 506,${my - 10} 506,${my + 10} 426,${my + 4} 404,${my + 4} 360,${my + 9}`}
                    fill="#1f2416"
                    stroke={hi("venturi", "#eab308")}
                    strokeWidth="1.6"
                  />
                  <line x1="368" y1={my} x2="498" y2={my} stroke="#38bdf8" strokeWidth="1.3" />
                </g>
              ))}
              <text x="433" y="264" textAnchor="middle" fill="#facc15" fontSize="9.5" fontWeight="700" fontFamily="monospace">
                STAGE 2: {arch.parallel_venturi.throat_count}× PARALLEL MINI-VENTURIS
              </text>
            </g>
          ) : (
            <g>
              {/* Single High-Velocity Wet Venturi (25° conv -> throat -> 7° diffuser) */}
              <polygon
                points="344,292 406,307 434,307 522,288 522,344 434,325 406,325 344,340"
                fill="#1e2315"
                stroke={hi("venturi", "#eab308")}
                strokeWidth={hiW("venturi", 2.2)}
              />
              {/* Throat velocity arrow */}
              <line x1="354" y1="316" x2="512" y2="316" stroke="#38bdf8" strokeWidth="2.4" markerEnd="url(#arrGas)" />

              {/* Opposed throat spray injection nozzles */}
              <polygon points="414,307 420,301 426,307" fill="#22d3ee" />
              <polygon points="414,325 420,331 426,325" fill="#22d3ee" />

              {/* Optional Deluxe Variable-Throat Flooded-Disc Plug + 24VDC Actuator */}
              {hasVarThroat && (
                <g>
                  <polygon points="412,316 424,310 424,322" fill="#facc15" stroke="#fef08a" strokeWidth="1" />
                  <line x1="420" y1="310" x2="420" y2="278" stroke="#facc15" strokeWidth="2" />
                  <rect x="404" y="256" width="32" height="22" rx="3" fill="#1e293b" stroke="#a855f7" strokeWidth="1.6" />
                  <text x="420" y="270" textAnchor="middle" fill="#e9d5ff" fontSize="7.5" fontWeight="700" fontFamily="monospace">
                    E7 VT
                  </text>
                </g>
              )}

              {/* Optional Shop Compressor Twin-Fluid Air-Atomizing FRL Line (A1) */}
              {isAirAtomize && showPlumbing && (
                <g>
                  <rect x="358" y="232" width="46" height="22" rx="3" fill="#1c1917" stroke="#f59e0b" strokeWidth="1.4" />
                  <text x="381" y="246" textAnchor="middle" fill="#fde68a" fontSize="7.5" fontWeight="700" fontFamily="monospace">
                    A1 40PSI
                  </text>
                  <path d="M 404 243 L 420 243 L 420 301" fill="none" stroke="#f59e0b" strokeWidth="1.6" strokeDasharray="4 2" markerEnd="url(#arrRet)" />
                </g>
              )}

              <text x="433" y="242" textAnchor="middle" fill="#facc15" fontSize="9.5" fontWeight="700" fontFamily="monospace">
                STAGE 2: WET VENTURI
              </text>
            </g>
          )}

          <text x="433" y="362" textAnchor="middle" fill="#cbd5e1" fontSize="8.2" fontFamily="monospace">
            Throat Ø{dim(mToIn(design.venturi.throat_diameter_m))} · {fmt(design.venturi.throat_velocity_m_s, 0)} m/s
          </text>
          <text x="433" y="374" textAnchor="middle" fill="#fde047" fontSize="8" fontFamily="monospace">
            d₅₀ {fmt(design.venturi.cut_diameter_d50_um, 2)} µm · ΔP {fmt(design.venturi.pressure_drop_in_wc, 1)} inWC
          </text>

          {/* Stage 2 -> Stage 3 flanged spool */}
          <rect x="522" y="302" width="46" height="28" fill="rgba(56,189,248,0.16)" stroke="#10b981" strokeWidth="1.6" />
          <line x1="528" y1="316" x2="562" y2="316" stroke="#38bdf8" strokeWidth="2.2" markerEnd="url(#arrGas)" />
          <Flange x={524} y={316} />
          <Flange x={564} y={316} />
        </g>
      )}

      {/* ================================================================= */}
      {/* STAGE 3: PACKED-BED ABSORBER COLUMN & DEMISTER (CUTAWAY)          */}
      {/* ================================================================= */}
      <g>
        {/* Outer Column / Shell-Less Uniseal Bundle */}
        <path
          d="M 568 122 L 606 94 L 664 94 L 702 122 L 702 368 L 664 388 L 606 388 L 568 368 Z"
          fill="#0b1f1a"
          stroke={hi("packed", "#10b981")}
          strokeWidth={hiW("packed", 2.2)}
        />
        <Flange x={635} y={94} horizontal color="#34d399" />

        {/* 3-Pass Chevron Mist Eliminator (Zig-Zag) */}
        <rect x="576" y="132" width="118" height="24" fill="#0f2d24" stroke="#34d399" strokeWidth="1.2" />
        <path
          d="M 580 150 L 592 138 L 604 150 L 616 138 L 628 150 L 640 138 L 652 150 L 664 138 L 676 150 L 688 138"
          fill="none"
          stroke="#6ee7b7"
          strokeWidth="1.8"
        />
        <text x="635" y="129" textAnchor="middle" fill="#a7f3d0" fontSize="7.5" fontFamily="monospace">
          3-PASS CHEVRON DEMISTER
        </text>

        {/* Top Liquid Distributor Header (P5) + Counter-Current Spray Rain */}
        <line x1="584" y1="172" x2="686" y2="172" stroke="#22d3ee" strokeWidth="2.6" />
        {[598, 622, 648, 672].map((dx) => (
          <g key={dx}>
            <line x1={dx} y1="172" x2={dx} y2="184" stroke="#22d3ee" strokeWidth="1.4" strokeDasharray="3 2" />
            <circle cx={dx} cy="172" r="2.2" fill="#bae6fd" />
          </g>
        ))}

        {/* Packing Bed Internals (Single Column OR Shell-Less Uniseal Parallel Pipes) */}
        {isParallelBed ? (
          <g>
            {/* Top & bottom collector plates + rubber Uniseals */}
            <rect x="572" y="184" width="126" height="7" fill="#334155" stroke="#94a3b8" strokeWidth="1" />
            <rect x="572" y="290" width="126" height="7" fill="#334155" stroke="#94a3b8" strokeWidth="1" />
            {[586, 610, 635, 660, 684].map((px) => (
              <g key={px}>
                <rect x={px - 8} y="191" width="16" height="99" fill="url(#packHatch)" stroke="#34d399" strokeWidth="1.4" />
                <rect x={px - 10} y="182" width="20" height="5" rx="1" fill="#0f172a" stroke="#64748b" strokeWidth="0.9" />
                <rect x={px - 10} y="293" width="20" height="5" rx="1" fill="#0f172a" stroke="#64748b" strokeWidth="0.9" />
              </g>
            ))}
          </g>
        ) : (
          <g>
            <rect x="576" y="188" width="118" height="106" fill="url(#packHatch)" stroke="#10b981" strokeWidth="1.4" strokeDasharray="4 2" />
            {/* Packing support grate */}
            <line x1="572" y1="296" x2="698" y2="296" stroke="#94a3b8" strokeWidth="3.2" strokeDasharray="6 3" />
          </g>
        )}

        {/* Upward scrubbing gas centerline */}
        <line x1="635" y1="322" x2="635" y2="106" stroke="#34d399" strokeWidth="2" strokeDasharray="6 4" markerEnd="url(#arrClean)" />

        {/* Lower column liquid collection pool */}
        <path d="M 570 356 L 700 356 L 664 386 L 606 386 Z" fill="url(#sumpLiquidGrad)" />

        {/* Stage 3 Labels */}
        <text x="635" y="64" textAnchor="middle" fill="#34d399" fontSize="10.5" fontWeight="700" fontFamily="monospace">
          {isParallelBed
            ? `STAGE 3: ${arch.parallel_bed.pipe_count}× ${arch.parallel_bed.pipe_nominal_in}" UNISEAL BUNDLE`
            : "STAGE 3: PACKED ABSORBER"}
        </text>
        <text x="635" y="76" textAnchor="middle" fill="#cbd5e1" fontSize="8.5" fontFamily="monospace">
          Ø{dim(cots.packed_column.nominal_in)} · {fmt(cots.actual_percent_of_flood, 1)}% flood · {fmt(design.packed_bed.hcl_removal_efficiency * 100, 1)}% HCl
        </text>
        <rect x="580" y="232" width="110" height="22" rx="3" fill="rgba(7,17,30,0.82)" stroke="#10b981" strokeWidth="0.8" />
        <text x="635" y="246" textAnchor="middle" fill="#a7f3d0" fontSize="8" fontWeight="700" fontFamily="monospace">
          {arch.packing_media.label.slice(0, 19)}
        </text>
      </g>

      {/* ================================================================= */}
      {/* STAGE 4: OVERHEAD DUCT -> BLOWER / DRIVE -> SILENCER -> COND HX   */}
      {/* ================================================================= */}
      {isEductor ? (
        /* Eductor mode: clean vertical stack directly atop Stage 3 (NO downstream wet blower!) */
        <g>
          <rect x="748" y="84" width="224" height="288" rx="6" fill="rgba(16, 33, 54, 0.55)" stroke="#1e3a5f" strokeDasharray="5 4" />
          <path d="M 635 94 L 635 62 L 856 62 L 856 320" fill="none" stroke="#34d399" strokeWidth="16" opacity="0.16" />
          <path d="M 635 94 L 635 62 L 856 62" fill="none" stroke="#34d399" strokeWidth="2.2" markerEnd="url(#arrClean)" />
          {/* Vertical Outdoor Exhaust Stack */}
          <rect x="838" y="62" width="36" height="296" fill="#0f253c" stroke={hi("blower", "#38bdf8")} strokeWidth={hiW("blower", 2)} />
          <line x1="856" y1="330" x2="856" y2="74" stroke="#34d399" strokeWidth="2.2" markerEnd="url(#arrClean)" />
          <text x="856" y="384" textAnchor="middle" fill="#7dd3fc" fontSize="9.5" fontWeight="700" fontFamily="monospace">
            OUTDOOR STACK (NO WET FAN)
          </text>
        </g>
      ) : (
        <g>
          {/* Overhead Flanged CPVC Duct Stage 3 Top -> Stage 4 Blower Inlet */}
          <path d="M 635 94 L 635 72 L 774 72 L 774 302 L 826 302" fill="none" stroke="#34d399" strokeWidth="16" opacity="0.18" />
          <path d="M 635 94 L 635 72 L 774 72 L 774 302 L 822 302" fill="none" stroke="#34d399" strokeWidth="2.2" markerEnd="url(#arrClean)" />
          <Flange x={715} y={72} />
          <Flange x={818} y={302} />

          {/* Optional Acoustic Blower Baffle Box (2" Rockwool) */}
          {hasSilencer && (
            <g>
              <rect
                x="814"
                y="246"
                width="124"
                height="136"
                rx="4"
                fill="url(#rockwoolHatch)"
                stroke="#c084fc"
                strokeWidth="1.6"
                strokeDasharray="5 3"
              />
              <text x="876" y="376" textAnchor="middle" fill="#d8b4fe" fontSize="7.5" fontWeight="700" fontFamily="monospace">
                {isDeluxe ? "2\" ROCKWOOL BOX (-22 dBA)" : "ACOUSTIC BOX (-13 dBA)"}
              </text>
            </g>
          )}

          {/* Stage 4 Blower / Drive Cutaway Geometry */}
          {isCompEductor ? (
            <g>
              <rect x="854" y="280" width="44" height="44" rx="4" fill="#1e293b" stroke={hi("blower", "#38bdf8")} strokeWidth={hiW("blower", 2.2)} />
              <text x="876" y="305" textAnchor="middle" fill="#bae6fd" fontSize="7.5" fontWeight="700" fontFamily="monospace">AIR AMP</text>
            </g>
          ) : isDualBypass ? (
            <g>
              <rect x="832" y="274" width="38" height="62" rx="4" fill="#102841" stroke={hi("blower", "#38bdf8")} strokeWidth="2" />
              <rect x="882" y="274" width="38" height="62" rx="4" fill="#102841" stroke={hi("blower", "#38bdf8")} strokeWidth="2" />
              <text x="851" y="308" textAnchor="middle" fill="#bae6fd" fontSize="7.5" fontFamily="monospace">M1</text>
              <text x="901" y="308" textAnchor="middle" fill="#bae6fd" fontSize="7.5" fontFamily="monospace">M2</text>
            </g>
          ) : isRegenRing ? (
            <g>
              <circle cx="874" cy="306" r="40" fill="#102841" stroke={hi("blower", "#38bdf8")} strokeWidth={hiW("blower", 2.4)} />
              <circle cx="874" cy="306" r="22" fill="#091320" stroke="#7dd3fc" strokeWidth="1.6" />
              <text x="874" y="309" textAnchor="middle" fill="#bae6fd" fontSize="7.5" fontWeight="700" fontFamily="monospace">RING</text>
            </g>
          ) : (
            /* Spiral Centrifugal Volute + 9-Blade Radial Impeller */
            <g>
              <path
                d="M 832 306 A 42 42 0 1 0 916 306 L 916 258 L 884 258 L 884 268 A 36 36 0 0 0 832 306 Z"
                fill="#0f253c"
                stroke={hi("blower", "#38bdf8")}
                strokeWidth={hiW("blower", 2.2)}
              />
              <circle cx="874" cy="306" r="24" fill="none" stroke="#7dd3fc" strokeWidth="1.4" />
              {[0, 40, 80, 120, 160, 200, 240, 280, 320].map((deg) => (
                <line
                  key={deg}
                  x1="874"
                  y1="306"
                  x2={874 + Math.cos((deg * Math.PI) / 180) * 23}
                  y2={306 + Math.sin((deg * Math.PI) / 180) * 23}
                  stroke="#bae6fd"
                  strokeWidth="1.3"
                />
              ))}
              <circle cx="874" cy="306" r="6" fill="#38bdf8" />
            </g>
          )}

          {/* Flexible PTFE/EPDM Vibration Boot + Vertical Exhaust Stack */}
          <rect x="884" y="248" width="32" height="10" fill="#334155" stroke="#a855f7" strokeWidth="1.4" />
          <rect x="886" y="54" width="28" height="194" fill="#0f253c" stroke={hi("blower", "#38bdf8")} strokeWidth={hiW("blower", 2)} />
          <line x1="900" y1="242" x2="900" y2="62" stroke="#34d399" strokeWidth="2.2" markerEnd="url(#arrClean)" />

          {/* Inline Stack Silencer */}
          {hasSilencer && (
            <g>
              <rect x="876" y="182" width="48" height="54" rx="3" fill="rgba(168,85,247,0.20)" stroke="#c084fc" strokeWidth="1.6" />
              <text x="900" y="212" textAnchor="middle" fill="#e9d5ff" fontSize="7.5" fontWeight="700" fontFamily="monospace">
                SILENCER
              </text>
            </g>
          )}

          {/* Deluxe Ceramic-Coated / Sintered-SiC Condensing Plume Abater (150°F) + Outdoor Dry-Cooler */}
          {hasCondHX && (
            <g>
              <rect x="866" y="106" width="68" height="62" rx="4" fill="#102c27" stroke="#34d399" strokeWidth="2" />
              {/* Internal ceramic/SiC coil zig-zag */}
              <path
                d="M 874 118 L 926 118 L 874 130 L 926 130 L 874 142 L 926 142 L 874 154 L 926 154"
                fill="none"
                stroke="#f59e0b"
                strokeWidth="1.6"
              />
              <text x="900" y="101" textAnchor="middle" fill="#6ee7b7" fontSize="8" fontWeight="700" fontFamily="monospace">
                CERAMIC/SiC HX (150°F)
              </text>

              {/* Outdoor Closed-Loop Dry-Cooler Box */}
              <rect x="976" y="114" width="88" height="46" rx="4" fill="#0c2438" stroke="#38bdf8" strokeWidth="1.6" />
              <text x="1020" y="133" textAnchor="middle" fill="#bae6fd" fontSize="8" fontWeight="700" fontFamily="monospace">
                DRY-COOLER
              </text>
              <text x="1020" y="146" textAnchor="middle" fill="#7dd3fc" fontSize="7.3" fontFamily="monospace">
                Closed Heat Sink
              </text>
              <line x1="934" y1="126" x2="976" y2="126" stroke="#38bdf8" strokeWidth="1.6" markerEnd="url(#arrGas)" />
              <line x1="976" y1="146" x2="934" y2="146" stroke="#22d3ee" strokeWidth="1.6" markerEnd="url(#arrLiq)" />
            </g>
          )}

          {/* Optional 120V Squirrel-Cage Plume Diluter + HV Exit Cone */}
          {hasDiluter && (
            <g>
              <polygon points="886,54 914,54 908,42 892,42" fill="#0ea5e9" opacity="0.45" stroke="#38bdf8" strokeWidth="1.4" />
              <rect x="914" y="68" width="52" height="16" fill="#0c2d48" stroke="#38bdf8" strokeWidth="1.4" />
              <circle cx="980" cy="76" r="14" fill="#102841" stroke="#38bdf8" strokeWidth="1.6" />
              <text x="980" y="79" textAnchor="middle" fill="#bae6fd" fontSize="6.8" fontWeight="700" fontFamily="monospace">DIL</text>
              <text x="1002" y="74" fill="#7dd3fc" fontSize="7.5" fontFamily="monospace">120V Diluter</text>
              <text x="1002" y="84" fill="#94a3b8" fontSize="7" fontFamily="monospace">{fmt(urban.plume.dilution_cfm_needed, 0)} CFM</text>
            </g>
          )}

          {/* Stage 4 Readout */}
          <text x="962" y="256" fill="#7dd3fc" fontSize="9.5" fontWeight="700" fontFamily="monospace">
            STAGE 4: {isCompEductor ? "AIR AMP" : isDualBypass ? "2× BYPASS" : isRegenRing ? "RING BLOWER" : "RADIAL ID"}
          </text>
          <text x="962" y="269" fill="#e2e8f0" fontSize="8.3" fontFamily="monospace">
            {fmt(design.blower.flow_acfm, 0)} ACFM @ {fmt(design.blower.total_static_pressure_in_wc, 1)} inWC
          </text>
          <text x="962" y="281" fill="#93c5fd" fontSize="8.3" fontFamily="monospace">
            {arch.blower_drive.motor_hp} HP · {fmt(urban.noise.property_line_dba, 1)} dBA @ {urban.noise.property_line_ft}ft
          </text>
          <text x="962" y="293" fill="#6ee7b7" fontSize="8" fontFamily="monospace">
            Exit {fmt(urban.plume.exhaust_temp_f, 0)}°F · Min Stack {fmt(urban.plume.min_stack_height_ft, 0)}ft
          </text>
        </g>
      )}

      {/* ================================================================= */}
      {/* LOWER PLUMBING LAYER: P1-P9 CLOSED-LOOP SUMP, PUMP, SEALS, VALVES */}
      {/* ================================================================= */}
      {showPlumbing && (
        <g>
          {/* P1: Potable Cold Makeup Supply + IPC 608 RPZ / >=1" Air Gap -> Sump Float Valve */}
          <rect x="26" y="458" width="96" height="34" rx="4" fill="#0b2238" stroke="#38bdf8" strokeWidth="1.5" />
          <text x="74" y="473" textAnchor="middle" fill="#bae6fd" fontSize="8" fontWeight="700" fontFamily="monospace">
            P1: POTABLE H₂O
          </text>
          <text x="74" y="485" textAnchor="middle" fill="#7dd3fc" fontSize="7.2" fontFamily="monospace">
            IPC 608 RPZ / 1&quot; GAP
          </text>
          <path d="M 122 475 L 256 475 L 256 490" fill="none" stroke="#38bdf8" strokeWidth="2.2" markerEnd="url(#arrGas)" />
          <Valve x={168} y={475} color="#38bdf8" label="P1" />
          {/* Air gap / float valve ball in sump */}
          <circle cx="256" cy="496" r="4" fill="#38bdf8" />

          {/* Atmospheric Baffled Chemical Neutralizing Sump Tank (pH 7.5-8.5) */}
          <rect x="234" y="476" width="240" height="86" rx="5" fill="#091a2c" stroke="#22d3ee" strokeWidth="2" />
          {/* Sump liquid body */}
          <rect x="237" y="502" width="234" height="58" rx="3" fill="url(#sumpLiquidGrad)" />
          {/* Internal weir baffle separating ash-settling zone from clean pump suction */}
          <line x1="358" y1="484" x2="358" y2="544" stroke="#94a3b8" strokeWidth="3.5" />
          <text x="296" y="528" textAnchor="middle" fill="#fde68a" fontSize="7.8" fontWeight="700" fontFamily="monospace">
            ASH SETTLING ZONE
          </text>
          <text x="296" y="541" textAnchor="middle" fill="#cbd5e1" fontSize="7.2" fontFamily="monospace">
            {isSalt ? `NaHCO₃ Dose (${fmt(urban.drain.baking_soda_lbs_per_firing, 0)} lb)` : "Alkaline Buffer"}
          </text>
          <text x="415" y="528" textAnchor="middle" fill="#bae6fd" fontSize="7.8" fontWeight="700" fontFamily="monospace">
            CLEAN SUCTION BAY
          </text>
          <text x="415" y="541" textAnchor="middle" fill="#7dd3fc" fontSize="7.2" fontFamily="monospace">
            pH 7.5–8.5 · {fmt(recirc.sump_volume_gal, 0)} gal
          </text>
          <text x="354" y="471" textAnchor="middle" fill="#22d3ee" fontSize="8.8" fontWeight="700" fontFamily="monospace">
            {project.shared_baffled_sump ? "SHARED BAFFLED ATMOSPHERIC SUMP (VENTED)" : "DUAL DEDICATED SUMPS (ASH + ABSORBER)"}
          </text>

          {/* P7: Cool-Down Sump Drain (<=140°F IPC 702.5, pH 6.5-8.5) */}
          <path d="M 234 548 L 154 548" fill="none" stroke="#f59e0b" strokeWidth="2.2" markerEnd="url(#arrRet)" />
          <Valve x={198} y={548} color="#f59e0b" label="P7 2&quot;" />
          <text x="148" y="545" textAnchor="end" fill="#fcd34d" fontSize="7.8" fontWeight="700" fontFamily="monospace">
            P7 DRAIN ≤140°F
          </text>
          <text x="148" y="556" textAnchor="end" fill="#94a3b8" fontSize="7" fontFamily="monospace">
            IPC 702.5 · pH 6.5–8.5
          </text>

          {/* P6: Gravity Return + Mandatory Barometric Water-Seal Leg (Stage 1 + Stage 3 -> Sump Settling Bay) */}
          <path
            d="M 247 384 L 247 446 L 318 446 L 318 516"
            fill="none"
            stroke="#f59e0b"
            strokeWidth="2.6"
            markerEnd="url(#arrRet)"
          />
          <path
            d="M 635 388 L 635 446 L 318 446"
            fill="none"
            stroke="#f59e0b"
            strokeWidth="2.6"
          />
          <rect x="444" y="434" width="156" height="17" rx="3" fill="#1c1409" stroke="#f59e0b" strokeWidth="1" />
          <text x="522" y="446" textAnchor="middle" fill="#fde68a" fontSize="7.6" fontWeight="700" fontFamily="monospace">
            P6: ≥{sealLegIn}&quot; BAROMETRIC SEAL LEG
          </text>

          {/* P9: Ceramic/SiC Condenser Acidic Condensate Return (CPVC -> Alkaline Sump!) */}
          {hasCondHX && (
            <g>
              <path
                d="M 900 168 L 900 176 L 950 176 L 950 456 L 336 456 L 336 516"
                fill="none"
                stroke="#34d399"
                strokeWidth="2"
                strokeDasharray="6 3"
                markerEnd="url(#arrClean)"
              />
              <text x="760" y="452" fill="#6ee7b7" fontSize="7.6" fontWeight="700" fontFamily="monospace">
                P9: CPVC ACID CONDENSATE (pH 2.5–3.5 → NEUTRALIZE IN SUMP)
              </text>
            </g>
          )}

          {/* P2: Flooded Sump Suction -> Isolation Valve -> Y-Strainer -> Mag-Drive Recirc Pump */}
          <path d="M 474 536 L 564 536" fill="none" stroke="#22d3ee" strokeWidth="2.8" markerEnd="url(#arrLiq)" />
          <Valve x={502} y={536} color="#22d3ee" label="P2" />
          {/* Y-Strainer symbol */}
          <g transform="translate(536,536)">
            <line x1="-8" y1="0" x2="8" y2="0" stroke="#22d3ee" strokeWidth="2.4" />
            <line x1="0" y1="0" x2="7" y2="11" stroke="#22d3ee" strokeWidth="2.4" />
            <text x="0" y="-8" textAnchor="middle" fill="#94a3b8" fontSize="7" fontFamily="monospace">Y-STR</text>
          </g>

          {/* Centrifugal Sealless Mag-Drive Pump Symbol */}
          <g transform="translate(584,536)">
            <circle cx="0" cy="0" r="18" fill="#0d2b3e" stroke="#22d3ee" strokeWidth="2.2" />
            <polygon points="-7,-9 12,0 -7,9" fill="#22d3ee" />
            <text x="0" y="30" textAnchor="middle" fill="#7dd3fc" fontSize="8" fontWeight="700" fontFamily="monospace">
              MAG-DRIVE PUMP
            </text>
            <text x="0" y="40" textAnchor="middle" fill="#94a3b8" fontSize="7.2" fontFamily="monospace">
              {fmt(totalPumpGPM, 1)} GPM Flooded
            </text>
          </g>

          {/* Pressurized Discharge Header splitting cleanly into P3 (Quench), P4 (Venturi), P5 (Packed Bed) */}
          <path d="M 584 518 L 584 494 L 732 494" fill="none" stroke="#22d3ee" strokeWidth="2.6" />

          {/* P3: Pressurized Quench Spray Supply -> Flow Switch (FS-103) -> Stage 1 Top Nozzles */}
          <path
            d="M 584 494 L 584 422 L 174 422 L 174 142 L 214 142"
            fill="none"
            stroke="#22d3ee"
            strokeWidth="2.2"
            markerEnd="url(#arrLiq)"
          />
          <Valve x={174} y={294} vertical color="#22d3ee" label="P3" />
          <text x="170" y="138" textAnchor="end" fill="#7dd3fc" fontSize="7.5" fontWeight="700" fontFamily="monospace">
            P3: {fmt(design.quench.total_spray_water_gpm, 2)} GPM
          </text>

          {/* P4: Venturi Throat Opposed Spray Injection */}
          <path
            d="M 584 494 L 584 422 L 420 422 L 420 334"
            fill="none"
            stroke="#22d3ee"
            strokeWidth="2"
            markerEnd="url(#arrLiq)"
          />
          <Valve x={420} y={390} vertical color="#22d3ee" label="P4" />

          {/* P5: Stage 3 Counter-Current Top Distributor Riser */}
          <path
            d="M 732 494 L 732 172 L 690 172"
            fill="none"
            stroke="#22d3ee"
            strokeWidth="2.4"
            markerEnd="url(#arrLiq)"
          />
          <Valve x={732} y={346} vertical color="#22d3ee" label="P5" />
          <text x="738" y="168" fill="#7dd3fc" fontSize="7.5" fontWeight="700" fontFamily="monospace">
            P5: {fmt(design.packed_bed.recirculation_flow_gpm, 1)} GPM
          </text>
        </g>
      )}

      {/* ================================================================= */}
      {/* SAFETY & CONTROLS LAYER: NFPA 86 200°F HARDWIRED INTERLOCK (E3)   */}
      {/* ================================================================= */}
      {showControls && (
        <g>
          {/* NEMA 4X / UL 508A Control & VFD Panel */}
          <rect x="976" y="454" width="146" height="106" rx="5" fill="#18122b" stroke="#f87171" strokeWidth="1.8" />
          <text x="1049" y="472" textAnchor="middle" fill="#fca5a5" fontSize="8.8" fontWeight="700" fontFamily="monospace">
            {isDeluxe ? "UL 508A NEMA 4X PANEL" : "NEMA 4X CONTROL / VFD"}
          </text>
          <text x="1049" y="486" textAnchor="middle" fill="#e2e8f0" fontSize="7.6" fontFamily="monospace">
            {urban.electrical.required_voltage}V · {fmt(urban.electrical.required_amps, 1)}A (NEC 430.122)
          </text>
          <text x="1049" y="499" textAnchor="middle" fill="#fde68a" fontSize="7.4" fontFamily="monospace">
            Tier: {project.control_tier === "electromechanical_only" ? "Hardwired Only" : "RPi Modbus + 24VDC"}
          </text>
          <rect x="988" y="508" width="122" height="20" rx="3" fill="#450a0a" stroke="#ef4444" strokeWidth="1" />
          <text x="1049" y="521" textAnchor="middle" fill="#fecaca" fontSize="7.5" fontWeight="700" fontFamily="monospace">
            E3: DE-ENERGIZE TO TRIP
          </text>
          <text x="1049" y="544" textAnchor="middle" fill="#94a3b8" fontSize="7" fontFamily="monospace">
            NEC 110.26: 36&quot;×30&quot; Clear
          </text>
          <text x="1049" y="554" textAnchor="middle" fill="#94a3b8" fontSize="7" fontFamily="monospace">
            Sight Disconnect ≤50 ft
          </text>

          {/* ISA-5.1 Instrument Bubbles */}
          <InstBubble x={76} y={132} tag="TI-101" sub="Type-K" color="#fb923c" tx={76} ty={168} />
          <InstBubble x={324} y={266} tag="TSH" sub="200°F" color="#f87171" tx={300} ty={296} />
          <InstBubble x={144} y={356} tag="FS-103" sub="0.25GPM" color="#f87171" tx={174} ty={356} />
          <InstBubble x={516} y={502} tag="pH" sub="7.5-8.5" color="#34d399" tx={474} ty={514} />

          {/* Dashed Red Hardwired Safety Signal Bus (TSH 200°F + FS-103 -> E3 Panel -> Belimo Bypass) */}
          <path
            d="M 145 138 L 145 48 L 1094 48 L 1094 454"
            fill="none"
            stroke="#ef4444"
            strokeWidth="1.4"
            strokeDasharray="6 4"
            markerStart="url(#arrSafe)"
          />
          <text x="530" y="45" textAnchor="middle" fill="#fca5a5" fontSize="7.5" fontWeight="700" fontFamily="monospace">
            E3 HARDWIRED 24VDC NFPA 86 SAFETY LOOP (≥200°F OR &lt;0.25 GPM OR POWER LOSS → SPRING BYPASS OPEN + VFD STO)
          </text>
        </g>
      )}

      {/* ================================================================= */}
      {/* DIMENSIONS CALLOUT OVERLAY                                        */}
      {/* ================================================================= */}
      {showDimensions && (
        <g opacity="0.9">
          <line x1="194" y1="390" x2="300" y2="390" stroke="#fb923c" strokeWidth="1" />
          <text x="247" y="388" textAnchor="middle" fill="#fdba74" fontSize="7.4" fontFamily="monospace">
            Ø{dim(cots.quench_vessel.nominal_in)}
          </text>
          <line x1="568" y1="390" x2="702" y2="390" stroke="#34d399" strokeWidth="1" />
          <text x="635" y="388" textAnchor="middle" fill="#6ee7b7" fontSize="7.4" fontFamily="monospace">
            Ø{dim(cots.packed_column.nominal_in)} · Z={dim(mToIn(design.packed_bed.packing_height_m))}
          </text>
        </g>
      )}

      {/* ================================================================= */}
      {/* BOTTOM P&ID LEGEND BAR                                            */}
      {/* ================================================================= */}
      <rect x="16" y="582" width="1128" height="32" rx="3" fill="#0b1727" stroke="#1e3a5f" />
      {[
        { x: 28, c: "#fb923c", lbl: "Hot Flue Gas (2400°F)" },
        { x: 185, c: "#38bdf8", lbl: "Quenched Wet Gas (184°F)" },
        { x: 365, c: "#34d399", lbl: "Clean Scrubbed Exhaust" },
        { x: 535, c: "#22d3ee", lbl: "P2–P5 Pressurized Liquor" },
        { x: 705, c: "#f59e0b", lbl: "P6–P7 Gravity / Barometric Seal" },
        { x: 910, c: "#ef4444", lbl: "E3 24VDC NFPA 86 Safety Loop" },
      ].map((it) => (
        <g key={it.lbl}>
          <line x1={it.x} y1="598" x2={it.x + 22} y2="598" stroke={it.c} strokeWidth="2.8" />
          <text x={it.x + 28} y="601" fill="#cbd5e1" fontSize="8.8" fontFamily="monospace">
            {it.lbl}
          </text>
        </g>
      ))}
    </svg>
  );
}
