package metrics_test

import (
	"fmt"
	"time"

	"go.withmatt.com/metrics"
)

func ExampleSetVec_SetIsActive() {
	sv := metrics.NewSetVecWithTTL("id", time.Hour)

	sv.SetIsActive(func(s *metrics.Set) bool {
		active, _ := s.GetMetricUint64("active")
		return active > 0
	})

	active := sv.NewUint64Vec("active")
	counter := sv.NewUint64Vec("counter")

	active.WithLabelValues("a").Inc()
	counter.WithLabelValues("a").Add(100)

	fmt.Println(active.WithLabelValues("a").Get())
	fmt.Println(counter.WithLabelValues("a").Get())

	// Output:
	// 1
	// 100
}

func ExampleSetVec_SetIsActive_multipleConditions() {
	sv := metrics.NewSetVecWithTTL("id", 30*time.Minute)

	sv.SetIsActive(func(s *metrics.Set) bool {
		a, _ := s.GetMetricUint64("gauge_a")
		b, _ := s.GetMetricUint64("gauge_b")
		return a > 0 || b > 0
	})

	gaugeA := sv.NewUint64Vec("gauge_a")
	gaugeB := sv.NewUint64Vec("gauge_b")

	gaugeA.WithLabelValues("x").Inc()
	gaugeB.WithLabelValues("x").Add(5)

	fmt.Println(gaugeA.WithLabelValues("x").Get())
	fmt.Println(gaugeB.WithLabelValues("x").Get())

	// Output:
	// 1
	// 5
}
