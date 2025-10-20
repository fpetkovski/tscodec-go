package main

import (
	"fmt"
	"math"
	"math/rand"

	"alp-go"
)

func main() {
	fmt.Println("=== ALP (Adaptive Lossless floating-Point compression) Demo ===")
	fmt.Println()

	// Example 1: Simple integer-like floats
	fmt.Println("Example 1: Integer-like floats")
	data1 := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0}
	demonstrateCompression("Integer-like floats", data1)

	// Example 2: Decimal values
	fmt.Println("\nExample 2: Decimal values with 1 digit precision")
	data2 := []float64{1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9}
	demonstrateCompression("Decimal values", data2)

	// Example 3: Small decimal values
	fmt.Println("\nExample 3: Small decimal values")
	data3 := []float64{0.001, 0.002, 0.003, 0.004, 0.005, 0.006, 0.007, 0.008}
	demonstrateCompression("Small decimals", data3)

	// Example 4: Scientific notation
	fmt.Println("\nExample 4: Scientific notation")
	data4 := []float64{1e-5, 2e-5, 3e-5, 4e-5, 5e-5}
	demonstrateCompression("Scientific notation", data4)

	// Example 5: Mixed range
	fmt.Println("\nExample 5: Mixed range")
	data5 := []float64{0.1, 1.0, 10.0, 100.0, 1000.0}
	demonstrateCompression("Mixed range", data5)

	// Example 6: Constant values
	fmt.Println("\nExample 6: Constant values (optimal compression)")
	data6 := make([]float64, 1000)
	for i := range data6 {
		data6[i] = 42.5
	}
	demonstrateCompression("Constant values", data6)

	// Example 7: Large dataset with pattern
	fmt.Println("\nExample 7: Large dataset (10,000 values)")
	data7 := make([]float64, 10000)
	for i := range data7 {
		data7[i] = float64(i) * 0.1
	}
	demonstrateCompression("Large dataset", data7)

	// Example 8: Random values
	fmt.Println("\nExample 8: Random sensor readings")
	data8 := make([]float64, 1000)
	for i := range data8 {
		// Simulate sensor readings between 20.0 and 30.0 with 2 decimal precision
		data8[i] = math.Round((20.0+rand.Float64()*10.0)*100) / 100
	}
	demonstrateCompression("Random sensor readings", data8)

	// Example 9: Time series data
	fmt.Println("\nExample 9: Time series data (incremental)")
	data9 := make([]float64, 1000)
	value := 100.0
	for i := range data9 {
		value += (rand.Float64() - 0.5) * 2.0 // Random walk
		data9[i] = math.Round(value*100) / 100
	}
	demonstrateCompression("Time series data", data9)
}

func demonstrateCompression(name string, data []float64) {
	// Show original data (first few and last few values if large)
	fmt.Printf("  Original data: ")
	if len(data) <= 10 {
		fmt.Printf("%v\n", data)
	} else {
		fmt.Printf("[%.3f, %.3f, %.3f, ... %.3f, %.3f, %.3f] (n=%d)\n",
			data[0], data[1], data[2],
			data[len(data)-3], data[len(data)-2], data[len(data)-1],
			len(data))
	}

	// Compress
	compressed := alp.Compress(data)

	// Decompress
	decompressed := alp.Decompress(compressed)

	// Calculate statistics
	originalSize := len(data) * 8 // float64 is 8 bytes
	compressedSize := len(compressed)
	ratio := alp.CompressionRatio(len(data), compressedSize)

	// Verify lossless
	lossless := true
	maxError := 0.0
	for i := range data {
		error := math.Abs(data[i] - decompressed[i])
		if error > maxError {
			maxError = error
		}
		if error > 1e-12 {
			lossless = false
		}
	}

	// Print results
	fmt.Printf("  Original size: %d bytes\n", originalSize)
	fmt.Printf("  Compressed size: %d bytes\n", compressedSize)
	fmt.Printf("  Compression ratio: %.2f%%\n", ratio*100)
	fmt.Printf("  Space saved: %d bytes (%.2f%%)\n", originalSize-compressedSize, (1-ratio)*100)
	fmt.Printf("  Lossless: %v (max error: %.15f)\n", lossless, maxError)

	// Verify a few values
	if len(data) <= 10 {
		fmt.Printf("  Decompressed: %v\n", decompressed)
	}
}
