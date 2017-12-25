// +build darwin,cgo

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

	vmnet "github.com/zchee/go-vmnet"
)

const (
	CONFIG_PLIST = "/Library/Preferences/SystemConfiguration/com.apple.vmnet"
	NET_ADDR_KEY = "Shared_Net_Address"
)

func GetMACAddressFromUUID(UUID string) (string, error) {
	return vmnet.GetMACAddressFromUUID(UUID)
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
