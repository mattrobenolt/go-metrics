package metrics

import (
	"bytes"
	"math"
	"testing"

	"go.withmatt.com/metrics/internal/assert"
)

func TestWriteMetricsUint64(t *testing.T) {
	w := ExpfmtWriter{
		B: bytes.NewBuffer(nil),
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
		w.B.Reset()
		w.WriteMetricName(MetricName{
			Family:       MustIdent(tc.family),
			Tags:         MustTags(tc.tags...),
			ConstantTags: tc.constantTags,
		})
		w.WriteUint64(tc.value)
		assert.Equal(t, tc.expected+"\n", w.B.String())
	}
}

func TestWriteMetricsFloat64(t *testing.T) {
	w := ExpfmtWriter{
		B: bytes.NewBuffer(nil),
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
		w.B.Reset()
		w.WriteMetricName(MetricName{
			Family:       MustIdent(tc.family),
			Tags:         MustTags(tc.tags...),
			ConstantTags: tc.constantTags,
		})
		w.WriteFloat64(tc.value)
		assert.Equal(t, tc.expected+"\n", w.B.String())
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
			Family:       MustIdent(tc.family),
			Tags:         MustTags(tc.tags...),
			ConstantTags: tc.constantTags,
		}))
	}
}
