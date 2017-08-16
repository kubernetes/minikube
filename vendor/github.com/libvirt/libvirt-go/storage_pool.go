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
#include "storage_pool_compat.h"
*/
import "C"

import (
	"reflect"
	"unsafe"
)

type StoragePoolState int

const (
	STORAGE_POOL_INACTIVE     = StoragePoolState(C.VIR_STORAGE_POOL_INACTIVE)     // Not running
	STORAGE_POOL_BUILDING     = StoragePoolState(C.VIR_STORAGE_POOL_BUILDING)     // Initializing pool,not available
	STORAGE_POOL_RUNNING      = StoragePoolState(C.VIR_STORAGE_POOL_RUNNING)      // Running normally
	STORAGE_POOL_DEGRADED     = StoragePoolState(C.VIR_STORAGE_POOL_DEGRADED)     // Running degraded
	STORAGE_POOL_INACCESSIBLE = StoragePoolState(C.VIR_STORAGE_POOL_INACCESSIBLE) // Running,but not accessible
)

type StoragePoolBuildFlags int

const (
	STORAGE_POOL_BUILD_NEW          = StoragePoolBuildFlags(C.VIR_STORAGE_POOL_BUILD_NEW)          // Regular build from scratch
	STORAGE_POOL_BUILD_REPAIR       = StoragePoolBuildFlags(C.VIR_STORAGE_POOL_BUILD_REPAIR)       // Repair / reinitialize
	STORAGE_POOL_BUILD_RESIZE       = StoragePoolBuildFlags(C.VIR_STORAGE_POOL_BUILD_RESIZE)       // Extend existing pool
	STORAGE_POOL_BUILD_NO_OVERWRITE = StoragePoolBuildFlags(C.VIR_STORAGE_POOL_BUILD_NO_OVERWRITE) // Do not overwrite existing pool
	STORAGE_POOL_BUILD_OVERWRITE    = StoragePoolBuildFlags(C.VIR_STORAGE_POOL_BUILD_OVERWRITE)    // Overwrite data
)

type StoragePoolCreateFlags int

const (
	STORAGE_POOL_CREATE_NORMAL                  = StoragePoolCreateFlags(C.VIR_STORAGE_POOL_CREATE_NORMAL)
	STORAGE_POOL_CREATE_WITH_BUILD              = StoragePoolCreateFlags(C.VIR_STORAGE_POOL_CREATE_WITH_BUILD)
	STORAGE_POOL_CREATE_WITH_BUILD_OVERWRITE    = StoragePoolCreateFlags(C.VIR_STORAGE_POOL_CREATE_WITH_BUILD_OVERWRITE)
	STORAGE_POOL_CREATE_WITH_BUILD_NO_OVERWRITE = StoragePoolCreateFlags(C.VIR_STORAGE_POOL_CREATE_WITH_BUILD_NO_OVERWRITE)
)

type StoragePoolDeleteFlags int

const (
	STORAGE_POOL_DELETE_NORMAL = StoragePoolDeleteFlags(C.VIR_STORAGE_POOL_DELETE_NORMAL)
	STORAGE_POOL_DELETE_ZEROED = StoragePoolDeleteFlags(C.VIR_STORAGE_POOL_DELETE_ZEROED)
)

type StoragePoolEventID int

const (
	STORAGE_POOL_EVENT_ID_LIFECYCLE = StoragePoolEventID(C.VIR_STORAGE_POOL_EVENT_ID_LIFECYCLE)
	STORAGE_POOL_EVENT_ID_REFRESH   = StoragePoolEventID(C.VIR_STORAGE_POOL_EVENT_ID_REFRESH)
)

type StoragePoolEventLifecycleType int

const (
	STORAGE_POOL_EVENT_DEFINED   = StoragePoolEventLifecycleType(C.VIR_STORAGE_POOL_EVENT_DEFINED)
	STORAGE_POOL_EVENT_UNDEFINED = StoragePoolEventLifecycleType(C.VIR_STORAGE_POOL_EVENT_UNDEFINED)
	STORAGE_POOL_EVENT_STARTED   = StoragePoolEventLifecycleType(C.VIR_STORAGE_POOL_EVENT_STARTED)
	STORAGE_POOL_EVENT_STOPPED   = StoragePoolEventLifecycleType(C.VIR_STORAGE_POOL_EVENT_STOPPED)
)

type StoragePool struct {
	ptr C.virStoragePoolPtr
}

type StoragePoolInfo struct {
	State      StoragePoolState
	Capacity   uint64
	Allocation uint64
	Available  uint64
}

func (p *StoragePool) Build(flags StoragePoolBuildFlags) error {
	result := C.virStoragePoolBuild(p.ptr, C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (p *StoragePool) Create(flags StoragePoolCreateFlags) error {
	result := C.virStoragePoolCreate(p.ptr, C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (p *StoragePool) Delete(flags StoragePoolDeleteFlags) error {
	result := C.virStoragePoolDelete(p.ptr, C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (p *StoragePool) Destroy() error {
	result := C.virStoragePoolDestroy(p.ptr)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (p *StoragePool) Free() error {
	ret := C.virStoragePoolFree(p.ptr)
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (c *StoragePool) Ref() error {
	ret := C.virStoragePoolRef(c.ptr)
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (p *StoragePool) GetAutostart() (bool, error) {
	var out C.int
	result := C.virStoragePoolGetAutostart(p.ptr, (*C.int)(unsafe.Pointer(&out)))
	if result == -1 {
		return false, GetLastError()
	}
	switch out {
	case 1:
		return true, nil
	default:
		return false, nil
	}
}

func (p *StoragePool) GetInfo() (*StoragePoolInfo, error) {
	var cinfo C.virStoragePoolInfo
	result := C.virStoragePoolGetInfo(p.ptr, &cinfo)
	if result == -1 {
		return nil, GetLastError()
	}
	return &StoragePoolInfo{
		State:      StoragePoolState(cinfo.state),
		Capacity:   uint64(cinfo.capacity),
		Allocation: uint64(cinfo.allocation),
		Available:  uint64(cinfo.available),
	}, nil
}

func (p *StoragePool) GetName() (string, error) {
	name := C.virStoragePoolGetName(p.ptr)
	if name == nil {
		return "", GetLastError()
	}
	return C.GoString(name), nil
}

func (p *StoragePool) GetUUID() ([]byte, error) {
	var cUuid [C.VIR_UUID_BUFLEN](byte)
	cuidPtr := unsafe.Pointer(&cUuid)
	result := C.virStoragePoolGetUUID(p.ptr, (*C.uchar)(cuidPtr))
	if result != 0 {
		return []byte{}, GetLastError()
	}
	return C.GoBytes(cuidPtr, C.VIR_UUID_BUFLEN), nil
}

func (p *StoragePool) GetUUIDString() (string, error) {
	var cUuid [C.VIR_UUID_STRING_BUFLEN](C.char)
	cuidPtr := unsafe.Pointer(&cUuid)
	result := C.virStoragePoolGetUUIDString(p.ptr, (*C.char)(cuidPtr))
	if result != 0 {
		return "", GetLastError()
	}
	return C.GoString((*C.char)(cuidPtr)), nil
}

func (p *StoragePool) GetXMLDesc(flags StorageXMLFlags) (string, error) {
	result := C.virStoragePoolGetXMLDesc(p.ptr, C.uint(flags))
	if result == nil {
		return "", GetLastError()
	}
	xml := C.GoString(result)
	C.free(unsafe.Pointer(result))
	return xml, nil
}

func (p *StoragePool) IsActive() (bool, error) {
	result := C.virStoragePoolIsActive(p.ptr)
	if result == -1 {
		return false, GetLastError()
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (p *StoragePool) IsPersistent() (bool, error) {
	result := C.virStoragePoolIsPersistent(p.ptr)
	if result == -1 {
		return false, GetLastError()
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (p *StoragePool) SetAutostart(autostart bool) error {
	var cAutostart C.int
	switch autostart {
	case true:
		cAutostart = 1
	default:
		cAutostart = 0
	}
	result := C.virStoragePoolSetAutostart(p.ptr, cAutostart)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (p *StoragePool) Refresh(flags uint32) error {
	result := C.virStoragePoolRefresh(p.ptr, C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (p *StoragePool) Undefine() error {
	result := C.virStoragePoolUndefine(p.ptr)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (p *StoragePool) StorageVolCreateXML(xmlConfig string, flags StorageVolCreateFlags) (*StorageVol, error) {
	cXml := C.CString(string(xmlConfig))
	defer C.free(unsafe.Pointer(cXml))
	ptr := C.virStorageVolCreateXML(p.ptr, cXml, C.uint(flags))
	if ptr == nil {
		return nil, GetLastError()
	}
	return &StorageVol{ptr: ptr}, nil
}

func (p *StoragePool) StorageVolCreateXMLFrom(xmlConfig string, clonevol *StorageVol, flags StorageVolCreateFlags) (*StorageVol, error) {
	cXml := C.CString(string(xmlConfig))
	defer C.free(unsafe.Pointer(cXml))
	ptr := C.virStorageVolCreateXMLFrom(p.ptr, cXml, clonevol.ptr, C.uint(flags))
	if ptr == nil {
		return nil, GetLastError()
	}
	return &StorageVol{ptr: ptr}, nil
}

func (p *StoragePool) LookupStorageVolByName(name string) (*StorageVol, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	ptr := C.virStorageVolLookupByName(p.ptr, cName)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &StorageVol{ptr: ptr}, nil
}

func (p *StoragePool) NumOfStorageVolumes() (int, error) {
	result := int(C.virStoragePoolNumOfVolumes(p.ptr))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (p *StoragePool) ListStorageVolumes() ([]string, error) {
	const maxVols = 1024
	var names [maxVols](*C.char)
	namesPtr := unsafe.Pointer(&names)
	numStorageVols := C.virStoragePoolListVolumes(
		p.ptr,
		(**C.char)(namesPtr),
		maxVols)
	if numStorageVols == -1 {
		return nil, GetLastError()
	}
	goNames := make([]string, numStorageVols)
	for k := 0; k < int(numStorageVols); k++ {
		goNames[k] = C.GoString(names[k])
		C.free(unsafe.Pointer(names[k]))
	}
	return goNames, nil
}

func (p *StoragePool) ListAllStorageVolumes(flags uint32) ([]StorageVol, error) {
	var cList *C.virStorageVolPtr
	numVols := C.virStoragePoolListAllVolumes(p.ptr, (**C.virStorageVolPtr)(&cList), C.uint(flags))
	if numVols == -1 {
		return nil, GetLastError()
	}
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cList)),
		Len:  int(numVols),
		Cap:  int(numVols),
	}
	var pools []StorageVol
	slice := *(*[]C.virStorageVolPtr)(unsafe.Pointer(&hdr))
	for _, ptr := range slice {
		pools = append(pools, StorageVol{ptr})
	}
	C.free(unsafe.Pointer(cList))
	return pools, nil
}
