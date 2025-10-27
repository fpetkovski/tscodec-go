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

// Encode compresses an array of float64 values using ALP
func Encode(dst []byte, src []float64) []byte {
	switch {
	case len(src) == 0:
		if cap(dst) < MetadataSize {
			dst = make([]byte, MetadataSize)
		}
		dst = dst[:MetadataSize]
		encodeMetadata(dst, CompressionMetadata{
			EncodingType: EncodingNone,
			Count:        0,
		})
		return dst
	case isConstant(src):
		if cap(dst) < MetadataSize {
			dst = make([]byte, MetadataSize)
		}
		encodeMetadata(dst, CompressionMetadata{
			EncodingType:  EncodingConstant,
			Count:         int32(len(src)),
			ConstantValue: src[0],
		})
		return dst
	}

	// Find best exponent
	exponent, valid := findBestExponent(src)
	if !valid {
		// No valid exponent found - store raw float64 bytes
		totalSize := MetadataSize + len(src)*8
		if cap(dst) < totalSize {
			dst = make([]byte, totalSize)
		}
		dst = dst[:totalSize]
		encodeMetadata(dst, CompressionMetadata{
			EncodingType: EncodingNone,
			Count:        int32(len(src)),
		})
		// Store raw float64 bytes after metadata
		for i, v := range src {
			binary.LittleEndian.PutUint64(dst[MetadataSize+i*8:], math.Float64bits(v))
		}
		return dst
	}
	
	factor := powersOf10[exponent+10]

	// Convert to integers
	forValues := encodeToIntegers(src, factor)

	// Apply frame-of-reference encoding
	minValue := forValues[0]
	for _, v := range forValues {
		minValue = min(minValue, v)
	}
	for i, v := range forValues {
		forValues[i] = v - minValue
	}

	// Find bit-width for signed integers.
	bitWidth := 0
	for _, v := range forValues {
		bits := CalculateBitWidth(uint64(v))
		bitWidth = max(bitWidth, bits)
	}

	// Pack using signed integer packing.
	packedSize := bitpack.ByteCount(uint(len(forValues)*bitWidth)) + bitpack.PaddingInt64
	if cap(dst) < packedSize+MetadataSize {
		dst = make([]byte, packedSize+MetadataSize)
	}
	dst = dst[:packedSize]
	bitpack.PackInt64(dst[MetadataSize:], forValues, uint(bitWidth))

	// Create metadata
	metadata := CompressionMetadata{
		EncodingType: EncodingALP,
		Count:        int32(len(src)),
		Exponent:     int8(exponent),
		BitWidth:     uint8(bitWidth),
		FrameOfRef:   minValue,
	}

	// Combine metadata and src
	encodeMetadata(dst, metadata)
	return dst
}

// Decode decompresses ALP-encoded data
func Decode(dst []float64, data []byte) []float64 {
	if len(data) == 0 {
		return dst[:0]
	}

	// Decode metadata
	metadata := DecodeMetadata(data)

	// Validate count is within bounds
	if metadata.Count < 0 || metadata.Count > int32(len(dst)) {
		return dst[:0]
	}

	switch metadata.EncodingType {
	case EncodingNone:
		// Read raw float64 bytes
		if len(data) < MetadataSize+int(metadata.Count)*8 {
			return dst[:0]
		}
		for i := range metadata.Count {
			bits := binary.LittleEndian.Uint64(data[MetadataSize+i*8:])
			dst[i] = math.Float64frombits(bits)
		}
		return dst[:metadata.Count]

	case EncodingConstant:
		for i := range metadata.Count {
			dst[i] = metadata.ConstantValue
		}
		return dst[:metadata.Count]

	case EncodingALP:
		// Validate we have enough data
		if len(data) < MetadataSize {
			return dst[:0]
		}
		
		result := dst[:metadata.Count]
		ints := unsafecast.Slice[int64](result)
		bitpack.UnpackInt64(ints, data[MetadataSize:], uint(metadata.BitWidth))

		minValue := metadata.FrameOfRef
		numValues := metadata.Count

		// Use lookup table for power of 10.
		invFactor := powersOf10[(10-metadata.Exponent+21)%21]

		// Combined loop: add minValue and convert to float64 in one pass
		// This reduces memory traffic and allows better optimization
		i := int32(0)
		for ; i+3 < numValues; i += 4 {
			// Bounds check hint for the group of 4
			_ = ints[i+3]
			_ = result[i+3]

			result[i] = float64(ints[i]+minValue) * invFactor
			result[i+1] = float64(ints[i+1]+minValue) * invFactor
			result[i+2] = float64(ints[i+2]+minValue) * invFactor
			result[i+3] = float64(ints[i+3]+minValue) * invFactor
		}
		for ; i < numValues; i++ {
			result[i] = float64(ints[i]+minValue) * invFactor
		}

		return dst[:metadata.Count]
	default:
		return dst[:0]
	}
}

// findBestExponent analyzes the data and finds the best exponent for encoding
// Returns the exponent and a boolean indicating if a valid encoding was found
func findBestExponent(data []float64) (int, bool) {
	if len(data) == 0 {
		return 0, true
	}

	// Sample data if too large
	sampleSize := min(len(data), SamplingSize)

	bestExponent := 0
	minBitWidth := 64
	foundValid := false

	// Try different exponents
	for exp := MinExponent; exp <= MaxExponent; exp++ {
		factor := powersOf10[exp+10]
		invFactor := powersOf10[(10-exp+21)%21]
		maxBits := 0
		valid := true

		for i := range sampleSize {
			idx := i * len(data) / sampleSize
			original := data[idx]
			scaled := original * factor

			// Check for overflow, NaN, or Inf
			if math.IsNaN(scaled) || math.IsInf(scaled, 0) {
				valid = false
				break
			}

			// Check if conversion is lossless
			rounded := math.Round(scaled)
			
			// Check if the rounded value fits in int64 range
			if rounded > 9.223372036854775807e18 || rounded < -9.223372036854775808e18 {
				valid = false
				break
			}
			
			intValue := int64(rounded)

			// Reconstruct and check if lossless using same method as decompression
			reconstructed := float64(intValue) * invFactor
			absError := math.Abs(original - reconstructed)
			
			// Use relative error for non-zero values, absolute error for zero
			if original != 0 {
				relativeError := absError / math.Abs(original)
				if relativeError > 1e-12 {
					valid = false
					break
				}
			} else if absError > 1e-12 {
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
			foundValid = true
		}
	}

	return bestExponent, foundValid
}

// encodeToIntegers converts float64 values to integers using the factor
func encodeToIntegers(src []float64, factor float64) []int64 {
	result := make([]int64, len(src))
	for i, v := range src {
		scaled := v * factor
		
		// Check for overflow, NaN, or Inf - clamp to int64 range
		if math.IsNaN(scaled) || math.IsInf(scaled, 0) {
			if math.IsInf(scaled, 1) {
				result[i] = math.MaxInt64
			} else if math.IsInf(scaled, -1) {
				result[i] = math.MinInt64
			} else {
				result[i] = 0 // NaN case
			}
			continue
		}
		
		rounded := math.Round(scaled)
		
		// Clamp to int64 range
		if rounded > 9.223372036854775807e18 {
			result[i] = math.MaxInt64
		} else if rounded < -9.223372036854775808e18 {
			result[i] = math.MinInt64
		} else {
			result[i] = int64(rounded)
		}
	}
	return result
}

// DecompressValues decompresses ALP-encoded data
func DecompressValues(result []float64, src []byte, metadata CompressionMetadata) {
	clear(result)
	if len(src) == 0 {
		return
	}

	// Validate count is within bounds
	if metadata.Count < 0 || metadata.Count > int32(len(result)) {
		return
	}

	switch metadata.EncodingType {
	case EncodingNone:
		// Read raw float64 bytes
		if len(src) < MetadataSize+int(metadata.Count)*8 {
			return
		}
		for i := range metadata.Count {
			bits := binary.LittleEndian.Uint64(src[MetadataSize+i*8:])
			result[i] = math.Float64frombits(bits)
		}
	case EncodingConstant:
		for i := range metadata.Count {
			result[i] = metadata.ConstantValue
		}
	case EncodingALP:
		// Validate we have enough data
		if len(src) < MetadataSize {
			return
		}
		
		// Unpack src
		packedData := src[MetadataSize:]
		unpacked := unsafecast.Slice[int64](result)
		bitpack.UnpackInt64(unpacked[:metadata.Count], packedData, uint(metadata.BitWidth))

		// Reverse frame-of-reference and convert back to float64 in one pass
		minValue := metadata.FrameOfRef
		invFactor := powersOf10[(10-metadata.Exponent+21)%21]
		numValues := metadata.Count

		// Unroll loop for better performance
		i := int32(0)
		for ; i+3 < numValues; i += 4 {
			// Bounds check hint for the group of 4
			_ = unpacked[i+3]
			_ = result[i+3]

			result[i] = float64(unpacked[i]+minValue) * invFactor
			result[i+1] = float64(unpacked[i+1]+minValue) * invFactor
			result[i+2] = float64(unpacked[i+2]+minValue) * invFactor
			result[i+3] = float64(unpacked[i+3]+minValue) * invFactor
		}
		for ; i < numValues; i++ {
			result[i] = float64(unpacked[i]+minValue) * invFactor
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
func encodeMetadata(buf []byte, metadata CompressionMetadata) {
	buf[0] = byte(metadata.EncodingType)
	binary.LittleEndian.PutUint32(buf[1:5], uint32(metadata.Count))
	buf[5] = byte(metadata.Exponent)
	buf[6] = metadata.BitWidth
	binary.LittleEndian.PutUint64(buf[7:15], uint64(metadata.FrameOfRef))
	binary.LittleEndian.PutUint64(buf[15:23], math.Float64bits(metadata.ConstantValue))
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
