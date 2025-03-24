package metrics

import (
	"bytes"
	"cmp"
)

// Metric is a single data point that can be written to the Prometheus
// text format.
type Metric interface {
	marshalTo(ExpfmtWriter, MetricName)
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

// MetricName represents a FQN of a metric in pieces.
type MetricName struct {
	Family Ident
	Tags   []Tag
}

// String returns the MetricName in fully quanfied format. Prefer
// [ExpfmtWriter.WriteMetricName] over this when marshalling.
func (n MetricName) String() string {
	if !n.HasTags() {
		return n.Family.String()
	}
	var b bytes.Buffer
	writeMetricName(&b, n, "")
	return b.String()
}

func (n MetricName) HasTags() bool {
	return len(n.Tags) > 0
}

func compareNamedMetrics(a, b *namedMetric) int {
	return cmp.Compare(
		a.family.String(),
		b.family.String(),
	)
}
