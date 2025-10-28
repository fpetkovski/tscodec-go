package delta

import (
	"encoding/binary"
	"math"
	"slices"

	"github.com/parquet-go/bitpack"

	"github.com/fpetkovski/tscodec-go/alp"
)

const (
	// Int32BlockSize is the maximum amount of values that can be encoded at once.
	Int32BlockSize = 4096

	// Int32SizeBytes is the size in bytes of an int32.
	Int32SizeBytes = 4
)

type Int32Block [Int32BlockSize]int32

func EncodeInt32(dst []byte, src []int32) []byte {
	switch len(src) {
	case 0:
		return dst
	case 1:
		dst = slices.Grow(dst, HeaderSize)[:HeaderSize]
		EncodeHeader(dst, 1, int64(src[0]), 0)
		return dst
	}

	// Use int64 to avoid overflow when computing adjusted deltas
	minVal := int64(math.MaxInt64)
	encoded := make([]int64, len(src))
	encoded[0] = int64(src[0])
	for i := 1; i < len(src); i++ {
		delta := int64(src[i]) - int64(src[i-1])
		encoded[i] = delta
		minVal = min(minVal, delta)
	}
	for i := 1; i < len(encoded); i++ {
		encoded[i] = encoded[i] - minVal
	}

	bitWidth := 0
	for _, v := range encoded[1:] {
		bw := alp.CalculateBitWidth(uint64(v))
		bitWidth = max(bitWidth, bw)
	}

	packedSize := bitpack.ByteCount(uint((len(encoded) - 1) * bitWidth))
	totalSize := packedSize + Int32SizeBytes + HeaderSize + bitpack.PaddingInt64
	dst = slices.Grow(dst, totalSize)[:totalSize]

	EncodeHeader(dst, uint16(len(src)), minVal, uint8(bitWidth))

	// Encode the first value as int32 and bitpack the rest as int64
	binary.LittleEndian.PutUint32(dst[HeaderSize:], uint32(encoded[0]))
	bitpack.Pack(dst[HeaderSize+Int32SizeBytes:], encoded[1:], uint(bitWidth))

	return dst
}

func DecodeInt32(dst []int32, src []byte) uint16 {
	if len(src) == 0 {
		return 0
	}

	header := DecodeHeader(src)

	if header.NumValues == 1 {
		dst[0] = int32(header.MinVal)
		return 1
	}
	dst[0] = int32(binary.LittleEndian.Uint32(src[HeaderSize:]))

	// Unpack the adjusted deltas as int64 to match encoding
	adjustedDeltas := make([]int64, header.NumValues-1)
	bitpack.Unpack(adjustedDeltas, src[HeaderSize+Int32SizeBytes:], uint(header.BitWidth))

	// Combine adding minVal and computing prefix sum with loop unrolling
	minVal := header.MinVal
	numVals := int(header.NumValues)
	_ = dst[numVals-1] // Bounds check hint
	i := 1
	prev := int64(dst[0])
	for ; i+3 < numVals; i += 4 {
		val := adjustedDeltas[i-1] + minVal + prev
		dst[i] = int32(val)
		prev = val
		val = adjustedDeltas[i] + minVal + prev
		dst[i+1] = int32(val)
		prev = val
		val = adjustedDeltas[i+1] + minVal + prev
		dst[i+2] = int32(val)
		prev = val
		val = adjustedDeltas[i+2] + minVal + prev
		dst[i+3] = int32(val)
		prev = val
	}
	for ; i < numVals; i++ {
		val := adjustedDeltas[i-1] + minVal + prev
		dst[i] = int32(val)
		prev = val
	}
	return header.NumValues
}
