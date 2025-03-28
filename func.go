package metrics

// UintFunc is a uint64 value returned from a function.
type UintFunc struct {
	fn func() uint64
}

// Get returns the current value for f.
func (f *UintFunc) Get() uint64 {
	return f.fn()
}

func (f *UintFunc) marshalTo(w ExpfmtWriter, name MetricName) {
	w.WriteMetricName(name)
	w.WriteUint64(f.Get())
}

// NewUintFunc registers and returns gauge with the given name in s, which calls fn
// to obtain gauge value.
//
// family must be a Prometheus compatible identifier format.
//
// fn is an optional callback for making observations.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	NewUintFunc("family", observeFn, "label1", "value1", "label2", "value2")
//
// The returned UintFunc is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *Set) NewUintFunc(family string, fn func() uint64, tags ...string) *UintFunc {
	f := &UintFunc{fn: fn}
	s.mustRegisterMetric(f, MetricName{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
	})
	return f
}

// IntFunc is a int64 value returned from a function.
type IntFunc struct {
	fn func() int64
}

// Get returns the current value for f.
func (f *IntFunc) Get() int64 {
	return f.fn()
}

func (f *IntFunc) marshalTo(w ExpfmtWriter, name MetricName) {
	w.WriteMetricName(name)
	w.WriteInt64(f.Get())
}

// NewIntFunc registers and returns gauge with the given name in s, which calls fn
// to obtain gauge value.
//
// family must be a Prometheus compatible identifier format.
//
// fn is an optional callback for making observations.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	NewIntFunc("family", observeFn, "label1", "value1", "label2", "value2")
//
// The returned IntFunc is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *Set) NewIntFunc(family string, fn func() int64, tags ...string) *IntFunc {
	f := &IntFunc{fn: fn}
	s.mustRegisterMetric(f, MetricName{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
	})
	return f
}

// FloatFunc is a float64 value returned from a function.
type FloatFunc struct {
	fn func() float64
}

// Get returns the current value for f.
func (f *FloatFunc) Get() float64 {
	return f.fn()
}

func (f *FloatFunc) marshalTo(w ExpfmtWriter, name MetricName) {
	w.WriteMetricName(name)
	w.WriteFloat64(f.Get())
}

// NewFloatFunc registers and returns gauge with the given name in s, which calls fn
// to obtain gauge value.
//
// family must be a Prometheus compatible identifier format.
//
// fn is an optional callback for making observations.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	NewFloatFunc("family", observeFn, "label1", "value1", "label2", "value2")
//
// The returned FloatFunc is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *Set) NewFloatFunc(family string, fn func() float64, tags ...string) *FloatFunc {
	f := &FloatFunc{fn: fn}
	s.mustRegisterMetric(f, MetricName{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
	})
	return f
}
