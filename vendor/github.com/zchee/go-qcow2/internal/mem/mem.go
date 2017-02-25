package mem

import "unsafe"

//go:noescape
func set(dst []byte, value byte)

// Set imitation C memset using plan9 assembly
func Set(dst []byte, value byte) {
	set(dst, value)
}

//go:noescape
func move(to uintptr, from uintptr, n uintptr)

// Move copies n bytes from "src" to "dst", which imitation C memmove using plan9 assembly.
func Move(dst, src []byte) {
	move(uintptr(unsafe.Pointer(&dst[0])), uintptr(unsafe.Pointer(&src[0])), uintptr(len(src)))
}

//go:noescape
func cpy(to uintptr, from uintptr, n uintptr)

// Cpy imitation C memcpy using plan9 assembly
func Cpy(dst, src []byte, n uintptr) {
	if n == 0 {
		return
	}
	cpy(uintptr(unsafe.Pointer(&dst[0])), uintptr(unsafe.Pointer(&src[0])), n)
}
