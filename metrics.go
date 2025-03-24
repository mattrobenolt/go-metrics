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
	Collect(w ExpfmtWriter, constantTags string)
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
	Family       Ident
	Tags         []Tag
	ConstantTags string
}

// With returns a new MetricName with constantTags appended to existing.
func (n MetricName) With(constantTags string) MetricName {
	nn := MetricName{
		Family:       n.Family,
		Tags:         n.Tags,
		ConstantTags: n.ConstantTags,
	}
	switch {
	case len(constantTags) == 0:
		// do nothing
	case len(nn.ConstantTags) == 0:
		nn.ConstantTags = constantTags
	default:
		nn.ConstantTags = nn.ConstantTags + "," + constantTags
	}
	return nn
}

// String returns the MetricName in fully quanfied format. Prefer
// [ExpfmtWriter.WriteMetricName] over this when marshalling.
func (n MetricName) String() string {
	if !n.HasTags() {
		return n.Family.String()
	}
	var b bytes.Buffer
	writeMetricName(&b, n)
	return b.String()
}

func (n MetricName) HasTags() bool {
	return len(n.ConstantTags) > 0 || len(n.Tags) > 0
}

func compareNamedMetrics(a, b *namedMetric) int {
	return cmp.Compare(
		a.family.String(),
		b.family.String(),
	)
}
