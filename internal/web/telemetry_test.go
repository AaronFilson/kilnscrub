package web_test

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AaronFilson/kilnscrub/internal/web"
	"github.com/AaronFilson/kilnscrub/pkg/model"
)

type faultCall struct {
	action string
	value  float64
}

type mockControlBackend struct {
	mu                     sync.Mutex
	halMode                string
	supportsFaultInjection bool
	status                 model.ControlStatus
	cfg                    model.ControlConfig
	resetErr               error
	setpointErr            error
	faultErr               error
	faultCalls             []faultCall

	nextSubID   int
	subscribers map[int]chan model.ControlStatus
	unsubCount  int
}

func newMockControlBackend(halMode string, supportsFault bool) *mockControlBackend {
	cfg := model.DefaultControlConfig()
	now := time.Now().UTC()
	return &mockControlBackend{
		halMode:                halMode,
		supportsFaultInjection: supportsFault,
		cfg:                    cfg,
		status: model.ControlStatus{
			State:             "NORMAL",
			FilteredDraftPa:   model.InWCToPascal(-0.05),
			FilteredDraftInWC: -0.05,
			Readings: model.SensorReadings{
				DraftPa:             model.InWCToPascal(-0.05),
				DraftInWC:           -0.05,
				QuenchInletTemp:     model.FahrenheitToKelvin(2400.0),
				QuenchOutletTemp:    model.FahrenheitToKelvin(165.0),
				WaterFlow:           model.GPMToCubicMetersPerSec(2.2),
				WaterFlowGPM:        2.2,
				SumpPH:              8.0,
				BlowerSpeedFraction: 0.45,
				Timestamp:           now,
			},
			Commands: model.ActuatorCommands{
				BlowerVFD:        0.45,
				QuenchSprayValve: 0.50,
				BypassDamperOpen: false,
				AlarmLatched:     false,
				TripReason:       model.TripNone,
			},
			InterlockActive: false,
			UpdatedAt:       now,
		},
		subscribers: make(map[int]chan model.ControlStatus),
	}
}

func (m *mockControlBackend) HALMode() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.halMode
}

func (m *mockControlBackend) SupportsFaultInjection() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.supportsFaultInjection
}

func (m *mockControlBackend) Status() model.ControlStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.status
}

func (m *mockControlBackend) Config() model.ControlConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cfg
}

func (m *mockControlBackend) Subscribe() (<-chan model.ControlStatus, func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := m.nextSubID
	m.nextSubID++
	ch := make(chan model.ControlStatus, 16)
	m.subscribers[id] = ch

	var once sync.Once
	unsub := func() {
		once.Do(func() {
			m.mu.Lock()
			delete(m.subscribers, id)
			m.unsubCount++
			m.mu.Unlock()
		})
	}
	return ch, unsub
}

func (m *mockControlBackend) Publish(st model.ControlStatus) {
	m.mu.Lock()
	m.status = st
	subs := make([]chan model.ControlStatus, 0, len(m.subscribers))
	for _, ch := range m.subscribers {
		subs = append(subs, ch)
	}
	m.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- st:
		default:
		}
	}
}

func (m *mockControlBackend) ActiveSubscribers() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.subscribers)
}

func (m *mockControlBackend) UnsubCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.unsubCount
}

func (m *mockControlBackend) ResetInterlock() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.resetErr != nil {
		return m.resetErr
	}
	m.status.State = "NORMAL"
	m.status.InterlockActive = false
	m.status.Commands.BypassDamperOpen = false
	m.status.Commands.AlarmLatched = false
	m.status.Commands.TripReason = model.TripNone
	m.status.UpdatedAt = time.Now().UTC()
	return nil
}

func (m *mockControlBackend) SetSetpoints(targetDraftInWC *float64, quenchSetpointF *float64) (model.ControlConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.setpointErr != nil {
		return model.ControlConfig{}, m.setpointErr
	}
	next := m.cfg
	if targetDraftInWC != nil {
		next.SetTargetDraftInWC(*targetDraftInWC)
	}
	if quenchSetpointF != nil {
		next.QuenchSetpoint = model.FahrenheitToKelvin(*quenchSetpointF)
	}
	if err := next.Validate(); err != nil {
		return model.ControlConfig{}, err
	}
	m.cfg = next
	return m.cfg, nil
}

func (m *mockControlBackend) InjectFault(action string, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.faultErr != nil {
		return m.faultErr
	}
	m.faultCalls = append(m.faultCalls, faultCall{action: action, value: value})
	switch action {
	case web.FaultActionHeatSpike:
		m.status.Readings.QuenchOutletTemp = model.FahrenheitToKelvin(215.0)
	case web.FaultActionWaterLoss:
		m.status.Readings.WaterFlowGPM = 0.0
		m.status.Readings.WaterFlow = 0
		m.status.State = "TRIPPED"
		m.status.InterlockActive = true
		m.status.Commands.BypassDamperOpen = true
		m.status.Commands.AlarmLatched = true
		m.status.Commands.TripReason = model.TripWaterLoss
	case web.FaultActionClearWaterLoss:
		m.status.Readings.WaterFlowGPM = 2.2
		m.status.Readings.WaterFlow = model.GPMToCubicMetersPerSec(2.2)
	case web.FaultActionSensorFault:
		m.status.State = "TRIPPED"
		m.status.InterlockActive = true
		m.status.Commands.BypassDamperOpen = true
		m.status.Commands.AlarmLatched = true
		m.status.Commands.TripReason = model.TripSensorFault
	case web.FaultActionClearSensorFault:
		// Fault cleared, awaits explicit ResetInterlock.
	case web.FaultActionSumpPH:
		m.status.Readings.SumpPH = value
		if value < m.cfg.Safety.MinSumpPH || value > m.cfg.Safety.MaxSumpPH {
			m.status.State = "WARNING"
		}
	case web.FaultActionClearAll:
		m.status.Readings.WaterFlowGPM = 2.2
		m.status.Readings.WaterFlow = model.GPMToCubicMetersPerSec(2.2)
		m.status.Readings.QuenchOutletTemp = model.FahrenheitToKelvin(165.0)
		m.status.Readings.SumpPH = 8.0
		if !m.status.InterlockActive {
			m.status.State = "NORMAL"
		}
	}
	m.status.UpdatedAt = time.Now().UTC()
	return nil
}

func TestControlStatusEndpoint(t *testing.T) {
	t.Run("disabled when no backend attached", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, newMockExporter())
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/control/status", nil))

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
			t.Errorf("Cache-Control = %q, want no-store", cc)
		}
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if enabled, ok := body["enabled"].(bool); !ok || enabled {
			t.Fatalf("enabled = %v (ok=%v), want false", body["enabled"], ok)
		}
		if len(body) != 1 {
			t.Fatalf("expected minimal disabled payload {\"enabled\":false}, got %v", body)
		}
	})

	t.Run("enabled sim backend via SetControlBackend", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, newMockExporter())
		backend := newMockControlBackend("sim", true)
		srv.SetControlBackend(backend)

		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/control/status", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}

		var got struct {
			Enabled                bool                `json:"enabled"`
			HALMode                string              `json:"hal_mode"`
			SupportsFaultInjection bool                `json:"supports_fault_injection"`
			Status                 model.ControlStatus `json:"status"`
			Config                 model.ControlConfig `json:"config"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("unmarshal status payload: %v", err)
		}
		if !got.Enabled || got.HALMode != "sim" || !got.SupportsFaultInjection {
			t.Fatalf("unexpected control metadata: enabled=%v hal=%q fault=%v", got.Enabled, got.HALMode, got.SupportsFaultInjection)
		}
		if got.Status.State != "NORMAL" || got.Status.FilteredDraftInWC != -0.05 {
			t.Fatalf("unexpected status snapshot: %+v", got.Status)
		}
		if got.Config.TargetDraftInWC != -0.05 {
			t.Fatalf("unexpected config target_draft_in_wc: %v", got.Config.TargetDraftInWC)
		}
	})

	t.Run("enabled modbus backend via NewServerWithControl", func(t *testing.T) {
		backend := newMockControlBackend("modbus", false)
		srv := web.NewServerWithControl(&mockDesigner{}, newMockExporter(), backend)

		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/control/status", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		var raw map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if raw["hal_mode"] != "modbus" {
			t.Fatalf("hal_mode = %v, want modbus", raw["hal_mode"])
		}
		if v, exists := raw["supports_fault_injection"]; !exists || v != false {
			t.Fatalf("supports_fault_injection = %v (exists=%v), want explicit false", v, exists)
		}
	})

	t.Run("wrong method returns 405", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, newMockExporter())
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/control/status", nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("POST /api/control/status = %d, want 405", rec.Code)
		}
	})
}

func TestControlStreamSSE(t *testing.T) {
	t.Run("disabled returns 503 JSON", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, newMockExporter())
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/control/stream", nil))
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want 503", rec.Code)
		}
		var errResp map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil || errResp["error"] == "" {
			t.Fatalf("expected JSON error body, got %q", rec.Body.String())
		}
	})

	t.Run("wrong method returns 405", func(t *testing.T) {
		srv := web.NewServerWithControl(&mockDesigner{}, newMockExporter(), newMockControlBackend("sim", true))
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/control/stream", nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want 405", rec.Code)
		}
	})

	t.Run("streams initial status, multiple updates, keepalive ping, and cleans up on disconnect", func(t *testing.T) {
		backend := newMockControlBackend("sim", true)
		srv := web.NewServerWithControl(&mockDesigner{}, newMockExporter(), backend)
		srv.SetSSEKeepaliveInterval(35 * time.Millisecond)

		ts := httptest.NewServer(srv.Handler())
		defer ts.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/api/control/stream", nil)
		if err != nil {
			t.Fatalf("new request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("GET /api/control/stream: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
			t.Fatalf("Content-Type = %q, want text/event-stream", ct)
		}
		if cc := resp.Header.Get("Cache-Control"); !strings.Contains(cc, "no-cache") {
			t.Fatalf("Cache-Control = %q, want no-cache", cc)
		}

		reader := bufio.NewReader(resp.Body)
		readDataEvent := func() model.ControlStatus {
			t.Helper()
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					t.Fatalf("read SSE line: %v", err)
				}
				if strings.HasPrefix(line, "data: ") {
					payload := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
					// Consume terminating blank line.
					blank, err := reader.ReadString('\n')
					if err != nil {
						t.Fatalf("read SSE blank line: %v", err)
					}
					if strings.TrimSpace(blank) != "" {
						t.Fatalf("expected blank line after SSE data, got %q", blank)
					}
					var st model.ControlStatus
					if err := json.Unmarshal([]byte(payload), &st); err != nil {
						t.Fatalf("unmarshal SSE status JSON %q: %v", payload, err)
					}
					return st
				}
			}
		}

		// 1. Initial status flushed immediately.
		initial := readDataEvent()
		if initial.State != "NORMAL" || initial.FilteredDraftInWC != -0.05 {
			t.Fatalf("unexpected initial SSE event: %+v", initial)
		}

		// 2. Publish two updates and verify both arrive in order.
		u1 := initial
		u1.FilteredDraftInWC = -0.08
		u1.FilteredDraftPa = model.InWCToPascal(-0.08)
		u1.State = "WARNING"
		u1.Readings.SumpPH = 5.8

		u2 := u1
		u2.State = "TRIPPED"
		u2.InterlockActive = true
		u2.Commands.BypassDamperOpen = true
		u2.Commands.TripReason = model.TripHighQuenchTemp

		backend.Publish(u1)
		backend.Publish(u2)

		got1 := readDataEvent()
		if got1.State != "WARNING" || got1.FilteredDraftInWC != -0.08 || got1.Readings.SumpPH != 5.8 {
			t.Fatalf("unexpected SSE event #2: %+v", got1)
		}
		got2 := readDataEvent()
		if got2.State != "TRIPPED" || !got2.InterlockActive || got2.Commands.TripReason != model.TripHighQuenchTemp {
			t.Fatalf("unexpected SSE event #3: %+v", got2)
		}

		// 3. Verify periodic keepalive comment ": ping\n\n" arrives.
		sawPing := false
		for i := 0; i < 6; i++ {
			line, err := reader.ReadString('\n')
			if err != nil {
				t.Fatalf("read SSE ping line: %v", err)
			}
			if line == ": ping\n" {
				sawPing = true
				break
			}
		}
		if !sawPing {
			t.Fatal("expected SSE keepalive ': ping' comment")
		}

		// 4. Cancel client context and confirm clean unsubscribe.
		cancel()
		resp.Body.Close()

		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			if backend.ActiveSubscribers() == 0 && backend.UnsubCount() >= 1 {
				break
			}
			time.Sleep(15 * time.Millisecond)
		}
		if backend.ActiveSubscribers() != 0 {
			t.Fatalf("active subscribers = %d, want 0 after disconnect", backend.ActiveSubscribers())
		}
	})

	t.Run("concurrent SSE clients connect, receive, and disconnect cleanly under -race", func(t *testing.T) {
		backend := newMockControlBackend("sim", true)
		srv := web.NewServerWithControl(&mockDesigner{}, newMockExporter(), backend)
		srv.SetSSEKeepaliveInterval(50 * time.Millisecond)

		ts := httptest.NewServer(srv.Handler())
		defer ts.Close()

		const numClients = 6
		var wg sync.WaitGroup
		readyCh := make(chan struct{}, numClients)

		for i := 0; i < numClients; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				req, _ := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/api/control/stream", nil)
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					t.Errorf("client %d connect failed: %v", idx, err)
					return
				}
				defer resp.Body.Close()

				reader := bufio.NewReader(resp.Body)
				gotEvents := 0
				for gotEvents < 2 {
					line, err := reader.ReadString('\n')
					if err != nil {
						t.Errorf("client %d read error after %d events: %v", idx, gotEvents, err)
						return
					}
					if strings.HasPrefix(line, "data: ") {
						gotEvents++
						if gotEvents == 1 {
							readyCh <- struct{}{}
						}
					}
				}
				cancel()
			}(i)
		}

		for i := 0; i < numClients; i++ {
			select {
			case <-readyCh:
			case <-time.After(2 * time.Second):
				t.Fatal("timed out waiting for SSE clients to receive initial snapshot")
			}
		}

		st := backend.Status()
		st.FilteredDraftInWC = -0.09
		backend.Publish(st)

		wg.Wait()

		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			if backend.ActiveSubscribers() == 0 {
				break
			}
			time.Sleep(15 * time.Millisecond)
		}
		if backend.ActiveSubscribers() != 0 {
			t.Fatalf("leaked subscribers = %d, want 0", backend.ActiveSubscribers())
		}
	})
}

func TestControlResetEndpoint(t *testing.T) {
	t.Run("disabled returns 503", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, newMockExporter())
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/control/reset", nil))
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want 503", rec.Code)
		}
	})

	t.Run("wrong method returns 405", func(t *testing.T) {
		srv := web.NewServerWithControl(&mockDesigner{}, newMockExporter(), newMockControlBackend("sim", true))
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/control/reset", nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want 405", rec.Code)
		}
	})

	t.Run("returns 409 Conflict when fault is still active", func(t *testing.T) {
		backend := newMockControlBackend("sim", true)
		backend.resetErr = errors.New(`cannot reset interlock while fault "QUENCH_WATER_LOSS" is still active`)
		srv := web.NewServerWithControl(&mockDesigner{}, newMockExporter(), backend)

		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/control/reset", nil))
		if rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, want 409; body=%s", rec.Code, rec.Body.String())
		}
		var errResp map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil || !strings.Contains(errResp["error"], "QUENCH_WATER_LOSS") {
			t.Fatalf("expected 409 JSON error mentioning QUENCH_WATER_LOSS, got %q", rec.Body.String())
		}
	})

	t.Run("returns 200 with updated status on successful reset", func(t *testing.T) {
		backend := newMockControlBackend("sim", true)
		backend.status.State = "TRIPPED"
		backend.status.InterlockActive = true
		backend.status.Commands.BypassDamperOpen = true
		backend.status.Commands.AlarmLatched = true
		backend.status.Commands.TripReason = model.TripWaterLoss

		srv := web.NewServerWithControl(&mockDesigner{}, newMockExporter(), backend)
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/control/reset", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		var st model.ControlStatus
		if err := json.Unmarshal(rec.Body.Bytes(), &st); err != nil {
			t.Fatalf("unmarshal ControlStatus: %v", err)
		}
		if st.InterlockActive || st.Commands.BypassDamperOpen || st.State != "NORMAL" {
			t.Fatalf("expected cleared interlock status, got %+v", st)
		}
	})
}

func TestControlSetpointEndpoint(t *testing.T) {
	t.Run("disabled returns 503", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, newMockExporter())
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/control/setpoint", strings.NewReader(`{"target_draft_in_wc":-0.08}`)))
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want 503", rec.Code)
		}
	})

	t.Run("wrong method returns 405", func(t *testing.T) {
		srv := web.NewServerWithControl(&mockDesigner{}, newMockExporter(), newMockControlBackend("sim", true))
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/control/setpoint", nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want 405", rec.Code)
		}
	})

	t.Run("valid updates both draft and quench setpoints", func(t *testing.T) {
		backend := newMockControlBackend("sim", true)
		srv := web.NewServerWithControl(&mockDesigner{}, newMockExporter(), backend)

		body := `{"target_draft_in_wc": -0.10, "quench_setpoint_f": 175}`
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/control/setpoint", strings.NewReader(body)))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}

		var cfg model.ControlConfig
		if err := json.Unmarshal(rec.Body.Bytes(), &cfg); err != nil {
			t.Fatalf("unmarshal ControlConfig: %v", err)
		}
		if cfg.TargetDraftInWC != -0.10 {
			t.Errorf("TargetDraftInWC = %v, want -0.10", cfg.TargetDraftInWC)
		}
		if gotF := cfg.QuenchSetpoint.Fahrenheit(); gotF < 174.9 || gotF > 175.1 {
			t.Errorf("QuenchSetpoint = %.2f F, want 175.0 F", gotF)
		}
	})

	t.Run("valid partial update only draft", func(t *testing.T) {
		backend := newMockControlBackend("sim", true)
		srv := web.NewServerWithControl(&mockDesigner{}, newMockExporter(), backend)

		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/control/setpoint", strings.NewReader(`{"target_draft_in_wc": -0.12}`)))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
		}
		var cfg model.ControlConfig
		if err := json.Unmarshal(rec.Body.Bytes(), &cfg); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if cfg.TargetDraftInWC != -0.12 {
			t.Fatalf("TargetDraftInWC = %v, want -0.12", cfg.TargetDraftInWC)
		}
	})

	badPayloads := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"malformed JSON", "{broken"},
		{"JSON array", "[]"},
		{"empty object (no fields)", "{}"},
		{"positive draft", `{"target_draft_in_wc": 0.05}`},
		{"draft below -2.0", `{"target_draft_in_wc": -2.5}`},
		{"quench at 210F CPVC limit", `{"quench_setpoint_f": 210}`},
		{"quench above 210F", `{"quench_setpoint_f": 225}`},
		{"quench at freezing", `{"quench_setpoint_f": 32}`},
	}

	for _, tc := range badPayloads {
		t.Run("invalid: "+tc.name, func(t *testing.T) {
			srv := web.NewServerWithControl(&mockDesigner{}, newMockExporter(), newMockControlBackend("sim", true))
			rec := httptest.NewRecorder()
			srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/control/setpoint", strings.NewReader(tc.body)))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
			}
			var errResp map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil || errResp["error"] == "" {
				t.Fatalf("expected JSON error, got %q", rec.Body.String())
			}
		})
	}
}

func TestControlFaultEndpoint(t *testing.T) {
	t.Run("disabled returns 503", func(t *testing.T) {
		srv := web.NewServer(&mockDesigner{}, newMockExporter())
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/control/fault", strings.NewReader(`{"action":"water_loss"}`)))
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want 503", rec.Code)
		}
	})

	t.Run("wrong method returns 405", func(t *testing.T) {
		srv := web.NewServerWithControl(&mockDesigner{}, newMockExporter(), newMockControlBackend("sim", true))
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/control/fault", nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want 405", rec.Code)
		}
	})

	t.Run("unsupported backend returns 403", func(t *testing.T) {
		srv := web.NewServerWithControl(&mockDesigner{}, newMockExporter(), newMockControlBackend("modbus", false))
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/control/fault", strings.NewReader(`{"action":"water_loss"}`)))
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403; body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("valid actions succeed and return updated status", func(t *testing.T) {
		backend := newMockControlBackend("sim", true)
		srv := web.NewServerWithControl(&mockDesigner{}, newMockExporter(), backend)

		cases := []struct {
			body   string
			action string
			value  float64
		}{
			{`{"action":"heat_spike","value":350}`, web.FaultActionHeatSpike, 350},
			{`{"action":"water_loss","value":1}`, web.FaultActionWaterLoss, 1},
			{`{"action":"clear_water_loss"}`, web.FaultActionClearWaterLoss, 0},
			{`{"action":"sensor_fault","value":1}`, web.FaultActionSensorFault, 1},
			{`{"action":"clear_sensor_fault"}`, web.FaultActionClearSensorFault, 0},
			{`{"action":"sump_ph","value":5.8}`, web.FaultActionSumpPH, 5.8},
			{`{"action":"clear_all"}`, web.FaultActionClearAll, 0},
		}

		for _, tc := range cases {
			rec := httptest.NewRecorder()
			srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/control/fault", strings.NewReader(tc.body)))
			if rec.Code != http.StatusOK {
				t.Fatalf("action %q status = %d, want 200; body=%s", tc.action, rec.Code, rec.Body.String())
			}
			var st model.ControlStatus
			if err := json.Unmarshal(rec.Body.Bytes(), &st); err != nil {
				t.Fatalf("action %q unmarshal status: %v", tc.action, err)
			}
		}

		if len(backend.faultCalls) != len(cases) {
			t.Fatalf("recorded %d fault calls, want %d", len(backend.faultCalls), len(cases))
		}
		for i, tc := range cases {
			got := backend.faultCalls[i]
			if got.action != tc.action || got.value != tc.value {
				t.Errorf("call[%d] = %+v, want action=%q value=%v", i, got, tc.action, tc.value)
			}
		}
	})

	t.Run("unknown action and invalid payloads return 400", func(t *testing.T) {
		srv := web.NewServerWithControl(&mockDesigner{}, newMockExporter(), newMockControlBackend("sim", true))
		for _, body := range []string{
			"",
			"{bad",
			`{"action":""}`,
			`{"action":"meltdown"}`,
			`{"action":"sump_ph","value":0}`,
			`{"action":"sump_ph","value":15.5}`,
		} {
			rec := httptest.NewRecorder()
			srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/control/fault", strings.NewReader(body)))
			if rec.Code != http.StatusBadRequest {
				t.Errorf("body %q status = %d, want 400; resp=%s", body, rec.Code, rec.Body.String())
			}
		}
	})
}

func TestEmbeddedHTMLControlTabAndEndpoints(t *testing.T) {
	srv := web.NewServer(&mockDesigner{}, newMockExporter())
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()

	var bundle strings.Builder
	bundle.WriteString(body)
	for _, assetPath := range extractNextStaticPaths(body) {
		aRec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(aRec, httptest.NewRequest(http.MethodGet, assetPath, nil))
		if aRec.Code == http.StatusOK {
			bundle.WriteString("\n")
			bundle.WriteString(aRec.Body.String())
		}
	}
	full := bundle.String()

	required := []string{
		"Design &amp; 3D Studio",
		"Live Control &amp; Telemetry",
		"/api/control/status",
		"/api/control/stream",
		"/api/control/reset",
		"/api/control/setpoint",
		"/api/control/fault",
		"EventSource",
		"INTERLOCK TRIPPED",
		"Reset Interlock",
		"Clear All Faults",
	}

	for _, needle := range required {
		if !strings.Contains(full, needle) {
			t.Errorf("encapsulated Next.js bundle missing Live Control marker %q", needle)
		}
	}
}

func TestListenAndServeShutdownWithActiveSSE(t *testing.T) {
	backend := newMockControlBackend("sim", true)
	srv := web.NewServerWithControl(&mockDesigner{}, newMockExporter(), backend)
	srv.SetSSEKeepaliveInterval(50 * time.Millisecond)

	// Verify BaseContext cancellation terminates an active SSE stream and unsubscribes cleanly.
	ts := httptest.NewUnstartedServer(srv.Handler())
	baseCtx, baseCancel := context.WithCancel(context.Background())
	ts.Config.BaseContext = func(net.Listener) context.Context { return baseCtx }
	ts.Start()
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/control/stream")
	if err != nil {
		t.Fatalf("GET /api/control/stream: %v", err)
	}
	reader := bufio.NewReader(resp.Body)
	if _, err := reader.ReadString('\n'); err != nil {
		t.Fatalf("read initial SSE line: %v", err)
	}

	baseCancel()
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if backend.ActiveSubscribers() == 0 {
			break
		}
		time.Sleep(15 * time.Millisecond)
	}
	if backend.ActiveSubscribers() != 0 {
		t.Fatalf("active subscribers = %d, want 0 after server base context cancel", backend.ActiveSubscribers())
	}

	// Verify ListenAndServe exits cleanly when canceled.
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe(ctx, "127.0.0.1:0")
	}()
	time.Sleep(40 * time.Millisecond)
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("ListenAndServe shutdown error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("ListenAndServe did not shut down in time")
	}
}

func TestWebSecurityHeadersAndCSRFGuards(t *testing.T) {
	backend := newMockControlBackend("sim", true)
	srv := web.NewServerWithControl(&mockDesigner{}, newMockExporter(), backend)
	h := srv.Handler()

	// 1. Security headers on index and API responses.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := rec.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Errorf("X-Frame-Options = %q, want DENY", got)
	}
	if got := rec.Header().Get("Content-Security-Policy"); !strings.Contains(got, "frame-ancestors 'none'") {
		t.Errorf("Content-Security-Policy = %q, want frame-ancestors 'none'", got)
	}

	// 2. Cross-origin POST to mutating control endpoint must be rejected with 403.
	reqCross := httptest.NewRequest(http.MethodPost, "/api/control/reset", nil)
	reqCross.Host = "127.0.0.1:8080"
	reqCross.Header.Set("Origin", "https://evil.example")
	recCross := httptest.NewRecorder()
	h.ServeHTTP(recCross, reqCross)
	if recCross.Code != http.StatusForbidden {
		t.Errorf("cross-origin POST /api/control/reset status = %d, want 403", recCross.Code)
	}

	// 3. Cross-site Sec-Fetch-Site must be rejected with 403.
	reqSecFetch := httptest.NewRequest(http.MethodPost, "/api/control/reset", nil)
	reqSecFetch.Header.Set("Sec-Fetch-Site", "cross-site")
	recSecFetch := httptest.NewRecorder()
	h.ServeHTTP(recSecFetch, reqSecFetch)
	if recSecFetch.Code != http.StatusForbidden {
		t.Errorf("Sec-Fetch-Site: cross-site POST status = %d, want 403", recSecFetch.Code)
	}

	// 4. Form/text-plain CSRF Content-Type on JSON body endpoint must be rejected with 415.
	reqPlain := httptest.NewRequest(http.MethodPost, "/api/control/setpoint", strings.NewReader(`{"target_draft_in_wc":-0.08}`))
	reqPlain.Header.Set("Content-Type", "text/plain")
	recPlain := httptest.NewRecorder()
	h.ServeHTTP(recPlain, reqPlain)
	if recPlain.Code != http.StatusUnsupportedMediaType {
		t.Errorf("text/plain POST /api/control/setpoint status = %d, want 415", recPlain.Code)
	}
}
