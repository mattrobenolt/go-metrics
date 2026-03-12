package metrics

// NewGoMetricsCollector is a Collector that yields Go runtime metrics.
// Metrics are prefixed with `go_`.
//
// Deprecated: Use the discrete collectors instead: [NewGoInfoCollector],
// [NewGoGCCollector], [NewGoMemoryCollector], [NewGoSchedCollector],
// [NewGoCPUCollector], and [NewGoMemstatsCollector].
func NewGoMetricsCollector() Collector {
	return &goMetricsCollector{
		info:     NewGoInfoCollector(),
		gc:       NewGoGCCollector(),
		memstats: NewGoMemstatsCollector(),
	}
}

type goMetricsCollector struct {
	info     Collector
	gc       Collector
	memstats Collector
}

func (c *goMetricsCollector) Collect(w ExpfmtWriter) {
	c.info.Collect(w)
	c.gc.Collect(w)
	c.memstats.Collect(w)
}
