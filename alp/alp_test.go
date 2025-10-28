package alp

import (
	"math"
	"testing"

	"github.com/parquet-go/bitpack"
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
			name: "empty",
			data: []float64{},
		},
		{
			name: "constant",
			data: []float64{5.0, 5.0, 5.0, 5.0, 5.0},
		},
		{
			name: "zero values",
			data: []float64{0.0, 0.0, 0.0, 0.0},
		},
		{
			name: "single",
			data: []float64{5.0},
		},
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
		{
			name: "scientific notation",
			data: []float64{1e-5, 2e-5, 3e-5, 4e-5, 5e-5},
		},
		{
			name: "mixed ranges",
			data: []float64{0.1, 10.0, 100.0, 1000.0, 0.01},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			compressed := Encode(nil, tt.data)

			// Decode
			decompressed := make([]float64, len(tt.data))
			decompressed = Decode(decompressed, compressed)

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

func TestALPRandomDataset(t *testing.T) {
	data := make([]float64, 1000)
	for i := range data {
		data[i] = randGen.Float64() * 1e10
	}

	compressed := Encode(nil, data)
	decompressed := make([]float64, len(data))
	decompressed = Decode(decompressed, compressed)

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

	compressed := Encode(nil, data)
	decompressed := make([]float64, len(data))
	decompressed = Decode(decompressed, compressed)
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

func TestDecodeRange(t *testing.T) {
	tests := []struct {
		name  string
		data  []float64
		start int
		end   int
	}{
		{
			name:  "middle section",
			data:  []float64{1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9, 10.0},
			start: 3,
			end:   7,
		},
		{
			name:  "from start",
			data:  []float64{1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9, 10.0},
			start: 0,
			end:   5,
		},
		{
			name:  "to end",
			data:  []float64{1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9, 10.0},
			start: 5,
			end:   10,
		},
		{
			name:  "entire range",
			data:  []float64{1.1, 2.2, 3.3, 4.4, 5.5},
			start: 0,
			end:   5,
		},
		{
			name:  "single value",
			data:  []float64{1.1, 2.2, 3.3, 4.4, 5.5},
			start: 2,
			end:   3,
		},
		{
			name:  "constant values middle",
			data:  []float64{5.0, 5.0, 5.0, 5.0, 5.0, 5.0, 5.0, 5.0},
			start: 2,
			end:   6,
		},
		{
			name:  "negative values range",
			data:  []float64{-10.5, -5.5, 0.0, 5.5, 10.5, 15.5, 20.5},
			start: 1,
			end:   5,
		},
		{
			name:  "small decimals range",
			data:  []float64{0.001, 0.002, 0.003, 0.004, 0.005, 0.006, 0.007, 0.008},
			start: 2,
			end:   6,
		},
		{
			name:  "large dataset middle section",
			data:  func() []float64 {
				result := make([]float64, 1000)
				for i := range result {
					result[i] = float64(i) * 0.1
				}
				return result
			}(),
			start: 400,
			end:   600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode the full dataset
			compressed := Encode(nil, tt.data)

			// Decode the full dataset for reference
			fullDecoded := make([]float64, len(tt.data))
			fullDecoded = Decode(fullDecoded, compressed)

			// Decode the range
			rangeSize := tt.end - tt.start
			rangeDecoded := make([]float64, rangeSize)
			rangeDecoded = DecodeRange(rangeDecoded, compressed, tt.start, tt.end)

			// Verify length
			if len(rangeDecoded) != rangeSize {
				t.Errorf("Length mismatch: got %d, want %d", len(rangeDecoded), rangeSize)
			}

			// Verify values match the corresponding section of full decode
			for i := 0; i < rangeSize; i++ {
				originalIdx := tt.start + i
				if math.Abs(rangeDecoded[i]-fullDecoded[originalIdx]) > 1e-10 {
					t.Errorf("Value mismatch at range index %d (original index %d): got %f, want %f",
						i, originalIdx, rangeDecoded[i], fullDecoded[originalIdx])
				}
			}

			// Also verify against original data
			for i := 0; i < rangeSize; i++ {
				originalIdx := tt.start + i
				if math.Abs(rangeDecoded[i]-tt.data[originalIdx]) > 1e-10 {
					t.Errorf("Value mismatch against original at range index %d (original index %d): got %f, want %f",
						i, originalIdx, rangeDecoded[i], tt.data[originalIdx])
				}
			}
		})
	}
}

func TestDecodeRangeEdgeCases(t *testing.T) {
	data := []float64{1.1, 2.2, 3.3, 4.4, 5.5}
	compressed := Encode(nil, data)

	tests := []struct {
		name          string
		start         int
		end           int
		expectedLen   int
	}{
		{
			name:        "invalid range: start >= end",
			start:       5,
			end:         3,
			expectedLen: 0,
		},
		{
			name:        "invalid range: start < 0",
			start:       -1,
			end:         3,
			expectedLen: 0,
		},
		{
			name:        "invalid range: end > count",
			start:       0,
			end:         10,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := make([]float64, 10)
			result := DecodeRange(dst, compressed, tt.start, tt.end)
			if len(result) != tt.expectedLen {
				t.Errorf("Expected length %d, got %d", tt.expectedLen, len(result))
			}
		})
	}

	// Test empty data
	t.Run("empty data", func(t *testing.T) {
		emptyCompressed := Encode(nil, []float64{})
		dst := make([]float64, 10)
		result := DecodeRange(dst, emptyCompressed, 0, 0)
		if len(result) != 0 {
			t.Errorf("Expected empty result, got length %d", len(result))
		}
	})
}
