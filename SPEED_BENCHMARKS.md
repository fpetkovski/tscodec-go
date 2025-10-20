# ALP Speed Benchmarks

Comprehensive performance benchmarks for compression and decompression speeds.

**Test Environment**: Apple M3, 8 cores, Go 1.x

---

## Executive Summary

### Peak Performance
- **Compression**: Up to **734 MB/s** (1M values)
- **Decompression**: Up to **1,143 MB/s** (small datasets)
- **Constant data compression**: **28,218 MB/s** (99.7% compression)
- **Constant data decompression**: **10,802 MB/s**

### Sweet Spot
- **Best compression throughput**: 100K-1M values (726-734 MB/s)
- **Best decompression throughput**: 10-10K values (1,005-1,145 MB/s)
- **Most efficient**: Constant and random sensor data

---

## Compression Speed by Dataset Size

| Size | Latency | Throughput | Memory | Allocs |
|------|---------|------------|--------|--------|
| **10 values** | 486 ns | 164 MB/s | 197 B | 4 |
| **100 values** | 2.7 µs | 300 MB/s | 2.2 KB | 5 |
| **1,000 values** | 22.3 µs | 359 MB/s | 18.9 KB | 4 |
| **10,000 values** | 126 µs | **636 MB/s** | 201 KB | 4 |
| **100,000 values** | 1.1 ms | **727 MB/s** | 2.0 MB | 4 |
| **1,000,000 values** | 10.9 ms | **734 MB/s** | 21.0 MB | 4 |

### Key Findings:
- ✅ **Linear scaling** from 10K+ values
- ✅ **Stable allocations** (only 4 allocs regardless of size)
- ✅ **Peak throughput** at 1M values: **734 MB/s**

---

## Decompression Speed by Dataset Size

| Size | Latency | Throughput | Memory | Allocs |
|------|---------|------------|--------|--------|
| **10 values** | 70 ns | **1,143 MB/s** | 240 B | 3 |
| **100 values** | 699 ns | **1,145 MB/s** | 2.7 KB | 3 |
| **1,000 values** | 7.0 µs | **1,135 MB/s** | 24.6 KB | 3 |
| **10,000 values** | 79.6 µs | **1,005 MB/s** | 246 KB | 3 |
| **100,000 values** | 979 µs | 817 MB/s | 2.4 MB | 3 |
| **1,000,000 values** | 9.9 ms | 807 MB/s | 24.0 MB | 3 |

### Key Findings:
- ✅ **Decompression faster than compression** (2-3x on small datasets)
- ✅ **Peak throughput**: **1,145 MB/s** on small-medium datasets
- ✅ **Consistent 3 allocations** regardless of size
- ⚠️ Throughput decreases slightly on very large datasets (memory pressure)

---

## Speed by Data Pattern (1,000 values)

### Compression

| Pattern | Latency | Throughput | Ratio |
|---------|---------|------------|-------|
| Integers | 23.3 µs | 343 MB/s | Baseline |
| One Decimal | 21.5 µs | **372 MB/s** | 1.08x faster |
| Two Decimals | 20.0 µs | **400 MB/s** | 1.16x faster |
| **Constant** | **284 ns** | **28,218 MB/s** | **82x faster!** |
| Random Sensor | 5.9 µs | **1,356 MB/s** | **3.95x faster** |

### Decompression

| Pattern | Latency | Throughput | Ratio |
|---------|---------|------------|-------|
| Integers | 7.1 µs | 1,126 MB/s | Baseline |
| One Decimal | 7.2 µs | 1,114 MB/s | 0.99x |
| Two Decimals | 7.2 µs | 1,112 MB/s | 0.99x |
| **Constant** | **741 ns** | **10,802 MB/s** | **9.6x faster!** |
| Random Sensor | 4.6 µs | **1,749 MB/s** | **1.55x faster** |

### Key Insights:
- 🚀 **Constant data is extremely fast** (special case optimization)
- 🚀 **Random sensor data** compresses/decompresses very efficiently
- ✅ All patterns show good performance (300-1,700 MB/s)

---

## Round-Trip Performance (Compress + Decompress)

| Size | Latency | Throughput | Memory | Allocs |
|------|---------|------------|--------|--------|
| 100 values | 3.2 µs | 500 MB/s | 4.9 KB | 8 |
| 1,000 values | 29.2 µs | 549 MB/s | 43.5 KB | 7 |
| 10,000 values | 192 µs | **831 MB/s** | 446 KB | 7 |

**Note**: Throughput counts both compression and decompression bytes.

---

## Parallel Performance (1,000 values)

| Operation | Single Thread | Parallel (8 cores) | Speedup |
|-----------|--------------|-------------------|---------|
| **Compression** | 359 MB/s | **1,248 MB/s** | **3.5x** |
| **Decompression** | 1,135 MB/s | **1,811 MB/s** | **1.6x** |

### Parallel Efficiency:
- ✅ **Compression scales well** with multiple cores (3.5x on 8 cores = 44% efficiency)
- ✅ **Decompression also benefits** from parallelism
- 💡 ALP is **CPU-bound**, making it ideal for parallel processing

---

## Memory Efficiency

### Compression

| Size | Memory per Op | Bytes per Value | Allocations |
|------|--------------|----------------|-------------|
| 100 | 2.2 KB | 22 B | 5 |
| 1,000 | 18.9 KB | 19 B | 4 |
| 10,000 | 201 KB | 20 B | 4 |

### Decompression

| Size | Memory per Op | Bytes per Value | Allocations |
|------|--------------|----------------|-------------|
| 100 | 2.7 KB | 27 B | 3 |
| 1,000 | 24.6 KB | 25 B | 3 |
| 10,000 | 246 KB | 25 B | 3 |

### Key Findings:
- ✅ **~20 bytes per value** during compression
- ✅ **~25 bytes per value** during decompression
- ✅ **Very low allocation count** (3-5 allocations total)
- ✅ **No allocation growth** with dataset size

---

## Performance Characteristics

### Latency Profile
```
10 values:      486 ns    (ultra-fast)
100 values:     2.7 µs    (very fast)
1,000 values:   22 µs     (fast)
10,000 values:  126 µs    (good)
100,000 values: 1.1 ms    (acceptable)
1,000,000 values: 10.9 ms (slower but acceptable)
```

### Throughput Profile
```
Compression:
  Small (10-100):      164-300 MB/s
  Medium (1K-10K):     359-636 MB/s
  Large (100K-1M):     727-734 MB/s  ← Peak

Decompression:
  Small (10-100):      1,143-1,145 MB/s  ← Peak
  Medium (1K-10K):     1,005-1,135 MB/s
  Large (100K-1M):     807-817 MB/s
```

### Scalability
- **Near-linear scaling** from 10K to 1M values
- **Compression gets faster** with larger datasets (better amortization)
- **Decompression slightly slower** on very large datasets (cache effects)

---

## Best Practices for Performance

### When to Use ALP

✅ **Ideal scenarios:**
- Time series data (sensors, metrics, logs)
- Financial data (prices, amounts)
- Scientific measurements
- Data with 1-3 decimal places
- Constant or near-constant values
- Batch processing of 1K-100K values

✅ **Performance sweet spots:**
- **Compression**: 100K-1M values (700+ MB/s)
- **Decompression**: 100-10K values (1,000+ MB/s)
- **Constant data**: Any size (10,000+ MB/s!)

### Optimization Tips

1. **Batch data** into 10K-100K value chunks for best throughput
2. **Use parallel processing** for multiple independent datasets (3.5x speedup)
3. **Pre-allocate buffers** if doing repeated operations
4. **Constant data** compresses 82x faster - detect and optimize
5. **Decompression is faster** - consider pre-compressing frequently-read data

### When NOT to Use ALP

❌ **Avoid for:**
- Truly random data with full precision (use Zstd instead)
- Very small datasets (<10 values) where overhead matters
- Data requiring >10 decimal places of precision
- Mixed data types (ALP is float64-only)

---

## Comparison with Industry Standards

| Library | Operation | Throughput | Notes |
|---------|-----------|------------|-------|
| **ALP** | Compress | **734 MB/s** | Float64-specific |
| **ALP** | Decompress | **1,145 MB/s** | Float64-specific |
| Zstd | Compress | 388 MB/s | General purpose |
| Zstd | Decompress | 603 MB/s | General purpose |
| LZ4 | Compress | ~500 MB/s | Fast general |
| LZ4 | Decompress | ~3,000 MB/s | Very fast |
| Snappy | Compress | ~500 MB/s | Fast general |

### ALP Advantages:
- ✅ **Faster compression** than Zstd (1.9x)
- ✅ **Faster decompression** than Zstd (1.9x)
- ✅ **Better compression ratio** on structured floats
- ✅ **Specialized** for float64 data

---

## Running the Benchmarks

```bash
# Run all speed benchmarks
go test -bench=BenchmarkCompressionSpeed -benchmem
go test -bench=BenchmarkDecompressionSpeed -benchmem

# Test different patterns
go test -bench=BenchmarkSpeedByPattern -benchmem

# Test parallel performance
go test -bench=BenchmarkParallel -benchmem

# Test round-trip
go test -bench=BenchmarkRoundTrip -benchmem

# Test scalability
go test -bench=BenchmarkScalability -benchmem

# Full suite
go test -bench=. -benchmem -benchtime=1s
```

---

## Conclusion

ALP delivers **excellent performance** for floating-point compression:

- 🚀 **Fast**: 300-1,100 MB/s for typical workloads
- 🚀 **Scalable**: Linear performance up to millions of values
- 🚀 **Efficient**: Only 3-5 allocations regardless of size
- 🚀 **Parallel-ready**: 3.5x speedup on multi-core systems
- 🚀 **Specialized**: Outperforms general-purpose compressors on float data

For **time series, sensor data, and structured floats**, ALP is a compelling choice offering both speed and excellent compression ratios.
