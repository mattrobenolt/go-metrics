package metrics

import (
	"io"
	"testing"
)

func BenchmarkWriteMetric(b *testing.B) {
	b.Run("uint64", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			WriteMetricUint64(io.Discard, "foobar", 1234)
		}
	})

	b.Run("int64", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			WriteMetricInt64(io.Discard, "foobar", -10)
		}
	})

	b.Run("float64", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			WriteMetricFloat64(io.Discard, "foobar", 1e6)
		}
	})
}
