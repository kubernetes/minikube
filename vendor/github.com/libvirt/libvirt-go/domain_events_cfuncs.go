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
#include "domain_events_cfuncs.h"
#include "callbacks_cfuncs.h"
#include <stdint.h>

extern void domainEventLifecycleCallback(virConnectPtr, virDomainPtr, int, int, int);
void domainEventLifecycleCallback_cgo(virConnectPtr c, virDomainPtr d,
                                     int event, int detail, void *data)
{
    domainEventLifecycleCallback(c, d, event, detail, (int)(intptr_t)data);
}

extern void domainEventGenericCallback(virConnectPtr, virDomainPtr, int);
void domainEventGenericCallback_cgo(virConnectPtr c, virDomainPtr d, void *data)
{
    domainEventGenericCallback(c, d, (int)(intptr_t)data);
}

extern void domainEventRTCChangeCallback(virConnectPtr, virDomainPtr, long long, int);
void domainEventRTCChangeCallback_cgo(virConnectPtr c, virDomainPtr d,
                                     long long utcoffset, void *data)
{
    domainEventRTCChangeCallback(c, d, utcoffset, (int)(intptr_t)data);
}

extern void domainEventWatchdogCallback(virConnectPtr, virDomainPtr, int, int);
void domainEventWatchdogCallback_cgo(virConnectPtr c, virDomainPtr d,
                                    int action, void *data)
{
    domainEventWatchdogCallback(c, d, action, (int)(intptr_t)data);
}

extern void domainEventIOErrorCallback(virConnectPtr, virDomainPtr, const char *, const char *, int, int);
void domainEventIOErrorCallback_cgo(virConnectPtr c, virDomainPtr d,
                                   const char *srcPath, const char *devAlias,
                                   int action, void *data)
{
    domainEventIOErrorCallback(c, d, srcPath, devAlias, action, (int)(intptr_t)data);
}

extern void domainEventGraphicsCallback(virConnectPtr, virDomainPtr, int, const virDomainEventGraphicsAddress *,
                                        const virDomainEventGraphicsAddress *, const char *,
                                        const virDomainEventGraphicsSubject *, int);
void domainEventGraphicsCallback_cgo(virConnectPtr c, virDomainPtr d,
                                    int phase, const virDomainEventGraphicsAddress *local,
                                    const virDomainEventGraphicsAddress *remote,
                                    const char *authScheme,
                                    const virDomainEventGraphicsSubject *subject, void *data)
{
    domainEventGraphicsCallback(c, d, phase, local, remote, authScheme, subject, (int)(intptr_t)data);
}

extern void domainEventIOErrorReasonCallback(virConnectPtr, virDomainPtr, const char *, const char *,
                                             int, const char *, int);
void domainEventIOErrorReasonCallback_cgo(virConnectPtr c, virDomainPtr d,
                                         const char *srcPath, const char *devAlias,
                                         int action, const char *reason, void *data)
{
    domainEventIOErrorReasonCallback(c, d, srcPath, devAlias, action, reason, (int)(intptr_t)data);
}

extern void domainEventBlockJobCallback(virConnectPtr, virDomainPtr, const char *, int, int, int);
void domainEventBlockJobCallback_cgo(virConnectPtr c, virDomainPtr d,
                                    const char *disk, int type, int status, void *data)
{
    domainEventBlockJobCallback(c, d, disk, type, status, (int)(intptr_t)data);
}

extern void domainEventDiskChangeCallback(virConnectPtr, virDomainPtr, const char *, const char *,
                                          const char *, int, int);
void domainEventDiskChangeCallback_cgo(virConnectPtr c, virDomainPtr d,
                                      const char *oldSrcPath, const char *newSrcPath,
                                      const char *devAlias, int reason, void *data)
{
    domainEventDiskChangeCallback(c, d, oldSrcPath, newSrcPath, devAlias, reason, (int)(intptr_t)data);
}

extern void domainEventTrayChangeCallback(virConnectPtr, virDomainPtr, const char *, int, int);
void domainEventTrayChangeCallback_cgo(virConnectPtr c, virDomainPtr d,
                                      const char *devAlias, int reason, void *data)
{
    domainEventTrayChangeCallback(c, d, devAlias, reason, (int)(intptr_t)data);
}

extern void domainEventPMSuspendCallback(virConnectPtr, virDomainPtr, int, int);
void domainEventPMSuspendCallback_cgo(virConnectPtr c, virDomainPtr d,
                                  int reason, void *data)
{
    domainEventPMSuspendCallback(c, d, reason, (int)(intptr_t)data);
}

extern void domainEventPMWakeupCallback(virConnectPtr, virDomainPtr, int, int);
void domainEventPMWakeupCallback_cgo(virConnectPtr c, virDomainPtr d,
                                     int reason, void *data)
{
    domainEventPMWakeupCallback(c, d, reason, (int)(intptr_t)data);
}

extern void domainEventPMSuspendDiskCallback(virConnectPtr, virDomainPtr, int, int);
void domainEventPMSuspendDiskCallback_cgo(virConnectPtr c, virDomainPtr d,
                                          int reason, void *data)
{
    domainEventPMSuspendDiskCallback(c, d, reason, (int)(intptr_t)data);
}

extern void domainEventBalloonChangeCallback(virConnectPtr, virDomainPtr, unsigned long long, int);
void domainEventBalloonChangeCallback_cgo(virConnectPtr c, virDomainPtr d,
                                         unsigned long long actual, void *data)
{
    domainEventBalloonChangeCallback(c, d, actual, (int)(intptr_t)data);
}

extern void domainEventDeviceRemovedCallback(virConnectPtr, virDomainPtr, const char *, int);
void domainEventDeviceRemovedCallback_cgo(virConnectPtr c, virDomainPtr d,
                                         const char *devAlias, void *data)
{
    domainEventDeviceRemovedCallback(c, d, devAlias, (int)(intptr_t)data);
}

extern void domainEventTunableCallback(virConnectPtr, virDomainPtr, virTypedParameterPtr, int, int);
void domainEventTunableCallback_cgo(virConnectPtr conn,
				    virDomainPtr dom,
				    virTypedParameterPtr params,
				    int nparams,
				    void *opaque)
{
    domainEventTunableCallback(conn, dom, params, nparams, (int)(intptr_t)opaque);
}

extern void domainEventAgentLifecycleCallback(virConnectPtr, virDomainPtr, int, int, int);
void domainEventAgentLifecycleCallback_cgo(virConnectPtr conn,
					   virDomainPtr dom,
					   int state,
					   int reason,
					   void *opaque)
{
    domainEventAgentLifecycleCallback(conn, dom, state, reason, (int)(intptr_t)opaque);
}

extern void domainEventDeviceAddedCallback(virConnectPtr, virDomainPtr, const char *, int);
void domainEventDeviceAddedCallback_cgo(virConnectPtr conn,
					virDomainPtr dom,
					const char *devAlias,
					void *opaque)
{
    domainEventDeviceAddedCallback(conn, dom, devAlias, (int)(intptr_t)opaque);
}

extern void domainEventMigrationIterationCallback(virConnectPtr, virDomainPtr, int, int);
void domainEventMigrationIterationCallback_cgo(virConnectPtr conn,
					       virDomainPtr dom,
					       int iteration,
					       void *opaque)
{
    domainEventMigrationIterationCallback(conn, dom, iteration, (int)(intptr_t)opaque);
}

extern void domainEventJobCompletedCallback(virConnectPtr, virDomainPtr, virTypedParameterPtr, int, int);
void domainEventJobCompletedCallback_cgo(virConnectPtr conn,
					 virDomainPtr dom,
					 virTypedParameterPtr params,
					 int nparams,
					 void *opaque)
{
    domainEventJobCompletedCallback(conn, dom, params, nparams, (int)(intptr_t)opaque);
}

extern void domainEventDeviceRemovalFailedCallback(virConnectPtr, virDomainPtr, const char *, int);
void domainEventDeviceRemovalFailedCallback_cgo(virConnectPtr conn,
						virDomainPtr dom,
						const char *devAlias,
						void *opaque)
{
    domainEventDeviceRemovalFailedCallback(conn, dom, devAlias, (int)(intptr_t)opaque);
}

extern void domainEventMetadataChangeCallback(virConnectPtr, virDomainPtr, int, const char *, int);
void domainEventMetadataChangeCallback_cgo(virConnectPtr conn,
					   virDomainPtr dom,
					   int type,
					   const char *nsuri,
					   void *opaque)
{
    domainEventMetadataChangeCallback(conn, dom, type, nsuri, (int)(intptr_t)opaque);
}

extern void domainEventBlockThresholdCallback(virConnectPtr, virDomainPtr, const char *, const char *, unsigned long long, unsigned long long, int);
void domainEventBlockThresholdCallback_cgo(virConnectPtr conn,
					   virDomainPtr dom,
					   const char *dev,
					   const char *path,
					   unsigned long long threshold,
					   unsigned long long excess,
					   void *opaque)
{
    domainEventBlockThresholdCallback(conn, dom, dev, path, threshold, excess, (int)(intptr_t)opaque);
}

int virConnectDomainEventRegisterAny_cgo(virConnectPtr c,  virDomainPtr d,
                                         int eventID, virConnectDomainEventGenericCallback cb,
                                         long goCallbackId) {
    void* id = (void*)goCallbackId;
    return virConnectDomainEventRegisterAny(c, d, eventID, cb, id, freeGoCallback_cgo);
}

*/
import "C"
