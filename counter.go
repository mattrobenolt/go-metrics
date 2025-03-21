package metrics

import (
	"sync/atomic"
)

// Counter is a counter.
//
// It may be used as a gauge if Dec and Set are called.
type Counter struct {
	v atomic.Uint64
}

// Inc increments c by 1.
func (c *Counter) Inc() {
	c.v.Add(1)
}

// Dec decrements c by 1.
func (c *Counter) Dec() {
	c.v.Add(^uint64(0))
}

// Add adds delta to c.
func (c *Counter) Add(delta uint64) {
	c.v.Add(delta)
}

// Get returns the current value for c.
func (c *Counter) Get() uint64 {
	return c.v.Load()
}

// Set sets c value to val.
func (c *Counter) Set(val uint64) {
	c.v.Store(val)
}

func (c *Counter) marshalTo(w ExpfmtWriter, family Ident, tags ...Tag) {
	w.WriteMetricName(family, tags...)
	w.WriteUint64(c.Get())
}
