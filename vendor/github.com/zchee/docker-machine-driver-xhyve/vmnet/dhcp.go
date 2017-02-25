// Originally Author is @ailispaw
// import by https://github.com/ailispaw/xhyvectl/tree/master/vmnet
package vmnet

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	DHCPD_LEASES_FILE = "/var/db/dhcpd_leases"
)

type DHCPEntry struct {
	Name      string
	IPAddress string
	HWAddress string
	ID        string
	Lease     string
}

func parseDHCPdLeasesFile() ([]DHCPEntry, error) {
	file, err := os.Open(DHCPD_LEASES_FILE)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var (
		dhcpEntry   *DHCPEntry
		dhcpEntries []DHCPEntry
	)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "{" {
			dhcpEntry = new(DHCPEntry)
		}
		if strings.HasPrefix(line, "name=") {
			dhcpEntry.Name = line[5:]
		}
		if strings.HasPrefix(line, "ip_address=") {
			dhcpEntry.IPAddress = line[11:]
		}
		if strings.HasPrefix(line, "hw_address=") {
			dhcpEntry.HWAddress = line[13:]
		}
		if strings.HasPrefix(line, "identifier=") {
			dhcpEntry.ID = line[11:]
		}
		if strings.HasPrefix(line, "lease=") {
			dhcpEntry.Lease = line[6:]
		}
		if line == "}" {
			dhcpEntries = append(dhcpEntries, *dhcpEntry)
		}
	}
	return dhcpEntries, scanner.Err()
}

func GetIPAddressByMACAddress(mac string) (string, error) {
	dhcpEntries, err := parseDHCPdLeasesFile()
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

func GetIPAddressByName(name string) (string, error) {
	dhcpEntries, err := parseDHCPdLeasesFile()
	if err != nil {
		return "", err
	}
	for _, dhcpEntry := range dhcpEntries {
		if dhcpEntry.Name == name {
			return dhcpEntry.IPAddress, nil
		}
	}
	return "", fmt.Errorf("Could not find an IP address for %s", name)
}
