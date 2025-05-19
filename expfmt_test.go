package metrics

import (
	"bytes"
	"math"
	"strings"
	"testing"
	"time"

	"go.withmatt.com/metrics/internal/assert"
)

func TestWriteMetricsUint64(t *testing.T) {
	w := ExpfmtWriter{
		b: bytes.NewBuffer(nil),
	}
	for _, tc := range []struct {
		family       string
		tags         []string
		constantTags string
		value        uint64
		expected     string
	}{
		{"foo", nil, "", 0, "foo 0"},
		{"foo", nil, "", 1, "foo 1"},
		{"foo", nil, `x="y"`, 1, `foo{x="y"} 1`},
		{"foo", []string{"a", "1"}, "", 10, `foo{a="1"} 10`},
		{"foo", []string{"a", "1", "b", "2"}, "", 10, `foo{a="1",b="2"} 10`},
		{"foo", []string{"a", "1", "b", "2"}, `x="y"`, 10, `foo{x="y",a="1",b="2"} 10`},
	} {
		w.b.Reset()
		w.constantTags = tc.constantTags
		w.WriteMetricName(MetricName{
			Family: MustIdent(tc.family),
			Tags:   MustTags(tc.tags...),
		})
		w.WriteUint64(tc.value)
		assert.Equal(t, tc.expected+"\n", w.b.String())
	}
}

func TestWriteMetricsFloat64(t *testing.T) {
	w := ExpfmtWriter{
		b: bytes.NewBuffer(nil),
	}
	for _, tc := range []struct {
		family       string
		tags         []string
		constantTags string
		value        float64
		expected     string
	}{
		{"foo", nil, "", 0, "foo 0"},
		{"foo", nil, "", 1, "foo 1"},
		{"foo", nil, `x="y"`, 1, `foo{x="y"} 1`},
		{"foo", []string{"a", "1"}, "", 10, `foo{a="1"} 10`},
		{"foo", []string{"a", "1", "b", "2"}, "", 10, `foo{a="1",b="2"} 10`},
		{"foo", []string{"a", "1", "b", "2"}, `x="y"`, 10, `foo{x="y",a="1",b="2"} 10`},
		{"foo", nil, "", 1.1, `foo 1.1`},
		{"foo", nil, "", math.Inf(1), `foo +Inf`},
		{"foo", nil, "", math.Inf(-1), `foo -Inf`},
		{"foo", nil, "", math.NaN(), `foo NaN`},
		{"foo", nil, "", -1, `foo -1`},
		{"foo", nil, "", -1.1, `foo -1.1`},
		{"foo", nil, "", 1e20, `foo 1e+20`},
	} {
		w.b.Reset()
		w.constantTags = tc.constantTags
		w.WriteMetricName(MetricName{
			Family: MustIdent(tc.family),
			Tags:   MustTags(tc.tags...),
		})
		w.WriteFloat64(tc.value)
		assert.Equal(t, tc.expected+"\n", w.b.String())
	}
}

func TestWriteMetricsInt64(t *testing.T) {
	w := ExpfmtWriter{
		b: bytes.NewBuffer(nil),
	}
	for _, tc := range []struct {
		family       string
		tags         []string
		constantTags string
		value        int64
		expected     string
	}{
		{"foo", nil, "", 0, "foo 0"},
		{"foo", nil, "", 1, "foo 1"},
		{"foo", nil, `x="y"`, -1, `foo{x="y"} -1`},
		{"foo", []string{"a", "1"}, "", 10, `foo{a="1"} 10`},
		{"foo", []string{"a", "1", "b", "2"}, "", 10, `foo{a="1",b="2"} 10`},
		{"foo", []string{"a", "1", "b", "2"}, `x="y"`, 10, `foo{x="y",a="1",b="2"} 10`},
	} {
		w.b.Reset()
		w.constantTags = tc.constantTags
		w.WriteMetricName(MetricName{
			Family: MustIdent(tc.family),
			Tags:   MustTags(tc.tags...),
		})
		w.WriteInt64(tc.value)
		assert.Equal(t, tc.expected+"\n", w.b.String())
	}
}

func TestWriteMetricsDuration(t *testing.T) {
	w := ExpfmtWriter{
		b: bytes.NewBuffer(nil),
	}
	for _, tc := range []struct {
		family       string
		tags         []string
		constantTags string
		value        time.Duration
		expected     string
	}{
		{"foo", nil, "", time.Duration(0), "foo 0"},
		{"foo", nil, "", time.Second, "foo 1"},
		{"foo", nil, `x="y"`, time.Minute, `foo{x="y"} 60`},
		{"foo", []string{"a", "1"}, "", time.Hour, `foo{a="1"} 3600`},
		{"foo", []string{"a", "1", "b", "2"}, "", 10 * time.Second, `foo{a="1",b="2"} 10`},
		{"foo", []string{"a", "1", "b", "2"}, `x="y"`, 10 * time.Second, `foo{x="y",a="1",b="2"} 10`},
	} {
		w.b.Reset()
		w.constantTags = tc.constantTags
		w.WriteMetricName(MetricName{
			Family: MustIdent(tc.family),
			Tags:   MustTags(tc.tags...),
		})
		w.WriteDuration(tc.value)
		assert.Equal(t, tc.expected+"\n", w.b.String())
	}
}

func TestWriteLine(t *testing.T) {
	w := ExpfmtWriter{
		b: bytes.NewBuffer(nil),
	}
	for _, tc := range []struct {
		in, constantTags, expected string
	}{
		{"foo 1", "", "foo 1"},
		{"# HELP foo", `x="y"`, "# HELP foo"},
		{"foo 1", `x="y"`, `foo{x="y"} 1`},
		{`foo{a="b"} 1`, `x="y"`, `foo{x="y",a="b"} 1`},
		{`xxx`, `x="y"`, `xxx`},
		{``, `x="y"`, ``},
	} {
		w.b.Reset()
		w.constantTags = tc.constantTags
		w.WriteLine([]byte(tc.in))
		assert.Equal(t, tc.expected, w.b.String())
	}
}

func TestWriteMetricNameWithVariableTags(t *testing.T) {
	w := ExpfmtWriter{
		b: bytes.NewBuffer(nil),
	}
	for _, tc := range []struct {
		family         string
		tags           []string
		constantTags   string
		labels, values []string
		expected       string
	}{
		{"foo", nil, "", nil, nil, `foo`},
		{"foo", nil, `x="y"`, nil, nil, `foo{x="y"}`},
		{"foo", nil, `x="y"`, []string{"label1"}, []string{"value1"}, `foo{x="y",label1="value1"}`},
		{"foo", []string{"a", "1"}, `x="y"`, []string{"label1"}, []string{"value1"}, `foo{x="y",a="1",label1="value1"}`},
		{"foo", []string{"a", "1"}, `x="y"`, nil, nil, `foo{x="y",a="1"}`},
		{"foo", nil, "", []string{"label1"}, []string{"value1"}, `foo{label1="value1"}`},
		{"foo", nil, "", []string{"label1", "label2"}, []string{"value1", "value2"}, `foo{label1="value1",label2="value2"}`},
	} {
		w.b.Reset()
		w.constantTags = tc.constantTags
		w.WriteMetricNameWithVariableTags(MetricName{
			Family: MustIdent(tc.family),
			Tags:   MustTags(tc.tags...),
		}, makeLabels(tc.labels), makeValues(tc.values))
		assert.Equal(t, w.b.String(), tc.expected)
	}
}

func TestSizeOf(t *testing.T) {
	for _, tc := range []struct {
		family       string
		tags         []string
		constantTags string
		expected     int
	}{
		{"foo", nil, "", len(`foo`)},
		{"foo", []string{"x", "y"}, "", len(`foo{x="y"}`)},
		{"foo", []string{"x", "y", "other_tag", "other_value"}, "", len(`foo{x="y",other_tag="other_value"}`)},
		{"foo", nil, `x="y"`, len(`foo{x="y"}`)},
		{"foo", []string{"x", "y"}, `foo="bar"`, len(`foo{x="y",foo="bar"}`)},
	} {
		assert.Equal(t, tc.expected, sizeOfMetricName(MetricName{
			Family: MustIdent(tc.family),
			Tags:   MustTags(tc.tags...),
		}, tc.constantTags))
	}
}

func TestAppendConstantTags(t *testing.T) {
	w := ExpfmtWriter{
		b: bytes.NewBuffer(nil),
	}
	w2 := w.AppendConstantTags("foo", "bar")

	w.WriteLazyMetricUint64("a", 1)
	w2.WriteLazyMetricUint64("b", 2)

	assert.LinesEqual(t,
		splitLines(strings.Trim(w.b.String(), "\n")),
		[]string{
			"a 1\n",
			"b{foo=\"bar\"} 2\n",
		},
	)
}
