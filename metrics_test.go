package metrics

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"go.withmatt.com/metrics/internal/assert"
)

// assertMarshal compares lines of output ignoring order
func assertMarshal(tb testing.TB, set *Set, expected []string) {
	var b bytes.Buffer
	set.WritePrometheusUnthrottled(&b)
	lines := strings.Split(strings.Trim(b.String(), "\n"), "\n")
	assert.Equal(tb,
		strings.Join(lines, "\n"),
		strings.Join(expected, "\n"),
	)
}

func hammer(t testing.TB, n int, f func()) {
	var wg sync.WaitGroup
	for range n {
		wg.Add(1)
		go func() {
			defer func() {
				assert.Nil(t, recover())
				wg.Done()
			}()
			f()
		}()
	}
	wg.Wait()
}
