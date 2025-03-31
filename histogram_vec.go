package metrics

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
	hash := hashFinish(h.partialHash, values...)

	nm, ok := h.s.metrics.Load(hash)
	if !ok {
		nm = h.s.loadOrStoreMetricFromVec(
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
		partialHash: hashStart(family, labels...),
	}}
}
