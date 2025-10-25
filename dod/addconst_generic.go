//go:build !arm64

package dod

// addConstInt64 adds a constant to all elements (generic fallback)
func addConstInt64(dst []int64, c int64) {
	for i := range dst {
		dst[i] += c
	}
}

// addConstInt32 adds a constant to all elements (generic fallback)
func addConstInt32(dst []int32, c int32) {
	for i := range dst {
		dst[i] += c
	}
}

// addConstUint64 adds a constant to all elements (generic fallback)
func addConstUint64(dst []uint64, c uint64) {
	for i := range dst {
		dst[i] += c
	}
}
