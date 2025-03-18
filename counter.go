package metrics

import (
	"fmt"
	"io"
	"sync/atomic"
)

// NewCounter registers and returns new counter with the given name.
//
// name must be valid Prometheus-compatible metric with possible labels.
// For instance,
//
//   - foo
//   - foo{bar="baz"}
//   - foo{bar="baz",aaa="b"}
//
// The returned counter is safe to use from concurrent goroutines.
func NewCounter(name string) *Counter {
	return defaultSet.NewCounter(name)
}

// Counter is a counter.
//
// It may be used as a gauge if Dec and Set are called.
type Counter struct {
	n atomic.Uint64
}

// Inc increments c.
func (c *Counter) Inc() {
	c.n.Add(1)
}

// Dec decrements c.
func (c *Counter) Dec() {
	c.n.Add(^uint64(0))
}

// Add adds n to c.
func (c *Counter) Add(n int) {
	c.n.Add(uint64(n))
}

// AddInt64 adds n to c.
func (c *Counter) AddInt64(n int64) {
	c.n.Add(uint64(n))
}

// Get returns the current value for c.
func (c *Counter) Get() uint64 {
	return c.n.Load()
}

// Set sets c value to n.
func (c *Counter) Set(n uint64) {
	c.n.Store(n)
}

// marshalTo marshals c with the given prefix to w.
func (c *Counter) marshalTo(prefix string, w io.Writer) {
	v := c.Get()
	fmt.Fprintf(w, "%s %d\n", prefix, v)
}

func (c *Counter) metricType() string {
	return "counter"
}

// GetOrCreateCounter returns registered counter with the given name
// or creates new counter if the registry doesn't contain counter with
// the given name.
//
// name must be valid Prometheus-compatible metric with possible labels.
// For instance,
//
//   - foo
//   - foo{bar="baz"}
//   - foo{bar="baz",aaa="b"}
//
// The returned counter is safe to use from concurrent goroutines.
//
// Performance tip: prefer NewCounter instead of GetOrCreateCounter.
func GetOrCreateCounter(name string) *Counter {
	return defaultSet.GetOrCreateCounter(name)
}
