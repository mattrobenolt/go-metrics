package metrics_test

import (
	"fmt"
	"time"

	"github.com/VictoriaMetrics/metrics"
)

func ExampleSummary() {
	set := metrics.NewSet()

	// Define a summary in global scope.
	s := set.NewSummary(`request_duration_seconds{path="/foo/bar"}`)

	// Update the summary with the duration of processRequest call.
	startTime := time.Now()
	processRequest()
	s.UpdateDuration(startTime)
}

func ExampleSummary_vec() {
	set := metrics.NewSet()

	for range 3 {
		// Dynamically construct metric name and pass it to GetOrCreateSummary.
		name := fmt.Sprintf(`response_size_bytes{path=%q}`, "/foo/bar")
		response := processRequest()
		set.GetOrCreateSummary(name).Update(float64(len(response)))
	}
}

func processRequest() string {
	return "foobar"
}
