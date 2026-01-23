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

func TestNewTag(t *testing.T) {
	tag := NewTag(MustLabel("foo"), SanitizeValue("bar"))
	assert.Equal(t, tag.String(), `foo="bar"`)
}

func TestSanitizeValue(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// Already valid - no escaping needed
		{"", ""},
		{"hello", "hello"},
		{"foo_bar", "foo_bar"},
		{"ü", "ü"},
		{"🍖", "🍖"},

		// Needs escaping
		{`\`, `\\`},
		{`"`, `\"`},
		{"\n", `\n`},
		{`foo\bar`, `foo\\bar`},
		{`foo"bar`, `foo\"bar`},
		{"foo\nbar", `foo\nbar`},
		{`a\b"c` + "\n" + `d`, `a\\b\"c\nd`},

		// Invalid UTF-8 replaced with replacement character
		{"\xff", "\uFFFD"},
		{"hello\xffworld", "hello\uFFFDworld"},
		// Invalid UTF-8 combined with escaping
		{"\xff\n\xfe", "\uFFFD\\n\uFFFD"},
		{"\xff\\\"\xfe", "\uFFFD\\\\\\\"\uFFFD"},
		// Invalid UTF-8 mixed with valid multi-byte UTF-8
		{"\xffü\xfe", "\uFFFDü\uFFFD"},
	}

	for _, tt := range tests {
		got := SanitizeValue(tt.input)
		assert.Equal(t, got.String(), tt.want)
		// Verify the result passes validation
		assert.Equal(t, MustValue(got.String()).String(), tt.want)
	}
}

func TestSanitizeValueNoAlloc(t *testing.T) {
	// Valid strings should not allocate
	valid := "hello_world_123"
	allocs := testing.AllocsPerRun(100, func() {
		_ = SanitizeValue(valid)
	})
	assert.Equal(t, allocs, float64(0))
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

	b.Run("SanitizeValue/clean", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			SanitizeValue(`some.other.value`)
		}
	})

	b.Run("SanitizeValue/escape", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			SanitizeValue("foo\nbar\"baz\\qux")
		}
	})
}
