package metrics

import (
	"bytes"
	"math/rand/v2"
	"strconv"
	"testing"
	"time"

	"go.withmatt.com/metrics/internal/fasttime"
)

// BenchmarkSetVecWithTTL measures existing-label lookups, without a counter
// increment that would introduce a second source of contention. Both TTL modes
// go through SetVec so the comparison isolates the renewal cost.
func BenchmarkSetVecWithTTL(b *testing.B) {
	for _, tc := range []struct {
		name string
		ttl  time.Duration
	}{
		{"no_ttl", 0},
		{"ttl", time.Hour},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.Run("serial", func(b *testing.B) {
				set := NewSet()
				v := set.NewSetVecWithTTL("group", tc.ttl).NewUint64Vec("count")
				v.WithLabelValues("a")

				b.ReportAllocs()
				for b.Loop() {
					v.WithLabelValues("a")
				}
			})

			for _, keyCount := range []int{1, 64, 1024} {
				b.Run("parallel/keys="+strconv.Itoa(keyCount), func(b *testing.B) {
					set := NewSet()
					v := set.NewSetVecWithTTL("group", tc.ttl).NewUint64Vec("count")
					// Every worker samples the same populated key space. Larger
					// spaces model many keys with occasional overlapping lookups.
					values := make([]string, keyCount)
					for i := range values {
						values[i] = strconv.Itoa(i)
						v.WithLabelValues(values[i])
					}

					b.ReportAllocs()
					b.ResetTimer()
					b.RunParallel(func(pb *testing.PB) {
						for pb.Next() {
							// Include the same key-selection cost in both TTL modes.
							//nolint:gosec // Benchmark inputs do not require secure randomness.
							v.WithLabelValues(values[rand.IntN(len(values))])
						}
					})
				})
			}
		})
	}
}

// BenchmarkWritePrometheusTTL measures a complete unthrottled scrape of 256
// child sets. Expiration setup is excluded, so expired measures removal rather
// than construction. Keep this setup aligned with the expiration implementation
// when changing how a set records idle time.
func BenchmarkWritePrometheusTTL(b *testing.B) {
	const childCount = 256

	modes := []struct {
		name           string
		setTTL         time.Duration
		isActiveFunc   IsActiveFunc
		setupChild     func(*Set)
		wantExpiration bool
	}{
		{
			// Base case: TTL is not enabled.
			"no_ttl",
			0,
			nil,
			func(*Set) {},
			false,
		},
		{
			// TTL is enabled, metric is not expired.
			"fresh",
			time.Hour,
			nil,
			func(*Set) {},
			false,
		},
		{
			// TTL is enabled, metric is expired but still active.
			"active",
			time.Hour,
			func(s *Set) bool {
				count, _ := s.GetMetricUint64("count")
				return count > 0
			},
			func(child *Set) {
				// Active sets must survive even with an old timestamp.
				child.lastUsed.Store(fastClock().Now() - fasttime.Instant(2*time.Hour))
			},
			false,
		},
		{
			// TTL is enabled, metric is expired and not still active.
			"expired",
			time.Hour,
			nil,
			func(child *Set) {
				child.lastUsed.Store(fastClock().Now() - fasttime.Instant(2*time.Hour))
			},
			true,
		},
	}

	for _, mode := range modes {
		b.Run(mode.name, func(b *testing.B) {
			newFixture := func() *Set {
				set := NewSet()
				sv := set.NewSetVecWithTTL("group", mode.setTTL)
				if mode.isActiveFunc != nil {
					sv.SetIsActive(mode.isActiveFunc)
				}
				for i := range childCount {
					child := sv.WithLabelValue(strconv.Itoa(i))
					child.NewUint64("count").Inc()
					mode.setupChild(child)
				}
				return set
			}

			set := newFixture()
			var buf bytes.Buffer
			if !mode.wantExpiration {
				// Warm the output buffer and metric ordering caches.
				if _, err := set.WritePrometheusUnthrottled(&buf); err != nil {
					b.Fatal(err)
				}
			}
			b.ReportAllocs()
			for b.Loop() {
				if mode.wantExpiration {
					// Each time through the loop, our Sets
					// will expire. To ensure we are doing
					// work each time, we need to create
					// the set again.
					b.StopTimer()
					set = newFixture()
					b.StartTimer()
				}
				buf.Reset()
				if _, err := set.WritePrometheusUnthrottled(&buf); err != nil {
					b.Fatal(err)
				}
			}
			b.ReportMetric(childCount, "sets/op")

			// Verify that the benchmark exercised collection or removal, rather
			// than accidentally timing an empty parent throughout the run.
			remaining := 0
			set.setsByHash.Range(func(_ metricHash, _ *Set) bool {
				remaining++
				return true
			})
			if mode.wantExpiration {
				if remaining != 0 {
					b.Fatalf("remaining sets: got %d, want 0", remaining)
				}
				if buf.Len() != 0 {
					b.Fatal("expected WritePrometheus not to export metrics, but got:\n" + buf.String())
				}
			} else {
				if remaining != childCount {
					b.Fatalf("remaining sets: got %d, want %d", remaining, childCount)
				}
				if buf.Len() == 0 {
					b.Fatal("expected WritePrometheus to export metrics, but none were exported")
				}
			}
		})
	}
}
