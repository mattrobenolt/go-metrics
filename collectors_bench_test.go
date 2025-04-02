package metrics

import "testing"

func BenchmarkCollectors(b *testing.B) {
	benchmarkCollector(b, "Go", NewGoMetricsCollector())
	benchmarkCollector(b, "Process", NewProcessMetricsCollector())
	benchmarkCollector(b, "Self", NewSelfMetricsCollector())
}

func benchmarkCollector(b *testing.B, name string, c Collector) {
	b.Helper()
	b.Run(name, func(b *testing.B) {
		w := NewTestingExpfmtWriter()
		b.ReportAllocs()
		for b.Loop() {
			w.b.Reset()
			c.Collect(w)
		}
	})
}
