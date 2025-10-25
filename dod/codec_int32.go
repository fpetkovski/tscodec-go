package dod

import (
	"alp-go/alp"
	"alp-go/bitpack"
	"alp-go/delta"
	"encoding/binary"
	"math"
	"slices"
)

type Int32Block [BlockSize]int32

func EncodeInt32(dst []byte, src []int32) []byte {
	switch len(src) {
	case 0:
		return dst
	case 1:
		offset := len(dst)
		dst = slices.Grow(dst, delta.Int64HeaderSize)[:len(dst)+delta.Int64HeaderSize]
		out := dst[offset:]
		delta.EncodeInt64Header(out, 1, int64(src[0]), 0)
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
	totalSize := packedSize + delta.Int64SizeBytes + delta.Int64HeaderSize + bitpack.PaddingInt64
	offset := len(dst)
	dst = slices.Grow(dst, totalSize)[:len(dst)+totalSize]
	out := dst[offset:]

	delta.EncodeInt64Header(out, uint16(len(src)), int64(minVal), uint8(bitWidth))

	// Encode the first value as is and bitpack the rest.
	binary.LittleEndian.PutUint64(out[delta.Int64HeaderSize:delta.Int64HeaderSize+delta.Int64SizeBytes], uint64(encoded[0]))
	bitpack.PackInt32(out[delta.Int64HeaderSize+delta.Int64SizeBytes:], encoded[1:], uint(bitWidth))

	return dst
}

func DecodeInt32(dst []int32, src []byte) uint16 {
	if len(src) == 0 {
		return 0
	}

	header := delta.DecodeInt64Header(src)

	if header.NumValues == 1 {
		dst[0] = int32(header.MinVal)
		return 1
	}
	dst[0] = int32(binary.LittleEndian.Uint32(src[delta.Int64HeaderSize : delta.Int64HeaderSize+delta.Int64SizeBytes]))
	bitpack.UnpackInt32(dst[1:header.NumValues], src[delta.Int64HeaderSize+delta.Int64SizeBytes:], uint(header.BitWidth))

	numVals := int(header.NumValues)
	// Bounds check hint
	_ = dst[numVals-1]

	// Add minVal to all unpacked values first using SIMD
	minVal := int32(header.MinVal)
	addConstInt32(dst[1:numVals], minVal)

	// Now reconstruct DoD without minVal in critical path
	d0 := int32(0)
	prev := dst[0]
	i := 1

	// Loop unrolling - process 4 elements at a time
	for ; i+3 < numVals; i += 4 {
		d1 := dst[i] + d0
		prev = d1 + prev
		dst[i] = prev
		d0 = d1

		d1 = dst[i+1] + d0
		prev = d1 + prev
		dst[i+1] = prev
		d0 = d1

		d1 = dst[i+2] + d0
		prev = d1 + prev
		dst[i+2] = prev
		d0 = d1

		d1 = dst[i+3] + d0
		prev = d1 + prev
		dst[i+3] = prev
		d0 = d1
	}

	// Handle remaining elements
	for ; i < numVals; i++ {
		d1 := dst[i] + d0
		prev = d1 + prev
		dst[i] = prev
		d0 = d1
	}
	return header.NumValues
}
