package benchmarks

import (
	pqdelta "github.com/parquet-go/parquet-go/encoding/delta"
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
	b.Run("prometheus_decode", func(b *testing.B) {
		b.ReportAllocs()
		chk := chunkenc.NewHistogramChunk()
		app, err := chk.Appender()
		require.NoError(b, err)
		for i := range ts {
			_, _, _, err := app.AppendHistogram(nil, ts[i], hs[i], true)
			require.NoError(b, err)
		}
		for b.Loop() {
			it := chk.Iterator(nil)
			h := histogram.Histogram{}
			for it.Next() != chunkenc.ValNone {
				_, _ = it.AtHistogram(&h)
			}
		}
	})
	b.Run("alp_encode", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			b.ReportMetric(float64(encodeHistograms(hs, numSamples)), "compressed_bytes")
		}
	})
	// TODO: Fix decoder to properly handle histogram span/bucket relationship
	// b.Run("alp_decode", func(b *testing.B) {
	// 	b.ReportAllocs()
	// 	encoded := encodeHistogramsToBuffer(hs, numSamples)
	// 	decoded := make([]*histogram.Histogram, numSamples)
	// 	for b.Loop() {
	// 		decodeHistograms(decoded, encoded, numSamples)
	// 	}
	// })
}
