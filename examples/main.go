package main

import (
	"net/http"
	"time"

	"go.withmatt.com/metrics"
	"go.withmatt.com/metrics/promhttp"
)

var (
	currentTime = metrics.NewFloatFunc("current_time", func() float64 {
		return float64(time.Now().UnixNano()) / 1e9
	})
	ticksA       = metrics.NewCounter("tick", "variant", "a")
	ticksB       = metrics.NewCounter("tick", "variant", "b")
	httpRequests = metrics.NewHistogramVec("http_requests", "path")
)

func observe(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := httpRequests.WithLabelValues(r.URL.Path)
		start := time.Now()
		defer func() { h.UpdateDuration(start) }()
		next.ServeHTTP(w, r)
	})
}

func init() {
	metrics.RegisterDefaultCollectors()
}

func main() {
	go func() {
		for {
			<-time.After(10 * time.Millisecond)
			ticksA.Inc()
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	panic(http.ListenAndServe("127.0.0.1:9091", observe(mux)))
}
