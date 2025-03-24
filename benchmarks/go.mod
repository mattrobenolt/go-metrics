module x

go 1.24.0

require (
	github.com/VictoriaMetrics/metrics v1.35.2
	github.com/prometheus/client_golang v1.21.1
	github.com/prometheus/common v0.62.0
	go.withmatt.com/metrics v0.0.0-00010101000000-000000000000
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/valyala/fastrand v1.1.0 // indirect
	github.com/valyala/histogram v1.2.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	google.golang.org/protobuf v1.36.1 // indirect
)

replace go.withmatt.com/metrics => ../
