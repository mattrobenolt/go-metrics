package promhttp

import (
	"bytes"
	"strings"
	"testing"

	"go.withmatt.com/metrics/internal/assert"
)

func TestTransformer(t *testing.T) {
	r := strings.NewReader(`


foo 1
# something
x 1

foo{a="1"} 1
foo{a="2"} 1
######

hist2_count{foo="bar"} 1
hist2_sum{foo="bar"} 1
hist2_bucket{le="+Inf",foo="bar"} 1
hist2_bucket{le="10",foo="bar"} 1
hist1_count{foo="bar"} 1
hist1_sum{foo="bar"} 1
hist1_bucket{le="+Inf",foo="bar"} 1
hist1_bucket{le="10",foo="bar"} 1
hist1_bucket{le="0.1",foo="bar"} 0
foo2{b="1"} 1

`)

	tr := NewTransformer(Mapping{
		"foo": {
			Type: Counter,
			Help: "hello",
		},
		"a": {
			Type: Gauge,
			Help: "cool gauge",
		},
		"hist1": {
			Type: Histogram,
			Help: "cool histogram",
		},
	})
	r.WriteTo(tr)

	var bb bytes.Buffer
	bb.ReadFrom(tr)

	assert.Equal(t, bb.String(), `# HELP foo hello
# TYPE foo counter
foo{a="2"} 1
foo{a="1"} 1
foo 1
# HELP foo2
# TYPE foo2 untyped
foo2{b="1"} 1
# HELP hist1 cool histogram
# TYPE hist1 histogram
hist1_bucket{le="0.1",foo="bar"} 0
hist1_bucket{le="10",foo="bar"} 1
hist1_bucket{le="+Inf",foo="bar"} 1
hist1_sum{foo="bar"} 1
hist1_count{foo="bar"} 1
# HELP hist2
# TYPE hist2 untyped
hist2_bucket{le="10",foo="bar"} 1
hist2_bucket{le="+Inf",foo="bar"} 1
hist2_sum{foo="bar"} 1
hist2_count{foo="bar"} 1
# HELP x
# TYPE x untyped
x 1
`)
}

func TestTransformerNoMapping(t *testing.T) {
	r := strings.NewReader("foo 1")
	var tr Transformer
	r.WriteTo(&tr)

	var bb bytes.Buffer
	bb.ReadFrom(&tr)

	assert.Equal(t, bb.String(), `# HELP foo
# TYPE foo untyped
foo 1
`)
	r.WriteTo(&tr)
}
