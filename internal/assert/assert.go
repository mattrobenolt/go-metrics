// Copyright 2021-2025 The Connect Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package assert is a minimal assert package using generics.
//
// This prevents connect from needing additional dependencies.
package assert

import (
	"cmp"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"go.withmatt.com/metrics/internal/assert/difflib"
)

// Equal asserts that two values are equal.
func Equal[T comparable](tb testing.TB, got, want T, options ...Option) bool {
	tb.Helper()
	if got == want {
		return true
	}
	report(tb, got, want, "assert.Equal", true /* showWant */, options...)
	return false
}

func LinesEqual(tb testing.TB, got, want []string, options ...Option) bool {
	tb.Helper()
	if slices.Equal(got, want) {
		return true
	}
	reportDiff(tb, got, want)
	return false
}

func SlicesEqual[S ~[]E, E comparable](tb testing.TB, got, want S, options ...Option) bool {
	tb.Helper()
	if slices.Equal(got, want) {
		return true
	}
	report(tb, got, want, "assert.SlicesEqual", true /* showWant */, options...)
	return false
}

// NotEqual asserts that two values aren't equal.
func NotEqual[T comparable](tb testing.TB, got, want T, options ...Option) bool {
	tb.Helper()
	if got != want {
		return true
	}
	report(tb, got, want, "assert.NotEqual", true /* showWant */, options...)
	return false
}

func Greater[T cmp.Ordered](tb testing.TB, a, b T, options ...Option) bool {
	tb.Helper()
	if a > b {
		return true
	}
	report(tb, fmt.Sprintf("%v <= %v", a, b), nil, "assert.Greater", false /* showWant */, options...)
	return false
}

func Less[T cmp.Ordered](tb testing.TB, a, b T, options ...Option) bool {
	tb.Helper()
	if a < b {
		return true
	}
	report(tb, fmt.Sprintf("%v >= %v", a, b), nil, "assert.Less", false /* showWant */, options...)
	return false
}

// Nil asserts that the value is nil.
func Nil(tb testing.TB, got any, options ...Option) bool {
	tb.Helper()
	if isNil(got) {
		return true
	}
	report(tb, got, nil, "assert.Nil", false /* showWant */, options...)
	return false
}

// NotNil asserts that the value isn't nil.
func NotNil(tb testing.TB, got any, options ...Option) bool {
	tb.Helper()
	if !isNil(got) {
		return true
	}
	report(tb, got, nil, "assert.NotNil", false /* showWant */, options...)
	return false
}

// ErrorIs asserts that "want" is in "got's" error chain. See the standard
// library's errors package for details on error chains. On failure, output is
// identical to Equal.
func ErrorIs(tb testing.TB, got, want error, options ...Option) bool {
	tb.Helper()
	if errors.Is(got, want) {
		return true
	}
	report(tb, got, want, "assert.ErrorIs", true /* showWant */, options...)
	return false
}

// False asserts that "got" is false.
func False(tb testing.TB, got bool, options ...Option) bool {
	tb.Helper()
	if !got {
		return true
	}
	report(tb, got, false, "assert.False", false /* showWant */, options...)
	return false
}

// True asserts that "got" is true.
func True(tb testing.TB, got bool, options ...Option) bool {
	tb.Helper()
	if got {
		return true
	}
	report(tb, got, true, "assert.True", false /* showWant */, options...)
	return false
}

// Panics asserts that the function called panics.
func Panics(tb testing.TB, panicker func(), options ...Option) {
	tb.Helper()
	defer func() {
		if r := recover(); r == nil {
			report(tb, r, nil, "assert.Panic", false /* showWant */, options...)
		}
	}()
	panicker()
}

// An Option configures an assertion.
type Option interface {
	// Only option we've needed so far is a formatted message, so we can keep
	// this simple.
	message() string
}

// Sprintf adds a user-defined message to the assertion's output. The arguments
// are passed directly to fmt.Sprintf for formatting.
//
// If Sprintf is passed multiple times, only the last message is used.
func Sprintf(template string, args ...any) Option {
	return &sprintfOption{fmt.Sprintf(template, args...)}
}

type sprintfOption struct {
	msg string
}

func (o *sprintfOption) message() string {
	return o.msg
}

func report(tb testing.TB, got, want any, desc string, showWant bool, options ...Option) {
	tb.Helper()
	var buffer strings.Builder
	if len(options) > 0 {
		buffer.WriteString(options[len(options)-1].message())
	}
	buffer.WriteString("\n")
	fmt.Fprintf(&buffer, "assertion:\t%s\n", desc)
	switch {
	case isStringSlice(got):
		reportStringSlice(&buffer, got.([]string), want.([]string))
	case isSlice(got):
		fmt.Fprintf(&buffer, "got (len=%d):\n %#v\n", reflect.ValueOf(got).Len(), got)
		if showWant {
			fmt.Fprintf(&buffer, "\nwant (len=%d):\n %#v\n", reflect.ValueOf(want).Len(), want)
		}
	default:
		fmt.Fprintf(&buffer, "got:\t%+v\n", got)
		if showWant {
			fmt.Fprintf(&buffer, "want:\t%+v\n", want)
		}
	}
	tb.Error(buffer.String())
}

func reportDiff(tb testing.TB, got, want []string) {
	tb.Helper()
	var buffer strings.Builder
	buffer.WriteString("\n")
	fmt.Fprintf(&buffer, "assertion:\tassert.LinesEqual\n")
	difflib.WriteUnifiedDiff(&buffer, difflib.UnifiedDiff{
		A:        got,
		B:        want,
		FromFile: "got",
		ToFile:   "want",
		Context:  2,
	})
	tb.Error(buffer.String())
}

func reportStringSlice(buffer *strings.Builder, got, want []string) {
	difflib.WriteUnifiedDiff(buffer, difflib.UnifiedDiff{
		A:        got,
		B:        want,
		FromFile: "got",
		ToFile:   "want",
		Context:  2,
	})
}

func isSlice(got any) bool {
	val := reflect.ValueOf(got)
	return val.Kind() == reflect.Slice
}

func isStringSlice(got any) bool {
	val := reflect.ValueOf(got)
	return val.Kind() == reflect.Slice && val.Type().Elem().Kind() == reflect.String
}

func isNil(got any) bool {
	// Simple case, true only when the user directly passes a literal nil.
	if got == nil {
		return true
	}
	// Possibly more complex. Interfaces are a pair of words: a pointer to a type
	// and a pointer to a value. Because we're passing got as an interface, it's
	// likely that we've gotten a non-nil type and a nil value. This makes got
	// itself non-nil, but the user's code passed a nil value.
	val := reflect.ValueOf(got)
	//nolint:exhaustive
	switch val.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return val.IsNil()
	default:
		return false
	}
}
