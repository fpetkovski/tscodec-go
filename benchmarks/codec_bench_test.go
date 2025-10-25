package benchmarks

import (
	pqdelta "github.com/parquet-go/parquet-go/encoding/delta"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/fpetkovski/tscodec-go/alp"
	"github.com/fpetkovski/tscodec-go/dod"

	"github.com/prometheus/prometheus/tsdb/chunkenc"
	"github.com/stretchr/testify/require"
)

func BenchmarkFloatEncoding(b *testing.B) {
	const numSamples = 120
	t := time.Now().UnixMilli()
	v := float64(10_000)
	var tsArray [maxSamples]int64
	var vsArray [maxSamples]float64
	ts := tsArray[:numSamples]
	vs := vsArray[:numSamples]
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
			ints        dod.Int64Block
			floatsArray [maxSamples]float64
		)
		floats := floatsArray[:numSamples]
		for b.Loop() {
			dod.DecodeInt64(ints[:], tsc)
			_ = alp.Decompress(floats, fsc)
		}
	})
}
