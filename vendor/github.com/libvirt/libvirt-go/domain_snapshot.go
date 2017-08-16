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
*/
import "C"

import (
	"reflect"
	"unsafe"
)

type DomainSnapshotCreateFlags int

const (
	DOMAIN_SNAPSHOT_CREATE_REDEFINE    = DomainSnapshotCreateFlags(C.VIR_DOMAIN_SNAPSHOT_CREATE_REDEFINE)
	DOMAIN_SNAPSHOT_CREATE_CURRENT     = DomainSnapshotCreateFlags(C.VIR_DOMAIN_SNAPSHOT_CREATE_CURRENT)
	DOMAIN_SNAPSHOT_CREATE_NO_METADATA = DomainSnapshotCreateFlags(C.VIR_DOMAIN_SNAPSHOT_CREATE_NO_METADATA)
	DOMAIN_SNAPSHOT_CREATE_HALT        = DomainSnapshotCreateFlags(C.VIR_DOMAIN_SNAPSHOT_CREATE_HALT)
	DOMAIN_SNAPSHOT_CREATE_DISK_ONLY   = DomainSnapshotCreateFlags(C.VIR_DOMAIN_SNAPSHOT_CREATE_DISK_ONLY)
	DOMAIN_SNAPSHOT_CREATE_REUSE_EXT   = DomainSnapshotCreateFlags(C.VIR_DOMAIN_SNAPSHOT_CREATE_REUSE_EXT)
	DOMAIN_SNAPSHOT_CREATE_QUIESCE     = DomainSnapshotCreateFlags(C.VIR_DOMAIN_SNAPSHOT_CREATE_QUIESCE)
	DOMAIN_SNAPSHOT_CREATE_ATOMIC      = DomainSnapshotCreateFlags(C.VIR_DOMAIN_SNAPSHOT_CREATE_ATOMIC)
	DOMAIN_SNAPSHOT_CREATE_LIVE        = DomainSnapshotCreateFlags(C.VIR_DOMAIN_SNAPSHOT_CREATE_LIVE)
)

type DomainSnapshotListFlags int

const (
	DOMAIN_SNAPSHOT_LIST_ROOTS       = DomainSnapshotListFlags(C.VIR_DOMAIN_SNAPSHOT_LIST_ROOTS)
	DOMAIN_SNAPSHOT_LIST_DESCENDANTS = DomainSnapshotListFlags(C.VIR_DOMAIN_SNAPSHOT_LIST_DESCENDANTS)
	DOMAIN_SNAPSHOT_LIST_LEAVES      = DomainSnapshotListFlags(C.VIR_DOMAIN_SNAPSHOT_LIST_LEAVES)
	DOMAIN_SNAPSHOT_LIST_NO_LEAVES   = DomainSnapshotListFlags(C.VIR_DOMAIN_SNAPSHOT_LIST_NO_LEAVES)
	DOMAIN_SNAPSHOT_LIST_METADATA    = DomainSnapshotListFlags(C.VIR_DOMAIN_SNAPSHOT_LIST_METADATA)
	DOMAIN_SNAPSHOT_LIST_NO_METADATA = DomainSnapshotListFlags(C.VIR_DOMAIN_SNAPSHOT_LIST_NO_METADATA)
	DOMAIN_SNAPSHOT_LIST_INACTIVE    = DomainSnapshotListFlags(C.VIR_DOMAIN_SNAPSHOT_LIST_INACTIVE)
	DOMAIN_SNAPSHOT_LIST_ACTIVE      = DomainSnapshotListFlags(C.VIR_DOMAIN_SNAPSHOT_LIST_ACTIVE)
	DOMAIN_SNAPSHOT_LIST_DISK_ONLY   = DomainSnapshotListFlags(C.VIR_DOMAIN_SNAPSHOT_LIST_DISK_ONLY)
	DOMAIN_SNAPSHOT_LIST_INTERNAL    = DomainSnapshotListFlags(C.VIR_DOMAIN_SNAPSHOT_LIST_INTERNAL)
	DOMAIN_SNAPSHOT_LIST_EXTERNAL    = DomainSnapshotListFlags(C.VIR_DOMAIN_SNAPSHOT_LIST_EXTERNAL)
)

type DomainSnapshotRevertFlags int

const (
	DOMAIN_SNAPSHOT_REVERT_RUNNING = DomainSnapshotRevertFlags(C.VIR_DOMAIN_SNAPSHOT_REVERT_RUNNING)
	DOMAIN_SNAPSHOT_REVERT_PAUSED  = DomainSnapshotRevertFlags(C.VIR_DOMAIN_SNAPSHOT_REVERT_PAUSED)
	DOMAIN_SNAPSHOT_REVERT_FORCE   = DomainSnapshotRevertFlags(C.VIR_DOMAIN_SNAPSHOT_REVERT_FORCE)
)

type DomainSnapshotDeleteFlags int

const (
	DOMAIN_SNAPSHOT_DELETE_CHILDREN      = DomainSnapshotDeleteFlags(C.VIR_DOMAIN_SNAPSHOT_DELETE_CHILDREN)
	DOMAIN_SNAPSHOT_DELETE_METADATA_ONLY = DomainSnapshotDeleteFlags(C.VIR_DOMAIN_SNAPSHOT_DELETE_METADATA_ONLY)
	DOMAIN_SNAPSHOT_DELETE_CHILDREN_ONLY = DomainSnapshotDeleteFlags(C.VIR_DOMAIN_SNAPSHOT_DELETE_CHILDREN_ONLY)
)

type DomainSnapshot struct {
	ptr C.virDomainSnapshotPtr
}

func (s *DomainSnapshot) Free() error {
	ret := C.virDomainSnapshotFree(s.ptr)
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (c *DomainSnapshot) Ref() error {
	ret := C.virDomainSnapshotRef(c.ptr)
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (s *DomainSnapshot) Delete(flags DomainSnapshotDeleteFlags) error {
	result := C.virDomainSnapshotDelete(s.ptr, C.uint(flags))
	if result != 0 {
		return GetLastError()
	}
	return nil
}

func (s *DomainSnapshot) RevertToSnapshot(flags DomainSnapshotRevertFlags) error {
	result := C.virDomainRevertToSnapshot(s.ptr, C.uint(flags))
	if result != 0 {
		return GetLastError()
	}
	return nil
}

func (s *DomainSnapshot) IsCurrent(flags uint32) (bool, error) {
	result := C.virDomainSnapshotIsCurrent(s.ptr, C.uint(flags))
	if result == -1 {
		return false, GetLastError()
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (s *DomainSnapshot) HasMetadata(flags uint32) (bool, error) {
	result := C.virDomainSnapshotHasMetadata(s.ptr, C.uint(flags))
	if result == -1 {
		return false, GetLastError()
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (s *DomainSnapshot) GetXMLDesc(flags DomainXMLFlags) (string, error) {
	result := C.virDomainSnapshotGetXMLDesc(s.ptr, C.uint(flags))
	if result == nil {
		return "", GetLastError()
	}
	xml := C.GoString(result)
	C.free(unsafe.Pointer(result))
	return xml, nil
}

func (s *DomainSnapshot) GetName() (string, error) {
	name := C.virDomainSnapshotGetName(s.ptr)
	if name == nil {
		return "", GetLastError()
	}
	return C.GoString(name), nil
}

func (s *DomainSnapshot) GetParent(flags uint32) (*DomainSnapshot, error) {
	ptr := C.virDomainSnapshotGetParent(s.ptr, C.uint(flags))
	if ptr == nil {
		return nil, GetLastError()
	}
	return &DomainSnapshot{ptr: ptr}, nil
}

func (s *DomainSnapshot) NumChildren(flags DomainSnapshotListFlags) (int, error) {
	result := int(C.virDomainSnapshotNumChildren(s.ptr, C.uint(flags)))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (s *DomainSnapshot) ListChildrenNames(flags DomainSnapshotListFlags) ([]string, error) {
	const maxNames = 1024
	var names [maxNames](*C.char)
	namesPtr := unsafe.Pointer(&names)
	numNames := C.virDomainSnapshotListChildrenNames(
		s.ptr,
		(**C.char)(namesPtr),
		maxNames, C.uint(flags))
	if numNames == -1 {
		return nil, GetLastError()
	}
	goNames := make([]string, numNames)
	for k := 0; k < int(numNames); k++ {
		goNames[k] = C.GoString(names[k])
		C.free(unsafe.Pointer(names[k]))
	}
	return goNames, nil
}

func (d *DomainSnapshot) ListAllChildren(flags DomainSnapshotListFlags) ([]DomainSnapshot, error) {
	var cList *C.virDomainSnapshotPtr
	numVols := C.virDomainSnapshotListAllChildren(d.ptr, (**C.virDomainSnapshotPtr)(&cList), C.uint(flags))
	if numVols == -1 {
		return nil, GetLastError()
	}
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cList)),
		Len:  int(numVols),
		Cap:  int(numVols),
	}
	var pools []DomainSnapshot
	slice := *(*[]C.virDomainSnapshotPtr)(unsafe.Pointer(&hdr))
	for _, ptr := range slice {
		pools = append(pools, DomainSnapshot{ptr})
	}
	C.free(unsafe.Pointer(cList))
	return pools, nil
}
