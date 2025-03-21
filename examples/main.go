package main

import (
	"net/http"
	"time"

	"go.withmatt.com/metrics"
	"go.withmatt.com/metrics/promhttp"
)

var global = metrics.NewSet()

var (
	currentTime = global.NewGauge("current_time", func() float64 {
		return float64(time.Now().UTC().Unix())
	})
	ticksA = global.NewCounter("tick", "variant", "a")
	ticksB = global.NewCounter("tick", "variant", "b")
)

func observe(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := global.GetOrCreateHistogram("http_request", "path", r.URL.Path)
		start := time.Now()
		defer func() { h.UpdateDuration(start) }()
		next.ServeHTTP(w, r)
	})
}

func main() {
	go func() {
		for {
			<-time.After(10 * time.Millisecond)
			ticksA.Inc()
		}
	}()
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler(global))
	panic(http.ListenAndServe("127.0.0.1:9091", observe(mux)))
}
