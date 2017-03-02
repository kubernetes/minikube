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
#include "events_cfuncs.h"

void eventHandleCallback(int watch, int fd, int events, int callbackID);

static void eventAddHandleHelper(int watch, int fd, int events, void *opaque)
{
    eventHandleCallback(watch, fd, events, (int)(intptr_t)opaque);
}

int virEventAddHandle_cgo(int fd, int events, int callbackID)
{
    return virEventAddHandle(fd, events, eventAddHandleHelper, (void *)(intptr_t)callbackID, NULL);
}

void eventTimeoutCallback(int timer, int callbackID);

static void eventAddTimeoutHelper(int timer, void *opaque)
{
    eventTimeoutCallback(timer, (int)(intptr_t)opaque);
}

int virEventAddTimeout_cgo(int freq, int callbackID)
{
    return virEventAddTimeout(freq, eventAddTimeoutHelper, (void *)(intptr_t)callbackID, NULL);
}

int eventAddHandleFunc(int fd, int event, uintptr_t callback, uintptr_t opaque, uintptr_t freecb);
void eventUpdateHandleFunc(int watch, int event);
int eventRemoveHandleFunc(int watch);
int eventAddTimeoutFunc(int freq, uintptr_t callback, uintptr_t opaque, uintptr_t freecb);
void eventUpdateTimeoutFunc(int timer, int freq);
int eventRemoveTimeoutFunc(int timer);

int eventAddHandleFuncHelper(int fd, int event, virEventHandleCallback callback, void *opaque, virFreeCallback freecb)
{
    return eventAddHandleFunc(fd, event, (uintptr_t)callback, (uintptr_t)opaque, (uintptr_t)freecb);
}

void eventUpdateHandleFuncHelper(int watch, int event)
{
    eventUpdateHandleFunc(watch, event);
}

int eventRemoveHandleFuncHelper(int watch)
{
    return eventRemoveHandleFunc(watch);
}

int eventAddTimeoutFuncHelper(int freq, virEventTimeoutCallback callback, void *opaque, virFreeCallback freecb)
{
    return eventAddTimeoutFunc(freq, (uintptr_t)callback, (uintptr_t)opaque, (uintptr_t)freecb);
}

void eventUpdateTimeoutFuncHelper(int timer, int freq)
{
    eventUpdateTimeoutFunc(timer, freq);
}

int eventRemoveTimeoutFuncHelper(int timer)
{
    return eventRemoveTimeoutFunc(timer);
}


void virEventRegisterImpl_cgo(void)
{
    virEventRegisterImpl(eventAddHandleFuncHelper,
                         eventUpdateHandleFuncHelper,
                         eventRemoveHandleFuncHelper,
                         eventAddTimeoutFuncHelper,
                         eventUpdateTimeoutFuncHelper,
                         eventRemoveTimeoutFuncHelper);
}

void eventHandleCallbackInvoke(int watch, int fd, int events, uintptr_t callback, uintptr_t opaque)
{
    ((virEventHandleCallback)callback)(watch, fd, events, (void *)opaque);
}

void eventTimeoutCallbackInvoke(int timer, uintptr_t callback, uintptr_t opaque)
{
    ((virEventTimeoutCallback)callback)(timer, (void *)opaque);
}


void eventHandleCallbackFree(uintptr_t callback, uintptr_t opaque)
{
    ((virFreeCallback)callback)((void *)opaque);
}

void eventTimeoutCallbackFree(uintptr_t callback, uintptr_t opaque)
{
    ((virFreeCallback)callback)((void *)opaque);
}


*/
import "C"
