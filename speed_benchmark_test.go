package alp

import (
	"fmt"
	"math/rand"
	"testing"
)

// Dataset sizes to test
var benchmarkSizes = []int{
	10,      // Tiny
	100,     // Small
	1000,    // Medium
	10000,   // Large
	100000,  // Very Large
	1000000, // Huge
}

// BenchmarkCompressionSpeed measures compression throughput across different sizes
func BenchmarkCompressionSpeed(b *testing.B) {
	for _, size := range benchmarkSizes {
		// Generate sequential data
		data := make([]float64, size)
		for i := range data {
			data[i] = float64(i) * 0.1
		}

		name := fmt.Sprintf("Sequential_%d", size)
		b.Run(name, func(b *testing.B) {
			b.SetBytes(int64(size * 8)) // 8 bytes per float64
			b.ResetTimer()
			for b.Loop() {
				_ = Compress(data)
			}
		})
	}
}

// BenchmarkDecompressionSpeed measures decompression throughput across different sizes
func BenchmarkDecompressionSpeed(b *testing.B) {
	for _, size := range benchmarkSizes {
		// Generate and compress sequential data
		data := make([]float64, size)
		for i := range data {
			data[i] = float64(i) * 0.1
		}
		compressed := Compress(data)

		name := fmt.Sprintf("Sequential_%d", size)
		b.Run(name, func(b *testing.B) {
			b.SetBytes(int64(size * 8)) // 8 bytes per float64
			b.ResetTimer()
			for b.Loop() {
				_ = Decompress(compressed)
			}
		})
	}
}

// BenchmarkCompressionLatency measures latency (time per operation) without throughput
func BenchmarkCompressionLatency(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		data := make([]float64, size)
		for i := range data {
			data[i] = float64(i) * 0.1
		}

		name := fmt.Sprintf("%d_values", size)
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			for b.Loop() {
				_ = Compress(data)
			}
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/1000.0, "µs/op")
		})
	}
}

// BenchmarkDecompressionLatency measures decompression latency
func BenchmarkDecompressionLatency(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		data := make([]float64, size)
		for i := range data {
			data[i] = float64(i) * 0.1
		}
		compressed := Compress(data)

		name := fmt.Sprintf("%d_values", size)
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			for b.Loop() {
				_ = Decompress(compressed)
			}
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/1000.0, "µs/op")
		})
	}
}

// BenchmarkRoundTripSpeed measures full compress + decompress cycle
func BenchmarkRoundTripSpeed(b *testing.B) {
	for _, size := range []int{100, 1000, 10000} {
		data := make([]float64, size)
		for i := range data {
			data[i] = float64(i) * 0.1
		}

		name := fmt.Sprintf("%d_values", size)
		b.Run(name, func(b *testing.B) {
			b.SetBytes(int64(size * 8 * 2)) // Count both compress and decompress
			b.ResetTimer()
			for b.Loop() {
				compressed := Compress(data)
				_ = Decompress(compressed)
			}
		})
	}
}

// BenchmarkSpeedByPattern measures speed across different data patterns
func BenchmarkSpeedByPattern(b *testing.B) {
	size := 1000

	patterns := map[string]func() []float64{
		"Integers": func() []float64 {
			data := make([]float64, size)
			for i := range data {
				data[i] = float64(i)
			}
			return data
		},
		"OneDecimal": func() []float64 {
			data := make([]float64, size)
			for i := range data {
				data[i] = float64(i) * 0.1
			}
			return data
		},
		"TwoDecimals": func() []float64 {
			data := make([]float64, size)
			for i := range data {
				data[i] = float64(i) * 0.01
			}
			return data
		},
		"Constant": func() []float64 {
			data := make([]float64, size)
			for i := range data {
				data[i] = 42.5
			}
			return data
		},
		"RandomSensor": func() []float64 {
			data := make([]float64, size)
			for i := range data {
				data[i] = 20.0 + rand.Float64()*10.0
			}
			return data
		},
	}

	for patternName, generator := range patterns {
		data := generator()

		b.Run("Compress/"+patternName, func(b *testing.B) {
			b.SetBytes(int64(size * 8))
			b.ResetTimer()
			for b.Loop() {
				_ = Compress(data)
			}
		})

		compressed := Compress(data)
		b.Run("Decompress/"+patternName, func(b *testing.B) {
			b.SetBytes(int64(size * 8))
			b.ResetTimer()
			for b.Loop() {
				_ = Decompress(compressed)
			}
		})
	}
}

// BenchmarkParallelCompression tests compression in parallel
func BenchmarkParallelCompression(b *testing.B) {
	data := make([]float64, 1000)
	for i := range data {
		data[i] = float64(i) * 0.1
	}

	b.SetBytes(int64(len(data) * 8))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Compress(data)
		}
	})
}

// BenchmarkParallelDecompression tests decompression in parallel
func BenchmarkParallelDecompression(b *testing.B) {
	data := make([]float64, 1000)
	for i := range data {
		data[i] = float64(i) * 0.1
	}
	compressed := Compress(data)

	b.SetBytes(int64(len(data) * 8))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Decompress(compressed)
		}
	})
}

// BenchmarkScalability shows how performance scales with data size
func BenchmarkScalability(b *testing.B) {
	// Test if performance is linear with data size
	sizes := []int{100, 500, 1000, 5000, 10000, 50000, 100000}

	for _, size := range sizes {
		data := make([]float64, size)
		for i := range data {
			data[i] = float64(i) * 0.1
		}

		b.Run(fmt.Sprintf("Compress_%d", size), func(b *testing.B) {
			b.SetBytes(int64(size * 8))
			b.ResetTimer()
			for b.Loop() {
				_ = Compress(data)
			}
		})
	}
}

// BenchmarkMemoryEfficiency measures allocations per operation
func BenchmarkMemoryEfficiency(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		data := make([]float64, size)
		for i := range data {
			data[i] = float64(i) * 0.1
		}

		b.Run(fmt.Sprintf("Compress_%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				_ = Compress(data)
			}
		})

		compressed := Compress(data)
		b.Run(fmt.Sprintf("Decompress_%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				_ = Decompress(compressed)
			}
		})
	}
}

// BenchmarkCPUBound tests if compression is CPU bound
func BenchmarkCPUBound(b *testing.B) {
	data := make([]float64, 10000)
	for i := range data {
		data[i] = float64(i) * 0.1
	}

	// Single threaded
	b.Run("SingleThread", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		b.ResetTimer()
		for b.Loop() {
			_ = Compress(data)
		}
	})

	// Parallel
	b.Run("Parallel", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		b.SetParallelism(4)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = Compress(data)
			}
		})
	})
}
