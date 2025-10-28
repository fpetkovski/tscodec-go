package alp

import (
	"io"

	"github.com/parquet-go/bitpack"
	"github.com/parquet-go/bitpack/unsafecast"
)

type StreamEncoder struct {
	data      []float64
	blockSize int
}

func (e *StreamEncoder) Reset(blockSize int) {
	e.data = e.data[:0]
	e.blockSize = blockSize
}

func (e *StreamEncoder) Encode(src []float64) {
	e.data = append(e.data, src...)
}

func (e *StreamEncoder) Flush() []byte {
	if len(e.data) == 0 {
		return nil
	}

	// Find global encoding parameters
	exponent := findBestExponent(e.data)
	factor := powersOf10[exponent+10]

	// Convert all to integers with global exponent
	forValues := encodeToIntegers(e.data, factor)

	// Find global frame-of-reference
	minValue := forValues[0]
	for _, v := range forValues {
		minValue = min(minValue, v)
	}

	// Apply frame-of-reference and find global bit-width
	bitWidth := 0
	for i, v := range forValues {
		forValues[i] = v - minValue
		bits := CalculateBitWidth(uint64(forValues[i]))
		bitWidth = max(bitWidth, bits)
	}

	// Calculate total packed size
	totalBlocks := (len(forValues) + e.blockSize - 1) / e.blockSize
	packedSize := 0
	for i := 0; i < totalBlocks; i++ {
		blockStart := i * e.blockSize
		blockEnd := min(blockStart+e.blockSize, len(forValues))
		blockLen := blockEnd - blockStart
		packedSize += bitpack.ByteCount(uint(blockLen*bitWidth)) + bitpack.PaddingInt64
	}

	// Create output buffer: metadata + packed blocks
	output := make([]byte, MetadataSize+packedSize)

	// Write global metadata
	metadata := CompressionMetadata{
		EncodingType: EncodingALP,
		Count:        int32(len(e.data)),
		Exponent:     int8(exponent),
		BitWidth:     uint8(bitWidth),
		FrameOfRef:   minValue,
	}
	encodeMetadata(output, metadata)

	// Pack data in blocks
	offset := MetadataSize
	for i := 0; i < totalBlocks; i++ {
		blockStart := i * e.blockSize
		blockEnd := min(blockStart+e.blockSize, len(forValues))
		blockData := forValues[blockStart:blockEnd]

		blockPackedSize := bitpack.ByteCount(uint(len(blockData)*bitWidth)) + bitpack.PaddingInt64
		bitpack.PackInt64(output[offset:offset+blockPackedSize], blockData, uint(bitWidth))
		offset += blockPackedSize
	}

	return output
}

type StreamDecoder struct {
	buf              []byte
	metadata         CompressionMetadata
	blockSize        int
	decodedBuf       []float64 // Buffer for decoded block
	decodedBufOffset int       // Current read position in decoded buffer
	valuesRead       int32     // Total values read so far
}

func (d *StreamDecoder) Reset(buf []byte, blockSize int) {
	d.buf = buf
	d.blockSize = blockSize
	d.decodedBuf = nil
	d.decodedBufOffset = 0
	d.valuesRead = 0

	// Read global metadata
	if len(buf) >= MetadataSize {
		d.metadata = DecodeMetadata(buf)
		d.buf = buf[MetadataSize:]
	}
}

func (d *StreamDecoder) Decode(dst []float64) ([]float64, error) {
	if d.valuesRead >= d.metadata.Count {
		return dst[:0], io.EOF
	}

	// If we've consumed all values from current block, decode next block
	if d.decodedBufOffset >= len(d.decodedBuf) {
		// Determine block size
		remaining := d.metadata.Count - d.valuesRead
		blockSize := min(int32(d.blockSize), remaining)

		// Calculate size of packed data for this block
		packedSize := bitpack.ByteCount(uint(int(blockSize) * int(d.metadata.BitWidth)))
		if d.metadata.BitWidth != 0 {
			packedSize += bitpack.PaddingInt64
		}

		// Allocate buffer for decoded block
		if cap(d.decodedBuf) < int(blockSize) {
			d.decodedBuf = make([]float64, blockSize)
		} else {
			d.decodedBuf = d.decodedBuf[:blockSize]
		}

		// Unpack entire block
		ints := unsafecast.Slice[int64](d.decodedBuf)
		bitpack.UnpackInt64(ints, d.buf, uint(d.metadata.BitWidth))
		d.buf = d.buf[packedSize:]

		// Convert to float64
		minValue := d.metadata.FrameOfRef
		invFactor := powersOf10[(10-d.metadata.Exponent+21)%21]

		for i := range d.decodedBuf {
			d.decodedBuf[i] = float64(ints[i]+minValue) * invFactor
		}

		d.decodedBufOffset = 0
	}

	// Return a chunk from the decoded buffer
	remaining := len(d.decodedBuf) - d.decodedBufOffset
	n := min(len(dst), remaining)

	copy(dst[:n], d.decodedBuf[d.decodedBufOffset:d.decodedBufOffset+n])
	d.decodedBufOffset += n
	d.valuesRead += int32(n)

	// Check if we're done
	var err error
	if d.valuesRead >= d.metadata.Count {
		err = io.EOF
	}

	return dst[:n], err
}
