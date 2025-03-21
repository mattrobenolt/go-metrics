package metrics

import (
	"bytes"
	"math"
	"strconv"
)

// Ident represents either a metric family or a tag label.
type Ident struct {
	v *string
}

func (i Ident) String() string {
	return *i.v
}

// Label is also an Ident.
type Label = Ident

// WithValueUnsafe joins a Label with value that _must_ already be known
// to be valid. No validation is done on the value and this will perform faster.
func (l Label) WithValueUnsafe(val string) Tag {
	return Tag{l, Value{val}}
}

// Tag represents a label/value pair for a metric.
type Tag struct {
	label Label
	value Value
}

func (t Tag) String() string {
	return t.label.String() + `="` + t.value.v + `"`
}

// Value represents a Tag value that has been validated as a correct string.
type Value struct {
	v string
}

func (v Value) String() string {
	return v.v
}

// ExpfmtWriter wraps a bytes.Buffer adds functionality to write
// the Prometheus text exposiiton format.
type ExpfmtWriter struct {
	B *bytes.Buffer
}

// WriteMetricName writes the family and optional tags.
func (w ExpfmtWriter) WriteMetricName(family Ident, tags ...Tag) {
	writeMetricName(w.B, family, tags)
}

// WriteUint64 writes a uint64 and signals the end of the metric.
func (w ExpfmtWriter) WriteUint64(value uint64) {
	w.B.WriteByte(' ')
	writeUint64(w.B, value)
	w.B.WriteByte('\n')
}

// WriteUint64 writes a float64 and signals the end of the metric.
func (w ExpfmtWriter) WriteFloat64(value float64) {
	w.B.WriteByte(' ')
	writeFloat64(w.B, value)
	w.B.WriteByte('\n')
}

func writeUint64(b *bytes.Buffer, value uint64) {
	b.Write(strconv.AppendUint(b.AvailableBuffer(), value, 10))
}

func writeFloat64(b *bytes.Buffer, value float64) {
	intvalue := int64(value)
	switch {
	case float64(intvalue) == value:
		b.Write(strconv.AppendInt(b.AvailableBuffer(), intvalue, 10))
	case value < -math.MaxFloat64:
		b.WriteString("-Inf")
	case value > math.MaxFloat64:
		b.WriteString("+Inf")
	case math.IsNaN(value):
		b.WriteString("NaN")
	default:
		b.Write(strconv.AppendFloat(b.AvailableBuffer(), value, 'g', -1, 64))
	}
}

func getMetricName(family Ident, tags []Tag) string {
	if len(tags) == 0 {
		return family.String()
	}
	var b bytes.Buffer
	writeMetricName(&b, family, tags)
	return b.String()
}

func sizeOfMetric(family string, tags []Tag) int {
	size := len(family)
	size += len("{}")
	return size + sizeOfTags(tags)
}

func sizeOfTags(tags []Tag) int {
	var size int
	for _, tag := range tags {
		size += len(tag.label.String())
		size += len(tag.value.String())
		size += len(`="",`)
	}
	// subtract 1 since the last tag does not have a trailing comma
	return size - 1
}

func writeTag(b *bytes.Buffer, tag Tag) {
	b.WriteString(tag.label.String())
	b.WriteString(`="`)
	b.WriteString(tag.value.String())
	b.WriteByte('"')
}

func writeMetricName(b *bytes.Buffer, family Ident, tags []Tag) {
	if len(tags) == 0 {
		b.WriteString(family.String())
		return
	}

	b.Grow(sizeOfMetric(family.String(), tags))

	b.WriteString(family.String())
	b.WriteByte('{')
	writeTags(b, tags)
	b.WriteByte('}')
}

func writeTags(b *bytes.Buffer, tags []Tag) {
	for i, tag := range tags {
		if i > 0 {
			b.WriteByte(',')
		}
		writeTag(b, tag)
	}
}
