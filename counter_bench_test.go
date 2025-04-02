package metrics

import "testing"

func BenchmarkUint64Vec(b *testing.B) {
	family := "http_request"
	labels := []string{"status"}
	value := "200"

	b.Run("hot", func(b *testing.B) {
		set := NewSet()
		v := set.NewUint64Vec(family, labels...)
		v.WithLabelValues(value)

		b.ReportAllocs()
		for b.Loop() {
			v.WithLabelValues("200")
		}
	})

	b.Run("cold", func(b *testing.B) {
		set := NewSet()
		v := set.NewUint64Vec(family, labels...)
		v.WithLabelValues(value)

		b.ReportAllocs()
		for b.Loop() {
			set.Reset()
			v.WithLabelValues(value)
		}
	})

	b.Run("verycold", func(b *testing.B) {
		set := NewSet()
		v := set.NewUint64Vec(family, labels...)

		b.ReportAllocs()
		for b.Loop() {
			set.Reset()
			identCache.Clear()
			v.WithLabelValues(value)
		}
	})
}

func BenchmarkCounterParallel(b *testing.B) {
	benchmarkCounterParallel(b, "Uint64.Add", NewSet().NewUint64, (*Uint64).Add, 1)
	benchmarkCounterParallel(b, "Uint64.Set", NewSet().NewUint64, (*Uint64).Set, 1)
	benchmarkCounterParallel(b, "Float64.Add", NewSet().NewFloat64, (*Float64).Add, 1)
	benchmarkCounterParallel(b, "Float64.Set", NewSet().NewFloat64, (*Float64).Set, 1)
	benchmarkCounterParallel(b, "Histogram.Update", NewSet().NewHistogram, (*Histogram).Update, 1)
}

func benchmarkCounterParallel[T any, V any](
	b *testing.B,
	name string,
	setup func(string, ...string) *T,
	do func(*T, V),
	value V,
) {
	b.Helper()
	thing := setup("foo")
	b.Run(name, func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				do(thing, value)
			}
		})
	})
}
