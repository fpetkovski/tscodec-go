package dod

import (
	"alp-go/alp"
	"alp-go/bitpack"
	"alp-go/delta"
	"encoding/binary"
	"math"
	"slices"
)

const (
	// BlockSize is the maximum amount of values that can be encoded at once.
	BlockSize = delta.Int64BlockSize
)

type Int64Block [BlockSize]int64

func EncodeInt64(dst []byte, src []int64) []byte {
	switch len(src) {
	case 0:
		return dst
	case 1:
		offset := len(dst)
		dst = slices.Grow(dst, delta.Int64HeaderSize)[:len(dst)+delta.Int64HeaderSize]
		out := dst[offset:]
		delta.EncodeInt64Header(out, 1, src[0], 0)
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
	totalSize := packedSize + delta.Int64SizeBytes + delta.Int64HeaderSize + bitpack.PaddingInt64
	offset := len(dst)
	dst = slices.Grow(dst, totalSize)[:len(dst)+totalSize]
	out := dst[offset:]

	delta.EncodeInt64Header(out, uint16(len(src)), minVal, uint8(bitWidth))

	// Encode the first value as is and bitpack the rest.
	binary.LittleEndian.PutUint64(out[delta.Int64HeaderSize:delta.Int64HeaderSize+delta.Int64SizeBytes], uint64(encoded[0]))
	bitpack.PackInt64(out[delta.Int64HeaderSize+delta.Int64SizeBytes:], encoded[1:], uint(bitWidth))

	return dst
}

func DecodeInt64(dst []int64, src []byte) uint16 {
	if len(src) == 0 {
		return 0
	}

	header := delta.DecodeInt64Header(src)

	if header.NumValues == 1 {
		dst[0] = header.MinVal
		return 1
	}
	dst[0] = int64(binary.LittleEndian.Uint64(src[delta.Int64HeaderSize : delta.Int64HeaderSize+delta.Int64SizeBytes]))
	bitpack.UnpackInt64(dst[1:header.NumValues], src[delta.Int64HeaderSize+delta.Int64SizeBytes:], uint(header.BitWidth))

	d0 := int64(0)
	prev := dst[0]
	minVal := header.MinVal
	i := 1
	_ = dst[header.NumValues-1]
	//for ; i+3 < int(header.NumValues); i += 4 {
	//	d1 := dst[i] + d0 + minVal
	//	dst[i] = d1 + prev
	//	d0 = d1
	//	prev = dst[i]
	//
	//	d1 = dst[i+1] + d0 + minVal
	//	dst[i+1] = d1 + prev
	//	d0 = d1
	//	prev = dst[i+1]
	//
	//	d1 = dst[i+2] + d0 + minVal
	//	dst[i+2] = d1 + prev
	//	d0 = d1
	//	prev = dst[i+2]
	//
	//	d1 = dst[i+3] + d0 + minVal
	//	dst[i+3] = d1 + prev
	//	d0 = d1
	//	prev = dst[i+3]
	//}
	for ; i < int(header.NumValues); i++ {
		d1 := dst[i] + d0 + minVal
		prev = d1 + prev
		dst[i] = prev
		d0 = d1
	}
	return header.NumValues
}
