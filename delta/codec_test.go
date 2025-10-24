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
			src:  []int64{100000, 100001, 100002, 100003},
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
			encoded := Encode(tc.src)

			var decoded Block
			n := Decode(decoded[:], encoded)
			if !slices.Equal(tc.src, decoded[:n]) {
				t.Fatalf("Slices are not equal: got: [%v] want: [%v]", decoded[:n], tc.src)
			}
		})
	}
}

func FuzzEncodeDecode(f *testing.F) {
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

		var orig Block
		n := Decode(orig[:], Encode(src))
		if !slices.Equal(src, orig[:n]) {
			t.Fatalf("Roundtrip failed: got %v, want %v", orig[:n], src)
		}
	})
}
