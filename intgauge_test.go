package metrics

import (
	"sync"
	"testing"

	"go.withmatt.com/metrics/internal/assert"
)

func TestIntGaugeNew(t *testing.T) {
	NewSet().NewIntGauge("foo", nil)
	NewSet().NewIntGauge("foo", nil, "bar", "baz")

	// invalid label pairs
	assert.Panics(t, func() { NewSet().NewIntGauge("foo", nil, "bar") })

	// duplicate
	set := NewSet()
	set.NewIntGauge("foo", nil)
	assert.Panics(t, func() { set.NewIntGauge("foo", nil) })
}

func TestIntGaugeGetOrCreate(t *testing.T) {
	set := NewSet()
	set.GetOrCreateIntGauge("foo").Inc()
	set.GetOrCreateIntGauge("foo").Inc()
	assert.Equal(t, 2, set.GetOrCreateIntGauge("foo").Get())

	set.GetOrCreateIntGauge("foo", "a", "1").Inc()
	assert.Equal(t, 2, set.GetOrCreateIntGauge("foo").Get())
	assert.Equal(t, 1, set.GetOrCreateIntGauge("foo", "a", "1").Get())
}

func TestIntGaugeVec(t *testing.T) {
	set := NewSet()
	g := set.NewIntGaugeVec(IntGaugeVecOpt{
		Family: "foo",
		Labels: []string{"a", "b"},
	})
	g.WithLabelValues("1", "2").Set(5)
	g.WithLabelValues("1", "2").Set(1)
	g.WithLabelValues("3", "4").Set(2)

	assert.Equal(t, g.WithLabelValues("1", "2").Get(), 1)
	assert.Equal(t, g.WithLabelValues("3", "4").Get(), 2)

	// order is unpredictable bc the tags aren't ordered
	assertMarshalUnordered(t, set, []string{
		`foo{a="1",b="2"} 1`,
		`foo{a="3",b="4"} 2`,
	})
}

func TestIntGaugeSerial(t *testing.T) {
	set := NewSet()
	c := set.NewIntGauge("GaugeSerial", nil)
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

	assertMarshal(t, set, []string{"GaugeSerial 125"})

	// 	// Verify big numbers marshaling
	c.Set(1234567899)
	assertMarshal(t, set, []string{"GaugeSerial 1234567899"})
}

func TestIntGaugeCallback(t *testing.T) {
	set := NewSet()
	g := set.NewIntGauge("GaugeSerial", func() uint64 {
		return 100
	})

	assert.Panics(t, func() { g.Inc() })
	assert.Panics(t, func() { g.Set(1) })
	assert.Panics(t, func() { g.Dec() })

	assertMarshal(t, set, []string{"GaugeSerial 100"})
}

func TestIntGaugeIncDec(t *testing.T) {
	s := NewSet()
	g := s.NewIntGauge("foo", nil)
	assert.Equal(t, g.Get(), 0)
	for i := 1; i <= 100; i++ {
		g.Inc()
		assert.Equal(t, g.Get(), uint64(i))
	}
	for i := 99; i >= 0; i-- {
		g.Dec()
		assert.Equal(t, g.Get(), uint64(i))
	}
}

func TestIntGaugeIncDecConcurrent(t *testing.T) {
	s := NewSet()
	g := s.NewIntGauge("foo", nil)

	workers := 5
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			for range 100 {
				g.Inc()
				g.Dec()
			}
			wg.Done()
		}()
	}
	wg.Wait()

	assert.Equal(t, g.Get(), 0)
}

func TestIntGaugeConcurrent(t *testing.T) {
	const n = 1000
	const inner = 5

	var x int
	var nLock sync.Mutex
	g := NewSet().NewIntGauge("x", func() uint64 {
		nLock.Lock()
		defer nLock.Unlock()
		x++
		return uint64(x)
	})
	hammer(t, n, func(_ int) {
		prevX := g.Get()
		for range inner {
			assert.Greater(t, g.Get(), prevX)
		}
	})

	assert.Equal(t, g.Get(), (n*inner)+n+1)
}

func TestIntGaugeGetOrCreateConcurrent(t *testing.T) {
	const n = 1000
	const inner = 10

	set := NewSet()
	fn := func() *IntGauge {
		return set.GetOrCreateIntGauge("x", "a", "1")
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
