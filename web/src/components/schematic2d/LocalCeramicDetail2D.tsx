"use client";

import React from "react";
import type { StudioEvaluation } from "@/types/kilnscrub";
import { fmt } from "@/lib/api";
import type { Detail2DProps } from "./IndustrialDetail2D";

/**
 * High-detail 2D Engineering Cutaway & P&ID for Local Ceramic / Masonry & Off-Grid builds
 * (Nairobi / Mumbai / Rural Pottery) showing exact masonry wall cross-sections,
 * concentric flue reheat vs certified 6/20/60 psig steam-jet ejectors,
 * all 4 boiler heat sources (dedicated hearth, 0 psig vented Stage 1 feedwater preheater,
 * ASME/IBR <=25 L Stage 1 water jacket w/ UG-28 stiffener rings, 20 psig SiC-muffled coil),
 * and all 4 water drives (manual treadle, Steam PPCP, Cold-Water Syphon w/ <120°F drum,
 * and High-Capture Zero-Pump Flooded Sump-Pool + 1.4 m Co-Current Trough + 2-Stage Ceramic Sieve).
 */
export function LocalCeramicDetail2D({
  evaluation,
  units,
  highlightStage,
  showPlumbing,
  showControls,
  showDimensions,
  svgRef,
}: Detail2DProps) {
  const { project, design, frugal } = evaluation;
  const fr = frugal;
  const isSteam = !!fr?.steam_enabled;
  const isSubsonic = isSteam && fr?.draft_mode === "steam_ejector_subsonic";
  const isHP = isSteam && fr?.draft_mode === "steam_ejector_hp";
  const boilerSrc = fr?.boiler_source || project.local_boiler_source || "dedicated_hearth";
  const pumpMode = fr?.pump_mode || project.local_pump_mode || "manual_treadle";
  const isSumpPoolOnly = isSteam && fr?.has_overhead_tank === false;
  const isPPCP = isSteam && pumpMode === "steam_ppcp";
  const isSyphon = isSteam && pumpMode === "steam_jet_syphon";
  const usesIntegratedBoiler =
    isSteam && (boilerSrc === "stage1_water_jacket" || boilerSrc === "kiln_exhaust_coil");

  const dim = (inches: number) =>
    units === "in" ? `${inches.toFixed(1)}"` : `${Math.round(inches * 25.4)} mm`;

  const hi = (stage: string, def: string) =>
    highlightStage === stage ? "#ffffff" : def;
  const hiW = (stage: string, def = 2) => (highlightStage === stage ? def + 1.2 : def);

  const Valve = ({
    x,
    y,
    vertical = false,
    color = "#eab308",
    label,
  }: {
    x: number;
    y: number;
    vertical?: boolean;
    color?: string;
    label?: string;
  }) => (
    <g transform={`translate(${x},${y})${vertical ? " rotate(90)" : ""}`}>
      <polygon points="-6,-4.5 0,0 -6,4.5" fill="#091320" stroke={color} strokeWidth="1.4" />
      <polygon points="6,-4.5 0,0 6,4.5" fill="#091320" stroke={color} strokeWidth="1.4" />
      <line x1="0" y1="0" x2="0" y2="-7" stroke={color} strokeWidth="1.2" />
      <line x1="-4" y1="-7" x2="4" y2="-7" stroke={color} strokeWidth="1.4" />
      {label && (
        <text
          x="0"
          y={vertical ? -11 : 13}
          transform={vertical ? "rotate(-90)" : undefined}
          textAnchor="middle"
          fill="#cbd5e1"
          fontSize="7.4"
          fontFamily="monospace"
        >
          {label}
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
      style={{ display: "block", background: "#09121d" }}
    >
      <defs>
        <pattern id="locGrid" width="20" height="20" patternUnits="userSpaceOnUse">
          <path d="M 20 0 L 0 0 0 20" fill="none" stroke="#1b2d42" strokeWidth="0.6" opacity="0.55" />
        </pattern>
        <pattern id="brickHatch" width="16" height="10" patternUnits="userSpaceOnUse">
          <rect width="16" height="10" fill="#3a2016" />
          <path d="M 0 5 H 16 M 8 0 V 5 M 0 5 V 10 M 16 5 V 10" stroke="#7c3d22" strokeWidth="0.8" />
        </pattern>
        <pattern id="ringHatch" width="12" height="12" patternUnits="userSpaceOnUse">
          <circle cx="4" cy="4" r="2.8" fill="none" stroke="#d97706" strokeWidth="1" opacity="0.75" />
          <circle cx="10" cy="10" r="2.8" fill="none" stroke="#f59e0b" strokeWidth="1" opacity="0.65" />
        </pattern>
        <linearGradient id="locWater" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="#0284c7" stopOpacity="0.45" />
          <stop offset="100%" stopColor="#0c4a6e" stopOpacity="0.78" />
        </linearGradient>
        <marker id="lArrHot" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto">
          <path d="M 0 1 L 10 5 L 0 9 z" fill="#fb923c" />
        </marker>
        <marker id="lArrGas" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto">
          <path d="M 0 1 L 10 5 L 0 9 z" fill="#38bdf8" />
        </marker>
        <marker id="lArrSteam" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto">
          <path d="M 0 1 L 10 5 L 0 9 z" fill="#f8fafc" />
        </marker>
        <marker id="lArrClean" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto">
          <path d="M 0 1 L 10 5 L 0 9 z" fill="#34d399" />
        </marker>
        <marker id="lArrLiq" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="5" markerHeight="5" orient="auto">
          <path d="M 0 1 L 10 5 L 0 9 z" fill="#22d3ee" />
        </marker>
      </defs>

      <rect width="1160" height="630" fill="url(#locGrid)" />
      <rect x="8" y="8" width="1144" height="614" rx="4" fill="none" stroke="#2c3e55" strokeWidth="1.2" />

      {/* Title Strip */}
      <rect x="16" y="14" width="1128" height="28" rx="3" fill="#171410" stroke="#b45309" />
      <text x="26" y="32" fill="#fef3c7" fontSize="11.2" fontWeight="700" fontFamily="monospace">
        2D MASONRY &amp; STEAM P&amp;ID — {fr?.region_label || "LOCAL CERAMIC"} · {fmt(fr?.power_watts, 0)} W · {fr?.local_currency_estimate || ""}
      </text>
      <text x="1134" y="32" textAnchor="end" fill="#86efac" fontSize="9.8" fontFamily="monospace">
        ALL-CERAMIC &gt;1800°F IMMUNE · HCl {fmt(fr?.hcl_removal_pct, 1)}% · NaCl {fmt(fr?.submicron_nacl_capture_pct, 0)}% · {isSumpPoolOnly ? "0 PUMPS · 0 TOWER" : `${fmt(fr?.gravity_tank_liters, 0)}L GRAVITY TANK`}
      </text>

      {/* Reinforced Masonry Slab Pad (y = 398) */}
      <rect x="34" y="396" width="1090" height="12" rx="2" fill="#292524" stroke="#78716c" strokeWidth="1.5" />
      <text x="50" y="405" fill="#d6d3d1" fontSize="7.6" fontFamily="monospace" fontWeight="700">
        150 mm (6&quot;) REINFORCED CONCRETE / COMPACTED STONE MASONRY SLAB FOOTING (OUTDOOR ONLY · ≥36&quot; FLUE CLEARANCE)
      </text>

      {/* ================================================================= */}
      {/* 1. KILN FLUE + STAGE 1 MASONRY QUENCH CHAMBER                     */}
      {/* ================================================================= */}
      <g>
        {/* Kiln Brick Wall + Refractory Flue */}
        <rect x="24" y="156" width="24" height="90" fill="url(#brickHatch)" stroke="#ea580c" strokeWidth="1.5" />
        <rect
          x="48"
          y="182"
          width="112"
          height="36"
          fill="#2e160d"
          stroke={hi("inlet", "#f97316")}
          strokeWidth={hiW("inlet", 2.2)}
        />
        <line x1="56" y1="200" x2="154" y2="200" stroke="#fb923c" strokeWidth="2.6" markerEnd="url(#lArrHot)" />
        <text x="104" y="176" textAnchor="middle" fill="#fdba74" fontSize="8.8" fontWeight="700" fontFamily="monospace">
          REFRACTORY FLUE ({Math.round(design.input.peak_exit_temp_f)}°F)
        </text>

        {/* Stage 0 Passive Secondary-Air Masonry Tar Afterburner (Phase 16) */}
        {(project.local_afterburner || fr?.afterburner_enabled) && (
          <g>
            <rect x="64" y="178" width="66" height="44" rx="2" fill="rgba(234,88,12,0.28)" stroke="#f59e0b" strokeWidth="1.8" />
            <line x1="78" y1="180" x2="78" y2="220" stroke="#fbbf24" strokeWidth="2.5" strokeDasharray="5 3" />
            <line x1="96" y1="180" x2="96" y2="220" stroke="#fbbf24" strokeWidth="2.5" strokeDasharray="5 3" />
            <line x1="114" y1="180" x2="114" y2="220" stroke="#fbbf24" strokeWidth="2.5" strokeDasharray="5 3" />
            {/* Secondary air inlet ports */}
            <path d="M 86 238 L 86 222" stroke="#38bdf8" strokeWidth="1.8" markerEnd="url(#lArrGas)" />
            <path d="M 106 238 L 106 222" stroke="#38bdf8" strokeWidth="1.8" markerEnd="url(#lArrGas)" />
            <text x="97" y="250" textAnchor="middle" fill="#fde68a" fontSize="7.4" fontWeight="700" fontFamily="monospace">
              STAGE 0: 2° AIR TAR BURNER (&gt;750°C · 0W)
            </text>
          </g>
        )}

        {/* Optional Kiln Exhaust Monotube Coil (20 psig LP ONLY) */}
        {isSteam && boilerSrc === "kiln_exhaust_coil" && (
          <g>
            <rect x="76" y="168" width="56" height="64" rx="3" fill="rgba(148,163,184,0.16)" stroke="#cbd5e1" strokeWidth="1.8" />
            <path d="M 84 182 Q 92 174 100 182 T 116 182 T 124 218" fill="none" stroke="#f8fafc" strokeWidth="2.2" />
            <rect x="92" y="130" width="24" height="34" rx="4" fill="#1e293b" stroke="#f8fafc" strokeWidth="1.6" />
            <text x="104" y="123" textAnchor="middle" fill="#fde68a" fontSize="7.5" fontWeight="700" fontFamily="monospace">
              &lt;2L SiC COIL (20 psig)
            </text>
          </g>
        )}

        {/* Stage 1 Brick Quench Chamber Walls */}
        <rect
          x="160"
          y="148"
          width="118"
          height="248"
          fill="#221510"
          stroke={hi("quench", "#f97316")}
          strokeWidth={hiW("quench", 2.4)}
        />
        {/* Brick masonry wall thickness hatch */}
        <rect x="160" y="148" width="10" height="248" fill="url(#brickHatch)" />
        <rect x="268" y="148" width="10" height="248" fill="url(#brickHatch)" />
        <rect x="156" y="140" width="126" height="9" fill="#57534e" stroke="#a8a29e" strokeWidth="1" />

        {/* Stage 1 Heat-Recovery Options: 0 psig Feedwater Preheater OR ASME/IBR <=25 L Water Jacket */}
        {isSteam && boilerSrc === "stage1_feedwater_preheater" && (
          <g>
            <rect x="152" y="160" width="134" height="72" rx="3" fill="none" stroke="#f59e0b" strokeWidth="2" strokeDasharray="4 2" />
            <text x="219" y="134" textAnchor="middle" fill="#fde68a" fontSize="8" fontWeight="700" fontFamily="monospace">
              0 PSIG VENTED PREHEATER COIL (-13% FUEL)
            </text>
          </g>
        )}
        {isSteam && boilerSrc === "stage1_water_jacket" && (
          <g>
            <rect x="150" y="154" width="138" height="86" rx="4" fill="rgba(56,189,248,0.14)" stroke="#e2e8f0" strokeWidth="2.2" />
            {/* UG-28 Stiffener Rings */}
            <line x1="147" y1="176" x2="291" y2="176" stroke="#94a3b8" strokeWidth="3" />
            <line x1="147" y1="216" x2="291" y2="216" stroke="#94a3b8" strokeWidth="3" />
            {/* Top steam separator dome + PRV */}
            <rect x="234" y="106" width="32" height="34" rx="6" fill="#1e293b" stroke="#f8fafc" strokeWidth="1.8" />
            <text x="219" y="98" textAnchor="middle" fill="#7dd3fc" fontSize="8" fontWeight="700" fontFamily="monospace">
              ASME/IBR ≤25L S1 JACKET (UG-28)
            </text>
          </g>
        )}

        {/* Internal Baffle / Surge-Damped 0.5" S-Vane Skimmer & Sump Pool #1 */}
        {isSumpPoolOnly ? (
          <g>
            {/* 2x Barometric Surge Plenum + 0.50" Sloped S-Vane Skimmer dipping into Pool #1 */}
            <rect x="170" y="150" width="46" height="24" fill="#57534e" />
            <line x1="226" y1="156" x2="214" y2="348" stroke="#ea580c" strokeWidth="5" />
            <rect x="170" y="340" width="98" height="54" fill="url(#locWater)" />
            <path d="M 192 200 L 192 356 Q 214 372 238 356 L 266 346" fill="none" stroke="#38bdf8" strokeWidth="2.2" strokeDasharray="5 3" markerEnd="url(#lArrGas)" />
            <text x="219" y="256" textAnchor="middle" fill="#fde68a" fontSize="7.8" fontWeight="700" fontFamily="monospace">
              2× SURGE PLENUM
            </text>
            <text x="219" y="268" textAnchor="middle" fill="#fdba74" fontSize="7.8" fontWeight="700" fontFamily="monospace">
              0.50&quot; S-VANE SKIMMER
            </text>
            <text x="219" y="374" textAnchor="middle" fill="#bae6fd" fontSize="7.6" fontWeight="700" fontFamily="monospace">
              POOL #1 (ASH SKIM)
            </text>
          </g>
        ) : (
          <g>
            <line x1="230" y1="152" x2="230" y2="278" stroke="#ea580c" strokeWidth="4.5" />
            {/* Brass Gravity Spray Rose */}
            <polygon points="198,158 214,158 206,166" fill="#eab308" />
            <polygon points="206,166 176,326 236,326" fill="rgba(56,189,248,0.18)" />
            <rect x="170" y="362" width="98" height="32" fill="url(#locWater)" />
            <path d="M 194 200 L 194 328 L 268 328" fill="none" stroke="#38bdf8" strokeWidth="2" strokeDasharray="5 3" markerEnd="url(#lArrGas)" />
          </g>
        )}

        <text x="219" y="122" textAnchor="middle" fill="#fb923c" fontSize="9.8" fontWeight="700" fontFamily="monospace">
          STAGE 1: BRICK QUENCH
        </text>
      </g>

      {/* ================================================================= */}
      {/* 2. STAGE 2: STEAM-JET EJECTOR + BOILER (OR PASSIVE MASONRY FLUE)  */}
      {/* ================================================================= */}
      {isSteam ? (
        <g>
          {/* Standalone Certified Catering Steamer / Baby Boiler / Monotube (omitted when integrated S1 jacket or flue coil replaces it!) */}
          {!usesIntegratedBoiler && (
            <g>
              <rect
                x="306"
                y="112"
                width="76"
                height="96"
                rx={isSubsonic ? 6 : 10}
                fill={isSubsonic ? "#334155" : "#1e293b"}
                stroke="#f8fafc"
                strokeWidth="2"
              />
              <text x="344" y="148" textAnchor="middle" fill="#f8fafc" fontSize="8.2" fontWeight="700" fontFamily="monospace">
                {isSubsonic ? "CATERING" : isHP ? "CERTIFIED" : "MONOTUBE"}
              </text>
              <text x="344" y="160" textAnchor="middle" fill="#fde047" fontSize="8.2" fontWeight="700" fontFamily="monospace">
                {isSubsonic ? "STEAMER" : isHP ? "BABY BOILER" : "FLASH COIL"}
              </text>
              <text x="344" y="175" textAnchor="middle" fill="#7dd3fc" fontSize="8" fontFamily="monospace">
                {fmt(fr?.steam_pressure_psig, 0)} psig
              </text>
              <text x="344" y="187" textAnchor="middle" fill="#cbd5e1" fontSize="7.4" fontFamily="monospace">
                {fmt(fr?.steam_flow_kg_hr, 1)} kg/hr
              </text>
              {/* Certified PRV on top */}
              <polygon points="366,112 362,98 370,98" fill="#ef4444" />
              <text x="374" y="104" fill="#fca5a5" fontSize="7.4" fontWeight="700" fontFamily="monospace">PRV</text>

              {/* 4.2 m Open Hydrostatic Relief Standpipe (S2) strictly for 6 psig Subsonic! */}
              {isSubsonic && (
                <g>
                  <path d="M 306 196 L 288 196 L 288 64" fill="none" stroke="#38bdf8" strokeWidth="4" />
                  <polygon points="280,58 296,58 292,66 284,66" fill="#eab308" />
                  <text x="288" y="54" textAnchor="middle" fill="#7dd3fc" fontSize="7.4" fontWeight="700" fontFamily="monospace">
                    S2: 4.2m OPEN STANDPIPE (6 psig CAP)
                  </text>
                </g>
              )}

              {/* P6: Clean 0 psig Feedwater Tank + Check Valve -> Boiler */}
              {showPlumbing && (
                <g>
                  <rect x="412" y="138" width="54" height="44" rx="3" fill="#0c263b" stroke="#38bdf8" strokeWidth="1.4" />
                  <text x="439" y="156" textAnchor="middle" fill="#bae6fd" fontSize="7.2" fontWeight="700" fontFamily="monospace">
                    0 PSIG
                  </text>
                  <text x="439" y="168" textAnchor="middle" fill="#7dd3fc" fontSize="7" fontFamily="monospace">
                    FEED H₂O
                  </text>
                  <line x1="412" y1="164" x2="384" y2="164" stroke="#22d3ee" strokeWidth="1.8" markerEnd="url(#lArrLiq)" />
                  <text x="398" y="158" textAnchor="middle" fill="#94a3b8" fontSize="6.8" fontFamily="monospace">P6+CV</text>
                </g>
              )}
            </g>
          )}

          {/* Insulated Copper Steam Header S1 -> Ejector Nozzle */}
          <path
            d={
              boilerSrc === "stage1_water_jacket"
                ? `M 250 106 L 250 82 L 324 82 L 324 ${isSumpPoolOnly ? 344 : 316}`
                : boilerSrc === "kiln_exhaust_coil"
                  ? `M 104 130 L 104 82 L 324 82 L 324 ${isSumpPoolOnly ? 344 : 316}`
                  : `M 344 112 L 344 96 L 392 96 L 392 252 L 324 252 L 324 ${isSumpPoolOnly ? 344 : 316}`
            }
            fill="none"
            stroke="#f8fafc"
            strokeWidth="2.8"
            markerEnd="url(#lArrSteam)"
          />
          <Valve x={324} y={278} vertical color="#f8fafc" label="S1" />

          {isSumpPoolOnly ? (
            /* HIGH-CAPTURE ZERO-PUMP SUMP-POOL CONFIG: Flooded Sump-Level 60 psig Ejector + 1.4 m Co-Current Drenched Trough */
            <g>
              {/* Flooded Sump-Level 60 psig Ejector right at Pool #2 waterline (0" lift!) */}
              <polygon
                points="278,332 322,340 344,340 374,330 374,366 344,356 322,356 278,364"
                fill="#1f2316"
                stroke={hi("venturi", "#eab308")}
                strokeWidth={hiW("venturi", 2.2)}
              />
              {/* 1.4 m Horizontal Co-Current Drenched Stoneware Raschig-Ring Trough */}
              <rect
                x="374"
                y="314"
                width="174"
                height="68"
                fill="url(#ringHatch)"
                stroke={hi("packed", "#10b981")}
                strokeWidth={hiW("packed", 2.2)}
              />
              <line x1="320" y1="348" x2="540" y2="348" stroke="#f8fafc" strokeWidth="2.6" markerEnd="url(#lArrSteam)" />
              <text x="461" y="306" textAnchor="middle" fill="#6ee7b7" fontSize="8.8" fontWeight="700" fontFamily="monospace">
                1.4m CO-CURRENT DRENCHED TROUGH
              </text>
              <text x="461" y="390" textAnchor="middle" fill="#fde68a" fontSize="7.6" fontFamily="monospace">
                1&quot; Stoneware Rings (0.15 inWC) · Flooded Ejector (0&quot; Lift!)
              </text>
            </g>
          ) : (
            /* Standard Terracotta/316L Steam-Jet Ejector Mixing Tube */
            <g>
              <polygon
                points="278,302 356,314 392,314 494,296 494,346 392,328 356,328 278,340"
                fill="#1f2316"
                stroke={hi("venturi", "#eab308")}
                strokeWidth={hiW("venturi", 2.2)}
              />
              <polygon points="318,314 344,321 318,328" fill="#eab308" />
              <line x1="344" y1="321" x2="486" y2="321" stroke="#f8fafc" strokeWidth="2.4" markerEnd="url(#lArrSteam)" />
              <text x="386" y="362" textAnchor="middle" fill="#fde047" fontSize="8.5" fontWeight="700" fontFamily="monospace">
                {isSubsonic ? "DRILLED BRASS ORIFICE (M<1)" : "316L DE LAVAL EJECTOR"}
              </text>
              <text x="386" y="374" textAnchor="middle" fill="#cbd5e1" fontSize="7.8" fontFamily="monospace">
                Ø{fmt(fr?.steam_nozzle_throat_mm, 2)} mm · {fmt(fr?.steam_nozzle_vel_m_s, 0)} m/s · +{fmt(fr?.ejector_draft_inwc, 2)} inWC
              </text>
            </g>
          )}
        </g>
      ) : (
        /* Passive Natural Chimney Mode: Refractory Baffle Transfer Flue */
        <g>
          <rect
            x="278"
            y="302"
            width="216"
            height="42"
            fill="#261812"
            stroke={hi("venturi", "#eab308")}
            strokeWidth={hiW("venturi", 2)}
          />
          <line x1="286" y1="323" x2="484" y2="323" stroke="#38bdf8" strokeWidth="2.4" markerEnd="url(#lArrGas)" />
          <text x="386" y="294" textAnchor="middle" fill="#facc15" fontSize="9" fontWeight="700" fontFamily="monospace">
            STAGE 2: MASONRY BAFFLE FLUE (0.03 inWC)
          </text>
        </g>
      )}

      {/* ================================================================= */}
      {/* 3. STAGE 3: WIDE STONEWARE TOWER (OR 2-STAGE CERAMIC SIEVE TOWER) */}
      {/* ================================================================= */}
      <g>
        <rect
          x={isSumpPoolOnly ? 548 : 494}
          y="102"
          width="164"
          height="294"
          fill="#1f1713"
          stroke={hi("packed", "#10b981")}
          strokeWidth={hiW("packed", 2.4)}
        />
        <rect x={isSumpPoolOnly ? 548 : 494} y="102" width="10" height="294" fill="url(#brickHatch)" />
        <rect x={(isSumpPoolOnly ? 548 : 494) + 154} y="102" width="10" height="294" fill="url(#brickHatch)" />

        {/* 3-Pass Terracotta Chevron Mist Baffles */}
        <path
          d={`M ${(isSumpPoolOnly ? 548 : 494) + 18} 142 L ${(isSumpPoolOnly ? 548 : 494) + 42} 128 L ${(isSumpPoolOnly ? 548 : 494) + 66} 142 L ${(isSumpPoolOnly ? 548 : 494) + 90} 128 L ${(isSumpPoolOnly ? 548 : 494) + 114} 142 L ${(isSumpPoolOnly ? 548 : 494) + 144} 128`}
          fill="none"
          stroke="#ea580c"
          strokeWidth="2.2"
        />
        <text x={(isSumpPoolOnly ? 548 : 494) + 82} y="122" textAnchor="middle" fill="#fdba74" fontSize="7.4" fontFamily="monospace">
          TERRACOTTA CHEVRON DEMISTER
        </text>

        {isSumpPoolOnly ? (
          /* 2-Stage Shallow (0.5"+0.5") Perforated Ceramic Sieve / Bubble Trays */
          <g>
            {[212, 274].map((ty, idx) => (
              <g key={ty}>
                <rect x="558" y={ty - 14} width="144" height="14" fill="url(#locWater)" />
                <line x1="558" y1={ty} x2="702" y2={ty} stroke="#ea580c" strokeWidth="4" strokeDasharray="8 4" />
                <text x="630" y={ty - 4} textAnchor="middle" fill="#bae6fd" fontSize="7.4" fontWeight="700" fontFamily="monospace">
                  SIEVE TRAY #{idx + 1} (0.5&quot; FROTH)
                </text>
              </g>
            ))}
            <line x1="630" y1="340" x2="630" y2="148" stroke="#34d399" strokeWidth="2.2" strokeDasharray="6 4" markerEnd="url(#lArrClean)" />
            <text x="630" y="92" textAnchor="middle" fill="#34d399" fontSize="9.5" fontWeight="700" fontFamily="monospace">
              STAGE 3: 2-STG CERAMIC SIEVE
            </text>
          </g>
        ) : (
          /* Wide Stoneware Raschig-Ring Bed + Brass Gravity Spray Rose + Firebrick Support Arch */
          <g>
            <line x1="514" y1="164" x2="638" y2="164" stroke="#22d3ee" strokeWidth="2.6" />
            <rect x="506" y="180" width="140" height="122" fill="url(#ringHatch)" stroke="#10b981" strokeWidth="1.4" />
            <rect x="514" y="226" width="124" height="24" rx="3" fill="rgba(9,18,29,0.86)" stroke="#d97706" strokeWidth="0.9" />
            <text x="576" y="241" textAnchor="middle" fill="#fde68a" fontSize="8" fontWeight="700" fontFamily="monospace">
              1&quot; STONEWARE RINGS
            </text>
            {/* Heavy Firebrick Support Arch */}
            <path d="M 504 312 Q 576 296 648 312" fill="none" stroke="#a8a29e" strokeWidth="5" strokeDasharray="9 4" />
            <line x1="576" y1="344" x2="576" y2="146" stroke="#34d399" strokeWidth="2.2" strokeDasharray="6 4" markerEnd="url(#lArrClean)" />
            <text x="576" y="84" textAnchor="middle" fill="#34d399" fontSize="9.8" fontWeight="700" fontFamily="monospace">
              STAGE 3: WIDE STONEWARE TOWER
            </text>
            <text x="576" y="96" textAnchor="middle" fill="#cbd5e1" fontSize="8.2" fontFamily="monospace">
              Ø{dim(fr?.bed_diameter_in || 26)} · v_s={fmt(fr?.bed_superficial_vel_m_s, 2)} m/s · {fmt(fr?.bed_dp_inwc, 3)} inWC
            </text>
          </g>
        )}
      </g>

      {/* ================================================================= */}
      {/* 4. STAGE 4: SWG DUCT -> BRICK / TERRACOTTA CHIMNEY                */}
      {/* ================================================================= */}
      <g>
        {/* Curved SWG Overhead Duct into Chimney Base */}
        <path
          d={`M ${isSumpPoolOnly ? 712 : 658} 122 L 764 122 L 764 328 L 806 328`}
          fill="none"
          stroke="#ea580c"
          strokeWidth="16"
          opacity="0.26"
        />
        <path
          d={`M ${isSumpPoolOnly ? 712 : 658} 122 L 764 122 L 764 328 L 802 328`}
          fill="none"
          stroke="#34d399"
          strokeWidth="2.2"
          markerEnd="url(#lArrClean)"
        />
        <text x="740" y="114" textAnchor="middle" fill="#fdba74" fontSize="7.6" fontFamily="monospace">SWG DUCT</text>

        {/* Brick Chimney Base (with Concentric Reheat Recuperator when Passive!) */}
        <rect x="806" y="294" width="66" height="102" fill="url(#brickHatch)" stroke={hi("blower", "#38bdf8")} strokeWidth={hiW("blower", 2)} />
        {!isSteam && (
          <rect x="816" y="320" width="46" height="32" rx="2" fill="rgba(249,115,22,0.35)" stroke="#fb923c" strokeWidth="1.4" />
        )}
        {/* Tapered SWG / Brick Chimney Shaft */}
        <polygon
          points="820,62 858,62 866,294 812,294"
          fill="#2c1c16"
          stroke={hi("blower", "#38bdf8")}
          strokeWidth={hiW("blower", 2.2)}
        />
        <line x1="839" y1="364" x2="839" y2="70" stroke="#34d399" strokeWidth="2.4" markerEnd="url(#lArrClean)" />
        {/* Guy-wires */}
        <line x1="822" y1="146" x2="782" y2="396" stroke="#64748b" strokeWidth="1" strokeDasharray="4 3" />
        <line x1="856" y1="146" x2="898" y2="396" stroke="#64748b" strokeWidth="1" strokeDasharray="4 3" />

        <text x="839" y="54" textAnchor="middle" fill="#7dd3fc" fontSize="9.5" fontWeight="700" fontFamily="monospace">
          CHIMNEY {fmt(fr?.chimney_height_m, 1)}m ({fmt(fr?.chimney_height_ft, 0)}ft)
        </text>
        <text x="839" y="384" textAnchor="middle" fill="#bae6fd" fontSize="7.6" fontWeight="700" fontFamily="monospace">
          {fmt(fr?.available_draft_inwc, 3)}&quot; Avail
        </text>
      </g>

      {/* ================================================================= */}
      {/* 5. WATER SYSTEM: ELEVATED 3.5m TANK OR DUAL GROUND SUMP POOLS     */}
      {/* ================================================================= */}
      {showPlumbing && (
        isSumpPoolOnly ? (
          /* ZERO-PUMP, ZERO-TOWER DUAL GROUND-LEVEL BRICK SUMP POOLS */
          <g>
            <rect x="190" y="458" width="430" height="100" rx="5" fill="#18120f" stroke="#22d3ee" strokeWidth="2" />
            <rect x="194" y="484" width="422" height="70" rx="3" fill="url(#locWater)" />
            {/* Weir partition between Pool #1 and Pool #2 */}
            <line x1="400" y1="468" x2="400" y2="544" stroke="#a8a29e" strokeWidth="5" />
            <text x="295" y="514" textAnchor="middle" fill="#fde68a" fontSize="8.8" fontWeight="700" fontFamily="monospace">
              SUMP POOL #1: ASH SKIMMING
            </text>
            <text x="295" y="528" textAnchor="middle" fill="#cbd5e1" fontSize="7.6" fontFamily="monospace">
              Weir Baffle · {fr?.alkali_name || "Magadi Soda Ash"}
            </text>
            <text x="506" y="514" textAnchor="middle" fill="#bae6fd" fontSize="8.8" fontWeight="700" fontFamily="monospace">
              SUMP POOL #2: FLOODED WEIR FEED
            </text>
            <text x="506" y="528" textAnchor="middle" fill="#86efac" fontSize="7.6" fontFamily="monospace">
              ~95 gal (~360 L) Combined · 0 Pumps · 0 Tower!
            </text>
            {/* Gravity-Flooded Side Weir (0" suction lift!) -> 60 psig Sump-Level Ejector */}
            <path d="M 480 484 L 480 434 L 330 434 L 330 364" fill="none" stroke="#eab308" strokeWidth="2.6" markerEnd="url(#lArrLiq)" />
            <Valve x={406} y={434} color="#eab308" label="FLOODED WEIR (0&quot; LIFT)" />
            {/* Gravity return from trough & sieve tower back into Pool #1/#2 */}
            <path d="M 630 396 L 630 518 L 620 518" fill="none" stroke="#22d3ee" strokeWidth="2.4" markerEnd="url(#lArrLiq)" />
            <path d="M 219 396 L 219 484" fill="none" stroke="#22d3ee" strokeWidth="2.4" markerEnd="url(#lArrLiq)" />
            {/* Ground-level makeup bucket note */}
            <rect x="676" y="476" width="236" height="68" rx="5" fill="#0b1f1a" stroke="#34d399" strokeWidth="1.6" />
            <text x="794" y="498" textAnchor="middle" fill="#6ee7b7" fontSize="8.8" fontWeight="700" fontFamily="monospace">
              GROUND-LEVEL BUCKET MAKEUP ONLY
            </text>
            <text x="794" y="514" textAnchor="middle" fill="#cbd5e1" fontSize="8" fontFamily="monospace">
              ~{fr?.jerrycans_20l_per_hr || 0} × 20L buckets/hr @ floor level
            </text>
            <text x="794" y="529" textAnchor="middle" fill="#86efac" fontSize="7.6" fontFamily="monospace">
              Zero overhead stand · Zero electric/manual pump
            </text>
          </g>
        ) : (
          /* ELEVATED 3.5 m KENTANK / SINTEX ON X-BRACED STAND + BRICK SUMP + PUMP */
          <g>
            {/* Elevated Kentank / Sintex HDPE Gravity Tank */}
            <rect x="938" y="88" width="144" height="86" rx="8" fill="#0c2d48" stroke="#38bdf8" strokeWidth="2.2" />
            <rect x="942" y="114" width="136" height="56" rx="4" fill="url(#locWater)" />
            <text x="1010" y="106" textAnchor="middle" fill="#bae6fd" fontSize="8.8" fontWeight="700" fontFamily="monospace">
              {fmt(fr?.gravity_tank_liters, 0)} L OVERHEAD TANK
            </text>
            <text x="1010" y="138" textAnchor="middle" fill="#e0f2fe" fontSize="8" fontWeight="700" fontFamily="monospace">
              3.5 m ({fmt(fr?.tank_elevation_ft, 1)} ft)
            </text>
            <text x="1010" y="152" textAnchor="middle" fill="#7dd3fc" fontSize="7.6" fontFamily="monospace">
              {fmt(fr?.gravity_head_psi, 2)} PSI Gravity Head
            </text>

            {/* Diagonally X-Braced 3.5 m Timber/Steel Stand */}
            <line x1="948" y1="174" x2="948" y2="396" stroke="#b45309" strokeWidth="5" />
            <line x1="1072" y1="174" x2="1072" y2="396" stroke="#b45309" strokeWidth="5" />
            <line x1="948" y1="174" x2="1072" y2="285" stroke="#d97706" strokeWidth="2.5" />
            <line x1="1072" y1="174" x2="948" y2="285" stroke="#d97706" strokeWidth="2.5" />
            <line x1="948" y1="285" x2="1072" y2="396" stroke="#d97706" strokeWidth="2.5" />
            <line x1="1072" y1="285" x2="948" y2="396" stroke="#d97706" strokeWidth="2.5" />
            <text x="1010" y="388" textAnchor="middle" fill="#fcd34d" fontSize="7.4" fontWeight="700" fontFamily="monospace">
              X-BRACED STAND ({Math.round((fr?.gravity_tank_liters || 250) * 1.05 + 22)} kg)
            </text>

            {/* P4: Gravity Downcomer -> Stage 3 Brass Spray Rose + Stage 1 Quench Rose */}
            <path d="M 938 164 L 642 164" fill="none" stroke="#22d3ee" strokeWidth="2.6" markerEnd="url(#lArrLiq)" />
            <path d="M 730 164 L 730 62 L 206 62 L 206 156" fill="none" stroke="#22d3ee" strokeWidth="2.2" markerEnd="url(#lArrLiq)" />
            <Valve x={906} y={164} color="#22d3ee" label="P4" />

            {/* Ground-Level Plastered Brick Baffled Sump */}
            <rect x="240" y="468" width="234" height="88" rx="5" fill="#18120f" stroke="#22d3ee" strokeWidth="2" />
            <rect x="244" y="494" width="226" height="58" rx="3" fill="url(#locWater)" />
            <line x1="354" y1="478" x2="354" y2="542" stroke="#a8a29e" strokeWidth="4" />
            <text x="357" y="463" textAnchor="middle" fill="#22d3ee" fontSize="8.5" fontWeight="700" fontFamily="monospace">
              PLASTERED BRICK BAFFLED SUMP (~208 L / 55 gal)
            </text>
            <text x="298" y="524" textAnchor="middle" fill="#fde68a" fontSize="7.6" fontWeight="700" fontFamily="monospace">
              ASH WEIR
            </text>
            <text x="412" y="524" textAnchor="middle" fill="#bae6fd" fontSize="7.6" fontWeight="700" fontFamily="monospace">
              {fr?.alkali_formula || "Na₂CO₃"} pH 8
            </text>

            {/* P5 Gravity Returns from Stage 1 & Stage 3 into Brick Sump */}
            <path d="M 219 396 L 219 492 L 240 492" fill="none" stroke="#f59e0b" strokeWidth="2.4" markerEnd="url(#lArrLiq)" />
            <path d="M 576 396 L 576 444 L 304 444 L 304 468" fill="none" stroke="#f59e0b" strokeWidth="2.4" markerEnd="url(#lArrLiq)" />

            {/* Pump Station (Treadle / PPCP / Cold Syphon + Cold Drum / 0.5HP Monoblock) */}
            {isSyphon && (
              <g>
                <rect x="520" y="486" width="76" height="62" rx="4" fill="#0c2d48" stroke="#38bdf8" strokeWidth="1.6" />
                <text x="558" y="512" textAnchor="middle" fill="#bae6fd" fontSize="7.4" fontWeight="700" fontFamily="monospace">
                  COLD DRUM
                </text>
                <text x="558" y="526" textAnchor="middle" fill="#7dd3fc" fontSize="7" fontFamily="monospace">
                  &lt;120°F ONLY!
                </text>
                <line x1="596" y1="518" x2="644" y2="518" stroke="#38bdf8" strokeWidth="2.2" markerEnd="url(#lArrLiq)" />
              </g>
            )}
            {!isSyphon && (
              <path d="M 474 518 L 644 518" fill="none" stroke="#22d3ee" strokeWidth="2.4" markerEnd="url(#lArrLiq)" />
            )}

            <rect x="646" y="484" width="146" height="68" rx="5" fill="#0e2236" stroke={isPPCP || isSyphon ? "#f8fafc" : "#22d3ee"} strokeWidth="1.8" />
            <text x="719" y="505" textAnchor="middle" fill="#f8fafc" fontSize="8.4" fontWeight="700" fontFamily="monospace">
              {isPPCP
                ? "STEAM PPCP (0 W)"
                : isSyphon
                  ? "STEAM SYPHON (0 W)"
                  : fr?.zero_electricity
                    ? "TREADLE / HAND PUMP"
                    : "0.5 HP MONOBLOCK"}
            </text>
            <text x="719" y="520" textAnchor="middle" fill="#7dd3fc" fontSize="7.6" fontFamily="monospace">
              {isPPCP
                ? `Hot Sump · +${fmt(fr?.pump_steam_kg_hr, 2)} kg/hr stm`
                : isSyphon
                  ? `Cold Makeup · +${fmt(fr?.pump_steam_kg_hr, 1)} kg/hr`
                  : fr?.zero_electricity
                    ? `${fmt(fr?.treadle_minutes_per_hr, 1)} min/hr (${fr?.jerrycans_20l_per_hr}×20L)`
                    : `${fmt(fr?.spray_flow_lpm, 1)} L/min (230V)`}
            </text>
            <text x="719" y="536" textAnchor="middle" fill="#86efac" fontSize="7.3" fontFamily="monospace">
              {isPPCP ? "Forbes Marshall / Spirax" : isSyphon ? "Never Feed Hot Sump!" : "Foot Valve + Ash Screen"}
            </text>

            {/* P3 Lift Riser -> Top of Overhead Gravity Tank */}
            <path d="M 792 518 L 1110 518 L 1110 108 L 1084 108" fill="none" stroke="#22d3ee" strokeWidth="2.4" markerEnd="url(#lArrLiq)" />
            <text x="955" y="512" textAnchor="middle" fill="#7dd3fc" fontSize="7.8" fontWeight="700" fontFamily="monospace">
              P3: REFILL RISER → 3.5m TANK ({fmt(fr?.spray_flow_lpm, 1)} L/min)
            </text>
          </g>
        )
      )}

      {/* Field Monitoring / Zero-Electronics Callout */}
      {showControls && (
        <g>
          <rect x="946" y="458" width="186" height="96" rx="5" fill="#171410" stroke="#f59e0b" strokeWidth="1.5" />
          <text x="1039" y="476" textAnchor="middle" fill="#fcd34d" fontSize="8.5" fontWeight="700" fontFamily="monospace">
            FIELD MONITORING (0 PLC)
          </text>
          <text x="1039" y="492" textAnchor="middle" fill="#e2e8f0" fontSize="7.6" fontFamily="monospace">
            • $10 Hydrometer (Dump SG≥1.105)
          </text>
          <text x="1039" y="506" textAnchor="middle" fill="#e2e8f0" fontSize="7.6" fontFamily="monospace">
            • Litmus Strips (Hold pH 7.5–8.5)
          </text>
          <text x="1039" y="520" textAnchor="middle" fill="#e2e8f0" fontSize="7.6" fontFamily="monospace">
            • Stem Thermometer (Drain ≤140°F)
          </text>
          <text x="1039" y="536" textAnchor="middle" fill="#86efac" fontSize="7.5" fontWeight="700" fontFamily="monospace">
            Alkali: {fmt(fr?.alkali_kg_per_firing, 2)} kg {fr?.alkali_formula}
          </text>
        </g>
      )}

      {showDimensions && (
        <g opacity="0.9">
          <text x="219" y="391" textAnchor="middle" fill="#fdba74" fontSize="7.4" fontFamily="monospace">
            Firebrick &gt;1800°F
          </text>
          <text x={isSumpPoolOnly ? 630 : 576} y="391" textAnchor="middle" fill="#6ee7b7" fontSize="7.4" fontFamily="monospace">
            Ø{dim(fr?.bed_diameter_in || 26)} Ceramic
          </text>
        </g>
      )}

      {/* Bottom Legend Bar */}
      <rect x="16" y="582" width="1128" height="32" rx="3" fill="#0b1727" stroke="#2c3e55" />
      {[
        { x: 28, c: "#fb923c", lbl: "Refractory Hot Flue" },
        { x: 182, c: "#38bdf8", lbl: "Cooled Gas / Spray" },
        { x: 332, c: "#f8fafc", lbl: "S1 Certified Steam Line" },
        { x: 500, c: "#34d399", lbl: "Clean Chimney Exhaust" },
        { x: 672, c: "#22d3ee", lbl: "Gravity / Sump Water Loop" },
        { x: 868, c: "#f59e0b", lbl: "Masonry / Ceramic Internals" },
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
