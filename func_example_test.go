package metrics_test

import (
	"fmt"
	"runtime"

	"go.withmatt.com/metrics"
)

func ExampleIntFunc() {
	// Define a function exporting the number of goroutines.
	f := metrics.NewIntFunc("goroutines_count", func() int64 {
		return int64(runtime.NumGoroutine())
	})

	// Obtain function value.
	fmt.Println(f.Get())
}

func ExampleUintFunc() {
	// Define a function exporting the number of goroutines.
	f := metrics.NewUintFunc("goroutines_count", func() uint64 {
		return uint64(runtime.NumGoroutine())
	})

	// Obtain function value.
	fmt.Println(f.Get())
}

func ExampleFloatFunc() {
	// Define a function exporting the number of goroutines.
	f := metrics.NewFloatFunc("goroutines_count", func() float64 {
		return float64(runtime.NumGoroutine())
	})

	// Obtain function value.
	fmt.Println(f.Get())
}
