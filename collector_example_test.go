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

func ExampleCollectorFunc() {
	// CollectorFunc allows you to easily create custom collectors
	// from regular functions
	customCollector := metrics.CollectorFunc(func(w metrics.ExpfmtWriter) {
		w.WriteLazyMetricUint64("custom_metric", 42, "type", "example")
	})

	// Register the collector on a set
	set := metrics.NewSet()
	set.RegisterCollector(customCollector)

	// Write metrics
	var buf bytes.Buffer
	set.WritePrometheus(&buf)
	fmt.Print(buf.String())

	// Output:
	// custom_metric{type="example"} 42
}

func ExampleSet_Collect() {
	// Create a set with application metrics
	appMetrics := metrics.NewSet("app", "myapp")
	requestCount := appMetrics.NewCounter("requests_total")
	requestCount.Add(100)

	// Create a set with database metrics
	dbMetrics := metrics.NewSet("component", "database")
	queryCount := dbMetrics.NewCounter("queries_total")
	queryCount.Add(50)

	// Create a main set and register the other sets as collectors
	mainSet := metrics.NewSet()
	mainSet.RegisterCollector(appMetrics, dbMetrics)

	// Write all metrics
	var buf bytes.Buffer
	mainSet.WritePrometheus(&buf)
	fmt.Print(buf.String())

	// Output:
	// requests_total{app="myapp"} 100
	// queries_total{component="database"} 50
}

func ExampleSet_Collect_withAppendConstantTags() {
	// Create submodule sets for different parts of an application
	authModule := metrics.NewSet()
	authModule.NewCounter("requests").Add(100)
	authModule.NewCounter("errors").Add(5)

	apiModule := metrics.NewSet()
	apiModule.NewCounter("requests").Add(500)
	apiModule.NewCounter("errors").Add(10)

	// Create a collector that adds module-specific tags
	moduleCollector := metrics.CollectorFunc(func(w metrics.ExpfmtWriter) {
		// Each module gets its own identifying tag
		authModule.Collect(w.AppendConstantTags("module", "auth"))
		apiModule.Collect(w.AppendConstantTags("module", "api"))
	})

	// Create main application set
	appSet := metrics.NewSet("app", "myservice")
	appSet.RegisterCollector(moduleCollector)

	// Write metrics
	var buf bytes.Buffer
	appSet.WritePrometheus(&buf)
	fmt.Print(buf.String())

	// Output:
	// errors{app="myservice",module="auth"} 5
	// requests{app="myservice",module="auth"} 100
	// errors{app="myservice",module="api"} 10
	// requests{app="myservice",module="api"} 500
}
