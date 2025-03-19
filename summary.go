package metrics

import (
	"fmt"
	"io"
	"math"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/VictoriaMetrics/metrics/internal/histogram"
)

const defaultSummaryWindow = 5 * time.Minute

var defaultSummaryQuantiles = []float64{0.5, 0.9, 0.97, 0.99, 1}

// Summary implements summary.
type Summary struct {
	mu sync.Mutex

	curr *histogram.Fast
	next *histogram.Fast

	quantiles      []float64
	quantileValues []float64

	sum   float64
	count uint64

	window time.Duration
}

func newSummary(window time.Duration, quantiles []float64) *Summary {
	// Make a copy of quantiles in order to prevent from their modification by the caller.
	quantiles = slices.Clone(quantiles)
	validateQuantiles(quantiles)
	sm := &Summary{
		curr:           histogram.NewFast(),
		next:           histogram.NewFast(),
		quantiles:      quantiles,
		quantileValues: make([]float64, len(quantiles)),
		window:         window,
	}
	return sm
}

func validateQuantiles(quantiles []float64) {
	for _, q := range quantiles {
		if q < 0 || q > 1 {
			panic(fmt.Errorf("BUG: quantile must be in the range [0..1]; got %v", q))
		}
	}
}

// Update updates the summary.
func (sm *Summary) Update(v float64) {
	sm.mu.Lock()
	sm.curr.Update(v)
	sm.next.Update(v)
	sm.sum += v
	sm.count++
	sm.mu.Unlock()
}

// UpdateDuration updates request duration based on the given startTime.
func (sm *Summary) UpdateDuration(startTime time.Time) {
	d := time.Since(startTime).Seconds()
	sm.Update(d)
}

func (sm *Summary) marshalTo(prefix string, w io.Writer) {
	// Marshal only *_sum and *_count values.
	// Quantile values should be already updated by the caller via sm.updateQuantiles() call.
	// sm.quantileValues will be marshaled later via quantileValue.marshalTo.
	sm.mu.Lock()
	sum := sm.sum
	count := sm.count
	sm.mu.Unlock()

	if count > 0 {
		name, filters := splitMetricName(prefix)
		if isFloatInteger(sum) {
			// Marshal integer sum without scientific notation
			fmt.Fprintf(w, "%s_sum%s %d\n", name, filters, int64(sum))
		} else {
			fmt.Fprintf(w, "%s_sum%s %g\n", name, filters, sum)
		}
		fmt.Fprintf(w, "%s_count%s %d\n", name, filters, count)
	}
}

func splitMetricName(name string) (string, string) {
	n := strings.IndexByte(name, '{')
	if n < 0 {
		return name, ""
	}
	return name[:n], name[n:]
}

func (sm *Summary) updateQuantiles() {
	sm.mu.Lock()
	sm.quantileValues = sm.curr.Quantiles(sm.quantileValues[:0], sm.quantiles)
	sm.mu.Unlock()
}

type quantileValue struct {
	sm  *Summary
	idx int
}

func (qv *quantileValue) marshalTo(prefix string, w io.Writer) {
	qv.sm.mu.Lock()
	v := qv.sm.quantileValues[qv.idx]
	qv.sm.mu.Unlock()
	if !math.IsNaN(v) {
		WriteMetricFloat64(w, prefix, v)
	}
}

func addTag(name, tag string) string {
	if len(name) == 0 || name[len(name)-1] != '}' {
		return name + "{" + tag + "}"
	}
	name = name[:len(name)-1]
	if len(name) == 0 {
		panic(fmt.Errorf("BUG: metric name cannot be empty"))
	}
	if name[len(name)-1] == '{' {
		// case for empty labels set metric_name{}
		return name + tag + "}"
	}
	return name + "," + tag + "}"
}

func registerSummaryLocked(sm *Summary) {
	window := sm.window
	summariesLock.Lock()
	summaries[window] = append(summaries[window], sm)
	if len(summaries[window]) == 1 {
		go summariesSwapCron(window)
	}
	summariesLock.Unlock()
}

func unregisterSummary(sm *Summary) {
	window := sm.window
	summariesLock.Lock()
	sms := summaries[window]
	found := false
	for i, xsm := range sms {
		if xsm == sm {
			sms = slices.Delete(sms, i, i+1)
			found = true
			break
		}
	}
	if !found {
		panic(fmt.Errorf("BUG: cannot find registered summary %p", sm))
	}
	summaries[window] = sms
	summariesLock.Unlock()
}

func summariesSwapCron(window time.Duration) {
	for {
		time.Sleep(window / 2)
		summariesLock.Lock()
		for _, sm := range summaries[window] {
			sm.mu.Lock()
			tmp := sm.curr
			sm.curr = sm.next
			sm.next = tmp
			sm.next.Reset()
			sm.mu.Unlock()
		}
		summariesLock.Unlock()
	}
}

var (
	summaries     = map[time.Duration][]*Summary{}
	summariesLock sync.Mutex
)
