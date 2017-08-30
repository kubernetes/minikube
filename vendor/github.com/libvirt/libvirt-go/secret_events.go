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
#include "secret_compat.h"
#include "secret_events_cfuncs.h"
*/
import "C"

type SecretEventLifecycle struct {
	Event SecretEventLifecycleType
	// TODO: we can make Detail typesafe somehow ?
	Detail int
}

type SecretEventLifecycleCallback func(c *Connect, n *Secret, event *SecretEventLifecycle)

type SecretEventGenericCallback func(c *Connect, n *Secret)

//export secretEventLifecycleCallback
func secretEventLifecycleCallback(c C.virConnectPtr, n C.virSecretPtr,
	event int, detail int,
	goCallbackId int) {

	secret := &Secret{ptr: n}
	connection := &Connect{ptr: c}

	eventDetails := &SecretEventLifecycle{
		Event:  SecretEventLifecycleType(event),
		Detail: detail,
	}

	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(SecretEventLifecycleCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, secret, eventDetails)
}

//export secretEventGenericCallback
func secretEventGenericCallback(c C.virConnectPtr, n C.virSecretPtr,
	goCallbackId int) {

	secret := &Secret{ptr: n}
	connection := &Connect{ptr: c}

	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(SecretEventGenericCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(connection, secret)
}

func (c *Connect) SecretEventLifecycleRegister(secret *Secret, callback SecretEventLifecycleCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)
	if C.LIBVIR_VERSION_NUMBER < 3000000 {
		return 0, GetNotImplementedError("virConnectSecretEventRegisterAny")
	}

	callbackPtr := unsafe.Pointer(C.secretEventLifecycleCallback_cgo)
	var csecret C.virSecretPtr
	if secret != nil {
		csecret = secret.ptr
	}
	ret := C.virConnectSecretEventRegisterAny_cgo(c.ptr, csecret,
		C.VIR_SECRET_EVENT_ID_LIFECYCLE,
		C.virConnectSecretEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) SecretEventValueChangedRegister(secret *Secret, callback SecretEventGenericCallback) (int, error) {
	goCallBackId := registerCallbackId(callback)
	if C.LIBVIR_VERSION_NUMBER < 3000000 {
		return 0, GetNotImplementedError("virConnectSecretEventRegisterAny")
	}

	callbackPtr := unsafe.Pointer(C.secretEventGenericCallback_cgo)
	var csecret C.virSecretPtr
	if secret != nil {
		csecret = secret.ptr
	}
	ret := C.virConnectSecretEventRegisterAny_cgo(c.ptr, csecret,
		C.VIR_SECRET_EVENT_ID_VALUE_CHANGED,
		C.virConnectSecretEventGenericCallback(callbackPtr),
		C.long(goCallBackId))
	if ret == -1 {
		freeCallbackId(goCallBackId)
		return 0, GetLastError()
	}
	return int(ret), nil
}

func (c *Connect) SecretEventDeregister(callbackId int) error {
	if C.LIBVIR_VERSION_NUMBER < 3000000 {
		return GetNotImplementedError("virConnectSecretEventDeregisterAny")
	}
	// Deregister the callback
	if i := int(C.virConnectSecretEventDeregisterAnyCompat(c.ptr, C.int(callbackId))); i != 0 {
		return GetLastError()
	}
	return nil
}

func (e SecretEventLifecycle) String() string {
	var event string
	switch e.Event {
	case SECRET_EVENT_DEFINED:
		event = "defined"

	case SECRET_EVENT_UNDEFINED:
		event = "undefined"

	default:
		event = "unknown"
	}

	return fmt.Sprintf("Secret event=%q", event)
}
