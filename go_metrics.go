package metrics

import (
	"fmt"
	"regexp"
	"runtime"
	runtimemetrics "runtime/metrics"
	"slices"
	"strings"
	"time"
)

var (
	quantileTags = [...]Tag{
		MustTag("quantile", "0"),
		MustTag("quantile", "0.25"),
		MustTag("quantile", "0.5"),
		MustTag("quantile", "0.75"),
		MustTag("quantile", "1"),
	}
	goCollectorDefaultRuntimeMetrics = regexp.MustCompile(
		`/gc/gogc:percent|/gc/gomemlimit:bytes|/sched/gomaxprocs:threads`,
	)
)

// NewGoMetricsCollector is a Collector that yields Go runtime metrics.
// Metrics are prefixed with `go_`.
func NewGoMetricsCollector() Collector {
	var c goMetricsCollector
	c.init()
	return &c
}

type runtimeMetricName struct {
	sample string
	name   MetricName
}

type goMetricsCollector struct {
	runtimeMetrics []runtimeMetricName
}

func makeRuntimeMetricName(name string, r *strings.Replacer) runtimeMetricName {
	return runtimeMetricName{
		sample: name,
		name: MetricName{
			Family: MustIdent("go" + r.Replace(name)),
		},
	}
}

func (c *goMetricsCollector) init() {
	r := strings.NewReplacer("/", "_", ":", "_")
	for _, d := range runtimemetrics.All() {
		if goCollectorDefaultRuntimeMetrics.MatchString(d.Name) {
			c.runtimeMetrics = append(c.runtimeMetrics, makeRuntimeMetricName(d.Name, r))
		}
	}
}

func (c *goMetricsCollector) Collect(w ExpfmtWriter) {
	w.WriteLazyMetricUint64("go_info", 1,
		"version", runtime.Version(),
	)
	w.WriteLazyMetricUint64("go_info_ext", 1,
		"compiler", runtime.Compiler,
		"GOARCH", runtime.GOARCH,
		"GOOS", runtime.GOOS,
	)

	w.WriteLazyMetricFloat64("go_goroutines", float64(runtime.NumGoroutine()))

	numThreads, _ := runtime.ThreadCreateProfile(nil)
	w.WriteLazyMetricFloat64("go_threads", float64(numThreads))

	c.collectMemoryStats(w)
	c.collectRuntimeMetrics(w)
}

func (c *goMetricsCollector) collectMemoryStats(w ExpfmtWriter) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	c.collectGCStats(&ms, w)

	w.WriteLazyMetricUint64("go_memstats_alloc_bytes", ms.Alloc)
	w.WriteLazyMetricUint64("go_memstats_alloc_bytes_total", ms.TotalAlloc)
	w.WriteLazyMetricUint64("go_memstats_buck_hash_sys_bytes", ms.BuckHashSys)
	w.WriteLazyMetricUint64("go_memstats_frees_total", ms.Frees)
	w.WriteLazyMetricFloat64("go_memstats_gc_cpu_fraction", ms.GCCPUFraction)
	w.WriteLazyMetricUint64("go_memstats_gc_sys_bytes", ms.GCSys)
	w.WriteLazyMetricUint64("go_memstats_heap_alloc_bytes", ms.HeapAlloc)
	w.WriteLazyMetricUint64("go_memstats_heap_idle_bytes", ms.HeapIdle)
	w.WriteLazyMetricUint64("go_memstats_heap_inuse_bytes", ms.HeapInuse)
	w.WriteLazyMetricUint64("go_memstats_heap_objects", ms.HeapObjects)
	w.WriteLazyMetricUint64("go_memstats_heap_released_bytes", ms.HeapReleased)
	w.WriteLazyMetricUint64("go_memstats_heap_sys_bytes", ms.HeapSys)
	w.WriteLazyMetricDuration("go_memstats_last_gc_time_seconds", time.Duration(ms.LastGC))
	w.WriteLazyMetricUint64("go_memstats_lookups_total", ms.Lookups)
	w.WriteLazyMetricUint64("go_memstats_mallocs_total", ms.Mallocs)
	w.WriteLazyMetricUint64("go_memstats_mcache_inuse_bytes", ms.MCacheInuse)
	w.WriteLazyMetricUint64("go_memstats_mcache_sys_bytes", ms.MCacheSys)
	w.WriteLazyMetricUint64("go_memstats_mspan_inuse_bytes", ms.MSpanInuse)
	w.WriteLazyMetricUint64("go_memstats_mspan_sys_bytes", ms.MSpanSys)
	w.WriteLazyMetricUint64("go_memstats_next_gc_bytes", ms.NextGC)
	w.WriteLazyMetricUint64("go_memstats_other_sys_bytes", ms.OtherSys)
	w.WriteLazyMetricUint64("go_memstats_stack_inuse_bytes", ms.StackInuse)
	w.WriteLazyMetricUint64("go_memstats_stack_sys_bytes", ms.StackSys)
	w.WriteLazyMetricUint64("go_memstats_sys_bytes", ms.Sys)
}

func (c *goMetricsCollector) collectGCStats(ms *runtime.MemStats, w ExpfmtWriter) {
	var pauses []uint64
	if n := slices.Index(ms.PauseNs[:], 0); n == -1 {
		// the entire ring buffer is full
		pauses = ms.PauseNs[:]
	} else if n == 0 {
		// edge case of no GC pauses at all, we want a slice
		// of length 1 with 0 values
		pauses = ms.PauseNs[:1]
	} else {
		pauses = ms.PauseNs[:n]
	}
	slices.Sort(pauses)

	goGCDurationSeconds := MustIdent("go_gc_duration_seconds")
	const nq = len(quantileTags) - 1

	for i := range nq {
		w.WriteMetricDuration(MetricName{
			Family: goGCDurationSeconds,
			Tags:   []Tag{quantileTags[i]},
		}, time.Duration(pauses[len(pauses)*i/nq]))
	}
	w.WriteMetricDuration(MetricName{
		Family: goGCDurationSeconds,
		Tags:   []Tag{quantileTags[nq]},
	}, time.Duration(pauses[len(pauses)-1]))

	w.WriteLazyMetricUint64("go_gc_duration_seconds_count", uint64(ms.NumGC))
	w.WriteLazyMetricDuration("go_gc_duration_seconds_sum", time.Duration(ms.PauseTotalNs))
}

func (c *goMetricsCollector) collectRuntimeMetrics(w ExpfmtWriter) {
	samples := make([]runtimemetrics.Sample, len(c.runtimeMetrics))
	for i, rm := range c.runtimeMetrics {
		samples[i].Name = rm.sample
	}
	runtimemetrics.Read(samples)
	for i, rm := range c.runtimeMetrics {
		writeRuntimeMetric(w, rm.name, &samples[i])
	}
}

func writeRuntimeMetric(w ExpfmtWriter, name MetricName, sample *runtimemetrics.Sample) {
	kind := sample.Value.Kind()
	switch kind {
	case runtimemetrics.KindBad:
		panic(fmt.Sprintf("metrics: unexpected runtimemetrics.KindBad for sample.Name=%q", sample.Name))
	case runtimemetrics.KindUint64:
		v := sample.Value.Uint64()
		w.WriteMetricUint64(name, v)
	case runtimemetrics.KindFloat64:
		v := sample.Value.Float64()
		w.WriteMetricFloat64(name, v)
	case runtimemetrics.KindFloat64Histogram:
		// h := sample.Value.Float64Histogram()
		// writeRuntimeHistogramMetric(w, name, h)
	default:
		panic(fmt.Sprintf("metrics: unexpected metric kind=%d", kind))
	}
}
