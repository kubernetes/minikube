// +build libvirt.1.2.14

package libvirt

/*
#cgo LDFLAGS: -lvirt
#include <libvirt/libvirt.h>
#include <libvirt/virterror.h>
#include <stdlib.h>
*/
import "C"

import (
	"reflect"
	"time"
	"unsafe"
)

func (n *VirNetwork) GetDHCPLeases() ([]VirNetworkDHCPLease, error) {
	var cLeases *C.virNetworkDHCPLeasePtr
	numLeases := C.virNetworkGetDHCPLeases(n.ptr, nil, (**C.virNetworkDHCPLeasePtr)(&cLeases), C.uint(0))
	if numLeases == -1 {
		return nil, GetLastError()
	}
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cLeases)),
		Len:  int(numLeases),
		Cap:  int(numLeases),
	}
	var leases []VirNetworkDHCPLease
	slice := *(*[]C.virNetworkDHCPLeasePtr)(unsafe.Pointer(&hdr))
	for _, ptr := range slice {
		leases = append(leases, VirNetworkDHCPLease{ptr})
	}
	C.free(unsafe.Pointer(cLeases))
	return leases, nil
}

type VirNetworkDHCPLease struct {
	ptr C.virNetworkDHCPLeasePtr
}

func (l *VirNetworkDHCPLease) GetIface() string {
	return C.GoString(l.ptr.iface)
}

func (l *VirNetworkDHCPLease) GetExpiryTime() time.Time {
	return time.Unix(int64(l.ptr.expirytime), 0)
}

func (l *VirNetworkDHCPLease) GetMACAddress() string {
	return C.GoString(l.ptr.mac)
}

func (l *VirNetworkDHCPLease) GetIPAddress() string {
	return C.GoString(l.ptr.ipaddr)
}

func (l *VirNetworkDHCPLease) GetIPPrefix() uint {
	return uint(l.ptr.prefix)
}

func (l *VirNetworkDHCPLease) GetHostname() string {
	return C.GoString(l.ptr.hostname)
}

func (l *VirNetworkDHCPLease) GetClientID() string {
	return C.GoString(l.ptr.clientid)
}
