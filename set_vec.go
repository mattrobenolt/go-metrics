package metrics

import "hash/maphash"

// A SetVec is a collection of Sets partitioned by label, but different value.
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
