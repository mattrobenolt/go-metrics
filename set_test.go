package metrics

import (
	"testing"
)

type testCollector struct{}

func (testCollector) Collect(w ExpfmtWriter, constantTags string) {
	w.WriteMetricName(MetricName{
		Family:       MustIdent("collector1"),
		ConstantTags: constantTags,
	})
	w.WriteUint64(10)
}

func TestSet(t *testing.T) {
	set := NewSet()

	set.RegisterCollector(&testCollector{})

	set.NewCounter("counter1").Inc()
	set.NewCounter("counter2", "a", "1").Inc()
	set.NewGauge("gauge1", nil).Set(1)
	set.NewGauge("gauge2", nil, "a", "1").Set(1)
	set.NewHistogram("hist1").Update(1)
	set.NewHistogram("hist2", "a", "1").Update(1)

	s2 := set.NewSet()
	s2.NewCounter("counter3").Inc()

	assertMarshal(t, set, []string{
		`counter1 1`,
		`counter2{a="1"} 1`,
		`gauge1 1`,
		`gauge2{a="1"} 1`,
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
	set.NewGauge("gauge1", nil).Set(1)
	set.NewGauge("gauge2", nil, "a", "1").Set(1)
	set.NewHistogram("hist1").Update(1)
	set.NewHistogram("hist2", "a", "1").Update(1)

	s2 := set.NewSet()
	s2.NewCounter("counter3").Inc()

	s3 := set.NewSetOpt(SetOpt{
		ConstantTags: MustTags("x", "y"),
	})
	s3.NewCounter("counter4").Inc()
	s3.NewCounter("counter5", "a", "1").Inc()

	assertMarshal(t, set, []string{
		`counter1{foo="bar"} 1`,
		`counter2{foo="bar",a="1"} 1`,
		`gauge1{foo="bar"} 1`,
		`gauge2{foo="bar",a="1"} 1`,
		`hist1_bucket{vmrange="8.799e-01...1.000e+00",foo="bar"} 1`,
		`hist1_sum{foo="bar"} 1`,
		`hist1_count{foo="bar"} 1`,
		`hist2_bucket{vmrange="8.799e-01...1.000e+00",foo="bar",a="1"} 1`,
		`hist2_sum{foo="bar",a="1"} 1`,
		`hist2_count{foo="bar",a="1"} 1`,
		`counter3{foo="bar"} 1`,
		`counter4{foo="bar",x="y"} 1`,
		`counter5{foo="bar",x="y",a="1"} 1`,
		`collector1{foo="bar"} 10`,
	})
}
