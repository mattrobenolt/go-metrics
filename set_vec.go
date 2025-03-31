package metrics

import (
	"hash/maphash"
)

// A SetVec is a collection of Sets partitioned by label, but different value.
// The primary use-case is being able to destroy/delete entire sets of metrics
// by a common label.
type SetVec struct {
	s           *Set
	label       Label
	partialHash *maphash.Hash
}

// NewSetVec creates a new SetVec on the global Set.
// See [Set.NewSetVec].
func NewSetVec(label string) *SetVec {
	return defaultSet.NewSetVec(label)
}

// NewSetVec creates a new [SetVec] with the given label.
func (s *Set) NewSetVec(label string) *SetVec {
	return &SetVec{
		s:           s,
		label:       MustLabel(label),
		partialHash: hashStart("", label),
	}
}

// WithLabelValues returns the Set for the corresponding label value.
// If the combination of values is seen for the first time, a new Set
// is created.
func (s *SetVec) WithLabelValue(value string) *Set {
	hash := hashFinish(s.partialHash, value)

	set, ok := s.s.setsByHash.Load(hash)
	if !ok {
		set = s.s.loadOrStoreSetFromVec(hash, s.label, value)
	}
	return set
}

// RemoveByLabelValue removes the Set for the corresponding label value.
func (s *SetVec) RemoveByLabelValue(value string) {
	s.s.setsByHash.Delete(hashFinish(s.partialHash, value))
}

// NewUint64Vec creates a new [Uint64Vec] with the supplied name.
func (sv *SetVec) NewUint64Vec(family string, labels ...string) *Uint64Vec {
	return &Uint64Vec{getCommonVecSetVec(sv, family, labels)}
}

// NewUint64 registers and returns new Uint64 using the label from the SetVec.
//
// family must be a Prometheus compatible identifier format.
//
//	NewUint64("family", "value1")
//
// The returned Uint64 is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *SetVec) NewUint64(family string, value string, tags ...string) *Uint64 {
	return s.WithLabelValue(value).NewUint64(family, tags...)
}

// NewCounter is an alias for [SetVec.NewUint64].
func (s *SetVec) NewCounter(family string, value string, tags ...string) *Uint64 {
	return s.WithLabelValue(value).NewCounter(family, tags...)
}

// NewInt64Vec creates a new [Int64Vec] with the supplied name.
func (sv *SetVec) NewInt64Vec(family string, labels ...string) *Int64Vec {
	return &Int64Vec{getCommonVecSetVec(sv, family, labels)}
}

// NewInt64 registers and returns new Int64 using the label from the SetVec.
//
// family must be a Prometheus compatible identifier format.
//
//	NewInt64("family", "value1")
//
// The returned Int64 is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *SetVec) NewInt64(family string, value string, tags ...string) *Int64 {
	return s.WithLabelValue(value).NewInt64(family, tags...)
}

// NewFloat64Vec creates a new [Float64Vec] with the supplied name.
func (sv *SetVec) NewFloat64Vec(family string, labels ...string) *Float64Vec {
	return &Float64Vec{getCommonVecSetVec(sv, family, labels)}
}

// NewFloat64 registers and returns new Float64 using the label from the SetVec.
//
// family must be a Prometheus compatible identifier format.
//
//	NewFloat64("family", "value1")
//
// The returned Float64 is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *SetVec) NewFloat64(family string, value string, tags ...string) *Float64 {
	return s.WithLabelValue(value).NewFloat64(family, tags...)
}

// NewFixedHistogram creates and returns new FixedHistogram using the label from the SetVec.
//
// family must be a Prometheus compatible identifier format.
//
//	NewFixedHistogram("family", []float64{0.1, 0.5, 1}, "value1")
//
// The returned FixedHistogram is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *SetVec) NewFixedHistogram(family string, buckets []float64, value string, tags ...string) *FixedHistogram {
	return s.WithLabelValue(value).NewFixedHistogram(family, buckets, tags...)
}

// NewFixedHistogramVec creates a new [FixedHistogramVec] with the supplied opt.
func (s *SetVec) NewFixedHistogramVec(family string, buckets []float64, labels ...string) *FixedHistogramVec {
	buckets = getBuckets(buckets)

	return &FixedHistogramVec{
		commonVec: getCommonVecSetVec(s, family, labels),
		buckets:   buckets,
		labels:    labelsForBuckets(buckets),
	}
}

// NewHistogram creates and returns new Histogram using the label from the SetVec.
//
// family must be a Prometheus compatible identifier format.
//
//	NewHistogram("family", "value1")
//
// The returned Histogram is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *SetVec) NewHistogram(family string, value string, tags ...string) *Histogram {
	return s.WithLabelValue(value).NewHistogram(family, tags...)
}

// NewHistogramVec creates a new [HistogramVec] with the supplied name.
func (sv *SetVec) NewHistogramVec(family string, labels ...string) *HistogramVec {
	return &HistogramVec{getCommonVecSetVec(sv, family, labels)}
}
