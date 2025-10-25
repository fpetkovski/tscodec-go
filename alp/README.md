# ALP Algorithm - Detailed Explanation
`
Adaptive Lossless floating-Point (ALP) Compression Algorithm

`## Overview

ALP compresses float64 data by:
1. Finding optimal scale factor (exponent)
2. Converting floats → integers losslessly
3. Applying frame-of-reference encoding
4. Bit-packing to minimal width

---

## Compression Algorithm

### High-Level Pseudocode

```
function Compress(data: []float64) -> []byte:
    // Special case: constant values
    if all_values_equal(data):
        return metadata(CONSTANT, count, value)

    // Find optimal exponent
    exponent = find_best_exponent(data)
    factor = 10^exponent

    // Convert to integers
    int_values = []
    for each value in data:
        int_values.append(round(value * factor))

    // Frame-of-reference encoding
    min_value = min(int_values)
    max_value = max(int_values)

    for_values = []
    for each v in int_values:
        for_values.append(v - min_value)  // Now all non-negative

    // Find bit width needed
    bit_width = bits_needed(max_value - min_value)

    // Bit-pack the values
    packed_data = bit_pack(for_values, bit_width)

    // Store metadata + data
    metadata = encode_metadata(exponent, bit_width, min_value, count)
    return metadata || packed_data
```

### Detailed Steps

#### 1. **Constant Detection** (O(n))
```
function is_constant(data: []float64) -> bool:
    first = data[0]
    for each value in data[1:]:
        if value != first:
            return false
    return true

// If constant: compress to just 23 bytes (metadata only)
// Compression ratio: ~99.7% for 1000 values!
```

#### 2. **Find Best Exponent** (O(21 × sample_size))
```
function find_best_exponent(data: []float64) -> int:
    best_exp = 0
    min_bits = 64

    // Try exponents from -10 to +10
    for exp in range(-10, 10):
        factor = 10^exp
        max_bits = 0
        valid = true

        // Sample up to 1024 values
        for each sampled value in data:
            scaled = value * factor
            rounded = round(scaled)

            // Check if lossless
            reconstructed = rounded / factor
            if abs(value - reconstructed) > 1e-12 * abs(value):
                valid = false
                break

            // Track max bits needed
            bits = bits_needed(rounded)
            max_bits = max(max_bits, bits)

            if max_bits > 63:
                valid = false
                break

        // Keep exponent with smallest bit width
        if valid and max_bits > 0 and max_bits < min_bits:
            min_bits = max_bits
            best_exp = exp

    return best_exp

// Examples:
// [1.0, 2.0, 3.0] → exp=0 (factor=1), direct integers
// [1.1, 2.2, 3.3] → exp=1 (factor=10), scaled to [11, 22, 33]
// [0.001, 0.002] → exp=3 (factor=1000), scaled to [1, 2]
```

#### 3. **Float to Integer Conversion** (O(n))
```
function encode_to_integers(data: []float64, factor: float) -> []int64:
    int_values = []
    for each value in data:
        scaled = value * factor
        int_values.append(round(scaled))
    return int_values

// Example: [1.1, 2.2, 3.3] with factor=10
// → [11, 22, 33]
```

#### 4. **Frame-of-Reference (FOR) Encoding** (O(n))
```
function apply_frame_of_reference(int_values: []int64) -> ([]uint64, int64):
    // Find range
    min_value = min(int_values)
    max_value = max(int_values)

    // Subtract minimum (makes all values non-negative)
    for_values = []
    for each v in int_values:
        for_values.append(uint64(v - min_value))

    return (for_values, min_value)

// Example: [-100, -50, 0, 50, 100]
// min = -100
// → [0, 50, 100, 150, 200]
// Reduces bit width from 8 bits to 8 bits (but handles negatives)

// Example: [1000, 1001, 1002, 1003]
// min = 1000
// → [0, 1, 2, 3]
// Reduces bit width from 11 bits to 2 bits! 🎉
```

#### 5. **Bit Width Calculation** (O(n))
```
function find_max_bit_width(values: []uint64) -> int:
    max_bits = 0
    for each value in values:
        bits = bits_needed(value)
        max_bits = max(max_bits, bits)
    return max_bits

function bits_needed(value: uint64) -> int:
    if value == 0:
        return 1
    return 64 - leading_zeros(value)

// Examples:
// 0 → 1 bit
// 1 → 1 bit
// 3 → 2 bits
// 7 → 3 bits
// 255 → 8 bits
```

#### 6. **Bit Packing** (O(n × bit_width))
```
function bit_pack(values: []uint64, bit_width: int) -> []byte:
    total_bits = len(values) * bit_width
    buffer = new byte array[ceil(total_bits / 8)]
    bit_position = 0

    for each value in values:
        // Write 'bit_width' bits to buffer at bit_position
        write_bits(buffer, bit_position, value, bit_width)
        bit_position += bit_width

    return buffer

// Example: values=[0,1,2,3], bit_width=2
// Binary: 00 01 10 11
// Packed: 00011011 = 0x1B (1 byte instead of 32 bytes!)
```

#### 7. **Metadata Encoding** (23 bytes)
```
structure Metadata:
    encoding_type: uint8     // 1 byte  (1=ALP, 2=Constant, etc)
    count:         int32     // 4 bytes (number of values)
    exponent:      int8      // 1 byte  (-10 to +10)
    bit_width:     uint8     // 1 byte  (1-64 bits)
    frame_of_ref:  int64     // 8 bytes (minimum value)
    constant:      float64   // 8 bytes (for constant encoding)
    // Total: 23 bytes
```

---

## Decompression Algorithm

### High-Level Pseudocode

```
function Decompress(data: []byte) -> []float64:
    // Read metadata (first 23 bytes)
    metadata = decode_metadata(data[0:23])

    // Handle special cases
    if metadata.type == CONSTANT:
        return array_of(metadata.constant, metadata.count)

    if metadata.type == NONE:
        return []

    // Unpack bit-packed data
    packed_data = data[23:]
    for_values = bit_unpack(packed_data, metadata.count, metadata.bit_width)

    // Reverse frame-of-reference
    int_values = []
    for each v in for_values:
        int_values.append(v + metadata.frame_of_ref)

    // Convert back to float64
    factor = 10^metadata.exponent
    result = []
    for each v in int_values:
        result.append(v / factor)

    return result
```

### Detailed Steps

#### 1. **Metadata Decoding** (O(1))
```
function decode_metadata(data: []byte) -> Metadata:
    return Metadata{
        encoding_type: data[0],
        count:        read_int32(data[1:5]),
        exponent:     read_int8(data[5]),
        bit_width:    data[6],
        frame_of_ref: read_int64(data[7:15]),
        constant:     read_float64(data[15:23])
    }
```

#### 2. **Bit Unpacking** (O(n × bit_width))
```
function bit_unpack(data: []byte, count: int, bit_width: int) -> []uint64:
    result = []
    bit_position = 0

    for i in range(count):
        value = read_bits(data, bit_position, bit_width)
        result.append(value)
        bit_position += bit_width

    return result

// Example: data=[0x1B], count=4, bit_width=2
// Binary: 00011011
// Extract: 00, 01, 10, 11
// Result: [0, 1, 2, 3]
```

#### 3. **Reverse Frame-of-Reference** (O(n))
```
function reverse_for(for_values: []uint64, min_value: int64) -> []int64:
    int_values = []
    for each v in for_values:
        int_values.append(int64(v) + min_value)
    return int_values

// Example: [0, 50, 100, 150, 200] with min=-100
// → [-100, -50, 0, 50, 100]
```

#### 4. **Integer to Float Conversion** (O(n))
```
function decode_to_floats(int_values: []int64, exponent: int) -> []float64:
    factor = 10^exponent
    result = []
    for each v in int_values:
        result.append(float64(v) / factor)
    return result

// Example: [11, 22, 33] with exponent=1 (factor=10)
// → [1.1, 2.2, 3.3]
```

---

## Complexity Analysis

### Time Complexity

| Operation | Complexity | Notes |
|-----------|------------|-------|
| **Compression** | O(n) | Dominated by exponent search (sampling) |
| - Exponent search | O(21 × min(n, 1024)) | Try 21 exponents on sample |
| - Integer conversion | O(n) | Linear scan |
| - Frame-of-reference | O(n) | Min/max + transform |
| - Bit packing | O(n) | Linear with small constant |
| **Decompression** | O(n) | Pure linear operations |
| - Metadata decode | O(1) | Fixed 23 bytes |
| - Bit unpacking | O(n) | Linear |
| - FOR reverse | O(n) | Linear |
| - Float conversion | O(n) | Linear |

### Space Complexity

| Component | Space | Notes |
|-----------|-------|-------|
| **Metadata** | 23 bytes | Fixed overhead |
| **Packed data** | ⌈n × w / 8⌉ | n=count, w=bit_width |
| **Working memory** | O(n) | Temporary arrays during compression |

### Compression Ratio

```
ratio = (23 + ⌈n × w / 8⌉) / (n × 8)

Where:
- n = number of float64 values
- w = bit width (1-64)
- 8 = bytes per float64

Examples:
1. Sequential [0, 0.1, 0.2, ..., 999.9] (n=10000)
   - w ≈ 17 bits
   - ratio = (23 + 21,250) / 80,000 = 21.9%

2. Constant [42.5] × 1000 (n=1000)
   - Only metadata stored
   - ratio = 23 / 8,000 = 0.29%

3. Random sensor [20.00-30.00] × 1000 (n=1000)
   - w ≈ 10 bits
   - ratio = (23 + 1,250) / 8,000 = 15.9%
```

---

## Key Optimizations

### 1. **Sampling for Exponent Selection**
- Only test up to 1024 values instead of all values
- Reduces O(21n) to O(21 × 1024) for large datasets

### 2. **Early Termination**
```
if max_bits > 63:
    valid = false
    break  // Don't waste time on invalid exponents
```

### 3. **Constant Detection**
```
// Special case: O(n) scan for constant values
// Achieves 99.7% compression with no bit-packing overhead
```

### 4. **Frame-of-Reference**
```
// Reduces range: [1000, 1001, 1002] → [0, 1, 2]
// 11 bits → 2 bits (5.5x improvement!)
```

### 5. **Zigzag Encoding** (not currently used but implemented)
```
// Maps signed to unsigned: ..., -2, -1, 0, 1, 2, ...
//                       →  ...,  3,  1, 0, 2, 4, ...
// Efficient for small signed values
```

---

## Example: Complete Flow

### Input Data
```
data = [1.1, 2.2, 3.3, 4.4, 5.5]  // 40 bytes
```

### Compression Steps
```
1. Check constant? → No
2. Find best exponent:
   - Try exp=0: 1.1*1=1.1 (not integer) ❌
   - Try exp=1: 1.1*10=11.0 (integer!) ✓
   - Best: exp=1, factor=10

3. Convert to integers:
   [11, 22, 33, 44, 55]

4. Frame-of-reference:
   min=11
   [0, 11, 22, 33, 44]

5. Find bit width:
   max=44 → needs 6 bits

6. Bit pack (6 bits each):
   000000 001011 010110 100001 101100
   = 4 bytes packed

7. Create output:
   metadata (23 bytes) + packed (4 bytes) = 27 bytes

Result: 40 bytes → 27 bytes (67.5% ratio, 32.5% saved)
```

### Decompression Steps
```
1. Read metadata:
   - type=ALP, count=5, exp=1, width=6, min=11

2. Unpack 4 bytes → 5 values (6 bits each):
   [0, 11, 22, 33, 44]

3. Reverse FOR (add min=11):
   [11, 22, 33, 44, 55]

4. Convert to float (divide by factor=10):
   [1.1, 2.2, 3.3, 4.4, 5.5]

Result: Perfect reconstruction! ✓
```

---

## When ALP Works Best

### ✅ Optimal Cases
1. **Constant values**: 99.7% compression
2. **Limited precision** (1-3 decimals): 15-20% ratio
3. **Sequential patterns**: Frame-of-reference is very effective
4. **Small value ranges**: Reduces bit width dramatically

### ❌ Poor Cases
1. **Truly random floats**: Can't find good exponent
2. **High precision** (>3 decimals): Large bit widths
3. **Very small datasets** (<10 values): Metadata overhead
4. **Exponential ranges**: Can't scale all values together

---

## Comparison to Other Algorithms

| Algorithm | Approach | Best For |
|-----------|----------|----------|
| **ALP** | Adaptive scaling + bit-packing | Float64 with patterns |
| **Gorilla** | XOR + variable-length encoding | Time series with small changes |
| **Chimp** | Improved Gorilla with flags | Same as Gorilla |
| **Zstd** | Dictionary + entropy coding | General purpose, large data |
| **LZ4** | Byte-oriented compression | Fast general purpose |

---

## Implementation Notes

### Language-Specific Details (Go)

```go
// Efficient bit operations using math/bits
bits_needed = 64 - bits.LeadingZeros64(value)

// In-place transformations
for i := range intValues {
    forValues[i] = uint64(intValues[i] - minValue)
}

// Minimal allocations (only 3-4 per compression)
```

### Portability
- Uses little-endian byte order
- Standard IEEE 754 float64 representation
- Cross-platform compatible

---

## Future Optimizations

1. **Parallel exponent search** - Test exponents concurrently
2. **SIMD bit-packing** - Use vector instructions
3. **Dictionary encoding** - For repeated patterns
4. **Cascading compression** - Apply ALP recursively
5. **Adaptive sampling** - Dynamic sample size based on data patterns
