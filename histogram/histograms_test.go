package histogram

//
//import (
//	"testing"
//
//	"github.com/prometheus/prometheus/model/histogram"
//	"github.com/prometheus/prometheus/tsdb/tsdbutil"
//	"github.com/stretchr/testify/require"
//)
//
//func TestHistogramRoundtrip(t *testing.T) {
//	const numSamples = 120
//
//	// Generate test histograms
//	hs := make([]*histogram.Histogram, numSamples)
//	for i := range numSamples {
//		hs[i] = tsdbutil.GenerateTestHistogram(int64(i*100 + i))
//	}
//
//	// Encode histograms
//	encoded := encodeHistogramsToBuffer(hs, numSamples)
//
//	// Decode histograms
//	decoded := make([]*histogram.Histogram, numSamples)
//	decodeHistograms(decoded, encoded, numSamples)
//
//	// Verify each histogram matches
//	for i := 0; i < numSamples; i++ {
//		original := hs[i]
//		restored := decoded[i]
//
//		require.NotNil(t, restored, "decoded histogram %d should not be nil", i)
//
//		// Compare basic fields
//		require.Equal(t, original.CounterResetHint, restored.CounterResetHint, "histogram %d: CounterResetHint mismatch", i)
//		require.Equal(t, original.Schema, restored.Schema, "histogram %d: Schema mismatch", i)
//		require.InDelta(t, original.ZeroThreshold, restored.ZeroThreshold, 0.0001, "histogram %d: ZeroThreshold mismatch", i)
//		require.Equal(t, original.ZeroCount, restored.ZeroCount, "histogram %d: ZeroCount mismatch", i)
//		require.Equal(t, original.Count, restored.Count, "histogram %d: Count mismatch", i)
//		require.InDelta(t, original.Sum, restored.Sum, 0.0001, "histogram %d: Sum mismatch", i)
//
//		// Compare spans
//		require.Equal(t, len(original.PositiveSpans), len(restored.PositiveSpans), "histogram %d: PositiveSpans length mismatch", i)
//		for j, span := range original.PositiveSpans {
//			require.Equal(t, span.Offset, restored.PositiveSpans[j].Offset, "histogram %d: PositiveSpans[%d].Offset mismatch", i, j)
//			require.Equal(t, span.Length, restored.PositiveSpans[j].Length, "histogram %d: PositiveSpans[%d].Length mismatch", i, j)
//		}
//
//		require.Equal(t, len(original.NegativeSpans), len(restored.NegativeSpans), "histogram %d: NegativeSpans length mismatch", i)
//		for j, span := range original.NegativeSpans {
//			require.Equal(t, span.Offset, restored.NegativeSpans[j].Offset, "histogram %d: NegativeSpans[%d].Offset mismatch", i, j)
//			require.Equal(t, span.Length, restored.NegativeSpans[j].Length, "histogram %d: NegativeSpans[%d].Length mismatch", i, j)
//		}
//
//		// Compare buckets
//		require.Equal(t, len(original.PositiveBuckets), len(restored.PositiveBuckets), "histogram %d: PositiveBuckets length mismatch", i)
//		for j, bucket := range original.PositiveBuckets {
//			require.Equal(t, bucket, restored.PositiveBuckets[j], "histogram %d: PositiveBuckets[%d] mismatch", i, j)
//		}
//
//		require.Equal(t, len(original.NegativeBuckets), len(restored.NegativeBuckets), "histogram %d: NegativeBuckets length mismatch", i)
//		for j, bucket := range original.NegativeBuckets {
//			require.Equal(t, bucket, restored.NegativeBuckets[j], "histogram %d: NegativeBuckets[%d] mismatch", i, j)
//		}
//	}
//}
//
//func TestHistogramRoundtripSmall(t *testing.T) {
//	testCases := []struct {
//		name       string
//		numSamples int
//	}{
//		{"single", 1},
//		{"two", 2},
//		{"ten", 10},
//		{"fifty", 50},
//	}
//
//	for _, tc := range testCases {
//		t.Run(tc.name, func(t *testing.T) {
//			// Generate test histograms
//			hs := make([]*histogram.Histogram, tc.numSamples)
//			for i := range tc.numSamples {
//				hs[i] = tsdbutil.GenerateTestHistogram(int64(i * 100))
//			}
//
//			// Encode histograms
//			encoded := encodeHistogramsToBuffer(hs, tc.numSamples)
//
//			// Decode histograms
//			decoded := make([]*histogram.Histogram, tc.numSamples)
//			decodeHistograms(decoded, encoded, tc.numSamples)
//
//			// Verify length matches
//			require.Len(t, decoded, tc.numSamples)
//
//			// Verify each histogram has expected structure
//			for i := 0; i < tc.numSamples; i++ {
//				require.NotNil(t, decoded[i], "decoded histogram %d should not be nil", i)
//				require.Equal(t, hs[i].Count, decoded[i].Count, "histogram %d: Count mismatch", i)
//				require.InDelta(t, hs[i].Sum, decoded[i].Sum, 0.0001, "histogram %d: Sum mismatch", i)
//			}
//		})
//	}
//}
//
//func TestHistogramEncodingSize(t *testing.T) {
//	const numSamples = 120
//
//	// Generate test histograms
//	hs := make([]*histogram.Histogram, numSamples)
//	for i := range numSamples {
//		hs[i] = tsdbutil.GenerateTestHistogram(int64(i * 100))
//	}
//
//	// Encode histograms
//	encoded := encodeHistogramsToBuffer(hs, numSamples)
//
//	// Verify encoded size is reasonable
//	t.Logf("Encoded %d histograms into %d bytes (%.1f bytes per histogram)",
//		numSamples, len(encoded), float64(len(encoded))/float64(numSamples))
//
//	// Should be under 20 bytes per histogram on average for this data
//	require.Less(t, len(encoded), numSamples*20, "encoded size should be under 20 bytes per histogram")
//}
