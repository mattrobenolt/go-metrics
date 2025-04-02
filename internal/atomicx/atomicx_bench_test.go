package atomicx

import "testing"

func BenchmarkLoadParallel(b *testing.B) {
	b.Run("Float64", func(b *testing.B) {
		var f Float64
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				f.Load()
			}
		})
	})

	b.Run("Sum", func(b *testing.B) {
		var s Sum
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				s.Load()
			}
		})
	})
}

func BenchmarkAddParallel(b *testing.B) {
	b.Run("float/type=Float64", func(b *testing.B) {
		var f Float64
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				f.Add(1.1)
			}
		})
	})

	b.Run("integer/type=Float64", func(b *testing.B) {
		var f Float64
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				f.Add(1)
			}
		})
	})

	b.Run("float/type=Sum", func(b *testing.B) {
		var s Sum
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				s.Add(1.1)
			}
		})
	})

	b.Run("integer/type=Sum", func(b *testing.B) {
		var s Sum
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				s.Add(1)
			}
		})
	})
}

func BenchmarkAdd(b *testing.B) {
	b.Run("float/type=Float64", func(b *testing.B) {
		var f Float64
		for b.Loop() {
			f.Add(1.1)
		}
	})

	b.Run("integer/type=Float64", func(b *testing.B) {
		var f Float64
		for b.Loop() {
			f.Add(1)
		}
	})

	b.Run("float/type=Sum", func(b *testing.B) {
		var s Sum
		for b.Loop() {
			s.Add(1.1)
		}
	})

	b.Run("integer/type=Sum", func(b *testing.B) {
		var s Sum
		for b.Loop() {
			s.Add(1)
		}
	})
}
