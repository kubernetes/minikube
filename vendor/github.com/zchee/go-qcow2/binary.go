// Copyright 2016 The go-qcow2 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// QCow2 image format specifications is under the QEMU license.

package qcow2

import "encoding/binary"

func BEUint16(b []byte) uint16 {
	return binary.BigEndian.Uint16(b)
}

func BEUint32(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
}

func BEUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

// BEUvarint8 convert the uint8 type of varint(varying-length integer) to the binary data of big endian format byte order.
func BEUvarint8(i uint8) []byte {
	dst := [1]byte{}
	binary.PutUvarint(dst[:], uint64(i))
	return dst[:]
}

// BEUvarint16 convert the uint16 type of varint(varying-length integer) to the binary data of big endian format byte order.
func BEUvarint16(i uint16) []byte {
	dst := [2]byte{}
	binary.BigEndian.PutUint16(dst[:], i)
	return dst[:]
}

// BEUvarint32 convert the uint32 type of varint(varying-length integer) to the binary data of big endian format byte order.
func BEUvarint32(i uint32) []byte {
	dst := [4]byte{}
	binary.BigEndian.PutUint32(dst[:], i)
	return dst[:]
}

// BEUvarint64 convert the uint64 type of varint(varying-length integer) to the binary data of big endian format byte order.
func BEUvarint64(i uint64) []byte {
	dst := [8]byte{}
	binary.BigEndian.PutUint64(dst[:], i)
	return dst[:]
}
