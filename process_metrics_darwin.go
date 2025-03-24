//go:build darwin

package metrics

func (c *processMetricsCollector) Collect(w ExpfmtWriter, constantTags string) {
	collectUnix(w, constantTags)
}
