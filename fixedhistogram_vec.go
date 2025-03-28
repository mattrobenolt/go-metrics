package metrics

import (
	"errors"
	"slices"
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
func (s *Set) NewFixedHistogramVec(family string, buckets []float64, labels ...string) *FixedHistogramVec {
	if len(buckets) == 0 {
		buckets = slices.Clone(DefBuckets)
	} else {
		buckets = slices.Clone(buckets)
		slices.Sort(buckets)
	}

	return &FixedHistogramVec{
		commonVec: commonVec{
			s:           s,
			family:      MustIdent(family),
			partialTags: makePartialTags(labels),
			partialHash: hashStart(family, labels),
		},
		buckets: buckets,
		labels:  labelsForBuckets(buckets),
	}
}
