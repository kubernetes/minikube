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
)

const (
	DHCPLeasesFile = "/var/db/dhcpd_leases"
	CONFIG_PLIST   = "/Library/Preferences/SystemConfiguration/com.apple.vmnet"
	NET_ADDR_KEY   = "Shared_Net_Address"
)

type DHCPEntry struct {
	Name      string
	IPAddress string
	HWAddress string
	ID        string
	Lease     string
}

func GetIPAddressByMACAddress(mac string) (string, error) {
	return getIpAddressFromFile(mac, DHCPLeasesFile)
}

func getIpAddressFromFile(mac, path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	dhcpEntries, err := parseDHCPdLeasesFile(file)
	if err != nil {
		return "", err
	}
	for _, dhcpEntry := range dhcpEntries {
		if dhcpEntry.HWAddress == mac {
			return dhcpEntry.IPAddress, nil
		}
	}
	return "", fmt.Errorf("Could not find an IP address for %s", mac)
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
			return dhcpEntries, fmt.Errorf("Unable to parse line: %s", line)
		}
	}
	return dhcpEntries, scanner.Err()
}

// trimMacAddress trimming "0" of the ten's digit
func trimMacAddress(rawUUID string) string {
	re := regexp.MustCompile(`0([A-Fa-f0-9](:|$))`)
	mac := re.ReplaceAllString(rawUUID, "$1")

	return mac
}

func GetNetAddr() (net.IP, error) {
	_, err := os.Stat(CONFIG_PLIST + ".plist")
	if err != nil {
		return nil, fmt.Errorf("Does not exist %s", CONFIG_PLIST+".plist")
	}

	out, err := exec.Command("defaults", "read", CONFIG_PLIST, NET_ADDR_KEY).Output()
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(strings.TrimSpace(string(out)))
	if ip == nil {
		return nil, fmt.Errorf("Could not get the network address for vmnet")
	}
	return ip, nil
}
