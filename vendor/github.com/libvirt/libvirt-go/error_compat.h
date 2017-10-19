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

#ifndef LIBVIRT_GO_ERROR_COMPAT_H__
#define LIBVIRT_GO_ERROR_COMPAT_H__

/* 1.2.2 */

#ifndef VIR_FROM_BHYVE
#define VIR_FROM_BHYVE 57
#endif


/* 1.2.3 */

#ifndef VIR_FROM_CRYPTO
#define VIR_FROM_CRYPTO 58
#endif


/* 1.2.4 */

#ifndef VIR_FROM_FIREWALL
#define VIR_FROM_FIREWALL 59
#endif


/* 1.2.6 */

#ifndef VIR_ERR_CPU_INCOMPATIBLE
#define VIR_ERR_CPU_INCOMPATIBLE 91
#endif


/* 1.2.9 */

#ifndef VIR_FROM_POLKIT
#define VIR_FROM_POLKIT 60
#endif


/* 1.2.12 */

#ifndef VIR_ERR_XML_INVALID_SCHEMA
#define VIR_ERR_XML_INVALID_SCHEMA 92
#endif


/* 1.2.14 */

#ifndef VIR_FROM_THREAD
#define VIR_FROM_THREAD 61
#endif


/* 1.2.17 */

#ifndef VIR_FROM_ADMIN
#define VIR_FROM_ADMIN 62
#endif


/* 1.2.18 */

#ifndef VIR_ERR_MIGRATE_FINISH_OK
#define VIR_ERR_MIGRATE_FINISH_OK 93
#endif


/* 1.3.0 */

#ifndef VIR_FROM_LOGGING
#define VIR_FROM_LOGGING 63
#endif

/* 1.3.2 */

#ifndef VIR_FROM_XENXL
#define VIR_FROM_XENXL 64
#endif


/* 1.3.3 */

#ifndef VIR_FROM_PERF
#define VIR_FROM_PERF 65
#endif

#ifndef VIR_ERR_AUTH_UNAVAILABLE
#define VIR_ERR_AUTH_UNAVAILABLE 94
#endif

#ifndef VIR_ERR_NO_SERVER
#define VIR_ERR_NO_SERVER 95
#endif


/* 1.3.5 */

#ifndef VIR_ERR_NO_CLIENT
#define VIR_ERR_NO_CLIENT 96
#endif


/* 2.3.0 */

#ifndef VIR_ERR_AGENT_UNSYNCED
#define VIR_ERR_AGENT_UNSYNCED 97
#endif

/* 2.5.0 */

#ifndef VIR_ERR_LIBSSH
#define VIR_ERR_LIBSSH 98
#endif

#ifndef VIR_FROM_LIBSSH
#define VIR_FROM_LIBSSH 66
#endif

#endif /* LIBVIRT_GO_ERROR_COMPAT_H__ */
