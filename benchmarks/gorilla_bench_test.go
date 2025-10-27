package benchmarks

import (
	"math/rand/v2"
	"testing"
	"time"

	"github.com/prometheus/prometheus/tsdb/chunkenc"
	"github.com/stretchr/testify/require"

	"github.com/fpetkovski/tscodec-go/alp"
	"github.com/fpetkovski/tscodec-go/dod"
)

func BenchmarkFloats(b *testing.B) {
	const numSamples = 120
	var (
		t = time.Now().UnixMilli()
		v = float64(10_000)

		tsArray [numSamples]int64
		vsArray [numSamples]float64
	)

	ts := tsArray[:numSamples]
	vs := vsArray[:numSamples]
	for i := range numSamples {
		ts[i] = t + int64(15_000*i) + int64(rand.Float64()*100)
		vs[i] = v + float64(i+1000) + rand.Float64()*100
	}

	b.Run("gorilla", func(b *testing.B) {
		b.Run("encode", func(b *testing.B) {
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
		b.Run("decode", func(b *testing.B) {
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
	})

	b.Run("alp", func(b *testing.B) {
		b.Run("encode", func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				tsc := dod.EncodeInt64(nil, ts)
				fsc := alp.Encode(nil, vs)

				b.ReportMetric(float64(len(tsc)+len(fsc)), "compressed_bytes")
			}
		})
		b.Run("alp-decode", func(b *testing.B) {
			b.ReportAllocs()

			tsc := dod.EncodeInt64(nil, ts)
			fsc := alp.Encode(nil, vs)
			b.SetBytes(int64(len(fsc) + len(tsc)))

			var (
				ints   dod.Int64Block
				floats [numSamples]float64
			)
			for b.Loop() {
				_ = dod.DecodeInt64(ints[:], tsc)
				_ = alp.Decode(floats[:], fsc)
			}
		})
	})
}
