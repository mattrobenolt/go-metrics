# metrics

[![Go Reference](https://pkg.go.dev/badge/go.withmatt.com/metrics.svg)](https://pkg.go.dev/go.withmatt.com/metrics)

_Extremely_ fast and lightweight package for recording and exporting metrics in Prometheus exposition format.

> Heavily inspired and based on [VictoriaMetrics/metrics](https://github.com/VictoriaMetrics/metrics).

```go
import "go.withmatt.com/metrics"
```

## Features
* Very fast, very few allocations. [Really](benchmarks.txt).
* Optional expiring of unobserved metrics
* HTTP exporter
* Built-in runtime metrics collectors
* Easy Prometheus-like API
* No dependencies

## Quick Start

```go
import (
	"net/http"
	"go.withmatt.com/metrics"
)

func main() {
	// exposes Go and process runtime metrics
	metrics.RegisterDefaultCollectors()

	// create a new uint64 counter "foo" with the tag a=b
	// This will emit:
	//   foo{a="b"} 1
	c := metrics.NewUint("foo", "a", "b")
	c.Inc()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	panic(http.ListenAndServe("127.0.0.1:9091", mux))
}
```

## Metric Types
* Counters (uint64/int64/float64)
* Gauges (uint64/int64/float64)
* Histograms
  - Prometheus-like (`le` label style)
  - [VictoriaMetrics-like](https://medium.com/@valyala/improving-histogram-usability-for-prometheus-and-grafana-bc7e5df0e350) (`vmrange` label style)

> [!NOTE]
> Summary type has not been implemented.

## Performance

- `mattware` = this library
- `vm` = [`github.com/VictoriaMetrics/metrics`](https://pkg.go.dev/github.com/VictoriaMetrics/metrics)
- `prom` = [`github.com/prometheus/client_golang/prometheus`](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus)

### Updating metrics

> [!NOTE]
> Updating metrics happen very typically within extremely hot paths, and
> nanoseconds matter.

Increment a counter through `UintVec.WithLabelValue` API

| package | sec/op | vs base | allocs/op |
| :------ | :----: | :-----: | :-------: |
| mattware | 27.54n | +0% | 0
| vm * | 95.75n | +247.76% | 1
| prom | 47.36n | +72.02 | 0

> * VM/metrics doesn't support this API natively, and their pattern for this is highly discouraged.

Update Prometheus-like histogram (`le` label style)

| package | sec/op | vs base | allocs/op |
| :------ | :----: | :-----: | :-------: |
| mattware | 8.127n | +0% | 0
| vm * | - | - | -
| prom | 10.035n | +23.48% | 0

> * VM/metrics does not support this.

Update VictoriaMetrics-like histogram (`vmrange` label style)

| package | sec/op | vs base | allocs/op |
| :------ | :----: | :-----: | :-------: |
| mattware | 15.91n | +0% | 0
| vm | 16.98n | +6.72% | 0
| prom * | - | - | -

> * Prom client does not support this.

### Exporting metrics

> [!NOTE]
> Exporting happens by Prometheus compatible scrapers typically.

Exporting 10,000 counters

| package | sec/op | vs base | allocs/op | B/s |
| :------ | :----: | :-----: | :-------: | :-: |
| mattware | 704.2µ | +0% | 1 | 689.13Mi
| vm | 1033.3µ | +46.72% | 10018 | 478.91Mi
| prom | 10068.0µ | +1329.60% | 41422 | 48.23Mi

Exporting 100 Prometheus-like histograms, with 100,000 observations each

| package | sec/op | vs base | allocs/op | B/s |
| :------ | :----: | :-----: | :-------: | :-: |
| mattware | 108.2µ | +0% | 1 | 811.6Mi
| prom | 437.4µ | +304.36% | 5434 | 200.6Mi

Exporting 100 VictoriaMetrics-like histograms, with 100,000 observations each

| package | sec/op | vs base | allocs/op | B/s |
| :------ | :----: | :-----: | :-------: | :-: |
| mattware | 462.8µ | +0% | 1 | 1069.8Mi
| vm | 2373.2µ | +412.83% | 41422 | 208.6Mi
