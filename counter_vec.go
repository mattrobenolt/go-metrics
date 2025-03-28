package metrics

import "errors"

// A UintVec is a collection of Uints that are partitioned
// by the same metric name and tag labels, but different tag values.
type UintVec struct {
	commonVec
}

// NewUintVec creates a new UintVec on the global Set.
// See [Set.NewUintVec].
func NewUintVec(family string, labels ...string) *UintVec {
	return defaultSet.NewUintVec(family, labels...)
}

// NewCounterVec creates a new UintVec on the global Set.
// See [Set.NewUintVec].
func NewCounterVec(family string, labels ...string) *UintVec {
	return defaultSet.NewCounterVec(family, labels...)
}

// WithLabelValues returns the Uint for the corresponding label values.
// If the combination of values is seen for the first time, a new Uint
// is created.
//
// This will panic if the values count doesn't match the number of labels.
func (c *UintVec) WithLabelValues(values ...string) *Uint {
	if len(values) != len(c.partialTags) {
		panic(errors.New("mismatch length of labels"))
	}
	hash := hashFinish(c.partialHash, values)

	c.s.metricsMu.Lock()
	nm := c.s.metrics[hash]
	c.s.metricsMu.Unlock()

	if nm == nil {
		nm = c.s.getOrRegisterMetricFromVec(
			&Uint{}, hash, c.family, c.partialTags, values,
		)
	}
	return nm.metric.(*Uint)
}

// NewUintVec creates a new [UintVec] with the supplied name.
func (s *Set) NewUintVec(family string, labels ...string) *UintVec {
	return &UintVec{commonVec{
		s:           s,
		family:      MustIdent(family),
		partialTags: makePartialTags(labels),
		partialHash: hashStart(family, labels),
	}}
}

// NewCounterVec is an alias for [Set.NewUintVec].
func (s *Set) NewCounterVec(family string, labels ...string) *UintVec {
	return s.NewUintVec(family, labels...)
}

// A IntVec is a collection of Ints that are partitioned
// by the same metric name and tag labels, but different tag values.
type IntVec struct {
	commonVec
}

// NewIntVec creates a new IntVec on the global Set.
// See [Set.NewIntVec].
func NewIntVec(family string, labels ...string) *IntVec {
	return defaultSet.NewIntVec(family, labels...)
}

// WithLabelValues returns the Int for the corresponding label values.
// If the combination of values is seen for the first time, a new Int
// is created.
//
// This will panic if the values count doesn't match the number of labels.
func (c *IntVec) WithLabelValues(values ...string) *Int {
	if len(values) != len(c.partialTags) {
		panic(errors.New("mismatch length of labels"))
	}
	hash := hashFinish(c.partialHash, values)

	c.s.metricsMu.Lock()
	nm := c.s.metrics[hash]
	c.s.metricsMu.Unlock()

	if nm == nil {
		nm = c.s.getOrRegisterMetricFromVec(
			&Int{}, hash, c.family, c.partialTags, values,
		)
	}
	return nm.metric.(*Int)
}

// NewIntVec creates a new [IntVec] with the supplied name.
func (s *Set) NewIntVec(family string, labels ...string) *IntVec {
	return &IntVec{commonVec{
		s:           s,
		family:      MustIdent(family),
		partialTags: makePartialTags(labels),
		partialHash: hashStart(family, labels),
	}}
}

// A FloatVec is a collection of Floats that are partitioned
// by the same metric name and tag labels, but different tag values.
type FloatVec struct {
	commonVec
}

// NewFloatVec creates a new FloatVec on the global Set.
// See [Set.NewFloatVec].
func NewFloatVec(family string, labels ...string) *FloatVec {
	return defaultSet.NewFloatVec(family, labels...)
}

// WithLabelValues returns the Float for the corresponding label values.
// If the combination of values is seen for the first time, a new Float
// is created.
//
// This will panic if the values count doesn't match the number of labels.
func (c *FloatVec) WithLabelValues(values ...string) *Float {
	if len(values) != len(c.partialTags) {
		panic(errors.New("mismatch length of labels"))
	}
	hash := hashFinish(c.partialHash, values)

	c.s.metricsMu.Lock()
	nm := c.s.metrics[hash]
	c.s.metricsMu.Unlock()

	if nm == nil {
		nm = c.s.getOrRegisterMetricFromVec(
			&Float{}, hash, c.family, c.partialTags, values,
		)
	}
	return nm.metric.(*Float)
}

// NewFloatVec creates a new [FloatVec] with the supplied name.
func (s *Set) NewFloatVec(family string, labels ...string) *FloatVec {
	return &FloatVec{commonVec{
		s:           s,
		family:      MustIdent(family),
		partialTags: makePartialTags(labels),
		partialHash: hashStart(family, labels),
	}}
}
