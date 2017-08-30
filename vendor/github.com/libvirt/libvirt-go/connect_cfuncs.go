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
#include "connect_cfuncs.h"
#include "callbacks_cfuncs.h"

extern void closeCallback(virConnectPtr, int, long);
void closeCallback_cgo(virConnectPtr conn, int reason, void *opaque)
{
    closeCallback(conn, reason, (long)opaque);
}

int virConnectRegisterCloseCallback_cgo(virConnectPtr c, virConnectCloseFunc cb, long goCallbackId)
{
    void *id = (void*)goCallbackId;
    return virConnectRegisterCloseCallback(c, cb, id, freeGoCallback_cgo);
}

#include <stdio.h>

extern int connectAuthCallback(virConnectCredentialPtr, unsigned int, int);
int connectAuthCallback_cgo(virConnectCredentialPtr cred, unsigned int ncred, void *cbdata)
{
    int *callbackID = cbdata;

    return connectAuthCallback(cred, ncred, *callbackID);
}

virConnectPtr virConnectOpenAuthWrap(const char *name, int *credtype, uint ncredtype, int callbackID, unsigned int flags)
{
    virConnectAuth auth = {
       .credtype = credtype,
       .ncredtype = ncredtype,
       .cb = connectAuthCallback_cgo,
       .cbdata = &callbackID,
    };

    return virConnectOpenAuth(name, &auth, flags);
}

*/
import "C"
