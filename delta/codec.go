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

	// BlockSize is the maximum amount of values that can be encoded at once.
	BlockSize = 255
)

type Block [BlockSize]int64

// The HeaderSize is the size of the header of the encoded data.
const HeaderSize = 8 + 1 + 1

type Header struct {
	BitWidth  uint8
	NumValues uint8
	MinVal    int64
}

func Encode(src []int64) []byte {
	return EncodeInt64(nil, src)
}

func EncodeInt64(dst []byte, src []int64) []byte {
	switch len(src) {
	case 0:
		return dst
	case 1:
		offset := len(dst)
		size := HeaderSize + Int64SizeBytes
		dst = slices.Grow(dst, size)[:len(dst)+size]
		out := dst[offset:]
		out[9] = uint8(len(src))
		binary.LittleEndian.PutUint64(out[HeaderSize:size], uint64(src[0]))
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
	totalSize := packedSize + Int64SizeBytes + HeaderSize + bitpack.PaddingInt64
	offset := len(dst)
	dst = slices.Grow(dst, totalSize)[:len(dst)+totalSize]
	out := dst[offset:]

	// Encode header.
	binary.LittleEndian.PutUint64(out[:8], uint64(minVal))
	out[8] = uint8(bitWidth)
	out[9] = uint8(len(src))

	// Encode the first value as is and bitpack the rest.
	binary.LittleEndian.PutUint64(out[HeaderSize:HeaderSize+Int64SizeBytes], uint64(encoded[0]))
	bitpack.PackInt64(out[HeaderSize+Int64SizeBytes:], encoded[1:], uint(bitWidth))

	return dst
}

func Decode(dst []int64, src []byte) uint8 {
	if len(src) == 0 {
		return 0
	}

	header := Header{
		MinVal:    int64(binary.LittleEndian.Uint64(src[:8])),
		BitWidth:  src[8],
		NumValues: src[9],
	}

	dst[0] = int64(binary.LittleEndian.Uint64(src[HeaderSize : HeaderSize+Int64SizeBytes]))
	if header.NumValues == 1 {
		return 1
	}
	bitpack.UnpackInt64(dst[1:header.NumValues], src[HeaderSize+Int64SizeBytes:], uint(header.BitWidth))
	for i := 1; i < int(header.NumValues); i++ {
		dst[i] = dst[i] + header.MinVal
	}

	for i := 1; i < int(header.NumValues); i++ {
		dst[i] = dst[i] + dst[i-1]
	}
	return header.NumValues
}
