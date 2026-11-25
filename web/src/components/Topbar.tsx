"use client";

import React, { useEffect, useRef, useState } from "react";

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
  onSaveJSON: () => void;
  onLoadJSON: (file: File) => void;
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
  onSaveJSON,
  onLoadJSON,
}: TopbarProps) {
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement | null>(null);
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  useEffect(() => {
    if (!menuOpen) return;
    const onDocClick = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false);
      }
    };
    document.addEventListener("mousedown", onDocClick);
    return () => document.removeEventListener("mousedown", onDocClick);
  }, [menuOpen]);

  return (
    <header className="topbar">
      <div className="brand">
        <div className="brand-mark" aria-hidden="true">
          KS
        </div>
        <div>
          <div className="brand-title">kilnscrub</div>
          <div className="brand-sub">Kiln Exhaust Scrubber Designer</div>
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
          id="tab-plans"
          type="button"
          role="tab"
          aria-selected={activeTab === "plans"}
          className={`tab-btn ${activeTab === "plans" ? "active" : ""}`}
          onClick={() => onSelectTab("plans")}
        >
          Shop Plans &amp; Cut Sheets
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
      </nav>

      <div className="topbar-actions">
        <button
          type="button"
          className="btn btn-sm btn-ghost"
          onClick={onToggleUnits}
          title="Switch between Imperial (in/ft) and Metric (mm/m)"
        >
          {units === "in" ? "in / ft" : "mm / m"}
        </button>

        <div className="popover-wrap" ref={menuRef}>
          <button
            type="button"
            className="btn btn-sm"
            onClick={() => setMenuOpen((v) => !v)}
            aria-expanded={menuOpen}
          >
            Export &amp; Files ▾
          </button>
          <input
            ref={fileInputRef}
            type="file"
            accept=".json,application/json"
            style={{ display: "none" }}
            onChange={(e) => {
              const f = e.target.files?.[0];
              if (f) onLoadJSON(f);
              e.target.value = "";
            }}
          />
          {menuOpen && (
            <div className="popover-menu" role="menu">
              <div className="popover-label">Project File</div>
              <button
                type="button"
                className="popover-item"
                onClick={() => {
                  setMenuOpen(false);
                  onSaveJSON();
                }}
              >
                <span>Save project.json</span>
              </button>
              <button
                type="button"
                className="popover-item"
                onClick={() => {
                  setMenuOpen(false);
                  fileInputRef.current?.click();
                }}
              >
                <span>Load project.json</span>
              </button>
              {!isLocalCeramic && (
                <>
                  <div className="popover-sep" />
                  <div className="popover-label">CAD &amp; Parts Export</div>
                  <button
                    type="button"
                    className="popover-item"
                    disabled={busy}
                    onClick={() => {
                      setMenuOpen(false);
                      onExport("stl");
                    }}
                  >
                    <span>3D Assembly Mesh</span>
                    <span className="pill">STL</span>
                  </button>
                  <button
                    type="button"
                    className="popover-item"
                    disabled={busy}
                    onClick={() => {
                      setMenuOpen(false);
                      onExport("dxf");
                    }}
                  >
                    <span>Unfolded Sheet Cut Patterns</span>
                    <span className="pill">DXF</span>
                  </button>
                  <button
                    type="button"
                    className="popover-item"
                    disabled={busy}
                    onClick={() => {
                      setMenuOpen(false);
                      onExport("bom_csv");
                    }}
                  >
                    <span>Bill of Materials</span>
                    <span className="pill">CSV</span>
                  </button>
                </>
              )}
            </div>
          )}
        </div>

        <button
          type="button"
          className="btn btn-sm btn-primary"
          disabled={busy}
          onClick={onDownloadPacket}
          title="Download complete tailored Build Packet (.zip) with guides, BOM & plans"
        >
          Build Packet (.zip)
        </button>
      </div>
    </header>
  );
}
