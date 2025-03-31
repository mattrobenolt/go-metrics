package metrics

import (
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"go.withmatt.com/metrics/internal/assert"
	"go.withmatt.com/metrics/internal/fasttime"
)

type testCollector struct{}

func (testCollector) Collect(w ExpfmtWriter) {
	w.WriteLazyMetricUint64("collector1", 10)
}

func TestSet(t *testing.T) {
	set := NewSet()
	setvec := set.NewSetVec("setvec1")

	set.RegisterCollector(&testCollector{})

	set.NewCounter("counter1").Inc()
	setvec.NewCounter("counter1", "x").Inc()

	set.NewCounter("counter2", "a", "1").Inc()
	setvec.NewCounter("counter2", "x", "a", "1").Inc()

	set.NewUint64Func("gauge1", func() uint64 {
		return 1
	})
	set.NewUint64Func("gauge2", func() uint64 {
		return 2
	}, "a", "1")

	set.NewHistogram("hist1").Update(1)
	setvec.NewHistogram("hist1", "x").Update(1)

	set.NewHistogram("hist2", "a", "1").Update(1)
	setvec.NewHistogram("hist2", "x", "a", "1").Update(1)

	set.NewFixedHistogram("fixedhist1", []float64{0, 10}).Update(1)
	setvec.NewFixedHistogram("fixedhist1", []float64{0, 10}, "x").Update(1)

	set.NewFixedHistogram("fixedhist2", []float64{0, 10}, "a", "1").Update(1)
	setvec.NewFixedHistogram("fixedhist2", []float64{0, 10}, "x", "a", "1").Update(1)

	s2 := set.NewSet()
	s2.NewCounter("counter3").Inc()

	assertMarshalUnordered(t, set, []string{
		`counter1 1`,
		`counter1{setvec1="x"} 1`,
		`counter2{a="1"} 1`,
		`counter2{setvec1="x",a="1"} 1`,
		`gauge1 1`,
		`gauge2{a="1"} 2`,
		`hist1_bucket{vmrange="8.799e-01...1.000e+00"} 1`,
		`hist1_sum 1`,
		`hist1_count 1`,
		`hist1_bucket{vmrange="8.799e-01...1.000e+00",setvec1="x"} 1`,
		`hist1_sum{setvec1="x"} 1`,
		`hist1_count{setvec1="x"} 1`,
		`hist2_bucket{vmrange="8.799e-01...1.000e+00",a="1"} 1`,
		`hist2_sum{a="1"} 1`,
		`hist2_count{a="1"} 1`,
		`hist2_bucket{vmrange="8.799e-01...1.000e+00",setvec1="x",a="1"} 1`,
		`hist2_sum{setvec1="x",a="1"} 1`,
		`hist2_count{setvec1="x",a="1"} 1`,
		`fixedhist1_bucket{le="0"} 0`,
		`fixedhist1_bucket{le="10"} 1`,
		`fixedhist1_bucket{le="+Inf"} 1`,
		`fixedhist1_count 1`,
		`fixedhist1_sum 1`,
		`fixedhist1_bucket{le="0",setvec1="x"} 0`,
		`fixedhist1_bucket{le="10",setvec1="x"} 1`,
		`fixedhist1_bucket{le="+Inf",setvec1="x"} 1`,
		`fixedhist1_count{setvec1="x"} 1`,
		`fixedhist1_sum{setvec1="x"} 1`,
		`fixedhist2_bucket{le="0",a="1"} 0`,
		`fixedhist2_bucket{le="10",a="1"} 1`,
		`fixedhist2_bucket{le="+Inf",a="1"} 1`,
		`fixedhist2_count{a="1"} 1`,
		`fixedhist2_sum{a="1"} 1`,
		`fixedhist2_bucket{le="0",setvec1="x",a="1"} 0`,
		`fixedhist2_bucket{le="10",setvec1="x",a="1"} 1`,
		`fixedhist2_bucket{le="+Inf",setvec1="x",a="1"} 1`,
		`fixedhist2_count{setvec1="x",a="1"} 1`,
		`fixedhist2_sum{setvec1="x",a="1"} 1`,
		`counter3 1`,
		`collector1 10`,
	})
}

func TestSetConstantTags(t *testing.T) {
	set := NewSet("foo", "bar")

	// invalid label pairs
	assert.Panics(t, func() { NewSet("foo", "bar", "baz") })

	set.RegisterCollector(&testCollector{})

	set.NewCounter("counter1").Inc()
	set.NewCounter("counter2", "a", "1").Inc()
	set.NewUint64Func("gauge1", func() uint64 {
		return 1
	})
	set.NewUint64Func("gauge2", func() uint64 {
		return 2
	}, "a", "1")
	set.NewHistogram("hist1").Update(1)
	set.NewHistogram("hist2", "a", "1").Update(1)

	s2 := set.NewSet()
	s2.NewCounter("counter3").Inc()

	s3 := set.NewSet("x", "y")
	s3.NewCounter("counter4").Inc()
	s3.NewCounter("counter5", "a", "1").Inc()

	s4 := s3.NewSet("i", "j")
	s4.NewCounter("counter6", "z", "10").Inc()

	// duplicate
	assert.Panics(t, func() { s3.NewSet("i", "j") })

	assertMarshalUnordered(t, set, []string{
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

func TestNewSetVecConcurrent(t *testing.T) {
	const n = 100
	set := NewSet()
	hammer(t, n, func(i int) {
		set.NewSetVec(fmt.Sprintf("vec%d", i)).NewCounter("foo", "x").Set(uint64(i))
	})

	expected := make([]string, n)
	for i := range n {
		expected[i] = fmt.Sprintf(`foo{vec%d="x"} %d`, i, i)
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
	set.NewUint64Func("gauge1", func() uint64 {
		return 1
	})
	set.NewUint64Func("gauge2", func() uint64 {
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
		set.NewUint64Func(fmt.Sprintf("gauge%d", i), func() uint64 {
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

func TestSetVec(t *testing.T) {
	set := NewSet()
	sv := set.NewSetVec("a")

	sv.WithLabelValue("1").NewCounter("foo").Inc()
	sv.WithLabelValue("2").NewCounter("foo").Inc()

	assertMarshalUnordered(t, set, []string{
		`foo{a="1"} 1`,
		`foo{a="2"} 1`,
	})

	sv.RemoveByLabelValue("1")
	assertMarshalUnordered(t, set, []string{
		`foo{a="2"} 1`,
	})
	sv.RemoveByLabelValue("2")
	assertMarshalUnordered(t, set, nil)

	// should not fail
	sv.RemoveByLabelValue("xxx")

	// carry over constant labels
	set = NewSet("x", "y")
	sv = set.NewSetVec("a")
	sv.WithLabelValue("1").NewCounter("foo").Inc()
	sv.WithLabelValue("2").NewCounter("foo").Inc()

	assertMarshalUnordered(t, set, []string{
		`foo{x="y",a="1"} 1`,
		`foo{x="y",a="2"} 1`,
	})

	set = NewSet()
	sv = set.NewSetVec("type")
	sv.NewUint64Vec("foo", "label1").WithLabelValues("uint64", "value1").Inc()
	sv.NewInt64Vec("foo", "label1").WithLabelValues("int64", "value1").Inc()
	sv.NewFloat64Vec("foo", "label1").WithLabelValues("float64", "value1").Inc()
	sv.NewHistogramVec("foo", "label1").WithLabelValues("hist", "value1").Update(1)
	sv.NewFixedHistogramVec("foo", []float64{0, 10}, "label1").WithLabelValues("fixedhist", "value1").Update(1)

	assertMarshalUnordered(t, set, []string{
		`foo{type="uint64",label1="value1"} 1`,
		`foo{type="int64",label1="value1"} 1`,
		`foo{type="float64",label1="value1"} 1`,
		`foo_bucket{vmrange="8.799e-01...1.000e+00",type="hist",label1="value1"} 1`,
		`foo_count{type="hist",label1="value1"} 1`,
		`foo_sum{type="hist",label1="value1"} 1`,
		`foo_bucket{le="0",type="fixedhist",label1="value1"} 0`,
		`foo_bucket{le="10",type="fixedhist",label1="value1"} 1`,
		`foo_bucket{le="+Inf",type="fixedhist",label1="value1"} 1`,
		`foo_count{type="fixedhist",label1="value1"} 1`,
		`foo_sum{type="fixedhist",label1="value1"} 1`,
	})
}

func stubFastClock(t *testing.T) {
	originalFastClock := fastClock
	testClock := fasttime.NewClock(time.Microsecond)
	t.Cleanup(func() {
		testClock.Stop()
		fastClock = originalFastClock
	})
	fastClock = sync.OnceValue(func() *fasttime.Clock { return testClock })
}

func TestSetVecTTL(t *testing.T) {
	stubFastClock(t)

	set := NewSet()
	sv := set.NewSetVecWithTTL("a", time.Millisecond)

	sv.WithLabelValue("1").NewCounter("foo").Inc()
	sv.WithLabelValue("2").NewCounter("foo").Inc()
	time.Sleep(500 * time.Microsecond)

	// keeps "1" alive
	sv.WithLabelValue("1").NewCounter("bar").Inc()

	assertMarshalUnordered(t, set, []string{
		`foo{a="1"} 1`,
		`foo{a="2"} 1`,
		`bar{a="1"} 1`,
	})

	// "2" expired away by now
	time.Sleep(750 * time.Microsecond)

	assertMarshalUnordered(t, set, []string{
		`foo{a="1"} 1`,
		`bar{a="1"} 1`,
	})

	time.Sleep(500 * time.Microsecond)
	assertMarshalUnordered(t, set, []string{})
}
