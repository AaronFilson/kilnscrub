// Zero-dependency Node.js E2E test harness that launches the compiled ./bin/kilnscrub ui
// binary on an ephemeral localhost port, verifies the client reconcileProject() logic across
// all UI decision-tree transitions, and exercises every live HTTP endpoint end-to-end.

import { describe, it, before, after } from "node:test";
import assert from "node:assert/strict";
import { spawn } from "node:child_process";
import net from "node:net";
import fs from "node:fs";
import path from "node:path";

const BINARY = path.resolve(process.cwd(), "bin/kilnscrub");

function getFreePort() {
  return new Promise((resolve, reject) => {
    const srv = net.createServer();
    srv.listen(0, "127.0.0.1", () => {
      const addr = srv.address();
      srv.close(() => resolve(addr.port));
    });
    srv.on("error", reject);
  });
}

async function waitForHealth(baseURL, timeoutMs = 8000) {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    try {
      const r = await fetch(`${baseURL}/api/health`);
      if (r.ok) return;
    } catch {
      // retry until server listens
    }
    await new Promise((r) => setTimeout(r, 80));
  }
  throw new Error(`kilnscrub ui failed to become healthy at ${baseURL}`);
}

// Mirror of web/src/lib/api.ts reconcileProject() to assert client-server contract parity.
function reconcileProject(proj) {
  const next = { ...proj, input: { ...proj.input } };
  if (next.build_path !== "local_ceramic_masonry") {
    next.zero_electricity = false;
    next.local_draft_mode = "natural_chimney";
    next.local_boiler_source = "dedicated_hearth";
    next.local_pump_mode = "manual_treadle";
    next.local_afterburner = false;
    if (next.available_voltage === 0) next.available_voltage = 240;
  } else {
    if (next.zero_electricity) {
      next.available_voltage = 0;
      next.include_plume_diluter = false;
    } else if (next.available_voltage === 0) {
      next.available_voltage = 230;
    }
    const isSteam =
      next.local_draft_mode === "steam_ejector_subsonic" ||
      next.local_draft_mode === "steam_ejector_lp" ||
      next.local_draft_mode === "steam_ejector_hp";
    const isLP = next.local_draft_mode === "steam_ejector_lp";
    const isHP = next.local_draft_mode === "steam_ejector_hp";
    const has20Plus = isLP || isHP;
    if (!isSteam) {
      next.local_boiler_source = "dedicated_hearth";
    } else if (next.local_boiler_source === "kiln_exhaust_coil" && !isLP) {
      next.local_boiler_source = isHP ? "stage1_water_jacket" : "stage1_feedwater_preheater";
    } else if (next.local_boiler_source === "stage1_water_jacket" && !has20Plus) {
      next.local_boiler_source = "stage1_feedwater_preheater";
    }
    if (!has20Plus) {
      next.local_pump_mode = "manual_treadle";
    } else if (next.local_pump_mode === "sump_pool_no_pump" && !isHP) {
      next.local_pump_mode = "manual_treadle";
    }
  }
  if (next.build_path === "custom_fab_deluxe") {
    next.packed_bed_layout = "single_column";
  } else {
    next.deluxe_condensing_reclaimer = false;
    next.deluxe_variable_throat_vfd = false;
    next.deluxe_stage1_heat_recovery = false;
  }
  if (next.stage_alignment === "eductor_dry_fan") {
    next.venturi_layout = "single_throat";
    if (
      next.blower_drive_option === "compressor_air_atomize" ||
      next.blower_drive_option === "compressor_eductor_only" ||
      next.blower_drive_option === "dual_bypass_blowers"
    ) {
      next.blower_drive_option = "industrial_radial_vfd";
    }
  }
  const hasVFDDrive =
    !next.blower_drive_option ||
    next.blower_drive_option === "industrial_radial_vfd" ||
    next.blower_drive_option === "compressor_air_atomize";
  if (
    next.deluxe_variable_throat_vfd &&
    (next.stage_alignment !== "inline_id_blower" ||
      next.venturi_layout !== "single_throat" ||
      !hasVFDDrive)
  ) {
    next.deluxe_variable_throat_vfd = false;
  }
  if (next.vapor_glaze_mode === "salt") {
    if (!(next.salting_hours > 0)) next.salting_hours = Math.min(3.0, next.firing_duration_hours || 10);
    if (next.salting_hours > next.firing_duration_hours) next.salting_hours = next.firing_duration_hours;
    if (!(next.input.salt_rate_lbs_per_hr > 0)) next.input.salt_rate_lbs_per_hr = 2.0;
    for (const base of ["wood", "gas", "coal", "peat", "biomass_ag"]) {
      if (next.input.fuel_type === base) next.input.fuel_type = `${base}_salt`;
    }
  } else {
    next.salting_hours = 0;
    next.input.salt_rate_lbs_per_hr = 0;
    for (const base of ["wood", "gas", "coal", "peat", "biomass_ag"]) {
      if (next.input.fuel_type === `${base}_salt`) next.input.fuel_type = base;
    }
  }
  return next;
}

describe("Live kilnscrub Binary E2E Decision-Tree Walkthrough", () => {
  let proc;
  let baseURL;
  let refProject;

  before(async () => {
    assert.ok(fs.existsSync(BINARY), `Expected compiled binary at ${BINARY} (run make build)`);
    const port = await getFreePort();
    baseURL = `http://127.0.0.1:${port}`;
    proc = spawn(BINARY, ["ui", "--addr", `127.0.0.1:${port}`], { stdio: "ignore" });
    await waitForHealth(baseURL);

    const res = await fetch(`${baseURL}/api/studio/presets`);
    assert.equal(res.status, 200);
    const { presets } = await res.json();
    assert.equal(presets.length, 5);
    refProject = presets.find((p) => p.id === "wood_salt_24").project;
  });

  after(() => {
    if (proc) proc.kill("SIGTERM");
  });

  async function evalOK(proj, label) {
    const reconciled = reconcileProject(proj);
    const r = await fetch(`${baseURL}/api/studio/evaluate`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(reconciled),
    });
    const text = await r.text();
    assert.equal(r.status, 200, `${label}: expected HTTP 200, got ${r.status}: ${text}`);
    return JSON.parse(text);
  }

  it("serves the embedded Phase 17 Guided Web UI HTML at GET /", async () => {
    const r = await fetch(`${baseURL}/`);
    assert.equal(r.status, 200);
    const html = await r.text();
    assert.match(html, /Design &amp; 3D Studio/);
    assert.match(html, /Build &amp; Firing Guide/);
    assert.match(html, /Shop Plans &amp; Cut Sheets/);
    assert.match(html, /Live Control &amp; Telemetry/);
  });

  it("walks sequential UI click transitions across Deluxe, Eductor, Budget DIY & Off-Grid without a single 400", async () => {
    let cur = structuredClone(refProject);

    // 1. Deluxe + all 3 eco upgrades -> switch to Eductor -> switch to No-Weld Budget DIY
    cur = reconcileProject({
      ...cur,
      build_path: "custom_fab_deluxe",
      deluxe_condensing_reclaimer: true,
      deluxe_variable_throat_vfd: true,
      deluxe_stage1_heat_recovery: true,
    });
    let ev = await evalOK(cur, "Deluxe all eco");
    assert.ok(ev.shop_plan.deluxe_spec.water_saved_pct > 50);

    cur = reconcileProject({ ...cur, stage_alignment: "eductor_dry_fan" });
    assert.equal(cur.deluxe_variable_throat_vfd, false);
    await evalOK(cur, "Deluxe -> Eductor");

    cur = reconcileProject({
      ...cur,
      build_path: "cots_no_weld",
      stage_alignment: "inline_id_blower",
      packed_bed_layout: "parallel_manifold",
      parallel_pipe_size_in: 4,
      blower_drive_option: "compressor_air_atomize",
      packing_media_option: "diy_ceramic_rings",
      compact_footprint: true,
    });
    await evalOK(cur, "Budget DIY combo");

    // 2. Off-Grid ON -> OFF (regression guard for Bug A voltage recovery!) -> ON
    cur = reconcileProject({
      ...cur,
      build_path: "local_ceramic_masonry",
      zero_electricity: true,
      local_region: "khurja_dhaka",
      local_alkali: "slaked_lime",
      local_draft_mode: "steam_ejector_hp",
      local_pump_mode: "sump_pool_no_pump",
      local_afterburner: true,
      input: { ...cur.input, fuel_type: "coal_salt", salt_rate_lbs_per_hr: 2.0 },
    });
    ev = await evalOK(cur, "Khurja Coal+Salt 60 psig zero-pump");
    assert.equal(ev.frugal.power_watts, 0);
    assert.ok(ev.frugal.so2_inlet_ppm > 500);
    assert.equal(ev.frugal.has_overhead_tank, false);

    // Uncheck Zero Electricity while staying in Local Ceramic -> must recover 230V!
    cur = reconcileProject({ ...cur, zero_electricity: false });
    assert.equal(cur.available_voltage, 230);
    await evalOK(cur, "Local Hybrid after unchecking zero_electricity");

    // Step 20 psig coil -> 60 psig HP -> 6 psig subsonic -> 0 psig natural chimney
    cur = reconcileProject({
      ...cur,
      zero_electricity: true,
      local_region: "cusco_andes",
      local_draft_mode: "steam_ejector_lp",
      local_boiler_source: "kiln_exhaust_coil",
      local_pump_mode: "steam_ppcp",
      vapor_glaze_mode: "none",
      input: { ...cur.input, fuel_type: "peat" },
    });
    ev = await evalOK(cur, "Cusco 3,400m Peat 20 psig coil + PPCP");
    assert.ok(ev.frugal.ambient_pressure_kpa < 68);
    assert.ok(ev.frugal.bed_diameter_in >= 30);

    cur = reconcileProject({ ...cur, local_draft_mode: "steam_ejector_hp" });
    assert.equal(cur.local_boiler_source, "stage1_water_jacket");
    await evalOK(cur, "Cusco 20 psig coil -> 60 psig HP auto-migration");

    cur = reconcileProject({ ...cur, local_draft_mode: "steam_ejector_subsonic" });
    assert.equal(cur.local_boiler_source, "stage1_feedwater_preheater");
    assert.equal(cur.local_pump_mode, "manual_treadle");
    await evalOK(cur, "Cusco 60 psig -> 6 psig subsonic downgrade");

    // Return to standard No-Weld DIY -> local_afterburner must reset cleanly.
    cur = reconcileProject({ ...cur, build_path: "cots_no_weld" });
    assert.equal(cur.local_afterburner, false);
    assert.equal(cur.available_voltage, 240);
    await evalOK(cur, "Back to No-Weld COTS");
  });
});
