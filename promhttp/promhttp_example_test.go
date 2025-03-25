package promhttp_test

import (
	"net/http"

	"go.withmatt.com/metrics"
	"go.withmatt.com/metrics/promhttp"
)

func ExampleHandler() {
	// Export all globally registered metrics.
	http.Handle("/metrics", promhttp.Handler())
}

func ExampleHandlerFor() {
	set := metrics.NewSet()
	set.NewCounter("foo").Inc()

	// Export all metrics from a specific Set.
	http.Handle("/metrics", promhttp.HandlerFor(set))
}
