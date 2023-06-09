//go:build linux

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
	"text/template"
)

// extraDisksTmpl ExtraDisks XML Template
const extraDisksTmpl = `
<disk type='file' device='disk'>
  <driver name='qemu' type='raw' cache='default' io='threads' />
  <source file='{{.DiskPath}}'/>
  <target dev='{{.DiskLogicalName}}' bus='virtio'/>
</disk>
`

// ExtraDisks holds the extra disks configuration
type ExtraDisks struct {
	DiskPath        string
	DiskLogicalName string
}

// getExtraDiskXML returns the XML that can be added to the libvirt domain XML
// for additional disks
func getExtraDiskXML(diskpath string, logicalName string) (string, error) {
	var extraDisk ExtraDisks
	extraDisk.DiskLogicalName = logicalName
	extraDisk.DiskPath = diskpath
	tmpl := template.Must(template.New("").Parse(extraDisksTmpl))
	var extraDisksXML bytes.Buffer
	if err := tmpl.Execute(&extraDisksXML, extraDisk); err != nil {
		return "", fmt.Errorf("couldn't generate extra disks XML: %v", err)
	}
	return extraDisksXML.String(), nil
}
