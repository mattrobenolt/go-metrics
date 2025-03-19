package metrics

import (
	"fmt"
	"io"
	"math"
	"sync/atomic"
)

// Gauge is a float64 gauge.
type Gauge struct {
	// valueBits contains uint64 representation of float64 passed to Gauge.Set.
	valueBits atomic.Uint64

	// f is a callback, which is called for returning the gauge value.
	f func() float64
}

// Get returns the current value for g.
func (g *Gauge) Get() float64 {
	if f := g.f; f != nil {
		return f()
	}
	n := g.valueBits.Load()
	return math.Float64frombits(n)
}

// Set sets g value to v.
//
// The g must be created with nil callback in order to be able to call this function.
func (g *Gauge) Set(v float64) {
	if g.f != nil {
		panic(fmt.Errorf("cannot call Set on gauge created with non-nil callback"))
	}
	n := math.Float64bits(v)
	g.valueBits.Store(n)
}

// Inc increments g by 1.
//
// The g must be created with nil callback in order to be able to call this function.
func (g *Gauge) Inc() {
	g.Add(1)
}

// Dec decrements g by 1.
//
// The g must be created with nil callback in order to be able to call this function.
func (g *Gauge) Dec() {
	g.Add(-1)
}

// Add adds fAdd to g. fAdd may be positive and negative.
//
// The g must be created with nil callback in order to be able to call this function.
func (g *Gauge) Add(fAdd float64) {
	if g.f != nil {
		panic(fmt.Errorf("cannot call Set on gauge created with non-nil callback"))
	}
	for {
		n := g.valueBits.Load()
		f := math.Float64frombits(n)
		fNew := f + fAdd
		nNew := math.Float64bits(fNew)
		if g.valueBits.CompareAndSwap(n, nNew) {
			break
		}
	}
}

func (g *Gauge) marshalTo(prefix string, w io.Writer) {
	v := g.Get()
	if float64(int64(v)) == v {
		// Marshal integer values without scientific notation
		fmt.Fprintf(w, "%s %d\n", prefix, int64(v))
	} else {
		fmt.Fprintf(w, "%s %g\n", prefix, v)
	}
}
