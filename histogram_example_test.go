package metrics_test

import (
	"time"

	"go.withmatt.com/metrics"
)

func ExampleHistogram() {
	// Define a histogram in global scope.
	h := metrics.NewHistogram(
		"request_duration_seconds",
		"path", "/foo/bar",
	)

	// Update the histogram with the duration of processRequest call.
	startTime := time.Now()
	processRequest()
	h.UpdateDuration(startTime)
}

func ExampleHistogramVec() {
	responseSizeBytes := metrics.NewHistogramVec(metrics.HistogramVecOpt{
		Family: "response_size_bytes",
		Labels: []string{"path"},
	})
	for range 3 {
		response := processRequest()
		// Dynamically construct metric name with label values
		responseSizeBytes.WithLabelValues(
			"/foo/bar",
		).Update(float64(len(response)))
	}
}

func processRequest() string {
	return "foobar"
}
