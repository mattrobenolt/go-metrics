package metrics

import (
	"bytes"
	"context"
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
	assert.Equal(t, 3, h.sum.Load())

	fn("foo", "a", "1").Update(1)
	assert.Equal(t, 3, fn("foo").sum.Load())
	assert.Equal(t, 1, fn("foo", "a", "1").sum.Load())
}

func TestHistogramVec(t *testing.T) {
	t.Skip(context.TODO())
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
	t.Skip(context.TODO())
}

func TestHistogramGetOrCreateConcurrent(t *testing.T) {
	t.Skip(context.TODO())
}
