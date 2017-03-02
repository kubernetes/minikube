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
#include "connect_compat.h"

int virNodeGetFreePagesCompat(virConnectPtr conn,
			      unsigned int npages,
			      unsigned int *pages,
			      int startcell,
			      unsigned int cellcount,
			      unsigned long long *counts,
			      unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1002006
    assert(0); // Caller should have checked version
#else
    return virNodeGetFreePages(conn, npages, pages, startcell, cellcount, counts, flags);
#endif
}

char * virConnectGetDomainCapabilitiesCompat(virConnectPtr conn,
					     const char *emulatorbin,
					     const char *arch,
					     const char *machine,
					     const char *virttype,
					     unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1002007
    assert(0); // Caller should have checked version
#else
    return virConnectGetDomainCapabilities(conn, emulatorbin, arch, machine, virttype, flags);
#endif
}

int virConnectGetAllDomainStatsCompat(virConnectPtr conn,
				      unsigned int stats,
				      virDomainStatsRecordPtr **retStats,
				      unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1002008
    assert(0); // Caller should have checked version
#else
    return virConnectGetAllDomainStats(conn, stats, retStats, flags);
#endif
}

int virDomainListGetStatsCompat(virDomainPtr *doms,
				unsigned int stats,
				virDomainStatsRecordPtr **retStats,
				unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1002008
    assert(0); // Caller should have checked version
#else
    return virDomainListGetStats(doms, stats, retStats, flags);
#endif
}

void virDomainStatsRecordListFreeCompat(virDomainStatsRecordPtr *stats)
{
}

int virNodeAllocPagesCompat(virConnectPtr conn,
			    unsigned int npages,
			    unsigned int *pageSizes,
			    unsigned long long *pageCounts,
			    int startCell,
			    unsigned int cellCount,
			    unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1002009
    assert(0); // Caller should have checked version
#else
    return virNodeAllocPages(conn, npages, pageSizes, pageCounts, startCell, cellCount, flags);
#endif
}


virDomainPtr virDomainDefineXMLFlagsCompat(virConnectPtr conn,
					   const char *xml,
					   unsigned int flags)
{
#if LIBVIR_VERSION_NUMBER < 1002012
    assert(0); // Caller should have checked version
#else
    return virDomainDefineXMLFlags(conn, xml, flags);
#endif
}

*/
import "C"
