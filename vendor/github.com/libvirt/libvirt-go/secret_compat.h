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

#ifndef LIBVIRT_GO_SECRET_COMPAT_H__
#define LIBVIRT_GO_SECRET_COMPAT_H__

/* 3.0.0 */

#ifndef VIR_SECRET_EVENT_DEFINED
#define VIR_SECRET_EVENT_DEFINED 0
#endif

#ifndef VIR_SECRET_EVENT_UNDEFINED
#define VIR_SECRET_EVENT_UNDEFINED 1
#endif

#ifndef VIR_SECRET_EVENT_ID_LIFECYCLE
#define VIR_SECRET_EVENT_ID_LIFECYCLE 0
#endif

#ifndef VIR_SECRET_EVENT_ID_VALUE_CHANGED
#define VIR_SECRET_EVENT_ID_VALUE_CHANGED 1
#endif


#if LIBVIR_VERSION_NUMBER < 3000000
typedef void (*virConnectSecretEventGenericCallback)(virConnectPtr conn,
						     virSecretPtr secret,
						     void *opaque);
#endif

int virConnectSecretEventDeregisterAnyCompat(virConnectPtr conn,
					     int callbackID);

/* 2.2.1 */

#ifndef VIR_SECRET_USAGE_TYPE_TLS
#define VIR_SECRET_USAGE_TYPE_TLS 4
#endif

#endif /* LIBVIRT_GO_SECRET_COMPAT_H__ */
