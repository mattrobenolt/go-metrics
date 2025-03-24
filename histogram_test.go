package metrics

import (
	"bytes"
	"math"
	"strings"
	"testing"
	"time"

	"go.withmatt.com/metrics/internal/assert"
)

func TestHistogramNew(t *testing.T) {
	NewSet().NewHistogram("foo")
	NewSet().NewHistogram("foo", "bar", "baz")

	// invalid label pairs
	assert.Panics(t, func() { NewSet().NewHistogram("foo", "bar") })

	// duplicate
	set := NewSet()
	set.NewHistogram("foo")
	assert.Panics(t, func() { set.NewHistogram("foo") })
}

func TestHistogramGetOrCreate(t *testing.T) {
	set := NewSet()
	fn := set.GetOrCreateHistogram

	fn("foo").Update(1)
	fn("foo").Update(2)

	h := fn("foo")
	assert.Equal(t, 3, h.sum())

	fn("foo", "a", "1").Update(1)
	assert.Equal(t, 3, fn("foo").sum())
	assert.Equal(t, 1, fn("foo", "a", "1").sum())
}

func TestHistogramVec(t *testing.T) {
	set := NewSet()
	h := set.NewHistogramVec(HistogramVecOpt{
		Family: "foo",
		Labels: []string{"a", "b"},
	})
	h.WithLabelValues("1", "2").Update(1)
	h.WithLabelValues("1", "2").Update(2)
	h.WithLabelValues("3", "4").Update(1)

	// order is unpredictable bc the tags aren't ordered
	assertMarshalUnordered(t, set, []string{
		`foo_bucket{vmrange="8.799e-01...1.000e+00",a="1",b="2"} 1`,
		`foo_bucket{vmrange="1.896e+00...2.154e+00",a="1",b="2"} 1`,
		`foo_sum{a="1",b="2"} 3`,
		`foo_count{a="1",b="2"} 2`,
		`foo_bucket{vmrange="8.799e-01...1.000e+00",a="3",b="4"} 1`,
		`foo_sum{a="3",b="4"} 1`,
		`foo_count{a="3",b="4"} 1`,
	})
}

func TestHistogramSerial(t *testing.T) {
	set := NewSet()
	h := set.NewHistogram("hist")

	// Verify that the histogram is invisible in the output of WritePrometheus when it has no data.
	assertMarshal(t, set, nil)
	for i := 98; i < 218; i++ {
		h.Update(float64(i))
	}

	assertMarshal(t, set, []string{
		`hist_bucket{vmrange="8.799e+01...1.000e+02"} 3`,
		`hist_bucket{vmrange="1.000e+02...1.136e+02"} 13`,
		`hist_bucket{vmrange="1.136e+02...1.292e+02"} 16`,
		`hist_bucket{vmrange="1.292e+02...1.468e+02"} 17`,
		`hist_bucket{vmrange="1.468e+02...1.668e+02"} 20`,
		`hist_bucket{vmrange="1.668e+02...1.896e+02"} 23`,
		`hist_bucket{vmrange="1.896e+02...2.154e+02"} 26`,
		`hist_bucket{vmrange="2.154e+02...2.448e+02"} 2`,
		`hist_sum 18900`,
		`hist_count 120`,
	})

	set = NewSet()
	h = set.NewHistogram("hist", "foo", "bar")

	// Verify that the histogram is invisible in the output of WritePrometheus when it has no data.
	assertMarshal(t, set, nil)
	for i := 98; i < 218; i++ {
		h.Update(float64(i))
	}

	assertMarshal(t, set, []string{
		`hist_bucket{vmrange="8.799e+01...1.000e+02",foo="bar"} 3`,
		`hist_bucket{vmrange="1.000e+02...1.136e+02",foo="bar"} 13`,
		`hist_bucket{vmrange="1.136e+02...1.292e+02",foo="bar"} 16`,
		`hist_bucket{vmrange="1.292e+02...1.468e+02",foo="bar"} 17`,
		`hist_bucket{vmrange="1.468e+02...1.668e+02",foo="bar"} 20`,
		`hist_bucket{vmrange="1.668e+02...1.896e+02",foo="bar"} 23`,
		`hist_bucket{vmrange="1.896e+02...2.154e+02",foo="bar"} 26`,
		`hist_bucket{vmrange="2.154e+02...2.448e+02",foo="bar"} 2`,
		`hist_sum{foo="bar"} 18900`,
		`hist_count{foo="bar"} 120`,
	})

	// Verify Reset
	h.Reset()
	assertMarshal(t, set, nil)

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

	// manually just fill every bucket
	for i := range h.buckets {
		if h.buckets[i].Load() == nil {
			var db decimalBucket
			h.buckets[i].Store(&db)
		}
		db := h.buckets[i].Load()
		for j := range db {
			if db[j].Load() == 0 {
				db[j].Add(1)
			}
		}
	}

	var b bytes.Buffer
	set.WritePrometheusUnthrottled(&b)
	lines := strings.Split(strings.Trim(b.String(), "\n"), "\n")
	assert.Equal(t, len(lines), maxNumSeries)
	for i, vmrange := range bucketRanges {
		assert.True(t, strings.Contains(lines[i], vmrange))
	}
	assert.True(t, strings.HasPrefix(lines[totalBuckets], `hist_sum{foo="bar"} `))
	assert.True(t, strings.HasPrefix(lines[totalBuckets+1], `hist_count{foo="bar"} `))
}

func TestHistogramConcurrent(t *testing.T) {
	const n = 5

	set := NewSet()
	h := set.NewHistogram("x")
	hammer(t, n, func(_ int) {
		for f := 0.6; f < 1.4; f += 0.1 {
			h.Update(f)
		}
	})

	assertMarshal(t, set, []string{
		`x_bucket{vmrange="5.995e-01...6.813e-01"} 5`,
		`x_bucket{vmrange="6.813e-01...7.743e-01"} 5`,
		`x_bucket{vmrange="7.743e-01...8.799e-01"} 5`,
		`x_bucket{vmrange="8.799e-01...1.000e+00"} 10`,
		`x_bucket{vmrange="1.000e+00...1.136e+00"} 5`,
		`x_bucket{vmrange="1.136e+00...1.292e+00"} 5`,
		`x_bucket{vmrange="1.292e+00...1.468e+00"} 5`,
		`x_sum 38`,
		`x_count 40`,
	})
}

func TestHistogramGetOrCreateConcurrent(t *testing.T) {
	const n = 5

	set := NewSet()
	fn := func() *Histogram {
		return set.GetOrCreateHistogram("x", "a", "1")
	}
	hammer(t, n, func(_ int) {
		for f := 0.6; f < 1.4; f += 0.1 {
			fn().Update(f)
		}
	})

	assertMarshal(t, set, []string{
		`x_bucket{vmrange="5.995e-01...6.813e-01",a="1"} 5`,
		`x_bucket{vmrange="6.813e-01...7.743e-01",a="1"} 5`,
		`x_bucket{vmrange="7.743e-01...8.799e-01",a="1"} 5`,
		`x_bucket{vmrange="8.799e-01...1.000e+00",a="1"} 10`,
		`x_bucket{vmrange="1.000e+00...1.136e+00",a="1"} 5`,
		`x_bucket{vmrange="1.136e+00...1.292e+00",a="1"} 5`,
		`x_bucket{vmrange="1.292e+00...1.468e+00",a="1"} 5`,
		`x_sum{a="1"} 38`,
		`x_count{a="1"} 40`,
	})
}

func TestHistogramMerge(t *testing.T) {
	set := NewSet()
	h := set.NewHistogram("foo")
	// Write data to histogram
	for i := 10; i < 100; i++ {
		h.Update(float64(i))
	}

	var v Histogram
	for i := 1; i < 300; i++ {
		v.Update(float64(i))
	}

	h.Merge(&v)

	assertMarshal(t, set, []string{
		`foo_bucket{vmrange="8.799e-01...1.000e+00"} 1`,
		`foo_bucket{vmrange="1.896e+00...2.154e+00"} 1`,
		`foo_bucket{vmrange="2.783e+00...3.162e+00"} 1`,
		`foo_bucket{vmrange="3.594e+00...4.084e+00"} 1`,
		`foo_bucket{vmrange="4.642e+00...5.275e+00"} 1`,
		`foo_bucket{vmrange="5.995e+00...6.813e+00"} 1`,
		`foo_bucket{vmrange="6.813e+00...7.743e+00"} 1`,
		`foo_bucket{vmrange="7.743e+00...8.799e+00"} 1`,
		`foo_bucket{vmrange="8.799e+00...1.000e+01"} 3`,
		`foo_bucket{vmrange="1.000e+01...1.136e+01"} 2`,
		`foo_bucket{vmrange="1.136e+01...1.292e+01"} 2`,
		`foo_bucket{vmrange="1.292e+01...1.468e+01"} 4`,
		`foo_bucket{vmrange="1.468e+01...1.668e+01"} 4`,
		`foo_bucket{vmrange="1.668e+01...1.896e+01"} 4`,
		`foo_bucket{vmrange="1.896e+01...2.154e+01"} 6`,
		`foo_bucket{vmrange="2.154e+01...2.448e+01"} 6`,
		`foo_bucket{vmrange="2.448e+01...2.783e+01"} 6`,
		`foo_bucket{vmrange="2.783e+01...3.162e+01"} 8`,
		`foo_bucket{vmrange="3.162e+01...3.594e+01"} 8`,
		`foo_bucket{vmrange="3.594e+01...4.084e+01"} 10`,
		`foo_bucket{vmrange="4.084e+01...4.642e+01"} 12`,
		`foo_bucket{vmrange="4.642e+01...5.275e+01"} 12`,
		`foo_bucket{vmrange="5.275e+01...5.995e+01"} 14`,
		`foo_bucket{vmrange="5.995e+01...6.813e+01"} 18`,
		`foo_bucket{vmrange="6.813e+01...7.743e+01"} 18`,
		`foo_bucket{vmrange="7.743e+01...8.799e+01"} 20`,
		`foo_bucket{vmrange="8.799e+01...1.000e+02"} 25`,
		`foo_bucket{vmrange="1.000e+02...1.136e+02"} 13`,
		`foo_bucket{vmrange="1.136e+02...1.292e+02"} 16`,
		`foo_bucket{vmrange="1.292e+02...1.468e+02"} 17`,
		`foo_bucket{vmrange="1.468e+02...1.668e+02"} 20`,
		`foo_bucket{vmrange="1.668e+02...1.896e+02"} 23`,
		`foo_bucket{vmrange="1.896e+02...2.154e+02"} 26`,
		`foo_bucket{vmrange="2.154e+02...2.448e+02"} 29`,
		`foo_bucket{vmrange="2.448e+02...2.783e+02"} 34`,
		`foo_bucket{vmrange="2.783e+02...3.162e+02"} 21`,
		`foo_sum 49755`,
		`foo_count 389`,
	})
}
