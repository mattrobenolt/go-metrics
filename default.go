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

// NewCounter creates a new Uint on the global Set.
// See [Set.NewUint].
func NewCounter(family string, tags ...string) *Uint {
	return defaultSet.NewCounter(family, tags...)
}

// NewCounterVec creates a new UintVec on the global Set.
// See [Set.NewUintVec].
func NewCounterVec(family string, labels ...string) *UintVec {
	return defaultSet.NewCounterVec(family, labels...)
}

// NewUint creates a new Uint on the global Set.
// See [Set.NewUint].
func NewUint(family string, tags ...string) *Uint {
	return defaultSet.NewUint(family, tags...)
}

// NewUintVec creates a new UintVec on the global Set.
// See [Set.NewUintVec].
func NewUintVec(family string, labels ...string) *UintVec {
	return defaultSet.NewUintVec(family, labels...)
}

// NewInt creates a new Int on the global Set.
// See [Set.NewInt].
func NewInt(family string, tags ...string) *Int {
	return defaultSet.NewInt(family, tags...)
}

// NewIntVec creates a new IntVec on the global Set.
// See [Set.NewIntVec].
func NewIntVec(family string, labels ...string) *IntVec {
	return defaultSet.NewIntVec(family, labels...)
}

// NewFloat creates a new Float on the global Set.
// See [Set.NewFloat].
func NewFloat(family string, tags ...string) *Float {
	return defaultSet.NewFloat(family, tags...)
}

// NewFloatVec creates a new FloatVec on the global Set.
// See [Set.NewFloatVec].
func NewFloatVec(family string, labels ...string) *FloatVec {
	return defaultSet.NewFloatVec(family, labels...)
}

// NewHistogram creates a new Histogram on the global Set.
// See [Set.NewHistogram].
func NewHistogram(family string, tags ...string) *Histogram {
	return defaultSet.NewHistogram(family, tags...)
}

// NewHistogramVec creates a new HistogramVec on the global Set.
// See [Set.NewHistogramVec].
func NewHistogramVec(family string, labels ...string) *HistogramVec {
	return defaultSet.NewHistogramVec(family, labels...)
}

// NewFixedHistogram creates a new FixedHistogram on the global Set.
// See [Set.NewFixedHistogram].
func NewFixedHistogram(family string, buckets []float64, tags ...string) *FixedHistogram {
	return defaultSet.NewFixedHistogram(family, buckets, tags...)
}

// NewFixedHistogramVec creates a new FixedHistogramVec on the global Set.
// See [Set.NewFixedHistogramVec].
func NewFixedHistogramVec(family string, buckets []float64, labels ...string) *FixedHistogramVec {
	return defaultSet.NewFixedHistogramVec(family, buckets, labels...)
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
