"use client";

import React, { useEffect, useState } from "react";
import type { StudioEvaluation } from "@/types/kilnscrub";
import { fmt, fmtUSD } from "@/lib/api";

interface GuideTabProps {
  evaluation: StudioEvaluation;
  busy: boolean;
  onDownloadPacket: () => void;
  onSaveJSON: () => void;
}

function InteractiveMarkdown({
  markdown,
  storageKey,
}: {
  markdown: string;
  storageKey: string;
}) {
  const [checkedMap, setCheckedMap] = useState<Record<string, boolean>>({});

  useEffect(() => {
    try {
      const raw = localStorage.getItem(storageKey);
      if (raw) setCheckedMap(JSON.parse(raw) as Record<string, boolean>);
    } catch {
      // ignore storage errors
    }
  }, [storageKey]);

  const toggle = (key: string) => {
    setCheckedMap((prev) => {
      const next = { ...prev, [key]: !prev[key] };
      try {
        localStorage.setItem(storageKey, JSON.stringify(next));
      } catch {
        // ignore storage errors
      }
      return next;
    });
  };

  const lines = (markdown || "").split("\n");
  return (
    <div className="md-doc">
      {lines.map((raw, idx) => {
        const line = raw.trimEnd();
        if (!line.trim()) return <div key={idx} style={{ height: 6 }} />;
        if (line.startsWith("### ")) return <h3 key={idx}>{line.slice(4)}</h3>;
        if (line.startsWith("## ")) return <h2 key={idx}>{line.slice(3)}</h2>;
        if (line.startsWith("# ")) return <h1 key={idx}>{line.slice(2)}</h1>;
        if (line.startsWith("> ")) {
          return (
            <div
              key={idx}
              style={{
                borderLeft: "3px solid var(--accent-blower)",
                paddingLeft: 10,
                color: "#bae6fd",
                fontSize: 12.5,
                margin: "4px 0",
              }}
            >
              {line.slice(2)}
            </div>
          );
        }
        const checkMatch = /^\s*-\s*\[( |x|X)\]\s*(.*)$/.exec(line);
        if (checkMatch) {
          const itemKey = `${idx}:${checkMatch[2].slice(0, 40)}`;
          const isChecked = !!checkedMap[itemKey];
          return (
            <label
              key={idx}
              className={`md-check-item ${isChecked ? "checked" : ""}`}
            >
              <input
                type="checkbox"
                checked={isChecked}
                onChange={() => toggle(itemKey)}
              />
              <span>{checkMatch[2]}</span>
            </label>
          );
        }
        if (/^\s*-\s+/.test(line)) {
          return (
            <div key={idx} style={{ paddingLeft: 14, margin: "3px 0" }}>
              • {line.replace(/^\s*-\s+/, "")}
            </div>
          );
        }
        return (
          <p key={idx} style={{ margin: "4px 0" }}>
            {line}
          </p>
        );
      })}
    </div>
  );
}

export function GuideTab({
  evaluation,
  busy,
  onDownloadPacket,
  onSaveJSON,
}: GuideTabProps) {
  const [docView, setDocView] = useState<"build" | "ops">("build");
  const {
    project,
    cots,
    urban,
    architecture_options: arch,
    recirculation: recirc,
    frugal,
  } = evaluation;
  const isLocal = project.build_path === "local_ceramic_masonry";

  return (
    <div className="page-container">
      {/* Hero Action Header */}
      <div className="card">
        <div className="card-header" style={{ marginBottom: 0 }}>
          <div>
            <h2 style={{ margin: 0, fontSize: 17 }}>
              Build Guide, COTS BOM &amp; Urban Site Readiness — {project.studio_name}
            </h2>
            <div style={{ fontSize: 12.5, color: "var(--text-muted)", marginTop: 4 }}>
              Build Path: <code>{project.build_path}</code> · Estimated Hardware:{" "}
              <strong>{fmtUSD(cots.estimated_total_usd)}</strong> · Flood Ratio:{" "}
              <strong>{fmt(cots.actual_percent_of_flood, 1)}%</strong> · Overall Readiness:{" "}
              <span className={`pill ${urban.overall_status}`}>
                {urban.overall_status.toUpperCase()}
              </span>
            </div>
          </div>
          <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }} className="no-print">
            <button type="button" className="btn" onClick={onSaveJSON}>
              Save project.json
            </button>
            <button
              type="button"
              className="btn"
              onClick={() => window.print()}
            >
              Print Guide
            </button>
            <button
              type="button"
              className="btn btn-primary"
              disabled={busy}
              onClick={onDownloadPacket}
            >
              Download Complete Build Packet (.zip)
            </button>
          </div>
        </div>
      </div>

      {/* 6-Card Urban Site & Neighbor Readiness */}
      <section className="grid-3">
        <div className="card">
          <div className="card-header">
            <h3 className="card-title">Electrical Service</h3>
            <span className={`pill ${urban.electrical.status}`}>{urban.electrical.status}</span>
          </div>
          <div style={{ fontWeight: 700, marginBottom: 4 }}>
            {urban.electrical.circuit_label} ({fmt(urban.electrical.required_amps, 1)} A)
          </div>
          <div style={{ fontSize: 12, color: "var(--text-muted)" }}>{urban.electrical.summary}</div>
        </div>

        <div className="card">
          <div className="card-header">
            <h3 className="card-title">Water Supply &amp; IPC 608</h3>
            <span className={`pill ${urban.water.status}`}>{urban.water.status}</span>
          </div>
          <div style={{ fontWeight: 700, marginBottom: 4 }}>
            {fmt(urban.water.required_gpm, 2)} GPM peak (~{fmt(urban.water.total_firing_gallons, 0)} gal)
          </div>
          <div style={{ fontSize: 12, color: "var(--text-muted)" }}>{urban.water.summary}</div>
        </div>

        <div className="card">
          <div className="card-header">
            <h3 className="card-title">Neighbor Acoustics</h3>
            <span className={`pill ${urban.noise.status}`}>{urban.noise.status}</span>
          </div>
          <div style={{ fontWeight: 700, marginBottom: 4 }}>
            {fmt(urban.noise.property_line_dba, 1)} dBA @ {urban.noise.property_line_ft} ft (limit {urban.noise.urban_limit_dba} dBA)
          </div>
          <div style={{ fontSize: 12, color: "var(--text-muted)" }}>{urban.noise.summary}</div>
        </div>

        <div className="card">
          <div className="card-header">
            <h3 className="card-title">Winter Steam Plume</h3>
            <span className={`pill ${urban.plume.status}`}>{urban.plume.status}</span>
          </div>
          <div style={{ fontWeight: 700, marginBottom: 4 }}>
            ~{fmt(urban.plume.bare_plume_length_ft, 0)} ft bare
            {!project.zero_electricity
              ? ` → ~${fmt(urban.plume.diluted_plume_length_ft, 0)} ft diluted`
              : " (passive chimney)"}
          </div>
          <div style={{ fontSize: 12, color: "var(--text-muted)" }}>{urban.plume.summary}</div>
        </div>

        <div className="card">
          <div className="card-header">
            <h3 className="card-title">Alkali &amp; Neutralized Drain</h3>
            <span className={`pill ${urban.drain.status}`}>{urban.drain.status}</span>
          </div>
          <div style={{ fontWeight: 700, marginBottom: 4 }}>{urban.drain.bulk_packaging_label}</div>
          <div style={{ fontSize: 12, color: "var(--text-muted)" }}>{urban.drain.summary}</div>
        </div>

        <div className="card">
          <div className="card-header">
            <h3 className="card-title">Corrosion &amp; Alloy Care</h3>
            <span className={`pill ${urban.corrosion.status}`}>{urban.corrosion.status}</span>
          </div>
          <div style={{ fontWeight: 700, marginBottom: 4 }}>{urban.corrosion.tier_label}</div>
          <div style={{ fontSize: 12, color: "var(--text-muted)" }}>
            {urban.corrosion.material_summary}
          </div>
        </div>
      </section>

      {/* Architecture Cost & Sump Recirculation Side-by-Side */}
      <section className="grid-2">
        <div className="card">
          <h3 className="card-title" style={{ marginBottom: 8 }}>
            Architecture Cost &amp; Trade-Off Comparison
          </h3>
          {isLocal && frugal ? (
            <div style={{ fontSize: 12.5, color: "#cbd5e1" }}>
              <p style={{ marginTop: 0 }}>{frugal.summary}</p>
              <p>
                <strong>Draft:</strong> {frugal.draft_mechanism} ({fmt(frugal.available_draft_inwc, 3)} inWC avail vs{" "}
                {fmt(frugal.required_draft_inwc, 3)} inWC req)
              </p>
              <p>
                <strong>Local Alkali:</strong> {frugal.alkali_name} ({frugal.alkali_formula}) —{" "}
                {fmt(frugal.alkali_kg_per_firing, 2)} kg/firing
              </p>
              {frugal.steam_enabled && (
                <p>
                  <strong>Steam Source:</strong> {frugal.steam_source_type} ({fmt(frugal.steam_pressure_psig, 0)} psig,{" "}
                  {fmt(frugal.steam_flow_kg_hr, 1)} kg/hr, Mach {fmt(frugal.steam_nozzle_mach, 2)})
                </p>
              )}
            </div>
          ) : (
            <div style={{ fontSize: 12.5, color: "#cbd5e1" }}>
              <p style={{ marginTop: 0 }}>
                <strong>Blower Drive ({arch.blower_drive.label}):</strong>{" "}
                {arch.blower_drive.summary}
              </p>
              <p>
                <strong>Packing Media ({arch.packing_media.label}):</strong>{" "}
                {arch.packing_media.summary}
              </p>
              <p>
                <strong>Shop Compressor ({fmt(arch.compressor.available_scfm, 0)} SCFM):</strong>{" "}
                {arch.compressor.atomize_verdict}
              </p>
              <p style={{ marginBottom: 0 }}>
                <strong>Footprint:</strong> {fmt(arch.footprint_length_ft, 1)} ft ×{" "}
                {fmt(arch.footprint_width_ft, 1)} ft ({fmt(arch.footprint_savings_pct, 0)}% length saved)
              </p>
            </div>
          )}
        </div>

        <div className="card">
          <h3 className="card-title" style={{ marginBottom: 8 }}>
            Sump Recirculation &amp; Dump Schedule
          </h3>
          <p style={{ fontSize: 12.5, color: "#cbd5e1", marginTop: 0 }}>{recirc.summary}</p>
          <div style={{ display: "flex", gap: 6, flexWrap: "wrap", marginBottom: 8 }}>
            <span className="pill ok">{fmt(recirc.water_saved_pct, 1)}% Water Saved</span>
            <span className="pill info">Dump every {recirc.firings_before_full_dump} firing(s)</span>
            <span className="pill warn">Hydrometer SG ≥ {recirc.dump_specific_gravity.toFixed(3)}</span>
          </div>
          <ul style={{ margin: 0, paddingLeft: 18, fontSize: 12, color: "var(--text-muted)" }}>
            {recirc.schedule_steps.map((s, i) => (
              <li key={i} style={{ marginBottom: 3 }}>
                {s}
              </li>
            ))}
          </ul>
        </div>
      </section>

      {/* COTS Parts & Shopping List BOM Table */}
      <section className="card">
        <div className="card-header">
          <h3 className="card-title">COTS Parts &amp; Shopping List</h3>
          <span className="pill ok">
            Total: {fmtUSD(cots.bom.total_cost_usd)} · {fmt(cots.bom.total_weight_kg, 1)} kg
          </span>
        </div>
        <div className="table-wrap">
          <table className="data-table">
            <thead>
              <tr>
                <th>Category</th>
                <th>Item</th>
                <th>Material</th>
                <th className="num">Qty</th>
                <th className="num">Unit USD</th>
                <th className="num">Total USD</th>
                <th className="num">Wt (kg)</th>
              </tr>
            </thead>
            <tbody>
              {cots.bom.items.map((it, idx) => (
                <tr key={idx}>
                  <td>{it.category}</td>
                  <td>
                    <div style={{ fontWeight: 600 }}>{it.name}</div>
                    <div style={{ fontSize: 11, color: "var(--text-muted)" }}>{it.spec}</div>
                  </td>
                  <td>{it.material}</td>
                  <td className="num">
                    {it.quantity} {it.unit}
                  </td>
                  <td className="num">{fmtUSD(it.unit_cost_usd)}</td>
                  <td className="num">{fmtUSD(it.total_cost_usd)}</td>
                  <td className="num">{fmt(it.weight_kg, 1)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      {/* Interactive Step-by-Step Build Guide & Firing-Day Checklist */}
      <section className="card">
        <div className="card-header">
          <div style={{ display: "flex", gap: 6 }}>
            <button
              type="button"
              className={`btn ${docView === "build" ? "btn-primary" : ""}`}
              onClick={() => setDocView("build")}
            >
              Step-by-Step Build Guide
            </button>
            <button
              type="button"
              className={`btn ${docView === "ops" ? "btn-primary" : ""}`}
              onClick={() => setDocView("ops")}
            >
              Firing-Day How-To Checklist
            </button>
          </div>
          <span className="pill info">Checkboxes saved locally in browser</span>
        </div>
        {docView === "build" ? (
          <InteractiveMarkdown
            markdown={evaluation.build_guide_md}
            storageKey={`kilnscrub:guide:${project.preset_id}:${project.build_path}`}
          />
        ) : (
          <InteractiveMarkdown
            markdown={evaluation.operations_checklist_md}
            storageKey={`kilnscrub:ops:${project.preset_id}:${project.build_path}`}
          />
        )}
      </section>
    </div>
  );
}
