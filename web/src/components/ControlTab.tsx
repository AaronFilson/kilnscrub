"use client";

import React, { useEffect, useRef, useState } from "react";
import type { ControlConfig, ControlStatus } from "@/types/kilnscrub";
import {
  fetchControlStatus,
  fmt,
  kToF,
  postControlFault,
  postControlReset,
  postControlSetpoint,
} from "@/lib/api";

interface TelemetrySample {
  draftInWC: number;
  quenchOutF: number;
  vfdPct: number;
  valvePct: number;
}

interface ControlTabProps {
  active: boolean;
  onStatusDotChange: (state: "off" | "ok" | "warn" | "tripped") => void;
  onToast: (msg: string) => void;
}

const MAX_SAMPLES = 90;

export function ControlTab({
  active,
  onStatusDotChange,
  onToast,
}: ControlTabProps) {
  const [enabled, setEnabled] = useState(false);
  const [halMode, setHalMode] = useState("sim");
  const [supportsFault, setSupportsFault] = useState(false);
  const [connState, setConnState] = useState<"offline" | "connecting" | "live">("offline");
  const [status, setStatus] = useState<ControlStatus | null>(null);
  const [config, setConfig] = useState<ControlConfig | null>(null);
  const [history, setHistory] = useState<TelemetrySample[]>([]);
  const [draftInput, setDraftInput] = useState(-0.05);
  const [quenchInputF, setQuenchInputF] = useState(188);
  const [busy, setBusy] = useState(false);
  const [errorMsg, setErrorMsg] = useState("");
  const sseRef = useRef<EventSource | null>(null);
  const syncedRef = useRef(false);

  const reportDot = (st: ControlStatus | null, isEn: boolean) => {
    if (!isEn || !st) {
      onStatusDotChange("off");
      return;
    }
    const tripped = !!(
      st.interlock_active ||
      st.commands?.bypass_damper_open ||
      st.state === "TRIPPED"
    );
    if (tripped) onStatusDotChange("tripped");
    else if (st.state === "WARNING") onStatusDotChange("warn");
    else onStatusDotChange("ok");
  };

  useEffect(() => {
    let cancelled = false;
    fetchControlStatus().then((res) => {
      if (cancelled) return;
      setEnabled(!!res.enabled);
      setHalMode(res.hal_mode || "sim");
      setSupportsFault(!!res.supports_fault_injection);
      if (res.status) {
        setStatus(res.status);
        reportDot(res.status, !!res.enabled);
      }
      if (res.config) {
        setConfig(res.config);
        if (!syncedRef.current) {
          syncedRef.current = true;
          setDraftInput(res.config.target_draft_in_wc ?? -0.05);
          if (res.config.quench_setpoint_k > 273.15) {
            setQuenchInputF(Math.round(kToF(res.config.quench_setpoint_k)));
          }
        }
      }
    });
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (!enabled) {
      sseRef.current?.close();
      sseRef.current = null;
      setConnState("offline");
      return;
    }

    setConnState("connecting");
    const es = new EventSource("/api/control/stream");
    sseRef.current = es;

    es.onopen = () => {
      setConnState("live");
      setErrorMsg("");
    };
    es.onmessage = (ev) => {
      try {
        const st = JSON.parse(ev.data) as ControlStatus;
        setStatus(st);
        setConnState("live");
        reportDot(st, true);
        setHistory((prev) => {
          const next = [
            ...prev,
            {
              draftInWC: st.filtered_draft_in_wc,
              quenchOutF: kToF(st.readings.quench_outlet_temp_k),
              vfdPct: st.commands.blower_vfd * 100,
              valvePct: st.commands.quench_spray_valve * 100,
            },
          ];
          return next.length > MAX_SAMPLES ? next.slice(next.length - MAX_SAMPLES) : next;
        });
      } catch {
        // ignore malformed event
      }
    };
    es.onerror = () => {
      setConnState(es.readyState === EventSource.CLOSED ? "offline" : "connecting");
    };

    return () => {
      es.close();
      sseRef.current = null;
    };
  }, [enabled]);

  const handleReset = async () => {
    setBusy(true);
    try {
      const st = await postControlReset();
      setStatus(st);
      setErrorMsg("");
      reportDot(st, true);
      onToast("Interlock reset — closed-loop control resumed");
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      setErrorMsg(msg);
      onToast(msg);
    } finally {
      setBusy(false);
    }
  };

  const handleApplySetpoints = async () => {
    setBusy(true);
    try {
      const cfg = await postControlSetpoint({
        target_draft_in_wc: draftInput,
        quench_setpoint_f: quenchInputF,
      });
      setConfig(cfg);
      setErrorMsg("");
      onToast(`Setpoints updated: ${draftInput.toFixed(2)} inWC, ${quenchInputF}°F`);
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      setErrorMsg(msg);
      onToast(msg);
    } finally {
      setBusy(false);
    }
  };

  const handleFault = async (action: string, value?: number) => {
    setBusy(true);
    try {
      const st = await postControlFault(action, value);
      setStatus(st);
      setErrorMsg("");
      reportDot(st, true);
      onToast(`Simulation action applied: ${action}`);
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      setErrorMsg(msg);
      onToast(msg);
    } finally {
      setBusy(false);
    }
  };

  const maxTempF =
    config?.safety?.max_quench_outlet_temp_k && config.safety.max_quench_outlet_temp_k > 273.15
      ? kToF(config.safety.max_quench_outlet_temp_k)
      : 200.0;

  const isTripped = !!(
    status &&
    (status.interlock_active ||
      status.commands.bypass_damper_open ||
      status.state === "TRIPPED")
  );

  const renderStripChart = (
    title: string,
    vals: number[],
    yMin: number,
    yMax: number,
    color: string,
    refLine?: { val: number; label: string; color: string }
  ) => {
    const W = 400;
    const H = 140;
    const mL = 40;
    const mR = 12;
    const mT = 14;
    const mB = 22;
    const pW = W - mL - mR;
    const pH = H - mT - mB;
    const yOf = (v: number) =>
      mT + pH - ((Math.min(yMax, Math.max(yMin, v)) - yMin) / Math.max(1e-6, yMax - yMin)) * pH;
    const d = vals
      .map((v, i) => {
        const x = mL + (vals.length > 1 ? (i / (vals.length - 1)) * pW : pW);
        return `${i === 0 ? "M" : "L"}${x.toFixed(1)},${yOf(v).toFixed(1)}`;
      })
      .join(" ");

    return (
      <div className="card">
        <div className="card-header" style={{ marginBottom: 6 }}>
          <h3 className="card-title">{title}</h3>
          <span className="pill info">
            {vals.length > 0 ? fmt(vals[vals.length - 1], 2) : "—"}
          </span>
        </div>
        <svg viewBox={`0 0 ${W} ${H}`} width="100%" height={H}>
          <rect x={mL} y={mT} width={pW} height={pH} fill="#0a1320" rx="4" />
          <text x={mL - 5} y={mT + 8} textAnchor="end" fill="#647c99" fontSize="9" fontFamily="monospace">
            {yMax}
          </text>
          <text x={mL - 5} y={mT + pH} textAnchor="end" fill="#647c99" fontSize="9" fontFamily="monospace">
            {yMin}
          </text>
          {refLine && (
            <>
              <line
                x1={mL}
                y1={yOf(refLine.val)}
                x2={mL + pW}
                y2={yOf(refLine.val)}
                stroke={refLine.color}
                strokeDasharray="4 3"
              />
              <text
                x={mL + pW - 4}
                y={yOf(refLine.val) - 3}
                textAnchor="end"
                fill={refLine.color}
                fontSize="9"
                fontFamily="monospace"
              >
                {refLine.label}
              </text>
            </>
          )}
          {d && <path d={d} fill="none" stroke={color} strokeWidth="2.2" />}
        </svg>
      </div>
    );
  };

  if (!enabled) {
    return (
      <div className="page-container">
        <div className="card">
          <h2 style={{ marginTop: 0 }}>Live Control Backend Disabled</h2>
          <p style={{ color: "var(--text-muted)" }}>
            Start <code>kilnscrub ui</code> without <code>--no-control</code> (or pass{" "}
            <code>--config configs/kilnscrub.example.json --hal modbus</code>) to stream live SSE
            telemetry and manage PID setpoints and safety interlocks.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="page-container" style={{ display: active ? "flex" : "none" }}>
      {/* Status Header & Fail-Safe Interlock Banner */}
      <div className={`banner ${isTripped ? "banner-danger" : "banner-ok"}`}>
        <div className="card-header" style={{ marginBottom: 4 }}>
          <div style={{ display: "flex", gap: 8, alignItems: "center", flexWrap: "wrap" }}>
            <strong style={{ fontSize: 15 }}>
              {isTripped
                ? "INTERLOCK TRIPPED — BYPASS DAMPER OPEN"
                : status?.state === "WARNING"
                  ? "WARNING — CHECK SUMP pH / READINGS"
                  : "CLOSED-LOOP CONTROL NORMAL"}
            </strong>
            <span className={`pill ${connState === "live" ? "ok" : "warn"}`}>
              SSE {connState.toUpperCase()}
            </span>
            <span className={`pill ${halMode === "modbus" ? "warn" : "info"}`}>
              HAL: {halMode.toUpperCase()}
            </span>
          </div>
          <button
            id="btn-ctrl-reset"
            type="button"
            className="btn btn-danger"
            disabled={busy || !enabled}
            onClick={handleReset}
          >
            Reset Interlock
          </button>
        </div>
        <div style={{ fontSize: 12.5, color: "#e2e8f0" }}>
          {isTripped
            ? `Trip Reason: ${status?.commands.trip_reason || "INTERLOCK_ACTIVE"} · Blower VFD 0%, Spray Valve 100%, Bypass Damper OPEN. Clear active fault before pressing Reset Interlock.`
            : `Aligned hardwired + software CPVC cutoff: ${fmt(maxTempF, 0)}°F · Min quench flow: 0.25 GPM · De-energize-to-trip fail-safe polarity active.`}
        </div>
        {errorMsg && (
          <div style={{ marginTop: 6, color: "#fca5a5", fontSize: 12 }}>
            Error: {errorMsg}
          </div>
        )}
      </div>

      {/* Live Telemetry KPIs */}
      {status && (
        <section className="kpi-grid">
          <div className="kpi-card">
            <div className="kpi-label">Filtered Kiln Draft</div>
            <div className="kpi-value">{fmt(status.filtered_draft_in_wc, 3)} inWC</div>
            <div className="kpi-sub">{fmt(status.filtered_draft_pa, 2)} Pa</div>
          </div>
          <div className="kpi-card" style={{ ["--kpi-accent" as string]: "var(--accent-quench)" }}>
            <div className="kpi-label">Quench Outlet Temp</div>
            <div className="kpi-value">{fmt(kToF(status.readings.quench_outlet_temp_k), 1)} °F</div>
            <div className="kpi-sub">Cutoff: {fmt(maxTempF, 0)}°F</div>
          </div>
          <div className="kpi-card">
            <div className="kpi-label">Quench Water Flow</div>
            <div className="kpi-value">{fmt(status.readings.water_flow_gpm, 2)} GPM</div>
            <div className="kpi-sub">Min trip: 0.25 GPM</div>
          </div>
          <div className="kpi-card" style={{ ["--kpi-accent" as string]: "var(--accent-packed)" }}>
            <div className="kpi-label">Sump pH</div>
            <div className="kpi-value">{fmt(status.readings.sump_ph, 2)}</div>
            <div className="kpi-sub">Target band: 6.5 – 9.5</div>
          </div>
          <div className="kpi-card">
            <div className="kpi-label">VFD / Spray Valve</div>
            <div className="kpi-value">
              {fmt(status.commands.blower_vfd * 100, 0)}% / {fmt(status.commands.quench_spray_valve * 100, 0)}%
            </div>
            <div className="kpi-sub">
              Bypass: {status.commands.bypass_damper_open ? "OPEN" : "CLOSED"}
            </div>
          </div>
        </section>
      )}

      {/* Rolling Strip Charts */}
      {active && (
        <section className="charts-grid">
          {renderStripChart(
            "Kiln Draft (inWC)",
            history.map((s) => s.draftInWC),
            -0.3,
            0.05,
            "#38bdf8",
            { val: config?.target_draft_in_wc ?? -0.05, label: "Target", color: "#10b981" }
          )}
          {renderStripChart(
            "Quench Outlet Temp (°F)",
            history.map((s) => s.quenchOutF),
            120,
            230,
            "#f97316",
            { val: maxTempF, label: `${Math.round(maxTempF)}°F Trip`, color: "#ef4444" }
          )}
          {renderStripChart(
            "Blower VFD Output (%)",
            history.map((s) => s.vfdPct),
            0,
            100,
            "#10b981"
          )}
        </section>
      )}

      {/* Live Setpoints & Sim Fault Injection */}
      <section className="grid-2">
        <div className="card">
          <h3 className="card-title" style={{ marginBottom: 10 }}>
            Live Controller Setpoints
          </h3>
          <div className="control" style={{ marginBottom: 10 }}>
            <div className="control-row">
              <label htmlFor="in-ctrl-draft">Target Kiln Draft (inWC)</label>
              <span className="control-val">{draftInput.toFixed(3)} inWC</span>
            </div>
            <input
              id="in-ctrl-draft"
              type="range"
              min={-0.3}
              max={-0.01}
              step={0.005}
              value={draftInput}
              onChange={(e) => setDraftInput(parseFloat(e.target.value))}
            />
          </div>
          <div className="control" style={{ marginBottom: 12 }}>
            <div className="control-row">
              <label htmlFor="in-ctrl-quench">Quench Outlet Setpoint (°F)</label>
              <span className="control-val">{quenchInputF} °F</span>
            </div>
            <input
              id="in-ctrl-quench"
              type="range"
              min={150}
              max={195}
              step={1}
              value={quenchInputF}
              onChange={(e) => setQuenchInputF(parseFloat(e.target.value))}
            />
          </div>
          <button
            id="btn-apply-setpoints"
            type="button"
            className="btn btn-primary"
            disabled={busy}
            onClick={handleApplySetpoints}
          >
            Apply Setpoints
          </button>
        </div>

        <div className="card">
          <h3 className="card-title" style={{ marginBottom: 10 }}>
            {supportsFault ? "Simulation Fault Injection Drill" : "Physical Modbus RTU Hardware Active"}
          </h3>
          {supportsFault ? (
            <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
              <button type="button" className="btn btn-danger" disabled={busy} onClick={() => handleFault("heat_spike", 220)}>
                Inject Heat Spike (220°F)
              </button>
              <button type="button" className="btn btn-danger" disabled={busy} onClick={() => handleFault("water_loss")}>
                Cut Quench Water
              </button>
              <button type="button" className="btn" disabled={busy} onClick={() => handleFault("clear_water_loss")}>
                Restore Water
              </button>
              <button type="button" className="btn btn-danger" disabled={busy} onClick={() => handleFault("sensor_fault")}>
                Inject Sensor Fault
              </button>
              <button type="button" className="btn" disabled={busy} onClick={() => handleFault("sump_ph", 5.8)}>
                Simulate Acid Sump (pH 5.8)
              </button>
              <button type="button" className="btn btn-accent" disabled={busy} onClick={() => handleFault("clear_all")}>
                Clear All Faults
              </button>
            </div>
          ) : (
            <p style={{ fontSize: 12.5, color: "var(--text-muted)", margin: 0 }}>
              Fault injection is disabled in physical Modbus RTU mode. Perform physical fail-safe
              drills using the hardwired 200°F snap-disc, 0.25 GPM paddle switch, and E-Stop.
            </p>
          )}
        </div>
      </section>
    </div>
  );
}
