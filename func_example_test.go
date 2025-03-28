package metrics_test

import (
	"fmt"
	"runtime"

	"go.withmatt.com/metrics"
)

func ExampleInt64Func() {
	// Define a function exporting the number of goroutines.
	f := metrics.NewInt64Func("goroutines_count", func() int64 {
		return int64(runtime.NumGoroutine())
	})

	// Obtain function value.
	fmt.Println(f.Get())
}

func ExampleUint64Func() {
	// Define a function exporting the number of goroutines.
	f := metrics.NewUint64Func("goroutines_count", func() uint64 {
		return uint64(runtime.NumGoroutine())
	})

	// Obtain function value.
	fmt.Println(f.Get())
}

func ExampleFloat64Func() {
	// Define a function exporting the number of goroutines.
	f := metrics.NewFloat64Func("goroutines_count", func() float64 {
		return float64(runtime.NumGoroutine())
	})

	// Obtain function value.
	fmt.Println(f.Get())
}
