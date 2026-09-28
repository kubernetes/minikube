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

// Package mac generates deterministic MAC addresses from VM names.
package mac

import (
	"crypto/sha256"
	"fmt"
)

// FromName returns a locally administered unicast MAC address derived from
// the given name. The same name always produces the same MAC address.
// https://en.wikipedia.org/wiki/MAC_address#Universal_vs._local_(U/L_bit)
func FromName(name string) string {
	fullName := fmt.Sprintf("minikube-mac-%s", name)
	b := sha256.Sum256([]byte(fullName))
	// | 2 sets the locally-administered bit.
	// & 0xfe clears the multicast bit (unicast only).
	b[0] = (b[0] | 2) & 0xfe
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", b[0], b[1], b[2], b[3], b[4], b[5])
}
