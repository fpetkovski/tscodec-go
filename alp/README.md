# Adaptive Lossless floating-Point (ALP) Compression Algorithm

## Overview 

The algorithm is originally described in a [CIW Paper](https://github.com/cwida/ALP).

ALP compresses float64 data by:
1. Finding optimal scale factor (exponent)
2. Converting floats → integers losslessly
3. Applying frame-of-reference encoding
4. Bit-packing to minimal width

---

## When ALP Works Best

### ✅ Optimal Cases
1. **Constant values**: 99.7% compression
2. **Limited precision** (1-3 decimals): 15-20% ratio
3. **Sequential patterns** by leveraging frame-of-reference effective
4. **Small value ranges**: by reduces bit width dramatically

### ❌ Poor Cases
1. **Truly random floats**: Unable to find a good exponent
2. **High precision** (>3 decimals): Large bit widths
3. **Very small datasets** (<10 values): Metadata overhead
4. **Exponential ranges**: Cannot scale all values together

