package metrics

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestWriteMetrics(t *testing.T) {
	t.Run("gauge_uint64", func(t *testing.T) {
		var bb bytes.Buffer

		WriteGaugeUint64(&bb, "foo", 123)
		sExpected := "foo 123\n"
		if s := bb.String(); s != sExpected {
			t.Fatalf("unexpected value; got\n%s\nwant\n%s", s, sExpected)
		}
	})
	t.Run("gauge_float64", func(t *testing.T) {
		var bb bytes.Buffer

		WriteGaugeFloat64(&bb, "foo", 1.23)
		sExpected := "foo 1.23\n"
		if s := bb.String(); s != sExpected {
			t.Fatalf("unexpected value; got\n%s\nwant\n%s", s, sExpected)
		}
	})
	t.Run("counter_uint64", func(t *testing.T) {
		var bb bytes.Buffer

		WriteCounterUint64(&bb, "foo_total", 123)
		sExpected := "foo_total 123\n"
		if s := bb.String(); s != sExpected {
			t.Fatalf("unexpected value; got\n%s\nwant\n%s", s, sExpected)
		}
	})
	t.Run("counter_float64", func(t *testing.T) {
		var bb bytes.Buffer

		WriteCounterFloat64(&bb, "foo_total", 1.23)
		sExpected := "foo_total 1.23\n"
		if s := bb.String(); s != sExpected {
			t.Fatalf("unexpected value; got\n%s\nwant\n%s", s, sExpected)
		}
	})
}

func TestRegisterUnregisterSet(t *testing.T) {
	const metricName = "metric_from_set"
	const metricValue = 123
	s := NewSet()
	c := s.NewCounter(metricName)
	c.Set(metricValue)

	RegisterSet(s)
	var bb bytes.Buffer
	WritePrometheus(&bb, false)
	data := bb.String()
	expectedLine := fmt.Sprintf("%s %d\n", metricName, metricValue)
	if !strings.Contains(data, expectedLine) {
		t.Fatalf("missing %q in\n%s", expectedLine, data)
	}

	UnregisterSet(s, true)
	bb.Reset()
	WritePrometheus(&bb, false)
	data = bb.String()
	if strings.Contains(data, expectedLine) {
		t.Fatalf("unepected %q in\n%s", expectedLine, data)
	}
}

func TestInvalidName(t *testing.T) {
	f := func(name string) {
		t.Helper()
		s := NewSet()
		expectPanic(t, fmt.Sprintf("NewCounter(%q)", name), func() { s.NewCounter(name) })
		expectPanic(t, fmt.Sprintf("NewGauge(%q)", name), func() { s.NewGauge(name, func() float64 { return 0 }) })
		expectPanic(t, fmt.Sprintf("NewSummary(%q)", name), func() { s.NewSummary(name) })
		expectPanic(t, fmt.Sprintf("GetOrCreateCounter(%q)", name), func() { s.GetOrCreateCounter(name) })
		expectPanic(t, fmt.Sprintf("GetOrCreateGauge(%q)", name), func() { s.GetOrCreateGauge(name, func() float64 { return 0 }) })
		expectPanic(t, fmt.Sprintf("GetOrCreateSummary(%q)", name), func() { s.GetOrCreateSummary(name) })
		expectPanic(t, fmt.Sprintf("GetOrCreateHistogram(%q)", name), func() { s.GetOrCreateHistogram(name) })
	}
	f("")
	f("foo{")
	f("foo}")
	f("foo{bar")
	f("foo{bar=")
	f(`foo{bar="`)
	f(`foo{bar="baz`)
	f(`foo{bar="baz"`)
	f(`foo{bar="baz",`)
	f(`foo{bar="baz",}`)
}

func TestDoubleRegister(t *testing.T) {
	t.Run("NewCounter", func(t *testing.T) {
		name := "NewCounterDoubleRegister"
		s := NewSet()
		s.NewCounter(name)
		expectPanic(t, name, func() { s.NewCounter(name) })
	})
	t.Run("NewGauge", func(t *testing.T) {
		name := "NewGaugeDoubleRegister"
		s := NewSet()
		s.NewGauge(name, func() float64 { return 0 })
		expectPanic(t, name, func() { s.NewGauge(name, func() float64 { return 0 }) })
	})
	t.Run("NewSummary", func(t *testing.T) {
		name := "NewSummaryDoubleRegister"
		s := NewSet()
		s.NewSummary(name)
		expectPanic(t, name, func() { s.NewSummary(name) })
	})
	t.Run("NewHistogram", func(t *testing.T) {
		name := "NewHistogramDoubleRegister"
		s := NewSet()
		s.NewHistogram(name)
		expectPanic(t, name, func() { s.NewSummary(name) })
	})
}

func TestGetOrCreateNotCounter(t *testing.T) {
	name := "GetOrCreateNotCounter"
	s := NewSet()
	s.NewSummary(name)
	expectPanic(t, name, func() { s.GetOrCreateCounter(name) })
}

func TestGetOrCreateNotGauge(t *testing.T) {
	name := "GetOrCreateNotGauge"
	s := NewSet()
	s.NewCounter(name)
	expectPanic(t, name, func() { s.GetOrCreateGauge(name, func() float64 { return 0 }) })
}

func TestGetOrCreateNotSummary(t *testing.T) {
	name := "GetOrCreateNotSummary"
	s := NewSet()
	s.NewCounter(name)
	expectPanic(t, name, func() { s.GetOrCreateSummary(name) })
}

func TestGetOrCreateNotHistogram(t *testing.T) {
	name := "GetOrCreateNotHistogram"
	s := NewSet()
	s.NewCounter(name)
	expectPanic(t, name, func() { s.GetOrCreateHistogram(name) })
}

func TestWritePrometheusSerial(t *testing.T) {
	if err := testWritePrometheus(); err != nil {
		t.Fatal(err)
	}
}

func TestWritePrometheusConcurrent(t *testing.T) {
	if err := testConcurrent(testWritePrometheus); err != nil {
		t.Fatal(err)
	}
}

func testWritePrometheus() error {
	var bb bytes.Buffer
	WritePrometheus(&bb, false)
	resultWithoutProcessMetrics := bb.String()
	bb.Reset()
	WritePrometheus(&bb, true)
	resultWithProcessMetrics := bb.String()
	if len(resultWithProcessMetrics) <= len(resultWithoutProcessMetrics) {
		return fmt.Errorf("result with process metrics must contain more data than the result without process metrics; got\n%q\nvs\n%q",
			resultWithProcessMetrics, resultWithoutProcessMetrics)
	}
	return nil
}

func expectPanic(t *testing.T, context string, f func()) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Fatalf("expecting panic in %s", context)
		}
	}()
	f()
}

func testConcurrent(f func() error) error {
	const concurrency = 5
	resultsCh := make(chan error, concurrency)
	for range concurrency {
		go func() {
			resultsCh <- f()
		}()
	}
	for range concurrency {
		select {
		case err := <-resultsCh:
			if err != nil {
				return fmt.Errorf("unexpected error: %s", err)
			}
		case <-time.After(time.Second * 5):
			return fmt.Errorf("timeout")
		}
	}
	return nil
}

func testMarshalTo(t *testing.T, m metric, prefix, resultExpected string) {
	t.Helper()
	var bb bytes.Buffer
	m.marshalTo(prefix, &bb)
	result := bb.String()
	if result != resultExpected {
		t.Fatalf("unexpected marshaled metric;\ngot\n%q\nwant\n%q", result, resultExpected)
	}
}
