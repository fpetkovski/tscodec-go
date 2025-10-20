package alp

import (
	"math/bits"
)

// BitPacker handles bit-packing of integer values
type BitPacker struct {
	buffer []byte
	pos    int // position in bits
}

// NewBitPacker creates a new bit packer
func NewBitPacker(capacity int) *BitPacker {
	return &BitPacker{
		buffer: make([]byte, 0, capacity),
		pos:    0,
	}
}

// PackUint64 packs a uint64 value using the specified number of bits
func (bp *BitPacker) PackUint64(value uint64, bitWidth int) {
	if bitWidth == 0 {
		return
	}

	// Mask to get only the bits we need
	mask := uint64(1<<bitWidth) - 1
	value &= mask

	bitsRemaining := bitWidth
	for bitsRemaining > 0 {
		byteIndex := bp.pos / 8
		bitOffset := bp.pos % 8

		// Ensure buffer has enough space
		for len(bp.buffer) <= byteIndex {
			bp.buffer = append(bp.buffer, 0)
		}

		// How many bits can we write to current byte
		bitsAvailable := 8 - bitOffset
		bitsToWrite := min(bitsRemaining, bitsAvailable)

		// Extract bits to write
		shift := bitsRemaining - bitsToWrite
		bitsValue := (value >> shift) & ((1 << bitsToWrite) - 1)

		// Write to buffer
		bp.buffer[byteIndex] |= byte(bitsValue << (bitsAvailable - bitsToWrite))

		bp.pos += bitsToWrite
		bitsRemaining -= bitsToWrite
	}
}

// PackInt64 packs a signed int64 value using the specified number of bits
func (bp *BitPacker) PackInt64(value int64, bitWidth int) {
	// Use zigzag encoding for signed values
	zigzag := uint64((value << 1) ^ (value >> 63))
	bp.PackUint64(zigzag, bitWidth)
}

// Bytes returns the packed bytes
func (bp *BitPacker) Bytes() []byte {
	return bp.buffer
}

// Reset resets the packer
func (bp *BitPacker) Reset() {
	bp.buffer = bp.buffer[:0]
	bp.pos = 0
}

// BitUnpacker handles unpacking of bit-packed values
type BitUnpacker struct {
	buffer []byte
	pos    int // position in bits
}

// NewBitUnpacker creates a new bit unpacker
func NewBitUnpacker(data []byte) *BitUnpacker {
	return &BitUnpacker{
		buffer: data,
		pos:    0,
	}
}

// UnpackUint64 unpacks a uint64 value using the specified number of bits
func (bu *BitUnpacker) UnpackUint64(bitWidth int) uint64 {
	if bitWidth == 0 {
		return 0
	}

	var value uint64
	bitsRemaining := bitWidth

	for bitsRemaining > 0 {
		byteIndex := bu.pos / 8
		bitOffset := bu.pos % 8

		if byteIndex >= len(bu.buffer) {
			return value
		}

		// How many bits can we read from current byte
		bitsAvailable := 8 - bitOffset
		bitsToRead := min(bitsRemaining, bitsAvailable)

		// Read bits
		mask := byte((1 << bitsToRead) - 1)
		shift := bitsAvailable - bitsToRead
		bitsValue := (bu.buffer[byteIndex] >> shift) & mask

		value = (value << bitsToRead) | uint64(bitsValue)

		bu.pos += bitsToRead
		bitsRemaining -= bitsToRead
	}

	return value
}

// UnpackInt64 unpacks a signed int64 value using the specified number of bits
func (bu *BitUnpacker) UnpackInt64(bitWidth int) int64 {
	zigzag := bu.UnpackUint64(bitWidth)
	// Decode zigzag
	return int64((zigzag >> 1) ^ (-(zigzag & 1)))
}

// Reset resets the unpacker to the beginning
func (bu *BitUnpacker) Reset() {
	bu.pos = 0
}

// CalculateBitWidth calculates the minimum bits needed to represent a uint64 value
func CalculateBitWidth(value uint64) int {
	if value == 0 {
		return 1
	}
	return 64 - bits.LeadingZeros64(value)
}

// CalculateBitWidthSigned calculates the minimum bits needed for a signed int64
func CalculateBitWidthSigned(value int64) int {
	zigzag := uint64((value << 1) ^ (value >> 63))
	return CalculateBitWidth(zigzag)
}

// FindMaxBitWidth finds the maximum bit width needed for an array of uint64
func FindMaxBitWidth(values []uint64) int {
	maxBits := 0
	for _, v := range values {
		maxBits = max(maxBits, CalculateBitWidth(v))
	}
	return maxBits
}

// PackUint64Array packs an array of uint64 values with the same bit width
func PackUint64Array(values []uint64, bitWidth int) []byte {
	packer := NewBitPacker(len(values) * bitWidth / 8)
	for _, v := range values {
		packer.PackUint64(v, bitWidth)
	}
	return packer.Bytes()
}

// UnpackUint64Array unpacks an array of uint64 values
func UnpackUint64Array(data []byte, count int, bitWidth int) []uint64 {
	unpacker := NewBitUnpacker(data)
	result := make([]uint64, count)
	for i := range count {
		result[i] = unpacker.UnpackUint64(bitWidth)
	}
	return result
}

// PackInt64Array packs an array of int64 values with the same bit width
func PackInt64Array(values []int64, bitWidth int) []byte {
	packer := NewBitPacker(len(values) * bitWidth / 8)
	for _, v := range values {
		packer.PackInt64(v, bitWidth)
	}
	return packer.Bytes()
}

// UnpackInt64Array unpacks an array of int64 values
func UnpackInt64Array(data []byte, count int, bitWidth int) []int64 {
	unpacker := NewBitUnpacker(data)
	result := make([]int64, count)
	for i := range count {
		result[i] = unpacker.UnpackInt64(bitWidth)
	}
	return result
}
