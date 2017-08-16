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
#include <assert.h>
#include "domain_compat.h"

int virDomainCoreDumpWithFormatCompat(virDomainPtr domain,
				      const char *to,
				      unsigned int dumpformat,
				      unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1002003
    assert(0); // Caller should have checked version
#else
    return virDomainCoreDumpWithFormat(domain, to, dumpformat, flags);
#endif
}


int virDomainGetTimeCompat(virDomainPtr dom,
			   long long *seconds,
			   unsigned int *nseconds,
			   unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1002005
    assert(0); // Caller should have checked version
#else
    return virDomainGetTime(dom, seconds, nseconds, flags);
#endif
}

int virDomainSetTimeCompat(virDomainPtr dom,
			   long long seconds,
			   unsigned int nseconds,
			   unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1002005
    assert(0); // Caller should have checked version
#else
    return virDomainSetTime(dom, seconds, nseconds, flags);
#endif
}

int virDomainFSFreezeCompat(virDomainPtr dom,
			    const char **mountpoints,
			    unsigned int nmountpoints,
			    unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1002005
    assert(0); // Caller should have checked version
#else
    return virDomainFSFreeze(dom, mountpoints, nmountpoints, flags);
#endif
}

int virDomainFSThawCompat(virDomainPtr dom,
			  const char **mountpoints,
			  unsigned int nmountpoints,
			  unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1002005
    assert(0); // Caller should have checked version
#else
    return virDomainFSThaw(dom, mountpoints, nmountpoints, flags);
#endif
}

int virDomainBlockCopyCompat(virDomainPtr dom, const char *disk,
			     const char *destxml,
			     virTypedParameterPtr params,
			     int nparams,
			     unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1002008
    assert(0); // Caller should have checked version
#else
    return virDomainBlockCopy(dom, disk, destxml, params, nparams, flags);
#endif
}

int virDomainOpenGraphicsFDCompat(virDomainPtr dom,
				  unsigned int idx,
				  unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1002008
    assert(0); // Caller should have checked version
#else
    return virDomainOpenGraphicsFD(dom, idx, flags);
#endif
}

void virDomainFSInfoFreeCompat(virDomainFSInfoPtr info)
{
}

int virDomainGetFSInfoCompat(virDomainPtr dom,
			     virDomainFSInfoPtr **info,
			     unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1002011
    assert(0); // Caller should have checked version
#else
    return virDomainGetFSInfo(dom, info, flags);
#endif
}

int virDomainInterfaceAddressesCompat(virDomainPtr dom,
				      virDomainInterfacePtr **ifaces,
				      unsigned int source,
				      unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1002014
    assert(0); // Caller should have checked version
#else
    return virDomainInterfaceAddresses(dom, ifaces, source, flags);
#endif
}

void virDomainInterfaceFreeCompat(virDomainInterfacePtr iface)
{
}

void virDomainIOThreadInfoFreeCompat(virDomainIOThreadInfoPtr info)
{
}

int virDomainGetIOThreadInfoCompat(virDomainPtr domain,
				   virDomainIOThreadInfoPtr **info,
				   unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1002014
    assert(0); // Caller should have checked version
#else
    return virDomainGetIOThreadInfo(domain, info, flags);
#endif
}
int virDomainPinIOThreadCompat(virDomainPtr domain,
			       unsigned int iothread_id,
			       unsigned char *cpumap,
			       int maplen,
			       unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1002014
    assert(0); // Caller should have checked version
#else
    return virDomainPinIOThread(domain, iothread_id, cpumap, maplen, flags);
#endif
}

int virDomainAddIOThreadCompat(virDomainPtr domain,
			       unsigned int iothread_id,
			       unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1002015
    assert(0); // Caller should have checked version
#else
    return virDomainAddIOThread(domain, iothread_id, flags);
#endif
}


int virDomainDelIOThreadCompat(virDomainPtr domain,
			       unsigned int iothread_id,
			       unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1002015
    assert(0); // Caller should have checked version
#else
    return virDomainDelIOThread(domain, iothread_id, flags);
#endif
}


int virDomainSetUserPasswordCompat(virDomainPtr dom,
				   const char *user,
				   const char *password,
				   unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1002016
    assert(0); // Caller should have checked version
#else
    return virDomainSetUserPassword(dom, user, password, flags);
#endif
}


int virDomainRenameCompat(virDomainPtr dom,
			  const char *new_name,
			  unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1002019
    assert(0); // Caller should have checked version
#else
    return virDomainRename(dom, new_name, flags);
#endif
}


int virDomainGetPerfEventsCompat(virDomainPtr dom,
				 virTypedParameterPtr *params,
				 int *nparams,
				 unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1003003
    assert(0); // Caller should have checked version
#else
    return virDomainGetPerfEvents(dom, params, nparams, flags);
#endif
}


int virDomainSetPerfEventsCompat(virDomainPtr dom,
				 virTypedParameterPtr params,
				 int nparams,
				 unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1003003
    assert(0); // Caller should have checked version
#else
    return virDomainSetPerfEvents(dom, params, nparams, flags);
#endif
}


int virDomainMigrateStartPostCopyCompat(virDomainPtr domain,
					unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1003003
    assert(0); // Caller should have checked version
#else
    return virDomainMigrateStartPostCopy(domain, flags);
#endif
}


int virDomainGetGuestVcpusCompat(virDomainPtr domain,
				 virTypedParameterPtr *params,
				 unsigned int *nparams,
				 unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 2000000
    assert(0); // Caller should have checked version
#else
    return virDomainGetGuestVcpus(domain, params, nparams, flags);
#endif
}


int virDomainSetGuestVcpusCompat(virDomainPtr domain,
				 const char *cpumap,
				 int state,
				 unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 2000000
    assert(0); // Caller should have checked version
#else
    return virDomainSetGuestVcpus(domain, cpumap, state, flags);
#endif
}

int virDomainSetVcpuCompat(virDomainPtr domain,
			   const char *cpumap,
			   int state,
			   unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 3001000
    assert(0); // Caller should have checked version
#else
    return virDomainSetVcpu(domain, cpumap, state, flags);
#endif
}


int virDomainSetBlockThresholdCompat(virDomainPtr domain,
                                     const char *dev,
                                     unsigned long long threshold,
                                     unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 3002000
    assert(0); // Caller should have checked version
#else
    return virDomainSetBlockThreshold(domain, dev, threshold, flags);
#endif
}

*/
import "C"
