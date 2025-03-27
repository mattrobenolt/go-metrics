package metrics

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"maps"
	"runtime"
	"slices"
	"sync"
	"sync/atomic"
)

const minimumWriteBuffer = 16 * 1024

// SetOpt are the options for creating a Set.
type SetOpt struct {
	ConstantTags []Tag
}

// Set is a collection of metrics. A single Set may have children Sets.
//
// [Set.WritePrometheus] must be called for exporting metrics from the set.
type Set struct {
	dirty       atomic.Bool
	hasChildren atomic.Bool

	metricsMu sync.Mutex
	metrics   map[metricHash]*namedMetric

	valuesMu sync.Mutex
	values   []*namedMetric

	childrenMu sync.Mutex
	children   []*Set
	collectors []Collector

	constantTags string
}

// NewSet creates new set of metrics.
func NewSet() *Set {
	return NewSetOpt(SetOpt{})
}

// NewSetOpt creates a new Set with the opts.
func NewSetOpt(opt SetOpt) *Set {
	var s Set
	s.Reset()

	if len(opt.ConstantTags) > 0 {
		s.constantTags = materializeTags(opt.ConstantTags)
	}
	return &s
}

// Reset resets the Set and retains allocated memory for reuse.
//
// Reset retains any ConstantTags if set.
func (s *Set) Reset() {
	s.metricsMu.Lock()
	clear(s.metrics)
	s.values = s.values[:0]
	s.dirty.Store(false)
	s.metricsMu.Unlock()

	s.childrenMu.Lock()
	s.children = s.children[:0]
	s.collectors = s.collectors[:0]
	s.hasChildren.Store(false)
	s.childrenMu.Unlock()
}

// NewSet creates a new child Set in s.
func (s *Set) NewSet() *Set {
	return s.NewSetOpt(SetOpt{})
}

// NewSetOpt creates a new child Set with the opts in s.
func (s *Set) NewSetOpt(opt SetOpt) *Set {
	s2 := NewSet()
	s2.constantTags = joinTags(s.constantTags, opt.ConstantTags)
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

	if s.dirty.Load() {
		s.metricsMu.Lock()
		tmp := maps.Clone(s.metrics)
		// mark dirty early since we've taken a clone of our
		// metrics, which is the point we have split off.
		s.dirty.Store(false)

		// lock our values before unlocking the metrics to enforce
		// that we run sequentially in the same goroutine.
		s.valuesMu.Lock()
		s.metricsMu.Unlock()

		s.values = slices.Grow(s.values, len(tmp))
		s.values = s.values[:0]
		for _, v := range tmp {
			s.values = append(s.values, v)
		}
		slices.SortStableFunc(s.values, compareNamedMetrics)
		s.valuesMu.Unlock()
	}

	exp := ExpfmtWriter{
		b:            bb,
		constantTags: s.constantTags,
	}

	s.valuesMu.Lock()
	values := slices.Clone(s.values)
	s.valuesMu.Unlock()

	for _, nm := range values {
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

	s.metricsMu.Lock()
	defer s.metricsMu.Unlock()

	if _, ok := s.metrics[nm.id]; ok {
		panic(fmt.Errorf("metric %q is already registered", name.String()))
	}

	s.addMetricLocked(nm)
}

// getOrAddMetricFromStrings will attempt to create a new Metric or return one that
// was potentially created in parallel. Prefer registerMetric for speed.
func (s *Set) getOrAddMetricFromStrings(m Metric, hash metricHash, family string, tags []string) *namedMetric {
	return s.getOrAddNamedMetric(&namedMetric{
		id: hash,
		name: MetricName{
			Family: MustIdent(family),
			Tags:   MustTags(tags...),
		},
		metric: m,
	})
}

// getOrAddMetricFromVec will attempt to create a new metric or return one that
// was potentially created in parallel from a Vec which is partially materialized.
// partialTags are tags with validated labels, but no values
func (s *Set) getOrRegisterMetricFromVec(m Metric, hash metricHash, family Ident, partialTags []Tag, values []string) *namedMetric {
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
	s.metricsMu.Lock()
	defer s.metricsMu.Unlock()
	nm := s.metrics[newNm.id]
	if nm == nil {
		nm = newNm
		s.addMetricLocked(nm)
	}
	return nm
}

func (s *Set) addMetricLocked(nm *namedMetric) {
	if s.metrics == nil {
		s.metrics = map[metricHash]*namedMetric{
			nm.id: nm,
		}
	} else {
		s.metrics[nm.id] = nm
	}
	s.dirty.Store(true)
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
