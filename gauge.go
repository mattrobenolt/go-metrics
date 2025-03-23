package metrics

import (
	"errors"

	"go.withmatt.com/metrics/internal/atomicx"
)

// Gauge is a float64 gauge.
type Gauge struct {
	v  atomicx.Float64
	fn func() float64
}

// Get returns the current value for g.
func (g *Gauge) Get() float64 {
	if f := g.fn; f != nil {
		return f()
	}
	return g.v.Load()
}

// Set sets g value to val.
//
// The g must be created with nil callback in order to be able to call this function.
func (g *Gauge) Set(val float64) {
	if g.fn != nil {
		panic(errors.New("cannot call Set on gauge created with non-nil callback"))
	}
	g.v.Store(val)
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

// Add adds val to g. val may be positive or negative.
//
// The g must be created with nil callback in order to be able to call this function.
func (g *Gauge) Add(val float64) {
	if g.fn != nil {
		panic(errors.New("cannot call Set on gauge created with non-nil callback"))
	}
	g.v.Add(val)
}

func (g *Gauge) marshalTo(w ExpfmtWriter, name MetricName) {
	w.WriteMetricName(name)
	w.WriteFloat64(g.Get())
}
