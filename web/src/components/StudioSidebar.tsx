"use client";

import React, { useRef } from "react";
import type { StudioPreset, StudioProject } from "@/types/kilnscrub";
import { CONE_OPTIONS, inferConeFromTemp, mToIn, reconcileProject } from "@/lib/api";

interface StudioSidebarProps {
  presets: StudioPreset[];
  project: StudioProject;
  onChange: (next: StudioProject) => void;
  onSelectPreset: (preset: StudioPreset) => void;
  onSaveJSON: () => void;
  onLoadJSON: (file: File) => void;
  onApplyBudgetCombo: () => void;
  onApplyDeluxe: () => void;
  onApplyOffGrid: () => void;
  onApplyOffGridSteam: () => void;
}

export function StudioSidebar({
  presets,
  project,
  onChange,
  onSelectPreset,
  onSaveJSON,
  onLoadJSON,
  onApplyBudgetCombo,
  onApplyDeluxe,
  onApplyOffGrid,
  onApplyOffGridSteam,
}: StudioSidebarProps) {
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  const isLocal = project.build_path === "local_ceramic_masonry";
  const isDeluxe = project.build_path === "custom_fab_deluxe";
  const isEductor = project.stage_alignment === "eductor_dry_fan";

  function patch(partial: Partial<StudioProject>) {
    onChange(reconcileProject({ ...project, ...partial }));
  }

  function patchInput(partial: Partial<StudioProject["input"]>) {
    onChange(
      reconcileProject({
        ...project,
        input: { ...project.input, ...partial },
      })
    );
  }

  return (
    <aside className="sidebar" aria-label="Studio & Engineering Configuration">
      {/* 1. Curated Presets + Quick Actions */}
      <div className="sidebar-section">
        <div className="control-row">
          <h3>Studio Kiln Presets</h3>
          <div style={{ display: "flex", gap: 4 }}>
            <button type="button" className="btn btn-sm" onClick={onSaveJSON}>
              Save project.json
            </button>
            <button
              type="button"
              className="btn btn-sm"
              onClick={() => fileInputRef.current?.click()}
            >
              Load project.json
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
          </div>
        </div>

        <div className="preset-list">
          {presets.map((pr) => (
            <button
              key={pr.id}
              type="button"
              className={`preset-card ${project.preset_id === pr.id ? "active" : ""}`}
              onClick={() => onSelectPreset(pr)}
            >
              <div className="preset-head">
                <span className="preset-name">{pr.name}</span>
                <span className="preset-badge">{pr.badge}</span>
              </div>
              <div className="preset-tag">{pr.tagline}</div>
            </button>
          ))}
        </div>

        <div className="seg-grid-2x2">
          <button
            id="btn-apply-budget-combo"
            type="button"
            className="btn btn-sm btn-accent"
            onClick={onApplyBudgetCombo}
          >
            Budget Studio Combo
          </button>
          <button
            id="btn-apply-deluxe-fab"
            type="button"
            className="btn btn-sm"
            onClick={onApplyDeluxe}
          >
            Apply Custom Fab Deluxe
          </button>
          <button
            id="btn-apply-offgrid-ceramic"
            type="button"
            className="btn btn-sm"
            onClick={onApplyOffGrid}
          >
            Off-Grid Ceramic (0W)
          </button>
          <button
            id="btn-apply-offgrid-steam"
            type="button"
            className="btn btn-sm"
            onClick={onApplyOffGridSteam}
          >
            Off-Grid Steam-Jet (60 psig)
          </button>
        </div>
      </div>

      {/* 2. Kiln & Firing Profile */}
      <div className="sidebar-section">
        <h3>Kiln &amp; Firing Profile</h3>

        <div className="control">
          <label htmlFor="in-studio-name">Studio Name</label>
          <input
            id="in-studio-name"
            type="text"
            value={project.studio_name}
            onChange={(e) => patch({ studio_name: e.target.value })}
          />
        </div>

        <div className="control">
          <div className="control-row">
            <label htmlFor="in-kiln-vol">Kiln Interior Volume</label>
            <span className="control-val">
              {project.input.kiln_volume_cu_ft.toFixed(1)} cu.ft.
            </span>
          </div>
          <input
            id="in-kiln-vol"
            type="range"
            min={2}
            max={120}
            step={1}
            value={project.input.kiln_volume_cu_ft}
            onChange={(e) =>
              patchInput({ kiln_volume_cu_ft: parseFloat(e.target.value) })
            }
          />
        </div>

        <div className="control">
          <div className="control-row">
            <label>Firing Cone Quick-Select</label>
            <span className="control-val">{project.firing_cone}</span>
          </div>
          <div className="seg-group" role="group" aria-label="Firing Cone Presets">
            {CONE_OPTIONS.map((c) => (
              <button
                key={c.cone}
                type="button"
                data-cone={c.cone}
                className={`seg-btn ${project.firing_cone === c.cone ? "active" : ""}`}
                onClick={() =>
                  onChange(
                    reconcileProject({
                      ...project,
                      firing_cone: c.cone,
                      input: { ...project.input, peak_exit_temp_f: c.tempF },
                    })
                  )
                }
              >
                {c.cone}
                <span className="seg-sub">{c.tempF}°F</span>
              </button>
            ))}
          </div>
        </div>

        <div className="control">
          <div className="control-row">
            <label htmlFor="in-peak-temp">Peak Flue Exit Temp</label>
            <span className="control-val">
              {Math.round(project.input.peak_exit_temp_f)} °F
            </span>
          </div>
          <input
            id="in-peak-temp"
            type="range"
            min={1200}
            max={2500}
            step={5}
            value={project.input.peak_exit_temp_f}
            onChange={(e) => {
              const t = parseFloat(e.target.value);
              onChange(
                reconcileProject({
                  ...project,
                  firing_cone: inferConeFromTemp(t),
                  input: { ...project.input, peak_exit_temp_f: t },
                })
              );
            }}
          />
        </div>

        <div className="control">
          <label>Glaze Vapor Mode</label>
          <div className="seg-group" role="group" aria-label="Glaze Vapor Mode">
            {(
              [
                { id: "none", label: "None", sub: "Clean / Ash" },
                { id: "soda", label: "Soda", sub: "Na₂CO₃" },
                { id: "salt", label: "Salt", sub: "NaCl" },
              ] as const
            ).map((m) => (
              <button
                key={m.id}
                type="button"
                data-vapor={m.id}
                className={`seg-btn ${project.vapor_glaze_mode === m.id ? "active" : ""}`}
                onClick={() => patch({ vapor_glaze_mode: m.id })}
              >
                {m.label}
                <span className="seg-sub unit-literal">{m.sub}</span>
              </button>
            ))}
          </div>
        </div>

        {project.vapor_glaze_mode === "salt" && (
          <>
            <div className="control">
              <div className="control-row">
                <label htmlFor="in-salt-rate">Salt Rate (<span className="unit-literal">NaCl</span>)</label>
                <span className="control-val">
                  {project.input.salt_rate_lbs_per_hr.toFixed(1)} lbs/hr
                </span>
              </div>
              <input
                id="in-salt-rate"
                type="range"
                min={0.5}
                max={10}
                step={0.1}
                value={project.input.salt_rate_lbs_per_hr}
                onChange={(e) =>
                  patchInput({ salt_rate_lbs_per_hr: parseFloat(e.target.value) })
                }
              />
            </div>

            <div className="control">
              <div className="control-row">
                <label htmlFor="in-salting-hours">Active Salting Window</label>
                <span className="control-val">
                  {project.salting_hours.toFixed(1)} hrs
                </span>
              </div>
              <input
                id="in-salting-hours"
                type="range"
                min={0.5}
                max={Math.max(1, project.firing_duration_hours)}
                step={0.5}
                value={project.salting_hours}
                onChange={(e) =>
                  patch({ salting_hours: parseFloat(e.target.value) })
                }
              />
            </div>
          </>
        )}

        <div className="control">
          <div className="control-row">
            <label htmlFor="in-firing-hours">Total Firing Duration</label>
            <span className="control-val">
              {project.firing_duration_hours.toFixed(0)} hrs
            </span>
          </div>
          <input
            id="in-firing-hours"
            type="range"
            min={2}
            max={48}
            step={1}
            value={project.firing_duration_hours}
            onChange={(e) =>
              patch({ firing_duration_hours: parseFloat(e.target.value) })
            }
          />
        </div>

        <div className="control">
          <label htmlFor="in-fuel-type">Fuel Stoichiometry</label>
          <select
            id="in-fuel-type"
            value={project.input.fuel_type}
            onChange={(e) => patchInput({ fuel_type: e.target.value })}
          >
            <option value="wood_salt">Wood + Salt (wood_salt)</option>
            <option value="wood">Wood Only (wood)</option>
            <option value="gas_salt">Gas + Salt (gas_salt)</option>
            <option value="gas">Gas / Electric / Soda (gas)</option>
          </select>
        </div>
      </div>

      {/* 3. Build Path & Phase 10 Local / Off-Grid */}
      <div className="sidebar-section">
        <h3>Build Path &amp; Materials</h3>
        <div className="seg-grid-2x2" role="group" aria-label="Build Path">
          {(
            [
              { id: "cots_no_weld", label: "No-Weld COTS", sub: "Bolted / CPVC" },
              { id: "welded_dxf", label: "Welded DXF", sub: "Sheet Metal" },
              { id: "custom_fab_deluxe", label: "Custom Fab Deluxe", sub: "TIG / FRP / Skid" },
              { id: "local_ceramic_masonry", label: "Local Ceramic / Brick", sub: "Nairobi / Mumbai" },
            ] as const
          ).map((bp) => (
            <button
              key={bp.id}
              type="button"
              data-build-path={bp.id}
              className={`seg-btn ${project.build_path === bp.id ? "active" : ""}`}
              onClick={() => patch({ build_path: bp.id })}
            >
              {bp.label}
              <span className="seg-sub">{bp.sub}</span>
            </button>
          ))}
        </div>

        {isLocal ? (
          <div
            id="panel-local-ceramic-settings"
            style={{
              display: "flex",
              flexDirection: "column",
              gap: 8,
              padding: 8,
              borderRadius: 6,
              background: "rgba(245, 158, 11, 0.08)",
              border: "1px solid rgba(245, 158, 11, 0.35)",
            }}
          >
            <label className="checkbox-row">
              <input
                id="in-zero-electricity"
                type="checkbox"
                checked={project.zero_electricity}
                onChange={(e) => patch({ zero_electricity: e.target.checked })}
              />
              <span>Zero Electricity On Site (0 W Off-Grid)</span>
            </label>

            <div className="control">
              <label htmlFor="in-local-region">Local Region &amp; Altitude</label>
              <select
                id="in-local-region"
                value={project.local_region}
                onChange={(e) => patch({ local_region: e.target.value })}
              >
                <option value="nairobi">Nairobi / East Africa (KES · 1,660m)</option>
                <option value="mumbai">Mumbai / India (INR · Sea Level)</option>
                <option value="general_local">General Local Pottery (USD)</option>
              </select>
            </div>

            <div className="control">
              <label htmlFor="in-local-alkali">Local Neutralizing Alkali</label>
              <select
                id="in-local-alkali"
                value={project.local_alkali}
                onChange={(e) => patch({ local_alkali: e.target.value })}
              >
                <option value="soda_ash">Magadi Soda Ash (Na₂CO₃)</option>
                <option value="slaked_lime">Slaked Lime / Chuna (Ca(OH)₂)</option>
                <option value="wood_ash_lye">Kiln Wood-Ash Lye (K₂CO₃ — Free)</option>
                <option value="baking_soda">Bulk Baking Soda (NaHCO₃)</option>
              </select>
            </div>

            <div className="control">
              <label htmlFor="in-local-draft-mode">
                Velocity / Draft Driver (0 W Options)
              </label>
              <select
                id="in-local-draft-mode"
                value={project.local_draft_mode}
                onChange={(e) => patch({ local_draft_mode: e.target.value })}
              >
                <option value="natural_chimney">
                  Natural Reheat Chimney (0 psig · ~30% NaCl)
                </option>
                <option value="steam_ejector_subsonic">
                  Subsonic Quiet Steam-Jet (6 psig · ~52 dBA · ~55% NaCl)
                </option>
                <option value="steam_ejector_lp">
                  LP Steam-Jet Ejector (20 psig · ~62% NaCl)
                </option>
                <option value="steam_ejector_hp">
                  HP Baby Boiler Hydro-Sonic (60 psig · ~80% NaCl)
                </option>
              </select>
            </div>
          </div>
        ) : (
          <>
            <div className="control">
              <label>Corrosion Tier</label>
              <div className="seg-group">
                <button
                  type="button"
                  data-corrosion="standard_316l"
                  className={`seg-btn ${project.corrosion_tier !== "heavy_c276_frp" ? "active" : ""}`}
                  onClick={() => patch({ corrosion_tier: "standard_316l" })}
                >
                  Standard 316L SS
                  <span className="seg-sub">Light / Occasional</span>
                </button>
                <button
                  type="button"
                  data-corrosion="heavy_c276_frp"
                  className={`seg-btn ${project.corrosion_tier === "heavy_c276_frp" ? "active" : ""}`}
                  onClick={() => patch({ corrosion_tier: "heavy_c276_frp" })}
                >
                  Heavy C-276 / FRP
                  <span className="seg-sub">Commercial Salt</span>
                </button>
              </div>
            </div>

            {/* Phase 7 & 8 Stage Options */}
            <div className="control">
              <label>Stage Alignment</label>
              <div className="seg-group">
                <button
                  type="button"
                  className={`seg-btn ${!isEductor ? "active" : ""}`}
                  onClick={() => patch({ stage_alignment: "inline_id_blower" })}
                >
                  Inline ID Blower
                  <span className="seg-sub">Negative Suction</span>
                </button>
                <button
                  type="button"
                  className={`seg-btn ${isEductor ? "active" : ""}`}
                  onClick={() => patch({ stage_alignment: "eductor_dry_fan" })}
                >
                  Dry-Fan Eductor
                  <span className="seg-sub">Zero Acid Fan</span>
                </button>
              </div>
            </div>

            <div className="control">
              <label>Stage 3 Packed-Bed Layout</label>
              <div className="seg-group">
                <button
                  type="button"
                  data-packed-bed="single_column"
                  className={`seg-btn ${project.packed_bed_layout === "single_column" ? "active" : ""}`}
                  onClick={() => patch({ packed_bed_layout: "single_column" })}
                >
                  Single Column
                </button>
                <button
                  type="button"
                  data-packed-bed="parallel_manifold"
                  disabled={isDeluxe}
                  title={
                    isDeluxe
                      ? "Custom Fab Deluxe uses a single filament-wound FRP column with bolted manway"
                      : ""
                  }
                  className={`seg-btn ${project.packed_bed_layout === "parallel_manifold" ? "active" : ""}`}
                  onClick={() => patch({ packed_bed_layout: "parallel_manifold" })}
                >
                  Uniseal Bundle
                </button>
              </div>
            </div>

            {project.packed_bed_layout === "parallel_manifold" && (
              <div className="control">
                <label>Parallel Pipe Size</label>
                <div className="seg-group">
                  {[2, 3, 4].map((sz) => (
                    <button
                      key={sz}
                      type="button"
                      className={`seg-btn ${project.parallel_pipe_size_in === sz ? "active" : ""}`}
                      onClick={() => patch({ parallel_pipe_size_in: sz })}
                    >
                      {sz}&quot; Pipe
                    </button>
                  ))}
                </div>
              </div>
            )}

            <div className="control">
              <label>Stage 2 Venturi Layout</label>
              <div className="seg-group">
                <button
                  type="button"
                  data-venturi-layout="single_throat"
                  className={`seg-btn ${project.venturi_layout === "single_throat" ? "active" : ""}`}
                  onClick={() => patch({ venturi_layout: "single_throat" })}
                >
                  Single Throat
                </button>
                <button
                  type="button"
                  data-venturi-layout="parallel_mini_venturis"
                  disabled={isEductor}
                  title={
                    isEductor
                      ? "Dry-Fan Eductor replaces Stage 2 with a single downflow eductor throat"
                      : ""
                  }
                  className={`seg-btn ${project.venturi_layout === "parallel_mini_venturis" ? "active" : ""}`}
                  onClick={() =>
                    patch({ venturi_layout: "parallel_mini_venturis" })
                  }
                >
                  Parallel Mini-Venturis
                </button>
              </div>
            </div>

            <div className="control">
              <label htmlFor="in-blower-drive">Stage 4 Blower / Motive Drive</label>
              <select
                id="in-blower-drive"
                value={project.blower_drive_option}
                onChange={(e) => patch({ blower_drive_option: e.target.value })}
              >
                <option value="industrial_radial_vfd">
                  Industrial Radial Pressure Blower + VFD
                </option>
                <option value="regen_side_channel">
                  Side-Channel Regenerative Ring Blower
                </option>
                <option value="dual_bypass_blowers" disabled={isEductor}>
                  Dual 3-Stage Tangential Bypass Blowers (2× Ametek)
                </option>
                <option value="compressor_air_atomize" disabled={isEductor}>
                  Shop Compressor Air-Atomize (25 µm) + Low-ΔP Fan
                </option>
                <option value="compressor_eductor_only" disabled={isEductor}>
                  Shop Compressor Main Eductor (Feasibility Check)
                </option>
              </select>
            </div>

            <div className="control">
              <label>Packing Media</label>
              <div className="seg-group">
                <button
                  type="button"
                  className={`seg-btn ${project.packing_media_option === "pp_pall_rings" ? "active" : ""}`}
                  onClick={() => patch({ packing_media_option: "pp_pall_rings" })}
                >
                  PP Pall
                </button>
                <button
                  type="button"
                  className={`seg-btn ${project.packing_media_option === "aquaculture_bioballs" ? "active" : ""}`}
                  onClick={() =>
                    patch({ packing_media_option: "aquaculture_bioballs" })
                  }
                >
                  Bio-Balls
                </button>
                <button
                  type="button"
                  className={`seg-btn ${project.packing_media_option === "diy_ceramic_rings" ? "active" : ""}`}
                  onClick={() =>
                    patch({ packing_media_option: "diy_ceramic_rings" })
                  }
                >
                  DIY Ceramic
                </button>
              </div>
            </div>

            <div className="control">
              <label>Controls &amp; Safety Tier</label>
              <div className="seg-group">
                <button
                  type="button"
                  data-control-tier="rpi_modbus"
                  className={`seg-btn ${project.control_tier === "rpi_modbus" ? "active" : ""}`}
                  onClick={() => patch({ control_tier: "rpi_modbus" })}
                >
                  RPi Modbus + 200°F Safety
                </button>
                <button
                  type="button"
                  data-control-tier="electromechanical_only"
                  className={`seg-btn ${project.control_tier === "electromechanical_only" ? "active" : ""}`}
                  onClick={() =>
                    patch({ control_tier: "electromechanical_only" })
                  }
                >
                  Hardwired Only
                </button>
              </div>
            </div>

            <label className="checkbox-row">
              <input
                id="in-compact-footprint"
                type="checkbox"
                checked={project.compact_footprint}
                onChange={(e) => patch({ compact_footprint: e.target.checked })}
              />
              <span>Compact U-Turn Folded Skid (~47% shorter)</span>
            </label>

            <label className="checkbox-row">
              <input
                id="in-shared-baffled-sump"
                type="checkbox"
                checked={project.shared_baffled_sump}
                onChange={(e) => patch({ shared_baffled_sump: e.target.checked })}
              />
              <span>Shared Baffled Sump (Closed-Loop Reuse)</span>
            </label>

            <label className="checkbox-row">
              <input
                id="in-include-silencer"
                type="checkbox"
                checked={project.include_silencer}
                onChange={(e) => patch({ include_silencer: e.target.checked })}
              />
              <span>Acoustic Silencer &amp; Blower Enclosure</span>
            </label>

            <label className="checkbox-row">
              <input
                id="in-include-plume-diluter"
                type="checkbox"
                checked={project.include_plume_diluter}
                onChange={(e) =>
                  patch({ include_plume_diluter: e.target.checked })
                }
              />
              <span>120V Squirrel-Cage Winter Plume Diluter</span>
            </label>
          </>
        )}
      </div>

      {/* 4. Site & Utilities */}
      <div className="sidebar-section">
        <h3>Site &amp; Utilities</h3>

        <div className="control">
          <div className="control-row">
            <label htmlFor="in-property-line">Property Line Distance</label>
            <span className="control-val">{project.property_line_ft} ft</span>
          </div>
          <input
            id="in-property-line"
            type="range"
            min={5}
            max={150}
            step={5}
            value={project.property_line_ft}
            onChange={(e) =>
              patch({ property_line_ft: parseFloat(e.target.value) })
            }
          />
        </div>

        <div className="control">
          <div className="control-row">
            <label htmlFor="in-ambient-winter">Coldest Firing Ambient</label>
            <span className="control-val">{project.ambient_winter_f} °F</span>
          </div>
          <input
            id="in-ambient-winter"
            type="range"
            min={0}
            max={85}
            step={1}
            value={project.ambient_winter_f}
            onChange={(e) =>
              patch({ ambient_winter_f: parseFloat(e.target.value) })
            }
          />
        </div>

        {!project.zero_electricity && (
          <div className="control">
            <label>Available Studio Voltage</label>
            <div className="seg-group">
              {[120, 230, 240].map((v) => (
                <button
                  key={v}
                  type="button"
                  className={`seg-btn ${project.available_voltage === v ? "active" : ""}`}
                  onClick={() => patch({ available_voltage: v })}
                >
                  {v}V
                </button>
              ))}
            </div>
          </div>
        )}

        <div className="control">
          <div className="control-row">
            <label htmlFor="in-garden-hose">Water Supply (<span className="unit-literal">GPM</span>)</label>
            <span className="control-val">
              {project.garden_hose_gpm.toFixed(1)} GPM
            </span>
          </div>
          <input
            id="in-garden-hose"
            type="range"
            min={0.5}
            max={15}
            step={0.5}
            value={project.garden_hose_gpm}
            onChange={(e) =>
              patch({ garden_hose_gpm: parseFloat(e.target.value) })
            }
          />
        </div>
      </div>

      {/* 5. Collapsed-by-default Advanced Engineering Drawer */}
      <details id="advanced-engineering-drawer" className="adv-drawer">
        <summary>Advanced Engineering Parameters</summary>
        <div style={{ display: "flex", flexDirection: "column", gap: 9 }}>
          <div className="control">
            <div className="control-row">
              <label htmlFor="in-venturi-vel">Venturi Throat Velocity</label>
              <span className="control-val">
                {project.input.venturi_throat_velocity_m_s.toFixed(0)} m/s
              </span>
            </div>
            <input
              id="in-venturi-vel"
              type="range"
              min={35}
              max={105}
              step={1}
              value={project.input.venturi_throat_velocity_m_s}
              onChange={(e) =>
                patchInput({
                  venturi_throat_velocity_m_s: parseFloat(e.target.value),
                })
              }
            />
          </div>

          <div className="control">
            <div className="control-row">
              <label htmlFor="in-venturi-lg">Venturi L/G Ratio</label>
              <span className="control-val">
                {project.input.venturi_lg_ratio.toFixed(4)}
              </span>
            </div>
            <input
              id="in-venturi-lg"
              type="range"
              min={0.0005}
              max={0.003}
              step={0.0001}
              value={project.input.venturi_lg_ratio}
              onChange={(e) =>
                patchInput({ venturi_lg_ratio: parseFloat(e.target.value) })
              }
            />
          </div>

          <div className="control">
            <div className="control-row">
              <label htmlFor="in-hcl-target">Target HCl Removal</label>
              <span className="control-val">
                {(project.input.target_hcl_removal * 100).toFixed(1)}%
              </span>
            </div>
            <input
              id="in-hcl-target"
              type="range"
              min={80}
              max={99.5}
              step={0.5}
              value={project.input.target_hcl_removal * 100}
              onChange={(e) =>
                patchInput({
                  target_hcl_removal: parseFloat(e.target.value) / 100,
                })
              }
            />
          </div>

          <div className="control">
            <div className="control-row">
              <label htmlFor="in-kiln-draft">Target Kiln Draft (<span className="unit-literal">inWC</span>)</label>
              <span className="control-val">
                {project.input.target_kiln_draft_in_wc.toFixed(3)} inWC
              </span>
            </div>
            <input
              id="in-kiln-draft"
              type="range"
              min={-0.3}
              max={-0.01}
              step={0.005}
              value={project.input.target_kiln_draft_in_wc}
              onChange={(e) =>
                patchInput({
                  target_kiln_draft_in_wc: parseFloat(e.target.value),
                })
              }
            />
          </div>

          <div className="control">
            <div className="control-row">
              <label htmlFor="in-compressor-scfm">Shop Air Compressor</label>
              <span className="control-val">
                {project.shop_compressor_scfm.toFixed(0)} SCFM
              </span>
            </div>
            <input
              id="in-compressor-scfm"
              type="range"
              min={5}
              max={40}
              step={1}
              value={project.shop_compressor_scfm}
              onChange={(e) =>
                patch({ shop_compressor_scfm: parseFloat(e.target.value) })
              }
            />
          </div>

          <div className="control">
            <div className="control-row">
              <label htmlFor="in-packing-void">Packing Voidage</label>
              <span className="control-val">
                {project.input.packing_voidage.toFixed(2)}
              </span>
            </div>
            <input
              id="in-packing-void"
              type="range"
              min={0.65}
              max={0.95}
              step={0.01}
              value={project.input.packing_voidage}
              onChange={(e) =>
                patchInput({ packing_voidage: parseFloat(e.target.value) })
              }
            />
          </div>

          <div className="control">
            <div className="control-row">
              <label htmlFor="in-packing-diam">Packing Size</label>
              <span className="control-val">
                {mToIn(project.input.packing_diameter_m || 0.0254).toFixed(2)}&quot;
              </span>
            </div>
            <input
              id="in-packing-diam"
              type="range"
              min={0.5}
              max={2.0}
              step={0.125}
              value={mToIn(project.input.packing_diameter_m || 0.0254)}
              onChange={(e) =>
                patchInput({
                  packing_diameter_m: parseFloat(e.target.value) * 0.0254,
                })
              }
            />
          </div>

          <label className="checkbox-row">
            <input
              id="in-snap-cots"
              type="checkbox"
              checked={project.snap_to_cots}
              onChange={(e) => patch({ snap_to_cots: e.target.checked })}
            />
            <span>Snap Vessels to Standard COTS Pipe Sizes</span>
          </label>
        </div>
      </details>
    </aside>
  );
}
