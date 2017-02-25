// Copyright 2016 The go-qcow2 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// QCow2 image format specifications is under the QEMU license.

package qcow2

/*
#include <stdint.h>
#include <stdlib.h>

static inline int ctz32(uint32_t val)
{
    // Binary search for the trailing one bit
    int cnt;

    cnt = 0;
    if (!(val & 0x0000FFFFUL)) {
        cnt += 16;
        val >>= 16;
    }
    if (!(val & 0x000000FFUL)) {
        cnt += 8;
        val >>= 8;
    }
    if (!(val & 0x0000000FUL)) {
        cnt += 4;
        val >>= 4;
    }
    if (!(val & 0x00000003UL)) {
        cnt += 2;
        val >>= 2;
    }
    if (!(val & 0x00000001UL)) {
        cnt++;
        val >>= 1;
    }
    if (!(val & 0x00000001UL)) {
        cnt++;
    }

    return cnt;
}
*/
import "C"
import "unsafe"

func ctz32(val uint32) int {
	i := C.ctz32(C.uint32_t(val))
	return int(i)
}

func posixMemalign(alignment uint64, size uint64) []byte {
	var ptr C.void
	ptr2 := unsafe.Pointer(&ptr)
	writesize := C.posix_memalign(&ptr2, C.size_t(uintptr(alignment)), C.size_t(uintptr(size)))
	return []byte(C.GoBytes(unsafe.Pointer(&ptr), writesize))
}
