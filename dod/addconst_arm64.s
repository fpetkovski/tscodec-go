//go:build arm64

#include "textflag.h"

// func addConstInt64(dst []int64, c int64)
TEXT ·addConstInt64(SB), NOSPLIT, $0-32
    MOVD dst_base+0(FP), R0   // dst pointer
    MOVD dst_len+8(FP), R1    // length
    MOVD c+24(FP), R2          // constant

    CBZ R1, done

    // Duplicate constant to vector register (2x int64)
    VMOV R2, V0.D[0]
    VMOV R2, V0.D[1]

    // Main loop - process 16 elements at a time using SIMD
    CMP $16, R1
    BLT tail8

loop16:
    // Prefetch 2 cache lines ahead
    PRFM 128(R0), PLDL1KEEP
    PRFM 192(R0), PLDL1KEEP

    // Load first 8 int64 values
    VLD1 (R0), [V1.D2, V2.D2, V3.D2, V4.D2]
    ADD $64, R0
    // Load next 8 int64 values
    VLD1 (R0), [V5.D2, V6.D2, V7.D2, V8.D2]

    // Add constant to all vectors
    VADD V0.D2, V1.D2, V1.D2
    VADD V0.D2, V2.D2, V2.D2
    VADD V0.D2, V3.D2, V3.D2
    VADD V0.D2, V4.D2, V4.D2
    VADD V0.D2, V5.D2, V5.D2
    VADD V0.D2, V6.D2, V6.D2
    VADD V0.D2, V7.D2, V7.D2
    VADD V0.D2, V8.D2, V8.D2

    // Store both halves back
    SUB $64, R0
    VST1 [V1.D2, V2.D2, V3.D2, V4.D2], (R0)
    ADD $64, R0
    VST1 [V5.D2, V6.D2, V7.D2, V8.D2], (R0)

    ADD $64, R0
    SUB $16, R1
    CMP $16, R1
    BGE loop16

tail8:
    CMP $8, R1
    BLT tail4

    PRFM 64(R0), PLDL1KEEP
    VLD1 (R0), [V1.D2, V2.D2, V3.D2, V4.D2]
    VADD V0.D2, V1.D2, V1.D2
    VADD V0.D2, V2.D2, V2.D2
    VADD V0.D2, V3.D2, V3.D2
    VADD V0.D2, V4.D2, V4.D2
    VST1 [V1.D2, V2.D2, V3.D2, V4.D2], (R0)
    ADD $64, R0
    SUB $8, R1

tail4:
    CMP $4, R1
    BLT tail2

    // Load 4 int64 values (2 vectors)
    VLD1 (R0), [V1.D2, V2.D2]
    VADD V0.D2, V1.D2, V1.D2
    VADD V0.D2, V2.D2, V2.D2
    VST1 [V1.D2, V2.D2], (R0)
    ADD $32, R0
    SUB $4, R1

tail2:
    CMP $2, R1
    BLT tail1

    // Load 2 int64 values (1 vector)
    VLD1 (R0), [V1.D2]
    VADD V0.D2, V1.D2, V1.D2
    VST1 [V1.D2], (R0)
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

    // Duplicate constant to vector register (4x int32)
    VMOV R2, V0.S[0]
    VMOV R2, V0.S[1]
    VMOV R2, V0.S[2]
    VMOV R2, V0.S[3]

    // Main loop - process 16 elements at a time using SIMD
    CMP $16, R1
    BLT tail32_8

loop32_16:
    // Prefetch next cache line
    PRFM 64(R0), PLDL1KEEP

    // Load 16 int32 values (4 vectors of 4 int32 each)
    VLD1 (R0), [V1.S4, V2.S4, V3.S4, V4.S4]

    // Add constant vector to each
    VADD V0.S4, V1.S4, V1.S4
    VADD V0.S4, V2.S4, V2.S4
    VADD V0.S4, V3.S4, V3.S4
    VADD V0.S4, V4.S4, V4.S4

    // Store back
    VST1 [V1.S4, V2.S4, V3.S4, V4.S4], (R0)

    ADD $64, R0
    SUB $16, R1
    CMP $16, R1
    BGE loop32_16

tail32_8:
    CMP $8, R1
    BLT tail32_4

    // Load 8 int32 values (2 vectors)
    VLD1 (R0), [V1.S4, V2.S4]
    VADD V0.S4, V1.S4, V1.S4
    VADD V0.S4, V2.S4, V2.S4
    VST1 [V1.S4, V2.S4], (R0)
    ADD $32, R0
    SUB $8, R1

tail32_4:
    CMP $4, R1
    BLT tail32_1

    // Load 4 int32 values (1 vector)
    VLD1 (R0), [V1.S4]
    VADD V0.S4, V1.S4, V1.S4
    VST1 [V1.S4], (R0)
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
