package metrics

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"runtime"
	"slices"
	"sync"
	"sync/atomic"

	"go.withmatt.com/metrics/internal/syncx"
)

const minimumWriteBuffer = 16 * 1024

var defaultSet Set

// ResetDefaultSet results the default global Set.
// See [Set.Reset].
func ResetDefaultSet() {
	defaultSet.Reset()
}

// RegisterDefaultCollectors registers the default Collectors
// onto the global Set.
func RegisterDefaultCollectors() {
	RegisterCollector(NewGoMetricsCollector())
	RegisterCollector(NewProcessMetricsCollector())
}

// RegisterCollector registers a Collector onto the global Set.
// See [Set.RegisterCollector].
func RegisterCollector(c Collector) {
	defaultSet.RegisterCollector(c)
}

// WritePrometheus writes the global Set to io.Writer.
// See [Set.WritePrometheus].
func WritePrometheus(w io.Writer) (int, error) {
	return defaultSet.WritePrometheus(w)
}

// Set is a collection of metrics. A single Set may have children Sets.
//
// [Set.WritePrometheus] must be called for exporting metrics from the set.
type Set struct {
	hasChildren atomic.Bool

	metrics syncx.SortedMap[metricHash, *namedMetric]

	childrenMu sync.Mutex
	children   []*Set
	collectors []Collector

	constantTags string
}

// NewSet creates new set of metrics.
func NewSet(constantTags ...string) *Set {
	var s Set
	s.Reset()

	if len(constantTags) > 0 {
		s.constantTags = materializeTags(MustTags(constantTags...))
	}

	return &s
}

// Reset resets the Set and retains allocated memory for reuse.
//
// Reset retains any ConstantTags if set.
func (s *Set) Reset() {
	s.metrics.Init(compareNamedMetrics)
	s.metrics.Clear()

	s.childrenMu.Lock()
	s.children = s.children[:0]
	s.collectors = s.collectors[:0]
	s.hasChildren.Store(false)
	s.childrenMu.Unlock()
}

// NewSet creates a new child Set in s.
func (s *Set) NewSet(constantTags ...string) *Set {
	s2 := NewSet()
	if len(constantTags) > 0 {
		s2.constantTags = joinTags(
			s.constantTags,
			MustTags(constantTags...),
		)
	} else {
		s2.constantTags = s.constantTags
	}
	s.childrenMu.Lock()
	s.children = append(s.children, s2)
	s.hasChildren.Store(true)
	s.childrenMu.Unlock()
	return s2
}

// UnregisterSet removes a previously registered child Set.
func (s *Set) UnregisterSet(set *Set) {
	s.childrenMu.Lock()
	if idx := slices.Index(s.children, set); idx >= 0 {
		s.children = slices.Delete(s.children, idx, idx+1)
	}
	s.hasChildren.Store(len(s.children) > 0 || len(s.collectors) > 0)
	s.childrenMu.Unlock()
}

// RegisterCollector registers a Collector.
// Registering the same collector more than once will panic.
func (s *Set) RegisterCollector(c Collector) {
	s.childrenMu.Lock()
	if idx := slices.Index(s.collectors, c); idx >= 0 {
		panic(errors.New("Collector already registered"))
	}
	s.collectors = append(s.collectors, c)
	s.hasChildren.Store(true)
	s.childrenMu.Unlock()
}

// UnregisterCollector removes a previously registered Collector from
// the Set.
func (s *Set) UnregisterCollector(c Collector) {
	s.childrenMu.Lock()
	if idx := slices.Index(s.collectors, c); idx >= 0 {
		s.collectors = slices.Delete(s.collectors, idx, idx+1)
	}
	s.hasChildren.Store(len(s.children) > 0 || len(s.collectors) > 0)
	s.childrenMu.Unlock()
}

// WritePrometheus writes the metrics along with all children to the io.Writer
// in Prometheus text exposition format.
//
// Metric writing and collecting is throttled by yielding the Go scheduler to
// not starve CPU. Use WritePrometheusUnthrottled if you don't want that.
func (s *Set) WritePrometheus(w io.Writer) (int, error) {
	return s.writePrometheus(w, true)
}

// WritePrometheusUnthrottled writes the metrics along with all children to the
// io.Writer in Prometheus text exposition format.
//
// This may starve the CPU and it's suggested to use [Set.WritePrometheus] instead.
func (s *Set) WritePrometheusUnthrottled(w io.Writer) (int, error) {
	return s.writePrometheus(w, false)
}

func (s *Set) writePrometheus(w io.Writer, throttle bool) (int, error) {
	// Optimize for the case where our io.Writer is already a bytes.Buffer,
	// but we always want to write into a Buffer first in case we have a slow
	// io.Writer.
	bb, isBuffer := w.(*bytes.Buffer)
	if !isBuffer {
		// if it's not, allocate a new one with a reasonable default
		bb = bytes.NewBuffer(make([]byte, 0, minimumWriteBuffer))
	} else {
		bb.Grow(minimumWriteBuffer)
	}

	exp := ExpfmtWriter{
		b:            bb,
		constantTags: s.constantTags,
	}

	for _, nm := range s.metrics.Values() {
		// yield the scheduler for each metric to not starve CPU
		if throttle {
			runtime.Gosched()
		}
		nm.metric.marshalTo(exp, nm.name)
	}

	if s.hasChildren.Load() {
		s.childrenMu.Lock()
		children := slices.Clone(s.children)
		s.childrenMu.Unlock()

		for _, s := range children {
			if _, err := s.writePrometheus(bb, throttle); err != nil {
				return 0, err
			}
		}

		s.childrenMu.Lock()
		collectors := slices.Clone(s.collectors)
		s.childrenMu.Unlock()

		for _, c := range collectors {
			// yield the scheduler for each Collector to not starve CPU
			if throttle {
				runtime.Gosched()
			}
			c.Collect(exp)
		}
	}

	if bb.Len() == 0 {
		return 0, nil
	}

	if !isBuffer {
		return w.Write(bb.Bytes())
	}

	return bb.Len(), nil
}

// mustRegisterMetric adds a new Metric, and will panic if the metric already has
// been registered.
func (s *Set) mustRegisterMetric(m Metric, name MetricName) {
	nm := &namedMetric{
		id:     getHashTags(name.Family.String(), name.Tags),
		name:   name,
		metric: m,
	}

	if _, loaded := s.metrics.LoadOrStore(nm.id, nm); loaded {
		panic(fmt.Errorf("metric %q is already registered", name.String()))
	}
}

// getOrRegisterMetricFromVec will attempt to create a new metric or return one that
// was potentially created in parallel from a Vec which is partially materialized.
// partialTags are tags with validated labels, but no values
func (s *Set) getOrRegisterMetricFromVec(m Metric, hash metricHash, family Ident, partialTags []Tag, values []string) *namedMetric {
	if len(values) != len(partialTags) {
		panic(errors.New("mismatch length of labels and values"))
	}
	// tags come in without values, so we need to stitch them together
	tags := slices.Clone(partialTags)
	for i := range tags {
		tags[i].value = MustValue(values[i])
	}
	return s.getOrAddNamedMetric(&namedMetric{
		id: hash,
		name: MetricName{
			Family: family,
			Tags:   tags,
		},
		metric: m,
	})
}

func (s *Set) getOrAddNamedMetric(newNm *namedMetric) *namedMetric {
	nm, _ := s.metrics.LoadOrStore(newNm.id, newNm)
	return nm
}

func joinTags(previous string, new []Tag) string {
	switch {
	case len(previous) == 0 && len(new) == 0:
		return ""
	case len(previous) == 0:
		return materializeTags(new)
	case len(new) == 0:
		return previous
	default:
		return previous + "," + materializeTags(new)
	}
}

// makePartialTags copy labels into partial tags. partial tags
// have a validated label, but no value.
func makePartialTags(labels []string) []Tag {
	partialTags := make([]Tag, len(labels))
	for i, label := range labels {
		partialTags[i].label = MustIdent(label)
	}
	return partialTags
}
