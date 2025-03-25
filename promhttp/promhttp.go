/*
Package promhttp provides a Prometheus-compatible HTTP endpoint.

promhttp doesn't support and compression out of the box. I would
recommend wrapping the http.Handler with something like
https://pkg.go.dev/github.com/klauspost/compress/gzhttp.

For example:

	gzhttp.GzipHandler(promhttp.Handler())
*/
package promhttp

import (
	"net/http"

	"go.withmatt.com/metrics"
)

// HTTP Content-Type header for this format.
const ContentType = "text/plain; version=0.0.4"

// Handler returns an http.Handler for the global metrics Set.
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", ContentType)
		metrics.WritePrometheus(w)
	})
}

// HandlerFor returns an http.Handler for a specific metrics Set.
func HandlerFor(set *metrics.Set) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", ContentType)
		set.WritePrometheus(w)
	})
}
