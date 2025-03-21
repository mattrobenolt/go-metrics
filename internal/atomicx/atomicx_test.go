package atomicx_test

import (
	"runtime"
	"testing"

	"go.withmatt.com/metrics/internal/assert"
	. "go.withmatt.com/metrics/internal/atomicx"
)

func TestFloat64(t *testing.T) {
	var f Float64
	v := float64(1.88)
	assert.Equal(t, 0, f.Load())
	f.Store(v)
	assert.Equal(t, v, f.Load())
	f.Add(v)
	assert.Equal(t, v+v, f.Load())
}

func TestHammerAdd(t *testing.T) {
	const p = 4
	n := 100000
	if testing.Short() {
		n = 1000
	}
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(p))

	c := make(chan int)
	var val Float64
	for range p {
		go func() {
			defer func() {
				assert.Nil(t, recover())
				c <- 1
			}()
			for range n {
				val.Add(1)
			}
		}()
	}
	for range p {
		<-c
	}
	assert.Equal(t, val.Load(), float64(n)*p)
}
