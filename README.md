# tscodec-go

High-performance timeseries compression codecs for Go, featuring adaptive algorithms optimized for modern CPU architectures.

## Features

This library implements several state-of-the-art compression algorithms for timeseries data:

- **ALP (Adaptive Lossless floating-Point)** - Lossless compression for float64 values using adaptive scaling and bit-packing
- **Delta Encoding** - First-order delta encoding for int32/int64 values
- **Delta-of-Delta (DoD)** - Second-order delta encoding for highly correlated timeseries
- **Bitpacking** - Low-level bit manipulation with architecture-specific optimizations (amd64, arm64)

## Installation

```bash
go get github.com/fpetkovski/tscodec-go
```

## Quick Start

### ALP Compression (Float64)

```go
package main

import (
    "fmt"
    "github.com/fpetkovski/tscodec-go/alp"
)

func main() {
    // Original data
    data := []float64{1.1, 2.2, 3.3, 4.4, 5.5}

    // Compress
    compressed := alp.Compress(data)

    // Decompress
    decompressed := make([]float64, len(data))
    alp.Decompress(decompressed, compressed)

    // Calculate compression ratio
    ratio := alp.CompressionRatio(len(data), len(compressed))
    fmt.Printf("Compression ratio: %.2f%%\n", ratio*100)
}
```

### Delta Encoding (Int64)

```go
package main

import (
    "github.com/fpetkovski/tscodec-go/delta"
)

func main() {
    // Original data
    data := []int64{1000, 1001, 1002, 1003, 1004}

    // Compress
    compressed := make([]byte, 0)
    compressed = delta.EncodeInt64(compressed, data)

    // Decompress
    decompressed := make([]int64, len(data))
    delta.DecodeInt64(decompressed, compressed)
}
```

## Algorithms

### ALP (Adaptive Lossless floating-Point)

ALP achieves high compression ratios for float64 data through:

1. **Adaptive scaling** - Finds optimal exponent for converting floats to integers losslessly
2. **Frame-of-reference encoding** - Reduces value range by subtracting minimum
3. **Bit-packing** - Packs values using minimal bit width

**Best for:**
- Sensor data with limited precision
- Price data with 2-4 decimal places
- Sequential patterns
- Constant or near-constant values (achieves 99.7% compression)

**See [ALGORITHM.md](alp/ALGORITHM.md) for detailed explanation.**

### Delta Encoding

Encodes differences between consecutive values instead of absolute values.

**Best for:**
- Monotonically increasing sequences (timestamps, counters)
- Values with small differences between consecutive elements

### Delta-of-Delta (DoD)

Applies delta encoding twice, encoding the difference of differences.

**Best for:**
- Highly regular timeseries (e.g., evenly-spaced timestamps)
- Data with constant or near-constant rate of change

## Performance

The library includes architecture-specific optimizations:

- **AMD64**: SIMD-optimized bit unpacking
- **ARM64**: NEON-optimized bit unpacking
- **Pure Go**: Portable fallback implementation

Benchmarks show compression and decompression throughput of several GB/s on modern CPUs.

## Module Structure

```
tscodec-go/
├── alp/           # ALP compression for float64
├── delta/         # Delta encoding for int32/int64
├── dod/           # Delta-of-Delta encoding
├── bitpack/       # Low-level bit-packing utilities
├── unsafecast/    # Type conversion utilities
└── benchmarks/    # Performance benchmarks
```

## API Documentation

### ALP Package

```go
// Compress float64 slice
func Compress(data []float64) []byte

// Decompress to float64 slice
func Decompress(dst []float64, data []byte) int

// Calculate compression ratio
func CompressionRatio(originalCount int, compressedSize int) float64
```

### Delta Package

```go
// Encode int64 slice with delta encoding
func EncodeInt64(dst []byte, src []int64) []byte

// Decode int64 slice
func DecodeInt64(dst []int64, src []byte) uint16

// Int32 variants also available
func EncodeInt32(dst []byte, src []int32) []byte
func DecodeInt32(dst []int32, src []byte) uint16
```

## Requirements

- Go 1.25.1 or later

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## References

- ALP paper: [Adaptive Lossless floating-Point Compression](https://www.vldb.org/pvldb/vol16/p2953-afroozeh.pdf)
- Delta encoding: Standard technique for timeseries compression
- Gorilla/Chimp: Related work on timeseries compression

## Acknowledgments

This implementation includes optimized bitpacking routines with architecture-specific SIMD optimizations.
