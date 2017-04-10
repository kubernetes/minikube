package virtualbox

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"runtime"

	"github.com/docker/machine/libmachine/log"
)

const (
	buggyNetmask = "0f000000"
	dhcpPrefix   = "HostInterfaceNetworking-"
)

var (
	reHostOnlyAdapterCreated        = regexp.MustCompile(`Interface '(.+)' was successfully created`)
	errNewHostOnlyAdapterNotVisible = errors.New("The host-only adapter we just created is not visible. This is a well known VirtualBox bug. You might want to uninstall it and reinstall at least version 5.0.12 that is is supposed to fix this issue")
)

// Host-only network.
type hostOnlyNetwork struct {
	Name        string
	GUID        string
	DHCP        bool
	IPv4        net.IPNet
	HwAddr      net.HardwareAddr
	Medium      string
	Status      string
	NetworkName string // referenced in DHCP.NetworkName
}

// HostInterfaces returns host network interface info.  By default delegates to net.Interfaces()
type HostInterfaces interface {
	Interfaces() ([]net.Interface, error)
	Addrs(iface *net.Interface) ([]net.Addr, error)
}

func NewHostInterfaces() HostInterfaces {
	return &defaultHostInterfaces{}
}

type defaultHostInterfaces struct {
}

func (ni *defaultHostInterfaces) Interfaces() ([]net.Interface, error) {
	return net.Interfaces()
}

func (ni *defaultHostInterfaces) Addrs(iface *net.Interface) ([]net.Addr, error) {
	return iface.Addrs()
}

// Save changes the configuration of the host-only network.
func (n *hostOnlyNetwork) Save(vbox VBoxManager) error {
	if err := n.SaveIPv4(vbox); err != nil {
		return err
	}

	if n.DHCP {
		vbox.vbm("hostonlyif", "ipconfig", n.Name, "--dhcp") // not implemented as of VirtualBox 4.3
	}

	return nil
}

// SaveIPv4 changes the ipv4 configuration of the host-only network.
func (n *hostOnlyNetwork) SaveIPv4(vbox VBoxManager) error {
	if n.IPv4.IP != nil && n.IPv4.Mask != nil {
		if runtime.GOOS == "windows" {
			log.Warn("Windows might ask for the permission to configure a network adapter. Sometimes, such confirmation window is minimized in the taskbar.")
		}

		if err := vbox.vbm("hostonlyif", "ipconfig", n.Name, "--ip", n.IPv4.IP.String(), "--netmask", net.IP(n.IPv4.Mask).String()); err != nil {
			return err
		}
	}

	return nil
}

// createHostonlyAdapter creates a new host-only network.
func createHostonlyAdapter(vbox VBoxManager) (*hostOnlyNetwork, error) {
	if runtime.GOOS == "windows" {
		log.Warn("Windows might ask for the permission to create a network adapter. Sometimes, such confirmation window is minimized in the taskbar.")
	}

	out, err := vbox.vbmOut("hostonlyif", "create")
	if err != nil {
		return nil, err
	}

	res := reHostOnlyAdapterCreated.FindStringSubmatch(string(out))
	if res == nil {
		return nil, errors.New("Failed to create host-only adapter")
	}

	return &hostOnlyNetwork{Name: res[1]}, nil
}

// listHostOnlyAdapters gets all host-only adapters in a  map keyed by NetworkName.
func listHostOnlyAdapters(vbox VBoxManager) (map[string]*hostOnlyNetwork, error) {
	out, err := vbox.vbmOut("list", "hostonlyifs")
	if err != nil {
		return nil, err
	}

	byName := map[string]*hostOnlyNetwork{}
	byIP := map[string]*hostOnlyNetwork{}
	n := &hostOnlyNetwork{}

	err = parseKeyValues(out, reColonLine, func(key, val string) error {
		switch key {
		case "Name":
			n.Name = val
		case "GUID":
			n.GUID = val
		case "DHCP":
			n.DHCP = (val != "Disabled")
		case "IPAddress":
			n.IPv4.IP = net.ParseIP(val)
		case "NetworkMask":
			n.IPv4.Mask = parseIPv4Mask(val)
		case "HardwareAddress":
			mac, err := net.ParseMAC(val)
			if err != nil {
				return err
			}
			n.HwAddr = mac
		case "MediumType":
			n.Medium = val
		case "Status":
			n.Status = val
		case "VBoxNetworkName":
			n.NetworkName = val

			if _, present := byName[n.NetworkName]; present {
				return fmt.Errorf("VirtualBox is configured with multiple host-only adapters with the same name %q. Please remove one", n.NetworkName)
			}
			byName[n.NetworkName] = n

			if len(n.IPv4.IP) != 0 {
				if _, present := byIP[n.IPv4.IP.String()]; present {
					return fmt.Errorf("VirtualBox is configured with multiple host-only adapters with the same IP %q. Please remove one", n.IPv4.IP)
				}
				byIP[n.IPv4.IP.String()] = n
			}

			n = &hostOnlyNetwork{}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return byName, nil
}

func getHostOnlyAdapter(nets map[string]*hostOnlyNetwork, hostIP net.IP, netmask net.IPMask) *hostOnlyNetwork {
	for _, n := range nets {
		// Second part of this conditional handles a race where
		// VirtualBox returns us the incorrect netmask value for the
		// newly created adapter.
		if hostIP.Equal(n.IPv4.IP) &&
			(netmask.String() == n.IPv4.Mask.String() || n.IPv4.Mask.String() == buggyNetmask) {
			log.Debugf("Found: %s", n.Name)
			return n
		}
	}

	log.Debug("Not found")
	return nil
}

func getOrCreateHostOnlyNetwork(hostIP net.IP, netmask net.IPMask, nets map[string]*hostOnlyNetwork, vbox VBoxManager) (*hostOnlyNetwork, error) {
	// Search for an existing host-only adapter.
	hostOnlyAdapter := getHostOnlyAdapter(nets, hostIP, netmask)
	if hostOnlyAdapter != nil {
		return hostOnlyAdapter, nil
	}

	// No existing host-only adapter found. Create a new one.
	_, err := createHostonlyAdapter(vbox)
	if err != nil {
		// Sometimes the host-only adapter fails to create. See https://www.virtualbox.org/ticket/14040
		// BUT, it is created in fact! So let's wait until it appears last in the list
		log.Warnf("Creating a new host-only adapter produced an error: %s", err)
		log.Warn("This is a known VirtualBox bug. Let's try to recover anyway...")
	}

	// It can take some time for an adapter to appear. Let's poll.
	hostOnlyAdapter, err = waitForNewHostOnlyNetwork(nets, vbox)
	if err != nil {
		// Sometimes, Vbox says it created it but then it cannot be found...
		return nil, errNewHostOnlyAdapterNotVisible
	}

	log.Warnf("Found a new host-only adapter: %q", hostOnlyAdapter.Name)

	hostOnlyAdapter.IPv4.IP = hostIP
	hostOnlyAdapter.IPv4.Mask = netmask
	if err := hostOnlyAdapter.Save(vbox); err != nil {
		return nil, err
	}

	return hostOnlyAdapter, nil
}

func waitForNewHostOnlyNetwork(oldNets map[string]*hostOnlyNetwork, vbox VBoxManager) (*hostOnlyNetwork, error) {
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)

		newNets, err := listHostOnlyAdapters(vbox)
		if err != nil {
			return nil, err
		}

		for name, latestNet := range newNets {
			if _, present := oldNets[name]; !present {
				return latestNet, nil
			}
		}
	}

	return nil, errors.New("Failed to find a new host-only adapter")
}

// DHCP server info.
type dhcpServer struct {
	NetworkName string
	IPv4        net.IPNet
	LowerIP     net.IP
	UpperIP     net.IP
	Enabled     bool
}

// removeOrphanDHCPServers removed the DHCP servers linked to no host-only adapter
func removeOrphanDHCPServers(vbox VBoxManager) error {
	dhcps, err := listDHCPServers(vbox)
	if err != nil {
		return err
	}

	if len(dhcps) == 0 {
		return nil
	}

	log.Debug("Removing orphan DHCP servers...")

	nets, err := listHostOnlyAdapters(vbox)
	if err != nil {
		return err
	}

	for name := range dhcps {
		if strings.HasPrefix(name, dhcpPrefix) {
			if _, present := nets[name]; !present {
				if err := vbox.vbm("dhcpserver", "remove", "--netname", name); err != nil {
					log.Warnf("Unable to remove orphan dhcp server %q: %s", name, err)
				}
			}
		}
	}

	return nil
}

// addHostOnlyDHCPServer adds a DHCP server to a host-only network.
func addHostOnlyDHCPServer(ifname string, d dhcpServer, vbox VBoxManager) error {
	name := dhcpPrefix + ifname

	dhcps, err := listDHCPServers(vbox)
	if err != nil {
		return err
	}

	// On some platforms (OSX), creating a host-only adapter adds a default dhcpserver,
	// while on others (Windows?) it does not.
	command := "add"
	if dhcp, ok := dhcps[name]; ok {
		command = "modify"
		if (dhcp.IPv4.IP.Equal(d.IPv4.IP)) && (dhcp.IPv4.Mask.String() == d.IPv4.Mask.String()) && (dhcp.LowerIP.Equal(d.LowerIP)) && (dhcp.UpperIP.Equal(d.UpperIP)) && (dhcp.Enabled == d.Enabled) {
			// dhcp is up to date
			return nil
		}
	}

	args := []string{"dhcpserver", command,
		"--netname", name,
		"--ip", d.IPv4.IP.String(),
		"--netmask", net.IP(d.IPv4.Mask).String(),
		"--lowerip", d.LowerIP.String(),
		"--upperip", d.UpperIP.String(),
	}
	if d.Enabled {
		args = append(args, "--enable")
	} else {
		args = append(args, "--disable")
	}

	if runtime.GOOS == "windows" {
		log.Warn("Windows might ask for the permission to configure a dhcp server. Sometimes, such confirmation window is minimized in the taskbar.")
	}

	return vbox.vbm(args...)
}

// listDHCPServers lists all DHCP server settings in a map keyed by DHCP.NetworkName.
func listDHCPServers(vbox VBoxManager) (map[string]*dhcpServer, error) {
	out, err := vbox.vbmOut("list", "dhcpservers")
	if err != nil {
		return nil, err
	}

	m := map[string]*dhcpServer{}
	dhcp := &dhcpServer{}

	err = parseKeyValues(out, reColonLine, func(key, val string) error {
		switch key {
		case "NetworkName":
			dhcp = &dhcpServer{}
			m[val] = dhcp
			dhcp.NetworkName = val
		case "IP":
			dhcp.IPv4.IP = net.ParseIP(val)
		case "upperIPAddress":
			dhcp.UpperIP = net.ParseIP(val)
		case "lowerIPAddress":
			dhcp.LowerIP = net.ParseIP(val)
		case "NetworkMask":
			dhcp.IPv4.Mask = parseIPv4Mask(val)
		case "Enabled":
			dhcp.Enabled = (val == "Yes")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}

// listHostInterfaces returns a map of net.IPNet addresses of host interfaces that are "UP" and not loopback adapters
// and not virtualbox host-only networks (given by excludeNets), keyed by CIDR string.
func listHostInterfaces(hif HostInterfaces, excludeNets map[string]*hostOnlyNetwork) (map[string]*net.IPNet, error) {
	ifaces, err := hif.Interfaces()
	if err != nil {
		return nil, err
	}
	m := map[string]*net.IPNet{}

	for _, iface := range ifaces {
		addrs, err := hif.Addrs(&iface)
		if err != nil {
			return nil, err
		}

		// Check if an address of the interface is in the list of excluded addresses
		ifaceExcluded := false
		for _, a := range addrs {
			switch ipnet := a.(type) {
			case *net.IPNet:
				_, excluded := excludeNets[ipnet.String()]
				if excluded {
					ifaceExcluded = true
					break
				}
			}
		}

		// If excluded, or not up, or a loopback interface, skip the interface
		if ifaceExcluded || iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// This is a host interface, so add all its addresses to the map
		for _, a := range addrs {
			switch ipnet := a.(type) {
			case *net.IPNet:
				m[ipnet.String()] = ipnet
			}
		}
	}
	return m, nil
}

// checkIPNetCollision returns true if any host interfaces conflict with the host-only network mask passed as a parameter.
// This works with IPv4 or IPv6 ip addresses.
func checkIPNetCollision(hostonly *net.IPNet, hostIfaces map[string]*net.IPNet) (bool, error) {
	for _, ifaceNet := range hostIfaces {
		if hostonly.IP.Equal(ifaceNet.IP.Mask(ifaceNet.Mask)) {
			return true, nil
		}
	}
	return false, nil
}

// parseIPv4Mask parses IPv4 netmask written in IP form (e.g. 255.255.255.0).
// This function should really belong to the net package.
func parseIPv4Mask(s string) net.IPMask {
	mask := net.ParseIP(s)
	if mask == nil {
		return nil
	}
	return net.IPv4Mask(mask[12], mask[13], mask[14], mask[15])
}
