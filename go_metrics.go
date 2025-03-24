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
	goCollectorDefaultRuntimeMetrics = regexp.MustCompile(`/gc/gogc:percent|/gc/gomemlimit:bytes|/sched/gomaxprocs:threads`)
)

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

func (c *goMetricsCollector) Collect(w ExpfmtWriter, constantTags string) {
	runtime.GC()
	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_info"),
		Tags:         MustTags("version", runtime.Version()),
		ConstantTags: constantTags,
	}, 1)
	w.WriteMetricUint64(MetricName{
		Family: MustIdent("go_info_ext"),
		Tags: MustTags(
			"compiler", runtime.Compiler,
			"GOARCH", runtime.GOARCH,
			"GOOS", runtime.GOOS,
		),
		ConstantTags: constantTags,
	}, 1)

	w.WriteMetricFloat64(MetricName{
		Family:       MustIdent("go_goroutines"),
		ConstantTags: constantTags,
	}, float64(runtime.NumGoroutine()))

	numThreads, _ := runtime.ThreadCreateProfile(nil)
	w.WriteMetricFloat64(MetricName{
		Family:       MustIdent("go_threads"),
		ConstantTags: constantTags,
	}, float64(numThreads))

	c.collectMemoryStats(w, constantTags)
	c.collectRuntimeMetrics(w, constantTags)
}

func (c *goMetricsCollector) collectMemoryStats(w ExpfmtWriter, constantTags string) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	c.collectGCStats(&ms, w, constantTags)

	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_memstats_alloc_bytes"),
		ConstantTags: constantTags,
	}, ms.Alloc)
	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_memstats_alloc_bytes_total"),
		ConstantTags: constantTags,
	}, ms.TotalAlloc)
	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_memstats_buck_hash_sys_bytes"),
		ConstantTags: constantTags,
	}, ms.BuckHashSys)
	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_memstats_frees_total"),
		ConstantTags: constantTags,
	}, ms.Frees)
	w.WriteMetricFloat64(MetricName{
		Family:       MustIdent("go_memstats_gc_cpu_fraction"),
		ConstantTags: constantTags,
	}, ms.GCCPUFraction)
	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_memstats_gc_sys_bytes"),
		ConstantTags: constantTags,
	}, ms.GCSys)
	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_memstats_heap_alloc_bytes"),
		ConstantTags: constantTags,
	}, ms.HeapAlloc)
	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_memstats_heap_idle_bytes"),
		ConstantTags: constantTags,
	}, ms.HeapIdle)
	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_memstats_heap_inuse_bytes"),
		ConstantTags: constantTags,
	}, ms.HeapInuse)
	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_memstats_heap_objects"),
		ConstantTags: constantTags,
	}, ms.HeapObjects)
	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_memstats_heap_released_bytes"),
		ConstantTags: constantTags,
	}, ms.HeapReleased)
	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_memstats_heap_sys_bytes"),
		ConstantTags: constantTags,
	}, ms.HeapSys)
	w.WriteMetricDuration(MetricName{
		Family:       MustIdent("go_memstats_last_gc_time_seconds"),
		ConstantTags: constantTags,
	}, time.Duration(ms.LastGC))
	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_memstats_lookups_total"),
		ConstantTags: constantTags,
	}, ms.Lookups)
	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_memstats_mallocs_total"),
		ConstantTags: constantTags,
	}, ms.Mallocs)
	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_memstats_mcache_inuse_bytes"),
		ConstantTags: constantTags,
	}, ms.MCacheInuse)
	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_memstats_mcache_sys_bytes"),
		ConstantTags: constantTags,
	}, ms.MCacheSys)
	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_memstats_mspan_inuse_bytes"),
		ConstantTags: constantTags,
	}, ms.MSpanInuse)
	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_memstats_mspan_sys_bytes"),
		ConstantTags: constantTags,
	}, ms.MSpanSys)
	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_memstats_next_gc_bytes"),
		ConstantTags: constantTags,
	}, ms.NextGC)
	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_memstats_other_sys_bytes"),
		ConstantTags: constantTags,
	}, ms.OtherSys)
	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_memstats_stack_inuse_bytes"),
		ConstantTags: constantTags,
	}, ms.StackInuse)
	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_memstats_stack_sys_bytes"),
		ConstantTags: constantTags,
	}, ms.StackSys)
	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_memstats_sys_bytes"),
		ConstantTags: constantTags,
	}, ms.Sys)
}

func (c *goMetricsCollector) collectGCStats(ms *runtime.MemStats, w ExpfmtWriter, constantTags string) {
	var pauses []uint64
	if n := slices.Index(ms.PauseNs[:], 0); n == -1 {
		// the entire ring buffer is full
		pauses = ms.PauseNs[:]
	} else if n == 0 {
		// edge case of no GC pausese at all, we want a slice
		// of length 1 with 0 values
		pauses = ms.PauseNs[:1]
	} else {
		pauses = ms.PauseNs[:n]
	}
	slices.Sort(pauses)

	const nq = len(quantileTags) - 1
	for i := range nq {
		w.WriteMetricDuration(MetricName{
			Family:       MustIdent("go_gc_duration_seconds"),
			Tags:         []Tag{quantileTags[i]},
			ConstantTags: constantTags,
		}, time.Duration(pauses[len(pauses)*i/nq]))
	}
	w.WriteMetricDuration(MetricName{
		Family:       MustIdent("go_gc_duration_seconds"),
		Tags:         []Tag{quantileTags[nq]},
		ConstantTags: constantTags,
	}, time.Duration(pauses[len(pauses)-1]))

	w.WriteMetricUint64(MetricName{
		Family:       MustIdent("go_gc_duration_seconds_count"),
		ConstantTags: constantTags,
	}, uint64(ms.NumGC))
	w.WriteMetricDuration(MetricName{
		Family:       MustIdent("go_gc_duration_seconds_sum"),
		ConstantTags: constantTags,
	}, time.Duration(ms.PauseTotalNs))
}

func (c *goMetricsCollector) collectRuntimeMetrics(w ExpfmtWriter, constantTags string) {
	samples := make([]runtimemetrics.Sample, len(c.runtimeMetrics))
	for i, rm := range c.runtimeMetrics {
		samples[i].Name = rm.sample
	}
	runtimemetrics.Read(samples)
	for i, rm := range c.runtimeMetrics {
		writeRuntimeMetric(w, rm.name.With(constantTags), &samples[i])
	}
}

func writeRuntimeMetric(w ExpfmtWriter, name MetricName, sample *runtimemetrics.Sample) {
	kind := sample.Value.Kind()
	switch kind {
	case runtimemetrics.KindBad:
		panic(fmt.Errorf("BUG: unexpected runtimemetrics.KindBad for sample.Name=%q", sample.Name))
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
		panic(fmt.Errorf("unexpected metric kind=%d", kind))
	}
}
