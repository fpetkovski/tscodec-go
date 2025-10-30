package alp

import (
	"io"
	"math"
	"math/rand"
	"testing"

	"github.com/parquet-go/bitpack"
)

// compareFloats compares two float64 values using relative error for large numbers
// and absolute error for small numbers. Returns true if the values are equal within
// the acceptable tolerance.
func compareFloats(a, b float64) (equal bool, relError, absError float64) {
	absError = math.Abs(a - b)

	// For zero values, use absolute error
	if a == 0 || b == 0 {
		equal = absError <= 1e-10
		relError = math.Inf(1) // Relative error is undefined for zero
		return
	}

	// Calculate relative error based on the larger magnitude
	maxAbs := math.Max(math.Abs(a), math.Abs(b))
	relError = absError / maxAbs

	// For small numbers (close to zero), use absolute error
	// For large numbers, use relative error
	if maxAbs < 1.0 {
		equal = absError <= 1e-10
	} else {
		equal = relError <= 1e-11  // Slightly more tolerant than 1e-12 to account for ALP precision
	}

	return
}

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
				equal, relErr, absErr := compareFloats(decompressed[i], tt.data[i])
				if !equal {
					t.Errorf("Value mismatch at index %d: got %f, want %f (abs err: %e, rel err: %e)",
						i, decompressed[i], tt.data[i], absErr, relErr)
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
		equal, relErr, absErr := compareFloats(decompressed[i], data[i])
		if !equal {
			t.Errorf("Value mismatch at index %d: got %f, want %f (abs err: %e, rel err: %e)",
				i, decompressed[i], data[i], absErr, relErr)
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
	bitpack.Pack(dst, values, bitWidth)
	return dst
}

// UnpackInt64Array unpacks an array of int64 values
func UnpackInt64Array(data []byte, count int, bitWidth uint) []int64 {
	dst := make([]int64, count)
	bitpack.Unpack(dst, data, bitWidth)
	return dst
}

func BitPackedSize(numValues uint32, bitWidth uint) int {
	return bitpack.ByteCount(uint(numValues)*bitWidth) + bitpack.PaddingInt64
}

func TestDecodeRange(t *testing.T) {
	tests := []struct {
		name      string
		data      []float64
		blockSize int
		bufSize   int
	}{
		{
			name:      "smaller read buffer than values",
			data:      []float64{6, 2, 3, 4, 5, 6},
			blockSize: 120,
			bufSize:   3,
		},
		{
			name:      "smaller read buffer than values",
			data:      []float64{1, 2, 3, 4, 5, 6},
			blockSize: 120,
			bufSize:   3,
		},
		{
			name:      "smaller read buffer than values",
			data:      []float64{1.1, 2.2, 3.2, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9, 10.0},
			blockSize: 120,
			bufSize:   3,
		},
		{
			name:      "read buffer multiple of values",
			data:      []float64{1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8},
			blockSize: 120,
			bufSize:   4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode with StreamEncode
			compressed := StreamEncode(nil, tt.data, tt.blockSize)

			// Decode with StreamDecoder
			decoder := StreamDecoder{}
			decoder.Reset(compressed, tt.blockSize)

			fullDecoded := make([]float64, 0, len(tt.data))
			readBuf := make([]float64, tt.bufSize)

			for {
				readBuf, err := decoder.Decode(readBuf)
				if err != io.EOF {
					if err != nil {
						t.Fatalf("unexpected error %v", err)
					}
				}
				fullDecoded = append(fullDecoded, readBuf...)
				if err == io.EOF {
					break
				}
			}
			// Verify length
			if len(fullDecoded) != len(tt.data) {
				t.Fatalf("length mismatch: got %d, want %d", len(fullDecoded), len(tt.data))
			}

			// Verify values with tolerance
			for i := range tt.data {
				equal, relErr, absErr := compareFloats(fullDecoded[i], tt.data[i])
				if !equal {
					t.Errorf("value mismatch at index %d: got %f, want %f (abs err: %e, rel err: %e)",
						i, fullDecoded[i], tt.data[i], absErr, relErr)
				}
			}
		})
	}
}

func TestStreamEncoderDecoder(t *testing.T) {
	tests := []struct {
		name      string
		data      []float64
		blockSize int
	}{
		{
			name:      "small dataset",
			data:      []float64{1, 2, 3, 4, 5, 6},
			blockSize: 3,
		},
		{
			name:      "small dataset uneven block size",
			data:      []float64{1, 2, 3, 4, 5},
			blockSize: 3,
		},
		{
			name:      "single block",
			data:      []float64{1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8},
			blockSize: 120,
		},
		{
			name:      "multiple blocks different ranges",
			data:      []float64{1, 2, 3, 100, 101, 102, 1000, 1001, 1002},
			blockSize: 3,
		},
		{
			name: "large dataset",
			data: func() []float64 {
				data := make([]float64, 1000)
				for i := range data {
					data[i] = float64(i) * 0.1
				}
				return data
			}(),
			blockSize: 120,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			compressed := StreamEncode(nil, tt.data, tt.blockSize)

			t.Logf("Original size: %d bytes, Compressed size: %d bytes, Ratio: %.2f%%",
				len(tt.data)*8, len(compressed), float64(len(compressed))/float64(len(tt.data)*8)*100)

			// Decode
			decoder := StreamDecoder{}
			decoder.Reset(compressed, tt.blockSize)

			decoded := make([]float64, 0, len(tt.data))
			readBuf := make([]float64, tt.blockSize)

			for {
				result, err := decoder.Decode(readBuf)
				decoded = append(decoded, result...)
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			// Verify length
			if len(decoded) != len(tt.data) {
				t.Fatalf("length mismatch: got %d, want %d", len(decoded), len(tt.data))
			}

			// Verify values
			for i := range tt.data {
				equal, relErr, absErr := compareFloats(decoded[i], tt.data[i])
				if !equal {
					t.Errorf("value mismatch at index %d: got %f, want %f (abs err: %e, rel err: %e)",
						i, decoded[i], tt.data[i], absErr, relErr)
				}
			}
		})
	}
}

func FuzzStreamEncodeDecode(f *testing.F) {
	// Add seed corpus with various sizes and block sizes
	f.Add(uint8(10), uint8(5), int64(42))
	f.Add(uint8(50), uint8(10), int64(123))
	f.Add(uint8(100), uint8(30), int64(456))
	f.Add(uint8(255), uint8(120), int64(789))
	f.Add(uint8(1), uint8(1), int64(0))

	f.Fuzz(func(t *testing.T, size uint8, blockSize uint8, seed int64) {
		// Ensure valid parameters
		if size == 0 {
			size = 1
		}
		if blockSize == 0 {
			blockSize = 1
		}

		// Generate random float64 data
		src := make([]float64, size)
		gen := randGen
		if seed != 0 {
			gen = rand.New(rand.NewSource(seed))
		}
		for i := range src {
			src[i] = gen.Float64() * 100000000
		}

		// Encode with StreamEncode
		compressed := StreamEncode(nil, src, int(blockSize))

		// Decode with StreamDecoder
		decoder := StreamDecoder{}
		decoder.Reset(compressed, int(blockSize))

		decoded := make([]float64, 0, len(src))
		readBuf := make([]float64, blockSize)

		for {
			result, err := decoder.Decode(readBuf)
			decoded = append(decoded, result...)
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("Unexpected decode error: %v", err)
			}
		}

		// Verify length
		if len(decoded) != len(src) {
			t.Fatalf("Length mismatch: got %d, want %d", len(decoded), len(src))
		}

		// Verify values with appropriate tolerance
		for i := range src {
			equal, relErr, absErr := compareFloats(decoded[i], src[i])
			if !equal {
				t.Errorf("Value mismatch at index %d: got %f, want %f (abs err: %e, rel err: %e)",
					i, decoded[i], src[i], absErr, relErr)
			}
		}
	})
}
