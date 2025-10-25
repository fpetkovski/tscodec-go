package delta

import (
	"alp-go/alp"
	"alp-go/bitpack"
	"encoding/binary"
	"math"
	"slices"
)

const (
	// Int64SizeBytes is the size in bytes of an int64 value.
	Int64SizeBytes = 8

	// Int64BlockSize is the maximum amount of values that can be encoded at once.
	Int64BlockSize = 4096
)

type Int64Block [Int64BlockSize]int64

// The Int64HeaderSize is the size of the header of the encoded data.
const Int64HeaderSize = 8 + 1 + 2

type Int64Header struct {
	MinVal    int64
	NumValues uint16
	BitWidth  uint8
}

func EncodeInt64(dst []byte, src []int64) []byte {
	switch len(src) {
	case 0:
		return dst
	case 1:
		dst = slices.Grow(dst, Int64HeaderSize)[:Int64HeaderSize]
		EncodeInt64Header(dst, 1, src[0], 0)
		return dst
	}

	minVal := int64(math.MaxInt64)
	encoded := make([]int64, len(src))
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

	EncodeInt64Header(dst, uint16(len(src)), minVal, uint8(bitWidth))

	// Encode the first value as is and bitpack the rest.
	binary.LittleEndian.PutUint64(dst[Int64HeaderSize:Int64HeaderSize+Int64SizeBytes], uint64(encoded[0]))
	bitpack.PackInt64(dst[Int64HeaderSize+Int64SizeBytes:], encoded[1:], uint(bitWidth))

	return dst
}

func DecodeInt64(dst []int64, src []byte) uint16 {
	if len(src) == 0 {
		return 0
	}

	header := DecodeInt64Header(src)

	if header.NumValues == 1 {
		dst[0] = header.MinVal
		return 1
	}
	dst[0] = int64(binary.LittleEndian.Uint64(src[Int64HeaderSize : Int64HeaderSize+Int64SizeBytes]))
	bitpack.UnpackInt64(dst[1:header.NumValues], src[Int64HeaderSize+Int64SizeBytes:], uint(header.BitWidth))

	numVals := int(header.NumValues)
	// Bounds check hint.
	_ = dst[numVals-1]
	i := 1
	prev := dst[0]
	for ; i+3 < numVals; i += 4 {
		dst[i] = dst[i] + header.MinVal + prev
		prev = dst[i]
		dst[i+1] = dst[i+1] + header.MinVal + prev
		prev = dst[i+1]
		dst[i+2] = dst[i+2] + header.MinVal + prev
		prev = dst[i+2]
		dst[i+3] = dst[i+3] + header.MinVal + prev
		prev = dst[i+3]
	}
	prev = dst[i-1]
	for ; i < numVals; i++ {
		dst[i] = dst[i] + header.MinVal + prev
		prev = dst[i]
	}
	return header.NumValues
}

func EncodeInt64Header(dst []byte, numVals uint16, minVal int64, bitWidth uint8) {
	binary.LittleEndian.PutUint64(dst, uint64(minVal))
	binary.LittleEndian.PutUint16(dst[Int64SizeBytes:], numVals)
	dst[Int64SizeBytes+2] = bitWidth
}

func DecodeInt64Header(dst []byte) Int64Header {
	return Int64Header{
		MinVal:    int64(binary.LittleEndian.Uint64(dst)),
		NumValues: binary.LittleEndian.Uint16(dst[Int64SizeBytes:]),
		BitWidth:  dst[Int64SizeBytes+2],
	}
}
