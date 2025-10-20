# ALP-Go

Adaptive Lossless floating-Point Compression (ALP) implementation in Go.

## Overview

ALP-Go is a high-performance, lossless compression library for floating-point data (float64). It uses adaptive techniques to find optimal encoding schemes for different data patterns, achieving excellent compression ratios while maintaining perfect reconstruction of original values.

## Features

- **Lossless Compression**: Perfect reconstruction of all float64 values
- **Adaptive Algorithm**: Automatically finds optimal encoding parameters for your data
- **High Performance**: Compression throughput of 300-700 MB/s on Apple M3
- **Multiple Encoding Schemes**:
  - Constant value optimization (99%+ compression for repeated values)
  - Frame-of-reference encoding
  - Bit-packing with variable bit widths
  - Adaptive exponent selection
- **Simple API**: Easy-to-use compress/decompress functions

## Installation

```bash
go get alp-go
```

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    alp "alp-go"
)

func main() {
    // Original data
    data := []float64{1.1, 2.2, 3.3, 4.4, 5.5}

    // Compress
    compressed := alp.Compress(data)

    // Decompress
    decompressed := alp.Decompress(compressed)

    fmt.Printf("Original: %v\n", data)
    fmt.Printf("Compressed size: %d bytes\n", len(compressed))
    fmt.Printf("Decompressed: %v\n", decompressed)
}
```

### With Statistics

```go
data := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
compressed, stats, err := alp.CompressWithStats(data)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Original size: %d bytes\n", stats.OriginalSize)
fmt.Printf("Compressed size: %d bytes\n", stats.CompressedSize)
fmt.Printf("Compression ratio: %.2f%%\n", stats.Ratio*100)
```

## Performance

Benchmarks on Apple M3:

```
BenchmarkCompressSmall-8              485.5 ns/op    216 B/op    5 allocs/op
BenchmarkCompressMedium-8            2525 ns/op    2176 B/op    5 allocs/op
BenchmarkCompressLarge-8           107114 ns/op  200707 B/op    4 allocs/op
BenchmarkCompressConstant-8          274.7 ns/op     32 B/op    1 allocs/op

Compression Throughput:
- 100 values:    302 MB/s
- 1,000 values:  368 MB/s
- 10,000 values: 721 MB/s
- 100,000 values: 707 MB/s
```

## Compression Ratios

Typical compression ratios for different data patterns:

| Data Pattern | Compression Ratio | Space Saved |
|-------------|------------------|-------------|
| Constant values | 0.3% | 99.7% |
| Integer-like floats | 35% | 65% |
| Decimal values (1 digit) | 43% | 57% |
| Small decimals | 41% | 59% |
| Random sensor data | 16% | 84% |
| Time series data | 19% | 81% |
| Large sequential dataset | 22% | 78% |

## How It Works

ALP compression works through several stages:

1. **Exponent Selection**: Analyzes the data to find the optimal power-of-10 multiplier that converts all floats to integers losslessly

2. **Integer Conversion**: Multiplies all values by the selected factor and converts to int64

3. **Frame-of-Reference Encoding**: Subtracts the minimum value to reduce the range

4. **Bit-Packing**: Packs the resulting values using the minimum number of bits required

5. **Metadata Storage**: Stores encoding parameters for reconstruction

## Running Examples

```bash
cd examples
go run main.go
```

## Running Tests

```bash
go test -v
```

## Running Benchmarks

```bash
go test -bench=. -benchmem
```

## API Reference

### Core Functions

#### `Compress(data []float64) []byte`
Compresses a slice of float64 values.

#### `Decompress(data []byte) []float64`
Decompresses ALP-encoded data.

#### `CompressWithStats(data []float64) ([]byte, *CompressionStats, error)`
Compresses data and returns compression statistics.

### Types

#### `CompressionStats`
```go
type CompressionStats struct {
    OriginalSize   int64
    CompressedSize int64
    Ratio          float64
}
```

#### `ALPCompressor`
```go
type ALPCompressor struct {
    // ...
}

func NewALPCompressor() *ALPCompressor
func (ac *ALPCompressor) Compress(data []float64) []byte
func (ac *ALPCompressor) Decompress(data []byte) []float64
```

## Limitations

- Currently supports only float64 (not float32)
- Works best with data that can be represented as scaled integers
- Exponent range limited to -10 to +10 (10^-10 to 10^10)
- Not suitable for random/noisy floating-point data with full precision requirements

## Use Cases

ALP compression is ideal for:

- **Time series data**: Sensor readings, metrics, telemetry
- **Financial data**: Prices, amounts, rates
- **Scientific data**: Measurements with limited precision
- **Database columns**: Columnar storage of floating-point values
- **Data analytics**: Reducing storage and transfer costs

## License

MIT License

## References

Based on the ALP (Adaptive Lossless floating-Point Compression) algorithm.

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.
