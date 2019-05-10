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

import "net"

// isInBlock checks if ip is a CIDIR block
func isInBlock(ip string, block string) (bool, error) {
	_, b, err := net.ParseCIDR(block)
	if err != nil {
		return false, err
	}
	i := net.ParseIP(ip)
	if b.Contains(i) {
		return true, nil
	}
	return false, nil
}
