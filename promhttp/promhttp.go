/*
Package promhttp provides a Prometheus-compatible HTTP endpoint.

By default, we do not write out any HELP or TYPE annotations. This
is compatible with VictoriaMetrics and potentially other scrapers,
but if these annotations are required, you can use an [AnnotatedHandler].

Prefer [Handler] and [HandlerFor] when annotations aren't explicitly required.

Compression is not supported out of the box. I would recommend wrapping the
http.Handler with something like
https://pkg.go.dev/github.com/klauspost/compress/gzhttp.

For example:

	gzhttp.GzipHandler(promhttp.Handler())
*/
package promhttp

import (
	"io"
	"net/http"

	"go.withmatt.com/metrics"
)

// ContentType is the HTTP Content-Type header for this format.
const ContentType = "text/plain; version=0.0.4"

// Handler returns an http.Handler for the global metrics Set.
func Handler() http.Handler {
	return handler(metrics.WritePrometheus)
}

// HandlerFor returns an http.Handler for a specific metrics Set.
func HandlerFor(set *metrics.Set) http.Handler {
	return handler(set.WritePrometheus)
}

// AnnotatedHandler returns an http.Handler for the global metrics Set and will
// add HELP and TYPE annotations according to the [Mapping].
func AnnotatedHandler(m Mapping) http.Handler {
	return annotationHandler(metrics.WritePrometheus, m)
}

// AnnotatedHandlerFor returns an http.Handler for a specific metrics Set and will
// add HELP and TYPE annotations according to the [Mapping].
func AnnotatedHandlerFor(set *metrics.Set, m Mapping) http.Handler {
	return annotationHandler(set.WritePrometheus, m)
}

type writerFunc func(w io.Writer) (int, error)

func handler(writePrometheus writerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", ContentType)
		writePrometheus(w)
	})
}

func annotationHandler(writePrometheus writerFunc, m Mapping) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", ContentType)
		tr := NewTransformer(m)
		writePrometheus(tr)
		io.Copy(w, tr)
	})
}
