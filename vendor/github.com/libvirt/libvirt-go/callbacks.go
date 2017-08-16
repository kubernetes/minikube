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

// Helpers functions to register a Go callback function to a C
// function. For a simple example, look at how SetErrorFunc works in
// error.go.
//
// - Create a struct that will contain at least the Go callback to
//   invoke (errorContext).
//
// - Create an exported Golang function whose job will be to retrieve
//   the context and execute the callback in it
//   (connErrCallback). Such a function should receive a callback ID
//   and will use it to retrive the context.
//
// - Create a CGO function similar to the above function but with the
//   appropriate signature to be registered as a callback in C code
//   (connErrCallback_cgo). Notably, it will have a void* argument
//   that should be cast to long to retrieve the callback ID. It
//   should be just a thin wrapper to transform the opaque argument to
//   a callback ID.
//
// - Create a CGO function which will be a wrapper around the C
//   function to register the callback (virConnSetErrorFunc_cgo). Its
//   only role is to transform a callback ID (long) to an opaque (void*)
//   and call the C function.
//
// - When setting up a callback (SetErrorFunc), register the struct from first step
//   with registerCallbackId and invoke the CGO function from the
//   previous step with the appropriate ID.
//
// - When unregistering the callback, don't forget to call freecallbackId.
//
// If you need to associate some additional data with the connection,
// look at saveConnectionData, getConnectionData and
// releaseConnectionData.

import "C"

import (
	"sync"
)

const firstGoCallbackId int = 100 // help catch some additional errors during test
var goCallbackLock sync.RWMutex
var goCallbacks = make(map[int]interface{})
var nextGoCallbackId int = firstGoCallbackId

//export freeCallbackId
func freeCallbackId(goCallbackId int) {
	goCallbackLock.Lock()
	delete(goCallbacks, goCallbackId)
	goCallbackLock.Unlock()
}

func getCallbackId(goCallbackId int) interface{} {
	goCallbackLock.RLock()
	ctx := goCallbacks[goCallbackId]
	goCallbackLock.RUnlock()
	if ctx == nil {
		// If this happens there must be a bug in libvirt
		panic("Callback arrived after freeCallbackId was called")
	}
	return ctx
}

func registerCallbackId(ctx interface{}) int {
	goCallbackLock.Lock()
	goCallBackId := nextGoCallbackId
	nextGoCallbackId++
	for goCallbacks[nextGoCallbackId] != nil {
		nextGoCallbackId++
	}
	goCallbacks[goCallBackId] = ctx
	goCallbackLock.Unlock()
	return goCallBackId
}
