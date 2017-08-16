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
#include "storage_volume_compat.h"
*/
import "C"

import (
	"unsafe"
)

type StorageVolCreateFlags int

const (
	STORAGE_VOL_CREATE_PREALLOC_METADATA = StorageVolCreateFlags(C.VIR_STORAGE_VOL_CREATE_PREALLOC_METADATA)
	STORAGE_VOL_CREATE_REFLINK           = StorageVolCreateFlags(C.VIR_STORAGE_VOL_CREATE_REFLINK)
)

type StorageVolDeleteFlags int

const (
	STORAGE_VOL_DELETE_NORMAL         = StorageVolDeleteFlags(C.VIR_STORAGE_VOL_DELETE_NORMAL)         // Delete metadata only (fast)
	STORAGE_VOL_DELETE_ZEROED         = StorageVolDeleteFlags(C.VIR_STORAGE_VOL_DELETE_ZEROED)         // Clear all data to zeros (slow)
	STORAGE_VOL_DELETE_WITH_SNAPSHOTS = StorageVolDeleteFlags(C.VIR_STORAGE_VOL_DELETE_WITH_SNAPSHOTS) // Force removal of volume, even if in use
)

type StorageVolResizeFlags int

const (
	STORAGE_VOL_RESIZE_ALLOCATE = StorageVolResizeFlags(C.VIR_STORAGE_VOL_RESIZE_ALLOCATE) // force allocation of new size
	STORAGE_VOL_RESIZE_DELTA    = StorageVolResizeFlags(C.VIR_STORAGE_VOL_RESIZE_DELTA)    // size is relative to current
	STORAGE_VOL_RESIZE_SHRINK   = StorageVolResizeFlags(C.VIR_STORAGE_VOL_RESIZE_SHRINK)   // allow decrease in capacity
)

type StorageVolType int

const (
	STORAGE_VOL_FILE    = StorageVolType(C.VIR_STORAGE_VOL_FILE)    // Regular file based volumes
	STORAGE_VOL_BLOCK   = StorageVolType(C.VIR_STORAGE_VOL_BLOCK)   // Block based volumes
	STORAGE_VOL_DIR     = StorageVolType(C.VIR_STORAGE_VOL_DIR)     // Directory-passthrough based volume
	STORAGE_VOL_NETWORK = StorageVolType(C.VIR_STORAGE_VOL_NETWORK) //Network volumes like RBD (RADOS Block Device)
	STORAGE_VOL_NETDIR  = StorageVolType(C.VIR_STORAGE_VOL_NETDIR)  // Network accessible directory that can contain other network volumes
	STORAGE_VOL_PLOOP   = StorageVolType(C.VIR_STORAGE_VOL_PLOOP)   // Ploop directory based volumes
)

type StorageVolWipeAlgorithm int

const (
	STORAGE_VOL_WIPE_ALG_ZERO       = StorageVolWipeAlgorithm(C.VIR_STORAGE_VOL_WIPE_ALG_ZERO)       // 1-pass, all zeroes
	STORAGE_VOL_WIPE_ALG_NNSA       = StorageVolWipeAlgorithm(C.VIR_STORAGE_VOL_WIPE_ALG_NNSA)       // 4-pass NNSA Policy Letter NAP-14.1-C (XVI-8)
	STORAGE_VOL_WIPE_ALG_DOD        = StorageVolWipeAlgorithm(C.VIR_STORAGE_VOL_WIPE_ALG_DOD)        // 4-pass DoD 5220.22-M section 8-306 procedure
	STORAGE_VOL_WIPE_ALG_BSI        = StorageVolWipeAlgorithm(C.VIR_STORAGE_VOL_WIPE_ALG_BSI)        // 9-pass method recommended by the German Center of Security in Information Technologies
	STORAGE_VOL_WIPE_ALG_GUTMANN    = StorageVolWipeAlgorithm(C.VIR_STORAGE_VOL_WIPE_ALG_GUTMANN)    // The canonical 35-pass sequence
	STORAGE_VOL_WIPE_ALG_SCHNEIER   = StorageVolWipeAlgorithm(C.VIR_STORAGE_VOL_WIPE_ALG_SCHNEIER)   // 7-pass method described by Bruce Schneier in "Applied Cryptography" (1996)
	STORAGE_VOL_WIPE_ALG_PFITZNER7  = StorageVolWipeAlgorithm(C.VIR_STORAGE_VOL_WIPE_ALG_PFITZNER7)  // 7-pass random
	STORAGE_VOL_WIPE_ALG_PFITZNER33 = StorageVolWipeAlgorithm(C.VIR_STORAGE_VOL_WIPE_ALG_PFITZNER33) // 33-pass random
	STORAGE_VOL_WIPE_ALG_RANDOM     = StorageVolWipeAlgorithm(C.VIR_STORAGE_VOL_WIPE_ALG_RANDOM)     // 1-pass random
	STORAGE_VOL_WIPE_ALG_TRIM       = StorageVolWipeAlgorithm(C.VIR_STORAGE_VOL_WIPE_ALG_TRIM)       // Trim the underlying storage
)

type StorageXMLFlags int

const (
	STORAGE_XML_INACTIVE = StorageXMLFlags(C.VIR_STORAGE_XML_INACTIVE)
)

type StorageVolInfoFlags int

const (
	STORAGE_VOL_USE_ALLOCATION = StorageVolInfoFlags(C.VIR_STORAGE_VOL_USE_ALLOCATION)
	STORAGE_VOL_GET_PHYSICAL   = StorageVolInfoFlags(C.VIR_STORAGE_VOL_GET_PHYSICAL)
)

type StorageVolUploadFlags int

const (
	STORAGE_VOL_UPLOAD_SPARSE_STREAM = StorageVolUploadFlags(C.VIR_STORAGE_VOL_UPLOAD_SPARSE_STREAM)
)

type StorageVolDownloadFlags int

const (
	STORAGE_VOL_DOWNLOAD_SPARSE_STREAM = StorageVolDownloadFlags(C.VIR_STORAGE_VOL_DOWNLOAD_SPARSE_STREAM)
)

type StorageVol struct {
	ptr C.virStorageVolPtr
}

type StorageVolInfo struct {
	Type       StorageVolType
	Capacity   uint64
	Allocation uint64
}

func (v *StorageVol) Delete(flags StorageVolDeleteFlags) error {
	result := C.virStorageVolDelete(v.ptr, C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (v *StorageVol) Free() error {
	ret := C.virStorageVolFree(v.ptr)
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (c *StorageVol) Ref() error {
	ret := C.virStorageVolRef(c.ptr)
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (v *StorageVol) GetInfo() (*StorageVolInfo, error) {
	var cinfo C.virStorageVolInfo
	result := C.virStorageVolGetInfo(v.ptr, &cinfo)
	if result == -1 {
		return nil, GetLastError()
	}
	return &StorageVolInfo{
		Type:       StorageVolType(cinfo._type),
		Capacity:   uint64(cinfo.capacity),
		Allocation: uint64(cinfo.allocation),
	}, nil
}

func (v *StorageVol) GetInfoFlags(flags StorageVolInfoFlags) (*StorageVolInfo, error) {
	if C.LIBVIR_VERSION_NUMBER < 3000000 {
		return nil, GetNotImplementedError("virStorageVolGetInfoFlags")
	}

	var cinfo C.virStorageVolInfo
	result := C.virStorageVolGetInfoFlagsCompat(v.ptr, &cinfo, C.uint(flags))
	if result == -1 {
		return nil, GetLastError()
	}
	return &StorageVolInfo{
		Type:       StorageVolType(cinfo._type),
		Capacity:   uint64(cinfo.capacity),
		Allocation: uint64(cinfo.allocation),
	}, nil
}

func (v *StorageVol) GetKey() (string, error) {
	key := C.virStorageVolGetKey(v.ptr)
	if key == nil {
		return "", GetLastError()
	}
	return C.GoString(key), nil
}

func (v *StorageVol) GetName() (string, error) {
	name := C.virStorageVolGetName(v.ptr)
	if name == nil {
		return "", GetLastError()
	}
	return C.GoString(name), nil
}

func (v *StorageVol) GetPath() (string, error) {
	result := C.virStorageVolGetPath(v.ptr)
	if result == nil {
		return "", GetLastError()
	}
	path := C.GoString(result)
	C.free(unsafe.Pointer(result))
	return path, nil
}

func (v *StorageVol) GetXMLDesc(flags uint32) (string, error) {
	result := C.virStorageVolGetXMLDesc(v.ptr, C.uint(flags))
	if result == nil {
		return "", GetLastError()
	}
	xml := C.GoString(result)
	C.free(unsafe.Pointer(result))
	return xml, nil
}

func (v *StorageVol) Resize(capacity uint64, flags StorageVolResizeFlags) error {
	result := C.virStorageVolResize(v.ptr, C.ulonglong(capacity), C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (v *StorageVol) Wipe(flags uint32) error {
	result := C.virStorageVolWipe(v.ptr, C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}
func (v *StorageVol) WipePattern(algorithm StorageVolWipeAlgorithm, flags uint32) error {
	result := C.virStorageVolWipePattern(v.ptr, C.uint(algorithm), C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (v *StorageVol) Upload(stream *Stream, offset, length uint64, flags StorageVolUploadFlags) error {
	if C.virStorageVolUpload(v.ptr, stream.ptr, C.ulonglong(offset),
		C.ulonglong(length), C.uint(flags)) == -1 {
		return GetLastError()
	}
	return nil
}

func (v *StorageVol) Download(stream *Stream, offset, length uint64, flags StorageVolDownloadFlags) error {
	if C.virStorageVolDownload(v.ptr, stream.ptr, C.ulonglong(offset),
		C.ulonglong(length), C.uint(flags)) == -1 {
		return GetLastError()
	}
	return nil
}

func (v *StorageVol) LookupPoolByVolume() (*StoragePool, error) {
	poolPtr := C.virStoragePoolLookupByVolume(v.ptr)
	if poolPtr == nil {
		return nil, GetLastError()
	}
	return &StoragePool{ptr: poolPtr}, nil
}
