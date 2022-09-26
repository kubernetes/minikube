//go:build darwin

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
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

const (
	// VMNetDomain is the domain for vmnet
	VMNetDomain = "/Library/Preferences/SystemConfiguration/com.apple.vmnet"
	// SharedNetAddrKey is the key for the network address
	SharedNetAddrKey = "Shared_Net_Address"
)

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
