package metrics_test

import (
	"bytes"
	"fmt"

	"go.withmatt.com/metrics"
)

type testCollector struct{}

func (c *testCollector) Collect(w metrics.ExpfmtWriter) {
	// Quick, less optimal way
	w.WriteLazyMetricUint64("foo", 1, "label1", "value1", "label2", "value2")

	// You'd want to save this MetricName and reuse it for
	// most optimal performance.
	name := metrics.MetricName{
		Family: metrics.MustIdent("other"),
		Tags:   metrics.MustTags("a", "b"),
	}
	w.WriteMetricUint64(name, 2)
}

func ExampleCollector() {
	metrics.RegisterCollector(&testCollector{})

	var b bytes.Buffer
	metrics.WritePrometheus(&b)
	fmt.Println(b.String())

	// Output:
	// foo{label1="value1",label2="value2"} 1
	// other{a="b"} 2
}
