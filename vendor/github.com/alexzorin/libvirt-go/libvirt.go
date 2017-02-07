package libvirt

import (
	"encoding/base64"
	"io/ioutil"
	"reflect"
	"sync"
	"unsafe"
)

/*
#cgo LDFLAGS: -lvirt
#include <libvirt/libvirt.h>
#include <libvirt/virterror.h>
#include <stdlib.h>
#include "go_libvirt.h"
*/
import "C"

type VirConnection struct {
	ptr C.virConnectPtr
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

func saveConnectionData(c *VirConnection, d *virConnectionData) {
	if c.ptr == nil {
		return // Or panic?
	}
	connectionsLock.Lock()
	defer connectionsLock.Unlock()
	connections[c.ptr] = d
}

func getConnectionData(c *VirConnection) *virConnectionData {
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

func releaseConnectionData(c *VirConnection) {
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

func NewVirConnection(uri string) (VirConnection, error) {
	var cUri *C.char
	if uri != "" {
		cUri = C.CString(uri)
		defer C.free(unsafe.Pointer(cUri))
	}
	ptr := C.virConnectOpen(cUri)
	if ptr == nil {
		return VirConnection{}, GetLastError()
	}
	obj := VirConnection{ptr: ptr}
	return obj, nil
}

func NewVirConnectionWithAuth(uri string, username string, password string) (VirConnection, error) {
	var cUri *C.char

	authMechs := C.authMechs()
	defer C.free(unsafe.Pointer(authMechs))
	cUsername := C.CString(username)
	defer C.free(unsafe.Pointer(cUsername))
	cPassword := C.CString(password)
	defer C.free(unsafe.Pointer(cPassword))
	cbData := C.authData(cUsername, C.uint(len(username)), cPassword, C.uint(len(password)))
	defer C.free(unsafe.Pointer(cbData))

	auth := C.virConnectAuth{
		credtype:  authMechs,
		ncredtype: C.uint(2),
		cb:        C.virConnectAuthCallbackPtr(unsafe.Pointer(C.authCb)),
		cbdata:    unsafe.Pointer(cbData),
	}

	if uri != "" {
		cUri = C.CString(uri)
		defer C.free(unsafe.Pointer(cUri))
	}
	ptr := C.virConnectOpenAuth(cUri, (*C.struct__virConnectAuth)(unsafe.Pointer(&auth)), C.uint(0))
	if ptr == nil {
		return VirConnection{}, GetLastError()
	}
	obj := VirConnection{ptr: ptr}
	return obj, nil
}

func NewVirConnectionReadOnly(uri string) (VirConnection, error) {
	var cUri *C.char
	if uri != "" {
		cUri = C.CString(uri)
		defer C.free(unsafe.Pointer(cUri))
	}
	ptr := C.virConnectOpenReadOnly(cUri)
	if ptr == nil {
		return VirConnection{}, GetLastError()
	}
	obj := VirConnection{ptr: ptr}
	return obj, nil
}

func (c *VirConnection) CloseConnection() (int, error) {
	c.UnsetErrorFunc()
	result := int(C.virConnectClose(c.ptr))
	if result == -1 {
		return result, GetLastError()
	}
	if result == 0 {
		// No more reference to this connection, release data.
		releaseConnectionData(c)
	}
	return result, nil
}

type CloseCallback func(conn VirConnection, reason int, opaque func())
type closeContext struct {
	cb CloseCallback
	f  func()
}

// Register a close callback for the given destination. Only one
// callback per connection is allowed. Setting a callback will remove
// the previous one.
func (c *VirConnection) RegisterCloseCallback(cb CloseCallback, opaque func()) error {
	c.UnregisterCloseCallback()
	context := &closeContext{
		cb: cb,
		f:  opaque,
	}
	goCallbackId := registerCallbackId(context)
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

func (c *VirConnection) UnregisterCloseCallback() error {
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
func closeCallback(conn C.virConnectPtr, reason int, goCallbackId int) {
	ctx := getCallbackId(goCallbackId)
	switch cctx := ctx.(type) {
	case *closeContext:
		cctx.cb(VirConnection{ptr: conn}, reason, cctx.f)
	default:
		panic("Inappropriate callback type called")
	}
}

func (c *VirConnection) GetCapabilities() (string, error) {
	str := C.virConnectGetCapabilities(c.ptr)
	if str == nil {
		return "", GetLastError()
	}
	capabilities := C.GoString(str)
	C.free(unsafe.Pointer(str))
	return capabilities, nil
}

func (c *VirConnection) GetNodeInfo() (VirNodeInfo, error) {
	ni := VirNodeInfo{}
	var ptr C.virNodeInfo
	result := C.virNodeGetInfo(c.ptr, (*C.virNodeInfo)(unsafe.Pointer(&ptr)))
	if result == -1 {
		return ni, GetLastError()
	}
	ni.ptr = ptr
	return ni, nil
}

func (c *VirConnection) GetHostname() (string, error) {
	str := C.virConnectGetHostname(c.ptr)
	if str == nil {
		return "", GetLastError()
	}
	hostname := C.GoString(str)
	C.free(unsafe.Pointer(str))
	return hostname, nil
}

func (c *VirConnection) GetLibVersion() (uint32, error) {
	var version C.ulong
	if err := C.virConnectGetLibVersion(c.ptr, &version); err < 0 {
		return 0, GetLastError()
	}
	return uint32(version), nil
}

func (c *VirConnection) GetType() (string, error) {
	str := C.virConnectGetType(c.ptr)
	if str == nil {
		return "", GetLastError()
	}
	hypDriver := C.GoString(str)
	return hypDriver, nil
}

func (c *VirConnection) IsAlive() (bool, error) {
	result := C.virConnectIsAlive(c.ptr)
	if result == -1 {
		return false, GetLastError()
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (c *VirConnection) IsEncrypted() (bool, error) {
	result := C.virConnectIsEncrypted(c.ptr)
	if result == -1 {
		return false, GetLastError()
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (c *VirConnection) IsSecure() (bool, error) {
	result := C.virConnectIsSecure(c.ptr)
	if result == -1 {
		return false, GetLastError()
	}
	if result == 1 {
		return true, nil
	}
	return false, nil
}

func (c *VirConnection) ListDefinedDomains() ([]string, error) {
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

func (c *VirConnection) ListDomains() ([]uint32, error) {
	var cDomainsIds [512](uint32)
	cDomainsPointer := unsafe.Pointer(&cDomainsIds)
	numDomains := C.virConnectListDomains(c.ptr, (*C.int)(cDomainsPointer), 512)
	if numDomains == -1 {
		return nil, GetLastError()
	}

	return cDomainsIds[:numDomains], nil
}

func (c *VirConnection) ListInterfaces() ([]string, error) {
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

func (c *VirConnection) ListNetworks() ([]string, error) {
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

func (c *VirConnection) ListStoragePools() ([]string, error) {
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

func (c *VirConnection) LookupDomainById(id uint32) (VirDomain, error) {
	ptr := C.virDomainLookupByID(c.ptr, C.int(id))
	if ptr == nil {
		return VirDomain{}, GetLastError()
	}
	return VirDomain{ptr: ptr}, nil
}

func (c *VirConnection) LookupDomainByName(id string) (VirDomain, error) {
	cName := C.CString(id)
	defer C.free(unsafe.Pointer(cName))
	ptr := C.virDomainLookupByName(c.ptr, cName)
	if ptr == nil {
		return VirDomain{}, GetLastError()
	}
	return VirDomain{ptr: ptr}, nil
}

func (c *VirConnection) LookupByUUIDString(uuid string) (VirDomain, error) {
	cUuid := C.CString(uuid)
	defer C.free(unsafe.Pointer(cUuid))
	ptr := C.virDomainLookupByUUIDString(c.ptr, cUuid)
	if ptr == nil {
		return VirDomain{}, GetLastError()
	}
	return VirDomain{ptr: ptr}, nil
}

func (c *VirConnection) DomainCreateXMLFromFile(xmlFile string, flags uint32) (VirDomain, error) {
	xmlConfig, err := ioutil.ReadFile(xmlFile)
	if err != nil {
		return VirDomain{}, err
	}
	return c.DomainCreateXML(string(xmlConfig), flags)
}

func (c *VirConnection) DomainCreateXML(xmlConfig string, flags uint32) (VirDomain, error) {
	cXml := C.CString(string(xmlConfig))
	defer C.free(unsafe.Pointer(cXml))
	ptr := C.virDomainCreateXML(c.ptr, cXml, C.uint(flags))
	if ptr == nil {
		return VirDomain{}, GetLastError()
	}
	return VirDomain{ptr: ptr}, nil
}

func (c *VirConnection) DomainDefineXMLFromFile(xmlFile string) (VirDomain, error) {
	xmlConfig, err := ioutil.ReadFile(xmlFile)
	if err != nil {
		return VirDomain{}, err
	}
	return c.DomainDefineXML(string(xmlConfig))
}

func (c *VirConnection) DomainDefineXML(xmlConfig string) (VirDomain, error) {
	cXml := C.CString(string(xmlConfig))
	defer C.free(unsafe.Pointer(cXml))
	ptr := C.virDomainDefineXML(c.ptr, cXml)
	if ptr == nil {
		return VirDomain{}, GetLastError()
	}
	return VirDomain{ptr: ptr}, nil
}

func (c *VirConnection) ListDefinedInterfaces() ([]string, error) {
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

func (c *VirConnection) ListDefinedNetworks() ([]string, error) {
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

func (c *VirConnection) ListDefinedStoragePools() ([]string, error) {
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

func (c *VirConnection) NumOfDefinedInterfaces() (int, error) {
	result := int(C.virConnectNumOfDefinedInterfaces(c.ptr))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (c *VirConnection) NumOfDefinedNetworks() (int, error) {
	result := int(C.virConnectNumOfDefinedNetworks(c.ptr))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (c *VirConnection) NumOfDefinedStoragePools() (int, error) {
	result := int(C.virConnectNumOfDefinedStoragePools(c.ptr))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (c *VirConnection) NumOfDomains() (int, error) {
	result := int(C.virConnectNumOfDomains(c.ptr))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (c *VirConnection) NumOfInterfaces() (int, error) {
	result := int(C.virConnectNumOfInterfaces(c.ptr))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (c *VirConnection) NumOfNetworks() (int, error) {
	result := int(C.virConnectNumOfNetworks(c.ptr))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (c *VirConnection) NumOfNWFilters() (int, error) {
	result := int(C.virConnectNumOfNWFilters(c.ptr))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (c *VirConnection) NumOfSecrets() (int, error) {
	result := int(C.virConnectNumOfSecrets(c.ptr))
	if result == -1 {
		return 0, GetLastError()
	}
	return result, nil
}

func (c *VirConnection) NetworkDefineXMLFromFile(xmlFile string) (VirNetwork, error) {
	xmlConfig, err := ioutil.ReadFile(xmlFile)
	if err != nil {
		return VirNetwork{}, err
	}
	return c.NetworkDefineXML(string(xmlConfig))
}

func (c *VirConnection) NetworkDefineXML(xmlConfig string) (VirNetwork, error) {
	cXml := C.CString(string(xmlConfig))
	defer C.free(unsafe.Pointer(cXml))
	ptr := C.virNetworkDefineXML(c.ptr, cXml)
	if ptr == nil {
		return VirNetwork{}, GetLastError()
	}
	return VirNetwork{ptr: ptr}, nil
}

func (c *VirConnection) NetworkCreateXML(xmlConfig string) (VirNetwork, error) {
	cXml := C.CString(string(xmlConfig))
	defer C.free(unsafe.Pointer(cXml))
	ptr := C.virNetworkCreateXML(c.ptr, cXml)
	if ptr == nil {
		return VirNetwork{}, GetLastError()
	}
	return VirNetwork{ptr: ptr}, nil
}

func (c *VirConnection) LookupNetworkByName(name string) (VirNetwork, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	ptr := C.virNetworkLookupByName(c.ptr, cName)
	if ptr == nil {
		return VirNetwork{}, GetLastError()
	}
	return VirNetwork{ptr: ptr}, nil
}

func (c *VirConnection) LookupNetworkByUUIDString(uuid string) (VirNetwork, error) {
	cUuid := C.CString(uuid)
	defer C.free(unsafe.Pointer(cUuid))
	ptr := C.virNetworkLookupByUUIDString(c.ptr, cUuid)
	if ptr == nil {
		return VirNetwork{}, GetLastError()
	}
	return VirNetwork{ptr: ptr}, nil
}

func (c *VirConnection) SetKeepAlive(interval int, count uint) error {
	res := int(C.virConnectSetKeepAlive(c.ptr, C.int(interval), C.uint(count)))
	switch res {
	case 0:
		return nil
	default:
		return GetLastError()
	}
}

func (c *VirConnection) GetSysinfo(flags uint) (string, error) {
	cStr := C.virConnectGetSysinfo(c.ptr, C.uint(flags))
	if cStr == nil {
		return "", GetLastError()
	}
	info := C.GoString(cStr)
	C.free(unsafe.Pointer(cStr))
	return info, nil
}

func (c *VirConnection) GetURI() (string, error) {
	cStr := C.virConnectGetURI(c.ptr)
	if cStr == nil {
		return "", GetLastError()
	}
	uri := C.GoString(cStr)
	C.free(unsafe.Pointer(cStr))
	return uri, nil
}

func (c *VirConnection) GetMaxVcpus(typeAttr string) (int, error) {
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

func (c *VirConnection) InterfaceDefineXMLFromFile(xmlFile string) (VirInterface, error) {
	xmlConfig, err := ioutil.ReadFile(xmlFile)
	if err != nil {
		return VirInterface{}, err
	}
	return c.InterfaceDefineXML(string(xmlConfig), 0)
}

func (c *VirConnection) InterfaceDefineXML(xmlConfig string, flags uint32) (VirInterface, error) {
	cXml := C.CString(string(xmlConfig))
	defer C.free(unsafe.Pointer(cXml))
	ptr := C.virInterfaceDefineXML(c.ptr, cXml, C.uint(flags))
	if ptr == nil {
		return VirInterface{}, GetLastError()
	}
	return VirInterface{ptr: ptr}, nil
}

func (c *VirConnection) LookupInterfaceByName(name string) (VirInterface, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	ptr := C.virInterfaceLookupByName(c.ptr, cName)
	if ptr == nil {
		return VirInterface{}, GetLastError()
	}
	return VirInterface{ptr: ptr}, nil
}

func (c *VirConnection) LookupInterfaceByMACString(mac string) (VirInterface, error) {
	cName := C.CString(mac)
	defer C.free(unsafe.Pointer(cName))
	ptr := C.virInterfaceLookupByMACString(c.ptr, cName)
	if ptr == nil {
		return VirInterface{}, GetLastError()
	}
	return VirInterface{ptr: ptr}, nil
}

func (c *VirConnection) StoragePoolDefineXMLFromFile(xmlFile string) (VirStoragePool, error) {
	xmlConfig, err := ioutil.ReadFile(xmlFile)
	if err != nil {
		return VirStoragePool{}, err
	}
	return c.StoragePoolDefineXML(string(xmlConfig), 0)
}

func (c *VirConnection) StoragePoolDefineXML(xmlConfig string, flags uint32) (VirStoragePool, error) {
	cXml := C.CString(string(xmlConfig))
	defer C.free(unsafe.Pointer(cXml))
	ptr := C.virStoragePoolDefineXML(c.ptr, cXml, C.uint(flags))
	if ptr == nil {
		return VirStoragePool{}, GetLastError()
	}
	return VirStoragePool{ptr: ptr}, nil
}

func (c *VirConnection) LookupStoragePoolByName(name string) (VirStoragePool, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	ptr := C.virStoragePoolLookupByName(c.ptr, cName)
	if ptr == nil {
		return VirStoragePool{}, GetLastError()
	}
	return VirStoragePool{ptr: ptr}, nil
}

func (c *VirConnection) LookupStoragePoolByUUIDString(uuid string) (VirStoragePool, error) {
	cUuid := C.CString(uuid)
	defer C.free(unsafe.Pointer(cUuid))
	ptr := C.virStoragePoolLookupByUUIDString(c.ptr, cUuid)
	if ptr == nil {
		return VirStoragePool{}, GetLastError()
	}
	return VirStoragePool{ptr: ptr}, nil
}

func (c *VirConnection) NWFilterDefineXMLFromFile(xmlFile string) (VirNWFilter, error) {
	xmlConfig, err := ioutil.ReadFile(xmlFile)
	if err != nil {
		return VirNWFilter{}, err
	}
	return c.NWFilterDefineXML(string(xmlConfig))
}

func (c *VirConnection) NWFilterDefineXML(xmlConfig string) (VirNWFilter, error) {
	cXml := C.CString(string(xmlConfig))
	defer C.free(unsafe.Pointer(cXml))
	ptr := C.virNWFilterDefineXML(c.ptr, cXml)
	if ptr == nil {
		return VirNWFilter{}, GetLastError()
	}
	return VirNWFilter{ptr: ptr}, nil
}

func (c *VirConnection) LookupNWFilterByName(name string) (VirNWFilter, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	ptr := C.virNWFilterLookupByName(c.ptr, cName)
	if ptr == nil {
		return VirNWFilter{}, GetLastError()
	}
	return VirNWFilter{ptr: ptr}, nil
}

func (c *VirConnection) LookupNWFilterByUUIDString(uuid string) (VirNWFilter, error) {
	cUuid := C.CString(uuid)
	defer C.free(unsafe.Pointer(cUuid))
	ptr := C.virNWFilterLookupByUUIDString(c.ptr, cUuid)
	if ptr == nil {
		return VirNWFilter{}, GetLastError()
	}
	return VirNWFilter{ptr: ptr}, nil
}

func (c *VirConnection) LookupStorageVolByKey(key string) (VirStorageVol, error) {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	ptr := C.virStorageVolLookupByKey(c.ptr, cKey)
	if ptr == nil {
		return VirStorageVol{}, GetLastError()
	}
	return VirStorageVol{ptr: ptr}, nil
}

func (c *VirConnection) LookupStorageVolByPath(path string) (VirStorageVol, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	ptr := C.virStorageVolLookupByPath(c.ptr, cPath)
	if ptr == nil {
		return VirStorageVol{}, GetLastError()
	}
	return VirStorageVol{ptr: ptr}, nil
}

func (c *VirConnection) SecretDefineXMLFromFile(xmlFile string) (VirSecret, error) {
	xmlConfig, err := ioutil.ReadFile(xmlFile)
	if err != nil {
		return VirSecret{}, err
	}
	return c.SecretDefineXML(string(xmlConfig), 0)
}

func (c *VirConnection) SecretDefineXML(xmlConfig string, flags uint32) (VirSecret, error) {
	cXml := C.CString(string(xmlConfig))
	defer C.free(unsafe.Pointer(cXml))
	ptr := C.virSecretDefineXML(c.ptr, cXml, C.uint(flags))
	if ptr == nil {
		return VirSecret{}, GetLastError()
	}
	return VirSecret{ptr: ptr}, nil
}

func (c *VirConnection) SecretSetValue(uuid, value string) error {
	cUuid := C.CString(uuid)
	defer C.free(unsafe.Pointer(cUuid))
	ptr := C.virSecretLookupByUUIDString(c.ptr, cUuid)
	if ptr == nil {
		return GetLastError()
	}

	secret, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return err
	}
	cSecret := C.CString(string(secret))
	defer C.free(unsafe.Pointer(cSecret))

	res := C.virSecretSetValue(ptr, (*C.uchar)(unsafe.Pointer(cSecret)), C.size_t(len(secret)), 0)
	if res != 0 {
		return GetLastError()
	}

	return nil
}

func (c *VirConnection) LookupSecretByUUIDString(uuid string) (VirSecret, error) {
	cUuid := C.CString(uuid)
	defer C.free(unsafe.Pointer(cUuid))
	ptr := C.virSecretLookupByUUIDString(c.ptr, cUuid)
	if ptr == nil {
		return VirSecret{}, GetLastError()
	}
	return VirSecret{ptr: ptr}, nil
}

func (c *VirConnection) LookupSecretByUsage(usageType int, usageID string) (VirSecret, error) {
	cUsageID := C.CString(usageID)
	defer C.free(unsafe.Pointer(cUsageID))
	ptr := C.virSecretLookupByUsage(c.ptr, C.int(usageType), cUsageID)
	if ptr == nil {
		return VirSecret{}, GetLastError()
	}
	return VirSecret{ptr: ptr}, nil
}

func (c *VirConnection) ListAllInterfaces(flags uint32) ([]VirInterface, error) {
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
	var ifaces []VirInterface
	slice := *(*[]C.virInterfacePtr)(unsafe.Pointer(&hdr))
	for _, ptr := range slice {
		ifaces = append(ifaces, VirInterface{ptr})
	}
	C.free(unsafe.Pointer(cList))
	return ifaces, nil
}

func (c *VirConnection) ListAllNetworks(flags uint32) ([]VirNetwork, error) {
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
	var nets []VirNetwork
	slice := *(*[]C.virNetworkPtr)(unsafe.Pointer(&hdr))
	for _, ptr := range slice {
		nets = append(nets, VirNetwork{ptr})
	}
	C.free(unsafe.Pointer(cList))
	return nets, nil
}

func (c *VirConnection) ListAllDomains(flags uint32) ([]VirDomain, error) {
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
	var domains []VirDomain
	slice := *(*[]C.virDomainPtr)(unsafe.Pointer(&hdr))
	for _, ptr := range slice {
		domains = append(domains, VirDomain{ptr})
	}
	C.free(unsafe.Pointer(cList))
	return domains, nil
}

func (c *VirConnection) ListAllNWFilters(flags uint32) ([]VirNWFilter, error) {
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
	var filters []VirNWFilter
	slice := *(*[]C.virNWFilterPtr)(unsafe.Pointer(&hdr))
	for _, ptr := range slice {
		filters = append(filters, VirNWFilter{ptr})
	}
	C.free(unsafe.Pointer(cList))
	return filters, nil
}

func (c *VirConnection) ListAllStoragePools(flags uint32) ([]VirStoragePool, error) {
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
	var pools []VirStoragePool
	slice := *(*[]C.virStoragePoolPtr)(unsafe.Pointer(&hdr))
	for _, ptr := range slice {
		pools = append(pools, VirStoragePool{ptr})
	}
	C.free(unsafe.Pointer(cList))
	return pools, nil
}
