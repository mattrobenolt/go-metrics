package metrics

import "io"

var defaultSet Set

func ResetDefaultSet() {
	defaultSet.Reset()
}

func RegisterDefaultCollectors() {
	RegisterCollector(NewGoMetricsCollector())
	RegisterCollector(NewProcessMetricsCollector())
}

func RegisterCollector(c Collector) {
	defaultSet.RegisterCollector(c)
}

func WritePrometheus(w io.Writer) (int, error) {
	return defaultSet.WritePrometheus(w)
}

func NewCounter(family string, tags ...string) *Counter {
	return defaultSet.NewCounter(family, tags...)
}

func NewCounterOpt(opt CounterOpt) *Counter {
	return defaultSet.NewCounterOpt(opt)
}

func GetOrCreateCounter(family string, tags ...string) *Counter {
	return defaultSet.GetOrCreateCounter(family, tags...)
}

func NewCounterVec(opt CounterVecOpt) *CounterVec {
	return defaultSet.NewCounterVec(opt)
}

func NewFloatCounter(family string, tags ...string) *FloatCounter {
	return defaultSet.NewFloatCounter(family, tags...)
}

func NewFloatCounterOpt(opt FloatCounterOpt) *FloatCounter {
	return defaultSet.NewFloatCounterOpt(opt)
}

func GetOrCreateFloatCounter(family string, tags ...string) *FloatCounter {
	return defaultSet.GetOrCreateFloatCounter(family, tags...)
}

func NewFloatCounterVec(opt FloatCounterVecOpt) *FloatCounterVec {
	return defaultSet.NewFloatCounterVec(opt)
}

func NewGauge(family string, fn func() float64, tags ...string) *Gauge {
	return defaultSet.NewGauge(family, fn, tags...)
}

func NewGaugeOpt(opt GaugeOpt) *Gauge {
	return defaultSet.NewGaugeOpt(opt)
}

func GetOrCreateGauge(family string, tags ...string) *Gauge {
	return defaultSet.GetOrCreateGauge(family, tags...)
}

func NewGaugeVec(opt GaugeVecOpt) *GaugeVec {
	return defaultSet.NewGaugeVec(opt)
}

func NewIntGauge(family string, fn func() uint64, tags ...string) *IntGauge {
	return defaultSet.NewIntGauge(family, fn, tags...)
}

func NewIntGaugeOpt(opt IntGaugeOpt) *IntGauge {
	return defaultSet.NewIntGaugeOpt(opt)
}

func GetOrCreateIntGauge(family string, tags ...string) *IntGauge {
	return defaultSet.GetOrCreateIntGauge(family, tags...)
}

func NewIntGaugeVec(opt IntGaugeVecOpt) *IntGaugeVec {
	return defaultSet.NewIntGaugeVec(opt)
}

func NewHistogram(family string, tags ...string) *Histogram {
	return defaultSet.NewHistogram(family, tags...)
}

func NewHistogramOpt(opt HistogramOpt) *Histogram {
	return defaultSet.NewHistogramOpt(opt)
}

func GetOrCreateHistogram(family string, tags ...string) *Histogram {
	return defaultSet.GetOrCreateHistogram(family, tags...)
}

func NewHistogramVec(opt HistogramVecOpt) *HistogramVec {
	return defaultSet.NewHistogramVec(opt)
}
