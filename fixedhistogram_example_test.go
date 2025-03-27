package metrics_test

import (
	"time"

	"go.withmatt.com/metrics"
)

func ExampleFixedHistogram() {
	// Define a histogram in global scope.
	h := metrics.NewFixedHistogram(
		"request_duration_seconds",
		[]float64{0.005, 0.01, 0.05, 0.1, 0.25, 0.5, 1},
		"path", "/foo/bar",
	)

	// Update the histogram with the duration of processRequest call.
	startTime := time.Now()
	processRequest()
	h.UpdateDuration(startTime)
}

func ExampleFixedHistogramVec() {
	responseSizeBytes := metrics.NewFixedHistogramVec(metrics.FixedHistogramVecOpt{
		Name: metrics.VecName{
			Family: "response_size_bytes",
			Labels: []string{"path"},
		},
		Buckets: []float64{0.005, 0.01, 0.05, 0.1, 0.25, 0.5, 1},
	})
	for range 3 {
		response := processRequest()
		// Dynamically construct metric name with label values
		responseSizeBytes.WithLabelValues(
			"/foo/bar",
		).Update(float64(len(response)))
	}
}
