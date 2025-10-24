package dod

import (
	"alp-go/alp"
	"alp-go/bitpack"
	"encoding/binary"
	"math"
	"slices"
)

const (
	// Int32SizeBytes is the size in bytes of an int32 value.
	Int32SizeBytes = 4
)

// The Int32HeaderSize is the size of the header of the encoded data.
const Int32HeaderSize = Int32SizeBytes + 1 + 1

type Int32Block [BlockSize]int32

type Int32Header struct {
	BitWidth  uint8
	NumValues uint8
	MinVal    int32
}

func EncodeInt32(dst []byte, src []int32) []byte {
	switch len(src) {
	case 0:
		return dst
	case 1:
		offset := len(dst)
		dst = slices.Grow(dst, Int32HeaderSize)[:len(dst)+Int32HeaderSize]
		out := dst[offset:]
		binary.LittleEndian.PutUint32(out[:Int32SizeBytes], uint32(src[0]))
		out[Int32SizeBytes+1] = uint8(len(src))
		return dst
	}

	d0 := int32(0)
	minVal := int32(math.MaxInt32)
	encoded := make([]int32, len(src))
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
	totalSize := packedSize + Int32SizeBytes + Int32HeaderSize + bitpack.PaddingInt32
	offset := len(dst)
	dst = slices.Grow(dst, totalSize)[:len(dst)+totalSize]
	out := dst[offset:]

	// Encode header.
	binary.LittleEndian.PutUint32(out[:Int32SizeBytes], uint32(minVal))
	out[Int32SizeBytes] = uint8(bitWidth)
	out[Int32SizeBytes+1] = uint8(len(src))

	// Encode the first value as is and bitpack the rest.
	binary.LittleEndian.PutUint32(out[Int32HeaderSize:Int32HeaderSize+Int32SizeBytes], uint32(encoded[0]))
	bitpack.PackInt32(out[Int32HeaderSize+Int32SizeBytes:], encoded[1:], uint(bitWidth))

	return dst
}

func DecodeInt32(dst []int32, src []byte) uint8 {
	if len(src) == 0 {
		return 0
	}

	header := Int32Header{
		MinVal:    int32(binary.LittleEndian.Uint32(src[:Int32SizeBytes])),
		BitWidth:  src[Int32SizeBytes],
		NumValues: src[Int32SizeBytes+1],
	}

	if header.NumValues == 1 {
		dst[0] = header.MinVal
		return 1
	}
	dst[0] = int32(binary.LittleEndian.Uint32(src[Int32HeaderSize : Int32HeaderSize+Int32SizeBytes]))
	bitpack.UnpackInt32(dst[1:header.NumValues], src[Int32HeaderSize+Int32SizeBytes:], uint(header.BitWidth))
	for i := 1; i < int(header.NumValues); i++ {
		dst[i] = dst[i] + header.MinVal
	}

	d0 := int32(0)
	for i := 1; i < int(header.NumValues); i++ {
		d1 := dst[i] + d0
		dst[i] = d1 + dst[i-1]
		d0 = d1
	}
	return header.NumValues
}
