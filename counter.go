package metrics

import (
	"sync/atomic"

	"go.withmatt.com/metrics/internal/atomicx"
)

// Uint is a uint64 counter.
//
// It may be used as a gauge if Dec and Set are called.
type Uint struct {
	v atomic.Uint64
}

// Inc increments c by 1.
func (c *Uint) Inc() {
	c.v.Add(1)
}

// Dec decrements c by 1.
func (c *Uint) Dec() {
	c.v.Add(^uint64(0))
}

// Add adds delta to c.
func (c *Uint) Add(delta uint64) {
	c.v.Add(delta)
}

// Get returns the current value for c.
func (c *Uint) Get() uint64 {
	return c.v.Load()
}

// Set sets c value to val.
func (c *Uint) Set(val uint64) {
	c.v.Store(val)
}

func (c *Uint) marshalTo(w ExpfmtWriter, name MetricName) {
	w.WriteMetricName(name)
	w.WriteUint64(c.Get())
}

// NewUint registers and returns new Counter with the given name in the s.
//
// family must be a Prometheus compatible identifier format.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	NewUint("family", "label1", "value1", "label2", "value2")
//
// The returned Uint is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *Set) NewUint(family string, tags ...string) *Uint {
	c := &Uint{}
	s.mustRegisterMetric(c, MetricName{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
	})
	return c
}

// NewCounter is an alias for [Set.NewUint].
func (s *Set) NewCounter(family string, tags ...string) *Uint {
	return s.NewUint(family, tags...)
}

// Int is an int64 counter.
//
// It may be used as a gauge if Dec and Set are called.
type Int struct {
	v atomic.Int64
}

// Inc increments c by 1.
func (c *Int) Inc() {
	c.v.Add(1)
}

// Dec decrements c by 1.
func (c *Int) Dec() {
	c.v.Add(-1)
}

// Add adds delta to c.
func (c *Int) Add(delta int64) {
	c.v.Add(delta)
}

// Get returns the current value for c.
func (c *Int) Get() int64 {
	return c.v.Load()
}

// Set sets c value to val.
func (c *Int) Set(val int64) {
	c.v.Store(val)
}

func (c *Int) marshalTo(w ExpfmtWriter, name MetricName) {
	w.WriteMetricName(name)
	w.WriteInt64(c.Get())
}

// NewInt registers and returns new Int with the given name in the s.
//
// family must be a Prometheus compatible identifier format.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	NewInt("family", "label1", "value1", "label2", "value2")
//
// The returned Int is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *Set) NewInt(family string, tags ...string) *Int {
	c := &Int{}
	s.mustRegisterMetric(c, MetricName{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
	})
	return c
}

// Float is a float64 counter.
//
// It may be used as a gauge if Dec and Set are called.
type Float struct {
	v atomicx.Float64
}

// Inc increments c by 1.
func (c *Float) Inc() {
	c.v.Add(1)
}

// Dec decrements c by 1.
func (c *Float) Dec() {
	c.v.Add(-1)
}

// Add adds delta to c.
func (c *Float) Add(delta float64) {
	c.v.Add(delta)
}

// Get returns the current value for c.
func (c *Float) Get() float64 {
	return c.v.Load()
}

// Set sets c value to val.
func (c *Float) Set(val float64) {
	c.v.Store(val)
}

func (c *Float) marshalTo(w ExpfmtWriter, name MetricName) {
	w.WriteMetricName(name)
	w.WriteFloat64(c.Get())
}

// NewFloat registers and returns new Float with the given name in the s.
//
// family must be a Prometheus compatible identifier format.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	NewFloat("family", "label1", "value1", "label2", "value2")
//
// The returned Float is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *Set) NewFloat(family string, tags ...string) *Float {
	c := &Float{}
	s.mustRegisterMetric(c, MetricName{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
	})
	return c
}
