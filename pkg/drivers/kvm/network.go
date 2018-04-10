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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"text/template"

	"github.com/docker/machine/libmachine/log"
	libvirt "github.com/libvirt/libvirt-go"
	"github.com/pkg/errors"
)

// Replace with hardcoded range with CIDR
// https://play.golang.org/p/m8TNTtygK0
const networkTmpl = `
<network>
  <name>{{.PrivateNetwork}}</name>
  <ip address='192.168.39.1' netmask='255.255.255.0'>
    <dhcp>
      <range start='192.168.39.2' end='192.168.39.254'/>
    </dhcp>
  </ip>
</network>
`

func (d *Driver) createNetwork() error {
	if d.MAC == "" {
		mac, err := randomMAC()
		if err != nil {
			return errors.Wrap(err, "generating mac address")
		}
		d.MAC = mac.String()
	}
	conn, err := getConnection()
	if err != nil {
		return errors.Wrap(err, "getting libvirt connection")
	}
	defer conn.Close()

	tmpl := template.Must(template.New("network").Parse(networkTmpl))
	var networkXML bytes.Buffer
	err = tmpl.Execute(&networkXML, d)
	if err != nil {
		return errors.Wrap(err, "executing network template")
	}

	//Check if network already exists
	if _, err := conn.LookupNetworkByName(d.PrivateNetwork); err == nil {
		return nil
	}

	network, err := conn.NetworkDefineXML(networkXML.String())
	if err != nil {
		return errors.Wrapf(err, "defining network from xml: %s", networkXML.String())
	}
	err = network.SetAutostart(true)
	if err != nil {
		return errors.Wrap(err, "setting network to autostart")
	}

	err = network.Create()
	if err != nil {
		return errors.Wrap(err, "creating network")
	}

	return nil
}

func (d *Driver) lookupIP() (string, error) {
	dom, conn, err := d.getDomain()
	if err != nil {
		return "", errors.Wrap(err, "getting connection and domain")
	}

	defer closeDomain(dom, conn)

	libVersion, err := conn.GetLibVersion()
	if err != nil {
		return "", errors.Wrap(err, "getting libversion")
	}

	switch {
	case libVersion >= 1002014:
		return d.lookupIPFromListInterface(dom)
	case libVersion >= 1002006:
		return d.lookupIPFromStatusFile(conn)
	default:
		return d.lookupIPFromLeasesFile()
	}
}

func (d *Driver) lookupIPFromListInterface(dom *libvirt.Domain) (string, error) {
	ifaces, err := dom.ListAllInterfaceAddresses(libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE)
	if err != nil {
		return "", errors.Wrap(err, "listing leased addresses")
	}

	_, subnet, _ := net.ParseCIDR("192.168.39.0/24")

	for _, iface := range ifaces {
		for _, ip := range iface.Addrs {
			if subnet.Contains(net.ParseIP(ip.Addr)) {
				return ip.Addr, nil
			}
		}
	}
	return "", nil
}

func (d *Driver) lookupIPFromStatusFile(conn *libvirt.Connect) (string, error) {
	network, err := conn.LookupNetworkByName(d.PrivateNetwork)
	if err != nil {
		return "", errors.Wrap(err, "looking up network by name")
	}

	bridge, err := network.GetBridgeName()
	if err != nil {
		log.Warnf("Failed to get network bridge: %s", err)
		return "", err
	}
	statusFile := fmt.Sprintf("/var/lib/libvirt/dnsmasq/%s.status", bridge)
	statuses, err := ioutil.ReadFile(statusFile)
	if err != nil {
		return "", errors.Wrap(err, "reading status file")
	}
	type StatusEntry struct {
		IPAddress string `json:"ip-address"`
		Hostname  string `json:"hostname"`
	}

	var statusEntries []StatusEntry

	// If the status file is empty, parsing will fail, ignore this error.
	_ = json.Unmarshal(statuses, &statusEntries)

	ipAddress := ""
	for _, status := range statusEntries {
		if status.Hostname == d.MachineName {
			ipAddress = status.IPAddress
		}
	}

	return ipAddress, nil
}

func (d *Driver) lookupIPFromLeasesFile() (string, error) {
	leasesFile := fmt.Sprintf("/var/lib/libvirt/dnsmasq/%s.leases", d.PrivateNetwork)
	leases, err := ioutil.ReadFile(leasesFile)
	if err != nil {
		return "", errors.Wrap(err, "reading leases file")
	}
	ipAddress := ""
	for _, lease := range strings.Split(string(leases), "\n") {
		if len(lease) == 0 {
			continue
		}
		// format for lease entry
		// ExpiryTime MAC IP Hostname ExtendedMAC
		entry := strings.Split(lease, " ")
		if len(entry) != 5 {
			return "", fmt.Errorf("Malformed leases entry: %s", entry)
		}
		if entry[1] == d.MAC {
			ipAddress = entry[2]
		}
	}
	return ipAddress, nil
}
