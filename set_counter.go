package metrics

import (
	"errors"
)

// NewCounter registers and returns new Counter with the given name in the s.
//
// family must be a Prometheus compatible identifier format.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	NewCounter("family", "label1", "value1", "label2", "value2")
//
// The returned Counter is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *Set) NewCounter(family string, tags ...string) *Counter {
	c := &Counter{}
	s.mustRegisterMetric(c, MetricName{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
	})
	return c
}

// A CounterVec is a collection of Counters that are partitioned
// by the same metric name and tag labels, but different tag values.
type CounterVec struct {
	commonVec
}

// WithLabelValues returns the Counter for the corresponding label values.
// If the combination of values is seen for the first time, a new Counter
// is created.
//
// This will panic if the values count doesn't match the number of labels.
func (c *CounterVec) WithLabelValues(values ...string) *Counter {
	if len(values) != len(c.partialTags) {
		panic(errors.New("mismatch length of labels"))
	}
	hash := hashFinish(c.partialHash, values)

	c.s.metricsMu.Lock()
	nm := c.s.metrics[hash]
	c.s.metricsMu.Unlock()

	if nm == nil {
		nm = c.s.getOrRegisterMetricFromVec(
			&Counter{}, hash, c.family, c.partialTags, values,
		)
	}
	return nm.metric.(*Counter)
}

// NewCounterVec creates a new [CounterVec] with the supplied name.
func (s *Set) NewCounterVec(family string, labels ...string) *CounterVec {
	return &CounterVec{commonVec{
		s:           s,
		family:      MustIdent(family),
		partialTags: makePartialTags(labels),
		partialHash: hashStart(family, labels),
	}}
}
