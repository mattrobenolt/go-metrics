//go:build goexperiment.synctest

package metrics

import (
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"go.withmatt.com/metrics/internal/fasttime"
)

func stubFastClock(t *testing.T, testClock *fasttime.Clock) {
	originalFastClock := fastClock
	t.Cleanup(func() {
		fastClock = originalFastClock
	})
	fastClock = sync.OnceValue(func() *fasttime.Clock { return testClock })
}

func TestSetVecTTL(t *testing.T) {
	synctest.Run(func() {
		testClock := fasttime.NewClock(time.Millisecond)
		defer testClock.Stop()

		stubFastClock(t, testClock)

		set := NewSet()
		sv := set.NewSetVecWithTTL("a", time.Second)

		sv.WithLabelValue("1").NewCounter("foo").Inc()
		sv.WithLabelValue("2").NewCounter("foo").Inc()

		time.Sleep(750 * time.Millisecond)

		// keeps "1" alive
		sv.WithLabelValue("1").NewCounter("bar").Inc()

		assertMarshalUnordered(t, set, []string{
			`foo{a="1"} 1`,
			`foo{a="2"} 1`,
			`bar{a="1"} 1`,
		})

		// "2" expired away by now
		time.Sleep(500 * time.Millisecond)

		assertMarshalUnordered(t, set, []string{
			`foo{a="1"} 1`,
			`bar{a="1"} 1`,
		})

		time.Sleep(time.Second)
		assertMarshalUnordered(t, set, []string{})
	})
}
