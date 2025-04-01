package promhttp_test

import (
	"bytes"
	"fmt"
	"io"

	"go.withmatt.com/metrics"
	"go.withmatt.com/metrics/promhttp"
)

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
