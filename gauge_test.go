package metrics

import (
	"sync"
	"testing"

	"go.withmatt.com/metrics/internal/assert"
)

func TestGaugeNew(t *testing.T) {
	NewSet().NewGauge("foo", nil)
	NewSet().NewGauge("foo", nil, "bar", "baz")

	// invalid label pairs
	assert.Panics(t, func() { NewSet().NewGauge("foo", nil, "bar") })

	// duplicate
	set := NewSet()
	set.NewGauge("foo", nil)
	assert.Panics(t, func() { set.NewGauge("foo", nil) })
}

func TestGaugeGetOrCreate(t *testing.T) {
	set := NewSet()
	set.GetOrCreateGauge("foo").Inc()
	set.GetOrCreateGauge("foo").Inc()
	assert.Equal(t, 2, set.GetOrCreateGauge("foo").Get())

	set.GetOrCreateGauge("foo", "a", "1").Inc()
	assert.Equal(t, 2, set.GetOrCreateGauge("foo").Get())
	assert.Equal(t, 1, set.GetOrCreateGauge("foo", "a", "1").Get())
}

func TestGaugeVec(t *testing.T) {
	set := NewSet()
	g := set.NewGaugeVec(GaugeVecOpt{
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

func TestGaugeSerial(t *testing.T) {
	set := NewSet()
	c := set.NewGauge("GaugeSerial", nil)
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

func TestGaugeCallback(t *testing.T) {
	set := NewSet()
	g := set.NewGauge("GaugeSerial", func() float64 {
		return 100
	})

	assert.Panics(t, func() { g.Inc() })
	assert.Panics(t, func() { g.Set(1) })
	assert.Panics(t, func() { g.Dec() })

	assertMarshal(t, set, []string{"GaugeSerial 100"})
}

func TestGaugeIncDec(t *testing.T) {
	s := NewSet()
	g := s.NewGauge("foo", nil)
	assert.Equal(t, g.Get(), 0)
	for i := 1; i <= 100; i++ {
		g.Inc()
		assert.Equal(t, g.Get(), float64(i))
	}
	for i := 99; i >= 0; i-- {
		g.Dec()
		assert.Equal(t, g.Get(), float64(i))
	}
}

func TestGaugeIncDecConcurrent(t *testing.T) {
	s := NewSet()
	g := s.NewGauge("foo", nil)

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

func TestGaugeConcurrent(t *testing.T) {
	const n = 1000
	const inner = 5

	var x int
	var nLock sync.Mutex
	g := NewSet().NewGauge("x", func() float64 {
		nLock.Lock()
		defer nLock.Unlock()
		x++
		return float64(x)
	})
	hammer(t, n, func(_ int) {
		prevX := g.Get()
		for range inner {
			assert.Greater(t, g.Get(), prevX)
		}
	})

	assert.Equal(t, g.Get(), (n*inner)+n+1)
}

func TestGaugeGetOrCreateConcurrent(t *testing.T) {
	const n = 1000
	const inner = 10

	set := NewSet()
	fn := func() *Gauge {
		return set.GetOrCreateGauge("x", "a", "1")
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
