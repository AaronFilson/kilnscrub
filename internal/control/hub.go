package control

import (
	"sync"
	"sync/atomic"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

const defaultHubBufferSize = 16

// TelemetryHub is a thread-safe fan-out broadcaster for live model.ControlStatus snapshots.
// Slow subscribers never stall the real-time PID control loop: when a subscriber's buffer is
// full, the oldest queued snapshot is dropped so the subscriber receives the freshest state.
type TelemetryHub struct {
	mu         sync.RWMutex
	bufferSize int
	nextID     uint64
	subs       map[uint64]chan model.ControlStatus
	closed     bool
	dropped    atomic.Uint64
}

// NewTelemetryHub constructs a TelemetryHub with per-subscriber channel capacity bufferSize.
// If bufferSize <= 0, a sensible default (16) is used.
func NewTelemetryHub(bufferSize int) *TelemetryHub {
	if bufferSize <= 0 {
		bufferSize = defaultHubBufferSize
	}
	return &TelemetryHub{
		bufferSize: bufferSize,
		subs:       make(map[uint64]chan model.ControlStatus),
	}
}

// Subscribe registers a new live telemetry subscriber and returns a read-only channel along
// with an idempotent unsubscribe callback. Callers should always invoke unsubscribe when done
// (e.g. when an HTTP SSE client disconnects) to release resources.
func (h *TelemetryHub) Subscribe() (<-chan model.ControlStatus, func()) {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan model.ControlStatus, h.bufferSize)
	if h.closed {
		close(ch)
		return ch, func() {}
	}

	id := h.nextID
	h.nextID++
	h.subs[id] = ch

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			h.mu.Lock()
			if subCh, ok := h.subs[id]; ok {
				delete(h.subs, id)
				close(subCh)
			}
			h.mu.Unlock()
		})
	}
	return ch, unsubscribe
}

// Publish fans out status to all active subscribers without ever blocking.
// If a subscriber channel is full, the oldest queued item is evicted to make room for status.
func (h *TelemetryHub) Publish(status model.ControlStatus) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.closed {
		return
	}
	for _, ch := range h.subs {
		select {
		case ch <- status:
		default:
			// Drop oldest buffered sample so slow consumers stay caught up without stalling PID ticks.
			select {
			case <-ch:
				h.dropped.Add(1)
			default:
			}
			select {
			case ch <- status:
			default:
				h.dropped.Add(1)
			}
		}
	}
}

// Subscribers returns the current number of active subscribers.
func (h *TelemetryHub) Subscribers() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subs)
}

// Dropped returns the total number of telemetry frames dropped due to slow subscribers.
func (h *TelemetryHub) Dropped() uint64 {
	return h.dropped.Load()
}

// Close closes all active subscriber channels and marks the hub closed.
// Subsequent Publish calls are no-ops and Subscribe returns an already-closed channel.
func (h *TelemetryHub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return
	}
	h.closed = true
	for id, ch := range h.subs {
		delete(h.subs, id)
		close(ch)
	}
}
