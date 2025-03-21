package metrics

import (
	"testing"

	"go.withmatt.com/metrics/internal/assert"
)

func TestValidateIdentSuccess(t *testing.T) {
	for _, s := range []string{
		"a",
		"_9:8",
		`:foo:bar`,
		`some.foo`,
		`aB`,
	} {
		assert.Equal(t, MustIdent(s).String(), s)
	}
}

func TestValidateIdentError(t *testing.T) {
	for _, s := range []string{
		"",
		"1abc",
		"a{}",
		"a b",
		"a=b",
		"ü",
		"🍖",
	} {
		assert.Panics(t, func() { MustIdent(s) })
	}
}

func TestValidateValueSuccess(t *testing.T) {
	for _, s := range []string{
		"",
		"1abc",
		"a{}",
		"a b",
		"a=b",
		"ü",
		"🍖",
		`\n`,
		`\\`,
		`foo\nbar`,
		`foo\"bar`,
		`foo\\bar`,
	} {
		assert.Equal(t, MustValue(s).String(), s)
	}
}

func TestValidateValueError(t *testing.T) {
	for _, s := range []string{
		`"`,
		"\n",
		`\`,
		"foo\nbar",
		`foo"bar`,
		`foo\`,
		`foo\bar`,
	} {
		assert.Panics(t, func() { MustValue(s) })
	}
}

func BenchmarkValidate(b *testing.B) {
	b.Run("MustIdent", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			MustIdent(`go_memstats_mspan_inuse_bytes`)
		}
	})

	b.Run("validateIdent", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			validateIdent(`go_memstats_mspan_inuse_bytes`)
		}
	})

	b.Run("MustValue", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			MustValue(`some.other.value`)
		}
	})
}
