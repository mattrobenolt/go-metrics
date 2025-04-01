package metrics

import "unicode/utf8"

var utf8ValidString = utf8.ValidString

// replaceUTF8ValidStringHook is meant to be hijacked by a go:linkname
// directive to replace the utf8 validation implementation.
//
//nolint:unused
func replaceUTF8ValidStringHook(fn func(string) bool) {
	utf8ValidString = fn
}
