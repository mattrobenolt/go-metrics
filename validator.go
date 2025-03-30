package metrics

import (
	"fmt"
	"unicode/utf8"
)

// MustIdent ensures that s is a valid identifier and returns an Ident.
//
// This will panic if s is an invalid identifier.
func MustIdent(s string) Ident {
	return makeIdentHandle(s)
}

// MustLabel ensures that s is a valid label and returns a Label.
//
// This will panic if s is an invalid identifier.
func MustLabel(s string) Label {
	return makeIdentHandle(s)
}

// MustTag ensures that the label value pair are a valid
// tag, and returns a Tag. Label must be a valid Ident and
// value must be a valid tag value.
//
// This will panic if label or value are invalid for a tag.
func MustTag(label, value string) Tag {
	return Tag{
		label: MustLabel(label),
		value: MustValue(value),
	}
}

// MustValue ensures that s is a valid tag value.
//
// This will panic if s is an invalid tag value.
func MustValue(s string) Value {
	if !validateLabelValue(s) {
		panic(fmt.Sprintf("metrics: invalid tag value: %q", s))
	}
	// Values are expected to vary quite a lot, and there's no use
	// in uniquing.
	return Value{s}
}

// MustTags converts label value pairs into Tags.
//
// This will panic if s is an invalid tag value.
func MustTags(tags ...string) []Tag {
	if len(tags) == 0 {
		return nil
	}

	if len(tags)%2 != 0 {
		panic(fmt.Sprintf("metrics: tag label/values must be in pairs, got: %v", tags))
	}

	pairs := make([]Tag, 0, len(tags)/2)
	for i := 0; i < len(tags); i += 2 {
		pairs = append(pairs, MustTag(tags[i], tags[i+1]))
	}
	return pairs
}

// labelValue can be any sequence of UTF-8 characters, but the backslash (\),
// double-quote ("), and line feed (\n) characters have to be escaped as
// \\, \", and \n, respectively.
func validateLabelValue(s string) bool {
	// XXX: This is marginally faster to do in two passes since
	// utf8.ValidString is so optimized.
	if !utf8.ValidString(s) {
		return false
	}

	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '"', '\n':
			// disallow double-quote and line feed
			return false
		case '\\':
			// started escaping
			i++
			if i == len(s) {
				// oh no, we got to the end
				return false
			}
			switch s[i] {
			case 'n', '\\', '"':
				// escaping line feed, another backslash, or a quote
			default:
				// anything else is invalid escape sequence
				return false
			}
		}
	}
	return true
}

// validateIdent validates effectively this pattern:
// ^[a-zA-Z_:.][a-zA-Z0-9_:.]*$
func validateIdent(s string) bool {
	if len(s) == 0 {
		return false
	}
	// first character is either alpha or symbol
	if !isAlpha(s[0]) && !isSymbol(s[0]) {
		return false
	}
	// every other character can include numbers
	for i := 1; i < len(s); i++ {
		if !isAlpha(s[i]) && !isNumeric(s[i]) && !isSymbol(s[i]) {
			return false
		}
	}
	return true
}

func isAlpha(c byte) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

func isNumeric(c byte) bool {
	return '0' <= c && c <= '9'
}

func isSymbol(c byte) bool {
	switch c {
	case '_', ':', '.':
		return true
	}
	return false
}
