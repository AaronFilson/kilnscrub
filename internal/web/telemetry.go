package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

const (
	defaultSSEKeepaliveInterval = 15 * time.Second
	sseWriteTimeout             = 10 * time.Second

	// Supported simulation fault-injection actions for POST /api/control/fault.
	FaultActionHeatSpike        = "heat_spike"
	FaultActionWaterLoss        = "water_loss"
	FaultActionClearWaterLoss   = "clear_water_loss"
	FaultActionSensorFault      = "sensor_fault"
	FaultActionClearSensorFault = "clear_sensor_fault"
	FaultActionSumpPH           = "sump_ph"
	FaultActionClearAll         = "clear_all"
)

// ControlBackend decouples the web server and SSE telemetry endpoints from the
// active control loop and HAL implementation. It depends only on pkg/model.
type ControlBackend interface {
	HALMode() string
	SupportsFaultInjection() bool
	Status() model.ControlStatus
	Config() model.ControlConfig
	Subscribe() (<-chan model.ControlStatus, func())
	ResetInterlock() error
	SetSetpoints(targetDraftInWC *float64, quenchSetpointF *float64) (model.ControlConfig, error)
	InjectFault(action string, value float64) error
}

type controlDisabledResponse struct {
	Enabled bool `json:"enabled"`
}

type controlStatusResponse struct {
	Enabled                bool                `json:"enabled"`
	HALMode                string              `json:"hal_mode"`
	SupportsFaultInjection bool                `json:"supports_fault_injection"`
	Status                 model.ControlStatus `json:"status"`
	Config                 model.ControlConfig `json:"config"`
}

type setpointRequest struct {
	TargetDraftInWC *float64 `json:"target_draft_in_wc"`
	QuenchSetpointF *float64 `json:"quench_setpoint_f"`
}

type faultRequest struct {
	Action string  `json:"action"`
	Value  float64 `json:"value"`
}

func (s *Server) handleControlStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	b := s.getControlBackend()
	if b == nil {
		writeJSONResponse(w, r, http.StatusOK, controlDisabledResponse{Enabled: false})
		return
	}

	resp := controlStatusResponse{
		Enabled:                true,
		HALMode:                b.HALMode(),
		SupportsFaultInjection: b.SupportsFaultInjection(),
		Status:                 b.Status(),
		Config:                 b.Config(),
	}
	writeJSONResponse(w, r, http.StatusOK, resp)
}

func (s *Server) handleControlStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	b := s.getControlBackend()
	if b == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "live control backend is not enabled")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "streaming unsupported by response writer")
		return
	}

	// Clear the HTTP server's WriteTimeout for this long-lived SSE response, then apply a
	// bounded per-write deadline on every event/ping so slow/stalled TCP clients never leak goroutines.
	rc := http.NewResponseController(w)
	_ = rc.SetWriteDeadline(time.Time{})

	// Subscribe before reading the initial snapshot so no intermediate transition is missed.
	ch, unsubscribe := b.Subscribe()
	if unsubscribe != nil {
		defer unsubscribe()
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)

	if err := writeSSEStatus(rc, w, flusher, b.Status()); err != nil {
		return
	}

	keepalive := s.getSSEKeepaliveInterval()
	ticker := time.NewTicker(keepalive)
	defer ticker.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case st, open := <-ch:
			if !open {
				return
			}
			if err := writeSSEStatus(rc, w, flusher, st); err != nil {
				return
			}
		case <-ticker.C:
			_ = rc.SetWriteDeadline(time.Now().Add(sseWriteTimeout))
			if _, err := io.WriteString(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
			_ = rc.SetWriteDeadline(time.Time{})
		}
	}
}

func writeSSEStatus(rc *http.ResponseController, w http.ResponseWriter, flusher http.Flusher, st model.ControlStatus) error {
	payload, err := json.Marshal(st)
	if err != nil {
		return err
	}
	if rc != nil {
		_ = rc.SetWriteDeadline(time.Now().Add(sseWriteTimeout))
		defer func() { _ = rc.SetWriteDeadline(time.Time{}) }()
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

func (s *Server) handleControlReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !enforceSameOrigin(w, r) {
		return
	}

	b := s.getControlBackend()
	if b == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "live control backend is not enabled")
		return
	}

	if err := b.ResetInterlock(); err != nil {
		writeJSONError(w, http.StatusConflict, err.Error())
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, b.Status())
}

func (s *Server) handleControlSetpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !enforceSameOrigin(w, r) || !enforceJSONContentType(w, r) {
		return
	}

	b := s.getControlBackend()
	if b == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "live control backend is not enabled")
		return
	}

	body, err := readLimitedBody(r)
	if err != nil {
		writeBodyReadError(w, err)
		return
	}
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		writeJSONError(w, http.StatusBadRequest, "request body must be a JSON object")
		return
	}

	var req setpointRequest
	if err := json.Unmarshal(trimmed, &req); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON payload: %v", err))
		return
	}

	if req.TargetDraftInWC == nil && req.QuenchSetpointF == nil {
		writeJSONError(w, http.StatusBadRequest, "at least one of target_draft_in_wc or quench_setpoint_f must be provided")
		return
	}

	if req.TargetDraftInWC != nil {
		d := *req.TargetDraftInWC
		if math.IsNaN(d) || math.IsInf(d, 0) || d > 0 || d < -2.0 {
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("target_draft_in_wc must be finite and in [-2.0, 0], got %v", d))
			return
		}
	}

	if req.QuenchSetpointF != nil {
		qF := *req.QuenchSetpointF
		maxTempK := b.Config().Safety.MaxQuenchOutletTemp
		if maxTempK <= 273.15 {
			maxTempK = model.FahrenheitToKelvin(200.0)
		}
		maxTempF := maxTempK.Fahrenheit()
		if math.IsNaN(qF) || math.IsInf(qF, 0) || qF <= 32.0 || qF >= maxTempF {
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("quench_setpoint_f must be finite and in (32.0, %.1f), got %v", maxTempF, qF))
			return
		}
	}

	cfg, err := b.SetSetpoints(req.TargetDraftInWC, req.QuenchSetpointF)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) handleControlFault(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !enforceSameOrigin(w, r) || !enforceJSONContentType(w, r) {
		return
	}

	b := s.getControlBackend()
	if b == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "live control backend is not enabled")
		return
	}

	if !b.SupportsFaultInjection() {
		writeJSONError(w, http.StatusForbidden, "fault injection is not supported by active HAL backend")
		return
	}

	body, err := readLimitedBody(r)
	if err != nil {
		writeBodyReadError(w, err)
		return
	}
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		writeJSONError(w, http.StatusBadRequest, "request body must be a JSON object")
		return
	}

	var req faultRequest
	if err := json.Unmarshal(trimmed, &req); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON payload: %v", err))
		return
	}

	action := strings.ToLower(strings.TrimSpace(req.Action))
	switch action {
	case FaultActionHeatSpike,
		FaultActionWaterLoss,
		FaultActionClearWaterLoss,
		FaultActionSensorFault,
		FaultActionClearSensorFault,
		FaultActionSumpPH,
		FaultActionClearAll:
	default:
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("unknown fault action %q", req.Action))
		return
	}

	if math.IsNaN(req.Value) || math.IsInf(req.Value, 0) {
		writeJSONError(w, http.StatusBadRequest, "fault value must be finite")
		return
	}
	if action == FaultActionSumpPH && (req.Value <= 0 || req.Value > 14.0) {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("sump_ph value must be in (0, 14], got %v", req.Value))
		return
	}

	if err := b.InjectFault(action, req.Value); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, b.Status())
}
