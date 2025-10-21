package alp

import (
	"math"
	"testing"

	"alp-go/bitpack"
)

func TestBitPacking(t *testing.T) {
	tests := []struct {
		name     string
		values   []int64
		bitWidth uint
	}{
		{
			name:     "8-bit values",
			values:   []int64{0, 1, 127, 255},
			bitWidth: 8,
		},
		{
			name:     "4-bit values",
			values:   []int64{0, 1, 7, 15},
			bitWidth: 4,
		},
		{
			name:     "16-bit values",
			values:   []int64{0, 100, 1000, 65535},
			bitWidth: 16,
		},
		{
			name:     "1-bit values",
			values:   []int64{0, 1, 1, 0, 1},
			bitWidth: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Pack
			packed := PackInt64Array(tt.values, tt.bitWidth)

			// Unpack
			unpacked := UnpackInt64Array(packed, len(tt.values), tt.bitWidth)

			// Verify
			if len(unpacked) != len(tt.values) {
				t.Errorf("Length mismatch: got %d, want %d", len(unpacked), len(tt.values))
			}

			for i := range tt.values {
				if unpacked[i] != tt.values[i] {
					t.Errorf("Value mismatch at index %d: got %d, want %d", i, unpacked[i], tt.values[i])
				}
			}
		})
	}
}

func TestALPCompression(t *testing.T) {
	tests := []struct {
		name string
		data []float64
	}{
		{
			name: "simple integers",
			data: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
		},
		{
			name: "decimal values",
			data: []float64{1.1, 2.2, 3.3, 4.4, 5.5},
		},
		{
			name: "larger range",
			data: []float64{100.5, 200.5, 300.5, 400.5, 500.5},
		},
		{
			name: "negative values",
			data: []float64{-10.5, -5.5, 0.0, 5.5, 10.5},
		},
		{
			name: "small decimals",
			data: []float64{0.001, 0.002, 0.003, 0.004, 0.005},
		},
		{
			name: "negative decimals",
			data: []float64{-0.001, 0.002, 0.003, -0.004, 0.005},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compress
			compressed := Compress(tt.data)

			// Decompress
			decompressed := Decompress(compressed)

			// Verify length
			if len(decompressed) != len(tt.data) {
				t.Errorf("Length mismatch: got %d, want %d", len(decompressed), len(tt.data))
			}

			// Verify values (lossless)
			for i := range tt.data {
				if math.Abs(decompressed[i]-tt.data[i]) > 1e-10 {
					t.Errorf("Value mismatch at index %d: got %f, want %f", i, decompressed[i], tt.data[i])
				}
			}

			// Check compression ratio
			originalSize := len(tt.data) * 8
			ratio := CompressionRatio(len(tt.data), len(compressed))
			t.Logf("Compression ratio: %.2f%% (original: %d bytes, compressed: %d bytes)",
				ratio*100, originalSize, len(compressed))
		})
	}
}

func TestALPEmptyData(t *testing.T) {
	data := []float64{}
	compressed := Compress(data)
	decompressed := Decompress(compressed)

	if len(decompressed) != 0 {
		t.Errorf("Expected empty result, got %d elements", len(decompressed))
	}
}

func TestALPConstantData(t *testing.T) {
	// All values are the same
	data := []float64{42.5, 42.5, 42.5, 42.5, 42.5}
	compressed := Compress(data)
	decompressed := Decompress(compressed)

	if len(decompressed) != len(data) {
		t.Errorf("Length mismatch: got %d, want %d", len(decompressed), len(data))
	}

	for i, v := range decompressed {
		if v != data[i] {
			t.Errorf("Value mismatch at index %d: got %f, want %f", i, v, data[i])
		}
	}

	// Constant compression should be very efficient
	if len(compressed) > 50 {
		t.Errorf("Constant compression not efficient: %d bytes", len(compressed))
	}
}

func TestALPSingleValue(t *testing.T) {
	data := []float64{123.456}
	compressed := Compress(data)
	decompressed := Decompress(compressed)

	if len(decompressed) != 1 {
		t.Errorf("Length mismatch: got %d, want 1", len(decompressed))
	}

	if math.Abs(decompressed[0]-data[0]) > 1e-10 {
		t.Errorf("Value mismatch: got %f, want %f", decompressed[0], data[0])
	}
}

func TestALPRandomDataset(t *testing.T) {
	data := make([]float64, 1000)
	for i := range data {
		data[i] = randGen.Float64() * 1e10
	}

	compressed := Compress(data)
	decompressed := Decompress(compressed)

	if len(decompressed) != len(data) {
		t.Errorf("Length mismatch: got %d, want %d", len(decompressed), len(data))
	}

	// Check a sample of values
	// Use relative error for large numbers instead of absolute error
	for i := 0; i < len(data); i += 100 {
		absError := math.Abs(decompressed[i] - data[i])
		relError := absError / math.Abs(data[i])

		// For large numbers, use relative error threshold
		// For small numbers, use absolute error threshold
		if data[i] != 0 && relError > 1e-12 {
			t.Errorf("Value mismatch at index %d: got %.15f, want %.15f (rel error: %e)", i, decompressed[i], data[i], relError)
		} else if data[i] == 0 && absError > 1e-9 {
			t.Errorf("Value mismatch at index %d: got %f, want %f (abs error: %e)", i, decompressed[i], data[i], absError)
		}
	}

	ratio := CompressionRatio(len(data), len(compressed))
	t.Logf("Large dataset compression ratio: %.2f%%", ratio*100)
}

func TestALPLargeDataset(t *testing.T) {
	// Generate a larger dataset
	data := make([]float64, 10000)
	for i := range data {
		data[i] = float64(i) * 0.1
	}

	compressed := Compress(data)
	decompressed := Decompress(compressed)

	if len(decompressed) != len(data) {
		t.Errorf("Length mismatch: got %d, want %d", len(decompressed), len(data))
	}

	// Check a sample of values
	for i := 0; i < len(data); i += 100 {
		if math.Abs(decompressed[i]-data[i]) > 1e-9 {
			t.Errorf("Value mismatch at index %d: got %f, want %f", i, decompressed[i], data[i])
		}
	}

	ratio := CompressionRatio(len(data), len(compressed))
	t.Logf("Large dataset compression ratio: %.2f%%", ratio*100)
}

func TestALPZeroValues(t *testing.T) {
	data := []float64{0.0, 0.0, 0.0, 0.0}
	compressed := Compress(data)
	decompressed := Decompress(compressed)

	if len(decompressed) != len(data) {
		t.Errorf("Length mismatch: got %d, want %d", len(decompressed), len(data))
	}

	for i, v := range decompressed {
		if v != 0.0 {
			t.Errorf("Expected 0.0 at index %d, got %f", i, v)
		}
	}
}

func TestCalculateBitWidth(t *testing.T) {
	tests := []struct {
		value    uint64
		expected int
	}{
		{0, 1},
		{1, 1},
		{2, 2},
		{3, 2},
		{4, 3},
		{7, 3},
		{8, 4},
		{255, 8},
		{256, 9},
		{65535, 16},
	}

	for _, tt := range tests {
		result := CalculateBitWidth(tt.value)
		if result != tt.expected {
			t.Errorf("CalculateBitWidth(%d) = %d, want %d", tt.value, result, tt.expected)
		}
	}
}

func TestFindMaxBitWidth(t *testing.T) {
	values := []uint64{1, 2, 3, 4, 255, 10}
	result := FindMaxBitWidth(values)
	expected := 8 // 255 requires 8 bits

	if result != expected {
		t.Errorf("FindMaxBitWidth() = %d, want %d", result, expected)
	}
}

func TestALPScientificNotation(t *testing.T) {
	data := []float64{1e-5, 2e-5, 3e-5, 4e-5, 5e-5}
	compressed := Compress(data)
	decompressed := Decompress(compressed)

	if len(decompressed) != len(data) {
		t.Errorf("Length mismatch: got %d, want %d", len(decompressed), len(data))
	}

	for i := range data {
		if math.Abs(decompressed[i]-data[i]) > 1e-15 {
			t.Errorf("Value mismatch at index %d: got %e, want %e", i, decompressed[i], data[i])
		}
	}
}

func TestALPMixedRange(t *testing.T) {
	data := []float64{0.1, 10.0, 100.0, 1000.0, 0.01}
	compressed := Compress(data)
	decompressed := Decompress(compressed)

	if len(decompressed) != len(data) {
		t.Errorf("Length mismatch: got %d, want %d", len(decompressed), len(data))
	}

	for i := range data {
		if math.Abs(decompressed[i]-data[i]) > 1e-10 {
			t.Errorf("Value mismatch at index %d: got %f, want %f", i, decompressed[i], data[i])
		}
	}
}

// PackInt64Array packs an array of uint64 values with the same bit width
func PackInt64Array(values []int64, bitWidth uint) []byte {
	dst := make([]byte, BitPackedSize(uint32(len(values)), bitWidth))
	bitpack.PackInt64(dst, values, bitWidth)
	return dst
}

// UnpackInt64Array unpacks an array of int64 values
func UnpackInt64Array(data []byte, count int, bitWidth uint) []int64 {
	dst := make([]int64, count)
	bitpack.UnpackInt64(dst, data, bitWidth)
	return dst
}

func BitPackedSize(numValues uint32, bitWidth uint) int {
	return bitpack.ByteCount(uint(numValues)*bitWidth) + bitpack.PaddingInt64
}
