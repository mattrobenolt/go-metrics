package metrics

func NewProcessMetricsCollector() Collector {
	var c processMetricsCollector
	return &c
}

type processMetricsCollector struct{}
