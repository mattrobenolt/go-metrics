package metrics

import (
	"bytes"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"

	"go.withmatt.com/metrics/internal/assert"
)

// assertMarshal compares lines of output ignoring order
func assertMarshal(tb testing.TB, set *Set, expected []string) {
	tb.Helper()
	var b bytes.Buffer
	set.WritePrometheusUnthrottled(&b)
	lines := strings.Split(strings.Trim(b.String(), "\n"), "\n")
	assert.Equal(tb,
		strings.Join(lines, "\n"),
		strings.Join(expected, "\n"),
	)
}

func assertMarshalUnordered(tb testing.TB, set *Set, expected []string) {
	tb.Helper()
	var b bytes.Buffer
	set.WritePrometheusUnthrottled(&b)
	lines := strings.Split(strings.Trim(b.String(), "\n"), "\n")
	slices.Sort(lines)
	slices.Sort(expected)
	assert.Equal(tb,
		strings.Join(lines, "\n"),
		strings.Join(expected, "\n"),
	)
}

func hammer(tb testing.TB, n int, f func(int)) {
	tb.Helper()
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(4))

	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer func() {
				assert.Nil(tb, recover())
				wg.Done()
			}()
			f(i)
		}(i)
	}
	wg.Wait()
}
