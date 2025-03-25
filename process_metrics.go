package metrics

// NewProcessMetricsCollector is a Collector that yields process runtime metrics.
// Metrics are prefixed with `process_`.
func NewProcessMetricsCollector() Collector {
	var c processMetricsCollector
	return &c
}

type processMetricsCollector struct{}
