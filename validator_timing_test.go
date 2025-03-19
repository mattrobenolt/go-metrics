package metrics

import "testing"

func BenchmarkValidateMetric(b *testing.B) {
	b.Run("no labels", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			validateMetric(`go_memstats_mspan_inuse_bytes`)
		}
	})
	b.Run("2 labels", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			validateMetric(`go_memstats_mspan_inuse_bytes{foo="bar",baz="other"}`)
		}
	})
}
