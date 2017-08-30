// +build !without_qemu

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
// Can't rely on pkg-config for libvirt-qemu since it was not
// installed until 2.6.0 onwards
#cgo LDFLAGS: -lvirt-qemu
#include <libvirt/libvirt.h>
#include <libvirt/libvirt-qemu.h>
#include <libvirt/virterror.h>
#include <stdlib.h>
#include "qemu_compat.h"
#include "qemu_cfuncs.h"
*/
import "C"

import (
	"unsafe"
)

/*
 * QMP has two different kinds of ways to talk to QEMU. One is legacy (HMP,
 * or 'human' monitor protocol. The default is QMP, which is all-JSON.
 *
 * QMP json commands are of the format:
 * 	{"execute" : "query-cpus"}
 *
 * whereas the same command in 'HMP' would be:
 *	'info cpus'
 */

type DomainQemuMonitorCommandFlags int

const (
	DOMAIN_QEMU_MONITOR_COMMAND_DEFAULT = DomainQemuMonitorCommandFlags(C.VIR_DOMAIN_QEMU_MONITOR_COMMAND_DEFAULT)
	DOMAIN_QEMU_MONITOR_COMMAND_HMP     = DomainQemuMonitorCommandFlags(C.VIR_DOMAIN_QEMU_MONITOR_COMMAND_HMP)
)

type DomainQemuAgentCommandTimeout int

const (
	DOMAIN_QEMU_AGENT_COMMAND_MIN      = DomainQemuAgentCommandTimeout(C.VIR_DOMAIN_QEMU_AGENT_COMMAND_MIN)
	DOMAIN_QEMU_AGENT_COMMAND_BLOCK    = DomainQemuAgentCommandTimeout(C.VIR_DOMAIN_QEMU_AGENT_COMMAND_BLOCK)
	DOMAIN_QEMU_AGENT_COMMAND_DEFAULT  = DomainQemuAgentCommandTimeout(C.VIR_DOMAIN_QEMU_AGENT_COMMAND_DEFAULT)
	DOMAIN_QEMU_AGENT_COMMAND_NOWAIT   = DomainQemuAgentCommandTimeout(C.VIR_DOMAIN_QEMU_AGENT_COMMAND_NOWAIT)
	DOMAIN_QEMU_AGENT_COMMAND_SHUTDOWN = DomainQemuAgentCommandTimeout(C.VIR_DOMAIN_QEMU_AGENT_COMMAND_SHUTDOWN)
)

type DomainQemuMonitorEventFlags int

const (
	CONNECT_DOMAIN_QEMU_MONITOR_EVENT_REGISTER_REGEX  = DomainQemuMonitorEventFlags(C.VIR_CONNECT_DOMAIN_QEMU_MONITOR_EVENT_REGISTER_REGEX)
	CONNECT_DOMAIN_QEMU_MONITOR_EVENT_REGISTER_NOCASE = DomainQemuMonitorEventFlags(C.VIR_CONNECT_DOMAIN_QEMU_MONITOR_EVENT_REGISTER_NOCASE)
)

func (d *Domain) QemuMonitorCommand(command string, flags DomainQemuMonitorCommandFlags) (string, error) {
	var cResult *C.char
	cCommand := C.CString(command)
	defer C.free(unsafe.Pointer(cCommand))
	result := C.virDomainQemuMonitorCommand(d.ptr, cCommand, &cResult, C.uint(flags))

	if result != 0 {
		return "", GetLastError()
	}

	rstring := C.GoString(cResult)
	C.free(unsafe.Pointer(cResult))
	return rstring, nil
}

func (d *Domain) QemuAgentCommand(command string, timeout DomainQemuAgentCommandTimeout, flags uint32) (string, error) {
	cCommand := C.CString(command)
	defer C.free(unsafe.Pointer(cCommand))
	result := C.virDomainQemuAgentCommand(d.ptr, cCommand, C.int(timeout), C.uint(flags))

	if result == nil {
		return "", GetLastError()
	}

	rstring := C.GoString(result)
	C.free(unsafe.Pointer(result))
	return rstring, nil
}

func (c *Connect) DomainQemuAttach(pid uint32, flags uint32) (*Domain, error) {

	ptr := C.virDomainQemuAttach(c.ptr, C.uint(pid), C.uint(flags))
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Domain{ptr: ptr}, nil
}

type DomainQemuMonitorEvent struct {
	Event   string
	Seconds int64
	Micros  uint
	Details string
}

type DomainQemuMonitorEventCallback func(c *Connect, d *Domain, event *DomainQemuMonitorEvent)

//export domainQemuMonitorEventCallback
func domainQemuMonitorEventCallback(c C.virConnectPtr, d C.virDomainPtr,
	event *C.char, seconds C.longlong, micros C.uint, details *C.char, goCallbackId int) {

	domain := &Domain{ptr: d}
	connection := &Connect{ptr: c}

	eventDetails := &DomainQemuMonitorEvent{
		Event:   C.GoString(event),
		Seconds: int64(seconds),
		Micros:  uint(micros),
		Details: C.GoString(details),
	}

	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(DomainQemuMonitorEventCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, domain, eventDetails)

}

func (c *Connect) DomainQemuMonitorEventRegister(dom *Domain, event string, callback DomainQemuMonitorEventCallback, flags DomainQemuMonitorEventFlags) (int, error) {
	if C.LIBVIR_VERSION_NUMBER < 1002003 {
		return 0, GetNotImplementedError("virConnectDomainQemuMonitorEventRegister")
	}

	cEvent := C.CString(event)
	defer C.free(unsafe.Pointer(cEvent))
	goCallBackId := registerCallbackId(callback)

	callbackPtr := unsafe.Pointer(C.domainQemuMonitorEventCallback_cgo)
	var cdom C.virDomainPtr
	if dom != nil {
		cdom = dom.ptr
	}
	ret := C.virConnectDomainQemuMonitorEventRegister_cgo(c.ptr, cdom,
		cEvent,
		C.virConnectDomainEventGenericCallback(callbackPtr),
		C.long(goCallBackId),
		C.uint(flags))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) DomainQemuEventDeregister(callbackId int) error {
	if C.LIBVIR_VERSION_NUMBER < 1002003 {
		return GetNotImplementedError("virConnectDomainQemuMonitorEventDeregister")
	}

	// Deregister the callback
	if i := int(C.virConnectDomainQemuMonitorEventDeregisterCompat(c.ptr, C.int(callbackId))); i != 0 {
		return GetLastError()
	}
	return nil
}
