//go:build darwin

package metrics

func (c *processMetricsCollector) Collect(w ExpfmtWriter) {
	collectUnix(w)
}
