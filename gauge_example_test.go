package metrics_test

import (
	"fmt"
	"runtime"
	"strconv"

	"go.withmatt.com/metrics"
)

func ExampleGauge() {
	// Define a gauge exporting the number of goroutines.
	g := metrics.NewGauge("goroutines_count", func() float64 {
		return float64(runtime.NumGoroutine())
	})

	// Obtain gauge value.
	fmt.Println(g.Get())
}

func ExampleGaugeVec() {
	metricGauge := metrics.NewGaugeVec(metrics.GaugeVecOpt{
		Name: metrics.VecName{
			Family: "metric",
			Labels: []string{"label1", "label2"},
		},
	})
	for i := range 3 {
		// Dynamically construct metric name with label values
		metricGauge.WithLabelValues(
			"value1", strconv.Itoa(i),
		).Set(float64(i + 1))
	}

	// Read gauge values.
	for i := range 3 {
		n := metricGauge.WithLabelValues(
			"value1", strconv.Itoa(i),
		).Get()
		fmt.Println(n)
	}

	// Output:
	// 1
	// 2
	// 3
}
