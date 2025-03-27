package metrics

import "errors"

// GaugeOpt are the options for creating a Gauge.
type GaugeOpt struct {
	Name MetricName
	// Func is an optional callback for making observations.
	Func func() float64
}

// NewGauge registers and returns gauge with the given name in s, which calls fn
// to obtain gauge value.
//
// family must be a Prometheus compatible identifier format.
//
// fn is an optional callback for making observations.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	NewGauge("family", observeFn, "label1", "value1", "label2", "value2")
//
// The returned Gauge is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *Set) NewGauge(family string, fn func() float64, tags ...string) *Gauge {
	return s.NewGaugeOpt(GaugeOpt{
		Name: MetricName{
			Family: MustIdent(family),
			Tags:   MustTags(tags...),
		},
		Func: fn,
	})
}

// NewGaugeOpt registers and returns new Gauge with the opts in the s.
//
// The returned Gauge is safe to use from concurrent goroutines.
//
// This will panic if already registered.
func (s *Set) NewGaugeOpt(opt GaugeOpt) *Gauge {
	g := &Gauge{fn: opt.Func}
	s.mustRegisterMetric(g, opt.Name)
	return g
}

// GetOrCreateGauge returns registered Gauge with the given name in s
// or creates new gauge if s doesn't contain Gauge with the given name.
//
// family must be a Prometheus compatible identifier format.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	GetOrCreateGauge("family", "label1", "value1", "label2", "value2")
//
// The returned Gauge is safe to use from concurrent goroutines.
//
// Prefer [NewGauge] or [NewGaugeOpt] when performance is critical.
//
// This will panic if values are invalid.
func (s *Set) GetOrCreateGauge(family string, tags ...string) *Gauge {
	hash := getHashStrings(family, tags)

	s.metricsMu.Lock()
	nm := s.metrics[hash]
	s.metricsMu.Unlock()

	if nm == nil {
		nm = s.getOrAddMetricFromStrings(&Gauge{}, hash, family, tags)
	}
	return nm.metric.(*Gauge)
}

// GaugeVecOpt are options for creating a new [GaugeVec].
type GaugeVecOpt struct {
	Name VecName
	// Func is an optional callback for making observations.
	Func func() float64
}

// A GaugeVec is a collection of Gauges that are partitioned
// by the same metric name and tag labels, but different tag values.
type GaugeVec struct {
	commonVec
	fn func() float64
}

// WithLabelValues returns the Gauge for the corresponding label values.
// If the combination of values is seen for the first time, a new Gauge
// is created.
//
// This will panic if the values count doesn't match the number of labels.
func (g *GaugeVec) WithLabelValues(values ...string) *Gauge {
	if len(values) != len(g.partialTags) {
		panic(errors.New("mismatch length of labels"))
	}
	hash := hashFinish(g.partialHash, values)

	g.s.metricsMu.Lock()
	nm := g.s.metrics[hash]
	g.s.metricsMu.Unlock()

	if nm == nil {
		nm = g.s.getOrRegisterMetricFromVec(
			&Gauge{fn: g.fn}, hash, g.family, g.partialTags, values,
		)
	}
	return nm.metric.(*Gauge)
}

// NewGaugeVec creates a new [GaugeVec] with the supplied opt.
func (s *Set) NewGaugeVec(opt GaugeVecOpt) *GaugeVec {
	family := MustIdent(opt.Name.Family)

	return &GaugeVec{
		commonVec: commonVec{
			s:           s,
			family:      family,
			partialTags: makePartialTags(opt.Name.Labels),
			partialHash: hashStart(family.String(), opt.Name.Labels),
		},
		fn: opt.Func,
	}
}
