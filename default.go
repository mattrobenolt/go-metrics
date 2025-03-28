package metrics

import "io"

var defaultSet Set

// ResetDefaultSet results the default global Set.
// See [Set.Reset].
func ResetDefaultSet() {
	defaultSet.Reset()
}

// RegisterDefaultCollectors registers the default Collectors
// onto the global Set.
func RegisterDefaultCollectors() {
	RegisterCollector(NewGoMetricsCollector())
	RegisterCollector(NewProcessMetricsCollector())
}

// RegisterCollector registers a Collector onto the global Set.
// See [Set.RegisterCollector].
func RegisterCollector(c Collector) {
	defaultSet.RegisterCollector(c)
}

// WritePrometheus writes the global Set to io.Writer.
// See [Set.WritePrometheus].
func WritePrometheus(w io.Writer) (int, error) {
	return defaultSet.WritePrometheus(w)
}

// NewCounter creates a new Counter on the global Set.
// See [Set.NewCounter].
func NewCounter(family string, tags ...string) *Counter {
	return defaultSet.NewCounter(family, tags...)
}

// NewCounterVec creates a new CounterVec on the global Set.
// See [Set.NewCounterVec].
func NewCounterVec(name VecName) *CounterVec {
	return defaultSet.NewCounterVec(name)
}

// NewFloatCounter creates a new FloatCounter on the global Set.
// See [Set.NewFloatCounter].
func NewFloatCounter(family string, tags ...string) *FloatCounter {
	return defaultSet.NewFloatCounter(family, tags...)
}

// NewFloatCounterVec creates a new FloatCounterVec on the global Set.
// See [Set.NewFloatCounterVec].
func NewFloatCounterVec(name VecName) *FloatCounterVec {
	return defaultSet.NewFloatCounterVec(name)
}

// NewHistogram creates a new Histogram on the global Set.
// See [Set.NewHistogram].
func NewHistogram(family string, tags ...string) *Histogram {
	return defaultSet.NewHistogram(family, tags...)
}

// NewHistogramVec creates a new HistogramVec on the global Set.
// See [Set.NewHistogramVec].
func NewHistogramVec(name VecName) *HistogramVec {
	return defaultSet.NewHistogramVec(name)
}

// NewFixedHistogram creates a new FixedHistogram on the global Set.
// See [Set.NewFixedHistogram].
func NewFixedHistogram(family string, buckets []float64, tags ...string) *FixedHistogram {
	return defaultSet.NewFixedHistogram(family, buckets, tags...)
}

// NewFixedHistogramVec creates a new FixedHistogramVec on the global Set.
// See [Set.NewFixedHistogramVec].
func NewFixedHistogramVec(opt FixedHistogramVecOpt) *FixedHistogramVec {
	return defaultSet.NewFixedHistogramVec(opt)
}

// NewFloatFunc creates a new FloatFunc on the global Set.
// See [Set.NewFloatFunc].
func NewFloatFunc(name string, fn func() float64) *FloatFunc {
	return defaultSet.NewFloatFunc(name, fn)
}

// NewIntFunc creates a new IntFunc on the global Set.
// See [Set.NewIntFunc].
func NewIntFunc(name string, fn func() int64) *IntFunc {
	return defaultSet.NewIntFunc(name, fn)
}

// NewUintFunc creates a new UintFunc on the global Set.
// See [Set.NewUintFunc].
func NewUintFunc(name string, fn func() uint64) *UintFunc {
	return defaultSet.NewUintFunc(name, fn)
}
