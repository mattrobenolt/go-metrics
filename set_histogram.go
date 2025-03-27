package metrics

import (
	"errors"
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

// A HistogramVec is a collection of Histograms that are partitioned
// by the same metric name and tag labels, but different tag values.
type HistogramVec struct {
	commonVec
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

// NewHistogramVec creates a new [HistogramVec] with the supplied name.
func (s *Set) NewHistogramVec(name VecName) *HistogramVec {
	family := MustIdent(name.Family)

	return &HistogramVec{commonVec{
		s:           s,
		family:      family,
		partialTags: makePartialTags(name.Labels),
		partialHash: hashStart(family.String(), name.Labels),
	}}
}
