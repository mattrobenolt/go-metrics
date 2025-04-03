package main

import (
	"net/http"
	"strconv"
	"time"

	"go.withmatt.com/metrics"
	"go.withmatt.com/metrics/promhttp"
)

var (
	currentTime = metrics.NewFloat64Func("current_time", func() float64 {
		return float64(time.Now().UnixNano()) / 1e9
	})
	ticksA          = metrics.NewUint64("tick", "variant", "a")
	ticksB          = metrics.NewUint64("tick", "variant", "b")
	requestDuration = metrics.NewHistogramVec("http_request_duration_seconds", "path", "code")
	responseSize    = metrics.NewHistogramVec("http_response_size_bytes", "path", "code")
)

type wrapper struct {
	w       http.ResponseWriter
	code    int
	written int
}

func (w *wrapper) Header() http.Header {
	return w.w.Header()
}

func (w *wrapper) WriteHeader(statusCode int) {
	w.code = statusCode
	w.w.WriteHeader(statusCode)
}

func (w *wrapper) Write(b []byte) (int, error) {
	n, err := w.w.Write(b)
	w.written += n
	return n, err
}

func observe(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := wrapper{w: w}
		next.ServeHTTP(&ww, r)

		var code string
		switch ww.code {
		case 0, 200:
			code = "200"
		default:
			code = strconv.Itoa(ww.code)
		}

		responseSize.WithLabelValues(
			r.URL.Path, code,
		).Update(float64(ww.written))

		requestDuration.WithLabelValues(
			r.URL.Path, code,
		).UpdateDuration(start)
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
