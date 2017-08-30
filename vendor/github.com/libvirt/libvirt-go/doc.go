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
 * Copyright (C) 2016 Red Hat, Inc.
 *
 */

// Package libvirt provides a Go binding to the libvirt C library
//
// Through conditional compilation it supports libvirt versions 1.2.0 onwards.
// This is done automatically, with no requirement to use magic Go build tags.
// If an API was not available in the particular version of libvirt this package
// was built against, an error will be returned with a code of ERR_NO_SUPPORT.
// This is the same code seen if using a new libvirt library to talk to an old
// libvirtd lacking the API, or if a hypervisor does not support a given feature,
// so an application can easily handle all scenarios together.
//
// The Go binding is a fairly direct mapping of the underling C API which seeks
// to maximise the use of the Go type system to allow strong compiler type
// checking. The following rules describe how APIs/constants are mapped from C
// to Go
//
// For structs, the 'vir' prefix and 'Ptr' suffix are removed from the name.
// e.g. virConnectPtr in C becomes 'Connect' in Go.
//
// For structs which are reference counted at the C level, it is neccessary to
// explicitly release the reference at the Go level. e.g. if a Go method returns
// a '* Domain' struct, it is neccessary to call 'Free' on this when no longer
// required. The use of 'defer' is recommended for this purpose
//
//  dom, err := conn.LookupDomainByName("myguest")
//  if err != nil {
//      ...
//  }
//  defer dom.Free()
//
// If multiple goroutines are using the same libvirt object struct, it may
// not be possible to determine which goroutine should call 'Free'. In such
// scenarios each new goroutine should call 'Ref' to obtain a private reference
// on the underlying C struct. All goroutines can call 'Free' unconditionally
// with the final one causing the release of the C object.
//
// For methods, the 'vir' prefix and object name prefix are remove from the name.
// The C functions become methods with an object receiver. e.g.
// 'virDomainScreenshot' in C becomes 'Screenshot' with a 'Domain *' receiver.
//
// For methods which accept a 'unsigned int flags' parameter in the C level,
// the corresponding Go parameter will be a named type corresponding to the
// C enum that defines the valid flags. For example, the ListAllDomains
// method takes a 'flags ConnectListAllDomainsFlags' parameter. If there are
// not currently any flags defined for a method in the C API, then the Go
// method parameter will be declared as a "flags uint32". Callers should always
// pass the literal integer value 0 for such parameters, without forcing any
// specific type. This will allow compatibility with future updates to the
// libvirt-go binding which may replace the 'uint32' type with a enum type
// at a later date.
//
// For enums, the VIR_ prefix is removed from the name. The enums get a dedicated
// type defined in Go. e.g. the VIR_NODE_SUSPEND_TARGET_MEM enum constant in C,
// becomes NODE_SUSPEND_TARGET_MEM with a type of NodeSuspendTarget.
//
// Methods accepting or returning virTypedParameter arrays in C will map the
// parameters into a Go struct. The struct will contain two fields for each
// possible parameter. One boolean field with a suffix of 'Set' indicates whether
// the parameter has a value set, and the other custom typed field provides the
// parameter value. This makes it possible to distinguish a parameter with a
// default value of '0' from a parameter which is 0 because it isn't supported by
// the hypervisor. If the C API defines additional typed parameters, then the
// corresponding Go struct will be extended to have further fields.
// e.g. the GetMemoryStats method in Go (which is backed by
// virNodeGetMemoryStats in C) will return a NodeMemoryStats struct containing
// the typed parameter values.
//
//  stats, err := conn.GetMemoryParameters()
//  if err != nil {
//     ....
//  }
//  if stats.TotalSet {
//     fmt.Printf("Total memory: %d KB", stats.Total)
//  }
//
// Every method that can fail will include an 'error' object as the last return
// value. This will be an instance of the Error struct if an error occurred. To
// check for specific libvirt error codes, it is neccessary to cast the error.
//
//  err := storage_vol.Wipe(0)
//  if err != nil {
//     lverr, ok := err.(libvirt.Error)
//     if ok && lverr.Code == libvirt.ERR_NO_SUPPORT {
//         fmt.Println("Wiping storage volumes is not supported");
//     } else {
//         fmt.Println("Error wiping storage volume: %s", err)
//     }
//  }
//
// Example usage
//
// To connect to libvirt
//
//   import (
//       libvirt "github.com/libvirt/libvirt-go"
//   )
//   conn, err := libvirt.NewConnect("qemu:///system")
//   if err != nil {
//       ...
//   }
//   defer conn.Close()
//
//   doms, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
//   if err != nil {
//       ...
//   }
//
//   fmt.Printf("%d running domains:\n", len(doms))
//   for _, dom := range doms {
//       name, err := dom.GetName()
//       if err == nil {
//           fmt.Printf("  %s\n", name)
//       }
//       dom.Free()
//   }
//
package libvirt
