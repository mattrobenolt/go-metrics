package metrics

import "testing"

func BenchmarkHistogramUpdate(b *testing.B) {
	h := NewSet().NewHistogram("foo")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; b.Loop(); i++ {
		h.Update(float64(i))
	}
}

func BenchmarkHistogramUpdateParallel(b *testing.B) {
	h := NewSet().NewHistogram("foo")

	var i float64
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// XXX: this is racy, but it doesn't matter,
			// I don't want to add synchronization with an
			// atomic to ruin the benchmark.
			i += 1
			h.Update(i)
		}
	})
}
