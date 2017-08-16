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

import (
	"fmt"
	"os"
	"reflect"
	"sync"
	"unsafe"
)

/*
#cgo pkg-config: libvirt
#include <libvirt/libvirt.h>
#include <libvirt/virterror.h>
#include <stdlib.h>
#include "connect_compat.h"
#include "connect_cfuncs.h"
*/
import "C"

func init() {
	C.virInitialize()
}

const (
	VERSION_NUMBER = uint32(C.LIBVIR_VERSION_NUMBER)
)

type ConnectCloseReason int

const (
	CONNECT_CLOSE_REASON_ERROR     = ConnectCloseReason(C.VIR_CONNECT_CLOSE_REASON_ERROR)
	CONNECT_CLOSE_REASON_EOF       = ConnectCloseReason(C.VIR_CONNECT_CLOSE_REASON_EOF)
	CONNECT_CLOSE_REASON_KEEPALIVE = ConnectCloseReason(C.VIR_CONNECT_CLOSE_REASON_KEEPALIVE)
	CONNECT_CLOSE_REASON_CLIENT    = ConnectCloseReason(C.VIR_CONNECT_CLOSE_REASON_CLIENT)
)

type ConnectListAllDomainsFlags int

const (
	CONNECT_LIST_DOMAINS_ACTIVE         = ConnectListAllDomainsFlags(C.VIR_CONNECT_LIST_DOMAINS_ACTIVE)
	CONNECT_LIST_DOMAINS_INACTIVE       = ConnectListAllDomainsFlags(C.VIR_CONNECT_LIST_DOMAINS_INACTIVE)
	CONNECT_LIST_DOMAINS_PERSISTENT     = ConnectListAllDomainsFlags(C.VIR_CONNECT_LIST_DOMAINS_PERSISTENT)
	CONNECT_LIST_DOMAINS_TRANSIENT      = ConnectListAllDomainsFlags(C.VIR_CONNECT_LIST_DOMAINS_TRANSIENT)
	CONNECT_LIST_DOMAINS_RUNNING        = ConnectListAllDomainsFlags(C.VIR_CONNECT_LIST_DOMAINS_RUNNING)
	CONNECT_LIST_DOMAINS_PAUSED         = ConnectListAllDomainsFlags(C.VIR_CONNECT_LIST_DOMAINS_PAUSED)
	CONNECT_LIST_DOMAINS_SHUTOFF        = ConnectListAllDomainsFlags(C.VIR_CONNECT_LIST_DOMAINS_SHUTOFF)
	CONNECT_LIST_DOMAINS_OTHER          = ConnectListAllDomainsFlags(C.VIR_CONNECT_LIST_DOMAINS_OTHER)
	CONNECT_LIST_DOMAINS_MANAGEDSAVE    = ConnectListAllDomainsFlags(C.VIR_CONNECT_LIST_DOMAINS_MANAGEDSAVE)
	CONNECT_LIST_DOMAINS_NO_MANAGEDSAVE = ConnectListAllDomainsFlags(C.VIR_CONNECT_LIST_DOMAINS_NO_MANAGEDSAVE)
	CONNECT_LIST_DOMAINS_AUTOSTART      = ConnectListAllDomainsFlags(C.VIR_CONNECT_LIST_DOMAINS_AUTOSTART)
	CONNECT_LIST_DOMAINS_NO_AUTOSTART   = ConnectListAllDomainsFlags(C.VIR_CONNECT_LIST_DOMAINS_NO_AUTOSTART)
	CONNECT_LIST_DOMAINS_HAS_SNAPSHOT   = ConnectListAllDomainsFlags(C.VIR_CONNECT_LIST_DOMAINS_HAS_SNAPSHOT)
	CONNECT_LIST_DOMAINS_NO_SNAPSHOT    = ConnectListAllDomainsFlags(C.VIR_CONNECT_LIST_DOMAINS_NO_SNAPSHOT)
)

type ConnectListAllNetworksFlags int

const (
	CONNECT_LIST_NETWORKS_INACTIVE     = ConnectListAllNetworksFlags(C.VIR_CONNECT_LIST_NETWORKS_INACTIVE)
	CONNECT_LIST_NETWORKS_ACTIVE       = ConnectListAllNetworksFlags(C.VIR_CONNECT_LIST_NETWORKS_ACTIVE)
	CONNECT_LIST_NETWORKS_PERSISTENT   = ConnectListAllNetworksFlags(C.VIR_CONNECT_LIST_NETWORKS_PERSISTENT)
	CONNECT_LIST_NETWORKS_TRANSIENT    = ConnectListAllNetworksFlags(C.VIR_CONNECT_LIST_NETWORKS_TRANSIENT)
	CONNECT_LIST_NETWORKS_AUTOSTART    = ConnectListAllNetworksFlags(C.VIR_CONNECT_LIST_NETWORKS_AUTOSTART)
	CONNECT_LIST_NETWORKS_NO_AUTOSTART = ConnectListAllNetworksFlags(C.VIR_CONNECT_LIST_NETWORKS_NO_AUTOSTART)
)

type ConnectListAllStoragePoolsFlags int

const (
	CONNECT_LIST_STORAGE_POOLS_INACTIVE     = ConnectListAllStoragePoolsFlags(C.VIR_CONNECT_LIST_STORAGE_POOLS_INACTIVE)
	CONNECT_LIST_STORAGE_POOLS_ACTIVE       = ConnectListAllStoragePoolsFlags(C.VIR_CONNECT_LIST_STORAGE_POOLS_ACTIVE)
	CONNECT_LIST_STORAGE_POOLS_PERSISTENT   = ConnectListAllStoragePoolsFlags(C.VIR_CONNECT_LIST_STORAGE_POOLS_PERSISTENT)
	CONNECT_LIST_STORAGE_POOLS_TRANSIENT    = ConnectListAllStoragePoolsFlags(C.VIR_CONNECT_LIST_STORAGE_POOLS_TRANSIENT)
	CONNECT_LIST_STORAGE_POOLS_AUTOSTART    = ConnectListAllStoragePoolsFlags(C.VIR_CONNECT_LIST_STORAGE_POOLS_AUTOSTART)
	CONNECT_LIST_STORAGE_POOLS_NO_AUTOSTART = ConnectListAllStoragePoolsFlags(C.VIR_CONNECT_LIST_STORAGE_POOLS_NO_AUTOSTART)
	CONNECT_LIST_STORAGE_POOLS_DIR          = ConnectListAllStoragePoolsFlags(C.VIR_CONNECT_LIST_STORAGE_POOLS_DIR)
	CONNECT_LIST_STORAGE_POOLS_FS           = ConnectListAllStoragePoolsFlags(C.VIR_CONNECT_LIST_STORAGE_POOLS_FS)
	CONNECT_LIST_STORAGE_POOLS_NETFS        = ConnectListAllStoragePoolsFlags(C.VIR_CONNECT_LIST_STORAGE_POOLS_NETFS)
	CONNECT_LIST_STORAGE_POOLS_LOGICAL      = ConnectListAllStoragePoolsFlags(C.VIR_CONNECT_LIST_STORAGE_POOLS_LOGICAL)
	CONNECT_LIST_STORAGE_POOLS_DISK         = ConnectListAllStoragePoolsFlags(C.VIR_CONNECT_LIST_STORAGE_POOLS_DISK)
	CONNECT_LIST_STORAGE_POOLS_ISCSI        = ConnectListAllStoragePoolsFlags(C.VIR_CONNECT_LIST_STORAGE_POOLS_ISCSI)
	CONNECT_LIST_STORAGE_POOLS_SCSI         = ConnectListAllStoragePoolsFlags(C.VIR_CONNECT_LIST_STORAGE_POOLS_SCSI)
	CONNECT_LIST_STORAGE_POOLS_MPATH        = ConnectListAllStoragePoolsFlags(C.VIR_CONNECT_LIST_STORAGE_POOLS_MPATH)
	CONNECT_LIST_STORAGE_POOLS_RBD          = ConnectListAllStoragePoolsFlags(C.VIR_CONNECT_LIST_STORAGE_POOLS_RBD)
	CONNECT_LIST_STORAGE_POOLS_SHEEPDOG     = ConnectListAllStoragePoolsFlags(C.VIR_CONNECT_LIST_STORAGE_POOLS_SHEEPDOG)
	CONNECT_LIST_STORAGE_POOLS_GLUSTER      = ConnectListAllStoragePoolsFlags(C.VIR_CONNECT_LIST_STORAGE_POOLS_GLUSTER)
	CONNECT_LIST_STORAGE_POOLS_ZFS          = ConnectListAllStoragePoolsFlags(C.VIR_CONNECT_LIST_STORAGE_POOLS_ZFS)
	CONNECT_LIST_STORAGE_POOLS_VSTORAGE     = ConnectListAllStoragePoolsFlags(C.VIR_CONNECT_LIST_STORAGE_POOLS_VSTORAGE)
)

type ConnectBaselineCPUFlags int

const (
	CONNECT_BASELINE_CPU_EXPAND_FEATURES = ConnectBaselineCPUFlags(C.VIR_CONNECT_BASELINE_CPU_EXPAND_FEATURES)
	CONNECT_BASELINE_CPU_MIGRATABLE      = ConnectBaselineCPUFlags(C.VIR_CONNECT_BASELINE_CPU_MIGRATABLE)
)

type ConnectCompareCPUFlags int

const (
	CONNECT_COMPARE_CPU_FAIL_INCOMPATIBLE = ConnectCompareCPUFlags(C.VIR_CONNECT_COMPARE_CPU_FAIL_INCOMPATIBLE)
)

type ConnectListAllInterfacesFlags int

const (
	CONNECT_LIST_INTERFACES_INACTIVE = ConnectListAllInterfacesFlags(C.VIR_CONNECT_LIST_INTERFACES_INACTIVE)
	CONNECT_LIST_INTERFACES_ACTIVE   = ConnectListAllInterfacesFlags(C.VIR_CONNECT_LIST_INTERFACES_ACTIVE)
)

type ConnectListAllNodeDeviceFlags int

const (
	CONNECT_LIST_NODE_DEVICES_CAP_SYSTEM        = ConnectListAllNodeDeviceFlags(C.VIR_CONNECT_LIST_NODE_DEVICES_CAP_SYSTEM)
	CONNECT_LIST_NODE_DEVICES_CAP_PCI_DEV       = ConnectListAllNodeDeviceFlags(C.VIR_CONNECT_LIST_NODE_DEVICES_CAP_PCI_DEV)
	CONNECT_LIST_NODE_DEVICES_CAP_USB_DEV       = ConnectListAllNodeDeviceFlags(C.VIR_CONNECT_LIST_NODE_DEVICES_CAP_USB_DEV)
	CONNECT_LIST_NODE_DEVICES_CAP_USB_INTERFACE = ConnectListAllNodeDeviceFlags(C.VIR_CONNECT_LIST_NODE_DEVICES_CAP_USB_INTERFACE)
	CONNECT_LIST_NODE_DEVICES_CAP_NET           = ConnectListAllNodeDeviceFlags(C.VIR_CONNECT_LIST_NODE_DEVICES_CAP_NET)
	CONNECT_LIST_NODE_DEVICES_CAP_SCSI_HOST     = ConnectListAllNodeDeviceFlags(C.VIR_CONNECT_LIST_NODE_DEVICES_CAP_SCSI_HOST)
	CONNECT_LIST_NODE_DEVICES_CAP_SCSI_TARGET   = ConnectListAllNodeDeviceFlags(C.VIR_CONNECT_LIST_NODE_DEVICES_CAP_SCSI_TARGET)
	CONNECT_LIST_NODE_DEVICES_CAP_SCSI          = ConnectListAllNodeDeviceFlags(C.VIR_CONNECT_LIST_NODE_DEVICES_CAP_SCSI)
	CONNECT_LIST_NODE_DEVICES_CAP_STORAGE       = ConnectListAllNodeDeviceFlags(C.VIR_CONNECT_LIST_NODE_DEVICES_CAP_STORAGE)
	CONNECT_LIST_NODE_DEVICES_CAP_FC_HOST       = ConnectListAllNodeDeviceFlags(C.VIR_CONNECT_LIST_NODE_DEVICES_CAP_FC_HOST)
	CONNECT_LIST_NODE_DEVICES_CAP_VPORTS        = ConnectListAllNodeDeviceFlags(C.VIR_CONNECT_LIST_NODE_DEVICES_CAP_VPORTS)
	CONNECT_LIST_NODE_DEVICES_CAP_SCSI_GENERIC  = ConnectListAllNodeDeviceFlags(C.VIR_CONNECT_LIST_NODE_DEVICES_CAP_SCSI_GENERIC)
	CONNECT_LIST_NODE_DEVICES_CAP_DRM           = ConnectListAllNodeDeviceFlags(C.VIR_CONNECT_LIST_NODE_DEVICES_CAP_DRM)
	CONNECT_LIST_NODE_DEVICES_CAP_MDEV          = ConnectListAllNodeDeviceFlags(C.VIR_CONNECT_LIST_NODE_DEVICES_CAP_MDEV)
	CONNECT_LIST_NODE_DEVICES_CAP_MDEV_TYPES    = ConnectListAllNodeDeviceFlags(C.VIR_CONNECT_LIST_NODE_DEVICES_CAP_MDEV_TYPES)
	CONNECT_LIST_NODE_DEVICES_CAP_CCW_DEV       = ConnectListAllNodeDeviceFlags(C.VIR_CONNECT_LIST_NODE_DEVICES_CAP_CCW_DEV)
)

type ConnectListAllSecretsFlags int

const (
	CONNECT_LIST_SECRETS_EPHEMERAL    = ConnectListAllSecretsFlags(C.VIR_CONNECT_LIST_SECRETS_EPHEMERAL)
	CONNECT_LIST_SECRETS_NO_EPHEMERAL = ConnectListAllSecretsFlags(C.VIR_CONNECT_LIST_SECRETS_NO_EPHEMERAL)
	CONNECT_LIST_SECRETS_PRIVATE      = ConnectListAllSecretsFlags(C.VIR_CONNECT_LIST_SECRETS_PRIVATE)
	CONNECT_LIST_SECRETS_NO_PRIVATE   = ConnectListAllSecretsFlags(C.VIR_CONNECT_LIST_SECRETS_NO_PRIVATE)
)

type ConnectGetAllDomainStatsFlags int

const (
	CONNECT_GET_ALL_DOMAINS_STATS_ACTIVE        = ConnectGetAllDomainStatsFlags(C.VIR_CONNECT_GET_ALL_DOMAINS_STATS_ACTIVE)
	CONNECT_GET_ALL_DOMAINS_STATS_INACTIVE      = ConnectGetAllDomainStatsFlags(C.VIR_CONNECT_GET_ALL_DOMAINS_STATS_INACTIVE)
	CONNECT_GET_ALL_DOMAINS_STATS_PERSISTENT    = ConnectGetAllDomainStatsFlags(C.VIR_CONNECT_GET_ALL_DOMAINS_STATS_PERSISTENT)
	CONNECT_GET_ALL_DOMAINS_STATS_TRANSIENT     = ConnectGetAllDomainStatsFlags(C.VIR_CONNECT_GET_ALL_DOMAINS_STATS_TRANSIENT)
	CONNECT_GET_ALL_DOMAINS_STATS_RUNNING       = ConnectGetAllDomainStatsFlags(C.VIR_CONNECT_GET_ALL_DOMAINS_STATS_RUNNING)
	CONNECT_GET_ALL_DOMAINS_STATS_PAUSED        = ConnectGetAllDomainStatsFlags(C.VIR_CONNECT_GET_ALL_DOMAINS_STATS_PAUSED)
	CONNECT_GET_ALL_DOMAINS_STATS_SHUTOFF       = ConnectGetAllDomainStatsFlags(C.VIR_CONNECT_GET_ALL_DOMAINS_STATS_SHUTOFF)
	CONNECT_GET_ALL_DOMAINS_STATS_OTHER         = ConnectGetAllDomainStatsFlags(C.VIR_CONNECT_GET_ALL_DOMAINS_STATS_OTHER)
	CONNECT_GET_ALL_DOMAINS_STATS_BACKING       = ConnectGetAllDomainStatsFlags(C.VIR_CONNECT_GET_ALL_DOMAINS_STATS_BACKING)
	CONNECT_GET_ALL_DOMAINS_STATS_ENFORCE_STATS = ConnectGetAllDomainStatsFlags(C.VIR_CONNECT_GET_ALL_DOMAINS_STATS_ENFORCE_STATS)
)

type ConnectFlags int

const (
	CONNECT_RO         = ConnectFlags(C.VIR_CONNECT_RO)
	CONNECT_NO_ALIASES = ConnectFlags(C.VIR_CONNECT_NO_ALIASES)
)

type ConnectDomainEventAgentLifecycleState int

const (
	CONNECT_DOMAIN_EVENT_AGENT_LIFECYCLE_STATE_CONNECTED    = ConnectDomainEventAgentLifecycleState(C.VIR_CONNECT_DOMAIN_EVENT_AGENT_LIFECYCLE_STATE_CONNECTED)
	CONNECT_DOMAIN_EVENT_AGENT_LIFECYCLE_STATE_DISCONNECTED = ConnectDomainEventAgentLifecycleState(C.VIR_CONNECT_DOMAIN_EVENT_AGENT_LIFECYCLE_STATE_DISCONNECTED)
)

type ConnectDomainEventAgentLifecycleReason int

const (
	CONNECT_DOMAIN_EVENT_AGENT_LIFECYCLE_REASON_UNKNOWN        = ConnectDomainEventAgentLifecycleReason(C.VIR_CONNECT_DOMAIN_EVENT_AGENT_LIFECYCLE_REASON_UNKNOWN)
	CONNECT_DOMAIN_EVENT_AGENT_LIFECYCLE_REASON_DOMAIN_STARTED = ConnectDomainEventAgentLifecycleReason(C.VIR_CONNECT_DOMAIN_EVENT_AGENT_LIFECYCLE_REASON_DOMAIN_STARTED)
	CONNECT_DOMAIN_EVENT_AGENT_LIFECYCLE_REASON_CHANNEL        = ConnectDomainEventAgentLifecycleReason(C.VIR_CONNECT_DOMAIN_EVENT_AGENT_LIFECYCLE_REASON_CHANNEL)
)

type CPUCompareResult int

const (
	CPU_COMPARE_ERROR        = CPUCompareResult(C.VIR_CPU_COMPARE_ERROR)
	CPU_COMPARE_INCOMPATIBLE = CPUCompareResult(C.VIR_CPU_COMPARE_INCOMPATIBLE)
	CPU_COMPARE_IDENTICAL    = CPUCompareResult(C.VIR_CPU_COMPARE_IDENTICAL)
	CPU_COMPARE_SUPERSET     = CPUCompareResult(C.VIR_CPU_COMPARE_SUPERSET)
)

type NodeAllocPagesFlags int

const (
	NODE_ALLOC_PAGES_ADD = NodeAllocPagesFlags(C.VIR_NODE_ALLOC_PAGES_ADD)
	NODE_ALLOC_PAGES_SET = NodeAllocPagesFlags(C.VIR_NODE_ALLOC_PAGES_SET)
)

type NodeSuspendTarget int

const (
	NODE_SUSPEND_TARGET_MEM    = NodeSuspendTarget(C.VIR_NODE_SUSPEND_TARGET_MEM)
	NODE_SUSPEND_TARGET_DISK   = NodeSuspendTarget(C.VIR_NODE_SUSPEND_TARGET_DISK)
	NODE_SUSPEND_TARGET_HYBRID = NodeSuspendTarget(C.VIR_NODE_SUSPEND_TARGET_HYBRID)
)

type NodeGetCPUStatsAllCPUs int

const (
	NODE_CPU_STATS_ALL_CPUS = NodeGetCPUStatsAllCPUs(C.VIR_NODE_CPU_STATS_ALL_CPUS)
)

const (
	NODE_MEMORY_STATS_ALL_CELLS = int(C.VIR_NODE_MEMORY_STATS_ALL_CELLS)
)

type ConnectCredentialType int

const (
	CRED_USERNAME     = ConnectCredentialType(C.VIR_CRED_USERNAME)
	CRED_AUTHNAME     = ConnectCredentialType(C.VIR_CRED_AUTHNAME)
	CRED_LANGUAGE     = ConnectCredentialType(C.VIR_CRED_LANGUAGE)
	CRED_CNONCE       = ConnectCredentialType(C.VIR_CRED_CNONCE)
	CRED_PASSPHRASE   = ConnectCredentialType(C.VIR_CRED_PASSPHRASE)
	CRED_ECHOPROMPT   = ConnectCredentialType(C.VIR_CRED_ECHOPROMPT)
	CRED_NOECHOPROMPT = ConnectCredentialType(C.VIR_CRED_NOECHOPROMPT)
	CRED_REALM        = ConnectCredentialType(C.VIR_CRED_REALM)
	CRED_EXTERNAL     = ConnectCredentialType(C.VIR_CRED_EXTERNAL)
)

type Connect struct {
	ptr C.virConnectPtr
}

type NodeInfo struct {
	Model   string
	Memory  uint64
	Cpus    uint
	MHz     uint
	Nodes   uint32
	Sockets uint32
	Cores   uint32
	Threads uint32
}

// Additional data associated to the connection.
type virConnectionData struct {
	errCallbackId   *int
	closeCallbackId *int
}

var connections map[C.virConnectPtr]*virConnectionData
var connectionsLock sync.RWMutex

func init() {
	connections = make(map[C.virConnectPtr]*virConnectionData)
}

func saveConnectionData(c *Connect, d *virConnectionData) {
	if c.ptr == nil {
		return // Or panic?
	}
	connectionsLock.Lock()
	defer connectionsLock.Unlock()
	connections[c.ptr] = d
}

func getConnectionData(c *Connect) *virConnectionData {
	connectionsLock.RLock()
	d := connections[c.ptr]
	connectionsLock.RUnlock()
	if d != nil {
		return d
	}
	d = &virConnectionData{}
	saveConnectionData(c, d)
	return d
}

func releaseConnectionData(c *Connect) {
	if c.ptr == nil {
		return
	}
	connectionsLock.Lock()
	defer connectionsLock.Unlock()
	delete(connections, c.ptr)
}

func GetVersion() (uint32, error) {
	var version C.ulong
	if err := C.virGetVersion(&version, nil, nil); err < 0 {
		return 0, GetLastError()
	}
	return uint32(version), nil
}

func NewConnect(uri string) (*Connect, error) {
	var cUri *C.char
	if uri != "" {
		cUri = C.CString(uri)
		defer C.free(unsafe.Pointer(cUri))
	}
	ptr := C.virConnectOpen(cUri)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Connect{ptr: ptr}, nil
}

type ConnectCredential struct {
	Type      ConnectCredentialType
	Prompt    string
	Challenge string
	DefResult string
	Result    string
	ResultLen int
}

type ConnectAuthCallback func(creds []*ConnectCredential)

type ConnectAuth struct {
	CredType []ConnectCredentialType
	Callback ConnectAuthCallback
}

//export connectAuthCallback
func connectAuthCallback(ccredlist C.virConnectCredentialPtr, ncred C.uint, callbackID C.int) C.int {
	cred := make([]*ConnectCredential, int(ncred))

	for i := 0; i < int(ncred); i++ {
		ccred := (C.virConnectCredentialPtr)(unsafe.Pointer((uintptr)(unsafe.Pointer(ccredlist)) + (unsafe.Sizeof(*ccredlist) * uintptr(i))))
		cred[i] = &ConnectCredential{
			Type:      ConnectCredentialType(ccred._type),
			Prompt:    C.GoString(ccred.prompt),
			Challenge: C.GoString(ccred.challenge),
			DefResult: C.GoString(ccred.defresult),
			ResultLen: -1,
		}
	}
	callbackEntry := getCallbackId(int(callbackID))
	callback, ok := callbackEntry.(ConnectAuthCallback)
	if !ok {
		panic("Unexpected callback type")
	}

	callback(cred)

	for i := 0; i < int(ncred); i++ {
		ccred := (C.virConnectCredentialPtr)(unsafe.Pointer((uintptr)(unsafe.Pointer(ccredlist)) + (unsafe.Sizeof(*ccredlist) * uintptr(i))))
		if cred[i].ResultLen >= 0 {
			ccred.result = C.CString(cred[i].Result)
			ccred.resultlen = C.uint(cred[i].ResultLen)
		}
	}

	return 0
}

func NewConnectWithAuth(uri string, auth *ConnectAuth, flags ConnectFlags) (*Connect, error) {
	var cUri *C.char

	ccredtype := make([]C.int, len(auth.CredType))

	for i := 0; i < len(auth.CredType); i++ {
		ccredtype[i] = C.int(auth.CredType[i])
	}

	if uri != "" {
		cUri = C.CString(uri)
		defer C.free(unsafe.Pointer(cUri))
	}

	callbackID := registerCallbackId(auth.Callback)

	ptr := C.virConnectOpenAuthWrap(cUri, &ccredtype[0], C.uint(len(auth.CredType)), C.int(callbackID), C.uint(flags))
	freeCallbackId(callbackID)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Connect{ptr: ptr}, nil
}

func NewConnectReadOnly(uri string) (*Connect, error) {
	var cUri *C.char
	if uri != "" {
		cUri = C.CString(uri)
		defer C.free(unsafe.Pointer(cUri))
	}
	ptr := C.virConnectOpenReadOnly(cUri)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Connect{ptr: ptr}, nil
}

func (c *Connect) Close() (int, error) {
	result := int(C.virConnectClose(c.ptr))
	if result == -1 {
		return result, GetLastError()
	}
	if result == 0 {
		// No more reference to this connection, release data.
		releaseConnectionData(c)
		c.ptr = nil
	}
	return result, nil
}

func (c *Connect) Ref() error {
	ret := C.virConnectRef(c.ptr)
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

type CloseCallback func(conn *Connect, reason ConnectCloseReason)

// Register a close callback for the given destination. Only one
// callback per connection is allowed. Setting a callback will remove
// the previous one.
func (c *Connect) RegisterCloseCallback(callback CloseCallback) error {
	c.UnregisterCloseCallback()
	goCallbackId := registerCallbackId(callback)
	callbackPtr := unsafe.Pointer(C.closeCallback_cgo)
	res := C.virConnectRegisterCloseCallback_cgo(c.ptr, C.virConnectCloseFunc(callbackPtr), C.long(goCallbackId))
	if res != 0 {
		freeCallbackId(goCallbackId)
		return GetLastError()
	}
	connData := getConnectionData(c)
	connData.closeCallbackId = &goCallbackId
	return nil
}

func (c *Connect) UnregisterCloseCallback() error {
	connData := getConnectionData(c)
	if connData.closeCallbackId == nil {
		return nil
	}
	callbackPtr := unsafe.Pointer(C.closeCallback_cgo)
	res := C.virConnectUnregisterCloseCallback(c.ptr, C.virConnectCloseFunc(callbackPtr))
	if res != 0 {
		return GetLastError()
	}
	connData.closeCallbackId = nil
	return nil
}

//export closeCallback
func closeCallback(conn C.virConnectPtr, reason ConnectCloseReason, goCallbackId int) {
	callbackFunc := getCallbackId(goCallbackId)
	callback, ok := callbackFunc.(CloseCallback)
	if !ok {
		panic("Inappropriate callback type called")
	}
	callback(&Connect{ptr: conn}, reason)
}

func (c *Connect) GetCapabilities() (string, error) {
	str := C.virConnectGetCapabilities(c.ptr)
	if str == nil {
		return "", GetLastError()
	}
	capabilities := C.GoString(str)
	C.free(unsafe.Pointer(str))
	return capabilities, nil
}

func (c *Connect) GetNodeInfo() (*NodeInfo, error) {
	var cinfo C.virNodeInfo
	result := C.virNodeGetInfo(c.ptr, &cinfo)
	if result == -1 {
		return nil, GetLastError()
	}
	return &NodeInfo{
		Model:   C.GoString((*C.char)(unsafe.Pointer(&cinfo.model[0]))),
		Memory:  uint64(cinfo.memory),
		Cpus:    uint(cinfo.cpus),
		MHz:     uint(cinfo.mhz),
		Nodes:   uint32(cinfo.nodes),
		Sockets: uint32(cinfo.sockets),
		Cores:   uint32(cinfo.cores),
		Threads: uint32(cinfo.threads),
	}, nil
}

func (ni *NodeInfo) GetMaxCPUs() uint32 {
	return ni.Nodes * ni.Sockets * ni.Cores * ni.Threads
}

func (c *Connect) GetHostname() (string, error) {
	str := C.virConnectGetHostname(c.ptr)
	if str == nil {
		return "", GetLastError()
	}
	hostname := C.GoString(str)
	C.free(unsafe.Pointer(str))
	return hostname, nil
}

func (c *Connect) GetLibVersion() (uint32, error) {
	var version C.ulong
	if err := C.virConnectGetLibVersion(c.ptr, &version); err < 0 {
		return 0, GetLastError()
	}
	return uint32(version), nil
}

func (c *Connect) GetType() (string, error) {
	str := C.virConnectGetType(c.ptr)
	if str == nil {
		return "", GetLastError()
	}
	hypDriver := C.GoString(str)
	return hypDriver, nil
}

func (c *Connect) IsAlive() (bool, error) {
	result := C.virConnectIsAlive(c.ptr)
	if result == -1 {
		return false, GetLastError()
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (c *Connect) IsEncrypted() (bool, error) {
	result := C.virConnectIsEncrypted(c.ptr)
	if result == -1 {
		return false, GetLastError()
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (c *Connect) IsSecure() (bool, error) {
	result := C.virConnectIsSecure(c.ptr)
	if result == -1 {
		return false, GetLastError()
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (c *Connect) ListDefinedDomains() ([]string, error) {
	var names [1024](*C.char)
	namesPtr := unsafe.Pointer(&names)
	numDomains := C.virConnectListDefinedDomains(
		c.ptr,
		(**C.char)(namesPtr),
		1024)
	if numDomains == -1 {
		return nil, GetLastError()
	}
	goNames := make([]string, numDomains)
	for k := 0; k < int(numDomains); k++ {
		goNames[k] = C.GoString(names[k])
		C.free(unsafe.Pointer(names[k]))
	}
	return goNames, nil
}

func (c *Connect) ListDomains() ([]uint32, error) {
	var cDomainsIds [512](uint32)
	cDomainsPointer := unsafe.Pointer(&cDomainsIds)
	numDomains := C.virConnectListDomains(c.ptr, (*C.int)(cDomainsPointer), 512)
	if numDomains == -1 {
		return nil, GetLastError()
	}

	return cDomainsIds[:numDomains], nil
}

func (c *Connect) ListInterfaces() ([]string, error) {
	const maxIfaces = 1024
	var names [maxIfaces](*C.char)
	namesPtr := unsafe.Pointer(&names)
	numIfaces := C.virConnectListInterfaces(
		c.ptr,
		(**C.char)(namesPtr),
		maxIfaces)
	if numIfaces == -1 {
		return nil, GetLastError()
	}
	goNames := make([]string, numIfaces)
	for k := 0; k < int(numIfaces); k++ {
		goNames[k] = C.GoString(names[k])
		C.free(unsafe.Pointer(names[k]))
	}
	return goNames, nil
}

func (c *Connect) ListNetworks() ([]string, error) {
	const maxNets = 1024
	var names [maxNets](*C.char)
	namesPtr := unsafe.Pointer(&names)
	numNetworks := C.virConnectListNetworks(
		c.ptr,
		(**C.char)(namesPtr),
		maxNets)
	if numNetworks == -1 {
		return nil, GetLastError()
	}
	goNames := make([]string, numNetworks)
	for k := 0; k < int(numNetworks); k++ {
		goNames[k] = C.GoString(names[k])
		C.free(unsafe.Pointer(names[k]))
	}
	return goNames, nil
}

func (c *Connect) ListNWFilters() ([]string, error) {
	const maxFilters = 1024
	var names [maxFilters](*C.char)
	namesPtr := unsafe.Pointer(&names)
	numNWFilters := C.virConnectListNWFilters(
		c.ptr,
		(**C.char)(namesPtr),
		maxFilters)
	if numNWFilters == -1 {
		return nil, GetLastError()
	}
	goNames := make([]string, numNWFilters)
	for k := 0; k < int(numNWFilters); k++ {
		goNames[k] = C.GoString(names[k])
		C.free(unsafe.Pointer(names[k]))
	}
	return goNames, nil
}

func (c *Connect) ListStoragePools() ([]string, error) {
	const maxPools = 1024
	var names [maxPools](*C.char)
	namesPtr := unsafe.Pointer(&names)
	numStoragePools := C.virConnectListStoragePools(
		c.ptr,
		(**C.char)(namesPtr),
		maxPools)
	if numStoragePools == -1 {
		return nil, GetLastError()
	}
	goNames := make([]string, numStoragePools)
	for k := 0; k < int(numStoragePools); k++ {
		goNames[k] = C.GoString(names[k])
		C.free(unsafe.Pointer(names[k]))
	}
	return goNames, nil
}

func (c *Connect) ListSecrets() ([]string, error) {
	const maxSecrets = 1024
	var uuids [maxSecrets](*C.char)
	uuidsPtr := unsafe.Pointer(&uuids)
	numSecrets := C.virConnectListSecrets(
		c.ptr,
		(**C.char)(uuidsPtr),
		maxSecrets)
	if numSecrets == -1 {
		return nil, GetLastError()
	}
	goUuids := make([]string, numSecrets)
	for k := 0; k < int(numSecrets); k++ {
		goUuids[k] = C.GoString(uuids[k])
		C.free(unsafe.Pointer(uuids[k]))
	}
	return goUuids, nil
}

func (c *Connect) ListDevices(cap string, flags uint32) ([]string, error) {
	ccap := C.CString(cap)
	defer C.free(unsafe.Pointer(ccap))
	const maxNodeDevices = 1024
	var uuids [maxNodeDevices](*C.char)
	uuidsPtr := unsafe.Pointer(&uuids)
	numNodeDevices := C.virNodeListDevices(
		c.ptr, ccap,
		(**C.char)(uuidsPtr),
		maxNodeDevices, C.uint(flags))
	if numNodeDevices == -1 {
		return nil, GetLastError()
	}
	goUuids := make([]string, numNodeDevices)
	for k := 0; k < int(numNodeDevices); k++ {
		goUuids[k] = C.GoString(uuids[k])
		C.free(unsafe.Pointer(uuids[k]))
	}
	return goUuids, nil
}

func (c *Connect) LookupDomainById(id uint32) (*Domain, error) {
	ptr := C.virDomainLookupByID(c.ptr, C.int(id))
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Domain{ptr: ptr}, nil
}

func (c *Connect) LookupDomainByName(id string) (*Domain, error) {
	cName := C.CString(id)
	defer C.free(unsafe.Pointer(cName))
	ptr := C.virDomainLookupByName(c.ptr, cName)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Domain{ptr: ptr}, nil
}

func (c *Connect) LookupDomainByUUIDString(uuid string) (*Domain, error) {
	cUuid := C.CString(uuid)
	defer C.free(unsafe.Pointer(cUuid))
	ptr := C.virDomainLookupByUUIDString(c.ptr, cUuid)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Domain{ptr: ptr}, nil
}

func (c *Connect) LookupDomainByUUID(uuid []byte) (*Domain, error) {
	if len(uuid) != C.VIR_UUID_BUFLEN {
		return nil, fmt.Errorf("UUID must be exactly %d bytes in size",
			int(C.VIR_UUID_BUFLEN))
	}
	cUuid := make([]C.uchar, C.VIR_UUID_BUFLEN)
	for i := 0; i < C.VIR_UUID_BUFLEN; i++ {
		cUuid[i] = C.uchar(uuid[i])
	}
	ptr := C.virDomainLookupByUUID(c.ptr, &cUuid[0])
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Domain{ptr: ptr}, nil
}

func (c *Connect) DomainCreateXML(xmlConfig string, flags DomainCreateFlags) (*Domain, error) {
	cXml := C.CString(string(xmlConfig))
	defer C.free(unsafe.Pointer(cXml))
	ptr := C.virDomainCreateXML(c.ptr, cXml, C.uint(flags))
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Domain{ptr: ptr}, nil
}

func (c *Connect) DomainCreateXMLWithFiles(xmlConfig string, files []os.File, flags DomainCreateFlags) (*Domain, error) {
	cXml := C.CString(string(xmlConfig))
	defer C.free(unsafe.Pointer(cXml))
	cfiles := make([]C.int, len(files))
	for i := 0; i < len(files); i++ {
		cfiles[i] = C.int(files[i].Fd())
	}
	ptr := C.virDomainCreateXMLWithFiles(c.ptr, cXml, C.uint(len(files)), (&cfiles[0]), C.uint(flags))
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Domain{ptr: ptr}, nil
}

func (c *Connect) DomainDefineXML(xmlConfig string) (*Domain, error) {
	cXml := C.CString(string(xmlConfig))
	defer C.free(unsafe.Pointer(cXml))
	ptr := C.virDomainDefineXML(c.ptr, cXml)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Domain{ptr: ptr}, nil
}

func (c *Connect) DomainDefineXMLFlags(xmlConfig string, flags DomainDefineFlags) (*Domain, error) {
	if C.LIBVIR_VERSION_NUMBER < 1002012 {
		return nil, GetNotImplementedError("virDomainDefineXMLFlags")
	}
	cXml := C.CString(string(xmlConfig))
	defer C.free(unsafe.Pointer(cXml))
	ptr := C.virDomainDefineXMLFlagsCompat(c.ptr, cXml, C.uint(flags))
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Domain{ptr: ptr}, nil
}

func (c *Connect) ListDefinedInterfaces() ([]string, error) {
	const maxIfaces = 1024
	var names [maxIfaces](*C.char)
	namesPtr := unsafe.Pointer(&names)
	numIfaces := C.virConnectListDefinedInterfaces(
		c.ptr,
		(**C.char)(namesPtr),
		maxIfaces)
	if numIfaces == -1 {
		return nil, GetLastError()
	}
	goNames := make([]string, numIfaces)
	for k := 0; k < int(numIfaces); k++ {
		goNames[k] = C.GoString(names[k])
		C.free(unsafe.Pointer(names[k]))
	}
	return goNames, nil
}

func (c *Connect) ListDefinedNetworks() ([]string, error) {
	const maxNets = 1024
	var names [maxNets](*C.char)
	namesPtr := unsafe.Pointer(&names)
	numNetworks := C.virConnectListDefinedNetworks(
		c.ptr,
		(**C.char)(namesPtr),
		maxNets)
	if numNetworks == -1 {
		return nil, GetLastError()
	}
	goNames := make([]string, numNetworks)
	for k := 0; k < int(numNetworks); k++ {
		goNames[k] = C.GoString(names[k])
		C.free(unsafe.Pointer(names[k]))
	}
	return goNames, nil
}

func (c *Connect) ListDefinedStoragePools() ([]string, error) {
	const maxPools = 1024
	var names [maxPools](*C.char)
	namesPtr := unsafe.Pointer(&names)
	numStoragePools := C.virConnectListDefinedStoragePools(
		c.ptr,
		(**C.char)(namesPtr),
		maxPools)
	if numStoragePools == -1 {
		return nil, GetLastError()
	}
	goNames := make([]string, numStoragePools)
	for k := 0; k < int(numStoragePools); k++ {
		goNames[k] = C.GoString(names[k])
		C.free(unsafe.Pointer(names[k]))
	}
	return goNames, nil
}

func (c *Connect) NumOfDefinedDomains() (int, error) {
	result := int(C.virConnectNumOfDefinedDomains(c.ptr))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (c *Connect) NumOfDefinedInterfaces() (int, error) {
	result := int(C.virConnectNumOfDefinedInterfaces(c.ptr))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (c *Connect) NumOfDefinedNetworks() (int, error) {
	result := int(C.virConnectNumOfDefinedNetworks(c.ptr))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (c *Connect) NumOfDefinedStoragePools() (int, error) {
	result := int(C.virConnectNumOfDefinedStoragePools(c.ptr))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (c *Connect) NumOfDomains() (int, error) {
	result := int(C.virConnectNumOfDomains(c.ptr))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (c *Connect) NumOfStoragePools() (int, error) {
	result := int(C.virConnectNumOfStoragePools(c.ptr))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (c *Connect) NumOfInterfaces() (int, error) {
	result := int(C.virConnectNumOfInterfaces(c.ptr))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (c *Connect) NumOfNetworks() (int, error) {
	result := int(C.virConnectNumOfNetworks(c.ptr))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (c *Connect) NumOfNWFilters() (int, error) {
	result := int(C.virConnectNumOfNWFilters(c.ptr))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (c *Connect) NumOfSecrets() (int, error) {
	result := int(C.virConnectNumOfSecrets(c.ptr))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (c *Connect) NumOfDevices(cap string, flags uint32) (int, error) {
	ccap := C.CString(cap)
	defer C.free(unsafe.Pointer(ccap))
	result := int(C.virNodeNumOfDevices(c.ptr, ccap, C.uint(flags)))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (c *Connect) NetworkDefineXML(xmlConfig string) (*Network, error) {
	cXml := C.CString(string(xmlConfig))
	defer C.free(unsafe.Pointer(cXml))
	ptr := C.virNetworkDefineXML(c.ptr, cXml)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Network{ptr: ptr}, nil
}

func (c *Connect) NetworkCreateXML(xmlConfig string) (*Network, error) {
	cXml := C.CString(string(xmlConfig))
	defer C.free(unsafe.Pointer(cXml))
	ptr := C.virNetworkCreateXML(c.ptr, cXml)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Network{ptr: ptr}, nil
}

func (c *Connect) LookupNetworkByName(name string) (*Network, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	ptr := C.virNetworkLookupByName(c.ptr, cName)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Network{ptr: ptr}, nil
}

func (c *Connect) LookupNetworkByUUIDString(uuid string) (*Network, error) {
	cUuid := C.CString(uuid)
	defer C.free(unsafe.Pointer(cUuid))
	ptr := C.virNetworkLookupByUUIDString(c.ptr, cUuid)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Network{ptr: ptr}, nil
}

func (c *Connect) LookupNetworkByUUID(uuid []byte) (*Network, error) {
	if len(uuid) != C.VIR_UUID_BUFLEN {
		return nil, fmt.Errorf("UUID must be exactly %d bytes in size",
			int(C.VIR_UUID_BUFLEN))
	}
	cUuid := make([]C.uchar, C.VIR_UUID_BUFLEN)
	for i := 0; i < C.VIR_UUID_BUFLEN; i++ {
		cUuid[i] = C.uchar(uuid[i])
	}
	ptr := C.virNetworkLookupByUUID(c.ptr, &cUuid[0])
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Network{ptr: ptr}, nil
}

func (c *Connect) SetKeepAlive(interval int, count uint) error {
	res := int(C.virConnectSetKeepAlive(c.ptr, C.int(interval), C.uint(count)))
	switch res {
	case 0:
		return nil
	default:
		return GetLastError()
	}
}

func (c *Connect) GetSysinfo(flags uint32) (string, error) {
	cStr := C.virConnectGetSysinfo(c.ptr, C.uint(flags))
	if cStr == nil {
		return "", GetLastError()
	}
	info := C.GoString(cStr)
	C.free(unsafe.Pointer(cStr))
	return info, nil
}

func (c *Connect) GetURI() (string, error) {
	cStr := C.virConnectGetURI(c.ptr)
	if cStr == nil {
		return "", GetLastError()
	}
	uri := C.GoString(cStr)
	C.free(unsafe.Pointer(cStr))
	return uri, nil
}

func (c *Connect) GetMaxVcpus(typeAttr string) (int, error) {
	var cTypeAttr *C.char
	if typeAttr != "" {
		cTypeAttr = C.CString(typeAttr)
		defer C.free(unsafe.Pointer(cTypeAttr))
	}
	result := int(C.virConnectGetMaxVcpus(c.ptr, cTypeAttr))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (c *Connect) InterfaceDefineXML(xmlConfig string, flags uint32) (*Interface, error) {
	cXml := C.CString(string(xmlConfig))
	defer C.free(unsafe.Pointer(cXml))
	ptr := C.virInterfaceDefineXML(c.ptr, cXml, C.uint(flags))
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Interface{ptr: ptr}, nil
}

func (c *Connect) LookupInterfaceByName(name string) (*Interface, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	ptr := C.virInterfaceLookupByName(c.ptr, cName)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Interface{ptr: ptr}, nil
}

func (c *Connect) LookupInterfaceByMACString(mac string) (*Interface, error) {
	cName := C.CString(mac)
	defer C.free(unsafe.Pointer(cName))
	ptr := C.virInterfaceLookupByMACString(c.ptr, cName)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Interface{ptr: ptr}, nil
}

func (c *Connect) StoragePoolDefineXML(xmlConfig string, flags uint32) (*StoragePool, error) {
	cXml := C.CString(string(xmlConfig))
	defer C.free(unsafe.Pointer(cXml))
	ptr := C.virStoragePoolDefineXML(c.ptr, cXml, C.uint(flags))
	if ptr == nil {
		return nil, GetLastError()
	}
	return &StoragePool{ptr: ptr}, nil
}

func (c *Connect) StoragePoolCreateXML(xmlConfig string, flags StoragePoolCreateFlags) (*StoragePool, error) {
	cXml := C.CString(string(xmlConfig))
	defer C.free(unsafe.Pointer(cXml))
	ptr := C.virStoragePoolCreateXML(c.ptr, cXml, C.uint(flags))
	if ptr == nil {
		return nil, GetLastError()
	}
	return &StoragePool{ptr: ptr}, nil
}

func (c *Connect) LookupStoragePoolByName(name string) (*StoragePool, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	ptr := C.virStoragePoolLookupByName(c.ptr, cName)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &StoragePool{ptr: ptr}, nil
}

func (c *Connect) LookupStoragePoolByUUIDString(uuid string) (*StoragePool, error) {
	cUuid := C.CString(uuid)
	defer C.free(unsafe.Pointer(cUuid))
	ptr := C.virStoragePoolLookupByUUIDString(c.ptr, cUuid)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &StoragePool{ptr: ptr}, nil
}

func (c *Connect) LookupStoragePoolByUUID(uuid []byte) (*StoragePool, error) {
	if len(uuid) != C.VIR_UUID_BUFLEN {
		return nil, fmt.Errorf("UUID must be exactly %d bytes in size",
			int(C.VIR_UUID_BUFLEN))
	}
	cUuid := make([]C.uchar, C.VIR_UUID_BUFLEN)
	for i := 0; i < C.VIR_UUID_BUFLEN; i++ {
		cUuid[i] = C.uchar(uuid[i])
	}
	ptr := C.virStoragePoolLookupByUUID(c.ptr, &cUuid[0])
	if ptr == nil {
		return nil, GetLastError()
	}
	return &StoragePool{ptr: ptr}, nil
}

func (c *Connect) NWFilterDefineXML(xmlConfig string) (*NWFilter, error) {
	cXml := C.CString(string(xmlConfig))
	defer C.free(unsafe.Pointer(cXml))
	ptr := C.virNWFilterDefineXML(c.ptr, cXml)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &NWFilter{ptr: ptr}, nil
}

func (c *Connect) LookupNWFilterByName(name string) (*NWFilter, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	ptr := C.virNWFilterLookupByName(c.ptr, cName)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &NWFilter{ptr: ptr}, nil
}

func (c *Connect) LookupNWFilterByUUIDString(uuid string) (*NWFilter, error) {
	cUuid := C.CString(uuid)
	defer C.free(unsafe.Pointer(cUuid))
	ptr := C.virNWFilterLookupByUUIDString(c.ptr, cUuid)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &NWFilter{ptr: ptr}, nil
}

func (c *Connect) LookupNWFilterByUUID(uuid []byte) (*NWFilter, error) {
	if len(uuid) != C.VIR_UUID_BUFLEN {
		return nil, fmt.Errorf("UUID must be exactly %d bytes in size",
			int(C.VIR_UUID_BUFLEN))
	}
	cUuid := make([]C.uchar, C.VIR_UUID_BUFLEN)
	for i := 0; i < C.VIR_UUID_BUFLEN; i++ {
		cUuid[i] = C.uchar(uuid[i])
	}
	ptr := C.virNWFilterLookupByUUID(c.ptr, &cUuid[0])
	if ptr == nil {
		return nil, GetLastError()
	}
	return &NWFilter{ptr: ptr}, nil
}

func (c *Connect) LookupStorageVolByKey(key string) (*StorageVol, error) {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	ptr := C.virStorageVolLookupByKey(c.ptr, cKey)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &StorageVol{ptr: ptr}, nil
}

func (c *Connect) LookupStorageVolByPath(path string) (*StorageVol, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	ptr := C.virStorageVolLookupByPath(c.ptr, cPath)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &StorageVol{ptr: ptr}, nil
}

func (c *Connect) SecretDefineXML(xmlConfig string, flags uint32) (*Secret, error) {
	cXml := C.CString(string(xmlConfig))
	defer C.free(unsafe.Pointer(cXml))
	ptr := C.virSecretDefineXML(c.ptr, cXml, C.uint(flags))
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Secret{ptr: ptr}, nil
}

func (c *Connect) LookupSecretByUUID(uuid []byte) (*Secret, error) {
	if len(uuid) != C.VIR_UUID_BUFLEN {
		return nil, fmt.Errorf("UUID must be exactly %d bytes in size",
			int(C.VIR_UUID_BUFLEN))
	}
	cUuid := make([]C.uchar, C.VIR_UUID_BUFLEN)
	for i := 0; i < C.VIR_UUID_BUFLEN; i++ {
		cUuid[i] = C.uchar(uuid[i])
	}
	ptr := C.virSecretLookupByUUID(c.ptr, &cUuid[0])
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Secret{ptr: ptr}, nil
}

func (c *Connect) LookupSecretByUUIDString(uuid string) (*Secret, error) {
	cUuid := C.CString(uuid)
	defer C.free(unsafe.Pointer(cUuid))
	ptr := C.virSecretLookupByUUIDString(c.ptr, cUuid)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Secret{ptr: ptr}, nil
}

func (c *Connect) LookupSecretByUsage(usageType SecretUsageType, usageID string) (*Secret, error) {
	cUsageID := C.CString(usageID)
	defer C.free(unsafe.Pointer(cUsageID))
	ptr := C.virSecretLookupByUsage(c.ptr, C.int(usageType), cUsageID)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &Secret{ptr: ptr}, nil
}

func (c *Connect) LookupDeviceByName(id string) (*NodeDevice, error) {
	cName := C.CString(id)
	defer C.free(unsafe.Pointer(cName))
	ptr := C.virNodeDeviceLookupByName(c.ptr, cName)
	if ptr == nil {
		return nil, GetLastError()
	}
	return &NodeDevice{ptr: ptr}, nil
}

func (c *Connect) LookupDeviceSCSIHostByWWN(wwnn, wwpn string, flags uint32) (*NodeDevice, error) {
	cWwnn := C.CString(wwnn)
	cWwpn := C.CString(wwpn)
	defer C.free(unsafe.Pointer(cWwnn))
	defer C.free(unsafe.Pointer(cWwpn))
	ptr := C.virNodeDeviceLookupSCSIHostByWWN(c.ptr, cWwnn, cWwpn, C.uint(flags))
	if ptr == nil {
		return nil, GetLastError()
	}
	return &NodeDevice{ptr: ptr}, nil
}

func (c *Connect) DeviceCreateXML(xmlConfig string, flags uint32) (*NodeDevice, error) {
	cXml := C.CString(string(xmlConfig))
	defer C.free(unsafe.Pointer(cXml))
	ptr := C.virNodeDeviceCreateXML(c.ptr, cXml, C.uint(flags))
	if ptr == nil {
		return nil, GetLastError()
	}
	return &NodeDevice{ptr: ptr}, nil
}

func (c *Connect) ListAllInterfaces(flags ConnectListAllInterfacesFlags) ([]Interface, error) {
	var cList *C.virInterfacePtr
	numIfaces := C.virConnectListAllInterfaces(c.ptr, (**C.virInterfacePtr)(&cList), C.uint(flags))
	if numIfaces == -1 {
		return nil, GetLastError()
	}
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cList)),
		Len:  int(numIfaces),
		Cap:  int(numIfaces),
	}
	var ifaces []Interface
	slice := *(*[]C.virInterfacePtr)(unsafe.Pointer(&hdr))
	for _, ptr := range slice {
		ifaces = append(ifaces, Interface{ptr})
	}
	C.free(unsafe.Pointer(cList))
	return ifaces, nil
}

func (c *Connect) ListAllNetworks(flags ConnectListAllNetworksFlags) ([]Network, error) {
	var cList *C.virNetworkPtr
	numNets := C.virConnectListAllNetworks(c.ptr, (**C.virNetworkPtr)(&cList), C.uint(flags))
	if numNets == -1 {
		return nil, GetLastError()
	}
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cList)),
		Len:  int(numNets),
		Cap:  int(numNets),
	}
	var nets []Network
	slice := *(*[]C.virNetworkPtr)(unsafe.Pointer(&hdr))
	for _, ptr := range slice {
		nets = append(nets, Network{ptr})
	}
	C.free(unsafe.Pointer(cList))
	return nets, nil
}

func (c *Connect) ListAllDomains(flags ConnectListAllDomainsFlags) ([]Domain, error) {
	var cList *C.virDomainPtr
	numDomains := C.virConnectListAllDomains(c.ptr, (**C.virDomainPtr)(&cList), C.uint(flags))
	if numDomains == -1 {
		return nil, GetLastError()
	}
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cList)),
		Len:  int(numDomains),
		Cap:  int(numDomains),
	}
	var domains []Domain
	slice := *(*[]C.virDomainPtr)(unsafe.Pointer(&hdr))
	for _, ptr := range slice {
		domains = append(domains, Domain{ptr})
	}
	C.free(unsafe.Pointer(cList))
	return domains, nil
}

func (c *Connect) ListAllNWFilters(flags uint32) ([]NWFilter, error) {
	var cList *C.virNWFilterPtr
	numNWFilters := C.virConnectListAllNWFilters(c.ptr, (**C.virNWFilterPtr)(&cList), C.uint(flags))
	if numNWFilters == -1 {
		return nil, GetLastError()
	}
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cList)),
		Len:  int(numNWFilters),
		Cap:  int(numNWFilters),
	}
	var filters []NWFilter
	slice := *(*[]C.virNWFilterPtr)(unsafe.Pointer(&hdr))
	for _, ptr := range slice {
		filters = append(filters, NWFilter{ptr})
	}
	C.free(unsafe.Pointer(cList))
	return filters, nil
}

func (c *Connect) ListAllStoragePools(flags ConnectListAllStoragePoolsFlags) ([]StoragePool, error) {
	var cList *C.virStoragePoolPtr
	numPools := C.virConnectListAllStoragePools(c.ptr, (**C.virStoragePoolPtr)(&cList), C.uint(flags))
	if numPools == -1 {
		return nil, GetLastError()
	}
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cList)),
		Len:  int(numPools),
		Cap:  int(numPools),
	}
	var pools []StoragePool
	slice := *(*[]C.virStoragePoolPtr)(unsafe.Pointer(&hdr))
	for _, ptr := range slice {
		pools = append(pools, StoragePool{ptr})
	}
	C.free(unsafe.Pointer(cList))
	return pools, nil
}

func (c *Connect) ListAllSecrets(flags ConnectListAllSecretsFlags) ([]Secret, error) {
	var cList *C.virSecretPtr
	numPools := C.virConnectListAllSecrets(c.ptr, (**C.virSecretPtr)(&cList), C.uint(flags))
	if numPools == -1 {
		return nil, GetLastError()
	}
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cList)),
		Len:  int(numPools),
		Cap:  int(numPools),
	}
	var pools []Secret
	slice := *(*[]C.virSecretPtr)(unsafe.Pointer(&hdr))
	for _, ptr := range slice {
		pools = append(pools, Secret{ptr})
	}
	C.free(unsafe.Pointer(cList))
	return pools, nil
}

func (c *Connect) ListAllNodeDevices(flags ConnectListAllNodeDeviceFlags) ([]NodeDevice, error) {
	var cList *C.virNodeDevicePtr
	numPools := C.virConnectListAllNodeDevices(c.ptr, (**C.virNodeDevicePtr)(&cList), C.uint(flags))
	if numPools == -1 {
		return nil, GetLastError()
	}
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cList)),
		Len:  int(numPools),
		Cap:  int(numPools),
	}
	var pools []NodeDevice
	slice := *(*[]C.virNodeDevicePtr)(unsafe.Pointer(&hdr))
	for _, ptr := range slice {
		pools = append(pools, NodeDevice{ptr})
	}
	C.free(unsafe.Pointer(cList))
	return pools, nil
}

func (c *Connect) InterfaceChangeBegin(flags uint32) error {
	ret := C.virInterfaceChangeBegin(c.ptr, C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (c *Connect) InterfaceChangeCommit(flags uint32) error {
	ret := C.virInterfaceChangeCommit(c.ptr, C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (c *Connect) InterfaceChangeRollback(flags uint32) error {
	ret := C.virInterfaceChangeRollback(c.ptr, C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (c *Connect) AllocPages(pageSizes map[int]int64, startCell int, cellCount uint, flags NodeAllocPagesFlags) (int, error) {
	if C.LIBVIR_VERSION_NUMBER < 1002009 {
		return 0, GetNotImplementedError("virNodeAllocPages")
	}
	cpages := make([]C.uint, len(pageSizes))
	ccounts := make([]C.ulonglong, len(pageSizes))

	i := 0
	for key, val := range pageSizes {
		cpages[i] = C.uint(key)
		ccounts[i] = C.ulonglong(val)
		i++
	}

	ret := C.virNodeAllocPagesCompat(c.ptr, C.uint(len(pageSizes)), (*C.uint)(unsafe.Pointer(&cpages)),
		(*C.ulonglong)(unsafe.Pointer(&ccounts)), C.int(startCell), C.uint(cellCount), C.uint(flags))
	if ret == -1 {
		return 0, GetLastError()
	}

	return int(ret), nil
}

func (c *Connect) GetCPUMap(flags uint32) (map[int]bool, uint, error) {
	var ccpumap *C.uchar
	var conline C.uint
	ret := C.virNodeGetCPUMap(c.ptr, &ccpumap, &conline, C.uint(flags))
	if ret == -1 {
		return map[int]bool{}, 0, GetLastError()
	}
	defer C.free(unsafe.Pointer(ccpumap))

	cpumapbytes := C.GoBytes(unsafe.Pointer(ccpumap), C.int(ret/8))

	cpumap := make(map[int]bool, 0)
	for i := 0; i < int(ret); i++ {
		idx := int(i / 8)
		val := byte(cpumapbytes[idx])
		shift := i % 8
		cpumap[i] = (val & (1 << uint(shift))) == 1
	}

	return cpumap, uint(conline), nil
}

type NodeCPUStats struct {
	KernelSet      bool
	Kernel         uint64
	UserSet        bool
	User           uint64
	IdleSet        bool
	Idle           uint64
	IowaitSet      bool
	Iowait         uint64
	IntrSet        bool
	Intr           uint64
	UtilizationSet bool
	Utilization    uint64
}

func (c *Connect) GetCPUStats(cpuNum int, flags uint32) (*NodeCPUStats, error) {
	var nparams C.int

	ret := C.virNodeGetCPUStats(c.ptr, C.int(cpuNum), nil, &nparams, C.uint(0))
	if ret == -1 {
		return nil, GetLastError()
	}

	params := make([]C.virNodeCPUStats, nparams)
	ret = C.virNodeGetCPUStats(c.ptr, C.int(cpuNum), (*C.virNodeCPUStats)(unsafe.Pointer(&params[0])), &nparams, C.uint(flags))
	if ret == -1 {
		return nil, GetLastError()
	}

	stats := &NodeCPUStats{}
	for i := 0; i < int(nparams); i++ {
		param := params[i]
		field := C.GoString((*C.char)(unsafe.Pointer(&param.field)))
		switch field {
		case C.VIR_NODE_CPU_STATS_KERNEL:
			stats.KernelSet = true
			stats.Kernel = uint64(param.value)
		case C.VIR_NODE_CPU_STATS_USER:
			stats.UserSet = true
			stats.User = uint64(param.value)
		case C.VIR_NODE_CPU_STATS_IDLE:
			stats.IdleSet = true
			stats.Idle = uint64(param.value)
		case C.VIR_NODE_CPU_STATS_IOWAIT:
			stats.IowaitSet = true
			stats.Iowait = uint64(param.value)
		case C.VIR_NODE_CPU_STATS_INTR:
			stats.IntrSet = true
			stats.Intr = uint64(param.value)
		case C.VIR_NODE_CPU_STATS_UTILIZATION:
			stats.UtilizationSet = true
			stats.Utilization = uint64(param.value)
		}
	}

	return stats, nil
}

func (c *Connect) GetCellsFreeMemory(startCell int, maxCells int) ([]uint64, error) {
	cmem := make([]C.ulonglong, maxCells)
	ret := C.virNodeGetCellsFreeMemory(c.ptr, (*C.ulonglong)(unsafe.Pointer(&cmem[0])), C.int(startCell), C.int(maxCells))
	if ret == -1 {
		return []uint64{}, GetLastError()
	}

	mem := make([]uint64, ret)
	for i := 0; i < int(ret); i++ {
		mem[i] = uint64(cmem[i])
	}

	return mem, nil
}

func (c *Connect) GetFreeMemory() (uint64, error) {
	ret := C.virNodeGetFreeMemory(c.ptr)
	if ret == 0 {
		return 0, GetLastError()
	}

	return (uint64)(ret), nil
}

func (c *Connect) GetFreePages(pageSizes []uint64, startCell int, maxCells uint, flags uint32) ([]uint64, error) {
	if C.LIBVIR_VERSION_NUMBER < 1002006 {
		return []uint64{}, GetNotImplementedError("virNodeGetFreePages")
	}
	cpageSizes := make([]C.uint, len(pageSizes))
	ccounts := make([]C.ulonglong, len(pageSizes)*int(maxCells))

	for i := 0; i < len(pageSizes); i++ {
		cpageSizes[i] = C.uint(pageSizes[i])
	}

	ret := C.virNodeGetFreePagesCompat(c.ptr, C.uint(len(pageSizes)), (*C.uint)(unsafe.Pointer(&cpageSizes)), C.int(startCell),
		C.uint(maxCells), (*C.ulonglong)(unsafe.Pointer(&ccounts)), C.uint(flags))
	if ret == -1 {
		return []uint64{}, GetLastError()
	}

	counts := make([]uint64, ret)
	for i := 0; i < int(ret); i++ {
		counts[i] = uint64(ccounts[i])
	}

	return counts, nil
}

type NodeMemoryParameters struct {
	ShmPagesToScanSet      bool
	ShmPagesToScan         uint
	ShmSleepMillisecsSet   bool
	ShmSleepMillisecs      uint
	ShmPagesSharedSet      bool
	ShmPagesShared         uint64
	ShmPagesSharingSet     bool
	ShmPagesSharing        uint64
	ShmPagesUnsharedSet    bool
	ShmPagesUnshared       uint64
	ShmPagesVolatileSet    bool
	ShmPagesVolatile       uint64
	ShmFullScansSet        bool
	ShmFullScans           uint64
	ShmMergeAcrossNodesSet bool
	ShmMergeAcrossNodes    uint
}

func getMemoryParameterFieldInfo(params *NodeMemoryParameters) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		C.VIR_NODE_MEMORY_SHARED_PAGES_TO_SCAN: typedParamsFieldInfo{
			set: &params.ShmPagesToScanSet,
			ui:  &params.ShmPagesToScan,
		},
		C.VIR_NODE_MEMORY_SHARED_SLEEP_MILLISECS: typedParamsFieldInfo{
			set: &params.ShmSleepMillisecsSet,
			ui:  &params.ShmSleepMillisecs,
		},
		C.VIR_NODE_MEMORY_SHARED_MERGE_ACROSS_NODES: typedParamsFieldInfo{
			set: &params.ShmMergeAcrossNodesSet,
			ui:  &params.ShmMergeAcrossNodes,
		},
		C.VIR_NODE_MEMORY_SHARED_PAGES_SHARED: typedParamsFieldInfo{
			set: &params.ShmPagesSharedSet,
			ul:  &params.ShmPagesShared,
		},
		C.VIR_NODE_MEMORY_SHARED_PAGES_SHARING: typedParamsFieldInfo{
			set: &params.ShmPagesSharingSet,
			ul:  &params.ShmPagesSharing,
		},
		C.VIR_NODE_MEMORY_SHARED_PAGES_UNSHARED: typedParamsFieldInfo{
			set: &params.ShmPagesUnsharedSet,
			ul:  &params.ShmPagesUnshared,
		},
		C.VIR_NODE_MEMORY_SHARED_PAGES_VOLATILE: typedParamsFieldInfo{
			set: &params.ShmPagesVolatileSet,
			ul:  &params.ShmPagesVolatile,
		},
		C.VIR_NODE_MEMORY_SHARED_FULL_SCANS: typedParamsFieldInfo{
			set: &params.ShmFullScansSet,
			ul:  &params.ShmFullScans,
		},
	}
}

func (c *Connect) GetMemoryParameters(flags uint32) (*NodeMemoryParameters, error) {
	params := &NodeMemoryParameters{}
	info := getMemoryParameterFieldInfo(params)

	var nparams C.int

	ret := C.virNodeGetMemoryParameters(c.ptr, nil, &nparams, C.uint(0))
	if ret == -1 {
		return nil, GetLastError()
	}

	cparams := make([]C.virTypedParameter, nparams)
	ret = C.virNodeGetMemoryParameters(c.ptr, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), &nparams, C.uint(flags))
	if ret == -1 {
		return nil, GetLastError()
	}

	defer C.virTypedParamsClear((*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams)

	_, err := typedParamsUnpack(cparams, info)
	if err != nil {
		return nil, err
	}

	return params, nil
}

type NodeMemoryStats struct {
	TotalSet   bool
	Total      uint64
	FreeSet    bool
	Free       uint64
	BuffersSet bool
	Buffers    uint64
	CachedSet  bool
	Cached     uint64
}

func (c *Connect) GetMemoryStats(cellNum int, flags uint32) (*NodeMemoryStats, error) {
	var nparams C.int

	ret := C.virNodeGetMemoryStats(c.ptr, C.int(cellNum), nil, &nparams, 0)
	if ret == -1 {
		return nil, GetLastError()
	}

	params := make([]C.virNodeMemoryStats, nparams)
	ret = C.virNodeGetMemoryStats(c.ptr, C.int(cellNum), (*C.virNodeMemoryStats)(unsafe.Pointer(&params[0])), &nparams, C.uint(flags))
	if ret == -1 {
		return nil, GetLastError()
	}

	stats := &NodeMemoryStats{}
	for i := 0; i < int(nparams); i++ {
		param := params[i]
		field := C.GoString((*C.char)(unsafe.Pointer(&param.field)))
		switch field {
		case C.VIR_NODE_MEMORY_STATS_TOTAL:
			stats.TotalSet = true
			stats.Total = uint64(param.value)
		case C.VIR_NODE_MEMORY_STATS_FREE:
			stats.FreeSet = true
			stats.Free = uint64(param.value)
		case C.VIR_NODE_MEMORY_STATS_BUFFERS:
			stats.BuffersSet = true
			stats.Buffers = uint64(param.value)
		case C.VIR_NODE_MEMORY_STATS_CACHED:
			stats.CachedSet = true
			stats.Cached = uint64(param.value)
		}
	}

	return stats, nil
}

type NodeSecurityModel struct {
	Model string
	Doi   string
}

func (c *Connect) GetSecurityModel() (*NodeSecurityModel, error) {
	var cmodel C.virSecurityModel
	ret := C.virNodeGetSecurityModel(c.ptr, &cmodel)
	if ret == -1 {
		return nil, GetLastError()
	}

	return &NodeSecurityModel{
		Model: C.GoString((*C.char)(unsafe.Pointer(&cmodel.model))),
		Doi:   C.GoString((*C.char)(unsafe.Pointer(&cmodel.doi))),
	}, nil
}

func (c *Connect) SetMemoryParameters(params *NodeMemoryParameters, flags uint32) error {
	info := getMemoryParameterFieldInfo(params)

	var nparams C.int

	ret := C.virNodeGetMemoryParameters(c.ptr, nil, &nparams, 0)
	if ret == -1 {
		return GetLastError()
	}

	cparams := make([]C.virTypedParameter, nparams)
	ret = C.virNodeGetMemoryParameters(c.ptr, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), &nparams, 0)
	if ret == -1 {
		return GetLastError()
	}

	defer C.virTypedParamsClear((*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams)

	err := typedParamsPack(cparams, info)
	if err != nil {
		return err
	}

	ret = C.virNodeSetMemoryParameters(c.ptr, (*C.virTypedParameter)(unsafe.Pointer(&cparams[0])), nparams, C.uint(flags))

	return nil
}

func (c *Connect) SuspendForDuration(target NodeSuspendTarget, duration uint64, flags uint32) error {
	ret := C.virNodeSuspendForDuration(c.ptr, C.uint(target), C.ulonglong(duration), C.uint(flags))
	if ret == -1 {
		return GetLastError()
	}
	return nil
}

func (c *Connect) DomainSaveImageDefineXML(file string, xml string, flags DomainSaveRestoreFlags) error {
	cfile := C.CString(file)
	defer C.free(unsafe.Pointer(cfile))
	cxml := C.CString(xml)
	defer C.free(unsafe.Pointer(cxml))

	ret := C.virDomainSaveImageDefineXML(c.ptr, cfile, cxml, C.uint(flags))

	if ret == -1 {
		return GetLastError()
	}

	return nil
}

func (c *Connect) DomainSaveImageGetXMLDesc(file string, flags DomainXMLFlags) (string, error) {
	cfile := C.CString(file)
	defer C.free(unsafe.Pointer(cfile))

	ret := C.virDomainSaveImageGetXMLDesc(c.ptr, cfile, C.uint(flags))

	if ret == nil {
		return "", GetLastError()
	}

	defer C.free(unsafe.Pointer(ret))

	return C.GoString(ret), nil
}

func (c *Connect) BaselineCPU(xmlCPUs []string, flags ConnectBaselineCPUFlags) (string, error) {
	cxmlCPUs := make([]*C.char, len(xmlCPUs))
	for i := 0; i < len(xmlCPUs); i++ {
		cxmlCPUs[i] = C.CString(xmlCPUs[i])
		defer C.free(unsafe.Pointer(cxmlCPUs[i]))
	}

	ret := C.virConnectBaselineCPU(c.ptr, &cxmlCPUs[0], C.uint(len(xmlCPUs)), C.uint(flags))
	if ret == nil {
		return "", GetLastError()
	}

	defer C.free(unsafe.Pointer(ret))

	return C.GoString(ret), nil
}

func (c *Connect) CompareCPU(xmlDesc string, flags ConnectCompareCPUFlags) (CPUCompareResult, error) {
	cxmlDesc := C.CString(xmlDesc)
	defer C.free(unsafe.Pointer(cxmlDesc))

	ret := C.virConnectCompareCPU(c.ptr, cxmlDesc, C.uint(flags))
	if ret == C.VIR_CPU_COMPARE_ERROR {
		return CPU_COMPARE_ERROR, GetLastError()
	}

	return CPUCompareResult(ret), nil
}

func (c *Connect) DomainXMLFromNative(nativeFormat string, nativeConfig string, flags uint32) (string, error) {
	cnativeFormat := C.CString(nativeFormat)
	defer C.free(unsafe.Pointer(cnativeFormat))
	cnativeConfig := C.CString(nativeConfig)
	defer C.free(unsafe.Pointer(cnativeConfig))

	ret := C.virConnectDomainXMLFromNative(c.ptr, cnativeFormat, cnativeConfig, C.uint(flags))
	if ret == nil {
		return "", GetLastError()
	}

	defer C.free(unsafe.Pointer(ret))

	return C.GoString(ret), nil
}

func (c *Connect) DomainXMLToNative(nativeFormat string, domainXml string, flags uint32) (string, error) {
	cnativeFormat := C.CString(nativeFormat)
	defer C.free(unsafe.Pointer(cnativeFormat))
	cdomainXml := C.CString(domainXml)
	defer C.free(unsafe.Pointer(cdomainXml))

	ret := C.virConnectDomainXMLToNative(c.ptr, cnativeFormat, cdomainXml, C.uint(flags))
	if ret == nil {
		return "", GetLastError()
	}

	defer C.free(unsafe.Pointer(ret))

	return C.GoString(ret), nil
}

func (c *Connect) GetCPUModelNames(arch string, flags uint32) ([]string, error) {
	carch := C.CString(arch)
	defer C.free(unsafe.Pointer(carch))

	var cmodels **C.char
	ret := C.virConnectGetCPUModelNames(c.ptr, carch, &cmodels, C.uint(flags))
	if ret == -1 {
		return []string{}, GetLastError()
	}

	models := make([]string, int(ret))
	for i := 0; i < int(ret); i++ {
		cmodel := *(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cmodels)) + (unsafe.Sizeof(*cmodels) * uintptr(i))))

		defer C.free(unsafe.Pointer(cmodel))
		models[i] = C.GoString(cmodel)
	}
	defer C.free(unsafe.Pointer(cmodels))

	return models, nil
}

func (c *Connect) GetDomainCapabilities(emulatorbin string, arch string, machine string, virttype string, flags uint32) (string, error) {
	if C.LIBVIR_VERSION_NUMBER < 1002007 {
		return "", GetNotImplementedError("virConnectGetDomainCapabilities")
	}
	var cemulatorbin *C.char
	if emulatorbin != "" {
		cemulatorbin = C.CString(emulatorbin)
		defer C.free(unsafe.Pointer(cemulatorbin))
	}
	var carch *C.char
	if arch != "" {
		carch = C.CString(arch)
		defer C.free(unsafe.Pointer(carch))
	}
	var cmachine *C.char
	if machine != "" {
		cmachine = C.CString(machine)
		defer C.free(unsafe.Pointer(cmachine))
	}
	var cvirttype *C.char
	if virttype != "" {
		cvirttype = C.CString(virttype)
		defer C.free(unsafe.Pointer(cvirttype))
	}

	ret := C.virConnectGetDomainCapabilitiesCompat(c.ptr, cemulatorbin, carch, cmachine, cvirttype, C.uint(flags))
	if ret == nil {
		return "", GetLastError()
	}

	defer C.free(unsafe.Pointer(ret))

	return C.GoString(ret), nil
}

func (c *Connect) GetVersion() (uint32, error) {
	var hvVer C.ulong
	ret := C.virConnectGetVersion(c.ptr, &hvVer)
	if ret == -1 {
		return 0, GetLastError()
	}

	return uint32(hvVer), nil
}

func (c *Connect) FindStoragePoolSources(pooltype string, srcSpec string, flags uint32) (string, error) {
	cpooltype := C.CString(pooltype)
	defer C.free(unsafe.Pointer(cpooltype))
	var csrcSpec *C.char
	if srcSpec != "" {
		csrcSpec := C.CString(srcSpec)
		defer C.free(unsafe.Pointer(csrcSpec))
	}
	ret := C.virConnectFindStoragePoolSources(c.ptr, cpooltype, csrcSpec, C.uint(flags))
	if ret == nil {
		return "", GetLastError()
	}

	defer C.free(unsafe.Pointer(ret))

	return C.GoString(ret), nil
}

func (c *Connect) DomainRestore(srcFile string) error {
	cPath := C.CString(srcFile)
	defer C.free(unsafe.Pointer(cPath))
	if result := C.virDomainRestore(c.ptr, cPath); result == -1 {
		return GetLastError()
	}
	return nil
}

func (c *Connect) DomainRestoreFlags(srcFile, xmlConf string, flags DomainSaveRestoreFlags) error {
	cPath := C.CString(srcFile)
	defer C.free(unsafe.Pointer(cPath))
	var cXmlConf *C.char
	if xmlConf != "" {
		cXmlConf = C.CString(xmlConf)
		defer C.free(unsafe.Pointer(cXmlConf))
	}
	if result := C.virDomainRestoreFlags(c.ptr, cPath, cXmlConf, C.uint(flags)); result == -1 {
		return GetLastError()
	}
	return nil
}

func (c *Connect) NewStream(flags StreamFlags) (*Stream, error) {
	virStream := C.virStreamNew(c.ptr, C.uint(flags))
	if virStream == nil {
		return nil, GetLastError()
	}

	return &Stream{
		ptr: virStream,
	}, nil
}

type DomainStatsState struct {
	StateSet  bool
	State     DomainState
	ReasonSet bool
	Reason    int
}

func getDomainStatsStateFieldInfo(params *DomainStatsState) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		"state.state": typedParamsFieldInfo{
			set: &params.StateSet,
			i:   (*int)(unsafe.Pointer(&params.State)),
		},
		"state.reason": typedParamsFieldInfo{
			set: &params.ReasonSet,
			i:   &params.Reason,
		},
	}
}

type DomainStatsCPU struct {
	TimeSet   bool
	Time      uint64
	UserSet   bool
	User      uint64
	SystemSet bool
	System    uint64
}

func getDomainStatsCPUFieldInfo(params *DomainStatsCPU) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		"cpu.time": typedParamsFieldInfo{
			set: &params.TimeSet,
			ul:  &params.Time,
		},
		"cpu.user": typedParamsFieldInfo{
			set: &params.UserSet,
			ul:  &params.User,
		},
		"cpu.system": typedParamsFieldInfo{
			set: &params.SystemSet,
			ul:  &params.System,
		},
	}
}

type DomainStatsBalloon struct {
	CurrentSet bool
	Current    uint64
	MaximumSet bool
	Maximum    uint64
}

func getDomainStatsBalloonFieldInfo(params *DomainStatsBalloon) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		"balloon.current": typedParamsFieldInfo{
			set: &params.CurrentSet,
			ul:  &params.Current,
		},
		"balloon.maximum": typedParamsFieldInfo{
			set: &params.MaximumSet,
			ul:  &params.Maximum,
		},
	}
}

type DomainStatsVcpu struct {
	StateSet bool
	State    VcpuState
	TimeSet  bool
	Time     uint64
}

func getDomainStatsVcpuFieldInfo(idx int, params *DomainStatsVcpu) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		fmt.Sprintf("vcpu.%d.state", idx): typedParamsFieldInfo{
			set: &params.StateSet,
			i:   (*int)(unsafe.Pointer(&params.State)),
		},
		fmt.Sprintf("vcpu.%d.time", idx): typedParamsFieldInfo{
			set: &params.TimeSet,
			ul:  &params.Time,
		},
	}
}

type DomainStatsNet struct {
	NameSet    bool
	Name       string
	RxBytesSet bool
	RxBytes    uint64
	RxPktsSet  bool
	RxPkts     uint64
	RxErrsSet  bool
	RxErrs     uint64
	RxDropSet  bool
	RxDrop     uint64
	TxBytesSet bool
	TxBytes    uint64
	TxPktsSet  bool
	TxPkts     uint64
	TxErrsSet  bool
	TxErrs     uint64
	TxDropSet  bool
	TxDrop     uint64
}

func getDomainStatsNetFieldInfo(idx int, params *DomainStatsNet) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		fmt.Sprintf("net.%d.name", idx): typedParamsFieldInfo{
			set: &params.NameSet,
			s:   &params.Name,
		},
		fmt.Sprintf("net.%d.rx.bytes", idx): typedParamsFieldInfo{
			set: &params.RxBytesSet,
			ul:  &params.RxBytes,
		},
		fmt.Sprintf("net.%d.rx.pkts", idx): typedParamsFieldInfo{
			set: &params.RxPktsSet,
			ul:  &params.RxPkts,
		},
		fmt.Sprintf("net.%d.rx.errs", idx): typedParamsFieldInfo{
			set: &params.RxErrsSet,
			ul:  &params.RxErrs,
		},
		fmt.Sprintf("net.%d.rx.drop", idx): typedParamsFieldInfo{
			set: &params.RxDropSet,
			ul:  &params.RxDrop,
		},
		fmt.Sprintf("net.%d.tx.bytes", idx): typedParamsFieldInfo{
			set: &params.TxBytesSet,
			ul:  &params.TxBytes,
		},
		fmt.Sprintf("net.%d.tx.pkts", idx): typedParamsFieldInfo{
			set: &params.TxPktsSet,
			ul:  &params.TxPkts,
		},
		fmt.Sprintf("net.%d.tx.errs", idx): typedParamsFieldInfo{
			set: &params.TxErrsSet,
			ul:  &params.TxErrs,
		},
		fmt.Sprintf("net.%d.tx.drop", idx): typedParamsFieldInfo{
			set: &params.TxDropSet,
			ul:  &params.TxDrop,
		},
	}
}

type DomainStatsBlock struct {
	NameSet         bool
	Name            string
	BackingIndexSet bool
	BackingIndex    uint
	PathSet         bool
	Path            string
	RdReqsSet       bool
	RdReqs          uint64
	RdBytesSet      bool
	RdBytes         uint64
	RdTimesSet      bool
	RdTimes         uint64
	WrReqsSet       bool
	WrReqs          uint64
	WrBytesSet      bool
	WrBytes         uint64
	WrTimesSet      bool
	WrTimes         uint64
	FlReqsSet       bool
	FlReqs          uint64
	FlTimesSet      bool
	FlTimes         uint64
	ErrorsSet       bool
	Errors          uint64
	AllocationSet   bool
	Allocation      uint64
	CapacitySet     bool
	Capacity        uint64
	PhysicalSet     bool
	Physical        uint64
}

func getDomainStatsBlockFieldInfo(idx int, params *DomainStatsBlock) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		fmt.Sprintf("block.%d.name", idx): typedParamsFieldInfo{
			set: &params.NameSet,
			s:   &params.Name,
		},
		fmt.Sprintf("block.%d.backingIndex", idx): typedParamsFieldInfo{
			set: &params.BackingIndexSet,
			ui:  &params.BackingIndex,
		},
		fmt.Sprintf("block.%d.path", idx): typedParamsFieldInfo{
			set: &params.PathSet,
			s:   &params.Path,
		},
		fmt.Sprintf("block.%d.rd.reqs", idx): typedParamsFieldInfo{
			set: &params.RdReqsSet,
			ul:  &params.RdReqs,
		},
		fmt.Sprintf("block.%d.rd.bytes", idx): typedParamsFieldInfo{
			set: &params.RdBytesSet,
			ul:  &params.RdBytes,
		},
		fmt.Sprintf("block.%d.rd.times", idx): typedParamsFieldInfo{
			set: &params.RdTimesSet,
			ul:  &params.RdTimes,
		},
		fmt.Sprintf("block.%d.wr.reqs", idx): typedParamsFieldInfo{
			set: &params.WrReqsSet,
			ul:  &params.WrReqs,
		},
		fmt.Sprintf("block.%d.wr.bytes", idx): typedParamsFieldInfo{
			set: &params.WrBytesSet,
			ul:  &params.WrBytes,
		},
		fmt.Sprintf("block.%d.wr.times", idx): typedParamsFieldInfo{
			set: &params.WrTimesSet,
			ul:  &params.WrTimes,
		},
		fmt.Sprintf("block.%d.fl.reqs", idx): typedParamsFieldInfo{
			set: &params.FlReqsSet,
			ul:  &params.FlReqs,
		},
		fmt.Sprintf("block.%d.fl.times", idx): typedParamsFieldInfo{
			set: &params.FlTimesSet,
			ul:  &params.FlTimes,
		},
		fmt.Sprintf("block.%d.errors", idx): typedParamsFieldInfo{
			set: &params.ErrorsSet,
			ul:  &params.Errors,
		},
		fmt.Sprintf("block.%d.allocation", idx): typedParamsFieldInfo{
			set: &params.AllocationSet,
			ul:  &params.Allocation,
		},
		fmt.Sprintf("block.%d.capacity", idx): typedParamsFieldInfo{
			set: &params.CapacitySet,
			ul:  &params.Capacity,
		},
		fmt.Sprintf("block.%d.physical", idx): typedParamsFieldInfo{
			set: &params.PhysicalSet,
			ul:  &params.Physical,
		},
	}
}

type DomainStatsPerf struct {
	CmtSet                   bool
	Cmt                      uint64
	MbmtSet                  bool
	Mbmt                     uint64
	MbmlSet                  bool
	Mbml                     uint64
	CacheMissesSet           bool
	CacheMisses              uint64
	CacheReferencesSet       bool
	CacheReferences          uint64
	InstructionsSet          bool
	Instructions             uint64
	CpuCyclesSet             bool
	CpuCycles                uint64
	BranchInstructionsSet    bool
	BranchInstructions       uint64
	BranchMissesSet          bool
	BranchMisses             uint64
	BusCyclesSet             bool
	BusCycles                uint64
	StalledCyclesFrontendSet bool
	StalledCyclesFrontend    uint64
	StalledCyclesBackendSet  bool
	StalledCyclesBackend     uint64
	RefCpuCyclesSet          bool
	RefCpuCycles             uint64
	CpuClockSet              bool
	CpuClock                 uint64
	TaskClockSet             bool
	TaskClock                uint64
	PageFaultsSet            bool
	PageFaults               uint64
	ContextSwitchesSet       bool
	ContextSwitches          uint64
	CpuMigrationsSet         bool
	CpuMigrations            uint64
	PageFaultsMinSet         bool
	PageFaultsMin            uint64
	PageFaultsMajSet         bool
	PageFaultsMaj            uint64
	AlignmentFaultsSet       bool
	AlignmentFaults          uint64
	EmulationFaultsSet       bool
	EmulationFaults          uint64
}

func getDomainStatsPerfFieldInfo(params *DomainStatsPerf) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		"perf.cmt": typedParamsFieldInfo{
			set: &params.CmtSet,
			ul:  &params.Cmt,
		},
		"perf.mbmt": typedParamsFieldInfo{
			set: &params.MbmtSet,
			ul:  &params.Mbmt,
		},
		"perf.mbml": typedParamsFieldInfo{
			set: &params.MbmlSet,
			ul:  &params.Mbml,
		},
		"perf.cache_misses": typedParamsFieldInfo{
			set: &params.CacheMissesSet,
			ul:  &params.CacheMisses,
		},
		"perf.cache_references": typedParamsFieldInfo{
			set: &params.CacheReferencesSet,
			ul:  &params.CacheReferences,
		},
		"perf.instructions": typedParamsFieldInfo{
			set: &params.InstructionsSet,
			ul:  &params.Instructions,
		},
		"perf.cpu_cycles": typedParamsFieldInfo{
			set: &params.CpuCyclesSet,
			ul:  &params.CpuCycles,
		},
		"perf.branch_instructions": typedParamsFieldInfo{
			set: &params.BranchInstructionsSet,
			ul:  &params.BranchInstructions,
		},
		"perf.branch_misses": typedParamsFieldInfo{
			set: &params.BranchMissesSet,
			ul:  &params.BranchMisses,
		},
		"perf.bus_cycles": typedParamsFieldInfo{
			set: &params.BusCyclesSet,
			ul:  &params.BusCycles,
		},
		"perf.stalled_cycles_frontend": typedParamsFieldInfo{
			set: &params.StalledCyclesFrontendSet,
			ul:  &params.StalledCyclesFrontend,
		},
		"perf.stalled_cycles_backend": typedParamsFieldInfo{
			set: &params.StalledCyclesBackendSet,
			ul:  &params.StalledCyclesBackend,
		},
		"perf.ref_cpu_cycles": typedParamsFieldInfo{
			set: &params.RefCpuCyclesSet,
			ul:  &params.RefCpuCycles,
		},
		"perf.cpu_clock": typedParamsFieldInfo{
			set: &params.CpuClockSet,
			ul:  &params.CpuClock,
		},
		"perf.task_clock": typedParamsFieldInfo{
			set: &params.TaskClockSet,
			ul:  &params.TaskClock,
		},
		"perf.page_faults": typedParamsFieldInfo{
			set: &params.PageFaultsSet,
			ul:  &params.PageFaults,
		},
		"perf.context_switches": typedParamsFieldInfo{
			set: &params.ContextSwitchesSet,
			ul:  &params.ContextSwitches,
		},
		"perf.cpu_migrations": typedParamsFieldInfo{
			set: &params.CpuMigrationsSet,
			ul:  &params.CpuMigrations,
		},
		"perf.page_faults_min": typedParamsFieldInfo{
			set: &params.PageFaultsMinSet,
			ul:  &params.PageFaultsMin,
		},
		"perf.page_faults_maj": typedParamsFieldInfo{
			set: &params.PageFaultsMajSet,
			ul:  &params.PageFaultsMaj,
		},
		"perf.alignment_faults": typedParamsFieldInfo{
			set: &params.AlignmentFaultsSet,
			ul:  &params.AlignmentFaults,
		},
		"perf.emulation_faults": typedParamsFieldInfo{
			set: &params.EmulationFaultsSet,
			ul:  &params.EmulationFaults,
		},
	}
}

type DomainStats struct {
	Domain  *Domain
	State   *DomainStatsState
	Cpu     *DomainStatsCPU
	Balloon *DomainStatsBalloon
	Vcpu    []DomainStatsVcpu
	Net     []DomainStatsNet
	Block   []DomainStatsBlock
	Perf    *DomainStatsPerf
}

type domainStatsLengths struct {
	VcpuCurrentSet bool
	VcpuCurrent    uint
	VcpuMaximumSet bool
	VcpuMaximum    uint
	NetCountSet    bool
	NetCount       uint
	BlockCountSet  bool
	BlockCount     uint
}

func getDomainStatsLengthsFieldInfo(params *domainStatsLengths) map[string]typedParamsFieldInfo {
	return map[string]typedParamsFieldInfo{
		"vcpu.current": typedParamsFieldInfo{
			set: &params.VcpuCurrentSet,
			ui:  &params.VcpuCurrent,
		},
		"vcpu.maximum": typedParamsFieldInfo{
			set: &params.VcpuMaximumSet,
			ui:  &params.VcpuMaximum,
		},
		"net.count": typedParamsFieldInfo{
			set: &params.NetCountSet,
			ui:  &params.NetCount,
		},
		"block.count": typedParamsFieldInfo{
			set: &params.BlockCountSet,
			ui:  &params.BlockCount,
		},
	}
}

func (c *Connect) GetAllDomainStats(doms []*Domain, statsTypes DomainStatsTypes, flags ConnectGetAllDomainStatsFlags) ([]DomainStats, error) {
	if C.LIBVIR_VERSION_NUMBER < 1002008 {
		return []DomainStats{}, GetNotImplementedError("virConnectGetAllDomainStats")
	}
	var ret C.int
	var cstats *C.virDomainStatsRecordPtr
	if len(doms) > 0 {
		cdoms := make([]C.virDomainPtr, len(doms)+1)
		for i := 0; i < len(doms); i++ {
			cdoms[i] = doms[i].ptr
		}

		ret = C.virDomainListGetStatsCompat(&cdoms[0], C.uint(statsTypes), &cstats, C.uint(flags))
	} else {
		ret = C.virConnectGetAllDomainStatsCompat(c.ptr, C.uint(statsTypes), &cstats, C.uint(flags))
	}
	if ret == -1 {
		return []DomainStats{}, GetLastError()
	}

	defer C.virDomainStatsRecordListFreeCompat(cstats)

	stats := make([]DomainStats, ret)
	for i := 0; i < int(ret); i++ {
		cdomstats := *(*C.virDomainStatsRecordPtr)(unsafe.Pointer(uintptr(unsafe.Pointer(cstats)) + (unsafe.Sizeof(*cstats) * uintptr(i))))

		domstats := DomainStats{
			Domain: &Domain{ptr: cdomstats.dom},
		}

		state := &DomainStatsState{}
		stateInfo := getDomainStatsStateFieldInfo(state)

		count, err := typedParamsUnpackLen(cdomstats.params, int(cdomstats.nparams), stateInfo)
		if err != nil {
			return []DomainStats{}, err
		}
		if count != 0 {
			domstats.State = state
		}

		cpu := &DomainStatsCPU{}
		cpuInfo := getDomainStatsCPUFieldInfo(cpu)

		count, err = typedParamsUnpackLen(cdomstats.params, int(cdomstats.nparams), cpuInfo)
		if err != nil {
			return []DomainStats{}, err
		}
		if count != 0 {
			domstats.Cpu = cpu
		}

		balloon := &DomainStatsBalloon{}
		balloonInfo := getDomainStatsBalloonFieldInfo(balloon)

		count, err = typedParamsUnpackLen(cdomstats.params, int(cdomstats.nparams), balloonInfo)
		if err != nil {
			return []DomainStats{}, err
		}
		if count != 0 {
			domstats.Balloon = balloon
		}

		perf := &DomainStatsPerf{}
		perfInfo := getDomainStatsPerfFieldInfo(perf)

		count, err = typedParamsUnpackLen(cdomstats.params, int(cdomstats.nparams), perfInfo)
		if err != nil {
			return []DomainStats{}, err
		}
		if count != 0 {
			domstats.Perf = perf
		}

		lengths := domainStatsLengths{}
		lengthsInfo := getDomainStatsLengthsFieldInfo(&lengths)

		count, err = typedParamsUnpackLen(cdomstats.params, int(cdomstats.nparams), lengthsInfo)
		if err != nil {
			return []DomainStats{}, err
		}

		if !lengths.VcpuMaximumSet && lengths.VcpuCurrentSet {
			lengths.VcpuMaximum = lengths.VcpuCurrent
		}

		if lengths.VcpuMaximum > 0 {

			domstats.Vcpu = make([]DomainStatsVcpu, lengths.VcpuMaximum)
			for j := 0; j < int(lengths.VcpuMaximum); j++ {
				vcpu := DomainStatsVcpu{}
				vcpuInfo := getDomainStatsVcpuFieldInfo(j, &vcpu)

				count, err = typedParamsUnpackLen(cdomstats.params, int(cdomstats.nparams), vcpuInfo)
				if err != nil {
					return []DomainStats{}, err
				}
				if count == 0 {
					vcpu.StateSet = true
					vcpu.State = VCPU_OFFLINE
				}
				domstats.Vcpu[j] = vcpu
			}
		}

		if lengths.BlockCountSet && lengths.BlockCount > 0 {
			domstats.Block = make([]DomainStatsBlock, lengths.BlockCount)
			for j := 0; j < int(lengths.BlockCount); j++ {
				block := DomainStatsBlock{}
				blockInfo := getDomainStatsBlockFieldInfo(j, &block)

				count, err = typedParamsUnpackLen(cdomstats.params, int(cdomstats.nparams), blockInfo)
				if err != nil {
					return []DomainStats{}, err
				}
				if count != 0 {
					domstats.Block[j] = block
				}
			}
		}

		if lengths.NetCountSet && lengths.NetCount > 0 {
			domstats.Net = make([]DomainStatsNet, lengths.NetCount)
			for j := 0; j < int(lengths.NetCount); j++ {
				net := DomainStatsNet{}
				netInfo := getDomainStatsNetFieldInfo(j, &net)

				count, err = typedParamsUnpackLen(cdomstats.params, int(cdomstats.nparams), netInfo)
				if err != nil {
					return []DomainStats{}, err
				}
				if count != 0 {
					domstats.Net[j] = net
				}
			}
		}

		stats[i] = domstats
	}

	return stats, nil
}
