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
  <dns enable='no'/>
  <ip address='192.168.39.1' netmask='255.255.255.0'>
    <dhcp>
      <range start='192.168.39.2' end='192.168.39.254'/>
    </dhcp>
  </ip>
</network>
`

// setupNetwork ensures that the network with `name` is started (active)
// and has the autostart feature set.
func setupNetwork(conn *libvirt.Connect, name string) error {
	n, err := conn.LookupNetworkByName(name)
	if err != nil {
		return errors.Wrapf(err, "checking network %s", name)
	}

	// always ensure autostart is set on the network
	autostart, err := n.GetAutostart()
	if err != nil {
		return errors.Wrapf(err, "checking network %s autostart", name)
	}
	if !autostart {
		if err := n.SetAutostart(true); err != nil {
			return errors.Wrapf(err, "setting autostart for network %s", name)
		}
	}

	// always ensure the network is started (active)
	active, err := n.IsActive()
	if err != nil {
		return errors.Wrapf(err, "checking network status for %s", name)
	}
	if !active {
		if err := n.Create(); err != nil {
			return errors.Wrapf(err, "starting network %s", name)
		}
	}

	return nil
}

// ensureNetwork is called on start of the VM
func (d *Driver) ensureNetwork() error {
	conn, err := getConnection()
	if err != nil {
		return errors.Wrap(err, "getting libvirt connection")
	}
	defer conn.Close()

	// network: default

	// Start the default network
	// It is assumed that the libvirt/kvm installation has already created this network
	log.Infof("Ensuring network %s is active", defaultNetworkName)
	if err := setupNetwork(conn, defaultNetworkName); err != nil {
		return err
	}

	// network: private

	// Start the private network
	log.Infof("Ensuring network %s is active", d.PrivateNetwork)
	if err := setupNetwork(conn, d.PrivateNetwork); err != nil {
		return err
	}

	return nil
}

// createNetwork is called during creation of the VM only (and not on start)
func (d *Driver) createNetwork() error {
	conn, err := getConnection()
	if err != nil {
		return errors.Wrap(err, "getting libvirt connection")
	}
	defer conn.Close()

	// network: default
	// It is assumed that the libvirt/kvm installation has already created this network

	// network: private

	// Only create the private network if it does not already exist
	if _, err := conn.LookupNetworkByName(d.PrivateNetwork); err != nil {
		// create the XML for the private network from our networkTmpl
		tmpl := template.Must(template.New("network").Parse(networkTmpl))
		var networkXML bytes.Buffer
		if err := tmpl.Execute(&networkXML, d); err != nil {
			return errors.Wrap(err, "executing network template")
		}

		// define the network using our template
		network, err := conn.NetworkDefineXML(networkXML.String())
		if err != nil {
			return errors.Wrapf(err, "defining network from xml: %s", networkXML.String())
		}

		// and finally create it
		if err := network.Create(); err != nil {
			return errors.Wrapf(err, "creating network %s", d.PrivateNetwork)
		}
	}

	return nil
}

func (d *Driver) lookupIP() (string, error) {
	conn, err := getConnection()
	if err != nil {
		return "", errors.Wrap(err, "getting connection and domain")
	}

	defer conn.Close()

	libVersion, err := conn.GetLibVersion()
	if err != nil {
		return "", errors.Wrap(err, "getting libversion")
	}

	// Earlier versions of libvirt use a lease file instead of a status file
	if libVersion < 1002006 {
		return d.lookupIPFromLeasesFile()
	}

	return d.lookupIPFromStatusFile(conn)
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
		IPAddress  string `json:"ip-address"`
		MacAddress string `json:"mac-address"`
	}

	var statusEntries []StatusEntry

	// If the status file is empty, parsing will fail, ignore this error.
	_ = json.Unmarshal(statuses, &statusEntries)

	ipAddress := ""
	for _, status := range statusEntries {
		if status.MacAddress == d.MAC {
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
