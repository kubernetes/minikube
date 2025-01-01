//go:build linux

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

	"github.com/docker/machine/libmachine/log"
	"github.com/pkg/errors"
	"libvirt.org/go/libvirt"
)

func (d *Driver) getDomain() (*libvirt.Domain, *libvirt.Connect, error) {
	conn, err := getConnection(d.ConnectionURI)
	if err != nil {
		return nil, nil, errors.Wrap(err, "getting libvirt connection")
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
		return nil, errors.Wrap(err, "connecting to libvirt socket")
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
		return nil, errors.Wrap(err, "getting libvirt connection")
	}
	defer func() {
		if _, err := conn.Close(); err != nil {
			log.Errorf("unable to close libvirt connection: %v", err)
		}
	}()

	log.Infof("define libvirt domain using xml: %v", domainXML.String())
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
