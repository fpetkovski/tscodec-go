package benchmarks

import (
	"alp-go/bitpack"
	"alp-go/delta"
	pqdelta "github.com/parquet-go/parquet-go/encoding/delta"
	"github.com/parquet-go/parquet-go/encoding/rle"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/tsdb/tsdbutil"
	"math/rand/v2"
	"testing"
	"time"

	"alp-go/alp"
	"alp-go/dod"

	"github.com/prometheus/prometheus/tsdb/chunkenc"
	"github.com/stretchr/testify/require"
)

func BenchmarkFloatEncoding(b *testing.B) {
	const numSamples = 120
	t := time.Now().UnixMilli()
	v := float64(10_000)
	ts := make([]int64, numSamples)
	vs := make([]float64, numSamples)
	for i := range numSamples {
		ts[i] = t + int64(15_000*i) + int64(rand.Float64()*100)
		vs[i] = v + rand.Float64()*1000
	}

	b.Run("prometheus_encode", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			chk := chunkenc.NewXORChunk()
			app, err := chk.Appender()
			require.NoError(b, err)
			for i := range ts {
				app.Append(ts[i], vs[i])
			}
			b.ReportMetric(float64(len(chk.Bytes())), "compressed_bytes")
		}
	})
	b.Run("prometheus_decode", func(b *testing.B) {
		b.ReportAllocs()

		chk := chunkenc.NewXORChunk()
		app, err := chk.Appender()
		require.NoError(b, err)

		for i := range ts {
			app.Append(ts[i], vs[i])
		}
		for b.Loop() {
			it := chk.Iterator(nil)
			for it.Next() != chunkenc.ValNone {
				_, _ = it.At()
			}
		}
	})

	b.Run("alp-encode", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			tsc := dod.EncodeInt64(nil, ts)
			fsc := alp.Compress(vs)

			b.ReportMetric(float64(len(tsc)+len(fsc)), "compressed_bytes")
		}
	})
	b.Run("parquet-encode", func(b *testing.B) {
		enc := pqdelta.BinaryPackedEncoding{}

		_, err := enc.EncodeInt64(nil, ts)
		require.NoError(b, err)
	})
	b.Run("alp-decode", func(b *testing.B) {
		b.ReportAllocs()

		tsc := dod.EncodeInt64(nil, ts)
		fsc := alp.Compress(vs)

		var (
			ints   dod.Int64Block
			floats = make([]float64, numSamples)
		)
		for b.Loop() {
			dod.DecodeInt64(ints[:], tsc)
			_ = alp.Decompress(floats, fsc)
		}
	})
}

func BenchmarkHistogramEncoding(b *testing.B) {
	h := prometheus.NewHistogram(prometheus.HistogramOpts{})
	h.Observe(10)
	reg := prometheus.NewRegistry()
	reg.Gather()

	const numSamples = 120
	t := time.Now().UnixMilli()
	ts := make([]int64, numSamples)
	hs := make([]*histogram.Histogram, numSamples)
	for i := range numSamples {
		ts[i] = t + int64(15_000*i) + int64(rand.Float64()*10)
		hs[i] = tsdbutil.GenerateTestHistogram(int64(i*100) + int64(rand.Float64()*100))
	}
	b.Run("prometheus_encode", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			chk := chunkenc.NewHistogramChunk()
			app, err := chk.Appender()
			require.NoError(b, err)
			for i := range ts {
				_, _, _, err := app.AppendHistogram(nil, ts[i], hs[i], true)
				require.NoError(b, err)
			}
			b.ReportMetric(float64(len(chk.Bytes())), "compressed_bytes")
		}
	})
	b.Run("alp_encode", func(b *testing.B) {
		for b.Loop() {
			b.ReportMetric(float64(encodeHistograms(hs, numSamples)), "compressed_bytes")
		}
	})

	//var (
	//	hints                                  []byte
	//	Schema                                 []int32
	//	ZeroThreshold                          []float64
	//	ZeroCount                              []uint64
	//	Count                                  []uint64
	//	Sum                                    []float64
	//	PositiveSpanOffset, NegativeSpanOffset []int32
	//	PositiveSpanLength, NegativeSpanLength []uint64
	//	PositiveBuckets, NegativeBuckets       []int64
	//)

	//fmt.Println(hs[0])
}

// trimPadding removes the padding bytes from the end of the buffer.
// Each encoding adds 32 bytes of padding, we'll add it back once at the end.
func trimPadding(buf []byte, paddingSize int) []byte {
	if len(buf) >= paddingSize {
		return buf[:len(buf)-paddingSize]
	}
	return buf
}

func encodeHistograms(hs []*histogram.Histogram, numSamples int) int {
	lengths := make([]int32, 0, 15)

	// Pre-allocate a single shared buffer for all encodings
	// We encode everything into this buffer and remove intermediate padding
	sharedBuf := make([]byte, 0, 8192)

	hints := make([]byte, 0, numSamples)
	for _, h := range hs {
		hints = append(hints, byte(h.CounterResetHint))
	}
	enc := rle.Encoding{}
	// RLE doesn't support appending yet, so encode separately
	buf, _ := enc.EncodeBoolean(nil, hints)
	sharedBuf = append(sharedBuf, buf...)
	lengths = append(lengths, int32(len(buf)))

	var startLen int
	schema := make([]int32, 0, numSamples)
	for _, h := range hs {
		schema = append(schema, h.Schema)
	}
	startLen = len(sharedBuf)
	sharedBuf = dod.EncodeInt32(sharedBuf, schema)
	sharedBuf = trimPadding(sharedBuf, bitpack.PaddingInt32) // Remove padding
	lengths = append(lengths, int32(len(sharedBuf)-startLen))

	zeroThresholds := make([]float64, 0, numSamples)
	for _, h := range hs {
		zeroThresholds = append(zeroThresholds, h.ZeroThreshold)
	}
	// ALP doesn't support appending yet, so encode separately
	alpBuf := alp.Compress(zeroThresholds)
	sharedBuf = append(sharedBuf, alpBuf...)
	lengths = append(lengths, int32(len(alpBuf)))

	zeroCounts := make([]uint64, 0, numSamples)
	for _, h := range hs {
		zeroCounts = append(zeroCounts, h.ZeroCount)
	}
	startLen = len(sharedBuf)
	sharedBuf = dod.EncodeUInt64(sharedBuf, zeroCounts)
	sharedBuf = trimPadding(sharedBuf, bitpack.PaddingInt64) // Remove padding
	lengths = append(lengths, int32(len(sharedBuf)-startLen))

	counts := make([]uint64, 0, numSamples)
	for _, h := range hs {
		counts = append(counts, h.Count)
	}
	startLen = len(sharedBuf)
	sharedBuf = dod.EncodeUInt64(sharedBuf, counts)
	sharedBuf = trimPadding(sharedBuf, bitpack.PaddingInt64) // Remove padding
	lengths = append(lengths, int32(len(sharedBuf)-startLen))

	sums := make([]float64, 0, numSamples)
	for _, h := range hs {
		sums = append(sums, h.Sum)
	}
	// ALP doesn't support appending yet, so encode separately
	alpBuf = alp.Compress(sums)
	sharedBuf = append(sharedBuf, alpBuf...)
	lengths = append(lengths, int32(len(alpBuf)))

	spanOffsets := make([]int64, 0, numSamples)
	spanLengths := make([]int64, 0, numSamples)
	spanCounts := make([]int64, 0, numSamples)
	for _, h := range hs {
		for _, s := range h.PositiveSpans {
			spanOffsets = append(spanOffsets, int64(s.Offset))
			spanLengths = append(spanLengths, int64(s.Length))
		}
		spanCounts = append(spanCounts, int64(len(h.PositiveSpans)))
	}
	startLen = len(sharedBuf)
	sharedBuf = delta.EncodeInt64(sharedBuf, spanOffsets)
	sharedBuf = trimPadding(sharedBuf, bitpack.PaddingInt64) // Remove padding
	lengths = append(lengths, int32(len(sharedBuf)-startLen))

	startLen = len(sharedBuf)
	sharedBuf = delta.EncodeInt64(sharedBuf, spanLengths)
	sharedBuf = trimPadding(sharedBuf, bitpack.PaddingInt64) // Remove padding
	lengths = append(lengths, int32(len(sharedBuf)-startLen))

	startLen = len(sharedBuf)
	sharedBuf = delta.EncodeInt64(sharedBuf, spanCounts)
	sharedBuf = trimPadding(sharedBuf, bitpack.PaddingInt64) // Remove padding
	lengths = append(lengths, int32(len(sharedBuf)-startLen))

	spanOffsets = spanOffsets[:0]
	spanLengths = spanLengths[:0]
	spanCounts = spanCounts[:0]
	for _, h := range hs {
		for _, s := range h.NegativeSpans {
			spanOffsets = append(spanOffsets, int64(s.Offset))
			spanLengths = append(spanLengths, int64(s.Length))
		}
		spanCounts = append(spanCounts, int64(len(h.PositiveSpans)))
	}
	startLen = len(sharedBuf)
	sharedBuf = delta.EncodeInt64(sharedBuf, spanOffsets)
	sharedBuf = trimPadding(sharedBuf, bitpack.PaddingInt64) // Remove padding
	lengths = append(lengths, int32(len(sharedBuf)-startLen))

	startLen = len(sharedBuf)
	sharedBuf = delta.EncodeInt64(sharedBuf, spanLengths)
	sharedBuf = trimPadding(sharedBuf, bitpack.PaddingInt64) // Remove padding
	lengths = append(lengths, int32(len(sharedBuf)-startLen))

	startLen = len(sharedBuf)
	sharedBuf = delta.EncodeInt64(sharedBuf, spanCounts)
	sharedBuf = trimPadding(sharedBuf, bitpack.PaddingInt64) // Remove padding
	lengths = append(lengths, int32(len(sharedBuf)-startLen))

	// Reorder positive buckets: first buckets together, then all deltas
	firstBuckets := make([]int64, 0, numSamples)
	deltas := make([]int64, 0, numSamples)
	for _, h := range hs {
		if len(h.PositiveBuckets) > 0 {
			firstBuckets = append(firstBuckets, h.PositiveBuckets[0])
			deltas = append(deltas, h.PositiveBuckets[1:]...)
		}
	}
	startLen = len(sharedBuf)
	sharedBuf = dod.EncodeInt64(sharedBuf, firstBuckets)
	sharedBuf = trimPadding(sharedBuf, bitpack.PaddingInt64) // Remove padding
	lengths = append(lengths, int32(len(sharedBuf)-startLen))

	startLen = len(sharedBuf)
	sharedBuf = delta.EncodeInt64(sharedBuf, deltas)
	sharedBuf = trimPadding(sharedBuf, bitpack.PaddingInt64) // Remove padding
	lengths = append(lengths, int32(len(sharedBuf)-startLen))

	// Reorder negative buckets: first buckets together, then all deltas
	firstBuckets = firstBuckets[:0]
	deltas = deltas[:0]
	for _, h := range hs {
		if len(h.NegativeBuckets) > 0 {
			firstBuckets = append(firstBuckets, h.NegativeBuckets[0])
			deltas = append(deltas, h.NegativeBuckets[1:]...)
		}
	}
	startLen = len(sharedBuf)
	sharedBuf = dod.EncodeInt64(sharedBuf, firstBuckets)
	sharedBuf = trimPadding(sharedBuf, bitpack.PaddingInt64) // Remove padding
	lengths = append(lengths, int32(len(sharedBuf)-startLen))

	startLen = len(sharedBuf)
	sharedBuf = delta.EncodeInt64(sharedBuf, deltas)
	// Keep padding on the last encoding!
	lengths = append(lengths, int32(len(sharedBuf)-startLen))

	// We removed intermediate padding and kept only the final 32-byte padding
	// Total size is now much smaller than before
	return len(sharedBuf)
	//fmt.Println(lengths, len(dod.EncodeInt32(nil, lengths)))
}
