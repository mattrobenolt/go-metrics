package metrics

import (
	"sync/atomic"

	"go.withmatt.com/metrics/internal/atomicx"
)

// Uint64 is a uint64 counter.
//
// It may be used as a gauge if Dec and Set are called.
type Uint64 struct {
	v atomic.Uint64
}

// Inc increments c by 1.
func (c *Uint64) Inc() {
	c.v.Add(1)
}

// Dec decrements c by 1.
func (c *Uint64) Dec() {
	c.v.Add(^uint64(0))
}

// Add adds delta to c.
func (c *Uint64) Add(delta uint64) {
	c.v.Add(delta)
}

// Get returns the current value for c.
func (c *Uint64) Get() uint64 {
	return c.v.Load()
}

// Set sets c value to val.
func (c *Uint64) Set(val uint64) {
	c.v.Store(val)
}

func (c *Uint64) marshalTo(w ExpfmtWriter, name MetricName) {
	w.WriteMetricName(name)
	w.WriteUint64(c.Get())
}

// NewCounter creates a new Uint on the global Set.
// See [Set.NewUint64].
func NewCounter(family string, tags ...string) *Uint64 {
	return defaultSet.NewCounter(family, tags...)
}

// NewUint64 creates a new Uint on the global Set.
// See [Set.NewUint64].
func NewUint64(family string, tags ...string) *Uint64 {
	return defaultSet.NewUint64(family, tags...)
}

// NewUint64 registers and returns new Counter with the given name in the s.
//
// family must be a Prometheus compatible identifier format.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	NewUint64("family", "label1", "value1", "label2", "value2")
//
// The returned Uint is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *Set) NewUint64(family string, tags ...string) *Uint64 {
	c := &Uint64{}
	s.mustRegisterMetric(c, MetricName{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
	})
	return c
}

// NewCounter is an alias for [Set.NewUint64].
func (s *Set) NewCounter(family string, tags ...string) *Uint64 {
	return s.NewUint64(family, tags...)
}

// Int64 is an int64 counter.
//
// It may be used as a gauge if Dec and Set are called.
type Int64 struct {
	v atomic.Int64
}

// Inc increments c by 1.
func (c *Int64) Inc() {
	c.v.Add(1)
}

// Dec decrements c by 1.
func (c *Int64) Dec() {
	c.v.Add(-1)
}

// Add adds delta to c.
func (c *Int64) Add(delta int64) {
	c.v.Add(delta)
}

// Get returns the current value for c.
func (c *Int64) Get() int64 {
	return c.v.Load()
}

// Set sets c value to val.
func (c *Int64) Set(val int64) {
	c.v.Store(val)
}

func (c *Int64) marshalTo(w ExpfmtWriter, name MetricName) {
	w.WriteMetricName(name)
	w.WriteInt64(c.Get())
}

// NewInt64 creates a new Int on the global Set.
// See [Set.NewInt64].
func NewInt64(family string, tags ...string) *Int64 {
	return defaultSet.NewInt64(family, tags...)
}

// NewInt64 registers and returns new Int with the given name in the s.
//
// family must be a Prometheus compatible identifier format.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	NewInt64("family", "label1", "value1", "label2", "value2")
//
// The returned Int is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *Set) NewInt64(family string, tags ...string) *Int64 {
	c := &Int64{}
	s.mustRegisterMetric(c, MetricName{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
	})
	return c
}

// Float64 is a float64 counter.
//
// It may be used as a gauge if Dec and Set are called.
type Float64 struct {
	v atomicx.Float64
}

// Inc increments c by 1.
func (c *Float64) Inc() {
	c.v.Add(1)
}

// Dec decrements c by 1.
func (c *Float64) Dec() {
	c.v.Add(-1)
}

// Add adds delta to c.
func (c *Float64) Add(delta float64) {
	c.v.Add(delta)
}

// Get returns the current value for c.
func (c *Float64) Get() float64 {
	return c.v.Load()
}

// Set sets c value to val.
func (c *Float64) Set(val float64) {
	c.v.Store(val)
}

func (c *Float64) marshalTo(w ExpfmtWriter, name MetricName) {
	w.WriteMetricName(name)
	w.WriteFloat64(c.Get())
}

// NewFloat64 creates a new Float on the global Set.
// See [Set.NewFloat64].
func NewFloat64(family string, tags ...string) *Float64 {
	return defaultSet.NewFloat64(family, tags...)
}

// NewFloat64 registers and returns new Float with the given name in the s.
//
// family must be a Prometheus compatible identifier format.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	NewFloat64("family", "label1", "value1", "label2", "value2")
//
// The returned Float is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *Set) NewFloat64(family string, tags ...string) *Float64 {
	c := &Float64{}
	s.mustRegisterMetric(c, MetricName{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
	})
	return c
}
