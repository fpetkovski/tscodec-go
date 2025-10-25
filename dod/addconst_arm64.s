//go:build arm64

#include "textflag.h"

// func addConstInt64(dst []int64, c int64)
TEXT ·addConstInt64(SB), NOSPLIT, $0-32
    MOVD dst_base+0(FP), R0   // dst pointer
    MOVD dst_len+8(FP), R1    // length
    MOVD c+24(FP), R2          // constant

    CBZ R1, done

    // Main loop - process 8 elements at a time
    CMP $8, R1
    BLT tail4

loop8:
    LDP (R0), (R3, R4)           // Load 2 int64
    LDP 16(R0), (R5, R6)         // Load 2 more
    LDP 32(R0), (R7, R8)         // Load 2 more
    LDP 48(R0), (R9, R10)        // Load 2 more

    ADD R2, R3, R3
    ADD R2, R4, R4
    ADD R2, R5, R5
    ADD R2, R6, R6
    ADD R2, R7, R7
    ADD R2, R8, R8
    ADD R2, R9, R9
    ADD R2, R10, R10

    STP (R3, R4), (R0)
    STP (R5, R6), 16(R0)
    STP (R7, R8), 32(R0)
    STP (R9, R10), 48(R0)

    ADD $64, R0
    SUB $8, R1
    CMP $8, R1
    BGE loop8

tail4:
    CMP $4, R1
    BLT tail2

    LDP (R0), (R3, R4)
    LDP 16(R0), (R5, R6)
    ADD R2, R3, R3
    ADD R2, R4, R4
    ADD R2, R5, R5
    ADD R2, R6, R6
    STP (R3, R4), (R0)
    STP (R5, R6), 16(R0)
    ADD $32, R0
    SUB $4, R1

tail2:
    CMP $2, R1
    BLT tail1

    LDP (R0), (R3, R4)
    ADD R2, R3, R3
    ADD R2, R4, R4
    STP (R3, R4), (R0)
    ADD $16, R0
    SUB $2, R1

tail1:
    CBZ R1, done
    MOVD (R0), R3
    ADD R2, R3
    MOVD R3, (R0)

done:
    RET

// func addConstInt32(dst []int32, c int32)
TEXT ·addConstInt32(SB), NOSPLIT, $0-28
    MOVD dst_base+0(FP), R0
    MOVD dst_len+8(FP), R1
    MOVW c+24(FP), R2

    CBZ R1, done32

    CMP $8, R1
    BLT tail32_4

loop32_8:
    MOVW (R0), R3
    MOVW 4(R0), R4
    MOVW 8(R0), R5
    MOVW 12(R0), R6
    MOVW 16(R0), R7
    MOVW 20(R0), R8
    MOVW 24(R0), R9
    MOVW 28(R0), R10

    ADD R2, R3, R3
    ADD R2, R4, R4
    ADD R2, R5, R5
    ADD R2, R6, R6
    ADD R2, R7, R7
    ADD R2, R8, R8
    ADD R2, R9, R9
    ADD R2, R10, R10

    MOVW R3, (R0)
    MOVW R4, 4(R0)
    MOVW R5, 8(R0)
    MOVW R6, 12(R0)
    MOVW R7, 16(R0)
    MOVW R8, 20(R0)
    MOVW R9, 24(R0)
    MOVW R10, 28(R0)

    ADD $32, R0
    SUB $8, R1
    CMP $8, R1
    BGE loop32_8

tail32_4:
    CMP $4, R1
    BLT tail32_1

    MOVW (R0), R3
    MOVW 4(R0), R4
    MOVW 8(R0), R5
    MOVW 12(R0), R6
    ADD R2, R3, R3
    ADD R2, R4, R4
    ADD R2, R5, R5
    ADD R2, R6, R6
    MOVW R3, (R0)
    MOVW R4, 4(R0)
    MOVW R5, 8(R0)
    MOVW R6, 12(R0)
    ADD $16, R0
    SUB $4, R1

tail32_1:
    CBZ R1, done32
    MOVW (R0), R3
    ADD R2, R3
    MOVW R3, (R0)
    ADD $4, R0
    SUB $1, R1
    B tail32_1

done32:
    RET

// func addConstUint64(dst []uint64, c uint64)
TEXT ·addConstUint64(SB), NOSPLIT, $0-32
    // Same as int64 version
    JMP ·addConstInt64(SB)
