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

// NewCounterOpt creates a new Counter on the global Set.
// See [Set.NewCounter].
func NewCounterOpt(opt CounterOpt) *Counter {
	return defaultSet.NewCounterOpt(opt)
}

// GetOrCreateCounter gets or creates a Counter on the global Set.
// See [Set.GetOrCreateCounter].
func GetOrCreateCounter(family string, tags ...string) *Counter {
	return defaultSet.GetOrCreateCounter(family, tags...)
}

// NewCounterVec creates a new CounterVec on the global Set.
// See [Set.NewCounterVec].
func NewCounterVec(opt CounterVecOpt) *CounterVec {
	return defaultSet.NewCounterVec(opt)
}

// NewFloatCounter creates a new FloatCounter on the global Set.
// See [Set.NewFloatCounter].
func NewFloatCounter(family string, tags ...string) *FloatCounter {
	return defaultSet.NewFloatCounter(family, tags...)
}

// NewFloatCounterOpt creates a new FloatCounter on the global Set.
// See [Set.NewFloatCounterOpt].
func NewFloatCounterOpt(opt FloatCounterOpt) *FloatCounter {
	return defaultSet.NewFloatCounterOpt(opt)
}

// GetOrCreateFloatCounter gets or creates a new FloatCounter on the global Set.
// See [Set.GetOrCreateFloatCounter].
func GetOrCreateFloatCounter(family string, tags ...string) *FloatCounter {
	return defaultSet.GetOrCreateFloatCounter(family, tags...)
}

// NewFloatCounterVec creates a new FloatCounterVec on the global Set.
// See [Set.NewFloatCounterVec].
func NewFloatCounterVec(opt FloatCounterVecOpt) *FloatCounterVec {
	return defaultSet.NewFloatCounterVec(opt)
}

// NewGauge creates a new Gauge on the global Set.
// See [Set.NewGauge].
func NewGauge(family string, fn func() float64, tags ...string) *Gauge {
	return defaultSet.NewGauge(family, fn, tags...)
}

// NewGaugeOpt creates a new Gauge on the global Set.
// See [Set.NewGaugeOpt].
func NewGaugeOpt(opt GaugeOpt) *Gauge {
	return defaultSet.NewGaugeOpt(opt)
}

// GetOrCreateGauge gets or creates a new Gauge on the global Set.
// See [Set.GetOrCreateGauge].
func GetOrCreateGauge(family string, tags ...string) *Gauge {
	return defaultSet.GetOrCreateGauge(family, tags...)
}

// NewGaugeVec creates a new GaugeVec on the global Set.
// See [Set.NewGaugeVec].
func NewGaugeVec(opt GaugeVecOpt) *GaugeVec {
	return defaultSet.NewGaugeVec(opt)
}

// NewIntGauge creates a new IntGauge on the global Set.
// See [Set.NewIntGauge].
func NewIntGauge(family string, fn func() uint64, tags ...string) *IntGauge {
	return defaultSet.NewIntGauge(family, fn, tags...)
}

// NewIntGaugeOpt creates a new IntGauge on the global Set.
// See [Set.NewIntGaugeOpt].
func NewIntGaugeOpt(opt IntGaugeOpt) *IntGauge {
	return defaultSet.NewIntGaugeOpt(opt)
}

// GetOrCreateIntGauge gets or creates a new IntGauge on the global Set.
// See [Set.GetOrCreateIntGauge].
func GetOrCreateIntGauge(family string, tags ...string) *IntGauge {
	return defaultSet.GetOrCreateIntGauge(family, tags...)
}

// NewIntGaugeVec creates a new IntGaugeVec on the global Set.
// See [Set.NewIntGaugeVec].
func NewIntGaugeVec(opt IntGaugeVecOpt) *IntGaugeVec {
	return defaultSet.NewIntGaugeVec(opt)
}

// NewHistogram creates a new Histogram on the global Set.
// See [Set.NewHistogram].
func NewHistogram(family string, tags ...string) *Histogram {
	return defaultSet.NewHistogram(family, tags...)
}

// NewHistogramOpt creates a new Histogram on the global Set.
// See [Set.NewHistogramOpt].
func NewHistogramOpt(opt HistogramOpt) *Histogram {
	return defaultSet.NewHistogramOpt(opt)
}

// GetOrCreateHistogram gets or creates a new Histogram on the global Set.
// See [Set.GetOrCreateHistogram].
func GetOrCreateHistogram(family string, tags ...string) *Histogram {
	return defaultSet.GetOrCreateHistogram(family, tags...)
}

// NewHistogramVec creates a new HistogramVec on the global Set.
// See [Set.NewHistogramVec].
func NewHistogramVec(opt HistogramVecOpt) *HistogramVec {
	return defaultSet.NewHistogramVec(opt)
}

// NewFixedHistogram creates a new FixedHistogram on the global Set.
// See [Set.NewFixedHistogram].
func NewFixedHistogram(family string, buckets []float64, tags ...string) *FixedHistogram {
	return defaultSet.NewFixedHistogram(family, buckets, tags...)
}

// NewFixedHistogramOpt creates a new FixedHistogram on the global Set.
// See [Set.NewFixedHistogramOpt].
func NewFixedHistogramOpt(opt FixedHistogramOpt) *FixedHistogram {
	return defaultSet.NewFixedHistogramOpt(opt)
}

// GetOrCreateFixedHistogram gets or creates a new FixedHistogram on the global Set.
// See [Set.GetOrCreateFixedHistogram].
func GetOrCreateFixedHistogram(family string, buckets []float64, tags ...string) *FixedHistogram {
	return defaultSet.GetOrCreateFixedHistogram(family, buckets, tags...)
}

// NewFixedHistogramVec creates a new HistogramVec on the global Set.
// See [Set.NewHistogramVec].
func NewFixedHistogramVec(opt FixedHistogramVecOpt) *FixedHistogramVec {
	return defaultSet.NewFixedHistogramVec(opt)
}
