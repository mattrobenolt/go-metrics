package promhttp_test

import (
	"bytes"
	"fmt"
	"io"
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

func ExampleTransformer() {
	set := metrics.NewSet()
	set.NewCounter("foo").Inc()

	// Create a Mapping for our metrics families.
	mapping := promhttp.Mapping{
		"foo": {
			Type: promhttp.Counter,
			Help: "This is a counter",
		},
	}

	// Create a Transformer for our metrics.
	tr := promhttp.NewTransformer(mapping)

	// Write our metrics into the transformer.
	set.WritePrometheus(tr)

	// Write transformed data from the transformer to a buffer.
	var bb bytes.Buffer
	io.Copy(&bb, tr)

	fmt.Println(bb.String())
	// Output:
	// # HELP foo This is a counter
	// # TYPE foo counter
	// foo 1
}
