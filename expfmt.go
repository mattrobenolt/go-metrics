package metrics

import (
	"bytes"
	"math"
	"strconv"
	"strings"
	"time"
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
	b            *bytes.Buffer
	constantTags string
}

func (w ExpfmtWriter) Buffer() *bytes.Buffer {
	return w.b
}

func (w ExpfmtWriter) ConstantTags() string {
	return w.constantTags
}

// WriteMetricName writes the family and optional tags.
func (w ExpfmtWriter) WriteMetricName(name MetricName) {
	writeMetricName(w.b, name, w.constantTags)
}

// WriteUint64 writes a uint64 and signals the end of the metric.
func (w ExpfmtWriter) WriteUint64(value uint64) {
	w.b.WriteByte(' ')
	writeUint64(w.b, value)
	w.b.WriteByte('\n')
}

// WriteUint64 writes a float64 and signals the end of the metric.
func (w ExpfmtWriter) WriteFloat64(value float64) {
	w.b.WriteByte(' ')
	writeFloat64(w.b, value)
	w.b.WriteByte('\n')
}

func (w ExpfmtWriter) WriteMetricUint64(name MetricName, value uint64) {
	w.WriteMetricName(name)
	w.WriteUint64(value)
}

func (w ExpfmtWriter) WriteMetricFloat64(name MetricName, value float64) {
	w.WriteMetricName(name)
	w.WriteFloat64(value)
}

func (w ExpfmtWriter) WriteMetricDuration(name MetricName, value time.Duration) {
	w.WriteMetricName(name)
	w.WriteFloat64(value.Seconds())
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

func sizeOfMetricName(name MetricName, constantTags string) int {
	if !name.HasTags() && len(constantTags) == 0 {
		return len(name.Family.String())
	}
	size := len(name.Family.String())
	size += len("{}")
	return size + sizeOfTags(name.Tags, constantTags)
}

func sizeOfTags(tags []Tag, constantTags string) int {
	var size int
	if len(constantTags) > 0 {
		size += len(constantTags) + 1
	}

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

func writeMetricName(b *bytes.Buffer, name MetricName, constantTags string) {
	if !name.HasTags() && len(constantTags) == 0 {
		b.WriteString(name.Family.String())
		return
	}

	b.Grow(sizeOfMetricName(name, constantTags))

	b.WriteString(name.Family.String())
	b.WriteByte('{')
	writeTags(b, constantTags, name.Tags)
	b.WriteByte('}')
}

func writeTags(b *bytes.Buffer, constantTags string, tags []Tag) {
	if len(constantTags) > 0 {
		b.WriteString(constantTags)
		if len(tags) == 0 {
			return
		}
		b.WriteByte(',')
	}

	for i, tag := range tags {
		if i > 0 {
			b.WriteByte(',')
		}
		writeTag(b, tag)
	}
}

func materializeTags(tags []Tag) string {
	var sb strings.Builder
	for i, tag := range tags {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(tag.label.String())
		sb.WriteString(`="`)
		sb.WriteString(tag.value.String())
		sb.WriteByte('"')
	}
	return sb.String()
}
