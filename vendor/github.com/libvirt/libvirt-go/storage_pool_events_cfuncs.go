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
#include <assert.h>
#include "storage_pool_compat.h"
#include "storage_pool_events_cfuncs.h"
#include "callbacks_cfuncs.h"
#include <stdint.h>

extern void storagePoolEventLifecycleCallback(virConnectPtr, virStoragePoolPtr, int, int, int);
void storagePoolEventLifecycleCallback_cgo(virConnectPtr c, virStoragePoolPtr d,
                                           int event, int detail, void *data)
{
    storagePoolEventLifecycleCallback(c, d, event, detail, (int)(intptr_t)data);
}

extern void storagePoolEventGenericCallback(virConnectPtr, virStoragePoolPtr, int);
void storagePoolEventGenericCallback_cgo(virConnectPtr c, virStoragePoolPtr d,
                                         void *data)
{
    storagePoolEventGenericCallback(c, d, (int)(intptr_t)data);
}

int virConnectStoragePoolEventRegisterAny_cgo(virConnectPtr c,  virStoragePoolPtr d,
                                              int eventID, virConnectStoragePoolEventGenericCallback cb,
                                              long goCallbackId) {
#if LIBVIR_VERSION_NUMBER < 2000000
    assert(0); // Caller should have checked version
#else
    void* id = (void*)goCallbackId;
    return virConnectStoragePoolEventRegisterAny(c, d, eventID, cb, id, freeGoCallback_cgo);
#endif
}

*/
import "C"
