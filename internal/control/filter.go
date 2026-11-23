package control

import (
	"fmt"
	"sync"
	"time"

	"github.com/brandondube/pctl"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

// DraftFilter wraps github.com/brandondube/pctl's first-order low-pass filter (pctl.LPF)
// to eliminate three upstream hazards:
//  1. Startup jump from 0 Pa to negative draft (e.g. -12.45 Pa): first valid sample seeds the filter.
//  2. Missing Reset() method in pctl.LPF: Reset() clears seeding and rebuilds the underlying LPF.
//  3. Divide-by-zero when cutoffFreq == 0: NewDraftFilter enforces finite cutoffHz > 0 and dt > 0.
//
// Note: pctl.HPF is intentionally NOT used anywhere in this package due to an upstream bug
// where HPF.Update ignores its input argument.
type DraftFilter struct {
	mu       sync.RWMutex
	cutoffHz float64
	dtSec    float64
	lpf      *pctl.LPF
	seeded   bool
	seed     float64
	value    model.Pascal
}

// NewDraftFilter constructs a low-pass filter for kiln draft pressure readings.
// cutoffHz must be finite and strictly > 0; dt must be > 0.
func NewDraftFilter(cutoffHz float64, dt time.Duration) (*DraftFilter, error) {
	if !isFinite(cutoffHz) || cutoffHz <= 0 {
		return nil, fmt.Errorf("draft filter cutoff frequency must be finite and > 0 Hz, got %v", cutoffHz)
	}
	if dt <= 0 {
		return nil, fmt.Errorf("draft filter sampling interval must be > 0, got %v", dt)
	}
	dtSec := dt.Seconds()
	return &DraftFilter{
		cutoffHz: cutoffHz,
		dtSec:    dtSec,
		lpf:      pctl.NewLPF(cutoffHz, dtSec),
	}, nil
}

// Update filters rawPa using the configured nominal sampling interval.
// The first finite sample seeds the filter output directly (no startup jump from 0 Pa).
// Non-finite samples (NaN/Inf) are ignored so internal filter state cannot be poisoned.
func (f *DraftFilter) Update(rawPa model.Pascal) model.Pascal {
	return f.UpdateWithDT(rawPa, 0)
}

// UpdateWithDT filters rawPa using dt when dt > 0 (otherwise using the nominal sampling interval).
// Because pctl.LPF is a linear time-invariant first-order filter with prev initialized to 0,
// shifting the input by the first sample x0 (y = x0 + lpf.Update(x - x0)) is algebraically
// identical to initializing prev = x0 while delegating the discrete RC filtering to pctl.LPF.
func (f *DraftFilter) UpdateWithDT(rawPa model.Pascal, dt time.Duration) model.Pascal {
	f.mu.Lock()
	defer f.mu.Unlock()

	x := float64(rawPa)
	if !isFinite(x) {
		return f.value
	}

	stepDT := f.dtSec
	if dt > 0 {
		stepDT = dt.Seconds()
	}

	if !f.seeded {
		f.seeded = true
		f.seed = x
		f.lpf = pctl.NewLPF(f.cutoffHz, stepDT)
		f.value = rawPa
		return f.value
	}

	f.lpf.DT = stepDT
	filtered := f.seed + f.lpf.Update(x-f.seed)
	f.value = model.Pascal(filtered)
	return f.value
}

// Reset clears the filter state so the next finite sample re-seeds the filter cleanly.
func (f *DraftFilter) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.seeded = false
	f.seed = 0
	f.value = 0
	f.lpf = pctl.NewLPF(f.cutoffHz, f.dtSec)
}

// Value returns the most recent filtered draft pressure in Pascals.
func (f *DraftFilter) Value() model.Pascal {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.value
}

// Seeded reports whether the filter has been seeded by at least one finite sample since creation or Reset.
func (f *DraftFilter) Seeded() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.seeded
}

// CutoffHz returns the configured -3 dB cutoff frequency in Hz.
func (f *DraftFilter) CutoffHz() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.cutoffHz
}
