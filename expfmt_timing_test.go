package metrics

import (
	"bytes"
	"testing"
)

func BenchmarkExpfmtWriter(b *testing.B) {
	w := ExpfmtWriter{
		B: bytes.NewBuffer(nil),
	}

	family := MustIdent("really_cool_metric")
	tags := MustTags(
		"foo", "bar",
		"a", "b",
	)
	manyTags := MustTags(
		"foo", "bar",
		"foo", "bar",
		"foo", "bar",
		"foo", "bar",
		"foo", "bar",
		"foo", "bar",
		"foo", "bar",
		"foo", "bar",
		"foo", "bar",
		"foo", "bar",
		"foo", "bar",
		"foo", "bar",
	)

	b.Run("name", func(b *testing.B) {
		w.WriteMetricName(family)

		b.ReportAllocs()
		for b.Loop() {
			w.B.Reset()

			w.WriteMetricName(family)
			b.SetBytes(int64(w.B.Len()))
		}
	})

	b.Run("name with tags", func(b *testing.B) {
		w.WriteMetricName(family, tags...)

		b.ReportAllocs()
		for b.Loop() {
			w.B.Reset()
			w.WriteMetricName(family, tags...)
			b.SetBytes(int64(w.B.Len()))
		}
	})

	b.Run("name with many tags", func(b *testing.B) {
		w.WriteMetricName(family, manyTags...)

		b.ReportAllocs()
		for b.Loop() {
			w.B.Reset()
			w.WriteMetricName(family, manyTags...)
			b.SetBytes(int64(w.B.Len()))
		}
	})

	b.Run("uint64", func(b *testing.B) {
		const value = 100
		w.WriteMetricName(family, tags...)
		w.WriteUint64(value)

		b.ReportAllocs()
		for b.Loop() {
			w.B.Reset()
			w.WriteMetricName(family, tags...)
			w.WriteUint64(value)
			b.SetBytes(int64(w.B.Len()))
		}
	})

	b.Run("float64", func(b *testing.B) {
		const value = 100.5
		w.WriteMetricName(family, tags...)
		w.WriteFloat64(value)

		b.ReportAllocs()
		for b.Loop() {
			w.B.Reset()
			w.WriteMetricName(family, tags...)
			w.WriteFloat64(value)
			b.SetBytes(int64(w.B.Len()))
		}
	})
}
