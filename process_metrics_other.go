//go:build !darwin && !linux

package metrics

func (c *processMetricsCollector) Collect(w ExpfmtWriter, constantTags string) {}
