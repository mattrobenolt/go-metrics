package metrics

import (
	"fmt"
	"math"
	"regexp"
	"runtime"
	runtimemetrics "runtime/metrics"
	"strconv"
	"strings"
)

// runtimeMetricName maps a runtime/metrics sample name to its metric name.
type runtimeMetricName struct {
	sample string
	name   MetricName
}

var runtimeMetricReplacer = strings.NewReplacer("/", "_", ":", "_", "-", "_")

func makeRuntimeMetricNames(prefix string, filter *regexp.Regexp) []runtimeMetricName {
	var names []runtimeMetricName
	for _, d := range runtimemetrics.All() {
		if !strings.HasPrefix(d.Name, prefix) {
			continue
		}
		if filter != nil && !filter.MatchString(d.Name) {
			continue
		}
		names = append(names, runtimeMetricName{
			sample: d.Name,
			name: MetricName{
				Family: MustIdent("go" + runtimeMetricReplacer.Replace(d.Name)),
			},
		})
	}
	return names
}

func collectRuntimeMetrics(w ExpfmtWriter, names []runtimeMetricName) {
	samples := make([]runtimemetrics.Sample, len(names))
	for i, rm := range names {
		samples[i].Name = rm.sample
	}
	runtimemetrics.Read(samples)
	for i, rm := range names {
		writeRuntimeMetric(w, rm.name, &samples[i])
	}
}

func writeRuntimeMetric(w ExpfmtWriter, name MetricName, sample *runtimemetrics.Sample) {
	kind := sample.Value.Kind()
	switch kind {
	case runtimemetrics.KindBad:
		panic(fmt.Sprintf("metrics: unexpected runtimemetrics.KindBad for sample.Name=%q", sample.Name))
	case runtimemetrics.KindUint64:
		w.WriteMetricUint64(name, sample.Value.Uint64())
	case runtimemetrics.KindFloat64:
		w.WriteMetricFloat64(name, sample.Value.Float64())
	case runtimemetrics.KindFloat64Histogram:
		writeRuntimeHistogramMetric(w, name, sample.Value.Float64Histogram())
	default:
		panic(fmt.Sprintf("metrics: unexpected metric kind=%d", kind))
	}
}

func writeRuntimeHistogramMetric(w ExpfmtWriter, name MetricName, h *runtimemetrics.Float64Histogram) {
	buckets := h.Buckets
	counts := h.Counts
	family := name.Family.String()
	b := w.b

	tagsSize := sizeOfTags(name.Tags, w.constantTags) + 1

	const (
		chunkVMRange = `_bucket{vmrange="`
		chunkSum     = "_sum"
		chunkCount   = "_count"
	)

	var totalCount uint64
	var sum float64
	var nonZero int

	for i, count := range counts {
		totalCount += count
		if count > 0 {
			nonZero++
			// Estimate sum using lower bound of each bucket (underestimate,
			// same approach as prometheus/client_golang).
			if !math.IsInf(buckets[i], -1) {
				sum += buckets[i] * float64(count)
			}
		}
	}

	if totalCount == 0 {
		return
	}

	b.Grow(
		(len(family)+len(chunkVMRange)+tagsSize+40)*nonZero +
			len(family) + len(chunkSum) + tagsSize + 20 +
			len(family) + len(chunkCount) + tagsSize + 20 +
			64,
	)

	for i, count := range counts {
		if count == 0 {
			continue
		}

		lower := strconv.FormatFloat(buckets[i], 'g', -1, 64)
		upper := strconv.FormatFloat(buckets[i+1], 'g', -1, 64)

		b.WriteString(family)
		b.WriteString(chunkVMRange)
		b.WriteString(lower)
		b.WriteString("...")
		b.WriteString(upper)
		b.WriteByte('"')
		if len(w.constantTags) > 0 {
			b.WriteByte(',')
			b.WriteString(w.constantTags)
		}
		for _, tag := range name.Tags {
			b.WriteByte(',')
			writeTag(b, tag)
		}
		b.WriteString(`} `)
		writeUint64(b, count)
		b.WriteByte('\n')
	}

	// _sum
	b.WriteString(family)
	b.WriteString(chunkSum)
	if tagsSize > 0 {
		b.WriteByte('{')
		writeTags(b, w.constantTags, name.Tags)
		b.WriteByte('}')
	}
	b.WriteByte(' ')
	writeFloat64(b, sum)
	b.WriteByte('\n')

	// _count
	b.WriteString(family)
	b.WriteString(chunkCount)
	if tagsSize > 0 {
		b.WriteByte('{')
		writeTags(b, w.constantTags, name.Tags)
		b.WriteByte('}')
	}
	b.WriteByte(' ')
	writeUint64(b, totalCount)
	b.WriteByte('\n')
}

// GoInfoCollector emits go_info, go_info_ext, go_goroutines, and go_threads.
type goInfoCollector struct{}

// NewGoInfoCollector returns a Collector that emits basic Go runtime
// information: go_info, go_info_ext, go_goroutines, and go_threads.
func NewGoInfoCollector() Collector {
	return &goInfoCollector{}
}

func (c *goInfoCollector) Collect(w ExpfmtWriter) {
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
}

// GoGCCollectorOption configures a [GoGCCollector].
type GoGCCollectorOption func(*goGCCollector)

// WithGCMetricFilter sets a regexp filter for which /gc/* metrics to collect.
func WithGCMetricFilter(re *regexp.Regexp) GoGCCollectorOption {
	return func(c *goGCCollector) { c.filter = re }
}

type goGCCollector struct {
	filter         *regexp.Regexp
	runtimeMetrics []runtimeMetricName
}

// NewGoGCCollector returns a Collector that emits /gc/* metrics from
// the [runtime/metrics] package.
func NewGoGCCollector(opts ...GoGCCollectorOption) Collector {
	c := &goGCCollector{}
	for _, opt := range opts {
		opt(c)
	}
	c.runtimeMetrics = makeRuntimeMetricNames("/gc/", c.filter)
	return c
}

func (c *goGCCollector) Collect(w ExpfmtWriter) {
	collectRuntimeMetrics(w, c.runtimeMetrics)
}

// GoMemoryCollectorOption configures a [GoMemoryCollector].
type GoMemoryCollectorOption func(*goMemoryCollector)

// WithMemoryMetricFilter sets a regexp filter for which /memory/* metrics to collect.
func WithMemoryMetricFilter(re *regexp.Regexp) GoMemoryCollectorOption {
	return func(c *goMemoryCollector) { c.filter = re }
}

type goMemoryCollector struct {
	filter         *regexp.Regexp
	runtimeMetrics []runtimeMetricName
}

// NewGoMemoryCollector returns a Collector that emits /memory/* metrics from
// the [runtime/metrics] package.
func NewGoMemoryCollector(opts ...GoMemoryCollectorOption) Collector {
	c := &goMemoryCollector{}
	for _, opt := range opts {
		opt(c)
	}
	c.runtimeMetrics = makeRuntimeMetricNames("/memory/", c.filter)
	return c
}

func (c *goMemoryCollector) Collect(w ExpfmtWriter) {
	collectRuntimeMetrics(w, c.runtimeMetrics)
}

// GoSchedCollectorOption configures a [GoSchedCollector].
type GoSchedCollectorOption func(*goSchedCollector)

// WithSchedMetricFilter sets a regexp filter for which /sched/* metrics to collect.
func WithSchedMetricFilter(re *regexp.Regexp) GoSchedCollectorOption {
	return func(c *goSchedCollector) { c.filter = re }
}

type goSchedCollector struct {
	filter         *regexp.Regexp
	runtimeMetrics []runtimeMetricName
}

// NewGoSchedCollector returns a Collector that emits /sched/* metrics from
// the [runtime/metrics] package.
func NewGoSchedCollector(opts ...GoSchedCollectorOption) Collector {
	c := &goSchedCollector{}
	for _, opt := range opts {
		opt(c)
	}
	c.runtimeMetrics = makeRuntimeMetricNames("/sched/", c.filter)
	return c
}

func (c *goSchedCollector) Collect(w ExpfmtWriter) {
	collectRuntimeMetrics(w, c.runtimeMetrics)
}

// GoCPUCollectorOption configures a [GoCPUCollector].
type GoCPUCollectorOption func(*goCPUCollector)

// WithCPUMetricFilter sets a regexp filter for which /cpu/* metrics to collect.
func WithCPUMetricFilter(re *regexp.Regexp) GoCPUCollectorOption {
	return func(c *goCPUCollector) { c.filter = re }
}

type goCPUCollector struct {
	filter         *regexp.Regexp
	runtimeMetrics []runtimeMetricName
}

// NewGoCPUCollector returns a Collector that emits /cpu/* metrics from
// the [runtime/metrics] package.
func NewGoCPUCollector(opts ...GoCPUCollectorOption) Collector {
	c := &goCPUCollector{}
	for _, opt := range opts {
		opt(c)
	}
	c.runtimeMetrics = makeRuntimeMetricNames("/cpu/", c.filter)
	return c
}

func (c *goCPUCollector) Collect(w ExpfmtWriter) {
	collectRuntimeMetrics(w, c.runtimeMetrics)
}

// GoCgoCollectorOption configures a [GoCgoCollector].
type GoCgoCollectorOption func(*goCgoCollector)

// WithCgoMetricFilter sets a regexp filter for which /cgo/* metrics to collect.
func WithCgoMetricFilter(re *regexp.Regexp) GoCgoCollectorOption {
	return func(c *goCgoCollector) { c.filter = re }
}

type goCgoCollector struct {
	filter         *regexp.Regexp
	runtimeMetrics []runtimeMetricName
}

// NewGoCgoCollector returns a Collector that emits /cgo/* metrics from
// the [runtime/metrics] package.
func NewGoCgoCollector(opts ...GoCgoCollectorOption) Collector {
	c := &goCgoCollector{}
	for _, opt := range opts {
		opt(c)
	}
	c.runtimeMetrics = makeRuntimeMetricNames("/cgo/", c.filter)
	return c
}

func (c *goCgoCollector) Collect(w ExpfmtWriter) {
	collectRuntimeMetrics(w, c.runtimeMetrics)
}

// GoSyncCollectorOption configures a [GoSyncCollector].
type GoSyncCollectorOption func(*goSyncCollector)

// WithSyncMetricFilter sets a regexp filter for which /sync/* metrics to collect.
func WithSyncMetricFilter(re *regexp.Regexp) GoSyncCollectorOption {
	return func(c *goSyncCollector) { c.filter = re }
}

type goSyncCollector struct {
	filter         *regexp.Regexp
	runtimeMetrics []runtimeMetricName
}

// NewGoSyncCollector returns a Collector that emits /sync/* metrics from
// the [runtime/metrics] package.
func NewGoSyncCollector(opts ...GoSyncCollectorOption) Collector {
	c := &goSyncCollector{}
	for _, opt := range opts {
		opt(c)
	}
	c.runtimeMetrics = makeRuntimeMetricNames("/sync/", c.filter)
	return c
}

func (c *goSyncCollector) Collect(w ExpfmtWriter) {
	collectRuntimeMetrics(w, c.runtimeMetrics)
}

// GoGodebugCollectorOption configures a [GoGodebugCollector].
type GoGodebugCollectorOption func(*goGodebugCollector)

// WithGodebugMetricFilter sets a regexp filter for which /godebug/* metrics to collect.
func WithGodebugMetricFilter(re *regexp.Regexp) GoGodebugCollectorOption {
	return func(c *goGodebugCollector) { c.filter = re }
}

type goGodebugCollector struct {
	filter         *regexp.Regexp
	runtimeMetrics []runtimeMetricName
}

// NewGoGodebugCollector returns a Collector that emits /godebug/* metrics from
// the [runtime/metrics] package.
func NewGoGodebugCollector(opts ...GoGodebugCollectorOption) Collector {
	c := &goGodebugCollector{}
	for _, opt := range opts {
		opt(c)
	}
	c.runtimeMetrics = makeRuntimeMetricNames("/godebug/", c.filter)
	return c
}

func (c *goGodebugCollector) Collect(w ExpfmtWriter) {
	collectRuntimeMetrics(w, c.runtimeMetrics)
}
