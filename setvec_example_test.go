package metrics_test

import (
	"fmt"
	"time"

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

func ExampleNewSetVecWithTTL() {
	// Create a Set faceted by unique user ID, and expires after an hour.
	byUserID := metrics.NewSetVecWithTTL("user_id", time.Hour)
	// Define a histogram for page view durations by user ID, path, and status.
	pageviewDurationByUserID := byUserID.NewHistogramVec(
		"pageview_duration_seconds",
		"path", "status",
	)

	userID := "1234"
	startTime := time.Now()
	processRequest()

	pageviewDurationByUserID.WithLabelValues(
		userID, "/foo/bar", "200",
	).UpdateDuration(startTime)
}
