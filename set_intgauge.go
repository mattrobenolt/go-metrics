package metrics

import "hash/maphash"

// IntGaugeOpt are the options for creating a Gauge.
type IntGaugeOpt struct {
	Family Ident
	Tags   []Tag
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

type IntGaugeVecOpt struct {
	Family string
	Labels []string
	Func   func() uint64
}

type IntGaugeVec struct {
	s           *Set
	family      Ident
	partialTags []Tag
	partialHash *maphash.Hash
	fn          func() uint64
}

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
