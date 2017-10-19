/*
 * This file is part of the libvirt-go project
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 *
 * Copyright (c) 2013 Alex Zorin
 * Copyright (C) 2016 Red Hat, Inc.
 *
 */

package libvirt

/*
#cgo pkg-config: libvirt
#include <libvirt/libvirt.h>
#include <libvirt/virterror.h>
#include <stdlib.h>
#include "secret_compat.h"
*/
import "C"

import (
	"unsafe"
)

type SecretUsageType int

const (
	SECRET_USAGE_TYPE_NONE   = SecretUsageType(C.VIR_SECRET_USAGE_TYPE_NONE)
	SECRET_USAGE_TYPE_VOLUME = SecretUsageType(C.VIR_SECRET_USAGE_TYPE_VOLUME)
	SECRET_USAGE_TYPE_CEPH   = SecretUsageType(C.VIR_SECRET_USAGE_TYPE_CEPH)
	SECRET_USAGE_TYPE_ISCSI  = SecretUsageType(C.VIR_SECRET_USAGE_TYPE_ISCSI)
	SECRET_USAGE_TYPE_TLS    = SecretUsageType(C.VIR_SECRET_USAGE_TYPE_TLS)
)

type SecretEventLifecycleType int

const (
	SECRET_EVENT_DEFINED   = SecretEventLifecycleType(C.VIR_SECRET_EVENT_DEFINED)
	SECRET_EVENT_UNDEFINED = SecretEventLifecycleType(C.VIR_SECRET_EVENT_UNDEFINED)
)

type SecretEventID int

const (
	SECRET_EVENT_ID_LIFECYCLE     = SecretEventID(C.VIR_SECRET_EVENT_ID_LIFECYCLE)
	SECRET_EVENT_ID_VALUE_CHANGED = SecretEventID(C.VIR_SECRET_EVENT_ID_VALUE_CHANGED)
)

type Secret struct {
	ptr C.virSecretPtr
}

func (s *Secret) Free() error {
	ret := C.virSecretFree(s.ptr)
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (c *Secret) Ref() error {
	ret := C.virSecretRef(c.ptr)
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (s *Secret) Undefine() error {
	result := C.virSecretUndefine(s.ptr)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (s *Secret) GetUUID() ([]byte, error) {
	var cUuid [C.VIR_UUID_BUFLEN](byte)
	cuidPtr := unsafe.Pointer(&cUuid)
	result := C.virSecretGetUUID(s.ptr, (*C.uchar)(cuidPtr))
	if result != 0 {
		return []byte{}, GetLastError()
	}
	return C.GoBytes(cuidPtr, C.VIR_UUID_BUFLEN), nil
}

func (s *Secret) GetUUIDString() (string, error) {
	var cUuid [C.VIR_UUID_STRING_BUFLEN](C.char)
	cuidPtr := unsafe.Pointer(&cUuid)
	result := C.virSecretGetUUIDString(s.ptr, (*C.char)(cuidPtr))
	if result != 0 {
		return "", GetLastError()
	}
	return C.GoString((*C.char)(cuidPtr)), nil
}

func (s *Secret) GetUsageID() (string, error) {
	result := C.virSecretGetUsageID(s.ptr)
	if result == nil {
		return "", GetLastError()
	}
	return C.GoString(result), nil
}

func (s *Secret) GetUsageType() (SecretUsageType, error) {
	result := SecretUsageType(C.virSecretGetUsageType(s.ptr))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (s *Secret) GetXMLDesc(flags uint32) (string, error) {
	result := C.virSecretGetXMLDesc(s.ptr, C.uint(flags))
	if result == nil {
		return "", GetLastError()
	}
	xml := C.GoString(result)
	C.free(unsafe.Pointer(result))
	return xml, nil
}

func (s *Secret) GetValue(flags uint32) ([]byte, error) {
	var cvalue_size C.size_t

	cvalue := C.virSecretGetValue(s.ptr, &cvalue_size, C.uint(flags))
	if cvalue == nil {
		return nil, GetLastError()
	}
	defer C.free(unsafe.Pointer(cvalue))
	ret := C.GoBytes(unsafe.Pointer(cvalue), C.int(cvalue_size))
	return ret, nil
}

func (s *Secret) SetValue(value []byte, flags uint32) error {
	cvalue := make([]C.uchar, len(value))

	for i := 0; i < len(value); i++ {
		cvalue[i] = C.uchar(value[i])
	}

	result := C.virSecretSetValue(s.ptr, &cvalue[0], C.size_t(len(value)), C.uint(flags))

	if result == -1 {
		return GetLastError()
	}

	return nil
}
