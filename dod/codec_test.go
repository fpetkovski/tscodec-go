package dod

import (
	"slices"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name string
		src  []int64
		want []byte
	}{
		{
			name: "empty source",
			src:  nil,
			want: nil,
		},
		{
			name: "small input",
			src:  []int64{10, 15, 22, 31, 55},
			want: nil,
		},
		{
			name: "large numbers",
			src:  []int64{100000, 100001, 100002, 100003},
			want: nil,
		},
		{
			name: "irregular increments",
			src:  []int64{99968, 100001, 100002, 100003, 100004},
			want: nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			encoded := Encode(nil, tc.src)

			var decoded Block
			n := Decode(decoded[:], encoded)
			if !slices.Equal(tc.src, decoded[:n]) {
				t.Fatalf("Slices are not equal: got: [%v] want: [%v]", decoded, tc.src)
			}
		})
	}
}

func FuzzEncodeDecode(f *testing.F) {
	// Add seed corpus
	f.Add(int64(10), int64(15), int64(22), int64(31), int64(55))
	f.Add(int64(100000), int64(100001), int64(100002), int64(100003), int64(100004))
	f.Add(int64(0), int64(0), int64(0), int64(0), int64(0))
	f.Add(int64(-100), int64(-50), int64(0), int64(50), int64(100))

	f.Fuzz(func(t *testing.T, v1, v2, v3, v4, v5 int64) {
		src := []int64{v1, v2, v3, v4, v5}

		// Encode.
		dst := Encode(nil, src)

		// Decode.
		var orig Block
		n := Decode(orig[:], dst)

		// Verify roundtrip
		if !slices.Equal(src, orig[:n]) {
			t.Fatalf("Roundtrip failed: got %v, want %v", orig, src)
		}
	})
}
