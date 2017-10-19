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

#ifndef LIBVIRT_GO_EVENTS_CFUNCS_H__
#define LIBVIRT_GO_EVENTS_CFUNCS_H__

int virEventAddHandle_cgo(int fd, int events, int callbackID);
int virEventAddTimeout_cgo(int freq, int callbackID);
void virEventRegisterImpl_cgo(void);

void eventHandleCallbackInvoke(int watch, int fd, int events, uintptr_t callback, uintptr_t opaque);
void eventTimeoutCallbackInvoke(int timer, uintptr_t callback, uintptr_t opaque);
void eventHandleCallbackFree(uintptr_t callback, uintptr_t opaque);
void eventTimeoutCallbackFree(uintptr_t callback, uintptr_t opaque);

#endif /* LIBVIRT_GO_EVENTS_CFUNCS_H__ */
