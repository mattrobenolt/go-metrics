package metrics_test

import (
	"fmt"
	"strconv"

	"go.withmatt.com/metrics"
)

func ExampleCounter() {
	// Define a counter in global scope.
	c := metrics.NewCounter(
		"metric_total",
		"label1", "value1",
		"label2", "value2",
	)

	// Increment the counter when needed.
	for range 10 {
		c.Inc()
	}
	n := c.Get()
	fmt.Println(n)

	// Output:
	// 10
}

func ExampleCounterVec() {
	metricTotal := metrics.NewCounterVec(metrics.CounterVecOpt{
		Family: "metric_total",
		Labels: []string{"label1", "label2"},
	})
	for i := range 3 {
		// Dynamically construct metric name with label values
		metricTotal.WithLabelValues(
			"value1", strconv.Itoa(i),
		).Add(uint64(i + 1))
	}

	// Read counter values.
	for i := range 3 {
		n := metricTotal.WithLabelValues(
			"value1", strconv.Itoa(i),
		).Get()
		fmt.Println(n)
	}

	// Output:
	// 1
	// 2
	// 3
}
