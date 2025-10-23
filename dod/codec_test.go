package dod

import (
	"fmt"
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
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dst, bitWidth, minVal := Encode(nil, tc.src)
			fmt.Println(minVal)

			orig := make([]int64, len(tc.src))
			Decode(orig, dst, bitWidth, minVal)
			if !slices.Equal(tc.src, orig) {
				t.Fatalf("Slices are not equal: got: [%v] want: [%v]", orig, tc.src)
			}
		})
	}
}
