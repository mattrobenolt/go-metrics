package metrics

// A Uint64Vec is a collection of Uint64s that are partitioned
// by the same metric name and tag labels, but different tag values.
type Uint64Vec struct {
	commonVec
}

// NewUint64Vec creates a new Uint64Vec on the global Set.
// See [Set.NewUint64Vec].
func NewUint64Vec(family string, labels ...string) *Uint64Vec {
	return defaultSet.NewUint64Vec(family, labels...)
}

// NewCounterVec creates a new Uint64Vec on the global Set.
// See [Set.NewUint64Vec].
func NewCounterVec(family string, labels ...string) *Uint64Vec {
	return defaultSet.NewCounterVec(family, labels...)
}

// WithLabelValues returns the Uint64 for the corresponding label values.
// If the combination of values is seen for the first time, a new Uint
// is created.
//
// This will panic if the values count doesn't match the number of labels.
func (c *Uint64Vec) WithLabelValues(values ...string) *Uint64 {
	hash := hashFinish(c.partialHash, values)

	nm, ok := c.s.metrics.Load(hash)
	if !ok {
		nm = c.s.getOrRegisterMetricFromVec(
			&Uint64{}, hash, c.family, c.partialTags, values,
		)
	}
	return nm.metric.(*Uint64)
}

// NewUint64Vec creates a new [Uint64Vec] with the supplied name.
func (s *Set) NewUint64Vec(family string, labels ...string) *Uint64Vec {
	return &Uint64Vec{commonVec{
		s:           s,
		family:      MustIdent(family),
		partialTags: makePartialTags(labels),
		partialHash: hashStart(family, labels),
	}}
}

// NewCounterVec is an alias for [Set.NewUint64Vec].
func (s *Set) NewCounterVec(family string, labels ...string) *Uint64Vec {
	return s.NewUint64Vec(family, labels...)
}

// A Int64Vec is a collection of Int64s that are partitioned
// by the same metric name and tag labels, but different tag values.
type Int64Vec struct {
	commonVec
}

// NewInt64Vec creates a new Int64Vec on the global Set.
// See [Set.NewInt64Vec].
func NewInt64Vec(family string, labels ...string) *Int64Vec {
	return defaultSet.NewInt64Vec(family, labels...)
}

// WithLabelValues returns the Int64 for the corresponding label values.
// If the combination of values is seen for the first time, a new Int
// is created.
//
// This will panic if the values count doesn't match the number of labels.
func (c *Int64Vec) WithLabelValues(values ...string) *Int64 {
	hash := hashFinish(c.partialHash, values)

	nm, ok := c.s.metrics.Load(hash)
	if !ok {
		nm = c.s.getOrRegisterMetricFromVec(
			&Int64{}, hash, c.family, c.partialTags, values,
		)
	}
	return nm.metric.(*Int64)
}

// NewInt64Vec creates a new [Int64Vec] with the supplied name.
func (s *Set) NewInt64Vec(family string, labels ...string) *Int64Vec {
	return &Int64Vec{commonVec{
		s:           s,
		family:      MustIdent(family),
		partialTags: makePartialTags(labels),
		partialHash: hashStart(family, labels),
	}}
}

// A Float64Vec is a collection of Float64s that are partitioned
// by the same metric name and tag labels, but different tag values.
type Float64Vec struct {
	commonVec
}

// NewFloat64Vec creates a new Float64Vec on the global Set.
// See [Set.NewFloat64Vec].
func NewFloat64Vec(family string, labels ...string) *Float64Vec {
	return defaultSet.NewFloat64Vec(family, labels...)
}

// WithLabelValues returns the Float for the corresponding label values.
// If the combination of values is seen for the first time, a new Float
// is created.
//
// This will panic if the values count doesn't match the number of labels.
func (c *Float64Vec) WithLabelValues(values ...string) *Float64 {
	hash := hashFinish(c.partialHash, values)

	nm, ok := c.s.metrics.Load(hash)
	if !ok {
		nm = c.s.getOrRegisterMetricFromVec(
			&Float64{}, hash, c.family, c.partialTags, values,
		)
	}
	return nm.metric.(*Float64)
}

// NewFloat64Vec creates a new [Float64Vec] with the supplied name.
func (s *Set) NewFloat64Vec(family string, labels ...string) *Float64Vec {
	return &Float64Vec{commonVec{
		s:           s,
		family:      MustIdent(family),
		partialTags: makePartialTags(labels),
		partialHash: hashStart(family, labels),
	}}
}
