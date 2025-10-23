package dod

import (
	"alp-go/alp"
	"alp-go/bitpack"
	"encoding/binary"
	"math"
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

func Encode(dst []byte, src []int64) []byte {
	switch len(src) {
	case 0:
		return dst[:0]
	case 1:
		size := HeaderSize + Int64SizeBytes
		dst = make([]byte, size)
		dst[9] = uint8(len(src))
		binary.LittleEndian.PutUint64(dst[HeaderSize:size], uint64(src[0]))
		return dst[:size]
	}

	d0 := int64(0)
	minVal := int64(math.MaxInt64)
	encoded := make([]int64, len(src))
	encoded[0] = src[0]
	for i := 1; i < len(src); i++ {
		d1 := src[i] - src[i-1]
		dod1 := d1 - d0
		d0 = d1
		encoded[i] = dod1
		minVal = min(minVal, dod1)
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
	packedData := make([]byte, totalSize)

	// Encode header.
	binary.LittleEndian.PutUint64(packedData[:8], uint64(minVal))
	packedData[8] = uint8(bitWidth)
	packedData[9] = uint8(len(src))

	// Encode the first value as is and bitpack the rest.
	binary.LittleEndian.PutUint64(packedData[HeaderSize:HeaderSize+Int64SizeBytes], uint64(encoded[0]))
	bitpack.PackInt64(packedData[HeaderSize+Int64SizeBytes:], encoded[1:], uint(bitWidth))

	return packedData
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

	d0 := int64(0)
	for i := 1; i < int(header.NumValues); i++ {
		d1 := dst[i] + d0
		dst[i] = d1 + dst[i-1]
		d0 = d1
	}
	return header.NumValues
}
