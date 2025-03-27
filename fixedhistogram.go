package metrics

import (
	"math"
	"slices"
	"strconv"
	"sync/atomic"
	"time"

	"go.withmatt.com/metrics/internal/atomicx"
)

// DefBuckets is the default set of buckets used with a [FixedHistogram].
var DefBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

// FixedHistogram is a Prometheus-like histogram with fixed buckets.
//
// If you would like VictoriaMetrics `vmrange` histogram buckets, see [Histogram].
type FixedHistogram struct {
	buckets      []float64
	labels       []string
	observations []atomic.Uint64

	upper    atomic.Uint64
	sumInt   atomic.Int64
	sumFloat atomicx.Float64
	count    atomic.Uint64
}

func newFixedHistogram(buckets []float64) *FixedHistogram {
	if len(buckets) == 0 {
		buckets = slices.Clone(DefBuckets)
	} else {
		buckets = slices.Clone(buckets)
		slices.Sort(buckets)
	}

	return &FixedHistogram{
		buckets:      buckets,
		labels:       labelsForBuckets(buckets),
		observations: make([]atomic.Uint64, len(buckets)),
	}
}

// Reset resets the given histogram.
func (h *FixedHistogram) Reset() {
	clear(h.observations)
	h.upper.Store(0)
	h.count.Store(0)
	h.sumInt.Store(0)
	h.sumFloat.Store(0)
}

// Update updates h with val.
//
// NaNs are ignored.
func (h *FixedHistogram) Update(val float64) {
	if math.IsNaN(val) {
		// Skip NaNs.
		return
	}

	n := h.findBucket(val)
	for ; n < len(h.buckets); n++ {
		h.observations[n].Add(1)
	}
	h.upper.Add(1)

	if val != 0 {
		if intval := int64(val); float64(intval) == val {
			h.sumInt.Add(intval)
		} else {
			h.sumFloat.Add(val)
		}
	}

	h.count.Add(1)
}

// Observe updates h with val, identical to [FixedHistogram.Update].
//
// Negative values and NaNs are ignored.
func (h *FixedHistogram) Observe(val float64) {
	h.Update(val)
}

// UpdateDuration updates request duration based on the given startTime.
func (h *FixedHistogram) UpdateDuration(startTime time.Time) {
	h.Update(time.Since(startTime).Seconds())
}

func (h *FixedHistogram) findBucket(v float64) int {
	n := len(h.buckets)
	switch {
	case n == 0:
		return 0
	case v < h.buckets[0]:
		return 0
	case v > h.buckets[n-1]:
		return n
	case n < 35:
		// For small arrays, use simple linear search
		// "magic number" 35 is result of tests on couple different (AWS and bare metal) servers
		// see more details here: https://github.com/prometheus/client_golang/pull/1662
		for i, bound := range h.buckets {
			if v <= bound {
				return i
			}
		}
		// If v is greater than all upper bounds, return len(h.upperBounds)
		return n
	}

	// For larger arrays, use stdlib's binary search
	i, _ := slices.BinarySearch(h.buckets, v)
	return i
}

func (h *FixedHistogram) marshalTo(w ExpfmtWriter, name MetricName) {
	sum := h.sum()
	count := h.count.Load()
	upper := h.upper.Load()
	family := name.Family.String()

	// 1 extra because we're always adding in the vmrange tag
	// and sizeOfTags doesn't include a trailing comma
	tagsSize := sizeOfTags(name.Tags, w.constantTags) + 1

	const (
		chunkLe    = `_bucket{le="`
		chunkUpper = `_bucket{le="+Inf"`
		chunkSum   = "_sum"
		chunkCount = "_count"
	)

	// we need the underlying bytes.Buffer
	b := w.b

	b.Grow(
		(len(family) * len(h.buckets)) +
			(tagsSize * len(h.buckets)) +
			(len(chunkLe) * len(h.buckets)) +
			len(family) + len(chunkUpper) + tagsSize + 3 +
			len(family) + len(chunkSum) + tagsSize + 3 +
			len(family) + len(chunkCount) + tagsSize + 3 +
			64, // extra margin of error
	)

	for i := range h.buckets {
		b.WriteString(family)
		b.WriteString(chunkLe)
		b.WriteString(h.labels[i])
		b.WriteByte('"')
		if len(w.constantTags) > 0 {
			b.WriteByte(',')
			b.WriteString(w.constantTags)
		}
		for _, tag := range name.Tags {
			b.WriteByte(',')
			writeTag(b, tag)
		}
		b.WriteString(`} `)
		writeUint64(b, h.observations[i].Load())
		b.WriteByte('\n')
	}

	// write the upper bucket +Inf
	b.WriteString(family)
	b.WriteString(chunkUpper)
	if len(w.constantTags) > 0 {
		b.WriteByte(',')
		b.WriteString(w.constantTags)
	}
	for _, tag := range name.Tags {
		b.WriteByte(',')
		writeTag(b, tag)
	}
	b.WriteString(`} `)
	writeUint64(b, upper)
	b.WriteByte('\n')

	// Write our `_sum` line
	// This ultimately constructs a line such as:
	//   foo_sum{foo="bar"} 5
	b.WriteString(family)
	b.WriteString(chunkSum)
	if tagsSize > 0 {
		b.WriteByte('{')
		writeTags(b, w.constantTags, name.Tags)
		b.WriteByte('}')
	}
	b.WriteByte(' ')
	writeFloat64(b, sum)
	b.WriteByte('\n')

	// Write our `_count` line
	// This ultimately constructs a line such as:
	//   foo_count{foo="bar"} 5
	b.WriteString(family)
	b.WriteString(chunkCount)
	if tagsSize > 0 {
		b.WriteByte('{')
		writeTags(b, w.constantTags, name.Tags)
		b.WriteByte('}')
	}
	b.WriteByte(' ')
	writeUint64(b, count)
	b.WriteByte('\n')
}

func (h *FixedHistogram) sum() float64 {
	return float64(h.sumInt.Load()) + h.sumFloat.Load()
}

func labelsForBuckets(buckets []float64) []string {
	labels := make([]string, len(buckets))
	for i, v := range buckets {
		labels[i] = strconv.FormatFloat(v, 'f', -1, 64)
	}
	return labels
}
