package delta

import (
	"math/rand/v2"
	"testing"
)

const benchmarkSize = 4096

func BenchmarkInt64(b *testing.B) {
	b.Run("sequential", func(b *testing.B) {
		src := make([]int64, benchmarkSize)
		for i := range src {
			src[i] = int64(i)
		}

		b.Run("encode", func(b *testing.B) {
			dstBuf := make([]byte, 0, 40000)
			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				dst := EncodeInt64(dstBuf, src)
				_ = dst
			}
		})

		b.Run("decode", func(b *testing.B) {
			encoded := EncodeInt64(nil, src)
			dst := make([]int64, benchmarkSize)

			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				DecodeInt64(dst, encoded)
			}
		})
	})

	b.Run("small_deltas", func(b *testing.B) {
		src := make([]int64, benchmarkSize)
		val := int64(1000000)
		for i := range src {
			val += int64(rand.IntN(10))
			src[i] = val
		}

		b.Run("encode", func(b *testing.B) {
			dstBuf := make([]byte, 0, 40000)
			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				dst := EncodeInt64(dstBuf, src)
				_ = dst
			}
		})

		b.Run("decode", func(b *testing.B) {
			encoded := EncodeInt64(nil, src)
			dst := make([]int64, benchmarkSize)

			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				DecodeInt64(dst, encoded)
			}
		})
	})

	b.Run("large_deltas", func(b *testing.B) {
		src := make([]int64, benchmarkSize)
		val := int64(0)
		for i := range src {
			val += int64(rand.IntN(100000))
			src[i] = val
		}

		b.Run("encode", func(b *testing.B) {
			dstBuf := make([]byte, 0, 40000)
			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				dst := EncodeInt64(dstBuf, src)
				_ = dst
			}
		})

		b.Run("decode", func(b *testing.B) {
			encoded := EncodeInt64(nil, src)
			dst := make([]int64, benchmarkSize)

			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				DecodeInt64(dst, encoded)
			}
		})
	})

	b.Run("timestamps", func(b *testing.B) {
		src := make([]int64, benchmarkSize)
		ts := int64(1700000000000) // Unix milliseconds
		for i := range src {
			ts += 15000 + int64(rand.IntN(100)) // ~15s intervals with jitter
			src[i] = ts
		}

		b.Run("encode", func(b *testing.B) {
			dstBuf := make([]byte, 0, 40000)
			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				dst := EncodeInt64(dstBuf, src)
				_ = dst
			}
		})

		b.Run("decode", func(b *testing.B) {
			encoded := EncodeInt64(nil, src)
			dst := make([]int64, benchmarkSize)

			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				DecodeInt64(dst, encoded)
			}
		})
	})
}

func BenchmarkInt32(b *testing.B) {
	b.Run("sequential", func(b *testing.B) {
		src := make([]int32, benchmarkSize)
		for i := range src {
			src[i] = int32(i)
		}

		b.Run("encode", func(b *testing.B) {
			dstBuf := make([]byte, 0, 40000)
			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				dst := EncodeInt32(dstBuf, src)
				_ = dst
			}
		})

		b.Run("decode", func(b *testing.B) {
			encoded := EncodeInt32(nil, src)
			dst := make([]int32, benchmarkSize)

			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				DecodeInt32(dst, encoded)
			}
		})
	})

	b.Run("small_deltas", func(b *testing.B) {
		src := make([]int32, benchmarkSize)
		val := int32(1000000)
		for i := range src {
			val += int32(rand.IntN(10))
			src[i] = val
		}

		b.Run("encode", func(b *testing.B) {
			dstBuf := make([]byte, 0, 40000)
			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				dst := EncodeInt32(dstBuf, src)
				_ = dst
			}
		})

		b.Run("decode", func(b *testing.B) {
			encoded := EncodeInt32(nil, src)
			dst := make([]int32, benchmarkSize)

			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				DecodeInt32(dst, encoded)
			}
		})
	})

	b.Run("large_deltas", func(b *testing.B) {
		src := make([]int32, benchmarkSize)
		val := int32(0)
		for i := range src {
			val += int32(rand.IntN(10000))
			src[i] = val
		}

		b.Run("encode", func(b *testing.B) {
			dstBuf := make([]byte, 0, 40000)
			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				dst := EncodeInt32(dstBuf, src)
				_ = dst
			}
		})

		b.Run("decode", func(b *testing.B) {
			encoded := EncodeInt32(nil, src)
			dst := make([]int32, benchmarkSize)

			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				DecodeInt32(dst, encoded)
			}
		})
	})
}
