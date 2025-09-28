/*
Copyright 2017 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
 * Copyright 2017 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package vmware

const vmx = `
.encoding = "UTF-8"
config.version = "8"
displayName = "{{.MachineName}}"
ethernet0.present = "TRUE"
ethernet0.connectionType = "{{.NetworkType}}"
ethernet0.virtualDev = "vmxnet3"
ethernet0.wakeOnPcktRcv = "FALSE"
ethernet0.addressType = "generated"
ethernet0.linkStatePropagation.enable = "TRUE"
pciBridge0.present = "TRUE"
pciBridge4.present = "TRUE"
pciBridge4.virtualDev = "pcieRootPort"
pciBridge4.functions = "8"
pciBridge5.present = "TRUE"
pciBridge5.virtualDev = "pcieRootPort"
pciBridge5.functions = "8"
pciBridge6.present = "TRUE"
pciBridge6.virtualDev = "pcieRootPort"
pciBridge6.functions = "8"
pciBridge7.present = "TRUE"
pciBridge7.virtualDev = "pcieRootPort"
pciBridge7.functions = "8"
pciBridge0.pciSlotNumber = "17"
pciBridge4.pciSlotNumber = "21"
pciBridge5.pciSlotNumber = "22"
pciBridge6.pciSlotNumber = "23"
pciBridge7.pciSlotNumber = "24"
scsi0.pciSlotNumber = "160"
usb.pciSlotNumber = "32"
ethernet0.pciSlotNumber = "192"
sound.pciSlotNumber = "33"
vmci0.pciSlotNumber = "35"
sata0.pciSlotNumber = "36"
floppy0.present = "FALSE"
guestOS = "other3xlinux-64"
hpet0.present = "TRUE"
sata0.present = "TRUE"
sata0:1.present = "TRUE"
sata0:1.fileName = "{{.ISO}}"
sata0:1.deviceType = "cdrom-image"
{{ if .ConfigDriveURL }}
sata0:2.present = "TRUE"
sata0:2.fileName = "{{.ConfigDriveISO}}"
sata0:2.deviceType = "cdrom-image"
{{ end }}
vmci0.present = "TRUE"
mem.hotadd = "TRUE"
memsize = "{{.Memory}}"
powerType.powerOff = "soft"
powerType.powerOn = "soft"
powerType.reset = "soft"
powerType.suspend = "soft"
scsi0.present = "TRUE"
scsi0.virtualDev = "pvscsi"
scsi0:0.fileName = "{{.MachineName}}.vmdk"
scsi0:0.present = "TRUE"
tools.synctime = "TRUE"
virtualHW.productCompatibility = "hosted"
virtualHW.version = "10"
msg.autoanswer = "TRUE"
uuid.action = "create"
numvcpus = "{{.CPU}}"
hgfs.mapRootShare = "FALSE"
hgfs.linkRootShare = "FALSE"
`
