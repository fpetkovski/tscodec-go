package delta

import (
	"math/rand"
	"slices"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name string
		src  []int64
	}{
		{
			name: "empty source",
			src:  nil,
		},
		{
			name: "single value",
			src:  []int64{3},
		},
		{
			name: "small input",
			src:  []int64{10, 15, 22, 31, 55},
		},
		{
			name: "large numbers",
			src:  []int64{100000, 100001, 100002, 100003, 100004, 100005},
		},
		{
			name: "irregular increments",
			src:  []int64{99968, 100001, 100002, 100003, 100004},
		},
		{
			name: "negative deltas",
			src:  []int64{100, 90, 80, 70, 60},
		},
		{
			name: "mixed deltas",
			src:  []int64{50, 100, 75, 125, 80},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("int64", func(t *testing.T) {
				encoded := EncodeInt64(nil, tc.src)

				var decoded Int64Block
				n := DecodeInt64(decoded[:], encoded)
				if !slices.Equal(tc.src, decoded[:n]) {
					t.Fatalf("Slices are not equal: got: [%v] want: [%v]", decoded[:n], tc.src)
				}
			})
			t.Run("int32", func(t *testing.T) {
				vals := make([]int32, len(tc.src))
				for i := range vals {
					vals[i] = int32(tc.src[i])
				}
				encoded := EncodeInt32(nil, vals)

				var decoded Int32Block
				n := DecodeInt32(decoded[:], encoded)
				if !slices.Equal(vals, decoded[:n]) {
					t.Fatalf("Slices are not equal: got: [%v] want: [%v]", decoded[:n], tc.src)
				}
			})
		})
	}
}

func FuzzInt64EncodeDecode(f *testing.F) {
	// Add seed corpus
	f.Add(uint8(10), int64(6))
	f.Add(uint8(20), int64(0))
	f.Add(uint8(30), int64(-300))

	f.Fuzz(func(t *testing.T, size uint8, seed int64) {
		src := make([]int64, size)
		gen := rand.New(rand.NewSource(seed))
		for i := range src {
			src[i] = gen.Int63()
		}

		var orig Int64Block
		n := DecodeInt64(orig[:], EncodeInt64(nil, src))
		if !slices.Equal(src, orig[:n]) {
			t.Fatalf("Roundtrip failed: got %v, want %v", orig[:n], src)
		}
	})
}

func FuzzInt32EncodeDecode(f *testing.F) {
	// Add seed corpus
	f.Add(uint8(10), int64(6))
	f.Add(uint8(20), int64(0))
	f.Add(uint8(30), int64(-300))

	f.Fuzz(func(t *testing.T, size uint8, seed int64) {
		src := make([]int32, size)
		gen := rand.New(rand.NewSource(seed))
		for i := range src {
			src[i] = gen.Int31()
		}

		var orig Int32Block
		n := DecodeInt32(orig[:], EncodeInt32(nil, src))
		if !slices.Equal(src, orig[:n]) {
			t.Fatalf("Roundtrip failed: got %v, want %v", orig[:n], src)
		}
	})
}
