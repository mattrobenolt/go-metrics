package metrics

import "errors"

// NewFloatCounter registers and returns new FloatCounter with the given name in the s.
//
// family must be a Prometheus compatible identifier format.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	NewFloatCounter("family", "label1", "value1", "label2", "value2")
//
// The returned FloatCounter is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *Set) NewFloatCounter(family string, tags ...string) *FloatCounter {
	return s.NewFloatCounterOpt(MetricName{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
	})
}

// NewFloatCounterOpt registers and returns new FloatCounter with the name in the s.
//
// The returned Counter is safe to use from concurrent goroutines.
//
// This will panic if already registered.
func (s *Set) NewFloatCounterOpt(name MetricName) *FloatCounter {
	c := &FloatCounter{}
	s.mustRegisterMetric(c, name)
	return c
}

// A FloatCounterVec is a collection of FloatCounters that are partitioned
// by the same metric name and tag labels, but different tag values.
type FloatCounterVec struct {
	commonVec
}

// WithLabelValues returns the FloatCounter for the corresponding label values.
// If the combination of values is seen for the first time, a new FloatCounter
// is created.
//
// This will panic if the values count doesn't match the number of labels.
func (c *FloatCounterVec) WithLabelValues(values ...string) *FloatCounter {
	if len(values) != len(c.partialTags) {
		panic(errors.New("mismatch length of labels"))
	}
	hash := hashFinish(c.partialHash, values)

	c.s.metricsMu.Lock()
	nm := c.s.metrics[hash]
	c.s.metricsMu.Unlock()

	if nm == nil {
		nm = c.s.getOrRegisterMetricFromVec(
			&FloatCounter{}, hash, c.family, c.partialTags, values,
		)
	}
	return nm.metric.(*FloatCounter)
}

// NewFloatCounterVec creates a new [FloatCounterVec] with the supplied name.
func (s *Set) NewFloatCounterVec(name VecName) *FloatCounterVec {
	family := MustIdent(name.Family)

	return &FloatCounterVec{commonVec{
		s:           s,
		family:      family,
		partialTags: makePartialTags(name.Labels),
		partialHash: hashStart(family.String(), name.Labels),
	}}
}
