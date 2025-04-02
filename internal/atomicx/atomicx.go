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
type Sum struct {
	i atomic.Uint64
	v [2]atomic.Uint64
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

	if intval := uint64(val); val == float64(intval) {
		x.i.Add(intval)
		return
	}

	//nolint:gosec
	idx := rand.Uint64() & 1
	for {
		oldBits := x.v[idx].Load()
		newBits := math.Float64bits(math.Float64frombits(oldBits) + val)
		if x.v[idx].CompareAndSwap(oldBits, newBits) {
			return
		}
	}
}

func (x *Sum) Reset() {
	x.i.Store(0)
	clear(x.v[:])
}

func (x *Sum) Load() float64 {
	return float64(x.i.Load()) +
		math.Float64frombits(x.v[0].Load()) +
		math.Float64frombits(x.v[1].Load())
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
