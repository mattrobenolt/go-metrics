package metrics

import (
	"testing"

	"go.withmatt.com/metrics/internal/assert"
)

func TestFloatCounterNew(t *testing.T) {
	NewSet().NewFloatCounter("foo")
	NewSet().NewFloatCounter("foo", "bar", "baz")

	// invalid label pairs
	assert.Panics(t, func() { NewSet().NewFloatCounter("foo", "bar") })

	// duplicate
	set := NewSet()
	set.NewFloatCounter("foo")
	assert.Panics(t, func() { set.NewFloatCounter("foo") })
}

func TestFloatCounterGetOrCreate(t *testing.T) {
	set := NewSet()
	set.GetOrCreateFloatCounter("foo").Inc()
	set.GetOrCreateFloatCounter("foo").Inc()
	assert.Equal(t, 2, set.GetOrCreateFloatCounter("foo").Get())

	set.GetOrCreateFloatCounter("foo", "a", "1").Inc()
	assert.Equal(t, 2, set.GetOrCreateFloatCounter("foo").Get())
	assert.Equal(t, 1, set.GetOrCreateFloatCounter("foo", "a", "1").Get())
}

func TestFloatCounterVec(t *testing.T) {
	set := NewSet()
	c := set.NewFloatCounterVec(VecName{
		Family: "foo",
		Labels: []string{"a", "b"},
	})
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

func TestFloatCounterSerial(t *testing.T) {
	const name = "CounterSerial"
	set := NewSet()
	c := set.NewFloatCounter(name)
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

func TestFloatCounterConcurrent(t *testing.T) {
	const n = 1000
	const inner = 10

	c := NewSet().NewFloatCounter("x")
	hammer(t, n, func(_ int) {
		nPrev := c.Get()
		for range inner {
			c.Inc()
			assert.Greater(t, c.Get(), nPrev)
		}
	})
	assert.Equal(t, c.Get(), n*inner)
}

func TestFloatCounterGetOrCreateConcurrent(t *testing.T) {
	const n = 1000
	const inner = 10

	set := NewSet()
	fn := func() *FloatCounter {
		return set.GetOrCreateFloatCounter("x", "a", "1")
	}
	hammer(t, n, func(_ int) {
		nPrev := fn().Get()
		for range inner {
			fn().Inc()
			assert.Greater(t, fn().Get(), nPrev)
		}
	})
	assert.Equal(t, fn().Get(), n*inner)
}
