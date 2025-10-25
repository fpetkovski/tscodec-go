//go:build !s390x

package bitpack

import "github.com/fpetkovski/tscodec-go/unsafecast"

func unsafecastBytesToUint32(src []byte) []uint32 {
	return unsafecast.Slice[uint32](src)
}
