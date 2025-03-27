package metrics

import (
	"errors"
	"hash/maphash"
)

// NewHistogram creates and returns new Histogram in s with the given name.
//
// family must be a Prometheus compatible identifier format.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	NewHistogram("family", "label1", "value1", "label2", "value2")
//
// The returned Histogram is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *Set) NewHistogram(family string, tags ...string) *Histogram {
	return s.NewHistogramOpt(MetricName{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
	})
}

// NewHistogramOpt registers and returns new Histogram with the name in the s.
//
// The returned Histogram is safe to use from concurrent goroutines.
//
// This will panic if already registered.
func (s *Set) NewHistogramOpt(name MetricName) *Histogram {
	h := &Histogram{}
	s.mustRegisterMetric(h, name)
	return h
}

// GetOrCreateHistogram returns registered Histogram in s with the given name
// and tags creates new Histogram if s doesn't contain Histogram with the given name.
//
// family must be a Prometheus compatible identifier format.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	GetOrCreateHistogram("family", "label1", "value1", "label2", "value2")
//
// The returned Histogram is safe to use from concurrent goroutines.
//
// Prefer [NewHistogram] or [NewHistogramOpt] when performance is critical.
//
// This will panic if values are invalid.
func (s *Set) GetOrCreateHistogram(family string, tags ...string) *Histogram {
	hash := getHashStrings(family, tags)

	s.metricsMu.Lock()
	nm := s.metrics[hash]
	s.metricsMu.Unlock()

	if nm == nil {
		nm = s.getOrAddMetricFromStrings(&Histogram{}, hash, family, tags)
	}
	return nm.metric.(*Histogram)
}

// HistogramVecOpt are options for creating a new [HistogramVec].
type HistogramVecOpt struct {
	// Family is the metric family name, e.g. `http_requests`
	Family string
	// Labels are the tag labels that you want to partition on, e.g. "status", "path"
	Labels []string
}

// A HistogramVec is a collection of Histograms that are partitioned
// by the same metric name and tag labels, but different tag values.
type HistogramVec struct {
	s           *Set
	family      Ident
	partialTags []Tag
	partialHash *maphash.Hash
}

// WithLabelValues returns the Histogram for the corresponding label values.
// If the combination of values is seen for the first time, a new Histogram
// is created.
//
// This will panic if the values count doesn't match the number of labels.
func (h *HistogramVec) WithLabelValues(values ...string) *Histogram {
	if len(values) != len(h.partialTags) {
		panic(errors.New("mismatch length of labels"))
	}
	hash := hashFinish(h.partialHash, values)

	h.s.metricsMu.Lock()
	nm := h.s.metrics[hash]
	h.s.metricsMu.Unlock()

	if nm == nil {
		nm = h.s.getOrRegisterMetricFromVec(
			&Histogram{}, hash, h.family, h.partialTags, values,
		)
	}
	return nm.metric.(*Histogram)
}

// NewHistogramVec creates a new [HistogramVec] with the supplied opt.
func (s *Set) NewHistogramVec(opt HistogramVecOpt) *HistogramVec {
	family := MustIdent(opt.Family)

	return &HistogramVec{
		s:           s,
		family:      family,
		partialTags: makePartialTags(opt.Labels),
		partialHash: hashStart(family.String(), opt.Labels),
	}
}
