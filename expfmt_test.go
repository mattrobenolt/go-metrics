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
		family   string
		tags     []string
		value    uint64
		expected string
	}{
		{"foo", nil, 0, "foo 0"},
		{"foo", nil, 1, "foo 1"},
		{"foo", []string{"a", "1"}, 10, `foo{a="1"} 10`},
		{"foo", []string{"a", "1", "b", "2"}, 10, `foo{a="1",b="2"} 10`},
	} {
		w.B.Reset()
		w.WriteMetricName(MustIdent(tc.family), MustTags(tc.tags...)...)
		w.WriteUint64(tc.value)
		assert.Equal(t, tc.expected+"\n", w.B.String())
	}
}

func TestWriteMetricsFloat64(t *testing.T) {
	w := ExpfmtWriter{
		B: bytes.NewBuffer(nil),
	}
	for _, tc := range []struct {
		family   string
		tags     []string
		value    float64
		expected string
	}{
		{"foo", nil, 0, "foo 0"},
		{"foo", nil, 1, "foo 1"},
		{"foo", []string{"a", "1"}, 10, `foo{a="1"} 10`},
		{"foo", []string{"a", "1", "b", "2"}, 10, `foo{a="1",b="2"} 10`},
		{"foo", nil, 1.1, `foo 1.1`},
		{"foo", nil, math.Inf(1), `foo +Inf`},
		{"foo", nil, math.Inf(-1), `foo -Inf`},
		{"foo", nil, math.NaN(), `foo NaN`},
		{"foo", nil, -1, `foo -1`},
		{"foo", nil, -1.1, `foo -1.1`},
		{"foo", nil, 1e20, `foo 1e+20`},
	} {
		w.B.Reset()
		w.WriteMetricName(MustIdent(tc.family), MustTags(tc.tags...)...)
		w.WriteFloat64(tc.value)
		assert.Equal(t, tc.expected+"\n", w.B.String())
	}
}
