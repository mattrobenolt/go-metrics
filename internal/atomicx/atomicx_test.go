package atomicx_test

import (
	"math"
	"runtime"
	"testing"

	"go.withmatt.com/metrics/internal/assert"
	. "go.withmatt.com/metrics/internal/atomicx"
)

func TestFloat64(t *testing.T) {
	var f Float64
	v := float64(1.88)
	assert.Equal(t, f.Load(), 0)
	f.Store(v)
	assert.Equal(t, f.Load(), v)
	f.Add(v)
	assert.Equal(t, f.Load(), v+v)
}

func TestSum(t *testing.T) {
	var s Sum
	assert.Equal(t, s.Load(), 0)
	s.Add(1.88)
	assert.Equal(t, s.Load(), 1.88)
	s.Add(1)
	assert.Equal(t, s.Load(), 2.88)
	s.Reset()
	assert.Equal(t, s.Load(), 0)
}

func TestHammerFloatAdd(t *testing.T) {
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

func TestHammerSumAdd(t *testing.T) {
	const p = 4
	n := 10000
	if testing.Short() {
		n = 1000
	}
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(p))

	const a = 1.1
	const b = 1

	c := make(chan int)
	var val Sum
	for range p {
		go func() {
			defer func() {
				assert.Nil(t, recover())
				c <- 1
			}()
			for range n {
				val.Add(a)
				val.Add(b)
				val.AddUint64(b)
			}
		}()
	}
	for range p {
		<-c
	}
	// XXX: Floating point precision, so need to round
	assert.Equal(t, math.Round(val.Load()), math.Round(float64(n)*(p*a+p*b*2)))
}
