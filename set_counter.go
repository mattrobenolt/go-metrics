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
	return s.NewCounterOpt(MetricName{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
	})
}

// NewCounterOpt registers and returns new Counter with the name in the s.
//
// The returned Counter is safe to use from concurrent goroutines.
//
// This will panic if already registered.
func (s *Set) NewCounterOpt(name MetricName) *Counter {
	c := &Counter{}
	s.mustRegisterMetric(c, name)
	return c
}

// GetOrCreateCounter returns registered Counter in s with the given name
// and tags creates new Counter if s doesn't contain Counter with the given name.
//
// family must be a Prometheus compatible identifier format.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	GetOrCreateCounter("family", "label1", "value1", "label2", "value2")
//
// The returned Counter is safe to use from concurrent goroutines.
//
// Prefer [NewCounter] or [NewCounterOpt] when performance is critical.
//
// This will panic if values are invalid.
func (s *Set) GetOrCreateCounter(family string, tags ...string) *Counter {
	hash := getHashStrings(family, tags)

	s.metricsMu.Lock()
	nm := s.metrics[hash]
	s.metricsMu.Unlock()

	if nm == nil {
		nm = s.getOrAddMetricFromStrings(&Counter{}, hash, family, tags)
	}
	return nm.metric.(*Counter)
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
func (s *Set) NewCounterVec(name VecName) *CounterVec {
	family := MustIdent(name.Family)

	return &CounterVec{commonVec{
		s:           s,
		family:      family,
		partialTags: makePartialTags(name.Labels),
		partialHash: hashStart(family.String(), name.Labels),
	}}
}
