package metrics

import (
	"runtime"
	"slices"
)

var quantileTags = [...]Tag{
	MustTag("quantile", "0"),
	MustTag("quantile", "0.25"),
	MustTag("quantile", "0.5"),
	MustTag("quantile", "0.75"),
	MustTag("quantile", "1"),
}

func NewGoMetricsCollector() Collector {
	c := &goMetricsCollector{}
	c.init()
	return c
}

type goMetricsCollector struct {
	goInfo MetricName
}

func (c *goMetricsCollector) init() {
	c.goInfo = MetricName{
		Family: MustIdent("go_info"),
		Tags:   MustTags("version", runtime.Version()),
	}
}

func (c *goMetricsCollector) Collect(w ExpfmtWriter, constantTags string) {
	w.WriteMetricName(c.goInfo.With(constantTags))
	w.WriteUint64(1)

	collectMemoryStats(w, constantTags)
}

func collectMemoryStats(w ExpfmtWriter, constantTags string) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	collectGCStats(&ms, w, constantTags)
}

func collectGCStats(ms *runtime.MemStats, w ExpfmtWriter, constantTags string) {
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

	goGCDurationSecondsIdent := MustIdent("go_gc_duration_seconds")

	const nq = len(quantileTags) - 1
	for i := range nq {
		w.WriteMetricName(MetricName{
			Family:       goGCDurationSecondsIdent,
			Tags:         []Tag{quantileTags[i]},
			ConstantTags: constantTags,
		})
		w.WriteFloat64(float64(pauses[len(pauses)*i/nq]) / 1e9)
	}
	w.WriteMetricName(MetricName{
		Family:       goGCDurationSecondsIdent,
		Tags:         []Tag{quantileTags[nq]},
		ConstantTags: constantTags,
	})
	w.WriteFloat64(float64(pauses[len(pauses)-1]) / 1e9)

	w.WriteMetricName(MetricName{
		Family:       MustIdent("go_gc_duration_seconds_count"),
		ConstantTags: constantTags,
	})
	w.WriteUint64(uint64(ms.NumGC))
	w.WriteMetricName(MetricName{
		Family:       MustIdent("go_gc_duration_seconds_sum"),
		ConstantTags: constantTags,
	})
	w.WriteFloat64(float64(ms.PauseTotalNs) / 1e9)
}
