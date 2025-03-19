package metrics

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestSummarySerial(t *testing.T) {
	name := `TestSummarySerial`
	set := NewSet()
	RegisterSet(set)
	defer UnregisterSet(set, true)

	s := set.NewSummary(name)

	// Verify that the summary isn't visible in the output of WritePrometheus,
	// since it doesn't contain any values yet.
	var bb bytes.Buffer
	WritePrometheus(&bb, false)
	result := bb.String()
	if strings.Contains(result, name) {
		t.Fatalf("summary %s shouldn't be visible in the WritePrometheus output; got\n%s", name, result)
	}

	// Write data to summary
	for i := range 2000 {
		s.Update(float64(i))
		t := time.Now()
		s.UpdateDuration(t.Add(-time.Millisecond * time.Duration(i)))
	}

	// Make sure the summary prints <prefix>_sum and <prefix>_count on marshalTo call
	testMarshalTo(t, s, "prefix", fmt.Sprintf("prefix_sum %g\nprefix_count %d\n", s.sum, s.count))
	testMarshalTo(t, s, `m{foo="bar"}`, fmt.Sprintf("m_sum{foo=\"bar\"} %g\nm_count{foo=\"bar\"} %d\n", s.sum, s.count))

	// Verify s.quantileValues
	s.updateQuantiles()
	if s.quantileValues[len(s.quantileValues)-1] != 1999 {
		t.Fatalf("unexpected quantileValues[last]; got %v; want %v", s.quantileValues[len(s.quantileValues)-1], 1999)
	}

	// Make sure the summary becomes visible in the output of WritePrometheus,
	// since now it contains values.
	bb.Reset()
	WritePrometheus(&bb, false)
	result = bb.String()
	if !strings.Contains(result, name) {
		t.Fatalf("missing summary %s in the WritePrometheus output; got\n%s", name, result)
	}
}

func TestSummaryConcurrent(t *testing.T) {
	name := "SummaryConcurrent"
	s := NewSet().NewSummary(name)
	err := testConcurrent(func() error {
		for i := range 10 {
			s.Update(float64(i))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	testMarshalTo(t, s, "prefix", "prefix_sum 225\nprefix_count 50\n")
}

func TestSummaryWithTags(t *testing.T) {
	name := `TestSummary{tag="foo"}`
	set := NewSet()
	RegisterSet(set)
	defer UnregisterSet(set, true)

	s := set.NewSummary(name)
	s.Update(123)

	var bb bytes.Buffer
	WritePrometheus(&bb, false)
	result := bb.String()
	namePrefixWithTag := `TestSummary{tag="foo",quantile="`
	if !strings.Contains(result, namePrefixWithTag) {
		t.Fatalf("missing summary prefix %s in the WritePrometheus output; got\n%s", namePrefixWithTag, result)
	}
}

func TestSummaryWithEmptyTags(t *testing.T) {
	name := `TestSummary{}`
	set := NewSet()
	RegisterSet(set)
	defer UnregisterSet(set, true)

	s := set.NewSummary(name)
	s.Update(123)

	var bb bytes.Buffer
	WritePrometheus(&bb, false)
	result := bb.String()
	namePrefixWithTag := `TestSummary{quantile="`
	if !strings.Contains(result, namePrefixWithTag) {
		t.Fatalf("missing summary prefix %s in the WritePrometheus output; got\n%s", namePrefixWithTag, result)
	}
}

func TestSummaryInvalidQuantiles(t *testing.T) {
	name := "SummaryInvalidQuantiles"
	expectPanic(t, name, func() {
		NewSet().NewSummaryExt(name, time.Minute, []float64{123, -234})
	})
}

func TestSummarySmallWindow(t *testing.T) {
	name := "SummarySmallWindow"
	window := time.Millisecond * 20
	quantiles := []float64{0.1, 0.2, 0.3}
	set := NewSet()
	RegisterSet(set)
	defer UnregisterSet(set, true)

	s := set.NewSummaryExt(name, window, quantiles)
	for range 2000 {
		s.Update(123)
	}
	// Wait for window update and verify that the summary has been cleared.
	time.Sleep(2 * window)
	var bb bytes.Buffer
	WritePrometheus(&bb, false)
	result := bb.String()
	// <name>_sum and <name>_count are present in the output.
	// Only <name>{quantile} shouldn't be present.
	name += "{"
	if strings.Contains(result, name) {
		t.Fatalf("summary %s cannot be present in the WritePrometheus output; got\n%s", name, result)
	}
}

func TestGetOrCreateSummaryInvalidWindow(t *testing.T) {
	name := "GetOrCreateSummaryInvalidWindow"
	s := NewSet()
	s.GetOrCreateSummaryExt(name, defaultSummaryWindow, defaultSummaryQuantiles)
	expectPanic(t, name, func() {
		s.GetOrCreateSummaryExt(name, defaultSummaryWindow/2, defaultSummaryQuantiles)
	})
}

func TestGetOrCreateSummaryInvalidQuantiles(t *testing.T) {
	name := "GetOrCreateSummaryInvalidQuantiles"
	s := NewSet()
	s.GetOrCreateSummaryExt(name, defaultSummaryWindow, defaultSummaryQuantiles)
	expectPanic(t, name, func() {
		s.GetOrCreateSummaryExt(name, defaultSummaryWindow, []float64{0.1, 0.2})
	})
	quantiles := slices.Clone(defaultSummaryQuantiles)
	quantiles[len(quantiles)-1] /= 2
	expectPanic(t, name, func() {
		s.GetOrCreateSummaryExt(name, defaultSummaryWindow, quantiles)
	})
}

func TestGetOrCreateSummarySerial(t *testing.T) {
	name := "GetOrCreateSummarySerial"
	if err := testGetOrCreateSummary(name); err != nil {
		t.Fatal(err)
	}
}

func TestGetOrCreateSummaryConcurrent(t *testing.T) {
	name := "GetOrCreateSummaryConcurrent"
	err := testConcurrent(func() error {
		return testGetOrCreateSummary(name)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func testGetOrCreateSummary(name string) error {
	s := NewSet()
	s1 := s.GetOrCreateSummary(name)
	for range 10 {
		s2 := s.GetOrCreateSummary(name)
		if s1 != s2 {
			return fmt.Errorf("unexpected summary returned; got %p; want %p", s2, s1)
		}
	}
	return nil
}
