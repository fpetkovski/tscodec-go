package delta

import (
	"encoding/binary"
	"math"
	"slices"

	"github.com/fpetkovski/tscodec-go/alp"
	"github.com/fpetkovski/tscodec-go/bitpack"
)

const (
	// Int32BlockSize is the maximum amount of values that can be encoded at once.
	Int32BlockSize = 4096
)

type Int32Block [Int32BlockSize]int32

func EncodeInt32(dst []byte, src []int32) []byte {
	switch len(src) {
	case 0:
		return dst
	case 1:
		dst = slices.Grow(dst, Int64HeaderSize)[:Int64HeaderSize]
		EncodeInt64Header(dst, 1, int64(src[0]), 0)
		return dst
	}

	minVal := int32(math.MaxInt32)
	encoded := make([]int32, len(src))
	encoded[0] = src[0]
	for i := 1; i < len(src); i++ {
		delta := src[i] - src[i-1]
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
	totalSize := packedSize + Int64SizeBytes + Int64HeaderSize + bitpack.PaddingInt64
	dst = slices.Grow(dst, totalSize)[:totalSize]

	EncodeInt64Header(dst, uint16(len(src)), int64(minVal), uint8(bitWidth))

	// Encode the first value as is and bitpack the rest.
	binary.LittleEndian.PutUint64(dst[Int64HeaderSize:Int64HeaderSize+Int64SizeBytes], uint64(encoded[0]))
	bitpack.PackInt32(dst[Int64HeaderSize+Int64SizeBytes:], encoded[1:], uint(bitWidth))

	return dst
}

func DecodeInt32(dst []int32, src []byte) uint16 {
	if len(src) == 0 {
		return 0
	}

	header := DecodeInt64Header(src)

	if header.NumValues == 1 {
		dst[0] = int32(header.MinVal)
		return 1
	}
	dst[0] = int32(binary.LittleEndian.Uint32(src[Int64HeaderSize : Int64HeaderSize+Int64SizeBytes]))
	bitpack.UnpackInt32(dst[1:header.NumValues], src[Int64HeaderSize+Int64SizeBytes:], uint(header.BitWidth))

	// Combine adding minVal and computing prefix sum with loop unrolling
	minVal := int32(header.MinVal)
	numVals := int(header.NumValues)
	_ = dst[numVals-1] // Bounds check hint
	i := 1
	prev := dst[0]
	for ; i+3 < numVals; i += 4 {
		dst[i] = dst[i] + minVal + prev
		prev = dst[i]
		dst[i+1] = dst[i+1] + minVal + prev
		prev = dst[i+1]
		dst[i+2] = dst[i+2] + minVal + prev
		prev = dst[i+2]
		dst[i+3] = dst[i+3] + minVal + prev
		prev = dst[i+3]
	}
	prev = dst[i-1]
	for ; i < numVals; i++ {
		dst[i] = dst[i] + minVal + prev
		prev = dst[i]
	}
	return header.NumValues
}
