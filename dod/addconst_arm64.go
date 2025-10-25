//go:build arm64

package dod

// addConstInt64 adds a constant to all elements in the slice using SIMD
func addConstInt64(dst []int64, c int64)

// addConstInt32 adds a constant to all elements in the slice using SIMD
func addConstInt32(dst []int32, c int32)

// addConstUint64 adds a constant to all elements in the slice using SIMD
func addConstUint64(dst []uint64, c uint64)
