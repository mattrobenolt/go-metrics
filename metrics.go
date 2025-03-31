/*
Package metrics provides an extremely fast and lightweight API for
recording and exporting metrics in Prometheus format.
*/
package metrics

import (
	"bytes"
	"cmp"
	"hash/maphash"
)

// Metric is a single data point that can be written to the Prometheus
// text format.
type Metric interface {
	marshalTo(ExpfmtWriter, MetricName)
}

// Collector is custom data collector that is called during [Set.WritePrometheus].
type Collector interface {
	Collect(ExpfmtWriter)
}

// namedMetric is a single data point.
type namedMetric struct {
	// id is the unique hash to represent metric series.
	// the hash is based on the family an tags
	id     metricHash
	name   MetricName
	metric Metric
}

// MetricName represents a fully qualified name of a metric in pieces.
type MetricName struct {
	// Family is the metric Ident, see [MustIdent].
	Family Ident
	// Tags are optional tags for the metric, see [MustTags].
	Tags []Tag
}

// String returns the MetricName in fully qualified format. Prefer
// [ExpfmtWriter.WriteMetricName] over this when marshalling.
func (n MetricName) String() string {
	if !n.hasTags() {
		return n.Family.String()
	}
	var b bytes.Buffer
	writeMetricName(&b, n, "")
	return b.String()
}

func (n MetricName) hasTags() bool {
	return len(n.Tags) > 0
}

func compareNamedMetrics(a, b *namedMetric) int {
	return cmp.Compare(
		a.name.Family.String(),
		b.name.Family.String(),
	)
}

type commonVec struct {
	s           *Set
	family      Ident
	partialTags []Label
	partialHash *maphash.Hash
}
