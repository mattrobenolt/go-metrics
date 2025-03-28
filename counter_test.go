package metrics

import (
	"testing"

	"go.withmatt.com/metrics/internal/assert"
)

func TestCounterNew(t *testing.T) {
	NewSet().NewCounter("foo")
	NewSet().NewCounter("foo", "bar", "baz")

	// invalid label pairs
	assert.Panics(t, func() { NewSet().NewCounter("foo", "bar") })

	// duplicate
	set := NewSet()
	set.NewCounter("foo")
	assert.Panics(t, func() { set.NewCounter("foo") })
}

func TestCounterVec(t *testing.T) {
	set := NewSet()
	c := set.NewCounterVec("foo", "a", "b")
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
}

func TestCounterSerial(t *testing.T) {
	const name = "CounterSerial"
	set := NewSet()
	c := set.NewCounter(name)
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

	assertMarshal(t, set, []string{"CounterSerial 125"})
}

func TestCounterConcurrent(t *testing.T) {
	const n = 1000
	const inner = 10

	c := NewSet().NewCounter("x")
	hammer(t, n, func(_ int) {
		nPrev := c.Get()
		for range inner {
			c.Inc()
			assert.Greater(t, c.Get(), nPrev)
		}
	})
	assert.Equal(t, c.Get(), n*inner)
}
