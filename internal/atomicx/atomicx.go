package atomicx

import (
	"math"
	"math/rand/v2"
	"sync/atomic"

	"go.withmatt.com/metrics/internal/fasttime"
)

// A Float64 is an atomic float64. The zero value is zero.
type Float64 struct {
	v atomic.Uint64
}

func (x *Float64) Add(val float64) {
	if val == 0 {
		return
	}
	for {
		oldBits := x.v.Load()
		newBits := math.Float64bits(math.Float64frombits(oldBits) + val)
		if x.v.CompareAndSwap(oldBits, newBits) {
			return
		}
	}
}

func (x *Float64) Load() float64 {
	return math.Float64frombits(x.v.Load())
}

func (x *Float64) Store(val float64) {
	x.v.Store(math.Float64bits(val))
}

// A Sum is an atomic float64 that can only Add. The zero value is zero.
// Values cannot be negative.
type Sum struct {
	// the integer part of the sum
	i atomic.Uint64

	// the floating point part of the sum, but this is split
	// across two atomic.Uint64 values to avoid contention
	f [2]atomic.Uint64
}

func (x *Sum) Inc() {
	x.i.Add(1)
}

func (x *Sum) Dec() {
	x.i.Add(^uint64(0))
}

func (x *Sum) AddUint64(val uint64) {
	if val == 0 {
		return
	}
	x.i.Add(val)
}

func (x *Sum) Add(val float64) {
	if val <= 0 {
		return
	}

	if val == 1 {
		x.Inc()
		return
	}

	// if we're less than 1, we must be a float
	if val >= 1 {
		// Fast path to first check if the value is actually a whole number in
		// disguise. This extra check is a few nanoseconds cost, but the benefit
		// of avoiding the CAS loop is significant especially under pressure.
		if intval := uint64(val); val == float64(intval) {
			x.i.Add(intval)
			return
		}
	}

	// Choose a random index to update the floating point part of the sum.
	// This is done to avoid contention when multiple goroutines are trying
	// to update the sum concurrently.
	//
	//nolint:gosec
	idx := rand.Uint64() & 1 // & 1 is getting us 0 or 1, with a single CPU instruction

	for {
		oldBits := x.f[idx].Load()
		newBits := math.Float64bits(math.Float64frombits(oldBits) + val)
		if x.f[idx].CompareAndSwap(oldBits, newBits) {
			return
		}
	}
}

// Reset resets the sum to zero.
func (x *Sum) Reset() {
	x.i.Store(0)
	clear(x.f[:])
}

func (x *Sum) Load() float64 {
	return float64(x.i.Load()) +
		math.Float64frombits(x.f[0].Load()) +
		math.Float64frombits(x.f[1].Load())
}

// An Instant is an atomic fasttime.Instant. The zero value is zero.
type Instant struct {
	v atomic.Int64
}

func (x *Instant) Load() fasttime.Instant {
	return fasttime.Instant(x.v.Load())
}

func (x *Instant) Store(val fasttime.Instant) {
	x.v.Store(int64(val))
}
