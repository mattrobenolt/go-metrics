package metrics

import (
	"testing"

	"go.withmatt.com/metrics/internal/assert"
)

func TestCounterNew(t *testing.T) {
	t.Run("uint", func(t *testing.T) {
		NewSet().NewUint("foo")
		NewSet().NewUint("foo", "bar", "baz")

		// invalid label pairs
		assert.Panics(t, func() { NewSet().NewUint("foo", "bar") })

		// duplicate
		set := NewSet()
		set.NewUint("foo")
		assert.Panics(t, func() { set.NewUint("foo") })
	})

	t.Run("int", func(t *testing.T) {
		NewSet().NewInt("foo")
		NewSet().NewInt("foo", "bar", "baz")

		// invalid label pairs
		assert.Panics(t, func() { NewSet().NewInt("foo", "bar") })

		// duplicate
		set := NewSet()
		set.NewInt("foo")
		assert.Panics(t, func() { set.NewInt("foo") })
	})

	t.Run("float", func(t *testing.T) {
		NewSet().NewFloat("foo")
		NewSet().NewFloat("foo", "bar", "baz")

		// invalid label pairs
		assert.Panics(t, func() { NewSet().NewFloat("foo", "bar") })

		// duplicate
		set := NewSet()
		set.NewFloat("foo")
		assert.Panics(t, func() { set.NewFloat("foo") })
	})
}

func TestCounterVec(t *testing.T) {
	t.Run("uint", func(t *testing.T) {
		set := NewSet()
		c := set.NewUintVec("foo", "a", "b")
		c.WithLabelValues("1", "2").Inc()
		c.WithLabelValues("1", "2").Inc()
		c.WithLabelValues("3", "4").Inc()

		assert.Equal(t, c.WithLabelValues("1", "2").Get(), 2)
		assert.Equal(t, c.WithLabelValues("3", "4").Get(), 1)

		// order is unpredictable bc the tags aren't ordered
		assertMarshalUnordered(t, set, []string{
			`foo{a="1",b="2"} 2`,
			`foo{a="3",b="4"} 1`,
		})
	})

	t.Run("int", func(t *testing.T) {
		set := NewSet()
		c := set.NewIntVec("foo", "a", "b")
		c.WithLabelValues("1", "2").Inc()
		c.WithLabelValues("1", "2").Inc()
		c.WithLabelValues("3", "4").Inc()

		assert.Equal(t, c.WithLabelValues("1", "2").Get(), 2)
		assert.Equal(t, c.WithLabelValues("3", "4").Get(), 1)

		// order is unpredictable bc the tags aren't ordered
		assertMarshalUnordered(t, set, []string{
			`foo{a="1",b="2"} 2`,
			`foo{a="3",b="4"} 1`,
		})
	})

	t.Run("float", func(t *testing.T) {
		set := NewSet()
		c := set.NewFloatVec("foo", "a", "b")
		c.WithLabelValues("1", "2").Inc()
		c.WithLabelValues("1", "2").Inc()
		c.WithLabelValues("3", "4").Inc()

		assert.Equal(t, c.WithLabelValues("1", "2").Get(), 2)
		assert.Equal(t, c.WithLabelValues("3", "4").Get(), 1)

		// order is unpredictable bc the tags aren't ordered
		assertMarshalUnordered(t, set, []string{
			`foo{a="1",b="2"} 2`,
			`foo{a="3",b="4"} 1`,
		})
	})
}

func TestCounterSerial(t *testing.T) {
	t.Run("uint", func(t *testing.T) {
		set := NewSet()
		c := set.NewUint("foo")
		c.Inc()
		assert.Equal(t, c.Get(), 1)
		c.Dec()
		assert.Equal(t, c.Get(), 0)
		c.Set(123)
		assert.Equal(t, c.Get(), 123)
		c.Dec()
		assert.Equal(t, c.Get(), 122)
		c.Add(3)
		assert.Equal(t, c.Get(), 125)

		assertMarshal(t, set, []string{"foo 125"})
	})

	t.Run("int", func(t *testing.T) {
		set := NewSet()
		c := set.NewInt("foo")
		c.Inc()
		assert.Equal(t, c.Get(), 1)
		c.Dec()
		assert.Equal(t, c.Get(), 0)
		c.Set(123)
		assert.Equal(t, c.Get(), 123)
		c.Dec()
		assert.Equal(t, c.Get(), 122)
		c.Add(3)
		assert.Equal(t, c.Get(), 125)
		c.Set(-1)

		assertMarshal(t, set, []string{"foo -1"})
	})

	t.Run("float", func(t *testing.T) {
		set := NewSet()
		c := set.NewFloat("foo")
		c.Inc()
		assert.Equal(t, c.Get(), 1)
		c.Dec()
		assert.Equal(t, c.Get(), 0)
		c.Set(123)
		assert.Equal(t, c.Get(), 123)
		c.Dec()
		assert.Equal(t, c.Get(), 122)
		c.Add(3)
		assert.Equal(t, c.Get(), 125)
		c.Set(1.1)

		assertMarshal(t, set, []string{"foo 1.1"})
	})
}

func TestCounterConcurrent(t *testing.T) {
	const n = 1000
	const inner = 10

	t.Run("uint", func(t *testing.T) {
		c := NewSet().NewUint("x")
		hammer(t, n, func(_ int) {
			nPrev := c.Get()
			for range inner {
				c.Inc()
				assert.Greater(t, c.Get(), nPrev)
			}
		})
		assert.Equal(t, c.Get(), n*inner)
	})

	t.Run("int", func(t *testing.T) {
		c := NewSet().NewInt("x")
		hammer(t, n, func(_ int) {
			nPrev := c.Get()
			for range inner {
				c.Inc()
				assert.Greater(t, c.Get(), nPrev)
			}
		})
		assert.Equal(t, c.Get(), n*inner)
	})

	t.Run("float", func(t *testing.T) {
		c := NewSet().NewFloat("x")
		hammer(t, n, func(_ int) {
			nPrev := c.Get()
			for range inner {
				c.Inc()
				assert.Greater(t, c.Get(), nPrev)
			}
		})
		assert.Equal(t, c.Get(), n*inner)
	})
}
