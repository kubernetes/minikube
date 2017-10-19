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
#include "qemu_compat.h"
#include "qemu_cfuncs.h"
#include "callbacks_cfuncs.h"
#include <assert.h>
#include <stdint.h>


extern void domainQemuMonitorEventCallback(virConnectPtr, virDomainPtr, const char *, long long, unsigned int, const char *, int);
void domainQemuMonitorEventCallback_cgo(virConnectPtr c, virDomainPtr d,
					const char *event, long long secs,
					unsigned int micros, const char *details, void *data)
{
    domainQemuMonitorEventCallback(c, d, event, secs, micros, details, (int)(intptr_t)data);
}

int virConnectDomainQemuMonitorEventRegister_cgo(virConnectPtr c,  virDomainPtr d,
                                                    const char *event, virConnectDomainQemuMonitorEventCallback cb,
                                                    long goCallbackId, unsigned int flags) {
#if LIBVIR_VERSION_NUMBER < 1002003
    assert(0); // Caller should have checked version
#else
    void* id = (void*)goCallbackId;
    return virConnectDomainQemuMonitorEventRegister(c, d, event, cb, id, freeGoCallback_cgo, flags);
#endif
}

*/
import "C"
