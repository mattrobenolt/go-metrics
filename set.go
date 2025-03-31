package metrics

import (
	"bytes"
	"fmt"
	"io"
	"runtime"
	"slices"
	"sync"
	"sync/atomic"

	"go.withmatt.com/metrics/internal/syncx"
)

const minimumWriteBuffer = 16 * 1024

var defaultSet = newSet()

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
	// a Set gets assigned an id if it has constant tags, otherwise
	// there is no uniqueness in a Set.
	id metricHash

	metrics syncx.SortedMap[metricHash, *namedMetric]

	// setsByHash contains children Sets that have ids, therefore have constant
	// tags and are unique.
	setsByHash syncx.Map[metricHash, *Set]

	// unorderedSets are children Sets that do not have ids, therefore have
	// no constant tags.
	unorderedSets syncx.Set[*Set]

	hasCollectors atomic.Bool
	collectorsMu  sync.Mutex
	collectors    []Collector

	// constantTags are tags that are constant for all metrics in the set.
	// Children sets inherit these base tags.
	constantTags string
}

// NewSet creates new set of metrics.
func NewSet(constantTags ...string) *Set {
	s := newSet()
	s.setConstantTags("", constantTags...)
	return s
}

func newSet() *Set {
	s := &Set{}
	s.metrics.Init(compareNamedMetrics)
	return s
}

func (s *Set) setConstantTags(previousConstantTags string, constantTags ...string) {
	s.metrics.Init(compareNamedMetrics)
	s.constantTags = joinTags(previousConstantTags, MustTags(constantTags...)...)

	// give the Set an id if it has new constant tags
	if len(constantTags) > 0 {
		s.id = getHashStrings("", constantTags)
	}
}

// Reset resets the Set and retains allocated memory for reuse.
//
// Reset retains any ConstantTags if set.
func (s *Set) Reset() {
	s.metrics.Clear()
	s.setsByHash.Clear()
	s.unorderedSets.Clear()

	s.collectorsMu.Lock()
	s.collectors = s.collectors[:0]
	s.hasCollectors.Store(false)
	s.collectorsMu.Unlock()
}

// NewSet creates a new child Set in s.
// This will panic if constant tags are not unique within the parent Set. If
// no constant tags are provided, this will never fail.
func (s *Set) NewSet(constantTags ...string) *Set {
	s2 := newSet()
	s2.setConstantTags(s.constantTags, constantTags...)

	s.mustStoreSet(s2)
	return s2
}

// UnregisterSet removes a previously registered child Set.
func (s *Set) UnregisterSet(set *Set) {
	if set.id == emptyHash {
		s.unorderedSets.Delete(set)
	} else {
		s.setsByHash.Delete(set.id)
	}
}

// RegisterCollector registers a Collector.
// Registering the same collector more than once will panic.
func (s *Set) RegisterCollector(c Collector) {
	s.collectorsMu.Lock()
	if idx := slices.Index(s.collectors, c); idx >= 0 {
		panic("metrics: Collector already registered")
	}
	s.collectors = append(s.collectors, c)
	s.hasCollectors.Store(true)
	s.collectorsMu.Unlock()
}

// UnregisterCollector removes a previously registered Collector from
// the Set.
func (s *Set) UnregisterCollector(c Collector) {
	s.collectorsMu.Lock()
	if idx := slices.Index(s.collectors, c); idx >= 0 {
		s.collectors = slices.Delete(s.collectors, idx, idx+1)
	}
	s.hasCollectors.Store(len(s.collectors) > 0)
	s.collectorsMu.Unlock()
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

	if err := s.writeChildrenSets(bb, throttle); err != nil {
		return 0, err
	}

	if s.hasCollectors.Load() {
		s.collectorsMu.Lock()
		collectors := slices.Clone(s.collectors)
		s.collectorsMu.Unlock()

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

// writeChildrenSets writes all sets to the buffer.
func (s *Set) writeChildrenSets(bb *bytes.Buffer, throttle bool) error {
	var err error
	s.setsByHash.Range(func(_ metricHash, child *Set) bool {
		_, err = child.writePrometheus(bb, throttle)
		return err == nil
	})
	if err != nil {
		return err
	}
	s.unorderedSets.Range(func(child *Set) bool {
		_, err = child.writePrometheus(bb, throttle)
		return err == nil
	})
	return err
}

// mustStoreSet adds a new Set, and will panic if the set has already been registered.
func (s *Set) mustStoreSet(set *Set) {
	if set.id == emptyHash {
		if !s.unorderedSets.Add(set) {
			panic(fmt.Sprintf("metrics: set %v is already registered", set))
		}
	} else {
		if _, loaded := s.setsByHash.LoadOrStore(set.id, set); loaded {
			panic(fmt.Sprintf("metrics: set %q is already registered", set.constantTags))
		}
	}
}

// mustStoreMetric adds a new Metric, and will panic if the metric already has
// been registered.
func (s *Set) mustStoreMetric(m Metric, name MetricName) {
	nm := &namedMetric{
		id:     getHashTags(name.Family.String(), name.Tags),
		name:   name,
		metric: m,
	}

	if _, loaded := s.metrics.LoadOrStore(nm.id, nm); loaded {
		panic(fmt.Sprintf("metrics: metric %q is already registered", name.String()))
	}
}

// loadOrStoreMetricFromVec will attempt to create a new metric or return one that
// was potentially created in parallel from a Vec which is partially materialized.
// partialTags are tags with validated labels, but no values
func (s *Set) loadOrStoreMetricFromVec(m Metric, hash metricHash, family Ident, partialTags []Label, values []string) *namedMetric {
	if len(values) != len(partialTags) {
		panic("metrics: mismatch length of labels and values")
	}
	// tags come in without values, so we need to stitch them together
	tags := make([]Tag, len(partialTags))
	for i, label := range partialTags {
		tags[i] = Tag{
			label: label,
			value: MustValue(values[i]),
		}
	}
	return s.loadOrStoreNamedMetric(&namedMetric{
		id: hash,
		name: MetricName{
			Family: family,
			Tags:   tags,
		},
		metric: m,
	})
}

// loadOrStoreSetFromVec will attempt to create a new set or return one that
// was potentially created in parallel from a SetVec which is partially materialized.
func (s *Set) loadOrStoreSetFromVec(hash metricHash, label Label, value string) *Set {
	set := newSet()
	set.id = hash
	set.constantTags = joinTags(s.constantTags, Tag{
		label: label,
		value: MustValue(value),
	})
	return s.loadOrStoreSet(set)
}

func (s *Set) loadOrStoreSet(newSet *Set) *Set {
	set, _ := s.setsByHash.LoadOrStore(newSet.id, newSet)
	return set
}

func (s *Set) loadOrStoreNamedMetric(newNm *namedMetric) *namedMetric {
	nm, _ := s.metrics.LoadOrStore(newNm.id, newNm)
	return nm
}

func joinTags(previous string, new ...Tag) string {
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

// makeLabels converts a list of string labels into a list of Label objects.
func makeLabels(labels []string) []Label {
	new := make([]Label, len(labels))
	for i, label := range labels {
		new[i] = MustLabel(label)
	}
	return new
}
