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

#ifndef LIBVIRT_GO_QEMU_COMPAT_H__
#define LIBVIRT_GO_QEMU_COMPAT_H__

/* 1.2.3 */

#if LIBVIR_VERSION_NUMBER < 1002003
typedef void (*virConnectDomainQemuMonitorEventCallback)(virConnectPtr conn,
                                                         virDomainPtr dom,
                                                         const char *event,
                                                         long long seconds,
                                                         unsigned int micros,
                                                         const char *details,
                                                         void *opaque);
#endif

int virConnectDomainQemuMonitorEventDeregisterCompat(virConnectPtr conn,
						     int callbackID);

#ifndef VIR_CONNECT_DOMAIN_QEMU_MONITOR_EVENT_REGISTER_REGEX
#define VIR_CONNECT_DOMAIN_QEMU_MONITOR_EVENT_REGISTER_REGEX (1 << 0)
#endif

#ifndef VIR_CONNECT_DOMAIN_QEMU_MONITOR_EVENT_REGISTER_NOCASE
#define VIR_CONNECT_DOMAIN_QEMU_MONITOR_EVENT_REGISTER_NOCASE (1 << 1)
#endif

/* 1.2.15 */

#ifndef VIR_DOMAIN_QEMU_AGENT_COMMAND_SHUTDOWN
#define VIR_DOMAIN_QEMU_AGENT_COMMAND_SHUTDOWN 60
#endif


#endif /* LIBVIRT_GO_QEMU_COMPAT_H__ */
