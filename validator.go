package metrics

import "fmt"

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

// NewTag creates a Tag from an already-validated Label and Value.
func NewTag(label Label, value Value) Tag {
	return Tag{label: label, value: value}
}

// UnsafeValue creates a Value, but does not validate it.
func UnsafeValue(s string) Value {
	return Value{s}
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

// SanitizeValue escapes a string to be a valid tag value.
// Backslashes, double-quotes, and line feeds are escaped.
// Invalid UTF-8 sequences are replaced with the Unicode replacement character.
//
// If the input is already valid, it is returned without allocation.
func SanitizeValue(s string) Value {
	// Fast path for valid UTF-8: just check for escape characters
	if utf8ValidString(s) {
		n := 0
		for i := range len(s) {
			switch s[i] {
			case '\\', '"', '\n':
				n++
			}
		}
		if n == 0 {
			return Value{s}
		}
		// Valid UTF-8 but needs escaping
		b := make([]byte, 0, len(s)+n)
		last := 0
		for i := range len(s) {
			var esc byte
			switch s[i] {
			case '\\':
				esc = '\\'
			case '"':
				esc = '"'
			case '\n':
				esc = 'n'
			default:
				continue
			}
			b = append(b, s[last:i]...)
			b = append(b, '\\', esc)
			last = i + 1
		}
		b = append(b, s[last:]...)
		return Value{string(b)}
	}

	// Slow path: invalid UTF-8, need to scan and replace
	extra := 0
	for i := 0; i < len(s); {
		c := s[i]
		if c < 0x80 {
			switch c {
			case '\\', '"', '\n':
				extra++
			}
			i++
			continue
		}
		r, size := decodeRuneInString(s[i:])
		if r == runeError && size == 1 {
			extra += 2 // replacement char is 3 bytes, invalid is 1
			i++
		} else {
			i += size
		}
	}

	b := make([]byte, 0, len(s)+extra)
	last := 0
	for i := 0; i < len(s); {
		c := s[i]
		if c < 0x80 {
			var esc byte
			switch c {
			case '\\':
				esc = '\\'
			case '"':
				esc = '"'
			case '\n':
				esc = 'n'
			default:
				i++
				continue
			}
			b = append(b, s[last:i]...)
			b = append(b, '\\', esc)
			i++
			last = i
			continue
		}

		r, size := decodeRuneInString(s[i:])
		if r == runeError && size == 1 {
			b = append(b, s[last:i]...)
			b = append(b, "\uFFFD"...)
			i++
			last = i
		} else {
			i += size
		}
	}
	b = append(b, s[last:]...)

	return Value{string(b)}
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
	if !utf8ValidString(s) {
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
