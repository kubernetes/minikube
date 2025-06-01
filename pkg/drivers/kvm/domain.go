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
		return nil, nil, fmt.Errorf("failed opening libvirt connection: %w", err)
	}

	dom, err := conn.LookupDomainByName(d.MachineName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed looking up domain: %w", lvErr(err))
	}

	return dom, conn, nil
}

func getConnection(connectionURI string) (*libvirt.Connect, error) {
	conn, err := libvirt.NewConnect(connectionURI)
	if err != nil {
		return nil, fmt.Errorf("failed connecting to libvirt socket: %w", lvErr(err))
	}

	return conn, nil
}

func closeDomain(dom *libvirt.Domain, conn *libvirt.Connect) error {
	if dom == nil {
		return fmt.Errorf("nil domain, cannot close")
	}

	if err := dom.Free(); err != nil {
		return err
	}
	res, err := conn.Close()
	if res != 0 {
		return fmt.Errorf("conn.Close() == %d, expected 0", res)
	}
	return err
}

// defineDomain defines the XML for the domain using our domainTmpl template
func (d *Driver) defineDomain() (*libvirt.Domain, error) {
	tmpl := template.Must(template.New("domain").Parse(domainTmpl))
	var domainXML bytes.Buffer
	dlog := struct {
		Driver
		ConsoleLogPath string
	}{
		Driver:         *d,
		ConsoleLogPath: consoleLogPath(*d),
	}
	if err := tmpl.Execute(&domainXML, dlog); err != nil {
		return nil, errors.Wrap(err, "executing domain xml")
	}
	conn, err := getConnection(d.ConnectionURI)
	if err != nil {
		return nil, fmt.Errorf("failed opening libvirt connection: %w", err)
	}
	defer func() {
		if _, err := conn.Close(); err != nil {
			log.Errorf("failed closing libvirt connection: %v", lvErr(err))
		}
	}()

	log.Infof("defining domain using XML: %v", domainXML.String())
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
