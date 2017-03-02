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

#ifndef LIBVIRT_GO_STORAGE_POOL_COMPAT_H__
#define LIBVIRT_GO_STORAGE_POOL_COMPAT_H__

/* 1.3.1 */

#ifndef VIR_STORAGE_POOL_CREATE_NORMAL
#define VIR_STORAGE_POOL_CREATE_NORMAL 0
#endif

#ifndef VIR_STORAGE_POOL_CREATE_WITH_BUILD
#define VIR_STORAGE_POOL_CREATE_WITH_BUILD 1 << 0
#endif

#ifndef VIR_STORAGE_POOL_CREATE_WITH_BUILD_OVERWRITE
#define VIR_STORAGE_POOL_CREATE_WITH_BUILD_OVERWRITE 1 << 1
#endif

#ifndef VIR_STORAGE_POOL_CREATE_WITH_BUILD_NO_OVERWRITE
#define VIR_STORAGE_POOL_CREATE_WITH_BUILD_NO_OVERWRITE 1 << 2
#endif


/* 2.0.0 */

#ifndef VIVIR_STORAGE_POOL_EVENT_DEFINED
#define VIR_STORAGE_POOL_EVENT_DEFINED 0
#endif

#ifndef VIR_STORAGE_POOL_EVENT_UNDEFINED
#define VIR_STORAGE_POOL_EVENT_UNDEFINED 1
#endif

#ifndef VIR_STORAGE_POOL_EVENT_STARTED
#define VIR_STORAGE_POOL_EVENT_STARTED 2
#endif

#ifndef VIR_STORAGE_POOL_EVENT_STOPPED
#define VIR_STORAGE_POOL_EVENT_STOPPED 3
#endif

#ifndef VIR_STORAGE_POOL_EVENT_ID_LIFECYCLE
#define VIR_STORAGE_POOL_EVENT_ID_LIFECYCLE 0
#endif

#ifndef VIR_STORAGE_POOL_EVENT_ID_REFRESH
#define VIR_STORAGE_POOL_EVENT_ID_REFRESH 1
#endif

#if LIBVIR_VERSION_NUMBER < 2000000
typedef void (*virConnectStoragePoolEventGenericCallback)(virConnectPtr conn,
                                                          virStoragePoolPtr pool,
                                                          void *opaque);
#endif

int virConnectStoragePoolEventDeregisterAnyCompat(virConnectPtr conn,
						  int callbackID);


#endif /* LIBVIRT_GO_STORAGE_POOL_COMPAT_H__ */
