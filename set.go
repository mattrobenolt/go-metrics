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
	"time"

	"go.withmatt.com/metrics/internal/atomicx"
	"go.withmatt.com/metrics/internal/fasttime"
	"go.withmatt.com/metrics/internal/syncx"
)

const minimumWriteBuffer = 16 * 1024

var defaultSet = newSet()

// fastClock is a clock used for TTL'ing Sets. The accuracy is
// 1 tick per second, so it's not possible to get accuracy better than 1 second,
// but this keeps the overhead very cheap.
var fastClock = sync.OnceValue(func() *fasttime.Clock {
	return fasttime.NewClock(time.Second)
})

// ErrSetExpired is returned from WritePrometheus when a Set has expired.
var ErrSetExpired = errors.New("set expired")

// IsActiveFunc determines if a [Set] is considered "active" and should be kept alive.
//
// Return true to keep the [Set] alive (prevents expiration even if the TTL has passed).
// Return false to allow normal TTL-based expiration.
//
// See [SetVec.SetIsActive].
type IsActiveFunc func(s *Set) bool

// ResetDefaultSet results the default global Set.
// See [Set.Reset].
func ResetDefaultSet() {
	defaultSet.Reset()
}

// RegisterDefaultCollectors registers the default Collectors
// onto the global Set.
func RegisterDefaultCollectors() {
	RegisterCollector(
		NewGoMetricsCollector(),
		NewProcessMetricsCollector(),
	)
}

// RegisterCollector registers one or more Collectors onto the global Set.
// See [Set.RegisterCollector].
func RegisterCollector(cs ...Collector) {
	defaultSet.RegisterCollector(cs...)
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

	collectors atomic.Pointer[[]Collector]

	// constantTags are tags that are constant for all metrics in the set.
	// Children sets inherit these base tags.
	constantTags string

	ttl      time.Duration
	lastUsed atomicx.Instant

	// isActive is an optional callback to determine if this Set should be kept alive.
	// If set, it will be called during expiration checks.
	isActive IsActiveFunc
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

// AppendConstantTags appends constant tags to the Set.
// The new tags will descend into any children sets.
// This is not thread-safe and should only be done as a part of initial
// metrics setup.
//
// This can also panic if any new child sets with the new tags become non-unique.
func (s *Set) AppendConstantTags(constantTags ...string) {
	if len(constantTags) == 0 {
		return
	}

	s.setConstantTags(s.constantTags, constantTags...)

	// We also need to descend into all children sets and append the new tags
	// there as well.
	s.rangeChildrenSets(func(child *Set) bool {
		if child.id == emptyHash {
			s.unorderedSets.Delete(child)
		} else {
			s.setsByHash.Delete(child.id)
		}

		child.AppendConstantTags(constantTags...)
		s.mustStoreSet(child)
		return true
	})
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
	defer s.KeepAlive()

	s.metrics.Clear()
	s.setsByHash.Clear()
	s.unorderedSets.Clear()
	s.collectors.Store(nil)
}

// NewSet creates a new child Set in s.
// This will panic if constant tags are not unique within the parent Set. If
// no constant tags are provided, this will never fail.
func (s *Set) NewSet(constantTags ...string) *Set {
	defer s.KeepAlive()

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

// RegisterCollector registers one or more Collectors.
// Registering the same collector more than once will panic.
func (s *Set) RegisterCollector(cs ...Collector) {
	var newValues []Collector
	var oldValues *[]Collector

	for {
		oldValues = s.collectors.Load()
		if oldValues != nil {
			for _, c := range cs {
				if slices.Contains(*oldValues, c) {
					panic("metrics: Collector already registered")
				}
			}
			newValues = slices.Clone(*oldValues)
		} else {
			newValues = nil
		}
		newValues = append(newValues, cs...)
		if s.collectors.CompareAndSwap(oldValues, &newValues) {
			return
		}
	}
}

// UnregisterCollector removes a previously registered Collector from
// the Set.
func (s *Set) UnregisterCollector(c Collector) {
	var newValues []Collector
	var oldValues *[]Collector

	for {
		oldValues = s.collectors.Load()
		if oldValues == nil {
			return
		}
		idx := slices.Index(*oldValues, c)
		if idx == -1 {
			return
		}

		newValues = slices.Clone(*oldValues)
		newValues = slices.Delete(newValues, idx, idx+1)
		newValues = slices.Clip(newValues)
		if s.collectors.CompareAndSwap(oldValues, &newValues) {
			return
		}
	}
}

// WritePrometheus writes the metrics along with all children to the io.Writer
// in Prometheus text exposition format.
//
// Metric writing and collecting is throttled by yielding the Go scheduler to
// not starve CPU. Use WritePrometheusUnthrottled if you don't want that.
func (s *Set) WritePrometheus(w io.Writer) (int, error) {
	if s.isExpired() {
		return 0, ErrSetExpired
	}
	return s.writePrometheus(w, true)
}

// WritePrometheusUnthrottled writes the metrics along with all children to the
// io.Writer in Prometheus text exposition format.
//
// This may starve the CPU and it's suggested to use [Set.WritePrometheus] instead.
func (s *Set) WritePrometheusUnthrottled(w io.Writer) (int, error) {
	if s.isExpired() {
		return 0, ErrSetExpired
	}
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

	s.collectInternal(exp, throttle)

	if bb.Len() == 0 {
		return 0, nil
	}

	if !isBuffer {
		return w.Write(bb.Bytes())
	}

	return bb.Len(), nil
}

// rangeChildrenSets iterates over all child sets with a single
// callback function. rangeChildrenSets also maintains expiration
// and deletes expired sets if applicable.
func (s *Set) rangeChildrenSets(f func(s *Set) bool) {
	keepGoing := true
	s.setsByHash.Range(func(key metricHash, child *Set) bool {
		if child.isExpired() {
			s.setsByHash.Delete(key)
			return true
		}
		keepGoing = f(child)
		return keepGoing
	})
	if !keepGoing {
		return
	}
	s.unorderedSets.Range(func(child *Set) bool {
		if child.isExpired() {
			s.unorderedSets.Delete(child)
			return true
		}
		return f(child)
	})
}

// collectInternal is the unified collection logic used by both Collect and WritePrometheus.
// It writes metrics, child sets, and collectors to the provided ExpfmtWriter.
// If throttle is true, it yields the scheduler periodically to avoid CPU starvation.
func (s *Set) collectInternal(w ExpfmtWriter, throttle bool) {
	// Write all metrics in this set
	for _, nm := range s.metrics.Values() {
		// yield the scheduler for each metric to not starve CPU
		if throttle {
			runtime.Gosched()
		}
		nm.metric.marshalTo(w, nm.name)
	}

	// Write all children sets recursively
	s.collectChildrenSets(w, throttle)

	// Collect from any registered collectors
	if collectors := s.collectors.Load(); collectors != nil {
		for _, c := range *collectors {
			// yield the scheduler for each Collector to not starve CPU
			if throttle {
				runtime.Gosched()
			}
			c.Collect(w)
		}
	}
}

func (s *Set) isExpired() bool {
	if s.ttl == 0 {
		return false
	}

	// Check if user considers this Set "active"
	if s.isActive != nil && s.isActive(s) {
		s.KeepAlive() // Bump lastUsed since it's active
		return false  // Don't expire while active
	}

	return fastClock().Since(s.lastUsed.Load()) > s.ttl
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
	defer s.KeepAlive()
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
func (s *Set) loadOrStoreMetricFromVec(
	m Metric,
	hash metricHash,
	family Ident,
	partialTags []Label,
	values []string,
) *namedMetric {
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
func (s *Set) loadOrStoreSetFromVec(
	hash metricHash,
	ttl time.Duration,
	isActive IsActiveFunc,
	label Label,
	value string,
) *Set {
	set := newSet()
	set.id = hash
	set.ttl = ttl
	set.isActive = isActive
	set.KeepAlive()
	set.constantTags = joinTags(s.constantTags, Tag{
		label: label,
		value: MustValue(value),
	})
	return s.loadOrStoreSet(set)
}

// KeepAlive is used to bump a Set's expiration when a TTL is set.
func (s *Set) KeepAlive() {
	if s.ttl > 0 {
		s.lastUsed.Store(fastClock().Now())
	}
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

// makeValues converts a list of string values into a list of Value objects.
func makeValues(values []string) []Value {
	new := make([]Value, len(values))
	for i, value := range values {
		new[i] = MustValue(value)
	}
	return new
}

// Collect implements the Collector interface, allowing a Set to be used as a Collector.
// It writes all metrics in the Set and its children to the provided ExpfmtWriter.
func (s *Set) Collect(w ExpfmtWriter) {
	// Append this Set's constant tags to the writer's existing tags
	constantTags := w.ConstantTags()
	if s.constantTags != "" {
		if constantTags != "" {
			constantTags = constantTags + "," + s.constantTags
		} else {
			constantTags = s.constantTags
		}
	}
	exp := ExpfmtWriter{
		b:            w.Buffer(),
		constantTags: constantTags,
	}

	s.collectInternal(exp, false)
}

// collectChildrenSets writes all child sets using the provided ExpfmtWriter,
// preserving any existing constant tags in the writer.
func (s *Set) collectChildrenSets(w ExpfmtWriter, throttle bool) {
	s.rangeChildrenSets(func(child *Set) bool {
		// Create a new writer with the child's tags appended to the current writer's tags
		childWriter := ExpfmtWriter{
			b:            w.Buffer(),
			constantTags: child.constantTags,
		}

		child.collectInternal(childWriter, throttle)

		return true
	})
}

// GetMetricUint64 returns the current value of a [Uint64] metric by family name.
//
// Returns (value, true) if the metric exists and is a [Uint64], or (0, false) if not found or wrong type.
//
// This is typically used within an [IsActiveFunc] to check metric values.
func (s *Set) GetMetricUint64(family string) (uint64, bool) {
	hash := getHashStrings(family, nil)
	if nm, ok := s.metrics.Load(hash); ok {
		if m, ok := nm.metric.(*Uint64); ok {
			return m.Get(), true
		}
	}
	return 0, false
}

// GetMetricInt64 returns the current value of an [Int64] metric by family name.
//
// Returns (value, true) if the metric exists and is an [Int64], or (0, false) if not found or wrong type.
//
// This is typically used within an [IsActiveFunc] to check metric values.
func (s *Set) GetMetricInt64(family string) (int64, bool) {
	hash := getHashStrings(family, nil)
	if nm, ok := s.metrics.Load(hash); ok {
		if m, ok := nm.metric.(*Int64); ok {
			return m.Get(), true
		}
	}
	return 0, false
}

// GetMetricFloat64 returns the current value of a [Float64] metric by family name.
//
// Returns (value, true) if the metric exists and is a [Float64], or (0, false) if not found or wrong type.
//
// This is typically used within an [IsActiveFunc] to check metric values.
func (s *Set) GetMetricFloat64(family string) (float64, bool) {
	hash := getHashStrings(family, nil)
	if nm, ok := s.metrics.Load(hash); ok {
		if m, ok := nm.metric.(*Float64); ok {
			return m.Get(), true
		}
	}
	return 0, false
}
