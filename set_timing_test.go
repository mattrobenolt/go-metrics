package metrics

import (
	"bytes"
	"math"
	"testing"
)

func BenchmarkWritePrometheus(b *testing.B) {
	set := NewSet()

	c1 := set.NewCounter(
		"http_request",
		"path", "/test",
		"status", "200",
	)
	c1.Inc()
	c1.Inc()

	c2 := set.NewCounter("counter_test", "bar", "baz")
	c2.Inc()

	c3 := set.NewCounter("other")
	c3.Inc()

	fc1 := set.NewFloatCounter("floatcounter_test", "foo", "bar")
	fc1.Set(1.5898)

	h1 := set.NewHistogram("hist_test", "foo", "bar")
	h1.Update(0.01)
	h1.Update(1.23)
	h1.Update(1.231)

	h2 := set.NewHistogram("hist_test2")
	for i := range 1000 {
		h2.Update(float64(i))
	}
	h2.Update(math.Inf(1))

	set.NewHistogram("hist_test3")

	var bb bytes.Buffer
	set.WritePrometheusUnthrottled(&bb)

	b.ReportAllocs()

	for b.Loop() {
		bb.Reset()
		set.WritePrometheusUnthrottled(&bb)
		b.SetBytes(int64(bb.Len()))
	}
}
