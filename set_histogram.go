package metrics

import "hash/maphash"

// HistogramOpt are the options for creating a Histogram.
type HistogramOpt struct {
	Family Ident
	Tags   []Tag
}

// NewHistogram creates and returns new Histogram in s with the given name.
//
// family must be a Prometheus compatible identifier format.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	NewHistogram("family", "label1", "value1", "label2", "value2")
//
// The returned Histogram is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *Set) NewHistogram(family string, tags ...string) *Histogram {
	return s.NewHistogramOpt(HistogramOpt{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
	})
}

// NewHistogramOpt registers and returns new Histogram with the opts in the s.
//
// The returned Histogram is safe to use from concurrent goroutines.
//
// This will panic if already registered.
func (s *Set) NewHistogramOpt(opt HistogramOpt) *Histogram {
	h := &Histogram{}
	s.mustRegisterMetric(h, opt.Family, opt.Tags)
	return h
}

// GetOrCreateHistogram returns registered Histogram in s with the given name
// and tags creates new Histogram if s doesn't contain Histogram with the given name.
//
// family must be a Prometheus compatible identifier format.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	GetOrCreateHistogram("family", "label1", "value1", "label2", "value2")
//
// The returned Histogram is safe to use from concurrent goroutines.
//
// Prefer [NewHistogram] or [NewHistogramOpt] when performance is critical.
//
// This will panic if values are invalid.
func (s *Set) GetOrCreateHistogram(family string, tags ...string) *Histogram {
	hash := getHashStrings(family, tags)

	s.mu.Lock()
	nm := s.metrics[hash]
	s.mu.Unlock()

	if nm == nil {
		nm = s.getOrAddMetricFromStrings(&Histogram{}, hash, family, tags)
	}
	return nm.metric.(*Histogram)
}

type HistogramVecOpt struct {
	Family string
	Labels []string
}

type HistogramVec struct {
	s           *Set
	family      Ident
	partialTags []Tag
	partialHash *maphash.Hash
}

func (h *HistogramVec) WithLabelValues(values ...string) *Histogram {
	hash := hashFinish(h.partialHash, values)

	h.s.mu.Lock()
	nm := h.s.metrics[hash]
	h.s.mu.Unlock()

	if nm == nil {
		nm = h.s.getOrRegisterMetricFromVec(
			&Histogram{}, hash, h.family, h.partialTags, values,
		)
	}
	return nm.metric.(*Histogram)
}

func (s *Set) NewHistogramVec(opt HistogramVecOpt) *HistogramVec {
	family := MustIdent(opt.Family)

	return &HistogramVec{
		s:           s,
		family:      family,
		partialTags: makePartialTags(opt.Labels),
		partialHash: hashStart(family.String(), opt.Labels),
	}
}
