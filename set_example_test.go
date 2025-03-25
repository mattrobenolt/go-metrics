package metrics_test

import (
	"bytes"
	"fmt"

	"go.withmatt.com/metrics"
)

func ExampleSet() {
	// Create a new collection of metrics.
	set := metrics.NewSet()
	// Create and increment a counter on this set.
	c := set.NewCounter("foo")
	c.Inc()

	fmt.Println(c.Get())

	// Output:
	// 1
}

func ExampleWritePrometheus() {
	set := metrics.NewSet()
	set.NewCounter("foo", "label1", "value1").Inc()

	// Export all the registered metrics into a bytes buffer.
	var b bytes.Buffer
	set.WritePrometheus(&b)
	fmt.Println(b.String())

	// Output:
	// foo{label1="value1"} 1
}
