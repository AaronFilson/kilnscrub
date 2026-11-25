"use client";

import React, { useState } from "react";
import type { StudioPreset, StudioProject } from "@/types/kilnscrub";
import { CONE_OPTIONS, inferConeFromTemp, mToIn, reconcileProject } from "@/lib/api";

interface StudioSidebarProps {
  presets: StudioPreset[];
  project: StudioProject;
  onChange: (next: StudioProject) => void;
  onSelectPreset: (preset: StudioPreset) => void;
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
  onApplyBudgetCombo,
  onApplyDeluxe,
  onApplyOffGrid,
  onApplyOffGridSteam,
}: StudioSidebarProps) {
  const [sidebarMode, setSidebarMode] = useState<"guided" | "all">("guided");
  const [openHardware, setOpenHardware] = useState(false);
  const [openSite, setOpenSite] = useState(false);
  const [openSteamOpts, setOpenSteamOpts] = useState(false);
  const [openGlobalExamples, setOpenGlobalExamples] = useState(false);

  const isAll = sidebarMode === "all";
  const isLocal = project.build_path === "local_ceramic_masonry";
  const isDeluxe = project.build_path === "custom_fab_deluxe";
  const isEductor = project.stage_alignment === "eductor_dry_fan";

  const activePreset = presets.find((p) => p.id === project.preset_id);

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

  const hardwareSummary = [
    project.packed_bed_layout === "parallel_manifold"
      ? `${project.parallel_pipe_size_in}" Bundle`
      : "Single Col",
    project.blower_drive_option === "compressor_air_atomize"
      ? "Air-Atomize"
      : project.blower_drive_option === "regen_side_channel"
        ? "Regen Fan"
        : "Radial VFD",
    project.compact_footprint ? "U-Turn" : "Inline",
  ].join(" · ");

  const siteSummary = `${project.property_line_ft} ft · ${project.ambient_winter_f}°F · ${
    project.zero_electricity ? "0V" : `${project.available_voltage}V`
  }`;

  return (
    <aside className="sidebar" aria-label="Studio & Engineering Configuration">
      {/* Guided vs All Controls Switch */}
      <div className="sidebar-mode-bar" role="group" aria-label="Control Detail Level">
        <button
          type="button"
          className={`mode-switch-btn ${sidebarMode === "guided" ? "active" : ""}`}
          onClick={() => setSidebarMode("guided")}
        >
          Guided Setup
        </button>
        <button
          type="button"
          className={`mode-switch-btn ${sidebarMode === "all" ? "active" : ""}`}
          onClick={() => setSidebarMode("all")}
        >
          All Controls
        </button>
      </div>

      {/* STEP 1: Pick a Starting Point */}
      <div className="sidebar-section">
        <div className="step-header">
          <div className="step-title">
            <span className="step-num">1</span>
            <span>Start From a Kiln Preset</span>
          </div>
        </div>

        <div className="control">
          <select
            id="in-preset-select"
            aria-label="Kiln Preset"
            value={project.preset_id}
            onChange={(e) => {
              const found = presets.find((p) => p.id === e.target.value);
              if (found) onSelectPreset(found);
            }}
          >
            {presets.map((pr) => (
              <option key={pr.id} value={pr.id}>
                {pr.name} ({pr.badge})
              </option>
            ))}
          </select>
          {activePreset && (
            <span className="control-hint">{activePreset.tagline}</span>
          )}
        </div>

        <div className="control">
          <span className="control-hint" style={{ fontWeight: 600 }}>
            Quick Build Profiles
          </span>
          <div className="quick-chips">
            <button
              id="btn-apply-budget-combo"
              type="button"
              className="quick-chip"
              onClick={onApplyBudgetCombo}
            >
              Budget DIY
            </button>
            <button
              id="btn-apply-deluxe-fab"
              type="button"
              className="quick-chip"
              onClick={onApplyDeluxe}
            >
              Pro Deluxe
            </button>
            <button
              id="btn-apply-offgrid-ceramic"
              type="button"
              className="quick-chip"
              onClick={onApplyOffGrid}
            >
              Off-Grid (0 W)
            </button>
            <button
              id="btn-apply-offgrid-steam"
              type="button"
              className="quick-chip"
              onClick={onApplyOffGridSteam}
            >
              Steam-Jet (60 psig)
            </button>
          </div>
        </div>
      </div>

      {/* STEP 2: Your Kiln & Firing */}
      <div className="sidebar-section">
        <div className="step-header">
          <div className="step-title">
            <span className="step-num">2</span>
            <span>Your Kiln &amp; Fuel</span>
          </div>
          <span className="control-val">
            {project.input.kiln_volume_cu_ft.toFixed(0)} cu.ft.
          </span>
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
          <label>Glaze / Vapor Mode</label>
          <div className="seg-group" role="group" aria-label="Vapor Glaze Mode">
            {(
              [
                { id: "none", label: "Clean", sub: "No salt" },
                { id: "soda", label: "Soda", sub: "0 HCl" },
                { id: "salt", label: "Salt", sub: "NaCl / HCl" },
              ] as const
            ).map((opt) => (
              <button
                key={opt.id}
                type="button"
                data-vapor-mode={opt.id}
                className={`seg-btn ${project.vapor_glaze_mode === opt.id ? "active" : ""}`}
                onClick={() => patch({ vapor_glaze_mode: opt.id })}
              >
                {opt.label}
                <span className="seg-sub">{opt.sub}</span>
              </button>
            ))}
          </div>
        </div>

        <div className="seg-grid-2x2">
          <div className="control">
            <label htmlFor="in-firing-cone">Firing Cone</label>
            <select
              id="in-firing-cone"
              value={project.firing_cone}
              onChange={(e) => {
                const chosen = CONE_OPTIONS.find((c) => c.cone === e.target.value);
                patch({
                  firing_cone: e.target.value,
                  input: {
                    ...project.input,
                    peak_exit_temp_f: chosen
                      ? chosen.tempF
                      : project.input.peak_exit_temp_f,
                  },
                });
              }}
            >
              {CONE_OPTIONS.map((c) => (
                <option key={c.cone} value={c.cone}>
                  {c.cone} ({c.tempF}°F)
                </option>
              ))}
            </select>
          </div>

          <div className="control">
            <label htmlFor="in-firing-hours">Firing Length</label>
            <select
              id="in-firing-hours"
              value={project.firing_duration_hours}
              onChange={(e) =>
                patch({ firing_duration_hours: parseFloat(e.target.value) })
              }
            >
              {[6, 8, 10, 12, 16, 20, 24, 36, 48].map((hr) => (
                <option key={hr} value={hr}>
                  {hr} hours
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="control">
          <label htmlFor="in-fuel-type">Fuel Type</label>
          <select
            id="in-fuel-type"
            value={project.input.fuel_type}
            onChange={(e) => {
              const ft = e.target.value;
              const dirty =
                ft.startsWith("coal") || ft.startsWith("peat") || ft.startsWith("biomass_ag");
              const salted = ft.endsWith("_salt");
              onChange(
                reconcileProject({
                  ...project,
                  vapor_glaze_mode: salted
                    ? "salt"
                    : project.vapor_glaze_mode === "salt"
                      ? "none"
                      : project.vapor_glaze_mode,
                  local_afterburner: dirty ? true : project.local_afterburner,
                  input: { ...project.input, fuel_type: ft },
                })
              );
            }}
          >
            <optgroup label="Standard Studio Fuels">
              <option value="wood_salt">Wood + Salt</option>
              <option value="wood">Wood Only</option>
              <option value="gas_salt">Gas + Salt</option>
              <option value="gas">Gas / Electric / Soda</option>
            </optgroup>
            <optgroup label="Coal, Peat &amp; Agricultural Biomass (SO₂ + Tar)">
              <option value="coal_salt">Bituminous Coal + Salt (~2% S)</option>
              <option value="coal">Bituminous Coal Only (~2% S)</option>
              <option value="peat_salt">Peat / Turf + Salt (~0.8% S)</option>
              <option value="peat">Peat / Turf / Andean Turba</option>
              <option value="biomass_ag_salt">Ag-Biomass (PKS / Pomace / Husk) + Salt</option>
              <option value="biomass_ag">Ag-Biomass Only (PKS / Pomace / Husk)</option>
            </optgroup>
          </select>
        </div>

        {project.vapor_glaze_mode === "salt" && (
          <div className="seg-grid-2x2">
            <div className="control">
              <div className="control-row">
                <label htmlFor="in-salt-rate">Salt Rate</label>
                <span className="control-val">
                  {project.input.salt_rate_lbs_per_hr.toFixed(1)} lb/hr
                </span>
              </div>
              <input
                id="in-salt-rate"
                type="range"
                min={0.5}
                max={8}
                step={0.5}
                value={project.input.salt_rate_lbs_per_hr}
                onChange={(e) =>
                  patchInput({ salt_rate_lbs_per_hr: parseFloat(e.target.value) })
                }
              />
            </div>

            <div className="control">
              <div className="control-row">
                <label htmlFor="in-salting-hours">Salt Window</label>
                <span className="control-val">
                  {project.salting_hours.toFixed(1)} hr
                </span>
              </div>
              <input
                id="in-salting-hours"
                type="range"
                min={0.5}
                max={Math.max(1, Math.min(12, project.firing_duration_hours))}
                step={0.5}
                value={project.salting_hours}
                onChange={(e) =>
                  patch({ salting_hours: parseFloat(e.target.value) })
                }
              />
            </div>
          </div>
        )}

        {isAll && (
          <>
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
                <label htmlFor="in-peak-temp">Peak Flue Temperature</label>
                <span className="control-val">
                  {Math.round(project.input.peak_exit_temp_f)} °F
                </span>
              </div>
              <input
                id="in-peak-temp"
                type="range"
                min={1400}
                max={2600}
                step={10}
                value={project.input.peak_exit_temp_f}
                onChange={(e) => {
                  const tf = parseFloat(e.target.value);
                  patch({
                    firing_cone: inferConeFromTemp(tf),
                    input: { ...project.input, peak_exit_temp_f: tf },
                  });
                }}
              />
            </div>
          </>
        )}
      </div>

      {/* STEP 3: Build Style & Power */}
      <div className="sidebar-section">
        <div className="step-header">
          <div className="step-title">
            <span className="step-num">3</span>
            <span>Build Style &amp; Power</span>
          </div>
        </div>

        <div className="seg-grid-2x2" role="group" aria-label="Build Path">
          {(
            [
              { id: "cots_no_weld", label: "No-Weld DIY", sub: "Bolted · CPVC" },
              { id: "welded_dxf", label: "Welded Shop", sub: "Sheet Metal" },
              { id: "custom_fab_deluxe", label: "Pro Deluxe", sub: "TIG · FRP · Quiet" },
              { id: "local_ceramic_masonry", label: "Local Brick", sub: "Off-Grid · 0 W" },
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
            style={{ display: "flex", flexDirection: "column", gap: 9 }}
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
              <label htmlFor="in-local-region">Region &amp; Altitude</label>
              <select
                id="in-local-region"
                value={project.local_region}
                onChange={(e) => patch({ local_region: e.target.value })}
              >
                <option value="nairobi">Nairobi / East Africa (1,660 m · KES)</option>
                <option value="mumbai">Mumbai / Western India (Sea Level · INR)</option>
                <option value="khurja_dhaka">Khurja NCR / Dhaka Coal Hub (INR)</option>
                <option value="kathmandu">Kathmandu / Bhaktapur (1,350 m · NPR)</option>
                <option value="lagos_accra">Lagos / Abeokuta &amp; Accra (PKS · NGN)</option>
                <option value="tonala_puebla">Tonalá / Puebla, Mexico (1,650 m · MXN)</option>
                <option value="cusco_andes">Cusco / Altiplano (3,400 m · PEN)</option>
                <option value="cairo_fez">Fustat Cairo / Fez MENA (EGP)</option>
                <option value="highland_peat">Scottish Highlands / Ireland Peat (GBP)</option>
                <option value="general_local">General Rural / Local Studio (USD)</option>
              </select>
            </div>

            <div className="control">
              <label htmlFor="in-local-draft-mode">Draft Driver</label>
              <select
                id="in-local-draft-mode"
                value={project.local_draft_mode}
                onChange={(e) => patch({ local_draft_mode: e.target.value })}
              >
                <option value="natural_chimney">
                  Natural Brick Chimney (0 psig · ~30% NaCl)
                </option>
                <option value="steam_ejector_subsonic">
                  Quiet Steamer Jet (6 psig · ~52 dBA · ~55% NaCl)
                </option>
                <option value="steam_ejector_lp">
                  LP Steam-Jet Ejector (20 psig · ~62% NaCl)
                </option>
                <option value="steam_ejector_hp">
                  HP Baby Boiler Jet (60 psig · ~80% NaCl)
                </option>
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
                <option value="slaked_lime">Slaked Lime / Chuna / Cal (Ca(OH)₂)</option>
                <option value="trona_kanwa">Trona / Kanwa / Tequesquite (0 Scale)</option>
                <option value="wood_ash_lye">Kiln Wood-Ash Lye (K₂CO₃ — Free)</option>
                <option value="baking_soda">Bulk Baking Soda (NaHCO₃)</option>
              </select>
            </div>

            <label className="checkbox-row">
              <input
                id="in-local-afterburner"
                type="checkbox"
                checked={!!project.local_afterburner}
                onChange={(e) => patch({ local_afterburner: e.target.checked })}
              />
              <span>Stage 0 Masonry Tar Afterburner &amp; Skim Weir (0 W · best for coal/peat/PKS)</span>
            </label>

            {project.local_draft_mode !== "natural_chimney" && (
              <div className="accordion-card">
                <button
                  type="button"
                  className="accordion-trigger"
                  onClick={() => setOpenSteamOpts((v) => !v)}
                >
                  <span style={{ fontSize: 12, fontWeight: 700 }}>
                    Boiler Heat &amp; Water Pump
                  </span>
                  <span className="accordion-meta">
                    {openSteamOpts || isAll ? "Hide ▴" : "Customize ▾"}
                  </span>
                </button>
                {(openSteamOpts || isAll) && (
                  <div className="accordion-body">
                    <div className="control">
                      <label htmlFor="in-local-boiler-source">Steam Heat Source</label>
                      <select
                        id="in-local-boiler-source"
                        value={project.local_boiler_source || "dedicated_hearth"}
                        onChange={(e) => patch({ local_boiler_source: e.target.value })}
                      >
                        <option value="dedicated_hearth">
                          Dedicated Firebox / Separate Burner
                        </option>
                        <option value="stage1_feedwater_preheater">
                          0 psig Vented Stage 1 Preheater (-13% Fuel)
                        </option>
                        <option
                          value="stage1_water_jacket"
                          disabled={project.local_draft_mode === "steam_ejector_subsonic"}
                        >
                          ASME/IBR ≤25 L Stage 1 Water Jacket (20/60 psig)
                        </option>
                        <option
                          value="kiln_exhaust_coil"
                          disabled={project.local_draft_mode !== "steam_ejector_lp"}
                        >
                          Kiln Flue Monotube Coil (&lt;2 L · 20 psig Only)
                        </option>
                      </select>
                    </div>

                    <div className="control">
                      <label htmlFor="in-local-pump-mode">Water Circulation</label>
                      <select
                        id="in-local-pump-mode"
                        value={project.local_pump_mode || "manual_treadle"}
                        disabled={project.local_draft_mode === "steam_ejector_subsonic"}
                        onChange={(e) => patch({ local_pump_mode: e.target.value })}
                      >
                        <option value="manual_treadle">
                          Manual Treadle / Hand Pump + Gravity Tank
                        </option>
                        <option
                          value="sump_pool_no_pump"
                          disabled={project.local_draft_mode !== "steam_ejector_hp"}
                        >
                          Zero-Pump Sump Pools + 2-Stage Sieve (60 psig)
                        </option>
                        <option value="steam_ppcp">
                          Steam PPCP Auto-Recirc (Forbes Marshall, 0 W)
                        </option>
                        <option value="steam_jet_syphon">
                          Cold-Water Steam Syphon (&lt;120°F Drum Only)
                        </option>
                      </select>
                    </div>
                  </div>
                )}
              </div>
            )}

            <div className="accordion-card">
              <button
                type="button"
                className="accordion-trigger"
                onClick={() => setOpenGlobalExamples((v) => !v)}
              >
                <span style={{ fontSize: 12, fontWeight: 700 }}>
                  Global Coal &amp; Peat Examples
                </span>
                <span className="accordion-meta">
                  {openGlobalExamples || isAll ? "Hide ▴" : "4 Hubs ▾"}
                </span>
              </button>
              {(openGlobalExamples || isAll) && (
                <div className="accordion-body">
                  <div className="seg-grid-2x2">
                    <button
                      type="button"
                      className="btn btn-sm"
                      onClick={() =>
                        onChange(
                          reconcileProject({
                            ...project,
                            studio_name: "Khurja Coal-Salt Pottery Cluster",
                            build_path: "local_ceramic_masonry",
                            zero_electricity: true,
                            available_voltage: 0,
                            vapor_glaze_mode: "salt",
                            local_region: "khurja_dhaka",
                            local_alkali: "slaked_lime",
                            local_draft_mode: "steam_ejector_hp",
                            local_pump_mode: "sump_pool_no_pump",
                            local_afterburner: true,
                            input: { ...project.input, fuel_type: "coal_salt", salt_rate_lbs_per_hr: 2.0 },
                          })
                        )
                      }
                    >
                      Khurja Coal+Salt
                    </button>
                    <button
                      type="button"
                      className="btn btn-sm"
                      onClick={() =>
                        onChange(
                          reconcileProject({
                            ...project,
                            studio_name: "Cusco 3,400m Andean Peat Studio",
                            build_path: "local_ceramic_masonry",
                            zero_electricity: true,
                            available_voltage: 0,
                            vapor_glaze_mode: "none",
                            local_region: "cusco_andes",
                            local_alkali: "slaked_lime",
                            local_draft_mode: "steam_ejector_lp",
                            local_pump_mode: "manual_treadle",
                            local_afterburner: true,
                            input: { ...project.input, fuel_type: "peat", salt_rate_lbs_per_hr: 0.0 },
                          })
                        )
                      }
                    >
                      Cusco 3,400m Peat
                    </button>
                    <button
                      type="button"
                      className="btn btn-sm"
                      onClick={() =>
                        onChange(
                          reconcileProject({
                            ...project,
                            studio_name: "Lagos / Accra PKS + Kanwa Studio",
                            build_path: "local_ceramic_masonry",
                            zero_electricity: true,
                            available_voltage: 0,
                            vapor_glaze_mode: "salt",
                            local_region: "lagos_accra",
                            local_alkali: "trona_kanwa",
                            local_draft_mode: "steam_ejector_subsonic",
                            local_pump_mode: "manual_treadle",
                            local_afterburner: true,
                            input: { ...project.input, fuel_type: "biomass_ag_salt", salt_rate_lbs_per_hr: 1.5 },
                          })
                        )
                      }
                    >
                      Lagos PKS+Kanwa
                    </button>
                    <button
                      type="button"
                      className="btn btn-sm"
                      onClick={() =>
                        onChange(
                          reconcileProject({
                            ...project,
                            studio_name: "Highland Peat & Seaweed-Salt Kiln",
                            build_path: "local_ceramic_masonry",
                            zero_electricity: true,
                            available_voltage: 0,
                            vapor_glaze_mode: "salt",
                            local_region: "highland_peat",
                            local_alkali: "wood_ash_lye",
                            local_draft_mode: "natural_chimney",
                            local_pump_mode: "manual_treadle",
                            local_afterburner: true,
                            ambient_winter_f: 36,
                            input: { ...project.input, fuel_type: "peat_salt", salt_rate_lbs_per_hr: 1.5 },
                          })
                        )
                      }
                    >
                      Highland Peat+Salt
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        ) : (
          <>
            {isDeluxe && (() => {
              const vfdCompatible =
                project.stage_alignment === "inline_id_blower" &&
                project.venturi_layout === "single_throat" &&
                (!project.blower_drive_option ||
                  project.blower_drive_option === "industrial_radial_vfd" ||
                  project.blower_drive_option === "compressor_air_atomize");
              return (
                <div
                  id="panel-deluxe-eco-settings"
                  style={{
                    display: "flex",
                    flexDirection: "column",
                    gap: 6,
                    padding: 9,
                    borderRadius: 8,
                    background: "rgba(16, 185, 129, 0.08)",
                    border: "1px solid rgba(16, 185, 129, 0.3)",
                  }}
                >
                  <strong style={{ fontSize: 11, color: "#6ee7b7", textTransform: "uppercase" }}>
                    Deluxe Energy &amp; Water Savers
                  </strong>
                  <label className="checkbox-row">
                    <input
                      id="in-deluxe-condensing"
                      type="checkbox"
                      checked={!!project.deluxe_condensing_reclaimer}
                      onChange={(e) => patch({ deluxe_condensing_reclaimer: e.target.checked })}
                    />
                    <span>Ceramic / SiC Condensing Plume Reclaimer (150°F)</span>
                  </label>
                  <label
                    className="checkbox-row"
                    title={!vfdCompatible ? "Requires Inline ID Blower + Single Throat + VFD drive" : ""}
                  >
                    <input
                      id="in-deluxe-var-throat"
                      type="checkbox"
                      disabled={!vfdCompatible}
                      checked={!!project.deluxe_variable_throat_vfd}
                      onChange={(e) => patch({ deluxe_variable_throat_vfd: e.target.checked })}
                    />
                    <span style={!vfdCompatible ? { opacity: 0.5 } : undefined}>
                      Variable-Throat Venturi + VFD Turndown (-60% kWh)
                    </span>
                  </label>
                  <label className="checkbox-row">
                    <input
                      id="in-deluxe-s1-heat"
                      type="checkbox"
                      checked={!!project.deluxe_stage1_heat_recovery}
                      onChange={(e) => patch({ deluxe_stage1_heat_recovery: e.target.checked })}
                    />
                    <span>Stage 1 Water-Jacket Pre-Cooler (-35% water)</span>
                  </label>
                </div>
              );
            })()}

            <div className="control">
              <label>Metal &amp; Corrosion Grade</label>
              <div className="seg-group">
                <button
                  type="button"
                  data-corrosion="standard_316l"
                  className={`seg-btn ${project.corrosion_tier !== "heavy_c276_frp" ? "active" : ""}`}
                  onClick={() => patch({ corrosion_tier: "standard_316l" })}
                >
                  Standard 316L SS
                  <span className="seg-sub">Studio Use</span>
                </button>
                <button
                  type="button"
                  data-corrosion="heavy_c276_frp"
                  className={`seg-btn ${project.corrosion_tier === "heavy_c276_frp" ? "active" : ""}`}
                  onClick={() => patch({ corrosion_tier: "heavy_c276_frp" })}
                >
                  Heavy C-276 / FRP
                  <span className="seg-sub">Heavy Salt</span>
                </button>
              </div>
            </div>
          </>
        )}
      </div>

      {/* Collapsible Hardware & Layout Fine-Tuning (Non-Local Modes) */}
      {!isLocal && (
        <div className="accordion-card">
          <button
            type="button"
            className="accordion-trigger"
            onClick={() => setOpenHardware((v) => !v)}
          >
            <span className="step-title">
              <span className="step-num">+</span>
              <span>Hardware &amp; Layout</span>
            </span>
            <span className="accordion-meta">
              {openHardware || isAll ? "Hide ▴" : hardwareSummary}
            </span>
          </button>
          {(openHardware || isAll) && (
            <div className="accordion-body">
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
                <label>Stage 3 Absorber Layout</label>
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
                    className={`seg-btn ${project.venturi_layout === "parallel_mini_venturis" ? "active" : ""}`}
                    onClick={() => patch({ venturi_layout: "parallel_mini_venturis" })}
                  >
                    Mini-Venturis
                  </button>
                </div>
              </div>

              <div className="control">
                <label htmlFor="in-blower-drive">Blower / Motive Drive</label>
                <select
                  id="in-blower-drive"
                  value={project.blower_drive_option}
                  onChange={(e) => patch({ blower_drive_option: e.target.value })}
                >
                  <option value="industrial_radial_vfd">
                    Industrial Radial Blower + VFD
                  </option>
                  <option value="regen_side_channel">
                    Side-Channel Regenerative Blower
                  </option>
                  <option value="dual_bypass_blowers" disabled={isEductor}>
                    Dual 3-Stage Tangential Bypass Blowers
                  </option>
                  <option value="compressor_air_atomize" disabled={isEductor}>
                    Shop Compressor Air-Atomize (25 µm)
                  </option>
                  <option value="compressor_eductor_only" disabled={isEductor}>
                    Shop Compressor Eductor (Feasibility)
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
                    onClick={() => patch({ packing_media_option: "aquaculture_bioballs" })}
                  >
                    Bio-Balls
                  </button>
                  <button
                    type="button"
                    className={`seg-btn ${project.packing_media_option === "diy_ceramic_rings" ? "active" : ""}`}
                    onClick={() => patch({ packing_media_option: "diy_ceramic_rings" })}
                  >
                    DIY Ceramic
                  </button>
                </div>
              </div>

              <div className="control">
                <label>Controls &amp; Safety</label>
                <div className="seg-group">
                  <button
                    type="button"
                    data-control-tier="rpi_modbus"
                    className={`seg-btn ${project.control_tier === "rpi_modbus" ? "active" : ""}`}
                    onClick={() => patch({ control_tier: "rpi_modbus" })}
                  >
                    RPi + 200°F Safety
                  </button>
                  <button
                    type="button"
                    data-control-tier="electromechanical_only"
                    className={`seg-btn ${project.control_tier === "electromechanical_only" ? "active" : ""}`}
                    onClick={() => patch({ control_tier: "electromechanical_only" })}
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
                  onChange={(e) => patch({ include_plume_diluter: e.target.checked })}
                />
                <span>120V Winter Plume Diluter Fan</span>
              </label>
            </div>
          )}
        </div>
      )}

      {/* STEP 4: Collapsible Backyard & Site Utilities */}
      <div className="accordion-card">
        <button
          type="button"
          className="accordion-trigger"
          onClick={() => setOpenSite((v) => !v)}
        >
          <span className="step-title">
            <span className="step-num">4</span>
            <span>Backyard &amp; Utilities</span>
          </span>
          <span className="accordion-meta">
            {openSite || isAll ? "Hide ▴" : siteSummary}
          </span>
        </button>
        {(openSite || isAll) && (
          <div className="accordion-body">
            <div className="control">
              <div className="control-row">
                <label htmlFor="in-property-line">Distance to Fence</label>
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
                <label htmlFor="in-ambient-winter">Coldest Firing Temp</label>
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
        )}
      </div>

      {/* STEP 5: Collapsed Advanced Engineering Parameters */}
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
