package metrics

import (
	"sync"
	"testing"

	"go.withmatt.com/metrics/internal/assert"
)

func TestFuncNew(t *testing.T) {
	t.Run("float", func(t *testing.T) {
		NewSet().NewFloat64Func("foo", nil)
		NewSet().NewFloat64Func("foo", nil, "bar", "baz")

		// invalid label pairs
		assert.Panics(t, func() { NewSet().NewFloat64Func("foo", nil, "bar") })

		// duplicate
		set := NewSet()
		set.NewFloat64Func("foo", nil)
		assert.Panics(t, func() { set.NewFloat64Func("foo", nil) })
	})

	t.Run("int", func(t *testing.T) {
		NewSet().NewInt64Func("foo", nil)
		NewSet().NewInt64Func("foo", nil, "bar", "baz")

		// invalid label pairs
		assert.Panics(t, func() { NewSet().NewInt64Func("foo", nil, "bar") })

		// duplicate
		set := NewSet()
		set.NewInt64Func("foo", nil)
		assert.Panics(t, func() { set.NewInt64Func("foo", nil) })
	})

	t.Run("uint", func(t *testing.T) {
		NewSet().NewUint64Func("foo", nil)
		NewSet().NewUint64Func("foo", nil, "bar", "baz")

		// invalid label pairs
		assert.Panics(t, func() { NewSet().NewUint64Func("foo", nil, "bar") })

		// duplicate
		set := NewSet()
		set.NewUint64Func("foo", nil)
		assert.Panics(t, func() { set.NewUint64Func("foo", nil) })
	})
}

func TestFuncCallback(t *testing.T) {
	t.Run("float", func(t *testing.T) {
		set := NewSet()
		set.NewFloat64Func("foo", func() float64 {
			return 1.1
		})

		assertMarshal(t, set, []string{"foo 1.1"})
	})

	t.Run("int", func(t *testing.T) {
		set := NewSet()
		set.NewInt64Func("foo", func() int64 {
			return -2
		})

		assertMarshal(t, set, []string{"foo -2"})
	})

	t.Run("uint", func(t *testing.T) {
		set := NewSet()
		set.NewUint64Func("foo", func() uint64 {
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
		g := NewSet().NewFloat64Func("x", func() float64 {
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
		g := NewSet().NewInt64Func("x", func() int64 {
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
		g := NewSet().NewUint64Func("x", func() uint64 {
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
