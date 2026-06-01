/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

package dhcp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"k8s.io/minikube/pkg/libmachine/log"
	"k8s.io/minikube/pkg/minikube/detect"
)

const (
	// leasesPath is the path to dhcpd leases file
	leasesPath = "/var/db/dhcpd_leases"

	// pollInterval is how long to wait between attempts in WaitForLease.
	pollInterval = 2 * time.Second
)

// Entry holds a parsed DHCP lease entry.
type Entry struct {
	Name      string
	IPAddress string
	HWAddress net.HardwareAddr
	ID        string
	Lease     string
}

// IPAddressForMAC returns the IP address leased to the given MAC address.
func IPAddressForMAC(mac string) (string, error) {
	return ipAddressFromFile(mac, leasesPath)
}

// WaitForLease polls the DHCP leases file until a lease for the given MAC
// address appears or the timeout expires. In nested VMs the timeout is tripled
// automatically.
func WaitForLease(mac string, timeout time.Duration) (string, error) {
	if detect.NestedVM() {
		log.Debugf("Nested VM detected, increasing timeout from %s to %s", timeout, timeout*3)
		timeout *= 3
	}
	log.Infof("Waiting for DHCP lease for %s (timeout %s)", mac, timeout)

	start := time.Now()
	deadline := start.Add(timeout)
	for i := 0; ; i++ {
		log.Debugf("Searching for %s in %s (attempt %d) ...", mac, leasesPath, i)
		ip, err := ipAddressFromFile(mac, leasesPath)
		if err == nil {
			log.Infof("Found DHCP lease for %s: %s in %.3f seconds", mac, ip, time.Since(start).Seconds())
			return ip, nil
		}
		if time.Now().After(deadline) {
			return "", err
		}
		time.Sleep(pollInterval)
	}
}

func ipAddressFromFile(mac, path string) (string, error) {
	// Due to https://openradar.appspot.com/FB15382970 we need to parse the MAC
	// address and compare the bytes.
	macAddress, err := parseMAC(mac)
	if err != nil {
		return "", err
	}

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	entries, err := parseLeases(file)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.HWAddress == nil {
			continue
		}
		if bytes.Equal(entry.HWAddress, macAddress) {
			return entry.IPAddress, nil
		}
	}
	return "", fmt.Errorf("could not find an IP address for %s", mac)
}

func parseLeases(r io.Reader) ([]Entry, error) {
	var (
		entry   *Entry
		entries []Entry
	)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "{" {
			entry = new(Entry)
			continue
		} else if line == "}" {
			entries = append(entries, *entry)
			continue
		}

		split := strings.SplitN(line, "=", 2)
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid line in dhcp leases file: %s", line)
		}
		key, val := split[0], split[1]
		switch key {
		case "name":
			entry.Name = val
		case "ip_address":
			entry.IPAddress = val
		case "hw_address":
			// The mac addresses have a '1,' at the start.
			macAddress, err := parseMAC(val[2:])
			if err != nil {
				log.Warnf("unable to parse hw_address in dhcp leases file: %q: %s",
					val[2:], err)
				continue
			}
			entry.HWAddress = macAddress
		case "identifier":
			entry.ID = val
		case "lease":
			entry.Lease = val
		default:
			return entries, fmt.Errorf("unable to parse line: %s", line)
		}
	}
	return entries, scanner.Err()
}

// parseMAC parses both standard fixed-size MAC address "%02x:..." and the
// variable-size MAC address on darwin "%x:...".
func parseMAC(mac string) (net.HardwareAddr, error) {
	hw := make(net.HardwareAddr, 6)
	n, err := fmt.Sscanf(mac, "%x:%x:%x:%x:%x:%x",
		&hw[0], &hw[1], &hw[2], &hw[3], &hw[4], &hw[5])
	if n != len(hw) {
		return nil, fmt.Errorf("invalid MAC address: %q", mac)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to parse MAC address: %q: %s", mac, err)
	}
	return hw, nil
}
