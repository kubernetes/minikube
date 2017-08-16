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
#include <stdint.h>
#include <stdlib.h>
#include <assert.h>
#include "stream_cfuncs.h"

int streamSourceCallback(virStreamPtr st, char *cdata, size_t nbytes, int callbackID);
int streamSourceHoleCallback(virStreamPtr st, int *inData, long long *length, int callbackID);
int streamSourceSkipCallback(virStreamPtr st, long long length, int callbackID);

int streamSinkCallback(virStreamPtr st, const char *cdata, size_t nbytes, int callbackID);
int streamSinkHoleCallback(virStreamPtr st, long long length, int callbackID);

struct CallbackData {
    int callbackID;
    int holeCallbackID;
    int skipCallbackID;
};

static int streamSourceCallbackHelper(virStreamPtr st, char *data, size_t nbytes, void *opaque)
{
    struct CallbackData *cbdata = opaque;

    return streamSourceCallback(st, data, nbytes, cbdata->callbackID);
}

static int streamSourceHoleCallbackHelper(virStreamPtr st, int *inData, long long *length, void *opaque)
{
    struct CallbackData *cbdata = opaque;

    return streamSourceHoleCallback(st, inData, length, cbdata->holeCallbackID);
}

static int streamSourceSkipCallbackHelper(virStreamPtr st, long long length, void *opaque)
{
    struct CallbackData *cbdata = opaque;

    return streamSourceSkipCallback(st, length, cbdata->skipCallbackID);
}

static int streamSinkCallbackHelper(virStreamPtr st, const char *data, size_t nbytes, void *opaque)
{
    struct CallbackData *cbdata = opaque;

    return streamSinkCallback(st, data, nbytes, cbdata->callbackID);
}

static int streamSinkHoleCallbackHelper(virStreamPtr st, long long length, void *opaque)
{
    struct CallbackData *cbdata = opaque;

    return streamSinkHoleCallback(st, length, cbdata->holeCallbackID);
}

int virStreamSendAll_cgo(virStreamPtr st, int callbackID)
{
    struct CallbackData cbdata = { .callbackID = callbackID };
    return virStreamSendAll(st, streamSourceCallbackHelper, &cbdata);
}

int virStreamSparseSendAll_cgo(virStreamPtr st, int callbackID, int holeCallbackID, int skipCallbackID)
{
    struct CallbackData cbdata = { .callbackID = callbackID, .holeCallbackID = holeCallbackID, .skipCallbackID = skipCallbackID };
#if LIBVIR_VERSION_NUMBER < 3004000
    assert(0); // Caller should have checked version
#else
    return virStreamSparseSendAll(st, streamSourceCallbackHelper, streamSourceHoleCallbackHelper, streamSourceSkipCallbackHelper, &cbdata);
#endif
}


int virStreamRecvAll_cgo(virStreamPtr st, int callbackID)
{
    struct CallbackData cbdata = { .callbackID = callbackID };
    return virStreamRecvAll(st, streamSinkCallbackHelper, &cbdata);
}

int virStreamSparseRecvAll_cgo(virStreamPtr st, int callbackID, int holeCallbackID)
{
    struct CallbackData cbdata = { .callbackID = callbackID, .holeCallbackID = holeCallbackID };
#if LIBVIR_VERSION_NUMBER < 3004000
    assert(0); // Caller should have checked version
#else
    return virStreamSparseRecvAll(st, streamSinkCallbackHelper, streamSinkHoleCallbackHelper, &cbdata);
#endif
}

void streamEventCallback(virStreamPtr st, int events, int callbackID);

static void streamEventCallbackHelper(virStreamPtr st, int events, void *opaque)
{
    streamEventCallback(st, events, (int)(intptr_t)opaque);
}

int virStreamEventAddCallback_cgo(virStreamPtr st, int events, int callbackID)
{
    return virStreamEventAddCallback(st, events, streamEventCallbackHelper, (void *)(intptr_t)callbackID, NULL);
}

*/
import "C"
