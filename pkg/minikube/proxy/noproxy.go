/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package proxy

import (
	"fmt"
	"os"
	"strings"
)

// UpdateNoProxy is used to whitelist minikube's VM ip from going through proxy
// It updates NO_PROXY environment variable, for the current run.
func UpdateNoProxy(ip string) error {
	yes, v := IsInNoProxyEnv(ip)
	if yes { // skip if already whitelisted
		return nil
	}
	return os.Setenv("NO_PROXY", fmt.Sprintf("%s,%s", v, ip))
}

// IsInNoProxyEnv checks if ip is set in NO_PROXY env variable
// Checks for both IP and IP ranges
func IsInNoProxyEnv(ip string) (bool, string) {
	v := os.Getenv("NO_PROXY")

	if v == "" {
		return false, ""
	}

	//  Checking for IP explicitly, i.e., 192.168.39.224
	if strings.Contains(v, ip) {
		return true, v
	}

	// Checks if ip included in IP ranges, i.e., 192.168.39.13/24
	noProxyBlocks := strings.Split(v, ",")
	for _, b := range noProxyBlocks {
		if yes, _ := isInBlock(ip, b); yes {
			return true, v
		}
	}

	return false, v
}
