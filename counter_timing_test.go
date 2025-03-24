package metrics

import "testing"

func BenchmarkCounterGetOrCreate(b *testing.B) {
	family := "http_request"
	tags := []string{"status", "200"}

	b.Run("hot", func(b *testing.B) {
		set := NewSet()
		set.GetOrCreateCounter(family, tags...)

		b.ReportAllocs()
		for b.Loop() {
			set.GetOrCreateCounter(family, tags...)
		}
	})

	b.Run("cold", func(b *testing.B) {
		set := NewSet()

		b.ReportAllocs()
		for b.Loop() {
			set.Reset()
			set.GetOrCreateCounter(family, tags...)
		}
	})

	b.Run("verycold", func(b *testing.B) {
		set := NewSet()

		b.ReportAllocs()
		for b.Loop() {
			set.Reset()
			identCache.Clear()
			set.GetOrCreateCounter(family, tags...)
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
