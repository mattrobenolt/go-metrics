package syncx

import (
	"cmp"
	"runtime"
	"sync"
	"testing"
)

func BenchmarkSortedMap(b *testing.B) {
	const p = 4
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(p))

	b.Run("parallel-insert-new", func(b *testing.B) {
		var sm SortedMap[int, int]
		sm.Init(cmp.Compare[int])
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			var i int
			for ; pb.Next(); i++ {
				sm.Store(i, i)
			}
		})
	})

	b.Run("series-insert-new", func(b *testing.B) {
		var sm SortedMap[int, int]
		sm.Init(cmp.Compare[int])
		b.ReportAllocs()
		for i := 0; b.Loop(); i++ {
			sm.Store(i, i)
		}
	})

	b.Run("parallel-insert-same", func(b *testing.B) {
		var sm SortedMap[int, int]
		sm.Init(cmp.Compare[int])
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				sm.Store(1, 1)
			}
		})
	})

	b.Run("series-insert-same", func(b *testing.B) {
		var sm SortedMap[int, int]
		sm.Init(cmp.Compare[int])
		b.ReportAllocs()
		for b.Loop() {
			sm.Store(1, 1)
		}
	})

	b.Run("parallel-load-hit", func(b *testing.B) {
		var sm SortedMap[int, int]
		sm.Init(cmp.Compare[int])
		sm.Store(1, 1)
		var value int
		var ok bool
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				value, ok = sm.Load(1)
			}
		})
		_ = value
		_ = ok
	})

	b.Run("series-load-hit", func(b *testing.B) {
		var sm SortedMap[int, int]
		sm.cmp = cmp.Compare[int]
		sm.Store(1, 1)
		b.ReportAllocs()
		for b.Loop() {
			sm.Load(1)
		}
	})

	b.Run("parallel-load-miss", func(b *testing.B) {
		var sm SortedMap[int, int]
		sm.cmp = cmp.Compare[int]
		sm.Store(1, 1)
		var value int
		var ok bool
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				value, ok = sm.Load(2)
			}
		})
		_ = value
		_ = ok
	})

	b.Run("series-load-miss", func(b *testing.B) {
		var sm SortedMap[int, int]
		sm.cmp = cmp.Compare[int]
		sm.Store(1, 1)
		b.ReportAllocs()
		for b.Loop() {
			sm.Load(2)
		}
	})

	b.Run("values-100", func(b *testing.B) {
		var sm SortedMap[int, int]
		sm.cmp = cmp.Compare[int]
		for i := range 100 {
			sm.Store(i, i)
		}
		for range sm.Values() {
		}
		b.ReportAllocs()
		for b.Loop() {
			for range sm.Values() {
			}
		}
	})
}

func BenchmarkSyncMaps(b *testing.B) {
	b.Run("sync.Map", func(b *testing.B) {
		b.Run("insert-same", func(b *testing.B) {
			var m sync.Map

			b.ReportAllocs()
			for b.Loop() {
				m.Store(1, 1)
			}
		})

		b.Run("load-hit", func(b *testing.B) {
			var m sync.Map
			m.Store(1, 1)

			for b.Loop() {
				m.Load(1)
			}
		})
	})

	b.Run("map", func(b *testing.B) {
		b.Run("insert-same", func(b *testing.B) {
			var mu sync.Mutex
			m := map[int]int{}

			b.ReportAllocs()
			for b.Loop() {
				mu.Lock()
				m[1] = 1
				mu.Unlock()
			}
		})

		b.Run("insert-new", func(b *testing.B) {
			var mu sync.Mutex
			m := map[int]int{}

			b.ReportAllocs()
			for i := 0; b.Loop(); i++ {
				mu.Lock()
				m[i] = i
				mu.Unlock()
			}
		})

		b.Run("load-hit", func(b *testing.B) {
			var mu sync.Mutex
			m := map[int]int{
				1: 1,
			}

			for b.Loop() {
				mu.Lock()
				_ = m[1]
				mu.Unlock()
			}
		})
	})
}
