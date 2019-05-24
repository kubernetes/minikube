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
	"encoding/json"
	"encoding/xml"
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

	// It is assumed that the libvirt/kvm installation has already created this network
	log.Infof("Ensuring network %s is active", d.Network)
	if err := setupNetwork(conn, d.Network); err != nil {
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
	if _, err := conn.LookupNetworkByName(d.Network); err != nil {
		return errors.Wrapf(err, "network %s doesn't exist", d.Network)
	}

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

func (d *Driver) deleteNetwork() error {
	type source struct {
		//XMLName xml.Name `xml:"source"`
		Network string `xml:"network,attr"`
	}
	type iface struct {
		//XMLName xml.Name `xml:"interface"`
		Source source `xml:"source"`
	}
	type result struct {
		//XMLName xml.Name `xml:"domain"`
		Name       string  `xml:"name"`
		Interfaces []iface `xml:"devices>interface"`
	}

	conn, err := getConnection()
	if err != nil {
		return errors.Wrap(err, "getting libvirt connection")
	}
	defer conn.Close()

	// network: default
	// It is assumed that the OS manages this network

	// network: private
	log.Debugf("Checking if network %s exists...", d.PrivateNetwork)
	network, err := conn.LookupNetworkByName(d.PrivateNetwork)
	if err != nil {
		if libvirtErr, ok := err.(libvirt.Error); ok && libvirtErr.Code == libvirt.ERR_NO_NETWORK {
			log.Warnf("Network %s does not exist. Skipping deletion", d.PrivateNetwork)
			return nil
		}

		return errors.Wrapf(err, "failed looking for network %s", d.PrivateNetwork)
	}
	log.Debugf("Network %s exists", d.PrivateNetwork)

	// iterate over every (also turned off) domains, and check if it
	// is using the private network. Do *not* delete the network if
	// that is the case
	log.Debug("Trying to list all domains...")
	doms, err := conn.ListAllDomains(0)
	if err != nil {
		return errors.Wrap(err, "list all domains")
	}
	log.Debugf("Listed all domains: total of %d domains", len(doms))

	// fail if there are 0 domains
	if len(doms) == 0 {
		log.Warn("list of domains is 0 length")
	}

	for _, dom := range doms {
		// get the name of the domain we iterate over
		log.Debug("Trying to get name of domain...")
		name, err := dom.GetName()
		if err != nil {
			return errors.Wrap(err, "failed to get name of a domain")
		}
		log.Debugf("Got domain name: %s", name)

		// skip the domain if it is our own machine
		if name == d.MachineName {
			log.Debug("Skipping domain as it is us...")
			continue
		}

		// unfortunately, there is no better way to retrieve a list of all defined interfaces
		// in domains than getting it from the defined XML of all domains
		// NOTE: conn.ListAllInterfaces does not help in this case
		log.Debugf("Getting XML for domain %s...", name)
		xmlString, err := dom.GetXMLDesc(libvirt.DOMAIN_XML_INACTIVE)
		if err != nil {
			return errors.Wrapf(err, "failed to get XML of domain '%s'", name)
		}
		log.Debugf("Got XML for domain %s", name)

		v := result{}
		err = xml.Unmarshal([]byte(xmlString), &v)
		if err != nil {
			return errors.Wrapf(err, "failed to unmarshal XML of domain '%s", name)
		}
		log.Debugf("Unmarshaled XML for domain %s: %#v", name, v)

		// iterate over the found interfaces
		for _, i := range v.Interfaces {
			if i.Source.Network == d.PrivateNetwork {
				log.Debugf("domain %s DOES use network %s, aborting...", name, d.PrivateNetwork)
				return fmt.Errorf("network still in use at least by domain '%s',", name)
			}
			log.Debugf("domain %s does not use network %s", name, d.PrivateNetwork)
		}
	}

	// when we reach this point, it means it is safe to delete the network
	log.Debugf("Trying to destroy network %s...", d.PrivateNetwork)
	err = network.Destroy()
	if err != nil {
		return errors.Wrap(err, "network destroy")
	}
	log.Debugf("Trying to undefine network %s...", d.PrivateNetwork)
	err = network.Undefine()
	if err != nil {
		return errors.Wrap(err, "network undefine")
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

	// TODO: for everything > 1002006, there is direct support in the libvirt-go for handling this
	return d.lookupIPFromStatusFile(conn)
}

func (d *Driver) lookupIPFromStatusFile(conn *libvirt.Connect) (string, error) {
	network, err := conn.LookupNetworkByName(d.PrivateNetwork)
	if err != nil {
		return "", errors.Wrap(err, "looking up network by name")
	}

	bridge, err := network.GetBridgeName()
	if err != nil {
		log.Warnf("Failed to get network bridge: %v", err)
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
		if status.MacAddress == d.PrivateMAC {
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
			return "", fmt.Errorf("malformed leases entry: %s", entry)
		}
		if entry[1] == d.PrivateMAC {
			ipAddress = entry[2]
		}
	}
	return ipAddress, nil
}
