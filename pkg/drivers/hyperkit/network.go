// +build darwin

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

package hyperkit

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/docker/machine/libmachine/log"
)

const (
	// LeasesPath is the path to dhcpd leases
	LeasesPath = "/var/db/dhcpd_leases"
	// VMNetDomain is the domain for vmnet
	VMNetDomain = "/Library/Preferences/SystemConfiguration/com.apple.vmnet"
	// SharedNetAddrKey is the key for the network address
	SharedNetAddrKey = "Shared_Net_Address"
)

var (
	leadingZeroRegexp = regexp.MustCompile(`0([A-Fa-f0-9](:|$))`)
)

// DHCPEntry holds a parsed DNS entry
type DHCPEntry struct {
	Name      string
	IPAddress string
	HWAddress string
	ID        string
	Lease     string
}

// GetIPAddressByMACAddress gets the IP address of a MAC address
func GetIPAddressByMACAddress(mac string) (string, error) {
	return getIPAddressFromFile(mac, LeasesPath)
}

func getIPAddressFromFile(mac, path string) (string, error) {
	log.Debugf("Searching for %s in %s ...", mac, path)
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	dhcpEntries, err := parseDHCPdLeasesFile(file)
	if err != nil {
		return "", err
	}
	log.Debugf("Found %d entries in %s!", len(dhcpEntries), path)
	for _, dhcpEntry := range dhcpEntries {
		log.Debugf("dhcp entry: %+v", dhcpEntry)
		if dhcpEntry.HWAddress == mac {
			log.Debugf("Found match: %s", mac)
			return dhcpEntry.IPAddress, nil
		}
	}
	return "", fmt.Errorf("could not find an IP address for %s", mac)
}

func parseDHCPdLeasesFile(file io.Reader) ([]DHCPEntry, error) {
	var (
		dhcpEntry   *DHCPEntry
		dhcpEntries []DHCPEntry
	)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "{" {
			dhcpEntry = new(DHCPEntry)
			continue
		} else if line == "}" {
			dhcpEntries = append(dhcpEntries, *dhcpEntry)
			continue
		}

		split := strings.SplitN(line, "=", 2)
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid line in dhcp leases file: %s", line)
		}
		key, val := split[0], split[1]
		switch key {
		case "name":
			dhcpEntry.Name = val
		case "ip_address":
			dhcpEntry.IPAddress = val
		case "hw_address":
			// The mac addresses have a '1,' at the start.
			dhcpEntry.HWAddress = val[2:]
		case "identifier":
			dhcpEntry.ID = val
		case "lease":
			dhcpEntry.Lease = val
		default:
			return dhcpEntries, fmt.Errorf("unable to parse line: %s", line)
		}
	}
	return dhcpEntries, scanner.Err()
}

// trimMacAddress trimming "0" of the ten's digit
func trimMacAddress(rawUUID string) string {
	return leadingZeroRegexp.ReplaceAllString(rawUUID, "$1")
}

// GetNetAddr gets the network address for vmnet
func GetNetAddr() (net.IP, error) {
	plistPath := VMNetDomain + ".plist"
	if _, err := os.Stat(plistPath); err != nil {
		return nil, fmt.Errorf("stat: %v", err)
	}
	out, err := exec.Command("defaults", "read", VMNetDomain, SharedNetAddrKey).Output()
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(strings.TrimSpace(string(out)))
	if ip == nil {
		return nil, fmt.Errorf("could not get the network address for vmnet")
	}
	return ip, nil
}
