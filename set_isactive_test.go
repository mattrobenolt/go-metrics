//go:build goexperiment.synctest

package metrics

import (
	"bytes"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"go.withmatt.com/metrics/internal/fasttime"
)

func TestSetVecTTL_GaugeWraparound(t *testing.T) {
	synctest.Run(func() {
		testClock := fasttime.NewClock(time.Millisecond)
		defer testClock.Stop()
		stubFastClock(t, testClock)

		set := NewSet()
		sv := set.NewSetVecWithTTL("id", time.Second)
		gauge := sv.NewUint64Vec("gauge")

		gauge.WithLabelValues("a").Inc()

		var buf bytes.Buffer
		set.WritePrometheus(&buf)
		if !strings.Contains(buf.String(), `gauge{id="a"} 1`) {
			t.Errorf("Expected gauge=1, got:\n%s", buf.String())
		}

		time.Sleep(3 * time.Second)

		gauge.WithLabelValues("a").Dec()

		buf.Reset()
		set.WritePrometheus(&buf)
		output := buf.String()

		if strings.Contains(output, "18446744073709551615") {
			t.Error("Gauge wrapped around to maxuint64")
		}
	})
}

func TestSetVecTTL_WithIsActive(t *testing.T) {
	synctest.Run(func() {
		testClock := fasttime.NewClock(time.Millisecond)
		defer testClock.Stop()
		stubFastClock(t, testClock)

		set := NewSet()
		sv := set.NewSetVecWithTTL("id", time.Second)

		sv.SetIsActive(func(s *Set) bool {
			val, _ := s.GetMetricUint64("gauge")
			return val > 0
		})

		gauge := sv.NewUint64Vec("gauge")
		gauge.WithLabelValues("a").Inc()

		var buf bytes.Buffer
		set.WritePrometheus(&buf)
		if !strings.Contains(buf.String(), `gauge{id="a"} 1`) {
			t.Errorf("Expected gauge=1, got:\n%s", buf.String())
		}

		time.Sleep(3 * time.Second)

		buf.Reset()
		set.WritePrometheus(&buf)
		if !strings.Contains(buf.String(), `gauge{id="a"} 1`) {
			t.Errorf("Expected Set to still exist, got:\n%s", buf.String())
		}

		gauge.WithLabelValues("a").Dec()

		buf.Reset()
		set.WritePrometheus(&buf)
		output := buf.String()

		if strings.Contains(output, "18446744073709551615") {
			t.Error("Gauge wrapped around")
		}
		if strings.Contains(output, `gauge{id="a"}`) {
			if !strings.Contains(output, `gauge{id="a"} 0`) {
				t.Errorf("Expected gauge=0, got:\n%s", output)
			}
		}

		time.Sleep(2 * time.Second)

		buf.Reset()
		set.WritePrometheus(&buf)
		if strings.Contains(buf.String(), `id="a"`) {
			t.Errorf("Expected Set to be expired, got:\n%s", buf.String())
		}
	})
}
