package metrics

import (
	"errors"
	"hash/maphash"
	"slices"
	"sync/atomic"
)

// FixedHistogramOpt are the options for creating a [FixedHistogram].
type FixedHistogramOpt struct {
	// Family is the metric Ident, see [MustIdent].
	Family Ident
	// Tags are optional tags for the metric, see [MustTags].
	Tags []Tag
	// Buckets are histogram buckets, e.g. []float64{0.1, 0.5, 1}
	Buckets []float64
}

// NewFixedHistogram creates and returns new FixedHistogram in s with the given name.
//
// family must be a Prometheus compatible identifier format.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	NewFixedHistogram("family", []float64{0.1, 0.5, 1}, "label1", "value1", "label2", "value2")
//
// The returned Histogram is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *Set) NewFixedHistogram(family string, buckets []float64, tags ...string) *FixedHistogram {
	return s.NewFixedHistogramOpt(FixedHistogramOpt{
		Family:  MustIdent(family),
		Tags:    MustTags(tags...),
		Buckets: buckets,
	})
}

// NewFixedHistogramOpt registers and returns new FixedHistogram with the opts in the s.
//
// The returned FixedHistogram is safe to use from concurrent goroutines.
//
// This will panic if already registered.
func (s *Set) NewFixedHistogramOpt(opt FixedHistogramOpt) *FixedHistogram {
	h := newFixedHistogram(opt.Buckets)
	s.mustRegisterMetric(h, opt.Family, opt.Tags)
	return h
}

// GetOrCreateFixedHistogram returns registered FixedHistogram in s with the given name
// and tags creates new Histogram if s doesn't contain FixedHistogram with the given name.
//
// family must be a Prometheus compatible identifier format.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	GetOrCreateFixedHistogram("family", []float64{0.1, 0.5, 1}, "label1", "value1", "label2", "value2")
//
// The returned FixedHistogram is safe to use from concurrent goroutines.
//
// Prefer [NewFixedHistogram] or [NewFixedHistogramOpt] when performance is critical.
//
// This will panic if values are invalid.
func (s *Set) GetOrCreateFixedHistogram(family string, buckets []float64, tags ...string) *FixedHistogram {
	hash := getHashStrings(family, tags)

	s.metricsMu.Lock()
	nm := s.metrics[hash]
	s.metricsMu.Unlock()

	if nm == nil {
		nm = s.getOrAddMetricFromStrings(newFixedHistogram(buckets), hash, family, tags)
	}
	return nm.metric.(*FixedHistogram)
}

// FixedHistogramVecOpt are options for creating a new [FixedHistogramVec].
type FixedHistogramVecOpt struct {
	Family  string
	Labels  []string
	Buckets []float64
}

// A FixedHistogramVec is a collection of FixedHistogramVecs that are partitioned
// by the same metric name and tag labels, but different tag values.
type FixedHistogramVec struct {
	s           *Set
	family      Ident
	partialTags []Tag
	partialHash *maphash.Hash

	buckets []float64
	labels  []string
}

// WithLabelValues returns the FixedHistogram for the corresponding label values.
// If the combination of values is seen for the first time, a new FixedHistogram
// is created.
//
// This will panic if the values count doesn't match the number of labels.
func (h *FixedHistogramVec) WithLabelValues(values ...string) *FixedHistogram {
	if len(values) != len(h.partialTags) {
		panic(errors.New("mismatch length of labels"))
	}
	hash := hashFinish(h.partialHash, values)

	h.s.metricsMu.Lock()
	nm := h.s.metrics[hash]
	h.s.metricsMu.Unlock()

	if nm == nil {
		nm = h.s.getOrRegisterMetricFromVec(
			&FixedHistogram{
				buckets:      slices.Clone(h.buckets),
				labels:       h.labels,
				observations: make([]atomic.Uint64, len(h.buckets)),
			}, hash, h.family, h.partialTags, values,
		)
	}
	return nm.metric.(*FixedHistogram)
}

// NewFixedHistogramVec creates a new [FixedHistogramVec] with the supplied opt.
func (s *Set) NewFixedHistogramVec(opt FixedHistogramVecOpt) *FixedHistogramVec {
	family := MustIdent(opt.Family)

	buckets := opt.Buckets
	if len(buckets) == 0 {
		buckets = slices.Clone(DefBuckets)
	} else {
		buckets = slices.Clone(buckets)
		slices.Sort(buckets)
	}

	return &FixedHistogramVec{
		s:           s,
		family:      family,
		partialTags: makePartialTags(opt.Labels),
		partialHash: hashStart(family.String(), opt.Labels),

		buckets: buckets,
		labels:  labelsForBuckets(buckets),
	}
}
