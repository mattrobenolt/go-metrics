package metrics_test

import (
	"fmt"

	"github.com/VictoriaMetrics/metrics"
)

func ExampleCounter() {
	set := metrics.NewSet()

	// Define a counter..
	c := set.NewCounter(`metric_total{label1="value1", label2="value2"}`)

	// Increment the counter when needed.
	for range 10 {
		c.Inc()
	}
	n := c.Get()
	fmt.Println(n)

	// Output:
	// 10
}

func ExampleCounter_vec() {
	set := metrics.NewSet()

	for i := range 3 {
		// Dynamically construct metric name and pass it to GetOrCreateCounter.
		name := fmt.Sprintf(`metric_total{label1=%q, label2="%d"}`, "value1", i)
		set.GetOrCreateCounter(name).Add(i + 1)
	}

	// Read counter values.
	for i := range 3 {
		name := fmt.Sprintf(`metric_total{label1=%q, label2="%d"}`, "value1", i)
		n := set.GetOrCreateCounter(name).Get()
		fmt.Println(n)
	}

	// Output:
	// 1
	// 2
	// 3
}
