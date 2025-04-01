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

func ExampleAnnotatedHandler() {
	// Export all globally registered metrics with default mappings.
	http.Handle("/metrics", promhttp.AnnotatedHandler(nil))

	// Create a Mapping for our metrics families.
	mapping := promhttp.Mapping{
		"foo": {
			Type: promhttp.Counter,
			Help: "This is a counter",
		},
	}

	// Export all globally registered metrics with our mapping.
	http.Handle("/metrics", promhttp.AnnotatedHandler(mapping))
}
