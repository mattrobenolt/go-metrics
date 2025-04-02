package metrics

import "weak"

type selfMetricsCollector struct{}

// NewSelfMetricsCollector is a Collector that yields our own runtime metrics.
// Metrics are prefixed with `gometrics_`.
func NewSelfMetricsCollector() Collector {
	return &selfMetricsCollector{}
}

func (*selfMetricsCollector) Collect(w ExpfmtWriter) {
	var size uint64
	rangeIdentCache(func(_ uint64, _ weak.Pointer[string]) bool {
		size++
		return true
	})
	w.WriteLazyMetricUint64("gometrics_ident_cache_size", size)
}
