package dod

import (
	"encoding/binary"
	"math"

	"github.com/parquet-go/bitpack"

	"github.com/fpetkovski/tscodec-go/alp"
	"github.com/fpetkovski/tscodec-go/delta"
)

type Int32Block [BlockSize]int32

func EncodeInt32(dst []byte, src []int32) []byte {
	switch len(src) {
	case 0:
		return dst
	case 1:
		if cap(dst) < delta.HeaderSize {
			dst = make([]byte, delta.HeaderSize)
		}
		dst = dst[:delta.HeaderSize]
		delta.EncodeHeader(dst, 1, int64(src[0]), 0)
		return dst
	}

	// Use int64 to avoid overflow when computing adjusted delta-of-deltas
	d0 := int64(0)
	minVal := int64(math.MaxInt64)
	encoded := make([]int64, len(src))
	encoded[0] = int64(src[0])
	for i := 1; i < len(src); i++ {
		d1 := int64(src[i]) - int64(src[i-1])
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
	totalSize := packedSize + delta.Int32SizeBytes + delta.HeaderSize + bitpack.PaddingInt64
	if cap(dst) < totalSize {
		dst = make([]byte, totalSize)
	}
	dst = dst[:totalSize]

	// Encode the first value as int32 and bitpack the rest as int64
	delta.EncodeHeader(dst, uint16(len(src)), minVal, uint8(bitWidth))
	binary.LittleEndian.PutUint32(dst[delta.HeaderSize:], uint32(encoded[0]))
	bitpack.Pack(dst[delta.HeaderSize+delta.Int32SizeBytes:], encoded[1:], uint(bitWidth))

	return dst
}

func DecodeInt32(dst []int32, src []byte) uint16 {
	if len(src) == 0 {
		return 0
	}

	header := delta.DecodeHeader(src)

	if header.NumValues == 1 {
		dst[0] = int32(header.MinVal)
		return 1
	}
	dst[0] = int32(binary.LittleEndian.Uint32(src[delta.HeaderSize:]))

	// Unpack the adjusted delta-of-deltas as int64 to match encoding
	adjustedDods := make([]int64, header.NumValues-1)
	bitpack.Unpack(adjustedDods, src[delta.HeaderSize+delta.Int32SizeBytes:], uint(header.BitWidth))

	numVals := int(header.NumValues)
	// Bounds check hint
	_ = dst[numVals-1]

	// Reconstruct DoD with minVal added back
	minVal := header.MinVal
	d0 := int64(0)
	prev := int64(dst[0])
	i := 1

	// Loop unrolling - process 4 elements at a time
	for ; i+3 < numVals; i += 4 {
		d1 := adjustedDods[i-1] + minVal + d0
		val := d1 + prev
		dst[i] = int32(val)
		prev = val
		d0 = d1

		d1 = adjustedDods[i] + minVal + d0
		val = d1 + prev
		dst[i+1] = int32(val)
		prev = val
		d0 = d1

		d1 = adjustedDods[i+1] + minVal + d0
		val = d1 + prev
		dst[i+2] = int32(val)
		prev = val
		d0 = d1

		d1 = adjustedDods[i+2] + minVal + d0
		val = d1 + prev
		dst[i+3] = int32(val)
		prev = val
		d0 = d1
	}

	// Handle remaining elements
	for ; i < numVals; i++ {
		d1 := adjustedDods[i-1] + minVal + d0
		val := d1 + prev
		dst[i] = int32(val)
		prev = val
		d0 = d1
	}
	return header.NumValues
}
