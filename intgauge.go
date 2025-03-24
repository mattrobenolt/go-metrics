package metrics

import (
	"errors"
	"sync/atomic"
)

// IntGauge is a uint64 gauge.
type IntGauge struct {
	v  atomic.Uint64
	fn func() uint64
}

// Get returns the current value for g.
func (g *IntGauge) Get() uint64 {
	if f := g.fn; f != nil {
		return f()
	}
	return g.v.Load()
}

// Set sets g value to val.
//
// The g must be created with nil callback in order to be able to call this function.
func (g *IntGauge) Set(val uint64) {
	if g.fn != nil {
		panic(errors.New("cannot call Set on gauge created with non-nil callback"))
	}
	g.v.Store(val)
}

// Inc increments g by 1.
//
// The g must be created with nil callback in order to be able to call this function.
func (g *IntGauge) Inc() {
	g.Add(1)
}

// Dec decrements g by 1.
//
// The g must be created with nil callback in order to be able to call this function.
func (g *IntGauge) Dec() {
	g.Add(^uint64(0))
}

// Add adds val to g. val may be positive or negative.
//
// The g must be created with nil callback in order to be able to call this function.
func (g *IntGauge) Add(val uint64) {
	if g.fn != nil {
		panic(errors.New("cannot call Set on gauge created with non-nil callback"))
	}
	g.v.Add(val)
}

func (g *IntGauge) marshalTo(w ExpfmtWriter, name MetricName) {
	w.WriteMetricName(name)
	w.WriteUint64(g.Get())
}
