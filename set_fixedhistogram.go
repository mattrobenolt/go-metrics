package metrics

import (
	"errors"
	"slices"
	"sync/atomic"
)

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
	h := newFixedHistogram(buckets)
	s.mustRegisterMetric(h, MetricName{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
	})
	return h
}

// A FixedHistogramVec is a collection of FixedHistogramVecs that are partitioned
// by the same metric name and tag labels, but different tag values.
type FixedHistogramVec struct {
	commonVec
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
