package alp

import (
	"math/rand"
	"strconv"
	"testing"
)

// Dataset sizes to test
var benchmarkSizes = []int{
	10,      // Tiny
	100,     // Small
	1000,    // Medium
	10000,   // Large
	100000,  // Very Large
	1000000, // Huge
}

var randGen = rand.New(rand.NewSource(1000))

// BenchmarkALP is the consolidated benchmark function that contains all sub-benchmarks
func BenchmarkALP(b *testing.B) {
	dataset := make([]float64, benchmarkSizes[len(benchmarkSizes)-1])
	for i := range dataset {
		dataset[i] = randGen.Float64() * 1000
	}

	compressed := Encode(nil, dataset)
	b.Run("CompressionSpeed", func(b *testing.B) {
		for _, size := range benchmarkSizes {
			data := dataset[:size]
			b.Run(strconv.Itoa(size), func(b *testing.B) {
				b.SetBytes(int64(size * 8))
				b.ResetTimer()
				for b.Loop() {
					_ = Encode(compressed, data)
				}
			})
		}
	})

	b.Run("DecompressionSpeed", func(b *testing.B) {
		for _, size := range benchmarkSizes {
			compressed := Encode(compressed, dataset[:size])
			decompressed := make([]float64, size)
			b.Run(strconv.Itoa(size), func(b *testing.B) {
				b.SetBytes(int64(size * 8))
				b.ResetTimer()
				for b.Loop() {
					_ = Decode(decompressed, compressed)
				}
			})
		}
	})

	b.Run("ByPattern", func(b *testing.B) {
		size := 1000

		patterns := map[string]func() []float64{
			"Integers": func() []float64 {
				data := make([]float64, size)
				for i := range data {
					data[i] = float64(i)
				}
				return data
			},
			"OneDecimal": func() []float64 {
				data := make([]float64, size)
				for i := range data {
					data[i] = float64(i) * 0.1
				}
				return data
			},
			"TwoDecimals": func() []float64 {
				data := make([]float64, size)
				for i := range data {
					data[i] = float64(i) * 0.01
				}
				return data
			},
			"Constant": func() []float64 {
				data := make([]float64, size)
				for i := range data {
					data[i] = 42.5
				}
				return data
			},
			"RandomSensor": func() []float64 {
				data := make([]float64, size)
				for i := range data {
					data[i] = 20.0 + randGen.Float64()*10.0
				}
				return data
			},
		}

		for patternName, generator := range patterns {
			data := generator()
			b.Run("Encode/"+patternName, func(b *testing.B) {
				b.SetBytes(int64(size * 8))
				b.ResetTimer()
				for b.Loop() {
					_ = Encode(compressed, data)
				}
			})
		}

		for patternName, generator := range patterns {
			data := generator()
			compressed = Encode(compressed, data)
			decompressed := make([]float64, size)
			b.Run("Decode/"+patternName, func(b *testing.B) {
				b.SetBytes(int64(size * 8))
				b.ResetTimer()
				for b.Loop() {
					_ = Decode(decompressed, compressed)
				}
			})
		}
	})
}
