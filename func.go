package metrics

// Uint64Func is a uint64 value returned from a function.
type Uint64Func struct {
	fn func() uint64
}

// Get returns the current value for f.
func (f *Uint64Func) Get() uint64 {
	return f.fn()
}

func (f *Uint64Func) marshalTo(w ExpfmtWriter, name MetricName) {
	w.WriteMetricName(name)
	w.WriteUint64(f.Get())
}

// NewUint64Func creates a new UintFunc on the global Set.
// See [Set.NewUint64Func].
func NewUint64Func(name string, fn func() uint64) *Uint64Func {
	return defaultSet.NewUint64Func(name, fn)
}

// NewUint64Func registers and returns gauge with the given name in s, which calls fn
// to obtain gauge value.
//
// family must be a Prometheus compatible identifier format.
//
// fn is an optional callback for making observations.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	NewUint64Func("family", observeFn, "label1", "value1", "label2", "value2")
//
// The returned UintFunc is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *Set) NewUint64Func(family string, fn func() uint64, tags ...string) *Uint64Func {
	f := &Uint64Func{fn: fn}
	s.mustStoreMetric(f, MetricName{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
	})
	return f
}

// Int64Func is a int64 value returned from a function.
type Int64Func struct {
	fn func() int64
}

// Get returns the current value for f.
func (f *Int64Func) Get() int64 {
	return f.fn()
}

func (f *Int64Func) marshalTo(w ExpfmtWriter, name MetricName) {
	w.WriteMetricName(name)
	w.WriteInt64(f.Get())
}

// NewInt64Func creates a new Int64Func on the global Set.
// See [Set.NewInt64Func].
func NewInt64Func(name string, fn func() int64) *Int64Func {
	return defaultSet.NewInt64Func(name, fn)
}

// NewInt64Func registers and returns gauge with the given name in s, which calls fn
// to obtain gauge value.
//
// family must be a Prometheus compatible identifier format.
//
// fn is an optional callback for making observations.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	NewInt64Func("family", observeFn, "label1", "value1", "label2", "value2")
//
// The returned IntFunc is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *Set) NewInt64Func(family string, fn func() int64, tags ...string) *Int64Func {
	f := &Int64Func{fn: fn}
	s.mustStoreMetric(f, MetricName{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
	})
	return f
}

// Float64Func is a float64 value returned from a function.
type Float64Func struct {
	fn func() float64
}

// Get returns the current value for f.
func (f *Float64Func) Get() float64 {
	return f.fn()
}

func (f *Float64Func) marshalTo(w ExpfmtWriter, name MetricName) {
	w.WriteMetricName(name)
	w.WriteFloat64(f.Get())
}

// NewFloat64Func creates a new Float64Func on the global Set.
// See [Set.NewFloat64Func].
func NewFloat64Func(name string, fn func() float64) *Float64Func {
	return defaultSet.NewFloat64Func(name, fn)
}

// NewFloat64Func registers and returns gauge with the given name in s, which calls fn
// to obtain gauge value.
//
// family must be a Prometheus compatible identifier format.
//
// fn is an optional callback for making observations.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	NewFloat64Func("family", observeFn, "label1", "value1", "label2", "value2")
//
// The returned FloatFunc is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *Set) NewFloat64Func(family string, fn func() float64, tags ...string) *Float64Func {
	f := &Float64Func{fn: fn}
	s.mustStoreMetric(f, MetricName{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
	})
	return f
}
