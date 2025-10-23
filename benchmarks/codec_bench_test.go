package benchmarks

import (
	"math/rand/v2"
	"testing"
	"time"

	"alp-go/alp"
	"alp-go/dod"

	"github.com/prometheus/prometheus/tsdb/chunkenc"
	"github.com/stretchr/testify/require"
)

func BenchmarkEncoding(b *testing.B) {
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
			tsc := dod.Encode(nil, ts)
			fsc := alp.Compress(vs)

			b.ReportMetric(float64(len(tsc)+len(fsc)), "compressed_bytes")
		}
	})
	b.Run("alp-decode", func(b *testing.B) {
		b.ReportAllocs()

		tsc := dod.Encode(nil, ts)
		fsc := alp.Compress(vs)

		var (
			ints   dod.Block
			floats = make([]float64, numSamples)
		)
		for b.Loop() {
			dod.Decode(ints[:], tsc)
			_ = alp.Decompress(floats, fsc)
		}
	})
}
