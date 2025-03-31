package metrics

import (
	"sync/atomic"
)

// A FixedHistogramVec is a collection of FixedHistogramVecs that are partitioned
// by the same metric name and tag labels, but different tag values.
type FixedHistogramVec struct {
	commonVec
	buckets []float64
	labels  []string
}

// NewFixedHistogramVec creates a new FixedHistogramVec on the global Set.
// See [Set.NewFixedHistogramVec].
func NewFixedHistogramVec(family string, buckets []float64, labels ...string) *FixedHistogramVec {
	return defaultSet.NewFixedHistogramVec(family, buckets, labels...)
}

// WithLabelValues returns the FixedHistogram for the corresponding label values.
// If the combination of values is seen for the first time, a new FixedHistogram
// is created.
//
// This will panic if the values count doesn't match the number of labels.
func (h *FixedHistogramVec) WithLabelValues(values ...string) *FixedHistogram {
	set := h.set
	if set == nil {
		set = h.setvec.WithLabelValue(values[0])
		values = values[1:]
	}
	return h.withLabelValues(set, values)
}

func (h *FixedHistogramVec) withLabelValues(set *Set, values []string) *FixedHistogram {
	hash := hashFinish(h.partialHash, values...)

	nm, ok := set.metrics.Load(hash)
	if !ok {
		nm = set.loadOrStoreMetricFromVec(
			&FixedHistogram{
				buckets:      h.buckets,
				labels:       h.labels,
				observations: make([]atomic.Uint64, len(h.buckets)),
			}, hash, h.family, h.partialTags, values,
		)
	}
	return nm.metric.(*FixedHistogram)
}

// NewFixedHistogramVec creates a new [FixedHistogramVec] with the supplied opt.
func (s *Set) NewFixedHistogramVec(family string, buckets []float64, labels ...string) *FixedHistogramVec {
	buckets = getBuckets(buckets)

	return &FixedHistogramVec{
		commonVec: getCommonVecSet(s, family, labels),
		buckets:   buckets,
		labels:    labelsForBuckets(buckets),
	}
}
