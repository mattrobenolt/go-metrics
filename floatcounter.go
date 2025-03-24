package metrics

import "go.withmatt.com/metrics/internal/atomicx"

// FloatCounter is a float64 counter.
//
// It may be used as a gauge if Dec and Set are called.
type FloatCounter struct {
	v atomicx.Float64
}

// Inc increments c by 1.
func (c *FloatCounter) Inc() {
	c.v.Add(1)
}

// Dec decrements c by 1.
func (c *FloatCounter) Dec() {
	c.v.Add(-1)
}

// Add adds delta to c.
func (c *FloatCounter) Add(delta float64) {
	c.v.Add(delta)
}

// Get returns the current value for c.
func (c *FloatCounter) Get() float64 {
	return c.v.Load()
}

// Set sets c value to val.
func (c *FloatCounter) Set(val float64) {
	c.v.Store(val)
}

func (c *FloatCounter) marshalTo(w ExpfmtWriter, name MetricName) {
	w.WriteMetricName(name)
	w.WriteFloat64(c.Get())
}
