package alp

import (
	"math/bits"
)

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
