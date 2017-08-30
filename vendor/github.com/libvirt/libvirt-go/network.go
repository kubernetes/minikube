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
#include "network_compat.h"
*/
import "C"

import (
	"reflect"
	"time"
	"unsafe"
)

type IPAddrType int

const (
	IP_ADDR_TYPE_IPV4 = IPAddrType(C.VIR_IP_ADDR_TYPE_IPV4)
	IP_ADDR_TYPE_IPV6 = IPAddrType(C.VIR_IP_ADDR_TYPE_IPV6)
)

type NetworkXMLFlags int

const (
	NETWORK_XML_INACTIVE = NetworkXMLFlags(C.VIR_NETWORK_XML_INACTIVE)
)

type NetworkUpdateCommand int

const (
	NETWORK_UPDATE_COMMAND_NONE      = NetworkUpdateCommand(C.VIR_NETWORK_UPDATE_COMMAND_NONE)
	NETWORK_UPDATE_COMMAND_MODIFY    = NetworkUpdateCommand(C.VIR_NETWORK_UPDATE_COMMAND_MODIFY)
	NETWORK_UPDATE_COMMAND_DELETE    = NetworkUpdateCommand(C.VIR_NETWORK_UPDATE_COMMAND_DELETE)
	NETWORK_UPDATE_COMMAND_ADD_LAST  = NetworkUpdateCommand(C.VIR_NETWORK_UPDATE_COMMAND_ADD_LAST)
	NETWORK_UPDATE_COMMAND_ADD_FIRST = NetworkUpdateCommand(C.VIR_NETWORK_UPDATE_COMMAND_ADD_FIRST)
)

type NetworkUpdateSection int

const (
	NETWORK_SECTION_NONE              = NetworkUpdateSection(C.VIR_NETWORK_SECTION_NONE)
	NETWORK_SECTION_BRIDGE            = NetworkUpdateSection(C.VIR_NETWORK_SECTION_BRIDGE)
	NETWORK_SECTION_DOMAIN            = NetworkUpdateSection(C.VIR_NETWORK_SECTION_DOMAIN)
	NETWORK_SECTION_IP                = NetworkUpdateSection(C.VIR_NETWORK_SECTION_IP)
	NETWORK_SECTION_IP_DHCP_HOST      = NetworkUpdateSection(C.VIR_NETWORK_SECTION_IP_DHCP_HOST)
	NETWORK_SECTION_IP_DHCP_RANGE     = NetworkUpdateSection(C.VIR_NETWORK_SECTION_IP_DHCP_RANGE)
	NETWORK_SECTION_FORWARD           = NetworkUpdateSection(C.VIR_NETWORK_SECTION_FORWARD)
	NETWORK_SECTION_FORWARD_INTERFACE = NetworkUpdateSection(C.VIR_NETWORK_SECTION_FORWARD_INTERFACE)
	NETWORK_SECTION_FORWARD_PF        = NetworkUpdateSection(C.VIR_NETWORK_SECTION_FORWARD_PF)
	NETWORK_SECTION_PORTGROUP         = NetworkUpdateSection(C.VIR_NETWORK_SECTION_PORTGROUP)
	NETWORK_SECTION_DNS_HOST          = NetworkUpdateSection(C.VIR_NETWORK_SECTION_DNS_HOST)
	NETWORK_SECTION_DNS_TXT           = NetworkUpdateSection(C.VIR_NETWORK_SECTION_DNS_TXT)
	NETWORK_SECTION_DNS_SRV           = NetworkUpdateSection(C.VIR_NETWORK_SECTION_DNS_SRV)
)

type NetworkUpdateFlags int

const (
	NETWORK_UPDATE_AFFECT_CURRENT = NetworkUpdateFlags(C.VIR_NETWORK_UPDATE_AFFECT_CURRENT)
	NETWORK_UPDATE_AFFECT_LIVE    = NetworkUpdateFlags(C.VIR_NETWORK_UPDATE_AFFECT_LIVE)
	NETWORK_UPDATE_AFFECT_CONFIG  = NetworkUpdateFlags(C.VIR_NETWORK_UPDATE_AFFECT_CONFIG)
)

type NetworkEventLifecycleType int

const (
	NETWORK_EVENT_DEFINED   = NetworkEventLifecycleType(C.VIR_NETWORK_EVENT_DEFINED)
	NETWORK_EVENT_UNDEFINED = NetworkEventLifecycleType(C.VIR_NETWORK_EVENT_UNDEFINED)
	NETWORK_EVENT_STARTED   = NetworkEventLifecycleType(C.VIR_NETWORK_EVENT_STARTED)
	NETWORK_EVENT_STOPPED   = NetworkEventLifecycleType(C.VIR_NETWORK_EVENT_STOPPED)
)

type NetworkEventID int

const (
	NETWORK_EVENT_ID_LIFECYCLE = NetworkEventID(C.VIR_NETWORK_EVENT_ID_LIFECYCLE)
)

type Network struct {
	ptr C.virNetworkPtr
}

type NetworkDHCPLease struct {
	Iface      string
	ExpiryTime time.Time
	Type       IPAddrType
	Mac        string
	Iaid       string
	IPaddr     string
	Prefix     uint
	Hostname   string
	Clientid   string
}

func (n *Network) Free() error {
	ret := C.virNetworkFree(n.ptr)
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (c *Network) Ref() error {
	ret := C.virNetworkRef(c.ptr)
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (n *Network) Create() error {
	result := C.virNetworkCreate(n.ptr)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (n *Network) Destroy() error {
	result := C.virNetworkDestroy(n.ptr)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (n *Network) IsActive() (bool, error) {
	result := C.virNetworkIsActive(n.ptr)
	if result == -1 {
		return false, GetLastError()
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (n *Network) IsPersistent() (bool, error) {
	result := C.virNetworkIsPersistent(n.ptr)
	if result == -1 {
		return false, GetLastError()
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (n *Network) GetAutostart() (bool, error) {
	var out C.int
	result := C.virNetworkGetAutostart(n.ptr, (*C.int)(unsafe.Pointer(&out)))
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

func (n *Network) SetAutostart(autostart bool) error {
	var cAutostart C.int
	switch autostart {
	case true:
		cAutostart = 1
	default:
		cAutostart = 0
	}
	result := C.virNetworkSetAutostart(n.ptr, cAutostart)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (n *Network) GetName() (string, error) {
	name := C.virNetworkGetName(n.ptr)
	if name == nil {
		return "", GetLastError()
	}
	return C.GoString(name), nil
}

func (n *Network) GetUUID() ([]byte, error) {
	var cUuid [C.VIR_UUID_BUFLEN](byte)
	cuidPtr := unsafe.Pointer(&cUuid)
	result := C.virNetworkGetUUID(n.ptr, (*C.uchar)(cuidPtr))
	if result != 0 {
		return []byte{}, GetLastError()
	}
	return C.GoBytes(cuidPtr, C.VIR_UUID_BUFLEN), nil
}

func (n *Network) GetUUIDString() (string, error) {
	var cUuid [C.VIR_UUID_STRING_BUFLEN](C.char)
	cuidPtr := unsafe.Pointer(&cUuid)
	result := C.virNetworkGetUUIDString(n.ptr, (*C.char)(cuidPtr))
	if result != 0 {
		return "", GetLastError()
	}
	return C.GoString((*C.char)(cuidPtr)), nil
}

func (n *Network) GetBridgeName() (string, error) {
	result := C.virNetworkGetBridgeName(n.ptr)
	if result == nil {
		return "", GetLastError()
	}
	bridge := C.GoString(result)
	C.free(unsafe.Pointer(result))
	return bridge, nil
}

func (n *Network) GetXMLDesc(flags NetworkXMLFlags) (string, error) {
	result := C.virNetworkGetXMLDesc(n.ptr, C.uint(flags))
	if result == nil {
		return "", GetLastError()
	}
	xml := C.GoString(result)
	C.free(unsafe.Pointer(result))
	return xml, nil
}

func (n *Network) Undefine() error {
	result := C.virNetworkUndefine(n.ptr)
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (n *Network) Update(cmd NetworkUpdateCommand, section NetworkUpdateSection, parentIndex int, xml string, flags NetworkUpdateFlags) error {
	cxml := C.CString(xml)
	defer C.free(unsafe.Pointer(cxml))
	result := C.virNetworkUpdate(n.ptr, C.uint(cmd), C.uint(section), C.int(parentIndex), cxml, C.uint(flags))
	if result == -1 {
		return GetLastError()
	}
	return nil
}

func (n *Network) GetDHCPLeases() ([]NetworkDHCPLease, error) {
	if C.LIBVIR_VERSION_NUMBER < 1002006 {
		return []NetworkDHCPLease{}, GetNotImplementedError("virNetworkGetDHCPLeases")
	}
	var cLeases *C.virNetworkDHCPLeasePtr
	numLeases := C.virNetworkGetDHCPLeasesCompat(n.ptr, nil, (**C.virNetworkDHCPLeasePtr)(&cLeases), C.uint(0))
	if numLeases == -1 {
		return nil, GetLastError()
	}
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cLeases)),
		Len:  int(numLeases),
		Cap:  int(numLeases),
	}
	var leases []NetworkDHCPLease
	slice := *(*[]C.virNetworkDHCPLeasePtr)(unsafe.Pointer(&hdr))
	for _, clease := range slice {
		leases = append(leases, NetworkDHCPLease{
			Iface:      C.GoString(clease.iface),
			ExpiryTime: time.Unix(int64(clease.expirytime), 0),
			Type:       IPAddrType(clease._type),
			Mac:        C.GoString(clease.mac),
			Iaid:       C.GoString(clease.iaid),
			IPaddr:     C.GoString(clease.ipaddr),
			Prefix:     uint(clease.prefix),
			Hostname:   C.GoString(clease.hostname),
			Clientid:   C.GoString(clease.clientid),
		})
		C.virNetworkDHCPLeaseFreeCompat(clease)
	}
	C.free(unsafe.Pointer(cLeases))
	return leases, nil
}
