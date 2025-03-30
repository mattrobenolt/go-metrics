package syncx

import (
	"slices"
	"sync"
	"sync/atomic"
)

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

func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}

type SortedMap[K comparable, V any] struct {
	m      Map[K, V]
	cmp    func(V, V) int
	dirty  atomic.Uint64
	values atomic.Pointer[[]V]
}

func (m *SortedMap[K, V]) Init(cmp func(V, V) int) {
	m.cmp = cmp
}

func (m *SortedMap[K, V]) Clear() {
	m.m.Clear()
	m.values.Store(nil)
	m.dirty.Store(0)
}

func (m *SortedMap[K, V]) Load(key K) (value V, ok bool) {
	return m.m.Load(key)
}

func (m *SortedMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	actual, loaded = m.m.LoadOrStore(key, value)
	if !loaded {
		m.dirty.Add(1)
	}
	return actual, loaded
}

func (m *SortedMap[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
	m.dirty.Add(1)
}

func (m *SortedMap[K, V]) Delete(key K) {
	m.m.Delete(key)
	m.dirty.Add(1)
}

func (m *SortedMap[K, V]) Values() []V {
	if d := m.dirty.Load(); d > 0 {
		m.updateDirtyValues(d)
	}
	if values := m.values.Load(); values != nil {
		return *values
	}
	return nil
}

func (m *SortedMap[K, V]) updateDirtyValues(dirty uint64) {
	var (
		prevValues *[]V
		newValues  []V
	)
	prevValues = m.values.Load()
	if prevValues == nil && newValues == nil {
		newValues = make([]V, 0, 16)
	} else if newValues == nil {
		newValues = make([]V, 0, len(*prevValues))
	} else {
		newValues = newValues[:0]
	}

	m.m.Range(func(key K, value V) bool {
		idx, _ := slices.BinarySearchFunc(newValues, value, m.cmp)
		newValues = slices.Insert(newValues, idx, value)
		return true
	})

	newValues = slices.Clip(newValues)

	if m.values.CompareAndSwap(prevValues, &newValues) {
		m.dirty.CompareAndSwap(dirty, 0)
	}
}
