// +build libvirt.1.2.14

package libvirt

/*
#cgo LDFLAGS: -lvirt-qemu -lvirt
#include <libvirt/libvirt.h>
#include <libvirt/libvirt-qemu.h>
#include <libvirt/virterror.h>
#include <stdlib.h>
*/
import "C"

import (
	"reflect"
	"unsafe"
)

type VirDomainIPAddress struct {
	Type   int
	Addr   string
	Prefix uint
}

type VirDomainInterface struct {
	Name   string
	Hwaddr string
	Addrs  []VirDomainIPAddress
}

func (d *VirDomain) ListAllInterfaceAddresses(src uint) ([]VirDomainInterface, error) {
	var cList *C.virDomainInterfacePtr
	numIfaces := int(C.virDomainInterfaceAddresses(d.ptr, (**C.virDomainInterfacePtr)(&cList), C.uint(src), 0))
	if numIfaces == -1 {
		return nil, GetLastError()
	}

	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cList)),
		Len:  int(numIfaces),
		Cap:  int(numIfaces),
	}

	ifaces := make([]VirDomainInterface, numIfaces)
	ifaceSlice := *(*[]C.virDomainInterfacePtr)(unsafe.Pointer(&hdr))

	for i := 0; i < numIfaces; i++ {
		ifaces[i].Name = C.GoString(ifaceSlice[i].name)
		ifaces[i].Hwaddr = C.GoString(ifaceSlice[i].hwaddr)

		numAddr := int(ifaceSlice[i].naddrs)
		addrHdr := reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(&ifaceSlice[i].addrs)),
			Len:  int(numAddr),
			Cap:  int(numAddr),
		}

		ifaces[i].Addrs = make([]VirDomainIPAddress, numAddr)
		addrSlice := *(*[]C.virDomainIPAddressPtr)(unsafe.Pointer(&addrHdr))

		for k := 0; k < numAddr; k++ {
			ifaces[i].Addrs[k] = VirDomainIPAddress{}
			ifaces[i].Addrs[k].Type = int(addrSlice[k]._type)
			ifaces[i].Addrs[k].Addr = C.GoString(addrSlice[k].addr)
			ifaces[i].Addrs[k].Prefix = uint(addrSlice[k].prefix)

		}
		C.virDomainInterfaceFree(ifaceSlice[i])
	}
	C.free(unsafe.Pointer(cList))
	return ifaces, nil
}
