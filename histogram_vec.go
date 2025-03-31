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
	set := h.set
	if set == nil {
		set = h.setvec.WithLabelValue(values[0])
		values = values[1:]
	}
	return h.withLabelValues(set, values)
}

func (h *HistogramVec) withLabelValues(set *Set, values []string) *Histogram {
	hash := hashFinish(h.partialHash, values...)

	nm, ok := set.metrics.Load(hash)
	if !ok {
		nm = set.loadOrStoreMetricFromVec(
			&Histogram{}, hash, h.family, h.partialTags, values,
		)
	}
	return nm.metric.(*Histogram)
}

// NewHistogramVec creates a new [HistogramVec] with the supplied name.
func (s *Set) NewHistogramVec(family string, labels ...string) *HistogramVec {
	return &HistogramVec{getCommonVecSet(s, family, labels)}
}
