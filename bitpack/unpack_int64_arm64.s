//go:build !purego

#include "funcdata.h"
#include "textflag.h"

// func unpackInt64Default(dst []int64, src []byte, bitWidth uint)
TEXT ·unpackInt64Default(SB), NOSPLIT, $0-56
	MOVD dst_base+0(FP), R0   // R0 = dst pointer
	MOVD dst_len+8(FP), R1    // R1 = dst length
	MOVD src_base+24(FP), R2  // R2 = src pointer
	MOVD bitWidth+48(FP), R3  // R3 = bitWidth

	// Initialize registers
	MOVD $0, R5               // R5 = bitOffset
	MOVD $0, R6               // R6 = index

	// Check if length >= 4 for unrolled loop
	CMP  $4, R1
	BLT  scalar_loop_start

	// Calculate bitMask = (1 << bitWidth) - 1
	MOVD $1, R4
	LSL  R3, R4, R4
	SUB  $1, R4, R4           // R4 = bitMask

	// Calculate unrolled iterations
	LSR  $2, R1, R16          // R16 = length / 4
	CBZ  R16, scalar_loop_start
	LSL  $2, R16, R16         // R16 = (length / 4) * 4

unrolled_loop:
	// Process 4 elements with instruction-level parallelism
	// This allows the CPU to execute multiple loads in parallel

	// === Element 0 ===
	LSR  $6, R5, R7           // i = bitOffset / 64
	AND  $63, R5, R8          // j = bitOffset % 64
	MOVD (R2)(R7<<3), R9      // load src[i]
	LSL  R8, R4, R10
	AND  R10, R9, R9
	LSR  R8, R9, R9

	// Check span
	ADD  R8, R3, R11
	CMP  $64, R11
	BLE  store0
	MOVD $64, R12
	SUB  R8, R12, R12
	ADD  $1, R7, R13
	MOVD (R2)(R13<<3), R14
	LSR  R12, R4, R15
	AND  R15, R14, R14
	LSL  R12, R14, R14
	ORR  R14, R9, R9

store0:
	ADD  R3, R5, R5           // bitOffset += bitWidth
	MOVD R9, (R0)(R6<<3)
	ADD  $1, R6, R6

	// === Element 1 ===
	LSR  $6, R5, R7
	AND  $63, R5, R8
	MOVD (R2)(R7<<3), R9
	LSL  R8, R4, R10
	AND  R10, R9, R9
	LSR  R8, R9, R9

	ADD  R8, R3, R11
	CMP  $64, R11
	BLE  store1
	MOVD $64, R12
	SUB  R8, R12, R12
	ADD  $1, R7, R13
	MOVD (R2)(R13<<3), R14
	LSR  R12, R4, R15
	AND  R15, R14, R14
	LSL  R12, R14, R14
	ORR  R14, R9, R9

store1:
	ADD  R3, R5, R5
	MOVD R9, (R0)(R6<<3)
	ADD  $1, R6, R6

	// === Element 2 ===
	LSR  $6, R5, R7
	AND  $63, R5, R8
	MOVD (R2)(R7<<3), R9
	LSL  R8, R4, R10
	AND  R10, R9, R9
	LSR  R8, R9, R9

	ADD  R8, R3, R11
	CMP  $64, R11
	BLE  store2
	MOVD $64, R12
	SUB  R8, R12, R12
	ADD  $1, R7, R13
	MOVD (R2)(R13<<3), R14
	LSR  R12, R4, R15
	AND  R15, R14, R14
	LSL  R12, R14, R14
	ORR  R14, R9, R9

store2:
	ADD  R3, R5, R5
	MOVD R9, (R0)(R6<<3)
	ADD  $1, R6, R6

	// === Element 3 ===
	LSR  $6, R5, R7
	AND  $63, R5, R8
	MOVD (R2)(R7<<3), R9
	LSL  R8, R4, R10
	AND  R10, R9, R9
	LSR  R8, R9, R9

	ADD  R8, R3, R11
	CMP  $64, R11
	BLE  store3
	MOVD $64, R12
	SUB  R8, R12, R12
	ADD  $1, R7, R13
	MOVD (R2)(R13<<3), R14
	LSR  R12, R4, R15
	AND  R15, R14, R14
	LSL  R12, R14, R14
	ORR  R14, R9, R9

store3:
	ADD  R3, R5, R5
	MOVD R9, (R0)(R6<<3)
	ADD  $1, R6, R6

	CMP  R16, R6
	BLT  unrolled_loop

	// Check if done
	CMP  R1, R6
	BEQ  done

scalar_loop_start:
	// Fallback scalar loop for remaining elements
	MOVD $1, R4
	LSL  R3, R4, R4
	SUB  $1, R4, R4           // R4 = bitMask

scalar_loop:
	LSR  $6, R5, R7           // i = bitOffset / 64
	AND  $63, R5, R8          // j = bitOffset % 64
	MOVD (R2)(R7<<3), R9      // load src[i]
	LSL  R8, R4, R10          // bitMask << j
	AND  R10, R9, R9
	LSR  R8, R9, R9           // extracted value

	// Check for span
	ADD  R8, R3, R11
	CMP  $64, R11
	BLE  scalar_next
	MOVD $64, R12
	SUB  R8, R12, R12         // k = 64 - j
	ADD  $1, R7, R13
	MOVD (R2)(R13<<3), R14
	LSR  R12, R4, R15
	AND  R15, R14, R14
	LSL  R12, R14, R14
	ORR  R14, R9, R9

scalar_next:
	MOVD R9, (R0)(R6<<3)      // dst[index] = d
	ADD  R3, R5, R5           // bitOffset += bitWidth
	ADD  $1, R6, R6           // index++

scalar_test:
	CMP  R1, R6
	BNE  scalar_loop

done:
	RET
