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

#ifndef LIBVIRT_GO_DOMAIN_EVENTS_CFUNCS_H__
#define LIBVIRT_GO_DOMAIN_EVENTS_CFUNCS_H__

void domainEventLifecycleCallback_cgo(virConnectPtr c, virDomainPtr d,
                                     int event, int detail, void* data);

void domainEventGenericCallback_cgo(virConnectPtr c, virDomainPtr d, void* data);

void domainEventRTCChangeCallback_cgo(virConnectPtr c, virDomainPtr d,
                                     long long utcoffset, void* data);

void domainEventWatchdogCallback_cgo(virConnectPtr c, virDomainPtr d,
                                    int action, void* data);

void domainEventIOErrorCallback_cgo(virConnectPtr c, virDomainPtr d,
                                   const char *srcPath, const char *devAlias,
                                   int action, void* data);

void domainEventGraphicsCallback_cgo(virConnectPtr c, virDomainPtr d,
                                    int phase, const virDomainEventGraphicsAddress *local,
                                    const virDomainEventGraphicsAddress *remote,
                                    const char *authScheme,
                                    const virDomainEventGraphicsSubject *subject, void* data);

void domainEventIOErrorReasonCallback_cgo(virConnectPtr c, virDomainPtr d,
                                         const char *srcPath, const char *devAlias,
                                         int action, const char *reason, void* data);

void domainEventBlockJobCallback_cgo(virConnectPtr c, virDomainPtr d,
                                    const char *disk, int type, int status, void* data);

void domainEventDiskChangeCallback_cgo(virConnectPtr c, virDomainPtr d,
                                      const char *oldSrcPath, const char *newSrcPath,
                                      const char *devAlias, int reason, void* data);

void domainEventTrayChangeCallback_cgo(virConnectPtr c, virDomainPtr d,
                                      const char *devAlias, int reason, void* data);

void domainEventPMSuspendCallback_cgo(virConnectPtr c, virDomainPtr d,
				      int reason, void* data);

void domainEventPMWakeupCallback_cgo(virConnectPtr c, virDomainPtr d,
				     int reason, void* data);

void domainEventPMSuspendDiskCallback_cgo(virConnectPtr c, virDomainPtr d,
					  int reason, void* data);

void domainEventBalloonChangeCallback_cgo(virConnectPtr c, virDomainPtr d,
                                         unsigned long long actual, void* data);

void domainEventDeviceRemovedCallback_cgo(virConnectPtr c, virDomainPtr d,
                                         const char *devAlias, void* data);
void domainEventTunableCallback_cgo(virConnectPtr conn,
				    virDomainPtr dom,
				    virTypedParameterPtr params,
				    int nparams,
				    void *opaque);
void domainEventAgentLifecycleCallback_cgo(virConnectPtr conn,
					   virDomainPtr dom,
					   int state,
					   int reason,
					   void *opaque);
void domainEventDeviceAddedCallback_cgo(virConnectPtr conn,
					virDomainPtr dom,
					const char *devAlias,
					void *opaque);
void domainEventMigrationIterationCallback_cgo(virConnectPtr conn,
					       virDomainPtr dom,
					       int iteration,
					       void *opaque);
void domainEventJobCompletedCallback_cgo(virConnectPtr conn,
					 virDomainPtr dom,
					 virTypedParameterPtr params,
					 int nparams,
					 void *opaque);
void domainEventDeviceRemovalFailedCallback_cgo(virConnectPtr conn,
						virDomainPtr dom,
						const char *devAlias,
						void *opaque);
void domainEventMetadataChangeCallback_cgo(virConnectPtr conn,
					   virDomainPtr dom,
					   int type,
					   const char *nsuri,
					   void *opaque);
void domainEventBlockThresholdCallback_cgo(virConnectPtr conn,
					   virDomainPtr dom,
					   const char *dev,
					   const char *path,
					   unsigned long long threshold,
					   unsigned long long excess,
					   void *opaque);

int virConnectDomainEventRegisterAny_cgo(virConnectPtr c,  virDomainPtr d,
                                         int eventID, virConnectDomainEventGenericCallback cb,
                                         long goCallbackId);


#endif /* LIBVIRT_GO_DOMAIN_EVENTS_CFUNCS_H__ */
