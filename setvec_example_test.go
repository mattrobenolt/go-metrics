package metrics_test

import (
	"fmt"

	"go.withmatt.com/metrics"
)

func ExampleSetVec() {
	// Create a new Set faceted by unique path.
	byPath := metrics.NewSetVec("path")
	// Create a Uint64 counter for page views by path and status.
	viewsByPath := byPath.NewUint64Vec("pageview", "status")

	// Increment a counter for /abc path, with a 200 status.
	viewsByPath.WithLabelValues("/abc", "200").Inc()
	// Increment a counter for /notfound path, with a 404 status.
	viewsByPath.WithLabelValues("/notfound", "404").Inc()

	fmt.Println(viewsByPath.WithLabelValues("/abc", "200").Get())

	// Delete entire set of metrics for the /abc path.
	byPath.RemoveByLabelValue("/abc")

	fmt.Println(viewsByPath.WithLabelValues("/abc", "200").Get())

	// Output:
	// 1
	// 0
}
