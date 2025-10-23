package alp

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/klauspost/compress/zstd"
)

// Benchmark ALP vs Zstd on sequential integers
func BenchmarkCompare_Sequential(b *testing.B) {
	data := make([]float64, 1000)
	for i := range data {
		data[i] = float64(i) * 0.1
	}

	b.Run("ALP/Compress", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		b.ResetTimer()
		for b.Loop() {
			_ = Compress(data)
		}
	})

	b.Run("ALP/Decompress", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		compressed := Compress(data)
		b.ResetTimer()
		for b.Loop() {
			_ = Decompress(compressed)
		}
	})

	b.Run("Zstd/Compress", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		encoder, _ := zstd.NewWriter(nil)
		dataBytes := float64sToBytes(data)
		b.ResetTimer()
		for b.Loop() {
			_ = encoder.EncodeAll(dataBytes, make([]byte, 0, len(dataBytes)))
		}
	})

	b.Run("Zstd/Decompress", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		encoder, _ := zstd.NewWriter(nil)
		decoder, _ := zstd.NewReader(nil)
		dataBytes := float64sToBytes(data)
		compressed := encoder.EncodeAll(dataBytes, make([]byte, 0, len(dataBytes)))
		b.ResetTimer()
		for b.Loop() {
			_, _ = decoder.DecodeAll(compressed, make([]byte, 0, len(dataBytes)))
		}
	})

	// Report compression ratios
	alpCompressed := Compress(data)
	encoder, _ := zstd.NewWriter(nil)
	dataBytes := float64sToBytes(data)
	zstdCompressed := encoder.EncodeAll(dataBytes, make([]byte, 0, len(dataBytes)))

	b.Logf("\nSequential Data (1000 values):")
	b.Logf("  Original: %d bytes", len(dataBytes))
	b.Logf("  ALP:  %d bytes (%.2f%% ratio, %.2f%% saved)",
		len(alpCompressed),
		float64(len(alpCompressed))/float64(len(dataBytes))*100,
		(1-float64(len(alpCompressed))/float64(len(dataBytes)))*100)
	b.Logf("  Zstd: %d bytes (%.2f%% ratio, %.2f%% saved)",
		len(zstdCompressed),
		float64(len(zstdCompressed))/float64(len(dataBytes))*100,
		(1-float64(len(zstdCompressed))/float64(len(dataBytes)))*100)
}

// Benchmark ALP vs Zstd on random sensor data
func BenchmarkCompare_RandomSensor(b *testing.B) {
	data := make([]float64, 1000)
	for i := range data {
		data[i] = math.Round((20.0+randGen.Float64()*10.0)*100) / 100
	}

	b.Run("ALP/Compress", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		b.ResetTimer()
		for b.Loop() {
			_ = Compress(data)
		}
	})

	b.Run("ALP/Decompress", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		compressed := Compress(data)
		b.ResetTimer()
		for b.Loop() {
			_ = Decompress(compressed)
		}
	})

	b.Run("Zstd/Compress", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		encoder, _ := zstd.NewWriter(nil)
		dataBytes := float64sToBytes(data)
		b.ResetTimer()
		for b.Loop() {
			_ = encoder.EncodeAll(dataBytes, make([]byte, 0, len(dataBytes)))
		}
	})

	b.Run("Zstd/Decompress", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		encoder, _ := zstd.NewWriter(nil)
		decoder, _ := zstd.NewReader(nil)
		dataBytes := float64sToBytes(data)
		compressed := encoder.EncodeAll(dataBytes, make([]byte, 0, len(dataBytes)))
		b.ResetTimer()
		for b.Loop() {
			_, _ = decoder.DecodeAll(compressed, make([]byte, 0, len(dataBytes)))
		}
	})

	// Report compression ratios
	alpCompressed := Compress(data)
	encoder, _ := zstd.NewWriter(nil)
	dataBytes := float64sToBytes(data)
	zstdCompressed := encoder.EncodeAll(dataBytes, make([]byte, 0, len(dataBytes)))

	b.Logf("\nRandom Sensor Data (1000 values, 2 decimal places):")
	b.Logf("  Original: %d bytes", len(dataBytes))
	b.Logf("  ALP:  %d bytes (%.2f%% ratio, %.2f%% saved)",
		len(alpCompressed),
		float64(len(alpCompressed))/float64(len(dataBytes))*100,
		(1-float64(len(alpCompressed))/float64(len(dataBytes)))*100)
	b.Logf("  Zstd: %d bytes (%.2f%% ratio, %.2f%% saved)",
		len(zstdCompressed),
		float64(len(zstdCompressed))/float64(len(dataBytes))*100,
		(1-float64(len(zstdCompressed))/float64(len(dataBytes)))*100)
}

// Benchmark ALP vs Zstd on constant values (best case for both)
func BenchmarkCompare_Constant(b *testing.B) {
	data := make([]float64, 1000)
	for i := range data {
		data[i] = 42.5
	}

	b.Run("ALP/Compress", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		b.ResetTimer()
		for b.Loop() {
			_ = Compress(data)
		}
	})

	b.Run("ALP/Decompress", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		compressed := Compress(data)
		b.ResetTimer()
		for b.Loop() {
			_ = Decompress(compressed)
		}
	})

	b.Run("Zstd/Compress", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		encoder, _ := zstd.NewWriter(nil)
		dataBytes := float64sToBytes(data)
		b.ResetTimer()
		for b.Loop() {
			_ = encoder.EncodeAll(dataBytes, make([]byte, 0, len(dataBytes)))
		}
	})

	b.Run("Zstd/Decompress", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		encoder, _ := zstd.NewWriter(nil)
		decoder, _ := zstd.NewReader(nil)
		dataBytes := float64sToBytes(data)
		compressed := encoder.EncodeAll(dataBytes, make([]byte, 0, len(dataBytes)))
		b.ResetTimer()
		for b.Loop() {
			_, _ = decoder.DecodeAll(compressed, make([]byte, 0, len(dataBytes)))
		}
	})

	// Report compression ratios
	alpCompressed := Compress(data)
	encoder, _ := zstd.NewWriter(nil)
	dataBytes := float64sToBytes(data)
	zstdCompressed := encoder.EncodeAll(dataBytes, make([]byte, 0, len(dataBytes)))

	b.Logf("\nConstant Values (1000 values):")
	b.Logf("  Original: %d bytes", len(dataBytes))
	b.Logf("  ALP:  %d bytes (%.2f%% ratio, %.2f%% saved)",
		len(alpCompressed),
		float64(len(alpCompressed))/float64(len(dataBytes))*100,
		(1-float64(len(alpCompressed))/float64(len(dataBytes)))*100)
	b.Logf("  Zstd: %d bytes (%.2f%% ratio, %.2f%% saved)",
		len(zstdCompressed),
		float64(len(zstdCompressed))/float64(len(dataBytes))*100,
		(1-float64(len(zstdCompressed))/float64(len(dataBytes)))*100)
}

// Benchmark ALP vs Zstd on truly random data (worst case)
func BenchmarkCompare_TrulyRandom(b *testing.B) {
	data := make([]float64, 1000)
	for i := range data {
		data[i] = randGen.Float64() * 1e10
	}

	b.Run("ALP/Compress", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		b.ResetTimer()
		for b.Loop() {
			_ = Compress(data)
		}
	})

	b.Run("ALP/Decompress", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		compressed := Compress(data)
		b.ResetTimer()
		for b.Loop() {
			_ = Decompress(compressed)
		}
	})

	b.Run("Zstd/Compress", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		encoder, _ := zstd.NewWriter(nil)
		dataBytes := float64sToBytes(data)
		b.ResetTimer()
		for b.Loop() {
			_ = encoder.EncodeAll(dataBytes, make([]byte, 0, len(dataBytes)))
		}
	})

	b.Run("Zstd/Decompress", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		encoder, _ := zstd.NewWriter(nil)
		decoder, _ := zstd.NewReader(nil)
		dataBytes := float64sToBytes(data)
		compressed := encoder.EncodeAll(dataBytes, make([]byte, 0, len(dataBytes)))
		b.ResetTimer()
		for b.Loop() {
			_, _ = decoder.DecodeAll(compressed, make([]byte, 0, len(dataBytes)))
		}
	})

	// Report compression ratios
	alpCompressed := Compress(data)
	encoder, _ := zstd.NewWriter(nil)
	dataBytes := float64sToBytes(data)
	zstdCompressed := encoder.EncodeAll(dataBytes, make([]byte, 0, len(dataBytes)))

	b.Logf("\nTruly Random Data (1000 values, full precision):")
	b.Logf("  Original: %d bytes", len(dataBytes))
	b.Logf("  ALP:  %d bytes (%.2f%% ratio, %.2f%% saved)",
		len(alpCompressed),
		float64(len(alpCompressed))/float64(len(dataBytes))*100,
		(1-float64(len(alpCompressed))/float64(len(dataBytes)))*100)
	b.Logf("  Zstd: %d bytes (%.2f%% ratio, %.2f%% saved)",
		len(zstdCompressed),
		float64(len(zstdCompressed))/float64(len(dataBytes))*100,
		(1-float64(len(zstdCompressed))/float64(len(dataBytes)))*100)
}

// Benchmark large dataset
func BenchmarkCompare_Large(b *testing.B) {
	data := make([]float64, 10000)
	for i := range data {
		data[i] = float64(i) * 0.1
	}

	b.Run("ALP/Compress", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		b.ResetTimer()
		for b.Loop() {
			_ = Compress(data)
		}
	})

	b.Run("ALP/Decompress", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		compressed := Compress(data)
		b.ResetTimer()
		for b.Loop() {
			_ = Decompress(compressed)
		}
	})

	b.Run("Zstd/Compress", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		encoder, _ := zstd.NewWriter(nil)
		dataBytes := float64sToBytes(data)
		b.ResetTimer()
		for b.Loop() {
			_ = encoder.EncodeAll(dataBytes, make([]byte, 0, len(dataBytes)))
		}
	})

	b.Run("Zstd/Decompress", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		encoder, _ := zstd.NewWriter(nil)
		decoder, _ := zstd.NewReader(nil)
		dataBytes := float64sToBytes(data)
		compressed := encoder.EncodeAll(dataBytes, make([]byte, 0, len(dataBytes)))
		b.ResetTimer()
		for b.Loop() {
			_, _ = decoder.DecodeAll(compressed, make([]byte, 0, len(dataBytes)))
		}
	})

	// Report compression ratios
	alpCompressed := Compress(data)
	encoder, _ := zstd.NewWriter(nil)
	dataBytes := float64sToBytes(data)
	zstdCompressed := encoder.EncodeAll(dataBytes, make([]byte, 0, len(dataBytes)))

	b.Logf("\nLarge Sequential Dataset (10,000 values):")
	b.Logf("  Original: %d bytes", len(dataBytes))
	b.Logf("  ALP:  %d bytes (%.2f%% ratio, %.2f%% saved)",
		len(alpCompressed),
		float64(len(alpCompressed))/float64(len(dataBytes))*100,
		(1-float64(len(alpCompressed))/float64(len(dataBytes)))*100)
	b.Logf("  Zstd: %d bytes (%.2f%% ratio, %.2f%% saved)",
		len(zstdCompressed),
		float64(len(zstdCompressed))/float64(len(dataBytes))*100,
		(1-float64(len(zstdCompressed))/float64(len(dataBytes)))*100)
}

// Test to verify lossless for ALP (Zstd should also be lossless for binary data)
func TestComparisonLossless(t *testing.T) {
	tests := []struct {
		name string
		data []float64
	}{
		{"sequential", []float64{1.0, 2.0, 3.0, 4.0, 5.0}},
		{"decimal", []float64{1.1, 2.2, 3.3, 4.4, 5.5}},
		{"negative", []float64{-10.5, -5.5, 0.0, 5.5, 10.5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test ALP
			alpCompressed := Compress(tt.data)
			alpDecompressed := Decompress(alpCompressed)

			for i := range tt.data {
				if alpDecompressed[i] != tt.data[i] {
					t.Errorf("ALP not lossless at index %d: got %v, want %v", i, alpDecompressed[i], tt.data[i])
				}
			}

			// Test Zstd
			encoder, _ := zstd.NewWriter(nil)
			decoder, _ := zstd.NewReader(nil)
			dataBytes := float64sToBytes(tt.data)
			zstdCompressed := encoder.EncodeAll(dataBytes, make([]byte, 0))
			zstdDecompressed, _ := decoder.DecodeAll(zstdCompressed, make([]byte, 0))
			decompressedData := bytesToFloat64s(zstdDecompressed)

			for i := range tt.data {
				if decompressedData[i] != tt.data[i] {
					t.Errorf("Zstd not lossless at index %d: got %v, want %v", i, decompressedData[i], tt.data[i])
				}
			}
		})
	}
}

// Helper function to convert float64 slice to bytes
func float64sToBytes(data []float64) []byte {
	buf := make([]byte, len(data)*8)
	for i, v := range data {
		binary.LittleEndian.PutUint64(buf[i*8:(i+1)*8], math.Float64bits(v))
	}
	return buf
}

// Helper function to convert bytes back to float64 slice
func bytesToFloat64s(data []byte) []float64 {
	result := make([]float64, len(data)/8)
	for i := range result {
		result[i] = math.Float64frombits(binary.LittleEndian.Uint64(data[i*8 : (i+1)*8]))
	}
	return result
}
