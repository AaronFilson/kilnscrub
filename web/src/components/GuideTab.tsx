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

type GuideSection = "readiness" | "bom" | "build" | "ops";

export function GuideTab({
  evaluation,
  busy,
  onDownloadPacket,
  onSaveJSON,
}: GuideTabProps) {
  const [section, setSection] = useState<GuideSection>("readiness");
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
      {/* Clean Hero Header */}
      <div className="card">
        <div className="card-header" style={{ marginBottom: 0 }}>
          <div>
            <h2 style={{ margin: 0, fontSize: 17 }}>{project.studio_name}</h2>
            <div style={{ fontSize: 12.5, color: "var(--text-muted)", marginTop: 4 }}>
              Estimated Hardware: <strong>{fmtUSD(cots.estimated_total_usd)}</strong> ·{" "}
              {cots.bom.items.length} parts ·{" "}
              <span className={`pill ${urban.overall_status}`}>
                Readiness: {urban.overall_status.replace("_", " ").toUpperCase()}
              </span>
            </div>
          </div>
          <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }} className="no-print">
            <button type="button" className="btn btn-sm" onClick={onSaveJSON}>
              Save project.json
            </button>
            <button type="button" className="btn btn-sm" onClick={() => window.print()}>
              Print
            </button>
            <button
              type="button"
              className="btn btn-sm btn-primary"
              disabled={busy}
              onClick={onDownloadPacket}
            >
              Download Complete Build Packet (.zip)
            </button>
          </div>
        </div>
      </div>

      {/* Focused Chapter Sub-Tabs */}
      <div className="card">
        <div className="subtab-bar no-print" role="tablist" aria-label="Guide Chapters">
          <button
            type="button"
            role="tab"
            aria-selected={section === "readiness"}
            className={`subtab-btn ${section === "readiness" ? "active" : ""}`}
            onClick={() => setSection("readiness")}
          >
            1. Site Readiness
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={section === "bom"}
            className={`subtab-btn ${section === "bom" ? "active" : ""}`}
            onClick={() => setSection("bom")}
          >
            2. COTS Parts &amp; Shopping List
            <span className="pill ok" style={{ padding: "1px 7px" }}>
              {fmtUSD(cots.bom.total_cost_usd)}
            </span>
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={section === "build"}
            className={`subtab-btn ${section === "build" ? "active" : ""}`}
            onClick={() => setSection("build")}
          >
            3. Step-by-Step Build Guide
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={section === "ops"}
            className={`subtab-btn ${section === "ops" ? "active" : ""}`}
            onClick={() => setSection("ops")}
          >
            4. Firing-Day How-To Checklist
          </button>
        </div>

        {/* Chapter 1: Urban Site Readiness + Sump Recirculation */}
        {section === "readiness" && (
          <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
            <div className="grid-3">
              <div className="site-check-item">
                <div className="control-row">
                  <strong>Electrical Service</strong>
                  <span className={`pill ${urban.electrical.status}`}>{urban.electrical.status}</span>
                </div>
                <div style={{ fontWeight: 700, fontSize: 13 }}>
                  {urban.electrical.circuit_label} ({fmt(urban.electrical.required_amps, 1)} A)
                </div>
                <div style={{ fontSize: 12, color: "var(--text-muted)" }}>{urban.electrical.summary}</div>
              </div>

              <div className="site-check-item">
                <div className="control-row">
                  <strong>Water Supply &amp; Backflow</strong>
                  <span className={`pill ${urban.water.status}`}>{urban.water.status}</span>
                </div>
                <div style={{ fontWeight: 700, fontSize: 13 }}>
                  {fmt(urban.water.required_gpm, 2)} GPM peak (~{fmt(urban.water.total_firing_gallons, 0)} gal)
                </div>
                <div style={{ fontSize: 12, color: "var(--text-muted)" }}>{urban.water.summary}</div>
              </div>

              <div className="site-check-item">
                <div className="control-row">
                  <strong>Neighbor Acoustics</strong>
                  <span className={`pill ${urban.noise.status}`}>{urban.noise.status}</span>
                </div>
                <div style={{ fontWeight: 700, fontSize: 13 }}>
                  {fmt(urban.noise.property_line_dba, 1)} dBA @ {urban.noise.property_line_ft} ft
                </div>
                <div style={{ fontSize: 12, color: "var(--text-muted)" }}>{urban.noise.summary}</div>
              </div>

              <div className="site-check-item">
                <div className="control-row">
                  <strong>Winter Steam Plume</strong>
                  <span className={`pill ${urban.plume.status}`}>{urban.plume.status}</span>
                </div>
                <div style={{ fontWeight: 700, fontSize: 13 }}>
                  ~{fmt(urban.plume.bare_plume_length_ft, 0)} ft bare
                  {!project.zero_electricity
                    ? ` → ~${fmt(urban.plume.diluted_plume_length_ft, 0)} ft diluted`
                    : " (passive chimney)"}
                </div>
                <div style={{ fontSize: 12, color: "var(--text-muted)" }}>{urban.plume.summary}</div>
              </div>

              <div className="site-check-item">
                <div className="control-row">
                  <strong>Alkali &amp; Neutralized Drain</strong>
                  <span className={`pill ${urban.drain.status}`}>{urban.drain.status}</span>
                </div>
                <div style={{ fontWeight: 700, fontSize: 13 }}>{urban.drain.bulk_packaging_label}</div>
                <div style={{ fontSize: 12, color: "var(--text-muted)" }}>{urban.drain.summary}</div>
              </div>

              <div className="site-check-item">
                <div className="control-row">
                  <strong>Corrosion &amp; Maintenance</strong>
                  <span className={`pill ${urban.corrosion.status}`}>{urban.corrosion.status}</span>
                </div>
                <div style={{ fontWeight: 700, fontSize: 13 }}>{urban.corrosion.tier_label}</div>
                <div style={{ fontSize: 12, color: "var(--text-muted)" }}>
                  {urban.corrosion.material_summary}
                </div>
              </div>
            </div>

            <div className="grid-2">
              <div className="site-check-item">
                <strong style={{ marginBottom: 4 }}>Architecture &amp; Trade-Offs</strong>
                {isLocal && frugal ? (
                  <div style={{ fontSize: 12.5, color: "#cbd5e1" }}>
                    <p style={{ marginTop: 0 }}>{frugal.summary}</p>
                    <p style={{ marginBottom: 4 }}>
                      <strong>Draft:</strong> {frugal.draft_mechanism}
                    </p>
                    <p style={{ marginBottom: 0 }}>
                      <strong>Local Alkali:</strong> {frugal.alkali_name} ({frugal.alkali_formula}) —{" "}
                      {fmt(frugal.alkali_kg_per_firing, 2)} kg/firing
                    </p>
                    {frugal.regional_machinery_note && (
                      <p style={{ marginBottom: 0, marginTop: 6, color: "var(--text-muted)" }}>
                        {frugal.regional_machinery_note}
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
                    <p style={{ marginBottom: 0 }}>
                      <strong>Footprint:</strong> {fmt(arch.footprint_length_ft, 1)} ft ×{" "}
                      {fmt(arch.footprint_width_ft, 1)} ft ({fmt(arch.footprint_savings_pct, 0)}% length saved)
                    </p>
                  </div>
                )}
              </div>

              <div className="site-check-item">
                <strong style={{ marginBottom: 4 }}>Water Reuse &amp; Sump Dump Schedule</strong>
                <p style={{ fontSize: 12.5, color: "#cbd5e1", marginTop: 0 }}>{recirc.summary}</p>
                <div style={{ display: "flex", gap: 6, flexWrap: "wrap", marginBottom: 6 }}>
                  <span className="pill ok">{fmt(recirc.water_saved_pct, 1)}% Water Saved</span>
                  <span className="pill info">Dump every {recirc.firings_before_full_dump} firing(s)</span>
                  <span className="pill warn">SG ≥ {recirc.dump_specific_gravity.toFixed(3)}</span>
                </div>
                <ul style={{ margin: 0, paddingLeft: 18, fontSize: 12, color: "var(--text-muted)" }}>
                  {recirc.schedule_steps.map((s, i) => (
                    <li key={i} style={{ marginBottom: 3 }}>
                      {s}
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          </div>
        )}

        {/* Chapter 2: COTS Parts & Shopping List */}
        {section === "bom" && (
          <div>
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
          </div>
        )}

        {/* Chapter 3: Step-by-Step Build Guide */}
        {section === "build" && (
          <InteractiveMarkdown
            markdown={evaluation.build_guide_md}
            storageKey={`kilnscrub:guide:${project.preset_id}:${project.build_path}`}
          />
        )}

        {/* Chapter 4: Firing-Day How-To Checklist */}
        {section === "ops" && (
          <div>
            <div className="card-header">
              <span style={{ fontSize: 12.5, color: "var(--text-muted)" }}>
                Check off items as you fire — progress is saved locally in your browser.
              </span>
            </div>
            <InteractiveMarkdown
              markdown={evaluation.operations_checklist_md}
              storageKey={`kilnscrub:ops:${project.preset_id}:${project.build_path}`}
            />
          </div>
        )}
      </div>
    </div>
  );
}
