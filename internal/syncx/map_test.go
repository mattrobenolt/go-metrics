package syncx

import (
	"cmp"
	"runtime"
	"slices"
	"testing"

	"go.withmatt.com/metrics/internal/assert"
)

func TestSortedMap(t *testing.T) {
	var sm SortedMap[int, int]
	sm.Init(cmp.Compare[int])
	sm.Clear()

	vals := sm.Values()
	assert.Equal(t, len(vals), 0)

	_, ok := sm.Load(0)
	assert.False(t, ok)

	sm.Store(0, 0)
	sm.Store(1, 1)
	v, _ := sm.Load(0)
	assert.Equal(t, v, 0)
	v, _ = sm.Load(1)
	assert.Equal(t, v, 1)

	assert.Equal(t, sm.dirty.Load(), 2)

	vals = sm.Values()
	assert.Equal(t, len(vals), 2)
	assert.Equal(t, sm.dirty.Load(), 0)
	assert.Equal(t, vals[0], 0)
	assert.Equal(t, vals[1], 1)

	assert.True(t, slices.IsSortedFunc(vals, sm.cmp))

	sm.Delete(0)
	assert.Equal(t, sm.dirty.Load(), 1)
	assert.Equal(t, len(sm.Values()), 1)
	assert.Equal(t, sm.dirty.Load(), 0)
	sm.Delete(0)
	assert.Equal(t, sm.dirty.Load(), 0)
	assert.Equal(t, len(sm.Values()), 1)

	v, loaded := sm.LoadOrStore(1, 2)
	assert.True(t, loaded)
	assert.Equal(t, v, 1)
	assert.Equal(t, sm.dirty.Load(), 0)

	v, loaded = sm.LoadOrStore(2, 2)
	assert.False(t, loaded)
	assert.Equal(t, v, 2)
	assert.Equal(t, sm.dirty.Load(), 1)

	sm.Clear()
	assert.Equal(t, sm.dirty.Load(), 0)
	assert.Equal(t, len(sm.Values()), 0)
}

func TestConcurrentSortedMap(t *testing.T) {
	var sm SortedMap[int, int]
	sm.Init(cmp.Compare[int])

	const p = 4
	const n = 1000
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(p))

	c := make(chan int)
	for range p {
		go func() {
			defer func() {
				assert.Nil(t, recover())
				c <- 1
			}()
			for i := range n {
				sm.Store(i, i)
				for range sm.Values() {
				}
			}
		}()
	}
	for range p {
		<-c
	}

	for i := range n {
		v, _ := sm.Load(i)
		assert.Equal(t, v, i)
	}

	vals := sm.Values()
	assert.True(t, slices.IsSortedFunc(vals, sm.cmp))
	assert.Equal(t, len(vals), n)

	for i, v := range vals {
		assert.Equal(t, v, i)
	}

	var i int
	for v := range sm.Values() {
		assert.Equal(t, v, i)
		i++
	}
}

func TestConcurrentSortedMapChaos(t *testing.T) {
	var sm SortedMap[int, int]
	sm.Init(cmp.Compare[int])
	const p = 4
	const n = 100000
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(p))

	c := make(chan int)
	for range p {
		go func() {
			defer func() {
				assert.Nil(t, recover())
				c <- 1
			}()
			for i := range n {
				sm.Clear()
				sm.Store(i, i)
				for range sm.Values() {
				}
				sm.Load(i)
				sm.LoadOrStore(i, i)
				sm.Delete(i)
			}
		}()
	}
	for range p {
		<-c
	}
}
