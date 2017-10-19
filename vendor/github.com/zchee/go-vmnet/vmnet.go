// +build darwin,cgo

package vmnet

// #cgo CFLAGS: -I${SRCDIR} -std=c11 -fblocks -mmacosx-version-min=10.10
// #cgo LDFLAGS: -framework vmnet
/*
#include <stdio.h>
#include <vmnet/vmnet.h>

#include "uuid.h"

static int
get_mac_address_from_uuid(const char *uuid_str, char *mac) {
  xpc_object_t interface_desc;
  uuid_t uuid;
  __block interface_ref iface;
  __block vmnet_return_t iface_status;
  dispatch_semaphore_t iface_created, iface_stopped;
  dispatch_queue_t if_create_q, if_stop_q;
  uint32_t uuid_status;

  interface_desc = xpc_dictionary_create(NULL, NULL, 0);
  xpc_dictionary_set_uint64(interface_desc, vmnet_operation_mode_key, VMNET_SHARED_MODE);

  uuid_from_string(uuid_str, &uuid, &uuid_status);
  if (uuid_status != uuid_s_ok) {
    fprintf(stderr, "Invalid UUID\n");
    return -1;
  }

  xpc_dictionary_set_uuid(interface_desc, vmnet_interface_id_key, uuid);
  iface = NULL;
  iface_status = 0;

  if_create_q = dispatch_queue_create("org.xhyve.vmnet.create", DISPATCH_QUEUE_SERIAL);

  iface_created = dispatch_semaphore_create(0);

  iface = vmnet_start_interface(interface_desc, if_create_q,
    ^(vmnet_return_t status, xpc_object_t interface_param)
  {
    iface_status = status;
    if (status != VMNET_SUCCESS || !interface_param) {
      dispatch_semaphore_signal(iface_created);
      return;
    }

    const char *s = xpc_dictionary_get_string(interface_param, vmnet_mac_address_key);
    strcpy(mac, s);

    dispatch_semaphore_signal(iface_created);
  });

  dispatch_semaphore_wait(iface_created, DISPATCH_TIME_FOREVER);
  dispatch_release(if_create_q);

  if (iface == NULL || iface_status != VMNET_SUCCESS) {
    fprintf(stderr, "virtio_net: Could not create vmnet interface, "
      "permission denied or no entitlement?\n");
    return -1;
  }

  iface_status = 0;

  if_stop_q = dispatch_queue_create("org.xhyve.vmnet.stop", DISPATCH_QUEUE_SERIAL);

  iface_stopped = dispatch_semaphore_create(0);

  iface_status = vmnet_stop_interface(iface, if_stop_q,
    ^(vmnet_return_t status)
  {
    iface_status = status;
    dispatch_semaphore_signal(iface_stopped);
  });

  dispatch_semaphore_wait(iface_stopped, DISPATCH_TIME_FOREVER);
  dispatch_release(if_stop_q);

  if (iface_status != VMNET_SUCCESS) {
    fprintf(stderr, "virtio_net: Could not stop vmnet interface, "
      "permission denied or no entitlement?\n");
    return -1;
  }

  return 0;
}

*/
import "C"
import "fmt"

func GetMACAddressFromUUID(uuid string) (string, error) {
	var mac C.char

	ret := C.get_mac_address_from_uuid(C.CString(uuid), &mac)
	if ret < 0 {
		return "", fmt.Errorf("vmnet: error from vmnet.framework: %d\n", ret)
	}

	return C.GoString(&mac), nil
}
