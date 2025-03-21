package metrics

import (
	"cmp"
)

// Metric is a single data point that can be written to the Prometheus
// text format.
type Metric interface {
	marshalTo(w ExpfmtWriter, family Ident, tags ...Tag)
}

type Collector interface {
	Collect(w ExpfmtWriter)
}

// namedMetric is a single data point.
type namedMetric struct {
	// id is the unique hash to represent metric series.
	// the hash is based on the family an tags
	id     metricHash
	family Ident
	tags   []Tag
	metric Metric
}

func compareNamedMetrics(a, b *namedMetric) int {
	return cmp.Compare(
		a.family.String(),
		b.family.String(),
	)
}
