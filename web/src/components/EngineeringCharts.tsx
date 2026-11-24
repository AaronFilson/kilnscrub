"use client";

import React from "react";
import type { ScrubberDesign } from "@/types/kilnscrub";
import { fmt } from "@/lib/api";

interface EngineeringChartsProps {
  design: ScrubberDesign;
}

export function EngineeringCharts({ design }: EngineeringChartsProps) {
  const bd = design.blower.breakdown;
  const dpBars = [
    { label: "Duct", val: bd.duct_in_wc, color: "#94a3b8" },
    { label: "Quench", val: bd.quench_in_wc, color: "#f97316" },
    { label: "Venturi", val: bd.venturi_in_wc, color: "#eab308" },
    { label: "Bed", val: bd.packed_bed_in_wc, color: "#10b981" },
    { label: "Demist", val: bd.demister_in_wc, color: "#14b8a6" },
    { label: "Draft", val: bd.kiln_draft_in_wc, color: "#38bdf8" },
  ];
  const maxDP = Math.max(1, ...dpBars.map((b) => b.val));

  const effPts = design.venturi.efficiency_curve || [];
  const maxUm = Math.max(2.0, ...effPts.map((p) => p.particle_diameter_um));

  const fanPts = design.blower.fan_curve || [];
  const sysPts = design.blower.system_curve || [];
  const maxFlow = Math.max(
    100,
    ...fanPts.map((p) => p.flow_acfm),
    ...sysPts.map((p) => p.flow_acfm)
  );
  const maxPress = Math.max(
    5,
    ...fanPts.map((p) => p.pressure_in_wc),
    ...sysPts.map((p) => p.pressure_in_wc)
  );

  const W = 340;
  const H = 185;
  const mL = 38;
  const mR = 14;
  const mT = 16;
  const mB = 28;
  const pW = W - mL - mR;
  const pH = H - mT - mB;

  const poly = (
    pts: Array<{ x: number; y: number }>,
    xMax: number,
    yMax: number
  ) =>
    pts
      .map((p, idx) => {
        const px = mL + (Math.min(xMax, Math.max(0, p.x)) / xMax) * pW;
        const py = mT + pH - (Math.min(yMax, Math.max(0, p.y)) / yMax) * pH;
        return `${idx === 0 ? "M" : "L"}${px.toFixed(1)},${py.toFixed(1)}`;
      })
      .join(" ");

  const op = design.blower.operating_point;
  const opX = mL + (Math.min(maxFlow, op.flow_acfm) / maxFlow) * pW;
  const opY = mT + pH - (Math.min(maxPress, op.pressure_in_wc) / maxPress) * pH;

  return (
    <div className="charts-grid">
      {/* 1. Pressure Drop Breakdown */}
      <div className="card">
        <div className="card-header">
          <h3 className="card-title">Static Pressure Breakdown (inWC)</h3>
          <span className="pill info">
            Total {fmt(design.blower.total_static_pressure_in_wc, 2)} inWC
          </span>
        </div>
        <svg id="chart-dp" viewBox={`0 0 ${W} ${H}`} width="100%" height={H}>
          <rect x={mL} y={mT} width={pW} height={pH} fill="#0a1320" rx="4" />
          {dpBars.map((b, i) => {
            const slot = pW / dpBars.length;
            const bw = slot * 0.62;
            const bh = (b.val / maxDP) * (pH - 12);
            const x = mL + i * slot + (slot - bw) / 2;
            const y = mT + pH - bh;
            return (
              <g key={b.label}>
                <rect x={x} y={y} width={bw} height={Math.max(2, bh)} fill={b.color} rx="3" />
                <text
                  x={x + bw / 2}
                  y={y - 4}
                  textAnchor="middle"
                  fill="#e2e8f0"
                  fontSize="9.5"
                  fontFamily="monospace"
                >
                  {b.val.toFixed(2)}
                </text>
                <text
                  x={x + bw / 2}
                  y={H - 10}
                  textAnchor="middle"
                  fill="#94a3b8"
                  fontSize="9.5"
                >
                  {b.label}
                </text>
              </g>
            );
          })}
        </svg>
      </div>

      {/* 2. Calvert Particulate Capture Curve */}
      <div className="card">
        <div className="card-header">
          <h3 className="card-title">Calvert Particulate Capture Curve</h3>
          <span className="pill ok">
            d₅₀ {fmt(design.venturi.cut_diameter_d50_um, 2)} µm
          </span>
        </div>
        <svg
          id="chart-efficiency"
          viewBox={`0 0 ${W} ${H}`}
          width="100%"
          height={H}
        >
          <rect x={mL} y={mT} width={pW} height={pH} fill="#0a1320" rx="4" />
          {[0.25, 0.5, 0.75, 1.0].map((g) => {
            const y = mT + pH - g * pH;
            return (
              <g key={g}>
                <line
                  x1={mL}
                  y1={y}
                  x2={mL + pW}
                  y2={y}
                  stroke="#1e344d"
                  strokeDasharray="3 3"
                />
                <text
                  x={mL - 5}
                  y={y + 3}
                  textAnchor="end"
                  fill="#647c99"
                  fontSize="9"
                  fontFamily="monospace"
                >
                  {Math.round(g * 100)}%
                </text>
              </g>
            );
          })}
          {effPts.length > 1 && (
            <>
              <path
                d={poly(
                  effPts.map((p) => ({
                    x: p.particle_diameter_um,
                    y: p.wood_ash_efficiency,
                  })),
                  maxUm,
                  1
                )}
                fill="none"
                stroke="#f97316"
                strokeWidth="2.2"
              />
              <path
                d={poly(
                  effPts.map((p) => ({
                    x: p.particle_diameter_um,
                    y: p.nacl_fume_efficiency,
                  })),
                  maxUm,
                  1
                )}
                fill="none"
                stroke="#38bdf8"
                strokeWidth="2.2"
              />
            </>
          )}
          <text x={mL + pW / 2} y={H - 8} textAnchor="middle" fill="#94a3b8" fontSize="9.5">
            Particle Diameter (0 – {maxUm.toFixed(1)} µm) · Orange=Ash · Cyan=NaCl
          </text>
        </svg>
      </div>

      {/* 3. Blower Fan vs System Curve */}
      <div className="card">
        <div className="card-header">
          <h3 className="card-title">Blower Fan vs. System Curve</h3>
          <span className="pill info">
            {fmt(op.flow_acfm, 0)} ACFM @ {fmt(op.pressure_in_wc, 1)}&quot;
          </span>
        </div>
        <svg id="chart-fan" viewBox={`0 0 ${W} ${H}`} width="100%" height={H}>
          <rect x={mL} y={mT} width={pW} height={pH} fill="#0a1320" rx="4" />
          {[0.25, 0.5, 0.75, 1.0].map((g) => {
            const y = mT + pH - g * pH;
            return (
              <g key={g}>
                <line
                  x1={mL}
                  y1={y}
                  x2={mL + pW}
                  y2={y}
                  stroke="#1e344d"
                  strokeDasharray="3 3"
                />
                <text
                  x={mL - 5}
                  y={y + 3}
                  textAnchor="end"
                  fill="#647c99"
                  fontSize="9"
                  fontFamily="monospace"
                >
                  {(g * maxPress).toFixed(0)}&quot;
                </text>
              </g>
            );
          })}
          {fanPts.length > 1 && (
            <path
              d={poly(
                fanPts.map((p) => ({ x: p.flow_acfm, y: p.pressure_in_wc })),
                maxFlow,
                maxPress
              )}
              fill="none"
              stroke="#38bdf8"
              strokeWidth="2.2"
            />
          )}
          {sysPts.length > 1 && (
            <path
              d={poly(
                sysPts.map((p) => ({ x: p.flow_acfm, y: p.pressure_in_wc })),
                maxFlow,
                maxPress
              )}
              fill="none"
              stroke="#10b981"
              strokeWidth="2.2"
            />
          )}
          <circle cx={opX} cy={opY} r="4.5" fill="#f97316" stroke="#fff" strokeWidth="1.5" />
          <text x={mL + pW / 2} y={H - 8} textAnchor="middle" fill="#94a3b8" fontSize="9.5">
            Flow (0 – {Math.round(maxFlow)} ACFM) · Cyan=Fan · Green=System
          </text>
        </svg>
      </div>
    </div>
  );
}
