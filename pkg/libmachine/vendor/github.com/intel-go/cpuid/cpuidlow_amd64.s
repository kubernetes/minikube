// Copyright 2017 Intel Corporation.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
#include "textflag.h"

// func cpuid_low(arg1, arg2 uint32) (eax, ebx, ecx, edx uint32)
TEXT ·cpuid_low(SB),NOSPLIT,$0-24
    MOVL    arg1+0(FP), AX
    MOVL    arg2+4(FP), CX
    CPUID
    MOVL AX, eax+8(FP)
    MOVL BX, ebx+12(FP)
    MOVL CX, ecx+16(FP)
    MOVL DX, edx+20(FP)
    RET 
// func xgetbv_low(arg1 uint32) (eax, edx uint32)
    TEXT ·xgetbv_low(SB),NOSPLIT,$0-16
    MOVL arg1+0(FP), CX
    BYTE $0x0F
    BYTE $0x01
    BYTE $0xD0
    MOVL AX,eax+8(FP)
    MOVL DX,edx+12(FP)
    RET
