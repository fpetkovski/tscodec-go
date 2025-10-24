package dod

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

type Int64Block [BlockSize]int64

// The Int64HeaderSize is the size of the header of the encoded data.
const Int64HeaderSize = Int64SizeBytes + 1 + 1

type Header struct {
	BitWidth  uint8
	NumValues uint8
	MinVal    int64
}

func EncodeInt64(dst []byte, src []int64) []byte {
	switch len(src) {
	case 0:
		return dst
	case 1:
		offset := len(dst)
		dst = slices.Grow(dst, Int64HeaderSize)[:len(dst)+Int64HeaderSize]
		out := dst[offset:]
		binary.LittleEndian.PutUint64(out[:Int64SizeBytes], uint64(src[0]))
		out[Int64SizeBytes+1] = uint8(len(src))
		return dst
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
	totalSize := packedSize + Int64SizeBytes + Int64HeaderSize + bitpack.PaddingInt64
	offset := len(dst)
	dst = slices.Grow(dst, totalSize)[:len(dst)+totalSize]
	out := dst[offset:]

	// Encode header.
	binary.LittleEndian.PutUint64(out[:8], uint64(minVal))
	out[Int64SizeBytes] = uint8(bitWidth)
	out[Int64SizeBytes+1] = uint8(len(src))

	// Encode the first value as is and bitpack the rest.
	binary.LittleEndian.PutUint64(out[Int64HeaderSize:Int64HeaderSize+Int64SizeBytes], uint64(encoded[0]))
	bitpack.PackInt64(out[Int64HeaderSize+Int64SizeBytes:], encoded[1:], uint(bitWidth))

	return dst
}

func DecodeInt64(dst []int64, src []byte) uint8 {
	if len(src) == 0 {
		return 0
	}

	header := Header{
		MinVal:    int64(binary.LittleEndian.Uint64(src[:8])),
		BitWidth:  src[Int64SizeBytes],
		NumValues: src[Int64SizeBytes+1],
	}

	if header.NumValues == 1 {
		dst[0] = header.MinVal
		return 1
	}
	dst[0] = int64(binary.LittleEndian.Uint64(src[Int64HeaderSize : Int64HeaderSize+Int64SizeBytes]))
	bitpack.UnpackInt64(dst[1:header.NumValues], src[Int64HeaderSize+Int64SizeBytes:], uint(header.BitWidth))
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
