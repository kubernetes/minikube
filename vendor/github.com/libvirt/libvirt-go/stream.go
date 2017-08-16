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
#include "stream_cfuncs.h"
#include "stream_compat.h"
*/
import "C"
import (
	"io"
	"unsafe"
)

type StreamFlags int

const (
	STREAM_NONBLOCK = StreamFlags(C.VIR_STREAM_NONBLOCK)
)

type StreamEventType int

const (
	STREAM_EVENT_READABLE = StreamEventType(C.VIR_STREAM_EVENT_READABLE)
	STREAM_EVENT_WRITABLE = StreamEventType(C.VIR_STREAM_EVENT_WRITABLE)
	STREAM_EVENT_ERROR    = StreamEventType(C.VIR_STREAM_EVENT_ERROR)
	STREAM_EVENT_HANGUP   = StreamEventType(C.VIR_STREAM_EVENT_HANGUP)
)

type StreamRecvFlagsValues int

const (
	STREAM_RECV_STOP_AT_HOLE = StreamRecvFlagsValues(C.VIR_STREAM_RECV_STOP_AT_HOLE)
)

type Stream struct {
	ptr C.virStreamPtr
}

func (v *Stream) Abort() error {
	result := C.virStreamAbort(v.ptr)
	if result == -1 {
		return GetLastError()
	}

	return nil
}

func (v *Stream) Finish() error {
	result := C.virStreamFinish(v.ptr)
	if result == -1 {
		return GetLastError()
	}

	return nil
}

func (v *Stream) Free() error {
	ret := C.virStreamFree(v.ptr)
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (c *Stream) Ref() error {
	ret := C.virStreamRef(c.ptr)
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (v *Stream) Recv(p []byte) (int, error) {
	n := C.virStreamRecv(v.ptr, (*C.char)(unsafe.Pointer(&p[0])), C.size_t(len(p)))
	if n < 0 {
		return 0, GetLastError()
	}
	if n == 0 {
		return 0, io.EOF
	}

	return int(n), nil
}

func (v *Stream) RecvFlags(p []byte, flags StreamRecvFlagsValues) (int, error) {
	if C.LIBVIR_VERSION_NUMBER < 3004000 {
		return 0, GetNotImplementedError("virStreamRecvFlags")
	}

	n := C.virStreamRecvFlagsCompat(v.ptr, (*C.char)(unsafe.Pointer(&p[0])), C.size_t(len(p)), C.uint(flags))
	if n < 0 {
		return 0, GetLastError()
	}
	if n == 0 {
		return 0, io.EOF
	}

	return int(n), nil
}

func (v *Stream) RecvHole(flags uint) (int64, error) {
	if C.LIBVIR_VERSION_NUMBER < 3004000 {
		return 0, GetNotImplementedError("virStreamSparseRecvHole")
	}

	var len C.longlong
	ret := C.virStreamRecvHoleCompat(v.ptr, &len, C.uint(flags))
	if ret < 0 {
		return 0, GetLastError()
	}

	return int64(len), nil
}

func (v *Stream) Send(p []byte) (int, error) {
	n := C.virStreamSend(v.ptr, (*C.char)(unsafe.Pointer(&p[0])), C.size_t(len(p)))
	if n < 0 {
		return 0, GetLastError()
	}
	if n == 0 {
		return 0, io.EOF
	}

	return int(n), nil
}

func (v *Stream) SendHole(len int64, flags uint32) error {
	if C.LIBVIR_VERSION_NUMBER < 3004000 {
		return GetNotImplementedError("virStreamSendHole")
	}

	ret := C.virStreamSendHoleCompat(v.ptr, C.longlong(len), C.uint(flags))
	if ret < 0 {
		return GetLastError()
	}

	return nil
}

type StreamSinkFunc func(*Stream, []byte) (int, error)
type StreamSinkHoleFunc func(*Stream, int64) error

//export streamSinkCallback
func streamSinkCallback(stream C.virStreamPtr, cdata *C.char, nbytes C.size_t, callbackID int) int {
	callbackFunc := getCallbackId(callbackID)

	callback, ok := callbackFunc.(StreamSinkFunc)
	if !ok {
		panic("Incorrect stream sink func callback")
	}

	data := make([]byte, int(nbytes))
	for i := 0; i < int(nbytes); i++ {
		cdatabyte := (*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cdata)) + (unsafe.Sizeof(*cdata) * uintptr(i))))
		data[i] = (byte)(*cdatabyte)
	}

	retnbytes, err := callback(&Stream{ptr: stream}, data)
	if err != nil {
		return -1
	}

	return retnbytes
}

//export streamSinkHoleCallback
func streamSinkHoleCallback(stream C.virStreamPtr, length C.longlong, callbackID int) int {
	callbackFunc := getCallbackId(callbackID)

	callback, ok := callbackFunc.(StreamSinkHoleFunc)
	if !ok {
		panic("Incorrect stream sink hole func callback")
	}

	err := callback(&Stream{ptr: stream}, int64(length))
	if err != nil {
		return -1
	}

	return 0
}

func (v *Stream) RecvAll(handler StreamSinkFunc) error {

	callbackID := registerCallbackId(handler)

	ret := C.virStreamRecvAll_cgo(v.ptr, (C.int)(callbackID))
	freeCallbackId(callbackID)
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (v *Stream) SparseRecvAll(handler StreamSinkFunc, holeHandler StreamSinkHoleFunc) error {
	if C.LIBVIR_VERSION_NUMBER < 3004000 {
		return GetNotImplementedError("virStreamSparseSendAll")
	}

	callbackID := registerCallbackId(handler)
	holeCallbackID := registerCallbackId(holeHandler)

	ret := C.virStreamSparseRecvAll_cgo(v.ptr, (C.int)(callbackID), (C.int)(holeCallbackID))
	freeCallbackId(callbackID)
	freeCallbackId(holeCallbackID)
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

type StreamSourceFunc func(*Stream, int) ([]byte, error)
type StreamSourceHoleFunc func(*Stream) (bool, int64, error)
type StreamSourceSkipFunc func(*Stream, int64) error

//export streamSourceCallback
func streamSourceCallback(stream C.virStreamPtr, cdata *C.char, nbytes C.size_t, callbackID int) int {
	callbackFunc := getCallbackId(callbackID)

	callback, ok := callbackFunc.(StreamSourceFunc)
	if !ok {
		panic("Incorrect stream sink func callback")
	}

	data, err := callback(&Stream{ptr: stream}, (int)(nbytes))
	if err != nil {
		return -1
	}

	nretbytes := int(nbytes)
	if len(data) < nretbytes {
		nretbytes = len(data)
	}

	for i := 0; i < nretbytes; i++ {
		cdatabyte := (*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cdata)) + (unsafe.Sizeof(*cdata) * uintptr(i))))
		*cdatabyte = (C.char)(data[i])
	}

	return nretbytes
}

//export streamSourceHoleCallback
func streamSourceHoleCallback(stream C.virStreamPtr, cinData *C.int, clength *C.longlong, callbackID int) int {
	callbackFunc := getCallbackId(callbackID)

	callback, ok := callbackFunc.(StreamSourceHoleFunc)
	if !ok {
		panic("Incorrect stream sink hole func callback")
	}

	inData, length, err := callback(&Stream{ptr: stream})
	if err != nil {
		return -1
	}

	if inData {
		*cinData = 1
	} else {
		*cinData = 0
	}
	*clength = C.longlong(length)

	return 0
}

//export streamSourceSkipCallback
func streamSourceSkipCallback(stream C.virStreamPtr, length C.longlong, callbackID int) int {
	callbackFunc := getCallbackId(callbackID)

	callback, ok := callbackFunc.(StreamSourceSkipFunc)
	if !ok {
		panic("Incorrect stream sink skip func callback")
	}

	err := callback(&Stream{ptr: stream}, int64(length))
	if err != nil {
		return -1
	}

	return 0
}

func (v *Stream) SendAll(handler StreamSourceFunc) error {

	callbackID := registerCallbackId(handler)

	ret := C.virStreamSendAll_cgo(v.ptr, (C.int)(callbackID))
	freeCallbackId(callbackID)
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (v *Stream) SparseSendAll(handler StreamSourceFunc, holeHandler StreamSourceHoleFunc, skipHandler StreamSourceSkipFunc) error {
	if C.LIBVIR_VERSION_NUMBER < 3004000 {
		return GetNotImplementedError("virStreamSparseSendAll")
	}

	callbackID := registerCallbackId(handler)
	holeCallbackID := registerCallbackId(holeHandler)
	skipCallbackID := registerCallbackId(skipHandler)

	ret := C.virStreamSparseSendAll_cgo(v.ptr, (C.int)(callbackID), (C.int)(holeCallbackID), (C.int)(skipCallbackID))
	freeCallbackId(callbackID)
	freeCallbackId(holeCallbackID)
	freeCallbackId(skipCallbackID)
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

type StreamEventCallback func(*Stream, StreamEventType)

func (v *Stream) EventAddCallback(events StreamEventType, callback StreamEventCallback) error {
	callbackID := registerCallbackId(callback)

	ret := C.virStreamEventAddCallback_cgo(v.ptr, (C.int)(events), (C.int)(callbackID))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

//export streamEventCallback
func streamEventCallback(st C.virStreamPtr, events int, callbackID int) {
	callbackFunc := getCallbackId(callbackID)

	callback, ok := callbackFunc.(StreamEventCallback)
	if !ok {
		panic("Incorrect stream event func callback")
	}

	callback(&Stream{ptr: st}, StreamEventType(events))
}

func (v *Stream) EventUpdateCallback(events StreamEventType) error {
	ret := C.virStreamEventUpdateCallback(v.ptr, (C.int)(events))
	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (v *Stream) EventRemoveCallback() error {
	ret := C.virStreamEventRemoveCallback(v.ptr)
	if ret == -1 {
		return GetLastError()
	}

	return nil
}
