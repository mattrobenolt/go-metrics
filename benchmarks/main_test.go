package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand/v2"
	"strconv"
	"testing"

	vmmetrics "github.com/VictoriaMetrics/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/common/expfmt"
	"go.withmatt.com/metrics"
)

const (
	modMattware   = "mod=mattware"
	modVmmetrics  = "mod=vmmetrics"
	modPrometheus = "mod=prometheus"
)

func BenchmarkIncWithLabelValues(b *testing.B) {
	b.Run(modMattware, func(b *testing.B) {
		set := metrics.NewSet()
		c := set.NewCounterVec(metrics.VecOpt{
			Family: "foo",
			Labels: []string{"label1", "label2", "label3"},
		})

		b.ResetTimer()
		b.ReportAllocs()

		for b.Loop() {
			c.WithLabelValues("a", "200", "something").Inc()
		}
	})

	b.Run(modVmmetrics, func(b *testing.B) {
		set := vmmetrics.NewSet()

		b.ResetTimer()
		b.ReportAllocs()

		for b.Loop() {
			set.GetOrCreateCounter(
				fmt.Sprintf(`foo{label1="%s",label2="%s",label3="%s"}`, "a", "200", "something"),
			).Inc()
		}
	})

	b.Run(modPrometheus, func(b *testing.B) {
		registry := prometheus.NewRegistry()
		factory := promauto.With(registry)
		c := factory.NewCounterVec(prometheus.CounterOpts{
			Name: "foo",
		}, []string{"label1", "label2", "label3"})

		b.ResetTimer()
		b.ReportAllocs()

		for b.Loop() {
			c.WithLabelValues("a", "200", "something").Inc()
		}
	})
}

func BenchmarkUpdateVMRangeHistogram(b *testing.B) {
	const numObservations = 1000

	var seed [32]byte
	r := rand.NewChaCha8(seed)
	observations := make([]uint64, numObservations)
	for i := range numObservations {
		observations[i] = r.Uint64()
	}
	b.Run(modMattware, func(b *testing.B) {
		set := metrics.NewSet()
		h := set.NewHistogram("foo")

		b.ResetTimer()
		b.ReportAllocs()

		for i := range b.N {
			h.Update(float64(observations[i%len(observations)]))
		}
	})

	b.Run(modVmmetrics, func(b *testing.B) {
		set := vmmetrics.NewSet()
		h := set.NewHistogram("foo")

		b.ResetTimer()
		b.ReportAllocs()

		for i := range b.N {
			h.Update(float64(observations[i%len(observations)]))
		}
	})
}

func BenchmarkUpdatePromHistogram(b *testing.B) {
	const numObservations = 1000

	var seed [32]byte
	r := rand.NewChaCha8(seed)
	observations := make([]uint64, numObservations)
	for i := range numObservations {
		observations[i] = r.Uint64()
	}

	b.Run(modMattware, func(b *testing.B) {
		set := metrics.NewSet()
		h := set.NewFixedHistogram("foo", nil)

		b.ResetTimer()
		b.ReportAllocs()

		for i := range b.N {
			h.Observe(float64(observations[i%len(observations)]))
		}
	})

	b.Run(modPrometheus, func(b *testing.B) {
		registry := prometheus.NewRegistry()
		factory := promauto.With(registry)
		h := factory.NewHistogram(prometheus.HistogramOpts{
			Name: "foo",
		})

		b.ResetTimer()
		b.ReportAllocs()

		for i := range b.N {
			h.Observe(float64(observations[i%len(observations)]))
		}
	})
}

func BenchmarkWriteMetricsCounters(b *testing.B) {
	const numMetrics = 10000

	b.Run(modMattware, func(b *testing.B) {
		set := metrics.NewSet()
		c := set.NewCounterVec(metrics.VecOpt{
			Family: "foo",
			Labels: []string{"label1", "label2", "label3"},
		})

		for i := range numMetrics {
			c.WithLabelValues("a", strconv.Itoa(i), "something").Inc()
		}

		var bb bytes.Buffer
		set.WritePrometheusUnthrottled(&bb)

		b.ResetTimer()
		b.ReportAllocs()

		for b.Loop() {
			bb.Reset()
			set.WritePrometheusUnthrottled(&bb)
			b.SetBytes(int64(bb.Len()))
		}
	})

	b.Run(modVmmetrics, func(b *testing.B) {
		set := vmmetrics.NewSet()
		for i := range numMetrics {
			set.NewCounter(
				fmt.Sprintf(`foo{label1="%s",label2="%d",labels3="%s"}`, "a", i, "something"),
			).Inc()
		}

		var bb bytes.Buffer
		set.WritePrometheus(&bb)

		b.ResetTimer()
		b.ReportAllocs()

		for b.Loop() {
			bb.Reset()
			set.WritePrometheus(&bb)
			b.SetBytes(int64(bb.Len()))
		}
	})

	b.Run(modPrometheus, func(b *testing.B) {
		registry := prometheus.NewRegistry()
		factory := promauto.With(registry)
		c := factory.NewCounterVec(prometheus.CounterOpts{
			Name: "foo",
		}, []string{"label1", "label2", "label3"})

		for i := range numMetrics {
			c.WithLabelValues("a", strconv.Itoa(i), "something").Inc()
		}

		writePrometheus := func(w io.Writer) {
			mfs := must(registry.Gather())
			for _, mf := range mfs {
				must(expfmt.MetricFamilyToText(w, mf))
			}
		}

		var bb bytes.Buffer
		writePrometheus(&bb)

		b.ResetTimer()
		b.ReportAllocs()

		for b.Loop() {
			bb.Reset()
			writePrometheus(&bb)
			b.SetBytes(int64(bb.Len()))
		}
	})
}

func BenchmarkWriteMetricsVMRangeHistograms(b *testing.B) {
	const numHistograms = 100
	const numObservations = 100000

	var seed [32]byte
	r := rand.NewChaCha8(seed)
	observations := make([]uint64, numObservations)
	for i := range numObservations {
		observations[i] = r.Uint64()
	}

	b.Run(modMattware, func(b *testing.B) {
		set := metrics.NewSet()
		v := set.NewHistogramVec(metrics.VecOpt{
			Family: "foo",
			Labels: []string{"label1", "label2", "label3"},
		})
		for i := range numHistograms {
			h := v.WithLabelValues("a", strconv.Itoa(i), "something")
			for j := range numObservations {
				h.Update(float64(observations[j]))
			}
		}

		var bb bytes.Buffer
		set.WritePrometheusUnthrottled(&bb)

		b.ResetTimer()
		b.ReportAllocs()

		for b.Loop() {
			bb.Reset()
			set.WritePrometheusUnthrottled(&bb)
			b.SetBytes(int64(bb.Len()))
		}
	})

	b.Run(modVmmetrics, func(b *testing.B) {
		set := vmmetrics.NewSet()
		for i := range numHistograms {
			h := set.NewHistogram(
				fmt.Sprintf(`foo{label1="a",label2="%d",label3="something"}`, i),
			)
			for j := range numObservations {
				h.Update(float64(observations[j]))
			}
		}

		var bb bytes.Buffer
		set.WritePrometheus(&bb)

		b.ResetTimer()
		b.ReportAllocs()

		for b.Loop() {
			bb.Reset()
			set.WritePrometheus(&bb)
			b.SetBytes(int64(bb.Len()))
		}
	})
}

func BenchmarkWriteMetricsPromHistograms(b *testing.B) {
	const numHistograms = 100
	const numObservations = 100000

	var seed [32]byte
	r := rand.NewChaCha8(seed)
	observations := make([]uint64, numObservations)
	for i := range numObservations {
		observations[i] = r.Uint64()
	}

	b.Run(modMattware, func(b *testing.B) {
		set := metrics.NewSet()
		v := set.NewFixedHistogramVec(metrics.FixedHistogramVecOpt{
			Family: "foo",
			Labels: []string{"label1", "label2", "label3"},
		})
		for i := range numHistograms {
			h := v.WithLabelValues("a", strconv.Itoa(i), "something")
			for j := range numObservations {
				h.Update(float64(observations[j]))
			}
		}

		var bb bytes.Buffer
		set.WritePrometheusUnthrottled(&bb)

		b.ResetTimer()
		b.ReportAllocs()

		for b.Loop() {
			bb.Reset()
			set.WritePrometheusUnthrottled(&bb)
			b.SetBytes(int64(bb.Len()))
		}
	})

	b.Run(modPrometheus, func(b *testing.B) {
		registry := prometheus.NewRegistry()
		factory := promauto.With(registry)
		v := factory.NewHistogramVec(prometheus.HistogramOpts{
			Name: "foo",
		}, []string{"label1", "label2", "label3"})

		for i := range numHistograms {
			h := v.WithLabelValues("a", strconv.Itoa(i), "something")
			for j := range numObservations {
				h.Observe(float64(observations[j]))
			}
		}

		writePrometheus := func(w io.Writer) {
			mfs := must(registry.Gather())
			for _, mf := range mfs {
				must(expfmt.MetricFamilyToText(w, mf))
			}
		}

		var bb bytes.Buffer
		writePrometheus(&bb)

		b.ResetTimer()
		b.ReportAllocs()

		for b.Loop() {
			bb.Reset()
			writePrometheus(&bb)
			b.SetBytes(int64(bb.Len()))
		}
	})
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
