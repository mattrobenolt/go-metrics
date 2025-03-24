package metrics

import (
	"math"
	"testing"
	"time"

	"go.withmatt.com/metrics/internal/assert"
)

func TestFixedHistogramNew(t *testing.T) {
	NewSet().NewFixedHistogram("foo", nil)
	NewSet().NewFixedHistogram("foo", nil, "bar", "baz")

	// invalid label pairs
	assert.Panics(t, func() { NewSet().NewFixedHistogram("foo", nil, "bar") })

	// duplicate
	set := NewSet()
	set.NewHistogram("foo")
	assert.Panics(t, func() { set.NewFixedHistogram("foo", nil) })
}

func TestFixedHistogramGetOrCreate(t *testing.T) {
	set := NewSet()
	fn := set.GetOrCreateFixedHistogram

	fn("foo", nil).Update(1)
	fn("foo", nil).Update(2)

	h := fn("foo", nil)
	assert.Equal(t, 3, h.sum())

	fn("foo", nil, "a", "1").Update(1)
	assert.Equal(t, 3, fn("foo", nil).sum())
	assert.Equal(t, 1, fn("foo", nil, "a", "1").sum())
}

func TestFixedHistogramVec(t *testing.T) {
	set := NewSet()
	h := set.NewFixedHistogramVec(FixedHistogramVecOpt{
		Family: "foo",
		Labels: []string{"a", "b"},
	})
	h.WithLabelValues("1", "2").Update(1)
	h.WithLabelValues("1", "2").Update(2)
	h.WithLabelValues("3", "4").Update(1)

	// order is unpredictable bc the tags aren't ordered
	assertMarshalUnordered(t, set, []string{
		`foo_bucket{le="0.005",a="1",b="2"} 0`,
		`foo_bucket{le="0.01",a="1",b="2"} 0`,
		`foo_bucket{le="0.025",a="1",b="2"} 0`,
		`foo_bucket{le="0.05",a="1",b="2"} 0`,
		`foo_bucket{le="0.1",a="1",b="2"} 0`,
		`foo_bucket{le="0.25",a="1",b="2"} 0`,
		`foo_bucket{le="0.5",a="1",b="2"} 0`,
		`foo_bucket{le="1",a="1",b="2"} 1`,
		`foo_bucket{le="2.5",a="1",b="2"} 2`,
		`foo_bucket{le="5",a="1",b="2"} 2`,
		`foo_bucket{le="10",a="1",b="2"} 2`,
		`foo_bucket{le="+Inf",a="1",b="2"} 2`,
		`foo_count{a="1",b="2"} 2`,
		`foo_sum{a="1",b="2"} 3`,
		`foo_bucket{le="0.005",a="3",b="4"} 0`,
		`foo_bucket{le="0.01",a="3",b="4"} 0`,
		`foo_bucket{le="0.025",a="3",b="4"} 0`,
		`foo_bucket{le="0.05",a="3",b="4"} 0`,
		`foo_bucket{le="0.1",a="3",b="4"} 0`,
		`foo_bucket{le="0.25",a="3",b="4"} 0`,
		`foo_bucket{le="0.5",a="3",b="4"} 0`,
		`foo_bucket{le="1",a="3",b="4"} 1`,
		`foo_bucket{le="2.5",a="3",b="4"} 1`,
		`foo_bucket{le="5",a="3",b="4"} 1`,
		`foo_bucket{le="10",a="3",b="4"} 1`,
		`foo_bucket{le="+Inf",a="3",b="4"} 1`,
		`foo_count{a="3",b="4"} 1`,
		`foo_sum{a="3",b="4"} 1`,
	})
}

func TestFixedHistogramSerial(t *testing.T) {
	set := NewSet()
	h := set.NewFixedHistogram("hist", []float64{
		100, 110, 120, 150, 200,
	})

	assertMarshal(t, set, []string{
		`hist_bucket{le="100"} 0`,
		`hist_bucket{le="110"} 0`,
		`hist_bucket{le="120"} 0`,
		`hist_bucket{le="150"} 0`,
		`hist_bucket{le="200"} 0`,
		`hist_bucket{le="+Inf"} 0`,
		`hist_sum 0`,
		`hist_count 0`,
	})

	for i := 98; i < 218; i++ {
		h.Update(float64(i))
	}

	assertMarshal(t, set, []string{
		`hist_bucket{le="100"} 3`,
		`hist_bucket{le="110"} 13`,
		`hist_bucket{le="120"} 23`,
		`hist_bucket{le="150"} 53`,
		`hist_bucket{le="200"} 103`,
		`hist_bucket{le="+Inf"} 120`,
		`hist_sum 18900`,
		`hist_count 120`,
	})

	h.Reset()
	assertMarshal(t, set, []string{
		`hist_bucket{le="100"} 0`,
		`hist_bucket{le="110"} 0`,
		`hist_bucket{le="120"} 0`,
		`hist_bucket{le="150"} 0`,
		`hist_bucket{le="200"} 0`,
		`hist_bucket{le="+Inf"} 0`,
		`hist_sum 0`,
		`hist_count 0`,
	})

	// Verify supported ranges
	for e10 := -100; e10 < 100; e10++ {
		for offset := range bucketsPerDecimal {
			m := 1 + math.Pow(bucketMultiplier, float64(offset))
			f1 := m * math.Pow10(e10)
			h.Update(f1)
			f2 := (m + 0.5*bucketMultiplier) * math.Pow10(e10)
			h.Update(f2)
			f3 := (m + 2*bucketMultiplier) * math.Pow10(e10)
			h.Update(f3)
		}
	}
	h.UpdateDuration(time.Now().Add(-time.Minute))

	// Verify edge cases
	h.Update(0)
	h.Update(math.Inf(1))
	h.Update(math.Inf(-1))
	h.Update(math.NaN())
	h.Update(-123)
	// See https://github.com/VictoriaMetrics/VictoriaMetrics/issues/1096
	h.Update(math.Float64frombits(0x3e112e0be826d695))

	assertMarshal(t, set, []string{
		`hist_bucket{le="100"} 5509`,
		`hist_bucket{le="110"} 5511`,
		`hist_bucket{le="120"} 5512`,
		`hist_bucket{le="150"} 5513`,
		`hist_bucket{le="200"} 5514`,
		`hist_bucket{le="+Inf"} 10807`,
		`hist_sum +Inf`,
		`hist_count 10807`,
	})
}

func TestFixedHistogramConcurrent(t *testing.T) {
	const n = 5

	set := NewSet()
	h := set.NewFixedHistogram("x", nil)
	hammer(t, n, func(_ int) {
		for f := 0.6; f < 1.4; f += 0.1 {
			h.Update(f)
		}
	})

	assertMarshal(t, set, []string{
		`x_bucket{le="0.005"} 0`,
		`x_bucket{le="0.01"} 0`,
		`x_bucket{le="0.025"} 0`,
		`x_bucket{le="0.05"} 0`,
		`x_bucket{le="0.1"} 0`,
		`x_bucket{le="0.25"} 0`,
		`x_bucket{le="0.5"} 0`,
		`x_bucket{le="1"} 25`,
		`x_bucket{le="2.5"} 40`,
		`x_bucket{le="5"} 40`,
		`x_bucket{le="10"} 40`,
		`x_bucket{le="+Inf"} 40`,
		`x_sum 38`,
		`x_count 40`,
	})
}

func TestFixedHistogramGetOrCreateConcurrent(t *testing.T) {
	const n = 5

	set := NewSet()
	fn := func() *FixedHistogram {
		return set.GetOrCreateFixedHistogram("x", nil, "a", "1")
	}
	hammer(t, n, func(_ int) {
		for f := 0.6; f < 1.4; f += 0.1 {
			fn().Update(f)
		}
	})

	assertMarshal(t, set, []string{
		`x_bucket{le="0.005",a="1"} 0`,
		`x_bucket{le="0.01",a="1"} 0`,
		`x_bucket{le="0.025",a="1"} 0`,
		`x_bucket{le="0.05",a="1"} 0`,
		`x_bucket{le="0.1",a="1"} 0`,
		`x_bucket{le="0.25",a="1"} 0`,
		`x_bucket{le="0.5",a="1"} 0`,
		`x_bucket{le="1",a="1"} 25`,
		`x_bucket{le="2.5",a="1"} 40`,
		`x_bucket{le="5",a="1"} 40`,
		`x_bucket{le="10",a="1"} 40`,
		`x_bucket{le="+Inf",a="1"} 40`,
		`x_sum{a="1"} 38`,
		`x_count{a="1"} 40`,
	})
}
