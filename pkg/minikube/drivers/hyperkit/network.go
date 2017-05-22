package hyperkit

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

const (
	DHCPLeasesFile = "/var/db/dhcpd_leases"
)

type DHCPEntry struct {
	Name      string
	IPAddress string
	HWAddress string
	ID        string
	Lease     string
}

func GetIPAddressByMACAddress(mac string) (string, error) {
	file, err := os.Open(DHCPLeasesFile)
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
