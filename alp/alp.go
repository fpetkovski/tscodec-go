package alp

import (
	"encoding/binary"
	"math"

	"alp-go/bitpack"
	"alp-go/unsafecast"
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

// ALPCompressor handles ALP compression of float64 data
type ALPCompressor struct {
	exponent int
	factor   float64
}

// NewALPCompressor creates a new ALP compressor
func NewALPCompressor() *ALPCompressor {
	return &ALPCompressor{}
}

// findBestExponent analyzes the data and finds the best exponent for encoding
func (ac *ALPCompressor) findBestExponent(data []float64) int {
	if len(data) == 0 {
		return 0
	}

	// Sample data if too large
	sampleSize := min(len(data), SamplingSize)

	bestExponent := 0
	minBitWidth := 64

	// Try different exponents
	for exp := MinExponent; exp <= MaxExponent; exp++ {
		factor := math.Pow(10, float64(exp))
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
func (ac *ALPCompressor) encodeToIntegers(data []float64) []int64 {
	result := make([]int64, len(data))
	for i, v := range data {
		scaled := v * ac.factor
		result[i] = int64(math.Round(scaled))
	}
	return result
}

// Compress compresses an array of float64 values using ALP
func (ac *ALPCompressor) Compress(data []float64) []byte {
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
	ac.exponent = ac.findBestExponent(data)
	ac.factor = math.Pow(10, float64(ac.exponent))

	// Convert to integers
	intValues := ac.encodeToIntegers(data)

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
		Exponent:     int8(ac.exponent),
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
func (ac *ALPCompressor) Decompress(dst []float64, data []byte) int {
	if len(data) == 0 {
		return 0
	}

	// Decode metadata
	metadata := DecodeMetadata(data)

	switch metadata.EncodingType {
	case EncodingNone:
		return int(metadata.Count)

	case EncodingConstant:
		for i := range dst {
			dst[i] = metadata.ConstantValue
		}
		return int(metadata.Count)

	case EncodingALP:
		result := dst[:metadata.Count]
		dst := unsafecast.Slice[int64](result)
		bitpack.UnpackInt64(dst, data[MetadataSize:], uint(metadata.BitWidth))

		minValue := metadata.FrameOfRef
		for i := range dst {
			dst[i] += minValue
		}

		// Convert back to float64
		factor := math.Pow(10, float64(metadata.Exponent))
		for i, v := range dst {
			result[i] = float64(v) / factor
		}

		return int(metadata.Count)

	default:
		return 0
	}
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

		// Reverse frame-of-reference
		minValue := metadata.FrameOfRef
		for i := range metadata.Count {
			unpacked[i] = unpacked[i] + minValue
		}

		// Convert back to float64
		factor := math.Pow(10, float64(metadata.Exponent))
		for i, v := range unpacked {
			result[i] = float64(v) / factor
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

// Compress is a convenience function to compress float64 data
func Compress(data []float64) []byte {
	compressor := NewALPCompressor()
	return compressor.Compress(data)
}

// Decompress is a convenience function to decompress ALP-encoded data
func Decompress(dst []float64, data []byte) int {
	compressor := NewALPCompressor()
	return compressor.Decompress(dst, data)
}

// CompressionRatio calculates the compression ratio
func CompressionRatio(originalCount int, compressedSize int) float64 {
	originalSize := originalCount * 8 // float64 is 8 bytes
	if originalSize == 0 {
		return 0
	}
	return float64(compressedSize) / float64(originalSize)
}
