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
#include "node_device_compat.h"
*/
import "C"

import (
	"unsafe"
)

type NodeDeviceEventID int

const (
	NODE_DEVICE_EVENT_ID_LIFECYCLE = NodeDeviceEventID(C.VIR_NODE_DEVICE_EVENT_ID_LIFECYCLE)
	NODE_DEVICE_EVENT_ID_UPDATE    = NodeDeviceEventID(C.VIR_NODE_DEVICE_EVENT_ID_UPDATE)
)

type NodeDeviceEventLifecycleType int

const (
	NODE_DEVICE_EVENT_CREATED = NodeDeviceEventLifecycleType(C.VIR_NODE_DEVICE_EVENT_CREATED)
	NODE_DEVICE_EVENT_DELETED = NodeDeviceEventLifecycleType(C.VIR_NODE_DEVICE_EVENT_DELETED)
)

type NodeDevice struct {
	ptr C.virNodeDevicePtr
}

func (n *NodeDevice) Free() error {
	ret := C.virNodeDeviceFree(n.ptr)
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (c *NodeDevice) Ref() error {
	ret := C.virNodeDeviceRef(c.ptr)
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (n *NodeDevice) Destroy() error {
	result := C.virNodeDeviceDestroy(n.ptr)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (n *NodeDevice) Reset() error {
	result := C.virNodeDeviceReset(n.ptr)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (n *NodeDevice) Detach() error {
	result := C.virNodeDeviceDettach(n.ptr)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (n *NodeDevice) DetachFlags(driverName string, flags uint32) error {
	cDriverName := C.CString(driverName)
	defer C.free(unsafe.Pointer(cDriverName))
	result := C.virNodeDeviceDetachFlags(n.ptr, cDriverName, C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (n *NodeDevice) ReAttach() error {
	result := C.virNodeDeviceReAttach(n.ptr)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (n *NodeDevice) GetName() (string, error) {
	name := C.virNodeDeviceGetName(n.ptr)
	if name == nil {
		return "", GetLastError()
	}
	return C.GoString(name), nil
}

func (n *NodeDevice) GetXMLDesc(flags uint32) (string, error) {
	result := C.virNodeDeviceGetXMLDesc(n.ptr, C.uint(flags))
	if result == nil {
		return "", GetLastError()
	}
	xml := C.GoString(result)
	C.free(unsafe.Pointer(result))
	return xml, nil
}

func (n *NodeDevice) GetParent() (string, error) {
	result := C.virNodeDeviceGetParent(n.ptr)
	if result == nil {
		return "", GetLastError()
	}
	defer C.free(unsafe.Pointer(result))
	return C.GoString(result), nil
}

func (p *NodeDevice) NumOfStorageCaps() (int, error) {
	result := int(C.virNodeDeviceNumOfCaps(p.ptr))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (p *NodeDevice) ListStorageCaps() ([]string, error) {
	const maxCaps = 1024
	var names [maxCaps](*C.char)
	namesPtr := unsafe.Pointer(&names)
	numCaps := C.virNodeDeviceListCaps(
		p.ptr,
		(**C.char)(namesPtr),
		maxCaps)
	if numCaps == -1 {
		return nil, GetLastError()
	}
	goNames := make([]string, numCaps)
	for k := 0; k < int(numCaps); k++ {
		goNames[k] = C.GoString(names[k])
		C.free(unsafe.Pointer(names[k]))
	}
	return goNames, nil
}
