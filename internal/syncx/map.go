package syncx

import (
	"slices"
	"sync"
	"sync/atomic"
)

// Map wraps a sync.Map and makes it typed, would really like to use
// internal/sync.HashTrieMap some day.
type Map[K comparable, V any] struct {
	m sync.Map
}

func (m *Map[K, V]) Clear() {
	m.m.Clear()
}

func (m *Map[K, V]) Load(key K) (value V, ok bool) {
	v, ok := m.m.Load(key)
	if v == nil {
		var zero V
		return zero, ok
	}
	return v.(V), ok
}

func (m *Map[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	v, loaded := m.m.LoadOrStore(key, value)
	if v == nil {
		var zero V
		return zero, loaded
	}
	return v.(V), loaded
}

func (m *Map[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

func (m *Map[K, V]) Delete(key K) {
	m.m.Delete(key)
}

func (m *Map[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	v, loaded := m.m.LoadAndDelete(key)
	if v == nil {
		var zero V
		return zero, loaded
	}
	return v.(V), loaded
}

func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}

// SortedMap is a concurrent Map implementation that maintains a sorted list of
// values.
//
// This implementation is optimized for read-heavy workloads while the values
// are eventually consistent. Adding and removing elements is entirely
// non-blocking.
type SortedMap[K comparable, V any] struct {
	m   Map[K, V]
	cmp func(V, V) int

	// dirty is used to track the number of dirty updates to the map. When dirty
	// is greater than zero, we know the values list is out of date and needs to
	// be recomputed. Every mutation to the map will increment this counter.
	dirty  atomic.Uint64
	values atomic.Pointer[[]V]
}

// Init sets the comparison function to use for sorting values.
func (m *SortedMap[K, V]) Init(cmp func(V, V) int) {
	m.cmp = cmp
}

// Clear deletes all the entries, resulting in an empty SortedMap.
func (m *SortedMap[K, V]) Clear() {
	m.m.Clear()
	m.values.Store(nil)
	m.dirty.Store(0)
}

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (m *SortedMap[K, V]) Load(key K) (value V, ok bool) {
	return m.m.Load(key)
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *SortedMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	actual, loaded = m.m.LoadOrStore(key, value)
	if !loaded {
		m.dirty.Add(1)
	}
	return actual, loaded
}

// Store sets the value for a key.
func (m *SortedMap[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
	m.dirty.Add(1)
}

// Delete deletes the value for a key.
func (m *SortedMap[K, V]) Delete(key K) {
	if _, loaded := m.m.LoadAndDelete(key); loaded {
		m.dirty.Add(1)
	}
}

// Values will return a sorted slice of values that represent a snapshot of the
// map. This is safe for concurrent use, and is not guaranteed to be exactly in
// sync with the map itself, it may have stale or missing values.
func (m *SortedMap[K, V]) Values() []V {
	if d := m.dirty.Load(); d > 0 {
		m.updateDirtyValues(d)
	}
	if values := m.values.Load(); values != nil {
		return *values
	}
	return nil
}

// updateDirtyValues eventually consistently updates the sorted values list.
// By the end of this function, if we were concurrent with another thread
// and lost, ultimately nothing will happen. If we win the race, the values
// list will be updated, and we will attempt to reset the dirty counter. If new
// changes have been made since we started, we will not reset the dirty counter.
func (m *SortedMap[K, V]) updateDirtyValues(oldDirty uint64) {
	startSize := 16 // start with something sensible to avoid early resizing
	oldValues := m.values.Load()
	if oldValues != nil {
		startSize = len(*oldValues)
	}

	newValues := make([]V, 0, startSize)

	// collect all values in sorted order
	m.m.Range(func(key K, value V) bool {
		idx, _ := slices.BinarySearchFunc(newValues, value, m.cmp)
		newValues = slices.Insert(newValues, idx, value)
		return true
	})

	// trim any excess capacity before storing
	newValues = slices.Clip(newValues)

	// only replace the values if another thread hasn't beat us
	if m.values.CompareAndSwap(oldValues, &newValues) {
		// if we did win the race and updated the values list, we can now
		// attempt to reset the dirty counter. We only want to reset the dirty
		// counter if nothing new has dirtied since we started.
		m.dirty.CompareAndSwap(oldDirty, 0)
	}
}
