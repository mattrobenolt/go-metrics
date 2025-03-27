package metrics

import (
	"fmt"
	"io"
	"testing"
)

type testCollector struct{}

func (testCollector) Collect(w ExpfmtWriter) {
	w.WriteLazyMetricUint64("collector1", 10)
}

func TestSet(t *testing.T) {
	set := NewSet()

	set.RegisterCollector(&testCollector{})

	set.NewCounter("counter1").Inc()
	set.NewCounter("counter2", "a", "1").Inc()
	set.NewUintFunc("gauge1", func() uint64 {
		return 1
	})
	set.NewUintFunc("gauge2", func() uint64 {
		return 2
	}, "a", "1")
	set.NewHistogram("hist1").Update(1)
	set.NewHistogram("hist2", "a", "1").Update(1)

	s2 := set.NewSet()
	s2.NewCounter("counter3").Inc()

	assertMarshal(t, set, []string{
		`counter1 1`,
		`counter2{a="1"} 1`,
		`gauge1 1`,
		`gauge2{a="1"} 2`,
		`hist1_bucket{vmrange="8.799e-01...1.000e+00"} 1`,
		`hist1_sum 1`,
		`hist1_count 1`,
		`hist2_bucket{vmrange="8.799e-01...1.000e+00",a="1"} 1`,
		`hist2_sum{a="1"} 1`,
		`hist2_count{a="1"} 1`,
		`counter3 1`,
		`collector1 10`,
	})
}

func TestSetConstantTags(t *testing.T) {
	set := NewSetOpt(SetOpt{
		ConstantTags: MustTags("foo", "bar"),
	})

	set.RegisterCollector(&testCollector{})

	set.NewCounter("counter1").Inc()
	set.NewCounter("counter2", "a", "1").Inc()
	set.NewUintFunc("gauge1", func() uint64 {
		return 1
	})
	set.NewUintFunc("gauge2", func() uint64 {
		return 2
	}, "a", "1")
	set.NewHistogram("hist1").Update(1)
	set.NewHistogram("hist2", "a", "1").Update(1)

	s2 := set.NewSet()
	s2.NewCounter("counter3").Inc()

	s3 := set.NewSetOpt(SetOpt{
		ConstantTags: MustTags("x", "y"),
	})
	s3.NewCounter("counter4").Inc()
	s3.NewCounter("counter5", "a", "1").Inc()

	s4 := s3.NewSetOpt(SetOpt{
		ConstantTags: MustTags("i", "j"),
	})
	s4.NewCounter("counter6", "z", "10").Inc()

	assertMarshal(t, set, []string{
		`counter1{foo="bar"} 1`,
		`counter2{foo="bar",a="1"} 1`,
		`gauge1{foo="bar"} 1`,
		`gauge2{foo="bar",a="1"} 2`,
		`hist1_bucket{vmrange="8.799e-01...1.000e+00",foo="bar"} 1`,
		`hist1_sum{foo="bar"} 1`,
		`hist1_count{foo="bar"} 1`,
		`hist2_bucket{vmrange="8.799e-01...1.000e+00",foo="bar",a="1"} 1`,
		`hist2_sum{foo="bar",a="1"} 1`,
		`hist2_count{foo="bar",a="1"} 1`,
		`counter3{foo="bar"} 1`,
		`counter4{foo="bar",x="y"} 1`,
		`counter5{foo="bar",x="y",a="1"} 1`,
		`counter6{foo="bar",x="y",i="j",z="10"} 1`,
		`collector1{foo="bar"} 10`,
	})
}

func TestNewSetConcurrent(t *testing.T) {
	const n = 100
	set := NewSet()
	hammer(t, n, func(i int) {
		set.NewSet().NewCounter(fmt.Sprintf("counter%d", i)).Set(uint64(i))
	})

	expected := make([]string, n)
	for i := range n {
		expected[i] = fmt.Sprintf("counter%d %d", i, i)
	}
	assertMarshalUnordered(t, set, expected)
}

func TestSetUnregister(t *testing.T) {
	set := NewSet()
	s1 := set.NewSet()
	s1.NewCounter("counter1").Inc()

	assertMarshalUnordered(t, set, []string{"counter1 1"})

	s2 := set.NewSet()
	s2.NewCounter("counter2").Inc()
	assertMarshalUnordered(t, set, []string{
		"counter1 1",
		"counter2 1",
	})

	set.UnregisterSet(s1)
	assertMarshalUnordered(t, set, []string{
		"counter2 1",
	})

	set.UnregisterSet(s2)
	assertMarshalUnordered(t, set, nil)
}

func TestSetUnregisterConcurrent(t *testing.T) {
	const n = 1000
	const inner = 10

	set := NewSet()
	hammer(t, n, func(_ int) {
		for range inner {
			s := set.NewSet()
			s.NewCounter("counter1").Inc()
			set.UnregisterSet(s)
		}
	})
	assertMarshalUnordered(t, set, nil)
}

func TestSetMarshalConcurrent(t *testing.T) {
	set := NewSet()

	set.RegisterCollector(&testCollector{})

	set.NewCounter("counter1").Inc()
	set.NewCounter("counter2", "a", "1").Inc()
	set.NewUintFunc("gauge1", func() uint64 {
		return 1
	})
	set.NewUintFunc("gauge2", func() uint64 {
		return 2
	}, "a", "1")
	set.NewHistogram("hist1").Update(1)
	set.NewHistogram("hist2", "a", "1").Update(1)

	s2 := set.NewSet()
	s2.NewCounter("counter3").Inc()

	hammer(t, 1000, func(_ int) {
		assertMarshal(t, set, []string{
			`counter1 1`,
			`counter2{a="1"} 1`,
			`gauge1 1`,
			`gauge2{a="1"} 2`,
			`hist1_bucket{vmrange="8.799e-01...1.000e+00"} 1`,
			`hist1_sum 1`,
			`hist1_count 1`,
			`hist2_bucket{vmrange="8.799e-01...1.000e+00",a="1"} 1`,
			`hist2_sum{a="1"} 1`,
			`hist2_count{a="1"} 1`,
			`counter3 1`,
			`collector1 10`,
		})
	})
}

func TestSetConcurrent(t *testing.T) {
	const n = 1000
	set := NewSet()

	hammer(t, n, func(i int) {
		set.NewCounter(fmt.Sprintf("counter%d", i)).Inc()
		set.NewUintFunc(fmt.Sprintf("gauge%d", i), func() uint64 {
			return 1
		})
		set.NewHistogram(fmt.Sprintf("hist%d", i)).Update(1)

		s2 := set.NewSet()
		s2.NewCounter(fmt.Sprintf("subcounter%d", i)).Inc()

		set.WritePrometheusUnthrottled(io.Discard)
	})

	var expected []string
	for i := range n {
		expected = append(expected, fmt.Sprintf("counter%d 1", i))
		expected = append(expected, fmt.Sprintf("gauge%d 1", i))
		expected = append(expected, fmt.Sprintf("subcounter%d 1", i))
		expected = append(expected, fmt.Sprintf(`hist%d_bucket{vmrange="8.799e-01...1.000e+00"} 1`, i))
		expected = append(expected, fmt.Sprintf("hist%d_sum 1", i))
		expected = append(expected, fmt.Sprintf("hist%d_count 1", i))
	}

	assertMarshalUnordered(t, set, expected)
}
