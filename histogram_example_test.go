package metrics_test

import (
	"fmt"
	"time"

	"github.com/VictoriaMetrics/metrics"
)

func ExampleHistogram() {
	set := metrics.NewSet()

	// Define a histogram.
	h := set.NewHistogram(`request_duration_seconds{path="/foo/bar"}`)

	// Update the histogram with the duration of processRequest call.
	startTime := time.Now()
	processRequest()
	h.UpdateDuration(startTime)
}

func ExampleHistogram_vec() {
	set := metrics.NewSet()

	for range 3 {
		// Dynamically construct metric name and pass it to GetOrCreateHistogram.
		name := fmt.Sprintf(`response_size_bytes{path=%q}`, "/foo/bar")
		response := processRequest()
		set.GetOrCreateHistogram(name).Update(float64(len(response)))
	}
}
