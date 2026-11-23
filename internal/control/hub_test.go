package control

import (
	"context"
	"errors"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

func TestTelemetryHubConcurrentPubSubAndSlowSubscriber(t *testing.T) {
	hub := NewTelemetryHub(4)
	defer hub.Close()

	// Slow subscriber that never drains during publishing burst.
	slowCh, slowUnsub := hub.Subscribe()
	defer slowUnsub()

	// Multiple concurrent fast subscribers.
	const fastCount = 4
	const totalFrames = 200

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		fastRecvd = make([]int, fastCount)
	)
	unsubs := make([]func(), fastCount)

	for i := 0; i < fastCount; i++ {
		ch, unsub := hub.Subscribe()
		unsubs[i] = unsub
		idx := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range ch {
				mu.Lock()
				fastRecvd[idx]++
				mu.Unlock()
			}
		}()
	}

	if got := hub.Subscribers(); got != fastCount+1 {
		t.Fatalf("Subscribers() = %d, want %d", got, fastCount+1)
	}

	// Publish 200 frames concurrently from 2 goroutines while slowCh is never drained.
	start := time.Now()
	var pubWG sync.WaitGroup
	pubWG.Add(2)
	for g := 0; g < 2; g++ {
		base := g * (totalFrames / 2)
		go func() {
			defer pubWG.Done()
			for k := 0; k < totalFrames/2; k++ {
				hub.Publish(model.ControlStatus{
					State:           StateNormal,
					FilteredDraftPa: model.Pascal(-(base + k + 1)),
					UpdatedAt:       time.Now(),
				})
			}
		}()
	}
	pubWG.Wait()
	elapsed := time.Since(start)

	// Publishing must never block on the stalled slow subscriber.
	if elapsed > 200*time.Millisecond {
		t.Fatalf("TelemetryHub.Publish took %v with slow subscriber (must be non-blocking)", elapsed)
	}
	if hub.Dropped() == 0 {
		t.Fatalf("expected dropped frames > 0 due to stalled slow subscriber")
	}

	// Slow subscriber should still have up to bufferSize (4) fresh snapshots buffered.
	var drained []model.ControlStatus
	for {
		select {
		case st := <-slowCh:
			drained = append(drained, st)
		default:
			goto doneDrain
		}
	}
doneDrain:
	if len(drained) == 0 || len(drained) > 4 {
		t.Fatalf("expected 1..4 buffered snapshots on slow subscriber, got %d", len(drained))
	}

	// Unsubscribe all fast subscribers idempotently.
	for _, unsub := range unsubs {
		unsub()
		unsub() // verify idempotence
	}
	wg.Wait()

	if got := hub.Subscribers(); got != 1 {
		t.Fatalf("expected 1 subscriber remaining (slow), got %d", got)
	}

	mu.Lock()
	for i, n := range fastRecvd {
		if n == 0 {
			t.Fatalf("fast subscriber %d received 0 frames", i)
		}
	}
	mu.Unlock()

	// Closing hub closes remaining subscriber channel cleanly.
	hub.Close()
	hub.Close() // idempotent
	if got := hub.Subscribers(); got != 0 {
		t.Fatalf("expected 0 subscribers after Close(), got %d", got)
	}
	if _, ok := <-slowCh; ok {
		t.Fatalf("expected slowCh to be closed after hub.Close()")
	}
}

func TestDaemonTelemetryHubSetpointsAndFaultInjection(t *testing.T) {
	cfg := model.DefaultControlConfig()
	cfg.LoopInterval = 10 * time.Millisecond
	simCfg := DefaultSimHALConfig()
	simCfg.AutoStepInterval = 100 * time.Millisecond
	simHAL := NewSimHAL(simCfg)

	d, err := NewDaemon(cfg, simHAL)
	if err != nil {
		t.Fatalf("NewDaemon: %v", err)
	}

	if d.HALMode() != HALModeSim {
		t.Fatalf("HALMode() = %q, want %q", d.HALMode(), HALModeSim)
	}
	if !d.SupportsFaultInjection() {
		t.Fatalf("expected SupportsFaultInjection() == true for SimHAL")
	}

	// Test live setpoint updates and validation atomicity.
	if err := d.SetTargetDraftInWC(-0.08); err != nil {
		t.Fatalf("SetTargetDraftInWC(-0.08): %v", err)
	}
	if err := d.SetQuenchSetpointF(170.0); err != nil {
		t.Fatalf("SetQuenchSetpointF(170): %v", err)
	}
	gotCfg := d.Config()
	if math.Abs(gotCfg.TargetDraftInWC-(-0.08)) > 1e-6 || math.Abs(float64(gotCfg.TargetDraftPa-model.InWCToPascal(-0.08))) > 1e-4 {
		t.Fatalf("unexpected updated draft setpoint: %+v", gotCfg)
	}
	if math.Abs(gotCfg.QuenchSetpoint.Fahrenheit()-170.0) > 1e-4 {
		t.Fatalf("unexpected updated quench setpoint: %.2f F", gotCfg.QuenchSetpoint.Fahrenheit())
	}

	// Atomic rejection: valid draft + invalid quench temp (> 210 F limit) must leave config untouched.
	goodDraft := -0.10
	badTempF := 225.0
	if _, err := d.SetSetpoints(&goodDraft, &badTempF); err == nil {
		t.Fatalf("expected SetSetpoints to reject quench temp above CPVC limit")
	}
	if math.Abs(d.Config().TargetDraftInWC-(-0.08)) > 1e-6 {
		t.Fatalf("SetSetpoints partially applied draft despite invalid quench temp")
	}

	// Subscribe before starting daemon loop.
	subCh, unsub := d.Subscribe()
	defer unsub()

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- d.Run(ctx)
	}()

	// Receive live snapshots over TelemetryHub.
	select {
	case st := <-subCh:
		if st.UpdatedAt.IsZero() {
			t.Fatalf("expected non-zero UpdatedAt from TelemetryHub")
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for telemetry from Daemon.Subscribe()")
	}

	// Exercise interactive fault injection hooks concurrently while daemon ticks.
	if err := d.InjectFault("sump_ph", 5.8); err != nil {
		t.Fatalf("InjectFault(sump_ph): %v", err)
	}
	waitFor(t, time.Second, func() bool {
		return d.Status().State == StateWarning && !d.Status().InterlockActive
	}, "expected StateWarning for low sump pH")

	if err := d.InjectFault("heat_spike", 400.0); err != nil {
		t.Fatalf("InjectFault(heat_spike): %v", err)
	}

	if err := d.InjectFault("water_loss", 0); err != nil {
		t.Fatalf("InjectFault(water_loss): %v", err)
	}
	waitFor(t, time.Second, func() bool {
		st := d.Status()
		return st.InterlockActive && st.Commands.TripReason == model.TripWaterLoss
	}, "expected TripWaterLoss on water_loss fault")

	// Clear all faults, wait for fault clearance, then ResetInterlock.
	if err := d.InjectFault("clear_all", 0); err != nil {
		t.Fatalf("InjectFault(clear_all): %v", err)
	}
	waitFor(t, time.Second, func() bool {
		return !d.Controller().InterlockState().FaultActive
	}, "expected active fault to clear after clear_all")

	if err := d.ResetInterlock(); err != nil {
		t.Fatalf("ResetInterlock failed after clear_all: %v", err)
	}
	if d.Status().InterlockActive {
		t.Fatalf("expected InterlockActive == false after ResetInterlock")
	}

	// Sensor fault + clear_sensor_fault round-trip.
	if err := d.InjectFault("sensor_fault", 0); err != nil {
		t.Fatalf("InjectFault(sensor_fault): %v", err)
	}
	waitFor(t, time.Second, func() bool {
		st := d.Status()
		return st.InterlockActive && st.Commands.TripReason == model.TripSensorFault
	}, "expected TripSensorFault on sensor_fault")

	if err := d.InjectFault("clear_sensor_fault", 0); err != nil {
		t.Fatalf("InjectFault(clear_sensor_fault): %v", err)
	}
	waitFor(t, time.Second, func() bool {
		return !d.Controller().InterlockState().FaultActive
	}, "expected sensor fault to clear after clear_sensor_fault")

	// Unknown action must error.
	if err := d.InjectFault("bogus_action", 0); err == nil {
		t.Fatalf("expected error for unknown fault action")
	}

	cancel()
	if runErr := <-errCh; runErr != nil {
		t.Fatalf("Daemon.Run returned error: %v", runErr)
	}
}

func TestDaemonModbusHALModeAndNoFaultInjection(t *testing.T) {
	slave := newMockModbusSlave(1)
	mbHAL, err := NewModbusHAL(DefaultModbusConfig(), slave)
	if err != nil {
		t.Fatalf("NewModbusHAL: %v", err)
	}
	defer mbHAL.Close()

	d, err := NewDaemon(model.DefaultControlConfig(), mbHAL)
	if err != nil {
		t.Fatalf("NewDaemon(ModbusHAL): %v", err)
	}
	if d.HALMode() != HALModeModbus {
		t.Fatalf("HALMode() = %q, want %q", d.HALMode(), HALModeModbus)
	}
	if d.SupportsFaultInjection() {
		t.Fatalf("expected SupportsFaultInjection() == false for ModbusHAL")
	}
	if err := d.InjectFault("water_loss", 0); !errors.Is(err, ErrFaultInjectionUnsupported) {
		t.Fatalf("expected ErrFaultInjectionUnsupported, got %v", err)
	}
}
