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

import (
	"fmt"
	"unsafe"
)

/*
#cgo pkg-config: libvirt
#include <libvirt/libvirt.h>
#include "storage_pool_compat.h"
#include "storage_pool_events_cfuncs.h"
*/
import "C"

type StoragePoolEventLifecycle struct {
	Event StoragePoolEventLifecycleType
	// TODO: we can make Detail typesafe somehow ?
	Detail int
}

type StoragePoolEventLifecycleCallback func(c *Connect, n *StoragePool, event *StoragePoolEventLifecycle)

type StoragePoolEventGenericCallback func(c *Connect, n *StoragePool)

//export storagePoolEventLifecycleCallback
func storagePoolEventLifecycleCallback(c C.virConnectPtr, s C.virStoragePoolPtr,
	event int, detail int,
	goCallbackId int) {

	storage_pool := &StoragePool{ptr: s}
	connection := &Connect{ptr: c}

	eventDetails := &StoragePoolEventLifecycle{
		Event:  StoragePoolEventLifecycleType(event),
		Detail: detail,
	}

	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(StoragePoolEventLifecycleCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, storage_pool, eventDetails)
}

//export storagePoolEventGenericCallback
func storagePoolEventGenericCallback(c C.virConnectPtr, s C.virStoragePoolPtr,
	goCallbackId int) {

	storage_pool := &StoragePool{ptr: s}
	connection := &Connect{ptr: c}

	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(StoragePoolEventGenericCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, storage_pool)
}

func (c *Connect) StoragePoolEventLifecycleRegister(pool *StoragePool, callback StoragePoolEventLifecycleCallback) (int, error) {
	if C.LIBVIR_VERSION_NUMBER < 2000000 {
		return 0, GetNotImplementedError("virConnectStoragePoolEventRegisterAny")
	}

	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.storagePoolEventLifecycleCallback_cgo)
	var cpool C.virStoragePoolPtr
	if pool != nil {
		cpool = pool.ptr
	}
	ret := C.virConnectStoragePoolEventRegisterAny_cgo(c.ptr, cpool,
		C.VIR_STORAGE_POOL_EVENT_ID_LIFECYCLE,
		C.virConnectStoragePoolEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) StoragePoolEventRefreshRegister(pool *StoragePool, callback StoragePoolEventGenericCallback) (int, error) {
	if C.LIBVIR_VERSION_NUMBER < 2000000 {
		return 0, GetNotImplementedError("virConnectStoragePoolEventRegisterAny")
	}

	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.storagePoolEventGenericCallback_cgo)
	var cpool C.virStoragePoolPtr
	if pool != nil {
		cpool = pool.ptr
	}
	ret := C.virConnectStoragePoolEventRegisterAny_cgo(c.ptr, cpool,
		C.VIR_STORAGE_POOL_EVENT_ID_REFRESH,
		C.virConnectStoragePoolEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) StoragePoolEventDeregister(callbackId int) error {
	if C.LIBVIR_VERSION_NUMBER < 2000000 {
		return GetNotImplementedError("virConnectStoragePoolEventDeregisterAny")
	}

	// Deregister the callback
	if i := int(C.virConnectStoragePoolEventDeregisterAnyCompat(c.ptr, C.int(callbackId))); i != 0 {
		return GetLastError()
	}
	return nil
}

func (e StoragePoolEventLifecycle) String() string {
	var event string
	switch e.Event {
	case STORAGE_POOL_EVENT_DEFINED:
		event = "defined"

	case STORAGE_POOL_EVENT_UNDEFINED:
		event = "undefined"

	case STORAGE_POOL_EVENT_STARTED:
		event = "started"

	case STORAGE_POOL_EVENT_STOPPED:
		event = "stopped"

	default:
		event = "unknown"
	}

	return fmt.Sprintf("StoragePool event=%q", event)
}
