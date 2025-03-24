package metrics

import "hash/maphash"

// FloatCounterOpt are the options for creating a Counter.
type FloatCounterOpt struct {
	Family Ident
	Tags   []Tag
}

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
	return s.NewFloatCounterOpt(FloatCounterOpt{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
	})
}

// NewFloatCounterOpt registers and returns new FloatCounter with the opts in the s.
//
// The returned Counter is safe to use from concurrent goroutines.
//
// This will panic if already registered.
func (s *Set) NewFloatCounterOpt(opt FloatCounterOpt) *FloatCounter {
	c := &FloatCounter{}
	s.mustRegisterMetric(c, opt.Family, opt.Tags)
	return c
}

// GetOrCreateFloatCounter returns registered FloatCounter in s with the given name
// and tags creates new FloatCounter if s doesn't contain FloatCounter with the given name.
//
// family must be a Prometheus compatible identifier format.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	GetOrCreateFloatCounter("family", "label1", "value1", "label2", "value2")
//
// The returned Counter is safe to use from concurrent goroutines.
//
// Prefer [NewFloatCounter] or [NewFloatCounterOpt] when performance is critical.
//
// This will panic if values are invalid.
func (s *Set) GetOrCreateFloatCounter(family string, tags ...string) *FloatCounter {
	hash := getHashStrings(family, tags)

	s.metricsMu.Lock()
	nm := s.metrics[hash]
	s.metricsMu.Unlock()

	if nm == nil {
		nm = s.getOrAddMetricFromStrings(&FloatCounter{}, hash, family, tags)
	}
	return nm.metric.(*FloatCounter)
}

type FloatCounterVecOpt struct {
	Family string
	Labels []string
}

type FloatCounterVec struct {
	s           *Set
	family      Ident
	partialTags []Tag
	partialHash *maphash.Hash
}

func (c *FloatCounterVec) WithLabelValues(values ...string) *FloatCounter {
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

func (s *Set) NewFloatCounterVec(opt FloatCounterVecOpt) *FloatCounterVec {
	family := MustIdent(opt.Family)

	return &FloatCounterVec{
		s:           s,
		family:      family,
		partialTags: makePartialTags(opt.Labels),
		partialHash: hashStart(family.String(), opt.Labels),
	}
}
