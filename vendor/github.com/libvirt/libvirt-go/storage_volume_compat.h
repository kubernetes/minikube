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

#ifndef LIBVIRT_GO_STORAGE_VOLUME_COMPAT_H__
#define LIBVIRT_GO_STORAGE_VOLUME_COMPAT_H__

/* 3.0.0 */

int virStorageVolGetInfoFlagsCompat(virStorageVolPtr vol,
				    virStorageVolInfoPtr info,
				    unsigned int flags);

#ifndef VIR_STORAGE_VOL_USE_ALLOCATION
#define VIR_STORAGE_VOL_USE_ALLOCATION 0
#endif

#ifndef VIR_STORAGE_VOL_GET_PHYSICAL
#define VIR_STORAGE_VOL_GET_PHYSICAL 1 << 0
#endif


/* 1.2.13 */

#ifndef VIR_STORAGE_VOL_CREATE_REFLINK
#define VIR_STORAGE_VOL_CREATE_REFLINK 1<< 1
#endif


/* 1.2.21 */

#ifndef VIR_STORAGE_VOL_DELETE_WITH_SNAPSHOTS
#define VIR_STORAGE_VOL_DELETE_WITH_SNAPSHOTS 1 << 1
#endif


/* 1.3.2 */

#ifndef VIR_STORAGE_VOL_WIPE_ALG_TRIM
#define VIR_STORAGE_VOL_WIPE_ALG_TRIM 9
#endif


/* 1.3.4 */

#ifndef VIR_STORAGE_VOL_PLOOP
#define VIR_STORAGE_VOL_PLOOP 5
#endif

/* 3.4.0 */

#ifndef VIR_STORAGE_VOL_UPLOAD_SPARSE_STREAM
#define VIR_STORAGE_VOL_UPLOAD_SPARSE_STREAM (1 << 0)
#endif

#ifndef VIR_STORAGE_VOL_DOWNLOAD_SPARSE_STREAM
#define VIR_STORAGE_VOL_DOWNLOAD_SPARSE_STREAM (1 << 0)
#endif

#endif /* LIBVIRT_GO_STORAGE_VOLUME_COMPAT_H__ */
