package metrics

import (
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"go.withmatt.com/metrics/internal/atomicx"
)

const (
	e10Min              = -9
	e10Max              = 18
	bucketsPerDecimal   = 18
	decimalBucketsCount = e10Max - e10Min

	// histBuckets is number of buckets within a single histogram
	histBuckets = decimalBucketsCount * bucketsPerDecimal

	// totalBuckets is histogram buckets + lower and upper buckets
	totalBuckets = histBuckets + 2

	// maxNumSeries is maximum possible series that can be emitted
	// by a single histogram, this includes all buckets and the
	// sum and count summaries
	maxNumSeries = totalBuckets + 2
)

var (
	bucketMultiplier = math.Pow(10, 1.0/bucketsPerDecimal)
	bucketRanges     [totalBuckets]string
)

type (
	decimalBucket       = [bucketsPerDecimal]atomic.Uint64
	atomicDecimalBucket = atomic.Pointer[decimalBucket]
)

// Histogram is a histogram for non-negative values with automatically created buckets.
//
// See https://medium.com/@valyala/improving-histogram-usability-for-prometheus-and-grafana-bc7e5df0e350
//
// Each bucket contains a counter for values in the given range.
// Each non-empty bucket is exposed via the following metric:
//
//	<metric_name>_bucket{<optional_tags>,vmrange="<start>...<end>"} <counter>
//
// Where:
//
//   - <metric_name> is the metric name passed to NewHistogram
//   - <optional_tags> is optional tags for the <metric_name>, which are passed to NewHistogram
//   - <start> and <end> - start and end values for the given bucket
//   - <counter> - the number of hits to the given bucket during Update* calls
//
// Histogram buckets can be converted to Prometheus-like buckets with `le` labels
// with `prometheus_buckets(<metric_name>_bucket)` function from PromQL extensions in VictoriaMetrics.
// (see https://docs.victoriametrics.com/metricsql/ ):
//
//	prometheus_buckets(request_duration_bucket)
//
// Time series produced by the Histogram have better compression ratio comparing to
// Prometheus histogram buckets with `le` labels, since they don't include counters
// for all the previous buckets.
//
// Zero histogram is usable.
type Histogram struct {
	// buckets contains counters for histogram buckets
	buckets [decimalBucketsCount]atomicDecimalBucket

	// lower is the number of values, which hit the lower bucket
	lower atomic.Uint64

	// upper is the number of values, which hit the upper bucket
	upper atomic.Uint64

	// sum is the sum of all the values put into Histogram
	sum atomicx.Float64
}

// Reset resets the given histogram.
func (h *Histogram) Reset() {
	clear(h.buckets[:])

	h.lower.Store(0)
	h.upper.Store(0)
	h.sum.Store(0)
}

// Update updates h with val.
//
// Negative values and NaNs are ignored.
func (h *Histogram) Update(val float64) {
	if math.IsNaN(val) || val < 0 {
		// Skip NaNs and negative values.
		return
	}

	bucketIdx := (math.Log10(val) - e10Min) * bucketsPerDecimal

	switch {
	case bucketIdx < 0:
		h.lower.Add(1)
		return
	case bucketIdx >= histBuckets:
		h.upper.Add(1)
	default:
		idx := uint(bucketIdx)
		if bucketIdx == float64(idx) && idx > 0 {
			// Edge case for 10^n values, which must go to the lower bucket
			// according to Prometheus logic for `le`-based histograms.
			idx--
		}
		decimalBucketIdx := idx / bucketsPerDecimal
		offset := idx % bucketsPerDecimal

		db := h.buckets[decimalBucketIdx].Load()
		if db == nil {
			// this bucket doesn't exist yet
			var dbNew decimalBucket
			if h.buckets[decimalBucketIdx].CompareAndSwap(db, &dbNew) {
				db = &dbNew
			} else {
				db = h.buckets[decimalBucketIdx].Load()
			}
		}
		db[offset].Add(1)
	}
	h.sum.Add(val)
}

// Merge merges src to h.
func (h *Histogram) Merge(src *Histogram) {
	h.lower.Add(src.lower.Load())
	h.upper.Add(src.upper.Load())
	h.sum.Add(src.sum.Load())

	for i := range src.buckets {
		if dbSrc := src.buckets[i].Load(); dbSrc != nil {
			dbDst := h.buckets[i].Load()
			if dbDst == nil {
				// this bucket doesn't exist yet
				var dbNew decimalBucket
				if h.buckets[i].CompareAndSwap(dbDst, &dbNew) {
					dbDst = &dbNew
				} else {
					dbDst = h.buckets[i].Load()
				}
			}
			for j := range dbSrc {
				dbDst[j].Add(dbSrc[j].Load())
			}
		}
	}
}

// UpdateDuration updates request duration based on the given startTime.
func (h *Histogram) UpdateDuration(startTime time.Time) {
	h.Update(time.Since(startTime).Seconds())
}

func (h *Histogram) marshalTo(w ExpfmtWriter, name MetricName) {
	card := punchCardPool.Get().(*punchCard)
	defer func() {
		clear(card[:])
		punchCardPool.Put(card)
	}()

	totalCounts, punches := h.punchBuckets(card)
	if totalCounts == 0 {
		return
	}

	sum := h.sum.Load()
	family := name.Family.String()

	// 1 extra because we're always adding in the vmrange tag
	// and sizeOfTags doesn't include a trailing comma
	tagsSize := sizeOfTags(name.Tags, w.constantTags) + 1

	const (
		chunkVMRange = `_bucket{vmrange="`
		chunkSum     = "_sum"
		chunkCount   = "_count"
	)

	// we need the underlying bytes.Buffer
	b := w.b

	// this is trying to compute up front how much we'll need to write
	// below with some margin of error to make sure we allocate enough
	// since we can't compute exactly
	b.Grow(
		(len(family) * punches) +
			(tagsSize * punches) +
			(len(chunkVMRange) * punches) +
			punches +
			len(family) + len(chunkSum) + tagsSize + 3 +
			len(family) + len(chunkCount) + tagsSize + 3 +
			64, // extra margin of error
	)

	// Write each `_bucket` count metric
	// This ultimately constructs a line such as:
	//   foo_bucket{vmrange="...",foo="bar"} 5
	for idx, count := range card {
		if count > 0 {
			vmrange := bucketRanges[idx]
			b.WriteString(family)
			b.WriteString(chunkVMRange)
			b.WriteString(vmrange)
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
			writeUint64(b, count)
			b.WriteByte('\n')
		}
	}

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
	writeUint64(b, totalCounts)
	b.WriteByte('\n')
}

// punchBuckets marks the counts on the punchCard corresponding to which
// histogram buckets have counts.
func (h *Histogram) punchBuckets(c *punchCard) (total uint64, punches int) {
	if count := h.lower.Load(); count > 0 {
		c[0] = count
		total += count
		punches++
	}

	if count := h.upper.Load(); count > 0 {
		c[len(c)-1] = count
		total += count
		punches++
	}

	for idx := range h.buckets {
		if db := h.buckets[idx].Load(); db != nil {
			for offset := range db {
				if count := db[offset].Load(); count > 0 {
					bucketIdx := idx*bucketsPerDecimal + offset
					c[bucketIdx+1] = count
					total += count
					punches++
				}
			}
		}
	}

	return
}

// punchCard is used internally to track counts per bucket when computing
// which histograms ranges have been hit.
type punchCard [totalBuckets]uint64

var punchCardPool = sync.Pool{
	New: func() any {
		var c punchCard
		return &c
	},
}

func init() {
	// pre-compute all bucket ranges
	v := math.Pow10(e10Min)
	bucketRanges[0] = "0..." + formatBucket(v)

	start := formatBucket(v)
	for i := range histBuckets {
		v *= bucketMultiplier
		end := formatBucket(v)
		bucketRanges[i+1] = start + "..." + end
		start = end
	}

	bucketRanges[totalBuckets-1] = formatBucket(math.Pow10(e10Max)) + "...+Inf"
}

func formatBucket(v float64) string {
	return strconv.FormatFloat(v, 'e', 3, 64)
}
