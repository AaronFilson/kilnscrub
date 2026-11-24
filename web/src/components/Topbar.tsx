"use client";

import React from "react";

export type ActiveTab = "design" | "guide" | "control" | "plans";

interface TopbarProps {
  activeTab: ActiveTab;
  onSelectTab: (tab: ActiveTab) => void;
  units: "in" | "mm";
  onToggleUnits: () => void;
  controlDotState: "off" | "ok" | "warn" | "tripped";
  busy: boolean;
  isLocalCeramic: boolean;
  onExport: (format: "stl" | "dxf" | "bom_csv" | "bom_md") => void;
  onDownloadPacket: () => void;
}

export function Topbar({
  activeTab,
  onSelectTab,
  units,
  onToggleUnits,
  controlDotState,
  busy,
  isLocalCeramic,
  onExport,
  onDownloadPacket,
}: TopbarProps) {
  return (
    <header className="topbar">
      <div className="brand">
        <div className="brand-mark" aria-hidden="true">
          KS
        </div>
        <div>
          <div className="brand-title">kilnscrub</div>
          <div className="brand-sub">
            Studio Kiln Scrubber Designer · Single-Binary Embedded UI
          </div>
        </div>
      </div>

      <nav className="tab-nav" role="tablist" aria-label="Workspace Tabs">
        <button
          id="tab-design"
          type="button"
          role="tab"
          aria-selected={activeTab === "design"}
          className={`tab-btn ${activeTab === "design" ? "active" : ""}`}
          onClick={() => onSelectTab("design")}
        >
          Design &amp; 3D Studio
        </button>
        <button
          id="tab-guide"
          type="button"
          role="tab"
          aria-selected={activeTab === "guide"}
          className={`tab-btn ${activeTab === "guide" ? "active" : ""}`}
          onClick={() => onSelectTab("guide")}
        >
          Build &amp; Firing Guide
        </button>
        <button
          id="tab-control"
          type="button"
          role="tab"
          aria-selected={activeTab === "control"}
          className={`tab-btn ${activeTab === "control" ? "active" : ""}`}
          onClick={() => onSelectTab("control")}
        >
          <span
            id="tab-control-dot"
            className={`tab-dot ${controlDotState !== "off" ? controlDotState : ""}`}
          />
          Live Control &amp; Telemetry
        </button>
        <button
          id="tab-plans"
          type="button"
          role="tab"
          aria-selected={activeTab === "plans"}
          className={`tab-btn ${activeTab === "plans" ? "active" : ""}`}
          onClick={() => onSelectTab("plans")}
        >
          Shop Plans &amp; Cut Sheets
        </button>
      </nav>

      <div className="topbar-actions">
        <button
          type="button"
          className="btn btn-sm"
          onClick={onToggleUnits}
          title="Toggle Imperial (in/ft) vs Metric (mm/m) display units"
        >
          Units: {units === "in" ? "in / ft" : "mm / m"}
        </button>
        {!isLocalCeramic && (
          <>
            <button
              type="button"
              className="btn btn-sm"
              disabled={busy}
              onClick={() => onExport("stl")}
              title="Download baseline 3D assembly STL (mm)"
            >
              STL
            </button>
            <button
              type="button"
              className="btn btn-sm"
              disabled={busy}
              onClick={() => onExport("dxf")}
              title="Download unfolded R12 DXF cut sheets (mm)"
            >
              DXF
            </button>
            <button
              type="button"
              className="btn btn-sm"
              disabled={busy}
              onClick={() => onExport("bom_csv")}
            >
              BOM CSV
            </button>
          </>
        )}
        <button
          type="button"
          className="btn btn-sm btn-primary"
          disabled={busy}
          onClick={onDownloadPacket}
          title="Download tailored Build Packet (.zip)"
        >
          Build Packet (.zip)
        </button>
      </div>
    </header>
  );
}
