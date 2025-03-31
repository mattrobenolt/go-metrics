package syncx

// Set is a collection of unique entries, safe for concurrent use.
type Set[E comparable] struct {
	m Map[E, struct{}]
}

// Clear removes all entries from the set.
func (s *Set[E]) Clear() {
	s.m.Clear()
}

// Add adds an entry to the set. Returns true if the entry was added,
// false if it already existed.
func (s *Set[E]) Add(entry E) (added bool) {
	_, loaded := s.m.LoadOrStore(entry, struct{}{})
	return !loaded
}

// Delete removes an element from the set.
func (s *Set[E]) Delete(entry E) {
	s.m.Delete(entry)
}

// Range iterates over all entries in the Set.
func (s *Set[E]) Range(f func(entry E) bool) {
	s.m.Range(func(entry E, _ struct{}) bool {
		return f(entry)
	})
}
