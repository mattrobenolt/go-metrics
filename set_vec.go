package metrics

import (
	"hash/maphash"
	"time"
)

// A SetVec is a collection of Sets partitioned by label, but different value.
// The primary use-case is being able to destroy/delete entire sets of metrics
// by a common label.
type SetVec struct {
	s           *Set
	label       Label
	partialHash *maphash.Hash
	ttl         time.Duration
}

// NewSetVec creates a new SetVec on the global Set.
// See [Set.NewSetVec].
func NewSetVec(label string) *SetVec {
	return defaultSet.NewSetVec(label)
}

// NewSetVecWithTTL creates a new SetVec on the global Set with a TTL.
// See [Set.NewSetVecWithTTL].
func NewSetVecWithTTL(label string, ttl time.Duration) *SetVec {
	return defaultSet.NewSetVecWithTTL(label, ttl)
}

// NewSetVec creates a new [SetVec] with the given label.
func (s *Set) NewSetVec(label string) *SetVec {
	return &SetVec{
		s:           s,
		label:       MustLabel(label),
		partialHash: hashStart("", label),
	}
}

// NewSetVecWithTTL creates a new [SetVec] with the given label and TTL.
// See [Set.KeepAlive] to manually keep a specific Set alive.
func (s *Set) NewSetVecWithTTL(label string, ttl time.Duration) *SetVec {
	sv := s.NewSetVec(label)
	sv.ttl = ttl
	return sv
}

// WithLabelValue returns the Set for the corresponding label value.
// If the combination of values is seen for the first time, a new Set
// is created.
//
// If the SetVec was created with a TTL, the Set will be automatically kept
// alive when this function is called. See [Set.KeepAlive] to manually keep a
// the returned Set alive if necessary.
//
// See [Set.KeepAlive] to manually keep a specific Set alive.
func (sv *SetVec) WithLabelValue(value string) *Set {
	hash := hashFinish(sv.partialHash, value)

	set, ok := sv.s.setsByHash.Load(hash)
	if !ok {
		set = sv.s.loadOrStoreSetFromVec(hash, sv.ttl, sv.label, value)
	}
	set.KeepAlive()
	return set
}

// RemoveByLabelValue removes the Set for the corresponding label value.
func (sv *SetVec) RemoveByLabelValue(value string) {
	sv.s.setsByHash.Delete(hashFinish(sv.partialHash, value))
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
func (sv *SetVec) NewUint64(family string, value string, tags ...string) *Uint64 {
	return sv.WithLabelValue(value).NewUint64(family, tags...)
}

// NewCounter is an alias for [SetVec.NewUint64].
func (sv *SetVec) NewCounter(family string, value string, tags ...string) *Uint64 {
	return sv.WithLabelValue(value).NewCounter(family, tags...)
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
func (sv *SetVec) NewInt64(family string, value string, tags ...string) *Int64 {
	return sv.WithLabelValue(value).NewInt64(family, tags...)
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
func (sv *SetVec) NewFloat64(family string, value string, tags ...string) *Float64 {
	return sv.WithLabelValue(value).NewFloat64(family, tags...)
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
func (sv *SetVec) NewFixedHistogram(family string, buckets []float64, value string, tags ...string) *FixedHistogram {
	return sv.WithLabelValue(value).NewFixedHistogram(family, buckets, tags...)
}

// NewFixedHistogramVec creates a new [FixedHistogramVec] with the supplied opt.
func (sv *SetVec) NewFixedHistogramVec(family string, buckets []float64, labels ...string) *FixedHistogramVec {
	buckets = getBuckets(buckets)

	return &FixedHistogramVec{
		commonVec: getCommonVecSetVec(sv, family, labels),
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
func (sv *SetVec) NewHistogram(family string, value string, tags ...string) *Histogram {
	return sv.WithLabelValue(value).NewHistogram(family, tags...)
}

// NewHistogramVec creates a new [HistogramVec] with the supplied name.
func (sv *SetVec) NewHistogramVec(family string, labels ...string) *HistogramVec {
	return &HistogramVec{getCommonVecSetVec(sv, family, labels)}
}
