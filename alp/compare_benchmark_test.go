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
	compressed := make([]byte, len(data))
	b.Run("ALP/Encode", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		b.ResetTimer()
		for b.Loop() {
			_ = Encode(compressed, data)
		}
	})

	b.Run("ALP/Decode", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		compressed := Encode(compressed, data)
		decompressed := make([]float64, len(data))
		b.ResetTimer()
		for b.Loop() {
			_ = Decode(decompressed, compressed)
		}
	})

	b.Run("Zstd/Encode", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		encoder, _ := zstd.NewWriter(nil)
		dataBytes := float64sToBytes(data)
		b.ResetTimer()
		for b.Loop() {
			_ = encoder.EncodeAll(dataBytes, make([]byte, 0, len(dataBytes)))
		}
	})

	b.Run("Zstd/Decode", func(b *testing.B) {
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
	alpCompressed := Encode(compressed, data)
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
	compressed := make([]byte, len(data))
	b.Run("ALP/Encode", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		b.ResetTimer()
		for b.Loop() {
			_ = Encode(compressed, data)
		}
	})

	b.Run("ALP/Decode", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		compressed := Encode(compressed, data)
		decompressed := make([]float64, len(data))
		b.ResetTimer()
		for b.Loop() {
			_ = Decode(decompressed, compressed)
		}
	})

	b.Run("Zstd/Encode", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		encoder, _ := zstd.NewWriter(nil)
		dataBytes := float64sToBytes(data)
		b.ResetTimer()
		for b.Loop() {
			_ = encoder.EncodeAll(dataBytes, make([]byte, 0, len(dataBytes)))
		}
	})

	b.Run("Zstd/Decode", func(b *testing.B) {
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
	alpCompressed := Encode(compressed, data)
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
	compressed := make([]byte, len(data))
	b.Run("ALP/Encode", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		b.ResetTimer()
		for b.Loop() {
			_ = Encode(compressed, data)
		}
	})

	b.Run("ALP/Decode", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		compressed = Encode(compressed, data)
		decompressed := make([]float64, len(data))
		b.ResetTimer()
		for b.Loop() {
			_ = Decode(decompressed, compressed)
		}
	})

	b.Run("Zstd/Encode", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		encoder, _ := zstd.NewWriter(nil)
		dataBytes := float64sToBytes(data)
		b.ResetTimer()
		for b.Loop() {
			_ = encoder.EncodeAll(dataBytes, make([]byte, 0, len(dataBytes)))
		}
	})

	b.Run("Zstd/Decode", func(b *testing.B) {
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
	alpCompressed := Encode(compressed, data)
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
	compressed := make([]byte, len(data))
	b.Run("ALP/Encode", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		b.ResetTimer()
		for b.Loop() {
			_ = Encode(compressed, data)
		}
	})

	b.Run("ALP/Decode", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		compressed := Encode(compressed, data)
		decompressed := make([]float64, len(data))
		b.ResetTimer()
		for b.Loop() {
			_ = Decode(decompressed, compressed)
		}
	})

	b.Run("Zstd/Encode", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		encoder, _ := zstd.NewWriter(nil)
		dataBytes := float64sToBytes(data)
		b.ResetTimer()
		for b.Loop() {
			_ = encoder.EncodeAll(dataBytes, make([]byte, 0, len(dataBytes)))
		}
	})

	b.Run("Zstd/Decode", func(b *testing.B) {
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
	alpCompressed := Encode(compressed, data)
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
	compressed := make([]byte, len(data))
	b.Run("ALP/Encode", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		b.ResetTimer()
		for b.Loop() {
			_ = Encode(compressed, data)
		}
	})

	b.Run("ALP/Decode", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		compressed := Encode(compressed, data)
		decompressed := make([]float64, len(data))
		b.ResetTimer()
		for b.Loop() {
			_ = Decode(decompressed, compressed)
		}
	})

	b.Run("Zstd/Encode", func(b *testing.B) {
		b.SetBytes(int64(len(data) * 8))
		encoder, _ := zstd.NewWriter(nil)
		dataBytes := float64sToBytes(data)
		b.ResetTimer()
		for b.Loop() {
			_ = encoder.EncodeAll(dataBytes, make([]byte, 0, len(dataBytes)))
		}
	})

	b.Run("Zstd/Decode", func(b *testing.B) {
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
	alpCompressed := Encode(compressed, data)
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

// floatEquals compares two float64 values with a relative tolerance
func floatEquals(a, b float64) bool {
	const epsilon = 1e-12
	if a == b {
		return true
	}
	diff := math.Abs(a - b)
	if a == 0 || b == 0 {
		return diff < epsilon
	}
	return diff/(math.Abs(a)+math.Abs(b)) < epsilon
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
			alpCompressed := make([]byte, len(tt.data))
			alpCompressed = Encode(alpCompressed, tt.data)
			alpDecompressed := make([]float64, len(tt.data))
			Decode(alpDecompressed, alpCompressed)

			for i := range tt.data {
				if !floatEquals(alpDecompressed[i], tt.data[i]) {
					t.Errorf("ALP not lossless at index %d: got %.17g, want %.17g", i, alpDecompressed[i], tt.data[i])
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
				if !floatEquals(decompressedData[i], tt.data[i]) {
					t.Errorf("Zstd not lossless at index %d: got %.17g, want %.17g", i, decompressedData[i], tt.data[i])
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

// Benchmark comparing regular ALP vs StreamEncoder with configurable block size
func BenchmarkCompare_StreamVsRegular(b *testing.B) {
	sizes := []int{1000, 10000}
	blockSize := 120

	patterns := map[string]func(int) []float64{
		"Sequential": func(size int) []float64 {
			data := make([]float64, size)
			for i := range data {
				data[i] = float64(i) * 0.1
			}
			return data
		},
		"Random": func(size int) []float64 {
			data := make([]float64, size)
			for i := range data {
				data[i] = randGen.Float64() * 1000
			}
			return data
		},
		"MixedRanges": func(size int) []float64 {
			data := make([]float64, size)
			for i := range data {
				if i%100 < 33 {
					data[i] = float64(i % 10)
				} else if i%100 < 66 {
					data[i] = float64(i % 1000)
				} else {
					data[i] = float64(i % 100000)
				}
			}
			return data
		},
	}

	for patternName, generator := range patterns {
		b.Logf("\n### Pattern: %s ###", patternName)
		for _, size := range sizes {
			data := generator(size)

			// Regular ALP
			regularCompressed := Encode(nil, data)

			// StreamEncoder with configured block size
			encoder := StreamEncoder{}
			encoder.Reset(blockSize)
			encoder.Encode(data)
			streamCompressed := encoder.Flush()

			overhead := len(streamCompressed) - len(regularCompressed)
			overheadPercent := float64(overhead) / float64(len(regularCompressed)) * 100

			b.Logf("Size %5d: Regular=%d bytes (%.2f%%), Stream=%d bytes (%.2f%%, blockSize=%d), overhead=%+d bytes (%+.1f%%)",
				size,
				len(regularCompressed),
				float64(len(regularCompressed))/float64(size*8)*100,
				len(streamCompressed),
				float64(len(streamCompressed))/float64(size*8)*100,
				blockSize,
				overhead,
				overheadPercent)
		}
	}
}
