package alp

import (
	"math/rand"
	"testing"
)

// BenchmarkSmall benchmarks small datasets (10 values)
func BenchmarkSmall(b *testing.B) {
	data := []float64{1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9, 10.0}

	b.Run("Compress", func(b *testing.B) {
		for b.Loop() {
			_ = Compress(data)
		}
	})

	b.Run("Decompress", func(b *testing.B) {
		compressed := Compress(data)
		b.ResetTimer()
		for b.Loop() {
			_ = Decompress(compressed)
		}
	})
}

// BenchmarkMedium benchmarks medium datasets (100 values)
func BenchmarkMedium(b *testing.B) {
	data := make([]float64, 100)
	for i := range data {
		data[i] = float64(i) * 0.1
	}

	b.Run("Compress", func(b *testing.B) {
		for b.Loop() {
			_ = Compress(data)
		}
	})

	b.Run("Decompress", func(b *testing.B) {
		compressed := Compress(data)
		b.ResetTimer()
		for b.Loop() {
			_ = Decompress(compressed)
		}
	})
}

// BenchmarkLarge benchmarks large datasets (10,000 values)
func BenchmarkLarge(b *testing.B) {
	data := make([]float64, 10000)
	for i := range data {
		data[i] = float64(i) * 0.1
	}

	b.Run("Compress", func(b *testing.B) {
		for b.Loop() {
			_ = Compress(data)
		}
	})

	b.Run("Decompress", func(b *testing.B) {
		compressed := Compress(data)
		b.ResetTimer()
		for b.Loop() {
			_ = Decompress(compressed)
		}
	})
}

// BenchmarkConstant benchmarks constant value datasets (1,000 values)
func BenchmarkConstant(b *testing.B) {
	data := make([]float64, 1000)
	for i := range data {
		data[i] = 42.5
	}

	b.Run("Compress", func(b *testing.B) {
		for b.Loop() {
			_ = Compress(data)
		}
	})

	b.Run("Decompress", func(b *testing.B) {
		compressed := Compress(data)
		b.ResetTimer()
		for b.Loop() {
			_ = Decompress(compressed)
		}
	})
}

// BenchmarkRandom benchmarks random value datasets (1,000 values)
func BenchmarkRandom(b *testing.B) {
	data := make([]float64, 100)
	for i := range data {
		data[i] = rand.Float64() * 1000
	}

	b.ResetTimer()
	b.Run("Compress", func(b *testing.B) {
		for b.Loop() {
			_ = Compress(data)
		}
	})

	b.Run("Decompress", func(b *testing.B) {
		compressed := Compress(data)
		b.ResetTimer()
		for b.Loop() {
			_ = Decompress(compressed)
		}
	})
}

// BenchmarkIntegerLike benchmarks integer-like floats (1,000 values)
func BenchmarkIntegerLike(b *testing.B) {
	data := make([]float64, 1000)
	for i := range data {
		data[i] = float64(i)
	}

	b.Run("Compress", func(b *testing.B) {
		for b.Loop() {
			_ = Compress(data)
		}
	})

	b.Run("Decompress", func(b *testing.B) {
		compressed := Compress(data)
		b.ResetTimer()
		for b.Loop() {
			_ = Decompress(compressed)
		}
	})
}

// BenchmarkOneDecimal benchmarks values with one decimal place (1,000 values)
func BenchmarkOneDecimal(b *testing.B) {
	data := make([]float64, 1000)
	for i := range data {
		data[i] = float64(i) * 0.1
	}

	b.Run("Compress", func(b *testing.B) {
		for b.Loop() {
			_ = Compress(data)
		}
	})

	b.Run("Decompress", func(b *testing.B) {
		compressed := Compress(data)
		b.ResetTimer()
		for b.Loop() {
			_ = Decompress(compressed)
		}
	})
}

// BenchmarkTwoDecimals benchmarks values with two decimal places (1,000 values)
func BenchmarkTwoDecimals(b *testing.B) {
	data := make([]float64, 1000)
	for i := range data {
		data[i] = float64(i) * 0.01
	}

	b.Run("Compress", func(b *testing.B) {
		for b.Loop() {
			_ = Compress(data)
		}
	})

	b.Run("Decompress", func(b *testing.B) {
		compressed := Compress(data)
		b.ResetTimer()
		for b.Loop() {
			_ = Decompress(compressed)
		}
	})
}

// BenchmarkBitPacking benchmarks bit packing operations
func BenchmarkBitPacking(b *testing.B) {
	values := make([]uint64, 1000)
	for i := range values {
		values[i] = uint64(i)
	}

	b.Run("Pack", func(b *testing.B) {
		for b.Loop() {
			_ = PackUint64Array(values, 16)
		}
	})

	b.Run("Unpack", func(b *testing.B) {
		packed := PackUint64Array(values, 16)
		b.ResetTimer()
		for b.Loop() {
			_ = UnpackUint64Array(packed, 1000, 16)
		}
	})
}

// BenchmarkExponentSearch benchmarks finding the best exponent
func BenchmarkExponentSearch(b *testing.B) {
	data := make([]float64, 1000)
	for i := range data {
		data[i] = float64(i) * 0.1
	}
	compressor := NewALPCompressor()

	for b.Loop() {
		_ = compressor.findBestExponent(data)
	}
}
