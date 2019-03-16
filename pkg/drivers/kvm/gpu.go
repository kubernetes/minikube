/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package kvm

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/docker/machine/libmachine/log"
)

var sysFsPCIDevicesPath = "/sys/bus/pci/devices/"
var sysKernelIOMMUGroupsPath = "/sys/kernel/iommu_groups/"

const nvidiaVendorID = "0x10de"

const devicesTmpl = `
<graphics type='spice' autoport='yes'>
  <listen type='address'/>
  <image compression='off'/>
</graphics>
<video>
  <model type='cirrus' vram='16384' heads='1' primary='yes'/>
</video>
{{range .}}
<hostdev mode='subsystem' type='pci' managed='yes'>
  <source>
    <address domain='{{.Domain}}' bus='{{.Bus}}' slot='{{.Slot}}' function='{{.Function}}'/>
  </source>
</hostdev>
{{end}}
`

// PCIDevice holds a parsed PCI device
type PCIDevice struct {
	Domain   string
	Bus      string
	Slot     string
	Function string
}

// getDevicesXML returns the XML that can be added to the libvirt domain XML to
// passthrough NVIDIA devices.
func getDevicesXML() (string, error) {
	unboundNVIDIADevices, err := getPassthroughableNVIDIADevices()
	if err != nil {
		return "", fmt.Errorf("couldn't generate devices XML: %v", err)
	}
	var pciDevices []PCIDevice
	for _, device := range unboundNVIDIADevices {
		splits := strings.Split(device, ":")
		if len(splits) != 3 {
			log.Infof("Error while parsing PCI device %q. Not splitable into domain:bus:slot.function.", device)
			continue
		}
		parts := strings.Split(splits[2], ".")
		if len(parts) != 2 {
			log.Infof("Error while parsing PCI device %q. Not splitable into domain:bus:slot.function.", device)
			continue
		}
		pciDevice := PCIDevice{
			Domain:   "0x" + splits[0],
			Bus:      "0x" + splits[1],
			Slot:     "0x" + parts[0],
			Function: "0x" + parts[1],
		}
		pciDevices = append(pciDevices, pciDevice)
	}
	if len(pciDevices) == 0 {
		return "", fmt.Errorf("couldn't generate devices XML: parsing failed")
	}
	tmpl := template.Must(template.New("").Parse(devicesTmpl))
	var devicesXML bytes.Buffer
	if err := tmpl.Execute(&devicesXML, pciDevices); err != nil {
		return "", fmt.Errorf("couldn't generate devices XML: %v", err)
	}
	return devicesXML.String(), nil
}

// getPassthroughableNVIDIADevices returns a list of NVIDIA devices that can be
// passthrough from the host to a VM. It returns an error if:
// - host doesn't support pci passthrough (IOMMU).
// - there are no passthorughable NVIDIA devices on the host.
func getPassthroughableNVIDIADevices() ([]string, error) {

	// Make sure the host supports IOMMU
	iommuGroups, err := ioutil.ReadDir(sysKernelIOMMUGroupsPath)
	if err != nil {
		return []string{}, fmt.Errorf("error reading %q: %v", sysKernelIOMMUGroupsPath, err)
	}
	if len(iommuGroups) == 0 {
		return []string{}, fmt.Errorf("no IOMMU groups found at %q. Make sure your host supports IOMMU. See instructions at https://github.com/kubernetes/minikube/blob/master/docs/gpu.md", sysKernelIOMMUGroupsPath)
	}

	// Get list of PCI devices
	devices, err := ioutil.ReadDir(sysFsPCIDevicesPath)
	if err != nil {
		return []string{}, fmt.Errorf("error reading %q: %v", sysFsPCIDevicesPath, err)
	}

	unboundNVIDIADevices := make(map[string]bool)
	found := false
	for _, device := range devices {
		vendorPath := filepath.Join(sysFsPCIDevicesPath, device.Name(), "vendor")
		content, err := ioutil.ReadFile(vendorPath)
		if err != nil {
			log.Infof("Error while reading %q: %v", vendorPath, err)
			continue
		}

		// Check if this is an NVIDIA device
		if strings.EqualFold(strings.TrimSpace(string(content)), nvidiaVendorID) {
			log.Infof("Found device %v with NVIDIA's vendorId %v", device.Name(), nvidiaVendorID)
			found = true

			// Check whether it's unbound. We don't want the device to be bound to nvidia/nouveau etc.
			if isUnbound(device.Name()) {
				// Add the unbound device to the map. The value is set to false initially,
				// it will be set to true later if the device is also isolated.
				unboundNVIDIADevices[device.Name()] = false
			}
		}
	}
	if !found {
		return []string{}, fmt.Errorf("no NVIDIA devices found")
	}
	if len(unboundNVIDIADevices) == 0 {
		return []string{}, fmt.Errorf("some NVIDIA devices were found but none of them were unbound. See instructions at https://github.com/kubernetes/minikube/blob/master/docs/gpu.md")
	}

	// Make sure all the unbound devices are in IOMMU groups that only contain unbound devices.
	for device := range unboundNVIDIADevices {
		unboundNVIDIADevices[device] = isIsolated(device)
	}

	isolatedNVIDIADevices := make([]string, 0, len(unboundNVIDIADevices))
	for unboundNVIDIADevice, isIsolated := range unboundNVIDIADevices {
		if isIsolated {
			isolatedNVIDIADevices = append(isolatedNVIDIADevices, unboundNVIDIADevice)
		}
	}
	if len(isolatedNVIDIADevices) == 0 {
		return []string{}, fmt.Errorf("some unbound NVIDIA devices were found but they had other devices in their IOMMU group that were bound. See instructoins at https://github.com/kubernetes/minikube/blob/master/docs/gpu.md")
	}

	return isolatedNVIDIADevices, nil
}

// isIsolated returns true if the device is an IOMMU group that only consists of unbound devices.
// The input device is expected to be a string like 0000:03:00.1 (Domain:Bus:Slot.Function)
func isIsolated(device string) bool {
	// Find out the other devices in the same IOMMU group as one of our unbound device.
	iommuGroupPath := filepath.Join(sysFsPCIDevicesPath, device, "iommu_group", "devices")
	otherDevices, err := ioutil.ReadDir(iommuGroupPath)
	if err != nil {
		log.Infof("Error reading %q: %v", iommuGroupPath)
		return false
	}

	for _, otherDevice := range otherDevices {
		// Check if the other device in the IOMMU group is unbound.
		if isUnbound(otherDevice.Name()) {
			continue
		}
		// If any of the other device in the IOMMU group is not unbound,
		// then our device is not isolated and cannot be safely passthrough.
		return false
	}
	return true
}

// isUnbound returns true if the device is not bound to any driver or if it's
// bound to a stub driver like pci-stub or vfio-pci.
// The input device is expected to be a string like 0000:03:00.1 (Domain:Bus:Slot.Function)
func isUnbound(device string) bool {
	modulePath, err := filepath.EvalSymlinks(filepath.Join(sysFsPCIDevicesPath, device, "driver", "module"))
	if os.IsNotExist(err) {
		log.Infof("%v is not bound to any driver", device)
		return true
	}
	module := filepath.Base(modulePath)
	if module == "pci_stub" || module == "vfio_pci" {
		log.Infof("%v is bound to a stub module: %v", device, module)
		return true
	}
	log.Infof("%v is bound to a non-stub module: %v", device, module)
	return false
}
