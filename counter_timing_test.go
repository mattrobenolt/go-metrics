package metrics

import "testing"

func BenchmarkCounterVec(b *testing.B) {
	name := VecName{
		Family: "http_request",
		Labels: []string{"status"},
	}
	value := "200"

	b.Run("hot", func(b *testing.B) {
		set := NewSet()
		cv := set.NewCounterVec(name)
		cv.WithLabelValues(value)

		b.ReportAllocs()
		for b.Loop() {
			cv.WithLabelValues("200")
		}
	})

	b.Run("cold", func(b *testing.B) {
		set := NewSet()
		v := set.NewCounterVec(name)
		v.WithLabelValues(value)

		b.ReportAllocs()
		for b.Loop() {
			set.Reset()
			v.WithLabelValues(value)
		}
	})

	b.Run("verycold", func(b *testing.B) {
		set := NewSet()
		v := set.NewCounterVec(name)

		b.ReportAllocs()
		for b.Loop() {
			set.Reset()
			identCache.Clear()
			v.WithLabelValues(value)
		}
	})
}

func BenchmarkCounterInc(b *testing.B) {
	c := NewSet().NewCounter("foo")

	b.ReportAllocs()
	for b.Loop() {
		c.Inc()
	}
}

func BenchmarkCounterIncParallel(b *testing.B) {
	c := NewSet().NewCounter("foo")

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Inc()
		}
	})
}
