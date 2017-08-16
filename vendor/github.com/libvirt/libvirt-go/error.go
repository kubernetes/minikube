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
#include "error_compat.h"

void ignoreErrorFunc(void *userData, virErrorPtr error) {
     // no-op
}
*/
import "C"

import (
	"fmt"
)

func init() {
	C.virSetErrorFunc(nil, (C.virErrorFunc)(C.ignoreErrorFunc))
	C.virInitialize()
}

type ErrorLevel int

const (
	ERR_NONE    = ErrorLevel(C.VIR_ERR_NONE)
	ERR_WARNING = ErrorLevel(C.VIR_ERR_WARNING)
	ERR_ERROR   = ErrorLevel(C.VIR_ERR_ERROR)
)

type ErrorNumber int

const (
	ERR_OK = ErrorNumber(C.VIR_ERR_OK)

	// internal error
	ERR_INTERNAL_ERROR = ErrorNumber(C.VIR_ERR_INTERNAL_ERROR)

	// memory allocation failure
	ERR_NO_MEMORY = ErrorNumber(C.VIR_ERR_NO_MEMORY)

	// no support for this function
	ERR_NO_SUPPORT = ErrorNumber(C.VIR_ERR_NO_SUPPORT)

	// could not resolve hostname
	ERR_UNKNOWN_HOST = ErrorNumber(C.VIR_ERR_UNKNOWN_HOST)

	// can't connect to hypervisor
	ERR_NO_CONNECT = ErrorNumber(C.VIR_ERR_NO_CONNECT)

	// invalid connection object
	ERR_INVALID_CONN = ErrorNumber(C.VIR_ERR_INVALID_CONN)

	// invalid domain object
	ERR_INVALID_DOMAIN = ErrorNumber(C.VIR_ERR_INVALID_DOMAIN)

	// invalid function argument
	ERR_INVALID_ARG = ErrorNumber(C.VIR_ERR_INVALID_ARG)

	// a command to hypervisor failed
	ERR_OPERATION_FAILED = ErrorNumber(C.VIR_ERR_OPERATION_FAILED)

	// a HTTP GET command to failed
	ERR_GET_FAILED = ErrorNumber(C.VIR_ERR_GET_FAILED)

	// a HTTP POST command to failed
	ERR_POST_FAILED = ErrorNumber(C.VIR_ERR_POST_FAILED)

	// unexpected HTTP error code
	ERR_HTTP_ERROR = ErrorNumber(C.VIR_ERR_HTTP_ERROR)

	// failure to serialize an S-Expr
	ERR_SEXPR_SERIAL = ErrorNumber(C.VIR_ERR_SEXPR_SERIAL)

	// could not open Xen hypervisor control
	ERR_NO_XEN = ErrorNumber(C.VIR_ERR_NO_XEN)

	// failure doing an hypervisor call
	ERR_XEN_CALL = ErrorNumber(C.VIR_ERR_XEN_CALL)

	// unknown OS type
	ERR_OS_TYPE = ErrorNumber(C.VIR_ERR_OS_TYPE)

	// missing kernel information
	ERR_NO_KERNEL = ErrorNumber(C.VIR_ERR_NO_KERNEL)

	// missing root device information
	ERR_NO_ROOT = ErrorNumber(C.VIR_ERR_NO_ROOT)

	// missing source device information
	ERR_NO_SOURCE = ErrorNumber(C.VIR_ERR_NO_SOURCE)

	// missing target device information
	ERR_NO_TARGET = ErrorNumber(C.VIR_ERR_NO_TARGET)

	// missing domain name information
	ERR_NO_NAME = ErrorNumber(C.VIR_ERR_NO_NAME)

	// missing domain OS information
	ERR_NO_OS = ErrorNumber(C.VIR_ERR_NO_OS)

	// missing domain devices information
	ERR_NO_DEVICE = ErrorNumber(C.VIR_ERR_NO_DEVICE)

	// could not open Xen Store control
	ERR_NO_XENSTORE = ErrorNumber(C.VIR_ERR_NO_XENSTORE)

	// too many drivers registered
	ERR_DRIVER_FULL = ErrorNumber(C.VIR_ERR_DRIVER_FULL)

	// not supported by the drivers (DEPRECATED)
	ERR_CALL_FAILED = ErrorNumber(C.VIR_ERR_CALL_FAILED)

	// an XML description is not well formed or broken
	ERR_XML_ERROR = ErrorNumber(C.VIR_ERR_XML_ERROR)

	// the domain already exist
	ERR_DOM_EXIST = ErrorNumber(C.VIR_ERR_DOM_EXIST)

	// operation forbidden on read-only connections
	ERR_OPERATION_DENIED = ErrorNumber(C.VIR_ERR_OPERATION_DENIED)

	// failed to open a conf file
	ERR_OPEN_FAILED = ErrorNumber(C.VIR_ERR_OPEN_FAILED)

	// failed to read a conf file
	ERR_READ_FAILED = ErrorNumber(C.VIR_ERR_READ_FAILED)

	// failed to parse a conf file
	ERR_PARSE_FAILED = ErrorNumber(C.VIR_ERR_PARSE_FAILED)

	// failed to parse the syntax of a conf file
	ERR_CONF_SYNTAX = ErrorNumber(C.VIR_ERR_CONF_SYNTAX)

	// failed to write a conf file
	ERR_WRITE_FAILED = ErrorNumber(C.VIR_ERR_WRITE_FAILED)

	// detail of an XML error
	ERR_XML_DETAIL = ErrorNumber(C.VIR_ERR_XML_DETAIL)

	// invalid network object
	ERR_INVALID_NETWORK = ErrorNumber(C.VIR_ERR_INVALID_NETWORK)

	// the network already exist
	ERR_NETWORK_EXIST = ErrorNumber(C.VIR_ERR_NETWORK_EXIST)

	// general system call failure
	ERR_SYSTEM_ERROR = ErrorNumber(C.VIR_ERR_SYSTEM_ERROR)

	// some sort of RPC error
	ERR_RPC = ErrorNumber(C.VIR_ERR_RPC)

	// error from a GNUTLS call
	ERR_GNUTLS_ERROR = ErrorNumber(C.VIR_ERR_GNUTLS_ERROR)

	// failed to start network
	WAR_NO_NETWORK = ErrorNumber(C.VIR_WAR_NO_NETWORK)

	// domain not found or unexpectedly disappeared
	ERR_NO_DOMAIN = ErrorNumber(C.VIR_ERR_NO_DOMAIN)

	// network not found
	ERR_NO_NETWORK = ErrorNumber(C.VIR_ERR_NO_NETWORK)

	// invalid MAC address
	ERR_INVALID_MAC = ErrorNumber(C.VIR_ERR_INVALID_MAC)

	// authentication failed
	ERR_AUTH_FAILED = ErrorNumber(C.VIR_ERR_AUTH_FAILED)

	// invalid storage pool object
	ERR_INVALID_STORAGE_POOL = ErrorNumber(C.VIR_ERR_INVALID_STORAGE_POOL)

	// invalid storage vol object
	ERR_INVALID_STORAGE_VOL = ErrorNumber(C.VIR_ERR_INVALID_STORAGE_VOL)

	// failed to start storage
	WAR_NO_STORAGE = ErrorNumber(C.VIR_WAR_NO_STORAGE)

	// storage pool not found
	ERR_NO_STORAGE_POOL = ErrorNumber(C.VIR_ERR_NO_STORAGE_POOL)

	// storage volume not found
	ERR_NO_STORAGE_VOL = ErrorNumber(C.VIR_ERR_NO_STORAGE_VOL)

	// failed to start node driver
	WAR_NO_NODE = ErrorNumber(C.VIR_WAR_NO_NODE)

	// invalid node device object
	ERR_INVALID_NODE_DEVICE = ErrorNumber(C.VIR_ERR_INVALID_NODE_DEVICE)

	// node device not found
	ERR_NO_NODE_DEVICE = ErrorNumber(C.VIR_ERR_NO_NODE_DEVICE)

	// security model not found
	ERR_NO_SECURITY_MODEL = ErrorNumber(C.VIR_ERR_NO_SECURITY_MODEL)

	// operation is not applicable at this time
	ERR_OPERATION_INVALID = ErrorNumber(C.VIR_ERR_OPERATION_INVALID)

	// failed to start interface driver
	WAR_NO_INTERFACE = ErrorNumber(C.VIR_WAR_NO_INTERFACE)

	// interface driver not running
	ERR_NO_INTERFACE = ErrorNumber(C.VIR_ERR_NO_INTERFACE)

	// invalid interface object
	ERR_INVALID_INTERFACE = ErrorNumber(C.VIR_ERR_INVALID_INTERFACE)

	// more than one matching interface found
	ERR_MULTIPLE_INTERFACES = ErrorNumber(C.VIR_ERR_MULTIPLE_INTERFACES)

	// failed to start nwfilter driver
	WAR_NO_NWFILTER = ErrorNumber(C.VIR_WAR_NO_NWFILTER)

	// invalid nwfilter object
	ERR_INVALID_NWFILTER = ErrorNumber(C.VIR_ERR_INVALID_NWFILTER)

	// nw filter pool not found
	ERR_NO_NWFILTER = ErrorNumber(C.VIR_ERR_NO_NWFILTER)

	// nw filter pool not found
	ERR_BUILD_FIREWALL = ErrorNumber(C.VIR_ERR_BUILD_FIREWALL)

	// failed to start secret storage
	WAR_NO_SECRET = ErrorNumber(C.VIR_WAR_NO_SECRET)

	// invalid secret
	ERR_INVALID_SECRET = ErrorNumber(C.VIR_ERR_INVALID_SECRET)

	// secret not found
	ERR_NO_SECRET = ErrorNumber(C.VIR_ERR_NO_SECRET)

	// unsupported configuration construct
	ERR_CONFIG_UNSUPPORTED = ErrorNumber(C.VIR_ERR_CONFIG_UNSUPPORTED)

	// timeout occurred during operation
	ERR_OPERATION_TIMEOUT = ErrorNumber(C.VIR_ERR_OPERATION_TIMEOUT)

	// a migration worked, but making the VM persist on the dest host failed
	ERR_MIGRATE_PERSIST_FAILED = ErrorNumber(C.VIR_ERR_MIGRATE_PERSIST_FAILED)

	// a synchronous hook script failed
	ERR_HOOK_SCRIPT_FAILED = ErrorNumber(C.VIR_ERR_HOOK_SCRIPT_FAILED)

	// invalid domain snapshot
	ERR_INVALID_DOMAIN_SNAPSHOT = ErrorNumber(C.VIR_ERR_INVALID_DOMAIN_SNAPSHOT)

	// domain snapshot not found
	ERR_NO_DOMAIN_SNAPSHOT = ErrorNumber(C.VIR_ERR_NO_DOMAIN_SNAPSHOT)

	// stream pointer not valid
	ERR_INVALID_STREAM = ErrorNumber(C.VIR_ERR_INVALID_STREAM)

	// valid API use but unsupported by the given driver
	ERR_ARGUMENT_UNSUPPORTED = ErrorNumber(C.VIR_ERR_ARGUMENT_UNSUPPORTED)

	// storage pool probe failed
	ERR_STORAGE_PROBE_FAILED = ErrorNumber(C.VIR_ERR_STORAGE_PROBE_FAILED)

	// storage pool already built
	ERR_STORAGE_POOL_BUILT = ErrorNumber(C.VIR_ERR_STORAGE_POOL_BUILT)

	// force was not requested for a risky domain snapshot revert
	ERR_SNAPSHOT_REVERT_RISKY = ErrorNumber(C.VIR_ERR_SNAPSHOT_REVERT_RISKY)

	// operation on a domain was canceled/aborted by user
	ERR_OPERATION_ABORTED = ErrorNumber(C.VIR_ERR_OPERATION_ABORTED)

	// authentication cancelled
	ERR_AUTH_CANCELLED = ErrorNumber(C.VIR_ERR_AUTH_CANCELLED)

	// The metadata is not present
	ERR_NO_DOMAIN_METADATA = ErrorNumber(C.VIR_ERR_NO_DOMAIN_METADATA)

	// Migration is not safe
	ERR_MIGRATE_UNSAFE = ErrorNumber(C.VIR_ERR_MIGRATE_UNSAFE)

	// integer overflow
	ERR_OVERFLOW = ErrorNumber(C.VIR_ERR_OVERFLOW)

	// action prevented by block copy job
	ERR_BLOCK_COPY_ACTIVE = ErrorNumber(C.VIR_ERR_BLOCK_COPY_ACTIVE)

	// The requested operation is not supported
	ERR_OPERATION_UNSUPPORTED = ErrorNumber(C.VIR_ERR_OPERATION_UNSUPPORTED)

	// error in ssh transport driver
	ERR_SSH = ErrorNumber(C.VIR_ERR_SSH)

	// guest agent is unresponsive, not running or not usable
	ERR_AGENT_UNRESPONSIVE = ErrorNumber(C.VIR_ERR_AGENT_UNRESPONSIVE)

	// resource is already in use
	ERR_RESOURCE_BUSY = ErrorNumber(C.VIR_ERR_RESOURCE_BUSY)

	// operation on the object/resource was denied
	ERR_ACCESS_DENIED = ErrorNumber(C.VIR_ERR_ACCESS_DENIED)

	// error from a dbus service
	ERR_DBUS_SERVICE = ErrorNumber(C.VIR_ERR_DBUS_SERVICE)

	// the storage vol already exists
	ERR_STORAGE_VOL_EXIST = ErrorNumber(C.VIR_ERR_STORAGE_VOL_EXIST)

	// given CPU is incompatible with host CPU
	ERR_CPU_INCOMPATIBLE = ErrorNumber(C.VIR_ERR_CPU_INCOMPATIBLE)

	// XML document doesn't validate against schema
	ERR_XML_INVALID_SCHEMA = ErrorNumber(C.VIR_ERR_XML_INVALID_SCHEMA)

	// Finish API succeeded but it is expected to return NULL */
	ERR_MIGRATE_FINISH_OK = ErrorNumber(C.VIR_ERR_MIGRATE_FINISH_OK)

	// authentication unavailable
	ERR_AUTH_UNAVAILABLE = ErrorNumber(C.VIR_ERR_AUTH_UNAVAILABLE)

	// Server was not found
	ERR_NO_SERVER = ErrorNumber(C.VIR_ERR_NO_SERVER)

	// Client was not found
	ERR_NO_CLIENT = ErrorNumber(C.VIR_ERR_NO_CLIENT)

	// guest agent replies with wrong id to guest sync command
	ERR_AGENT_UNSYNCED = ErrorNumber(C.VIR_ERR_AGENT_UNSYNCED)

	// error in libssh transport driver
	ERR_LIBSSH = ErrorNumber(C.VIR_ERR_LIBSSH)
)

type ErrorDomain int

const (
	FROM_NONE = ErrorDomain(C.VIR_FROM_NONE)

	// Error at Xen hypervisor layer
	FROM_XEN = ErrorDomain(C.VIR_FROM_XEN)

	// Error at connection with xend daemon
	FROM_XEND = ErrorDomain(C.VIR_FROM_XEND)

	// Error at connection with xen store
	FROM_XENSTORE = ErrorDomain(C.VIR_FROM_XENSTORE)

	// Error in the S-Expression code
	FROM_SEXPR = ErrorDomain(C.VIR_FROM_SEXPR)

	// Error in the XML code
	FROM_XML = ErrorDomain(C.VIR_FROM_XML)

	// Error when operating on a domain
	FROM_DOM = ErrorDomain(C.VIR_FROM_DOM)

	// Error in the XML-RPC code
	FROM_RPC = ErrorDomain(C.VIR_FROM_RPC)

	// Error in the proxy code; unused since 0.8.6
	FROM_PROXY = ErrorDomain(C.VIR_FROM_PROXY)

	// Error in the configuration file handling
	FROM_CONF = ErrorDomain(C.VIR_FROM_CONF)

	// Error at the QEMU daemon
	FROM_QEMU = ErrorDomain(C.VIR_FROM_QEMU)

	// Error when operating on a network
	FROM_NET = ErrorDomain(C.VIR_FROM_NET)

	// Error from test driver
	FROM_TEST = ErrorDomain(C.VIR_FROM_TEST)

	// Error from remote driver
	FROM_REMOTE = ErrorDomain(C.VIR_FROM_REMOTE)

	// Error from OpenVZ driver
	FROM_OPENVZ = ErrorDomain(C.VIR_FROM_OPENVZ)

	// Error at Xen XM layer
	FROM_XENXM = ErrorDomain(C.VIR_FROM_XENXM)

	// Error in the Linux Stats code
	FROM_STATS_LINUX = ErrorDomain(C.VIR_FROM_STATS_LINUX)

	// Error from Linux Container driver
	FROM_LXC = ErrorDomain(C.VIR_FROM_LXC)

	// Error from storage driver
	FROM_STORAGE = ErrorDomain(C.VIR_FROM_STORAGE)

	// Error from network config
	FROM_NETWORK = ErrorDomain(C.VIR_FROM_NETWORK)

	// Error from domain config
	FROM_DOMAIN = ErrorDomain(C.VIR_FROM_DOMAIN)

	// Error at the UML driver
	FROM_UML = ErrorDomain(C.VIR_FROM_UML)

	// Error from node device monitor
	FROM_NODEDEV = ErrorDomain(C.VIR_FROM_NODEDEV)

	// Error from xen inotify layer
	FROM_XEN_INOTIFY = ErrorDomain(C.VIR_FROM_XEN_INOTIFY)

	// Error from security framework
	FROM_SECURITY = ErrorDomain(C.VIR_FROM_SECURITY)

	// Error from VirtualBox driver
	FROM_VBOX = ErrorDomain(C.VIR_FROM_VBOX)

	// Error when operating on an interface
	FROM_INTERFACE = ErrorDomain(C.VIR_FROM_INTERFACE)

	// The OpenNebula driver no longer exists. Retained for ABI/API compat only
	FROM_ONE = ErrorDomain(C.VIR_FROM_ONE)

	// Error from ESX driver
	FROM_ESX = ErrorDomain(C.VIR_FROM_ESX)

	// Error from IBM power hypervisor
	FROM_PHYP = ErrorDomain(C.VIR_FROM_PHYP)

	// Error from secret storage
	FROM_SECRET = ErrorDomain(C.VIR_FROM_SECRET)

	// Error from CPU driver
	FROM_CPU = ErrorDomain(C.VIR_FROM_CPU)

	// Error from XenAPI
	FROM_XENAPI = ErrorDomain(C.VIR_FROM_XENAPI)

	// Error from network filter driver
	FROM_NWFILTER = ErrorDomain(C.VIR_FROM_NWFILTER)

	// Error from Synchronous hooks
	FROM_HOOK = ErrorDomain(C.VIR_FROM_HOOK)

	// Error from domain snapshot
	FROM_DOMAIN_SNAPSHOT = ErrorDomain(C.VIR_FROM_DOMAIN_SNAPSHOT)

	// Error from auditing subsystem
	FROM_AUDIT = ErrorDomain(C.VIR_FROM_AUDIT)

	// Error from sysinfo/SMBIOS
	FROM_SYSINFO = ErrorDomain(C.VIR_FROM_SYSINFO)

	// Error from I/O streams
	FROM_STREAMS = ErrorDomain(C.VIR_FROM_STREAMS)

	// Error from VMware driver
	FROM_VMWARE = ErrorDomain(C.VIR_FROM_VMWARE)

	// Error from event loop impl
	FROM_EVENT = ErrorDomain(C.VIR_FROM_EVENT)

	// Error from libxenlight driver
	FROM_LIBXL = ErrorDomain(C.VIR_FROM_LIBXL)

	// Error from lock manager
	FROM_LOCKING = ErrorDomain(C.VIR_FROM_LOCKING)

	// Error from Hyper-V driver
	FROM_HYPERV = ErrorDomain(C.VIR_FROM_HYPERV)

	// Error from capabilities
	FROM_CAPABILITIES = ErrorDomain(C.VIR_FROM_CAPABILITIES)

	// Error from URI handling
	FROM_URI = ErrorDomain(C.VIR_FROM_URI)

	// Error from auth handling
	FROM_AUTH = ErrorDomain(C.VIR_FROM_AUTH)

	// Error from DBus
	FROM_DBUS = ErrorDomain(C.VIR_FROM_DBUS)

	// Error from Parallels
	FROM_PARALLELS = ErrorDomain(C.VIR_FROM_PARALLELS)

	// Error from Device
	FROM_DEVICE = ErrorDomain(C.VIR_FROM_DEVICE)

	// Error from libssh2 connection transport
	FROM_SSH = ErrorDomain(C.VIR_FROM_SSH)

	// Error from lockspace
	FROM_LOCKSPACE = ErrorDomain(C.VIR_FROM_LOCKSPACE)

	// Error from initctl device communication
	FROM_INITCTL = ErrorDomain(C.VIR_FROM_INITCTL)

	// Error from identity code
	FROM_IDENTITY = ErrorDomain(C.VIR_FROM_IDENTITY)

	// Error from cgroups
	FROM_CGROUP = ErrorDomain(C.VIR_FROM_CGROUP)

	// Error from access control manager
	FROM_ACCESS = ErrorDomain(C.VIR_FROM_ACCESS)

	// Error from systemd code
	FROM_SYSTEMD = ErrorDomain(C.VIR_FROM_SYSTEMD)

	// Error from bhyve driver
	FROM_BHYVE = ErrorDomain(C.VIR_FROM_BHYVE)

	// Error from crypto code
	FROM_CRYPTO = ErrorDomain(C.VIR_FROM_CRYPTO)

	// Error from firewall
	FROM_FIREWALL = ErrorDomain(C.VIR_FROM_FIREWALL)

	// Erorr from polkit code
	FROM_POLKIT = ErrorDomain(C.VIR_FROM_POLKIT)

	// Error from thread utils
	FROM_THREAD = ErrorDomain(C.VIR_FROM_THREAD)

	// Error from admin backend
	FROM_ADMIN = ErrorDomain(C.VIR_FROM_ADMIN)

	// Error from log manager
	FROM_LOGGING = ErrorDomain(C.VIR_FROM_LOGGING)

	// Error from Xen xl config code
	FROM_XENXL = ErrorDomain(C.VIR_FROM_XENXL)

	// Error from perf
	FROM_PERF = ErrorDomain(C.VIR_FROM_PERF)

	// Error from libssh
	FROM_LIBSSH = ErrorDomain(C.VIR_FROM_LIBSSH)
)

type Error struct {
	Code    ErrorNumber
	Domain  ErrorDomain
	Message string
	Level   ErrorLevel
}

func (err Error) Error() string {
	return fmt.Sprintf("virError(Code=%d, Domain=%d, Message='%s')",
		err.Code, err.Domain, err.Message)
}

func GetLastError() Error {
	err := C.virGetLastError()
	if err == nil {
		return Error{
			Code:    ERR_OK,
			Domain:  FROM_NONE,
			Message: "Missing error",
			Level:   ERR_NONE,
		}
	}
	virErr := Error{
		Code:    ErrorNumber(err.code),
		Domain:  ErrorDomain(err.domain),
		Message: C.GoString(err.message),
		Level:   ErrorLevel(err.level),
	}
	C.virResetError(err)
	return virErr
}

func GetNotImplementedError(apiname string) Error {
	return Error{
		Code:    ERR_NO_SUPPORT,
		Domain:  FROM_NONE,
		Message: fmt.Sprintf("Function '%s' not available in the libvirt library used during Go build", apiname),
		Level:   ERR_ERROR,
	}
}
