// +build linux

/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

	libvirt "github.com/libvirt/libvirt-go"
	"github.com/pkg/errors"
)

const domainTmpl = `
<domain type='kvm'>
  <name>{{.MachineName}}</name>
  <memory unit='MiB'>{{.Memory}}</memory>
  <vcpu>{{.CPU}}</vcpu>
  <features>
    <acpi/>
    <apic/>
    <pae/>
    {{if .Hidden}}
    <kvm>
      <hidden state='on'/>
    </kvm>
    {{end}}
  </features>
  <cpu mode='host-passthrough'>
  {{if gt .NUMANodeCount 1}}
  {{.NUMANodeXML}}
  {{end}}
  </cpu>
  <os>
    <type>hvm</type>
    <boot dev='cdrom'/>
    <boot dev='hd'/>
    <bootmenu enable='no'/>
  </os>
  <devices>
    <disk type='file' device='cdrom'>
      <source file='{{.ISO}}'/>
      <target dev='hdc' bus='scsi'/>
      <readonly/>
    </disk>
    <disk type='file' device='disk'>
      <driver name='qemu' type='raw' cache='default' io='threads' />
      <source file='{{.DiskPath}}'/>
      <target dev='hda' bus='virtio'/>
    </disk>
    <interface type='network'>
      <source network='{{.PrivateNetwork}}'/>
      <model type='virtio'/>
    </interface>
    <interface type='network'>
      <source network='{{.Network}}'/>
      <model type='virtio'/>
    </interface>
    <serial type='pty'>
      <target port='0'/>
    </serial>
    <console type='pty'>
      <target type='serial' port='0'/>
    </console>
    <rng model='virtio'>
      <backend model='random'>/dev/random</backend>
    </rng>
    {{if .GPU}}
    {{.DevicesXML}}
    {{end}}
  </devices>
</domain>
`

func (d *Driver) getDomain() (*libvirt.Domain, *libvirt.Connect, error) {
	conn, err := getConnection(d.ConnectionURI)
	if err != nil {
		return nil, nil, errors.Wrap(err, "getting domain")
	}

	dom, err := conn.LookupDomainByName(d.MachineName)
	if err != nil {
		return nil, nil, errors.Wrap(err, "looking up domain")
	}

	return dom, conn, nil
}

func getConnection(connectionURI string) (*libvirt.Connect, error) {
	conn, err := libvirt.NewConnect(connectionURI)
	if err != nil {
		return nil, errors.Wrap(err, "error connecting to libvirt socket.")
	}

	return conn, nil
}

func closeDomain(dom *libvirt.Domain, conn *libvirt.Connect) error {
	if err := dom.Free(); err != nil {
		return err
	}
	res, err := conn.Close()
	if res != 0 {
		return fmt.Errorf("conn.Close() == %d, expected 0", res)
	}
	return err
}

func (d *Driver) createDomain() (*libvirt.Domain, error) {
	// create the XML for the domain using our domainTmpl template
	tmpl := template.Must(template.New("domain").Parse(domainTmpl))
	var domainXML bytes.Buffer
	if err := tmpl.Execute(&domainXML, d); err != nil {
		return nil, errors.Wrap(err, "executing domain xml")
	}
	conn, err := getConnection(d.ConnectionURI)
	if err != nil {
		return nil, errors.Wrap(err, "error getting libvirt connection")
	}
	defer conn.Close()

	// define the domain in libvirt using the generated XML
	dom, err := conn.DomainDefineXML(domainXML.String())
	if err != nil {
		return nil, errors.Wrapf(err, "error defining domain xml: %s", domainXML.String())
	}

	// save MAC address
	dmac, err := macFromXML(conn, d.MachineName, d.Network)
	if err != nil {
		return nil, fmt.Errorf("failed saving MAC address: %w", err)
	}
	d.MAC = dmac
	pmac, err := macFromXML(conn, d.MachineName, d.PrivateNetwork)
	if err != nil {
		return nil, fmt.Errorf("failed saving MAC address: %w", err)
	}
	d.PrivateMAC = pmac

	return dom, nil
}
