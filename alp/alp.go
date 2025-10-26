package alp

import (
	"encoding/binary"
	"math"

	"github.com/parquet-go/bitpack"
	"github.com/parquet-go/bitpack/unsafecast"
)

const (
	// MaxExponent is the maximum exponent to try for encoding
	MaxExponent = 10
	// MinExponent is the minimum exponent to try
	MinExponent = -10
	// SamplingSize is the number of values to sample for finding optimal encoding
	SamplingSize = 1024
	// MetadataSize is the size of metadata in bytes.
	MetadataSize = 23
)

// Pre-computed powers of 10 for fast lookup
var powersOf10 = [21]float64{
	1e-10, 1e-9, 1e-8, 1e-7, 1e-6, 1e-5, 1e-4, 1e-3, 1e-2, 1e-1,
	1e0,
	1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9, 1e10,
}

// EncodingType represents the type of encoding used
type EncodingType uint8

const (
	EncodingNone         EncodingType = 0
	EncodingALP          EncodingType = 1
	EncodingConstant     EncodingType = 2
	EncodingUncompressed EncodingType = 3
)

// CompressionMetadata contains metadata about the compressed data
type CompressionMetadata struct {
	EncodingType  EncodingType
	Count         int32
	Exponent      int8
	BitWidth      uint8
	FrameOfRef    int64
	ConstantValue float64
}

// Compress compresses an array of float64 values using ALP
func Compress(dst []byte, data []float64) []byte {
	if len(data) == 0 {
		return encodeMetadata(CompressionMetadata{
			EncodingType: EncodingNone,
			Count:        0,
		})
	}

	// Check for constant values
	if isConstant(data) {
		metadata := CompressionMetadata{
			EncodingType:  EncodingConstant,
			Count:         int32(len(data)),
			ConstantValue: data[0],
		}
		return encodeMetadata(metadata)
	}

	// Find best exponent
	exponent := findBestExponent(data)
	factor := powersOf10[exponent+10]

	// Convert to integers
	intValues := encodeToIntegers(data, factor)

	// Apply frame-of-reference encoding
	minValue := intValues[0]
	for _, v := range intValues {
		minValue = min(minValue, v)
	}

	// Apply frame-of-reference to create adjusted int64 values
	forValues := make([]int64, len(intValues))
	for i, v := range intValues {
		forValues[i] = v - minValue
	}

	// Find bit width for signed integers
	maxBits := 0
	for _, v := range forValues {
		bits := CalculateBitWidth(uint64(v))
		if bits > maxBits {
			maxBits = bits
		}
	}
	bitWidth := maxBits

	// Pack data using signed integer packing
	packedSize := bitpack.ByteCount(uint(len(forValues)*bitWidth)) + bitpack.PaddingInt64
	packedData := make([]byte, packedSize)
	bitpack.PackInt64(packedData, forValues, uint(bitWidth))

	// Create metadata
	metadata := CompressionMetadata{
		EncodingType: EncodingALP,
		Count:        int32(len(data)),
		Exponent:     int8(exponent),
		BitWidth:     uint8(bitWidth),
		FrameOfRef:   minValue,
	}

	// Combine metadata and data
	metadataBytes := encodeMetadata(metadata)
	result := make([]byte, len(metadataBytes)+len(packedData))
	copy(result, metadataBytes)
	copy(result[len(metadataBytes):], packedData)

	return result
}

// Decompress decompresses ALP-encoded data
func Decompress(dst []float64, data []byte) []float64 {
	if len(data) == 0 {
		return dst[:0]
	}

	// Decode metadata
	metadata := DecodeMetadata(data)

	switch metadata.EncodingType {
	case EncodingNone:
		return dst[:metadata.Count]

	case EncodingConstant:
		for i := range dst {
			dst[i] = metadata.ConstantValue
		}
		return dst[:metadata.Count]

	case EncodingALP:
		result := dst[:metadata.Count]
		ints := unsafecast.Slice[int64](result)
		bitpack.UnpackInt64(ints, data[MetadataSize:], uint(metadata.BitWidth))

		minValue := metadata.FrameOfRef
		numValues := metadata.Count

		// Use lookup table for power of 10.
		factor := powersOf10[metadata.Exponent+10]

		// Combined loop: add minValue and convert to float64 in one pass
		// This reduces memory traffic and allows better optimization
		i := int32(0)
		for ; i+3 < numValues; i += 4 {
			// Bounds check hint for the group of 4
			_ = ints[i+3]
			_ = result[i+3]

			result[i] = float64(ints[i]+minValue) / factor
			result[i+1] = float64(ints[i+1]+minValue) / factor
			result[i+2] = float64(ints[i+2]+minValue) / factor
			result[i+3] = float64(ints[i+3]+minValue) / factor
		}
		for ; i < numValues; i++ {
			result[i] = float64(ints[i]+minValue) / factor
		}

		return dst[:metadata.Count]
	default:
		return dst[:0]
	}
}

// findBestExponent analyzes the data and finds the best exponent for encoding
func findBestExponent(data []float64) int {
	if len(data) == 0 {
		return 0
	}

	// Sample data if too large
	sampleSize := min(len(data), SamplingSize)

	bestExponent := 0
	minBitWidth := 64

	// Try different exponents
	for exp := MinExponent; exp <= MaxExponent; exp++ {
		factor := powersOf10[exp+10]
		maxBits := 0
		valid := true

		for i := range sampleSize {
			idx := i * len(data) / sampleSize
			original := data[idx]
			scaled := original * factor

			// Check if conversion is lossless
			rounded := math.Round(scaled)
			intValue := int64(rounded)

			// Reconstruct and check if lossless
			reconstructed := float64(intValue) / factor
			relativeError := math.Abs(original - reconstructed)
			if original != 0 {
				relativeError /= math.Abs(original)
			}

			if relativeError > 1e-12 {
				// Not lossless at this exponent - skip this exponent entirely
				valid = false
				break
			}

			bits := CalculateBitWidthSigned(intValue)
			if bits > maxBits {
				maxBits = bits
			}

			// If bits are too large, this exponent won't be good
			if maxBits > 63 {
				valid = false
				break
			}
		}

		if valid && maxBits > 0 && maxBits < minBitWidth {
			minBitWidth = maxBits
			bestExponent = exp
		}
	}

	return bestExponent
}

// encodeToIntegers converts float64 values to integers using the factor
func encodeToIntegers(data []float64, factor float64) []int64 {
	result := make([]int64, len(data))
	for i, v := range data {
		scaled := v * factor
		result[i] = int64(math.Round(scaled))
	}
	return result
}

// DecompressValues decompresses ALP-encoded data
func DecompressValues(result []float64, src []byte, metadata CompressionMetadata) {
	clear(result)
	if len(src) == 0 {
		return
	}

	switch metadata.EncodingType {
	case EncodingNone:
	case EncodingConstant:
		for i := range result {
			result[i] = metadata.ConstantValue
		}
	case EncodingALP:
		// Unpack src
		packedData := src[MetadataSize:]
		unpacked := unsafecast.Slice[int64](result)
		bitpack.UnpackInt64(unpacked, packedData, uint(metadata.BitWidth))

		// Reverse frame-of-reference and convert back to float64 in one pass
		minValue := metadata.FrameOfRef
		factor := powersOf10[metadata.Exponent+10]
		for i := range metadata.Count {
			result[i] = float64(unpacked[i]+minValue) / factor
		}
	}
}

// isConstant checks if all values in the array are the same
func isConstant(data []float64) bool {
	if len(data) <= 1 {
		return true
	}

	first := data[0]
	for _, v := range data[1:] {
		if v != first {
			return false
		}
	}
	return true
}

// encodeMetadata encodes compression metadata to bytes
func encodeMetadata(metadata CompressionMetadata) []byte {
	buf := make([]byte, 32)
	buf[0] = byte(metadata.EncodingType)
	binary.LittleEndian.PutUint32(buf[1:5], uint32(metadata.Count))
	buf[5] = byte(metadata.Exponent)
	buf[6] = metadata.BitWidth
	binary.LittleEndian.PutUint64(buf[7:15], uint64(metadata.FrameOfRef))
	binary.LittleEndian.PutUint64(buf[15:23], math.Float64bits(metadata.ConstantValue))
	return buf[:23]
}

// DecodeMetadata decodes compression metadata from bytes
func DecodeMetadata(data []byte) CompressionMetadata {
	if len(data) < MetadataSize {
		return CompressionMetadata{EncodingType: EncodingNone}
	}

	return CompressionMetadata{
		EncodingType:  EncodingType(data[0]),
		Count:         int32(binary.LittleEndian.Uint32(data[1:5])),
		Exponent:      int8(data[5]),
		BitWidth:      data[6],
		FrameOfRef:    int64(binary.LittleEndian.Uint64(data[7:15])),
		ConstantValue: math.Float64frombits(binary.LittleEndian.Uint64(data[15:23])),
	}
}

// CompressionRatio calculates the compression ratio
func CompressionRatio(originalCount int, compressedSize int) float64 {
	originalSize := originalCount * 8 // float64 is 8 bytes
	if originalSize == 0 {
		return 0
	}
	return float64(compressedSize) / float64(originalSize)
}
