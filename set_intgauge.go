package metrics

import "hash/maphash"

// IntGaugeOpt are the options for creating a Gauge.
type IntGaugeOpt struct {
	// Family is the metric Ident, see [MustIdent].
	Family Ident
	// Tags are optional tags for the metric, see [MustTags].
	Tags []Tag
	// Func is an optional callback for making observations.
	Func func() uint64
}

// NewIntGauge registers and returns gauge with the given name in s, which calls fn
// to obtain gauge value.
//
// family must be a Prometheus compatible identifier format.
//
// fn is an optional callback for making observations.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	NewIntGauge("family", observeFn, "label1", "value1", "label2", "value2")
//
// The returned IntGauge is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *Set) NewIntGauge(family string, fn func() uint64, tags ...string) *IntGauge {
	return s.NewIntGaugeOpt(IntGaugeOpt{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
		Func:   fn,
	})
}

// NewIntGaugeOpt registers and returns new IntGauge with the opts in the s.
//
// The returned IntGauge is safe to use from concurrent goroutines.
//
// This will panic if already registered.
func (s *Set) NewIntGaugeOpt(opt IntGaugeOpt) *IntGauge {
	g := &IntGauge{fn: opt.Func}
	s.mustRegisterMetric(g, opt.Family, opt.Tags)
	return g
}

// GetOrCreateIntGauge returns registered IntGauge with the given name in s
// or creates new gauge if s doesn't contain IntGauge with the given name.
//
// family must be a Prometheus compatible identifier format.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	GetOrCreateGauge("family", "label1", "value1", "label2", "value2")
//
// The returned Gauge is safe to use from concurrent goroutines.
//
// Prefer [NewIntGauge] or [NewGaugeIntOpt] when performance is critical.
//
// This will panic if values are invalid.
func (s *Set) GetOrCreateIntGauge(family string, tags ...string) *IntGauge {
	hash := getHashStrings(family, tags)

	s.metricsMu.Lock()
	nm := s.metrics[hash]
	s.metricsMu.Unlock()

	if nm == nil {
		nm = s.getOrAddMetricFromStrings(&IntGauge{}, hash, family, tags)
	}
	return nm.metric.(*IntGauge)
}

// IntGaugeVecOpt are options for creating a new [IntGaugeVec].
type IntGaugeVecOpt struct {
	// Family is the metric family name, e.g. `http_requests`
	Family string
	// Labels are the tag labels that you want to partition on, e.g. "status", "path"
	Labels []string
	// Func is an optional callback for making observations.
	Func func() uint64
}

// A IntGaugeVec is a collection of IntGauges that are partitioned
// by the same metric name and tag labels, but different tag values.
type IntGaugeVec struct {
	s           *Set
	family      Ident
	partialTags []Tag
	partialHash *maphash.Hash
	fn          func() uint64
}

// WithLabelValues returns the IntGauge for the corresponding label values.
// If the combination of values is seen for the first time, a new IntGauge
// is created.
func (g *IntGaugeVec) WithLabelValues(values ...string) *IntGauge {
	hash := hashFinish(g.partialHash, values)

	g.s.metricsMu.Lock()
	nm := g.s.metrics[hash]
	g.s.metricsMu.Unlock()

	if nm == nil {
		nm = g.s.getOrRegisterMetricFromVec(
			&IntGauge{fn: g.fn}, hash, g.family, g.partialTags, values,
		)
	}
	return nm.metric.(*IntGauge)
}

// NewIntGaugeVec creates a new [IntGaugeVec] with the supplied opt.
func (s *Set) NewIntGaugeVec(opt IntGaugeVecOpt) *IntGaugeVec {
	family := MustIdent(opt.Family)

	return &IntGaugeVec{
		s:           s,
		family:      family,
		partialTags: makePartialTags(opt.Labels),
		partialHash: hashStart(family.String(), opt.Labels),
		fn:          opt.Func,
	}
}
