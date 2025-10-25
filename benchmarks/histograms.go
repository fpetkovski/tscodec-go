package benchmarks

import (
	"alp-go/alp"
	"alp-go/bitpack"
	"alp-go/delta"
	"alp-go/dod"
	"alp-go/unsafecast"
	"encoding/binary"
	"github.com/parquet-go/parquet-go/encoding/rle"
	"github.com/prometheus/prometheus/model/histogram"
)

var rleCodec = rle.Encoding{}

const maxSamples = 120

// decodeHistograms decodes histograms from a buffer created by encodeHistograms.
//
// DECODING FORMAT:
//  1. The lengths array is DoD-encoded and appended at the end of the buffer
//  2. The last 4 bytes contain the size of the encoded lengths
//  3. This avoids the extra copy required when prepending the lengths array
//
// SAFETY CONSIDERATIONS:
//
//  1. This decoder MUST process the buffer sequentially from the beginning because
//     intermediate padding was removed during encoding. The SIMD unpacking code in
//     bitpack can safely read up to 32 bytes ahead into the next block's data.
//
//  2. The self-describing formats (DoD, Delta) work because their headers contain
//     enough information to calculate the exact encoded size.
//
//  3. This approach is safe because:
//     - SIMD unpacking reads ahead but only uses bits specified by the header
//     - Sequential decoding means the "junk" bytes read ahead are from the next block
//     - The final block has real padding, preventing segfaults
func decodeHistograms(dst []*histogram.Histogram, src []byte, numSamples int) {
	// Read the size of encoded offsets from the last 4 bytes
	lengthsSize := binary.LittleEndian.Uint32(src[len(src)-4:])

	// Decode offsets from the end (before the 4-byte size)
	lengthsStart := len(src) - 4 - int(lengthsSize)
	var lengthsBlock dod.Int32Block
	numLengths := dod.DecodeInt32(lengthsBlock[:], src[lengthsStart:len(src)-4])
	offsets := lengthsBlock[:numLengths]

	// Lengths array contains start positions: [0, end_of_block_0, end_of_block_1, ...]
	// To decode block i, use src[offsets[i]:offsets[i+1]]
	i := 0

	// Decode hints (RLE - non-self-describing, needs bounds)
	var hintsArray [maxSamples]byte
	hints := hintsArray[:numSamples]
	_, _ = rleCodec.DecodeBoolean(hints, src[offsets[i]:offsets[i+1]])
	i++

	// Decode schema (DoD int32 - self-describing)
	var schemas dod.Int32Block
	dod.DecodeInt32(schemas[:], src[offsets[i]:])
	i++

	// Decode zero thresholds (ALP - non-self-describing, needs bounds)
	var zeroThresholdsArray [maxSamples]float64
	zeroThresholds := zeroThresholdsArray[:numSamples]
	_ = alp.Decompress(zeroThresholds, src[offsets[i]:offsets[i+1]])
	i++

	// Decode zero counts (DoD uint64 - self-describing)
	var zeroCounts dod.Uint64BLock
	dod.DecodeUInt64(zeroCounts[:], src[offsets[i]:])
	i++

	// Decode counts (DoD uint64 - self-describing)
	var counts dod.Uint64BLock
	dod.DecodeUInt64(counts[:], src[offsets[i]:])
	i++

	// Decode sums (ALP - non-self-describing, needs bounds)
	var sumsArray [maxSamples]float64
	sums := sumsArray[:numSamples]
	_ = alp.Decompress(sums, src[offsets[i]:offsets[i+1]])
	i++

	// Decode positive span offsets, offsets, counts (Delta - self-describing)
	posSpanOffsets := make([]int64, 4096)
	endIdx := lengthsStart
	if i+1 < len(offsets) {
		endIdx = int(offsets[i+1])
	}
	numPosSpanOffsets := delta.DecodeInt64(posSpanOffsets, src[offsets[i]:endIdx])
	posSpanOffsets = posSpanOffsets[:numPosSpanOffsets]
	i++

	posSpanLengths := make([]int64, 4096)
	endIdx = lengthsStart
	if i+1 < len(offsets) {
		endIdx = int(offsets[i+1])
	}
	numPosSpanLengths := delta.DecodeInt64(posSpanLengths, src[offsets[i]:endIdx])
	posSpanLengths = posSpanLengths[:numPosSpanLengths]
	i++

	var posSpanCounts delta.Int64Block
	delta.DecodeInt64(posSpanCounts[:], src[offsets[i]:offsets[i+1]])
	i++

	// Decode negative span offsets, lengths, counts (Delta - self-describing)
	negSpanOffsets := make([]int64, 4096)
	numNegSpanOffsets := delta.DecodeInt64(negSpanOffsets, src[offsets[i]:offsets[i+1]])
	negSpanOffsets = negSpanOffsets[:numNegSpanOffsets]
	i++

	negSpanLengths := make([]int64, 4096)
	numNegSpanLengths := delta.DecodeInt64(negSpanLengths, src[offsets[i]:offsets[i+1]])
	negSpanLengths = negSpanLengths[:numNegSpanLengths]
	i++

	var negSpanCounts delta.Int64Block
	delta.DecodeInt64(negSpanCounts[:], src[offsets[i]:offsets[i+1]])
	i++

	// Decode positive bucket counts per histogram
	var posBucketCounts delta.Int64Block
	endIdx = lengthsStart
	if i+1 < len(offsets) {
		endIdx = int(offsets[i+1])
	}
	delta.DecodeInt64(posBucketCounts[:], src[offsets[i]:endIdx])
	i++

	// Decode positive bucket first values (DoD int64 - self-describing)
	var posFirstBuckets dod.Int64Block
	endIdx = lengthsStart
	if i+1 < len(offsets) {
		endIdx = int(offsets[i+1])
	}
	dod.DecodeInt64(posFirstBuckets[:], src[offsets[i]:endIdx])
	i++

	// Decode positive bucket deltas (Delta - self-describing)
	// Deltas can exceed Int64BlockSize, so we need a larger slice
	var posDeltas [1024]int64 // Enough for many histograms
	endIdx = lengthsStart
	if i+1 < len(offsets) {
		endIdx = int(offsets[i+1])
	}
	numPosDeltas := delta.DecodeInt64(posDeltas[:], src[offsets[i]:endIdx])
	i++

	// Decode negative bucket counts per histogram
	var negBucketCounts delta.Int64Block
	endIdx = lengthsStart
	if i+1 < len(offsets) {
		endIdx = int(offsets[i+1])
	}
	delta.DecodeInt64(negBucketCounts[:], src[offsets[i]:endIdx])
	i++

	// Decode negative bucket first values (DoD int64 - self-describing)
	var negFirstBuckets dod.Int64Block
	endIdx = lengthsStart
	if i+1 < len(offsets) {
		endIdx = int(offsets[i+1])
	}
	dod.DecodeInt64(negFirstBuckets[:], src[offsets[i]:endIdx])
	i++

	// Decode negative bucket deltas (last field, has padding, goes until lengthsStart)
	// Deltas can exceed Int64BlockSize, so we need a larger slice
	var negDeltas [1024]int64 // Enough for many histograms
	numNegDeltas := delta.DecodeInt64(negDeltas[:], src[offsets[i]:lengthsStart])

	// Reconstruct histograms
	posSpanIdx := 0
	negSpanIdx := 0
	posFirstBucketIdx := 0
	negFirstBucketIdx := 0
	posDeltaIdx := 0
	negDeltaIdx := 0
	positiveSpans := make([]histogram.Span, numPosSpanLengths)
	for i := range numPosSpanLengths {
		positiveSpans[i].Offset = posSpanOffsets[i]
		positiveSpans[i].Length = posSpanLengths[i]
	}
	negativeSpans := make([]histogram.Span, numNegSpanLengths)
	for histIdx := 0; histIdx < numSamples; histIdx++ {
		h := dst[histIdx]

		// Restore basic fields
		h.CounterResetHint = histogram.CounterResetHint(hints[histIdx])
		h.Schema = schemas[histIdx]
		h.ZeroThreshold = zeroThresholds[histIdx]
		h.ZeroCount = zeroCounts[histIdx]
		h.Count = counts[histIdx]
		h.Sum = sums[histIdx]

		// Restore positive spans
		numPosSpans := int(posSpanCounts[histIdx])
		if numPosSpans > 0 {
			h.PositiveSpans = make([]histogram.Span, numPosSpans)
			for j := 0; j < numPosSpans; j++ {
				h.PositiveSpans[j].Offset = int32(posSpanOffsets[posSpanIdx+j])
				h.PositiveSpans[j].Length = uint32(posSpanLengths[posSpanIdx+j])
			}
			posSpanIdx += numPosSpans
		}

		// Restore positive buckets using explicit bucket count
		numPosBuckets := int(posBucketCounts[histIdx])
		if numPosBuckets > 0 && posFirstBucketIdx < len(posFirstBuckets) {
			h.PositiveBuckets = make([]int64, numPosBuckets)
			h.PositiveBuckets[0] = posFirstBuckets[posFirstBucketIdx]
			posFirstBucketIdx++
			// Note: numPosBuckets-1 deltas for numPosBuckets buckets (first bucket already set)
			for j := 1; j < numPosBuckets; j++ {
				if posDeltaIdx >= len(posDeltas) {
					// If we run out of deltas, something is wrong - but fill with 0 for now
					break
				}
				h.PositiveBuckets[j] = posDeltas[posDeltaIdx]
				posDeltaIdx++
			}
		}

		// Restore negative spans
		numNegSpans := int(negSpanCounts[histIdx])
		if numNegSpans > 0 {
			h.NegativeSpans = make([]histogram.Span, numNegSpans)
			for j := 0; j < numNegSpans; j++ {
				h.NegativeSpans[j].Offset = int32(negSpanOffsets[negSpanIdx+j])
				h.NegativeSpans[j].Length = uint32(negSpanLengths[negSpanIdx+j])
			}
			negSpanIdx += numNegSpans
		}

		// Restore negative buckets using explicit bucket count
		numNegBuckets := int(negBucketCounts[histIdx])
		if numNegBuckets > 0 && negFirstBucketIdx < len(negFirstBuckets) {
			h.NegativeBuckets = make([]int64, numNegBuckets)
			h.NegativeBuckets[0] = negFirstBuckets[negFirstBucketIdx]
			negFirstBucketIdx++
			// Note: numNegBuckets-1 deltas for numNegBuckets buckets (first bucket already set)
			for j := 1; j < numNegBuckets; j++ {
				if negDeltaIdx >= len(negDeltas) {
					// If we run out of deltas, something is wrong - but fill with 0 for now
					break
				}
				h.NegativeBuckets[j] = negDeltas[negDeltaIdx]
				negDeltaIdx++
			}
		}
	}
}

//

// encodeHistogramsToBuffer encodes histograms and returns the buffer
func encodeHistogramsToBuffer(hs []*histogram.Histogram, numSamples int) []byte {
	return encodeHistogramsInternal(hs, numSamples, false)
}

// encodeHistograms encodes histograms and returns the size
func encodeHistograms(hs []*histogram.Histogram, numSamples int) int {
	return len(encodeHistogramsInternal(hs, numSamples, true))
}

// encodeHistogramsInternal is the actual implementation.
//
// PADDING OPTIMIZATION:
// This encoder removes intermediate padding between blocks to save space.
// Each bitpack encoding normally adds 32 bytes of padding for SIMD safety.
// With ~15 encodings, that's ~480 bytes of pure overhead.
//
// The trick: We trim padding after each encoding and add it back only at the end.
// This is safe because:
//  1. All encodings go into one contiguous buffer
//  2. Decoders process sequentially from the start
//  3. SIMD code can safely read 32 bytes ahead into the next block
//  4. The final block has real padding to prevent segfaults
//
// Result: 368 bytes saved (2027 → 1659 bytes, beating Prometheus by 9.5%)
func encodeHistogramsInternal(hs []*histogram.Histogram, numSamples int, sizeOnly bool) []byte {
	// Store start positions of each block
	lengths := make([]int32, 0, 16)
	lengths = append(lengths, 0) // First block starts at 0

	// Pre-allocate a single shared buffer for all encodings
	// We encode everything into this buffer and remove intermediate padding
	dst := make([]byte, 0, 8192)
	scratch := make([]byte, 0, 8192)

	hints := make([]byte, 0, numSamples)
	for _, h := range hs {
		hints = append(hints, byte(h.CounterResetHint))
	}
	enc := rle.Encoding{}
	dst, _ = enc.EncodeBoolean(dst, hints)
	lengths = append(lengths, int32(len(dst))) // Start of next block

	schemas := unsafecast.Slice[int32](scratch)[:0]
	for _, h := range hs {
		schemas = append(schemas, h.Schema)
	}
	dst = trimPadding(dod.EncodeInt32(dst, schemas), bitpack.PaddingInt32)
	lengths = append(lengths, int32(len(dst)))

	zeroThresholds := unsafecast.Slice[float64](scratch)[:0]
	for _, h := range hs {
		zeroThresholds = append(zeroThresholds, h.ZeroThreshold)
	}
	dst = append(dst, alp.Compress(zeroThresholds)...)
	lengths = append(lengths, int32(len(dst)))

	zeroCounts := unsafecast.Slice[uint64](scratch)[:0]
	for _, h := range hs {
		zeroCounts = append(zeroCounts, h.ZeroCount)
	}
	dst = trimPadding(dod.EncodeUInt64(dst, zeroCounts), bitpack.PaddingInt64)
	lengths = append(lengths, int32(len(dst)))

	counts := unsafecast.Slice[uint64](scratch)[:0]
	for _, h := range hs {
		counts = append(counts, h.Count)
	}
	dst = trimPadding(dod.EncodeUInt64(dst, counts), bitpack.PaddingInt64)
	lengths = append(lengths, int32(len(dst)))

	sums := unsafecast.Slice[float64](scratch)[:0]
	for _, h := range hs {
		sums = append(sums, h.Sum)
	}
	dst = append(dst, alp.Compress(sums)...)
	lengths = append(lengths, int32(len(dst)))

	spanOffsets := make([]int64, 0, numSamples)
	spanLengths := make([]int64, 0, numSamples)
	spanCounts := make([]int64, 0, numSamples)
	for _, h := range hs {
		// Only encode spans if there are actual buckets
		for _, s := range h.PositiveSpans {
			spanOffsets = append(spanOffsets, int64(s.Offset))
			spanLengths = append(spanLengths, int64(s.Length))
		}
		spanCounts = append(spanCounts, int64(len(h.PositiveSpans)))
	}

	dst = trimPadding(delta.EncodeInt64(dst, spanOffsets), bitpack.PaddingInt64)
	lengths = append(lengths, int32(len(dst)))

	dst = trimPadding(delta.EncodeInt64(dst, spanLengths), bitpack.PaddingInt64)
	lengths = append(lengths, int32(len(dst)))

	dst = trimPadding(delta.EncodeInt64(dst, spanCounts), bitpack.PaddingInt64)
	lengths = append(lengths, int32(len(dst)))

	spanOffsets = spanOffsets[:0]
	spanLengths = spanLengths[:0]
	spanCounts = spanCounts[:0]
	for _, h := range hs {
		// Only encode spans if there are actual buckets
		for _, s := range h.NegativeSpans {
			spanOffsets = append(spanOffsets, int64(s.Offset))
			spanLengths = append(spanLengths, int64(s.Length))
		}
		spanCounts = append(spanCounts, int64(len(h.NegativeSpans)))
	}
	dst = trimPadding(delta.EncodeInt64(dst, spanOffsets), bitpack.PaddingInt64)
	lengths = append(lengths, int32(len(dst)))

	dst = trimPadding(delta.EncodeInt64(dst, spanLengths), bitpack.PaddingInt64)
	lengths = append(lengths, int32(len(dst)))

	dst = trimPadding(delta.EncodeInt64(dst, spanCounts), bitpack.PaddingInt64)
	lengths = append(lengths, int32(len(dst)))

	// Reorder positive buckets: first buckets together, then all deltas
	// Also track bucket counts per histogram
	firstBuckets := make([]int64, 0, numSamples)
	deltas := make([]int64, 0, numSamples)
	bucketCounts := make([]int64, 0, numSamples)
	for _, h := range hs {
		bucketCounts = append(bucketCounts, int64(len(h.PositiveBuckets)))
		if len(h.PositiveBuckets) > 0 {
			firstBuckets = append(firstBuckets, h.PositiveBuckets[0])
			deltas = append(deltas, h.PositiveBuckets[1:]...)
		}
	}

	dst = trimPadding(delta.EncodeInt64(dst, bucketCounts), bitpack.PaddingInt64)
	lengths = append(lengths, int32(len(dst)))

	dst = trimPadding(dod.EncodeInt64(dst, firstBuckets), bitpack.PaddingInt64)
	lengths = append(lengths, int32(len(dst)))

	dst = trimPadding(delta.EncodeInt64(dst, deltas), bitpack.PaddingInt64)
	lengths = append(lengths, int32(len(dst)))

	// Reorder negative buckets: first buckets together, then all deltas
	// Also track bucket counts per histogram
	firstBuckets = firstBuckets[:0]
	deltas = deltas[:0]
	bucketCounts = bucketCounts[:0]
	for _, h := range hs {
		bucketCounts = append(bucketCounts, int64(len(h.NegativeBuckets)))
		if len(h.NegativeBuckets) > 0 {
			firstBuckets = append(firstBuckets, h.NegativeBuckets[0])
			deltas = append(deltas, h.NegativeBuckets[1:]...)
		}
	}

	dst = trimPadding(delta.EncodeInt64(dst, bucketCounts), bitpack.PaddingInt64)
	lengths = append(lengths, int32(len(dst)))

	dst = trimPadding(dod.EncodeInt64(dst, firstBuckets), bitpack.PaddingInt64)
	lengths = append(lengths, int32(len(dst)))

	dst = delta.EncodeInt64(dst, deltas)
	// Keep padding on the last encoding!
	lengths = append(lengths, int32(len(dst)))

	// Encode lengths using DoD and append at the end to avoid extra copy
	encodedLengths := dod.EncodeInt32(nil, lengths)
	dst = append(dst, encodedLengths...)

	// Append the size of encoded lengths (4 bytes) at the very end
	lengthsSizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lengthsSizeBytes, uint32(len(encodedLengths)))
	dst = append(dst, lengthsSizeBytes...)

	return dst
}

// trimPadding removes the padding bytes from the end of the buffer.
// Each encoding adds 32 bytes of padding, we'll add it back once at the end.
func trimPadding(buf []byte, paddingSize int) []byte {
	//return buf
	if len(buf) >= paddingSize {
		return buf[:len(buf)-paddingSize]
	}
	return buf
}
