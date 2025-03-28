package metrics

import (
	"sync"
	"testing"

	"go.withmatt.com/metrics/internal/assert"
)

func TestFuncNew(t *testing.T) {
	t.Run("float", func(t *testing.T) {
		NewSet().NewFloatFunc("foo", nil)
		NewSet().NewFloatFunc("foo", nil, "bar", "baz")

		// invalid label pairs
		assert.Panics(t, func() { NewSet().NewFloatFunc("foo", nil, "bar") })

		// duplicate
		set := NewSet()
		set.NewFloatFunc("foo", nil)
		assert.Panics(t, func() { set.NewFloatFunc("foo", nil) })
	})

	t.Run("int", func(t *testing.T) {
		NewSet().NewIntFunc("foo", nil)
		NewSet().NewIntFunc("foo", nil, "bar", "baz")

		// invalid label pairs
		assert.Panics(t, func() { NewSet().NewIntFunc("foo", nil, "bar") })

		// duplicate
		set := NewSet()
		set.NewIntFunc("foo", nil)
		assert.Panics(t, func() { set.NewIntFunc("foo", nil) })
	})

	t.Run("uint", func(t *testing.T) {
		NewSet().NewUintFunc("foo", nil)
		NewSet().NewUintFunc("foo", nil, "bar", "baz")

		// invalid label pairs
		assert.Panics(t, func() { NewSet().NewUintFunc("foo", nil, "bar") })

		// duplicate
		set := NewSet()
		set.NewUintFunc("foo", nil)
		assert.Panics(t, func() { set.NewUintFunc("foo", nil) })
	})
}

func TestFuncCallback(t *testing.T) {
	t.Run("float", func(t *testing.T) {
		set := NewSet()
		set.NewFloatFunc("foo", func() float64 {
			return 1.1
		})

		assertMarshal(t, set, []string{"foo 1.1"})
	})

	t.Run("int", func(t *testing.T) {
		set := NewSet()
		set.NewIntFunc("foo", func() int64 {
			return -2
		})

		assertMarshal(t, set, []string{"foo -2"})
	})

	t.Run("uint", func(t *testing.T) {
		set := NewSet()
		set.NewUintFunc("foo", func() uint64 {
			return 100
		})

		assertMarshal(t, set, []string{"foo 100"})
	})
}

func TestFuncConcurrent(t *testing.T) {
	const n = 1000
	const inner = 5

	t.Run("float", func(t *testing.T) {
		var x float64
		var nLock sync.Mutex
		g := NewSet().NewFloatFunc("x", func() float64 {
			nLock.Lock()
			defer nLock.Unlock()
			x++
			return x
		})
		hammer(t, n, func(_ int) {
			prevX := g.Get()
			for range inner {
				assert.Greater(t, g.Get(), prevX)
			}
		})

		assert.Equal(t, g.Get(), (n*inner)+n+1)
	})

	t.Run("int", func(t *testing.T) {
		var x int64
		var nLock sync.Mutex
		g := NewSet().NewIntFunc("x", func() int64 {
			nLock.Lock()
			defer nLock.Unlock()
			x++
			return x
		})
		hammer(t, n, func(_ int) {
			prevX := g.Get()
			for range inner {
				assert.Greater(t, g.Get(), prevX)
			}
		})

		assert.Equal(t, g.Get(), (n*inner)+n+1)
	})

	t.Run("uint", func(t *testing.T) {
		var x uint64
		var nLock sync.Mutex
		g := NewSet().NewUintFunc("x", func() uint64 {
			nLock.Lock()
			defer nLock.Unlock()
			x++
			return x
		})
		hammer(t, n, func(_ int) {
			prevX := g.Get()
			for range inner {
				assert.Greater(t, g.Get(), prevX)
			}
		})

		assert.Equal(t, g.Get(), (n*inner)+n+1)
	})
}
