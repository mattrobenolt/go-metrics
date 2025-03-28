package metrics

import "errors"

// A HistogramVec is a collection of Histograms that are partitioned
// by the same metric name and tag labels, but different tag values.
type HistogramVec struct {
	commonVec
}

// NewHistogramVec creates a new HistogramVec on the global Set.
// See [Set.NewHistogramVec].
func NewHistogramVec(family string, labels ...string) *HistogramVec {
	return defaultSet.NewHistogramVec(family, labels...)
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
func (s *Set) NewHistogramVec(family string, labels ...string) *HistogramVec {
	return &HistogramVec{commonVec{
		s:           s,
		family:      MustIdent(family),
		partialTags: makePartialTags(labels),
		partialHash: hashStart(family, labels),
	}}
}
