package dod

import (
	"math"

	"alp-go/alp"
	"alp-go/bitpack"
)

func Encode(dst []byte, src []int64) ([]byte, uint, int64) {
	if len(src) == 0 {
		return dst[:0], 0, 0
	}

	// d1 = v1 - v0
	// dod1 = d1 - d0
	d0 := src[0]
	minVal := int64(math.MaxInt64)
	forValues := make([]int64, len(src))
	forValues[0] = src[0]
	for i := 1; i < len(src); i++ {
		d1 := src[i] - src[i-1]
		dod1 := d1 - d0
		d0 = d1
		forValues[i] = dod1
		minVal = min(minVal, dod1)
	}
	for i, v := range forValues {
		forValues[i] = v - minVal
	}

	bitWidth := 0
	for _, v := range forValues {
		bw := alp.CalculateBitWidthSigned(v)
		bitWidth = max(bitWidth, bw)
	}

	n := bitpack.ByteCount(uint(len(forValues) * bitWidth))
	packedSize := n + bitpack.PaddingInt64
	packedData := make([]byte, packedSize)
	bitpack.PackInt64(packedData, forValues, uint(bitWidth))
	return packedData, uint(bitWidth), minVal
}

func Decode(dst []int64, src []byte, bitWidth uint, minVal int64) []int64 {
	if len(src) == 0 {
		return dst[:0]
	}

	// 10, 15, 22, 31, 55
	// 10, 5, 7, 9, 24
	// 10, -5, 2, 2, 15

	// dod = d1 - d0 -> d1 = dod + d0
	// d1 = v1 - v0  -> v1 = d1 + v0
	bitpack.UnpackInt64(dst, src, bitWidth)
	for i, v := range dst {
		dst[i] = v + minVal
	}

	d0 := dst[0]
	for i := 1; i < len(dst); i++ {
		dod := dst[i]
		d1 := dod + d0
		dst[i] = d1 + dst[i-1]
		d0 = d1
	}
	return dst
}
