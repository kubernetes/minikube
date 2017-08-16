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
	"unsafe"
)

type InterfaceXMLFlags int

const (
	INTERFACE_XML_INACTIVE = InterfaceXMLFlags(C.VIR_INTERFACE_XML_INACTIVE)
)

type Interface struct {
	ptr C.virInterfacePtr
}

func (n *Interface) Create(flags uint32) error {
	result := C.virInterfaceCreate(n.ptr, C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (n *Interface) Destroy(flags uint32) error {
	result := C.virInterfaceDestroy(n.ptr, C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (n *Interface) IsActive() (bool, error) {
	result := C.virInterfaceIsActive(n.ptr)
	if result == -1 {
		return false, GetLastError()
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (n *Interface) GetMACString() (string, error) {
	result := C.virInterfaceGetMACString(n.ptr)
	if result == nil {
		return "", GetLastError()
	}
	mac := C.GoString(result)
	return mac, nil
}

func (n *Interface) GetName() (string, error) {
	result := C.virInterfaceGetName(n.ptr)
	if result == nil {
		return "", GetLastError()
	}
	name := C.GoString(result)
	return name, nil
}

func (n *Interface) GetXMLDesc(flags InterfaceXMLFlags) (string, error) {
	result := C.virInterfaceGetXMLDesc(n.ptr, C.uint(flags))
	if result == nil {
		return "", GetLastError()
	}
	xml := C.GoString(result)
	C.free(unsafe.Pointer(result))
	return xml, nil
}

func (n *Interface) Undefine() error {
	result := C.virInterfaceUndefine(n.ptr)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (n *Interface) Free() error {
	ret := C.virInterfaceFree(n.ptr)
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (c *Interface) Ref() error {
	ret := C.virInterfaceRef(c.ptr)
	if ret == -1 {
		return GetLastError()
	}
	return nil
}
