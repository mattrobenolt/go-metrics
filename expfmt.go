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

// ExpfmtWriter wraps a [bytes.Buffer] adds functionality to write
// the Prometheus text exposition format.
type ExpfmtWriter struct {
	b            *bytes.Buffer
	constantTags string
}

// Buffer returns the underlying [bytes.Buffer].
func (w ExpfmtWriter) Buffer() *bytes.Buffer {
	return w.b
}

// ConstantTags returns the materialized set of tags that are written
// for every metric.
func (w ExpfmtWriter) ConstantTags() string {
	return w.constantTags
}

// WriteMetricName writes the family name, optional tags, and constant tags.
func (w ExpfmtWriter) WriteMetricName(name MetricName) {
	writeMetricName(w.b, name, w.constantTags)
}

// WriteUint64 writes a uint64 and signals the end of the metric.
func (w ExpfmtWriter) WriteUint64(value uint64) {
	w.b.WriteByte(' ')
	writeUint64(w.b, value)
	w.b.WriteByte('\n')
}

// WriteInt64 writes a int64 and signals the end of the metric.
func (w ExpfmtWriter) WriteInt64(value int64) {
	w.b.WriteByte(' ')
	writeInt64(w.b, value)
	w.b.WriteByte('\n')
}

// WriteFloat64 writes a float64 and signals the end of the metric.
func (w ExpfmtWriter) WriteFloat64(value float64) {
	w.b.WriteByte(' ')
	writeFloat64(w.b, value)
	w.b.WriteByte('\n')
}

// WriteMetricUint64 writes a full MetricName and uint64 value.
func (w ExpfmtWriter) WriteMetricUint64(name MetricName, value uint64) {
	w.WriteMetricName(name)
	w.WriteUint64(value)
}

// WriteMetricInt64 writes a full MetricName and int64 value.
func (w ExpfmtWriter) WriteMetricInt64(name MetricName, value int64) {
	w.WriteMetricName(name)
	w.WriteInt64(value)
}

// WriteMetricFloat64 writes a full MetricName and float64 value.
func (w ExpfmtWriter) WriteMetricFloat64(name MetricName, value float64) {
	w.WriteMetricName(name)
	w.WriteFloat64(value)
}

// WriteMetricDuration writes a full MetricName and time.Duration value as seconds.
func (w ExpfmtWriter) WriteMetricDuration(name MetricName, value time.Duration) {
	w.WriteMetricName(name)
	w.WriteFloat64(value.Seconds())
}

// WriteLazyMetricUint64 writes a full metric name and uint64 value.
// Tags are passed as interleaving [label value] pairs.
// Prefer [ExpfmtWriter.WriteMetricUint64] when performance is critical.
// This will panic if family or tag labels are invalid.
func (w ExpfmtWriter) WriteLazyMetricUint64(family string, value uint64, tags ...string) {
	w.WriteMetricUint64(MetricName{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
	}, value)
}

// WriteLazyMetricFloat64 writes a full metric name and float64 value.
// Tags are passed as interleaving [label value] pairs.
// Prefer [ExpfmtWriter.WriteMetricFloat64] when performance is critical.
// This will panic if family or tag labels are invalid.
func (w ExpfmtWriter) WriteLazyMetricFloat64(family string, value float64, tags ...string) {
	w.WriteMetricFloat64(MetricName{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
	}, value)
}

// WriteLazyMetricDuration writes a full metric name and time.Duration value as seconds.
// Tags are passed as interleaving [label value] pairs.
// Prefer [ExpfmtWriter.WriteMetricDuration] when performance is critical.
// This will panic if family or tag labels are invalid.
func (w ExpfmtWriter) WriteLazyMetricDuration(family string, value time.Duration, tags ...string) {
	w.WriteLazyMetricFloat64(family, value.Seconds(), tags...)
}

func writeUint64(b *bytes.Buffer, value uint64) {
	b.Write(strconv.AppendUint(b.AvailableBuffer(), value, 10))
}

func writeInt64(b *bytes.Buffer, value int64) {
	b.Write(strconv.AppendInt(b.AvailableBuffer(), value, 10))
}

func writeFloat64(b *bytes.Buffer, value float64) {
	intvalue := int64(value)
	switch {
	case float64(intvalue) == value:
		writeInt64(b, intvalue)
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
	if !name.hasTags() && len(constantTags) == 0 {
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
	if !name.hasTags() && len(constantTags) == 0 {
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
