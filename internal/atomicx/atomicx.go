package atomicx

import (
	"math"
	"sync/atomic"

	"go.withmatt.com/metrics/internal/fasttime"
)

// An Float64 is an atomic float64. The zero value is zero.
type Float64 struct {
	v atomic.Uint64
}

func (x *Float64) Load() float64 {
	return math.Float64frombits(x.v.Load())
}

func (x *Float64) Store(val float64) {
	x.v.Store(math.Float64bits(val))
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
