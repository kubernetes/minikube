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
	"encoding/xml"
	"fmt"
	"net"
	"text/template"
	"time"

	"github.com/docker/machine/libmachine/log"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/network"
	"k8s.io/minikube/pkg/util/retry"
	"libvirt.org/go/libvirt"
)

// Replace with hardcoded range with CIDR
// https://go.dev/play/p/m8TNTtygK0
const networkTmpl = `
<network>
  <name>{{.Name}}</name>
  <dns enable='no'/>
  {{- with .Parameters}}
  <ip address='{{.Gateway}}' netmask='{{.Netmask}}'>
    <dhcp>
      <range start='{{.ClientMin}}' end='{{.ClientMax}}'/>
    </dhcp>
  </ip>
  {{- end}}
</network>
`

type kvmNetwork struct {
	Name string
	network.Parameters
}

type kvmIface struct {
	Type string `xml:"type,attr"`
	Mac  struct {
		Address string `xml:"address,attr"`
	} `xml:"mac"`
	Source struct {
		Network string `xml:"network,attr"`
		Portid  string `xml:"portid,attr"`
		Bridge  string `xml:"bridge,attr"`
	} `xml:"source"`
	Target struct {
		Dev string `xml:"dev,attr"`
	} `xml:"target"`
	Model struct {
		Type string `xml:"type,attr"`
	} `xml:"model"`
	Alias struct {
		Name string `xml:"name,attr"`
	} `xml:"alias"`
}

// firstSubnetAddr is starting subnet to try for new KVM cluster,
// avoiding possible conflict with other local networks by further incrementing it up to 20 times by 10.
const firstSubnetAddr = "192.168.39.0"

// setupNetwork ensures that the network with `name` is started (active)
// and has the autostart feature set.
func setupNetwork(conn *libvirt.Connect, name string) error {
	n, err := conn.LookupNetworkByName(name)
	if err != nil {
		return fmt.Errorf("failed looking up network %s: %w", name, lvErr(err))
	}
	defer func() {
		if n == nil {
			log.Warnf("nil network, cannot free")
		} else if err := n.Free(); err != nil {
			log.Errorf("failed freeing %s network: %v", name, lvErr(err))
		}
	}()

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
		log.Debugf("network %s is not active, trying to start it...", name)
		if err := n.Create(); err != nil {
			return errors.Wrapf(err, "starting network %s", name)
		}
	}
	return nil
}

// ensureNetwork is called on start of the VM
func (d *Driver) ensureNetwork() error {
	conn, err := getConnection(d.ConnectionURI)
	if err != nil {
		return fmt.Errorf("failed opening libvirt connection: %w", err)
	}
	defer func() {
		if _, err := conn.Close(); err != nil {
			log.Errorf("failed closing libvirt connection: %v", lvErr(err))
		}
	}()

	// network: default

	// It is assumed that the libvirt/kvm installation has already created this network
	log.Infof("Ensuring network %s is active", d.Network)
	if err := setupNetwork(conn, d.Network); err != nil {
		return err
	}

	// network: private

	// Start the private network
	log.Infof("Ensuring network %s is active", d.PrivateNetwork)
	// retry once to recreate the network, but only if is not used by another minikube instance
	if err := setupNetwork(conn, d.PrivateNetwork); err != nil {
		log.Debugf("Network %s is inoperable, will try to recreate it: %v", d.PrivateNetwork, err)
		if err := d.deleteNetwork(); err != nil {
			return errors.Wrapf(err, "deleting inoperable network %s", d.PrivateNetwork)
		}
		log.Debugf("Successfully deleted %s network", d.PrivateNetwork)
		if err := d.createNetwork(); err != nil {
			return errors.Wrapf(err, "recreating inoperable network %s", d.PrivateNetwork)
		}
		log.Debugf("Successfully recreated %s network", d.PrivateNetwork)
		if err := setupNetwork(conn, d.PrivateNetwork); err != nil {
			return err
		}
		log.Debugf("Successfully activated %s network", d.PrivateNetwork)
	}

	return nil
}

// createNetwork is called during creation of the VM only (and not on start)
func (d *Driver) createNetwork() error {
	if d.Network == defaultPrivateNetworkName {
		return fmt.Errorf("KVM network can't be named %s. This is the name of the private network created by minikube", defaultPrivateNetworkName)
	}

	conn, err := getConnection(d.ConnectionURI)
	if err != nil {
		return fmt.Errorf("failed opening libvirt connection: %w", err)
	}
	defer func() {
		if _, err := conn.Close(); err != nil {
			log.Errorf("failed closing libvirt connection: %v", lvErr(err))
		}
	}()

	// network: default
	// It is assumed that the libvirt/kvm installation has already created this network
	netd, err := conn.LookupNetworkByName(d.Network)
	if err != nil {
		return fmt.Errorf("failed looking up network %s, cannot continue: %w", d.Network, lvErr(err))
	}
	log.Debugf("found existing %s network", d.Network)

	if netdXML, err := netd.GetXMLDesc(0); err != nil {
		log.Debugf("failed getting %s network XML: %v", d.Network, lvErr(err))
	} else {
		log.Debug(netdXML)
	}

	if err := netd.Free(); err != nil {
		log.Errorf("failed freeing %s network: %v", d.Network, lvErr(err))
	}

	// network: private
	// Only create the private network if it does not already exist
	if netp, err := conn.LookupNetworkByName(d.PrivateNetwork); err == nil {
		log.Warnf("found existing %s private network, skipping creation", d.PrivateNetwork)

		if netpXML, err := netp.GetXMLDesc(0); err != nil {
			log.Debugf("failed getting %s private network XML: %v", d.PrivateNetwork, lvErr(err))
		} else {
			log.Debug(netpXML)
		}

		if err := netp.Free(); err != nil {
			log.Errorf("failed freeing %s private network: %v", d.PrivateNetwork, lvErr(err))
		}
		return nil
	}

	// retry up to 5 times to create kvm network
	for attempts, subnetAddr := 0, firstSubnetAddr; attempts < 5; attempts++ {
		// Rather than iterate through all of the valid subnets, give up at 20 to avoid a lengthy user delay for something that is unlikely to work.
		// will be like 192.168.39.0/24,..., 192.168.248.0/24 (in increment steps of 11)
		var subnet *network.Parameters
		subnet, err = network.FreeSubnet(subnetAddr, 11, 20)
		if err != nil {
			log.Debugf("failed finding free subnet for private network %s after %d attempts: %v", d.PrivateNetwork, 20, err)
			return fmt.Errorf("un-retryable: %w", err)
		}

		// reserve last client ip address for multi-control-plane loadbalancer vip address in ha cluster
		clientMaxIP := net.ParseIP(subnet.ClientMax)
		clientMaxIP.To4()[3]--
		subnet.ClientMax = clientMaxIP.String()

		// create the XML for the private network from our networkTmpl
		tryNet := kvmNetwork{
			Name:       d.PrivateNetwork,
			Parameters: *subnet,
		}
		tmpl := template.Must(template.New("network").Parse(networkTmpl))
		var networkXML bytes.Buffer
		if err = tmpl.Execute(&networkXML, tryNet); err != nil {
			return fmt.Errorf("executing private network template: %w", err)
		}

		// define the network using our template
		log.Debugf("defining private network:\n%s", networkXML.String())
		libvirtNet, err := conn.NetworkDefineXML(networkXML.String())
		if err != nil {
			return fmt.Errorf("defining private network %s %s from xml %s: %w", d.PrivateNetwork, subnet.CIDR, networkXML.String(), err)
		}

		// and finally create & start it
		log.Debugf("creating private network %s %s...", d.PrivateNetwork, subnet.CIDR)
		if err = libvirtNet.Create(); err == nil {
			log.Debugf("private network %s %s created", d.PrivateNetwork, subnet.CIDR)
			if netpXML, err := libvirtNet.GetXMLDesc(0); err != nil {
				log.Debugf("failed getting %s private network XML: %v", d.PrivateNetwork, lvErr(err))
			} else {
				log.Debug(netpXML)
			}

			return nil
		}
		log.Debugf("failed creating private network %s %s, will retry: %v", d.PrivateNetwork, subnet.CIDR, err)
		subnetAddr = subnet.IP
	}
	return fmt.Errorf("failed creating private network %s: %w", d.PrivateNetwork, err)
}

func (d *Driver) deleteNetwork() error {
	conn, err := getConnection(d.ConnectionURI)
	if err != nil {
		return fmt.Errorf("failed opening libvirt connection: %w", err)
	}
	defer func() {
		if _, err := conn.Close(); err != nil {
			log.Errorf("failed closing libvirt connection: %v", lvErr(err))
		}
	}()

	// network: default
	// It is assumed that the OS manages this network
	if d.PrivateNetwork == defaultNetworkName {
		log.Debugf("Using the default network, skipping deletion")
		return nil
	}

	// network: private
	log.Debugf("Checking if network %s exists...", d.PrivateNetwork)
	libvirtNet, err := conn.LookupNetworkByName(d.PrivateNetwork)
	if err != nil {
		if lvErr(err).Code == libvirt.ERR_NO_NETWORK {
			log.Warnf("Network %s does not exist. Skipping deletion", d.PrivateNetwork)
			return nil
		}
		return errors.Wrapf(err, "failed looking up network %s", d.PrivateNetwork)
	}
	defer func() {
		if libvirtNet == nil {
			log.Warnf("nil network, cannot free")
		} else if err := libvirtNet.Free(); err != nil {
			log.Errorf("failed freeing %s network: %v", d.PrivateNetwork, lvErr(err))
		}
	}()

	log.Debugf("Network %s exists", d.PrivateNetwork)

	err = d.checkDomains(conn)
	if err != nil {
		return err
	}

	// when we reach this point, it means it is safe to delete the network

	log.Debugf("Trying to delete network %s...", d.PrivateNetwork)
	deleteFunc := func() error {
		active, err := libvirtNet.IsActive()
		if err != nil {
			return err
		}
		if active {
			log.Debugf("Destroying active network %s", d.PrivateNetwork)
			if err := libvirtNet.Destroy(); err != nil {
				return err
			}
		}
		log.Debugf("Undefining inactive network %s", d.PrivateNetwork)
		return libvirtNet.Undefine()
	}
	if err := retry.Local(deleteFunc, 10*time.Second); err != nil {
		return errors.Wrap(err, "deleting network")
	}
	log.Debugf("Network %s deleted", d.PrivateNetwork)

	return nil
}

func (d *Driver) checkDomains(conn *libvirt.Connect) error {
	type source struct {
		// XMLName xml.Name `xml:"source"`
		Network string `xml:"network,attr"`
	}
	type iface struct {
		// XMLName xml.Name `xml:"interface"`
		Source source `xml:"source"`
	}
	type result struct {
		// XMLName xml.Name `xml:"domain"`
		Name       string  `xml:"name"`
		Interfaces []iface `xml:"devices>interface"`
	}

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

	return nil
}

// Static IP management
// "Update ... existing network definition, with the changes ... taking effect immediately, without needing to destroy and re-start the network."
// ref: https://libvirt.org/manpages/virsh.html#net-update
// ref: https://libvirt.org/html/libvirt-libvirt-network.html#virNetworkUpdate
// ref: https://wiki.libvirt.org/page/Networking#Applying_modifications_to_the_network

// ref: https://libvirt.org/formatnetwork.html#elementsAddress
// ref: https://libvirt.org/html/libvirt-libvirt-domain.html#virDomainInterfaceAddresses
// ref: https://libvirt.org/manpages/virsh.html#domifaddr

// addStaticIP appends new host's name, MAC and static IP address record to list of network DHCP leases.
// It will return nil if host record already exists.
func addStaticIP(conn *libvirt.Connect, networkName, hostname, mac, ip string) error {
	l, err := dhcpLease(conn, networkName, hostname, mac, ip)
	if err != nil {
		return fmt.Errorf("failed looking up network %s for host DHCP lease {name: %q, mac: %q, ip: %q}: %w", networkName, hostname, mac, ip, err)
	}
	if l != nil {
		log.Debugf("skip adding static IP to network %s - found existing host DHCP lease matching {name: %q, mac: %q, ip: %q}", networkName, hostname, mac, ip)
		return nil
	}

	libvirtNet, err := conn.LookupNetworkByName(networkName)
	if err != nil {
		return fmt.Errorf("failed looking up network %s: %w", networkName, err)
	}
	defer func() {
		if libvirtNet == nil {
			log.Warnf("nil network, cannot free")
		} else if err := libvirtNet.Free(); err != nil {
			log.Errorf("failed freeing %s network: %v", networkName, lvErr(err))
		}
	}()

	return libvirtNet.Update(
		libvirt.NETWORK_UPDATE_COMMAND_ADD_LAST,
		libvirt.NETWORK_SECTION_IP_DHCP_HOST,
		-1,
		fmt.Sprintf("<host mac=%q name=%q ip=%q/>", mac, hostname, ip),
		libvirt.NETWORK_UPDATE_AFFECT_LIVE+libvirt.NETWORK_UPDATE_AFFECT_CONFIG)
}

// delStaticIP deletes static IP address record that matches given combination of host's name, MAC and IP from list of network DHCP leases.
// It will return nil if record doesn't exist.
func delStaticIP(conn *libvirt.Connect, networkName, hostname, mac, ip string) error {
	l, err := dhcpLease(conn, networkName, hostname, mac, ip)
	if err != nil {
		return fmt.Errorf("failed looking up network %s for host DHCP lease {name: %q, mac: %q, ip: %q}: %w", networkName, hostname, mac, ip, err)
	}
	if l == nil {
		log.Debugf("skip deleting static IP from network %s - couldn't find host DHCP lease matching {name: %q, mac: %q, ip: %q}", networkName, hostname, mac, ip)
		return nil
	}

	libvirtNet, err := conn.LookupNetworkByName(networkName)
	if err != nil {
		return fmt.Errorf("failed looking up network %s: %w", networkName, err)
	}
	defer func() {
		if libvirtNet == nil {
			log.Warnf("nil network, cannot free")
		} else if err := libvirtNet.Free(); err != nil {
			log.Errorf("failed freeing %s network: %v", networkName, lvErr(err))
		}
	}()

	return libvirtNet.Update(
		libvirt.NETWORK_UPDATE_COMMAND_DELETE,
		libvirt.NETWORK_SECTION_IP_DHCP_HOST,
		-1,
		fmt.Sprintf("<host mac=%q name=%q ip=%q/>", l.Mac, l.Hostname, l.IPaddr),
		libvirt.NETWORK_UPDATE_AFFECT_LIVE+libvirt.NETWORK_UPDATE_AFFECT_CONFIG)
}

// dhcpLease returns network DHCP lease that matches given combination of host's name, MAC and IP.
func dhcpLease(conn *libvirt.Connect, networkName, hostname, mac, ip string) (lease *libvirt.NetworkDHCPLease, err error) {
	if hostname == "" && mac == "" && ip == "" {
		return nil, nil
	}

	libvirtNet, err := conn.LookupNetworkByName(networkName)
	if err != nil {
		return nil, fmt.Errorf("failed looking up network %s: %w", networkName, err)
	}
	defer func() {
		if libvirtNet == nil {
			log.Warnf("nil network, cannot free")
		} else if err := libvirtNet.Free(); err != nil {
			log.Errorf("failed freeing %s network: %v", networkName, lvErr(err))
		}
	}()

	leases, err := libvirtNet.GetDHCPLeases()
	if err != nil {
		return nil, fmt.Errorf("failed getting host DHCP leases: %w", err)
	}

	for _, l := range leases {
		if (hostname == "" || hostname == l.Hostname) && (mac == "" || mac == l.Mac) && (ip == "" || ip == l.IPaddr) {
			log.Debugf("found host DHCP lease matching {name: %q, mac: %q, ip: %q} in network %s: %+v", hostname, mac, ip, networkName, l)
			return &l, nil
		}
	}

	log.Debugf("unable to find host DHCP lease matching {name: %q, mac: %q, ip: %q} in network %s", hostname, mac, ip, networkName)
	return nil, nil
}

// ipFromAPI returns current primary IP address of domain interface in network.
func ipFromAPI(conn *libvirt.Connect, domain, networkName string) (string, error) {
	mac, err := macFromXML(conn, domain, networkName)
	if err != nil {
		return "", fmt.Errorf("failed getting MAC address: %w", err)
	}

	ifaces, err := ifListFromAPI(conn, domain)
	if err != nil {
		return "", fmt.Errorf("failed getting network %s interfaces using API of domain %s: %w", networkName, domain, err)
	}
	for _, i := range ifaces {
		if i.Hwaddr == mac {
			if i.Addrs != nil {
				log.Debugf("domain %s has current primary IP address %s and MAC address %s in network %s", domain, i.Addrs[0].Addr, mac, networkName)
				return i.Addrs[0].Addr, nil
			}
			log.Debugf("domain %s with MAC address %s doesn't have current IP address in network %s: %+v", domain, mac, networkName, i)
			return "", nil
		}
	}

	log.Debugf("unable to find current IP address of domain %s in network %s (interfaces detected: %+v)", domain, networkName, ifaces)
	return "", nil
}

// ifListFromAPI returns current domain interfaces.
func ifListFromAPI(conn *libvirt.Connect, domain string) ([]libvirt.DomainInterface, error) {
	dom, err := conn.LookupDomainByName(domain)
	if err != nil {
		return nil, fmt.Errorf("failed looking up domain %s: %w", domain, lvErr(err))
	}
	defer func() {
		if dom == nil {
			log.Warnf("nil domain, cannot free")
		} else if err := dom.Free(); err != nil {
			log.Errorf("failed freeing %s domain: %v", domain, lvErr(err))
		}
	}()

	ifs, err := dom.ListAllInterfaceAddresses(libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE)
	if len(ifs) == 0 {
		if err != nil {
			log.Debugf("failed listing network interface addresses of domain %s (source=lease): %v", domain, lvErr(err))
		} else {
			log.Debugf("no network interface addresses found for domain %s (source=lease)", domain)
		}
		log.Debugf("trying to list again with source=arp")

		if ifs, err = dom.ListAllInterfaceAddresses(libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_ARP); err != nil {
			return nil, fmt.Errorf("failed listing network interface addresses of domain %s (source=arp): %w", domain, lvErr(err))
		}
	}

	return ifs, nil
}

// ipFromXML returns defined IP address of interface in network.
func ipFromXML(conn *libvirt.Connect, domain, networkName string) (string, error) {
	mac, err := macFromXML(conn, domain, networkName)
	if err != nil {
		return "", fmt.Errorf("failed getting MAC address: %w", err)
	}

	lease, err := dhcpLease(conn, networkName, "", mac, "")
	if err != nil {
		return "", fmt.Errorf("failed looking up network %s for host DHCP lease {name: <any>, mac: %q, ip: <any>}: %w", networkName, mac, err)
	}
	if lease == nil {
		log.Debugf("unable to find defined IP address of network %s interface with MAC address %s", networkName, mac)
		return "", nil
	}

	log.Debugf("domain %s has defined IP address %s and MAC address %s in network %s", domain, lease.IPaddr, mac, networkName)
	return lease.IPaddr, nil
}

// macFromXML returns defined MAC address of interface in network from domain XML.
func macFromXML(conn *libvirt.Connect, domain, networkName string) (string, error) {
	domIfs, err := ifListFromXML(conn, domain)
	if err != nil {
		return "", fmt.Errorf("failed getting network %s interfaces using XML of domain %s: %w", networkName, domain, err)
	}

	for _, i := range domIfs {
		if i.Source.Network == networkName {
			log.Debugf("domain %s has defined MAC address %s in network %s", domain, i.Mac.Address, networkName)
			return i.Mac.Address, nil
		}
	}

	return "", fmt.Errorf("unable to get defined MAC address of network %s interface using XML of domain %s: network %s not found", networkName, domain, networkName)
}

// ifListFromXML returns defined domain interfaces from domain XML.
func ifListFromXML(conn *libvirt.Connect, domain string) ([]kvmIface, error) {
	dom, err := conn.LookupDomainByName(domain)
	if err != nil {
		return nil, fmt.Errorf("failed looking up domain %s: %w", domain, err)
	}
	defer func() {
		if dom == nil {
			log.Warnf("nil domain, cannot free")
		} else if err := dom.Free(); err != nil {
			log.Errorf("failed freeing %s domain: %v", domain, lvErr(err))
		}
	}()

	domXML, err := dom.GetXMLDesc(0)
	if err != nil {
		return nil, fmt.Errorf("failed getting XML of domain %s: %w", domain, err)
	}

	var d struct {
		Interfaces []kvmIface `xml:"devices>interface"`
	}
	err = xml.Unmarshal([]byte(domXML), &d)
	if err != nil {
		return nil, fmt.Errorf("failed parsing XML of domain %s: %w", domain, err)
	}

	return d.Interfaces, nil
}
