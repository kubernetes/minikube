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

package node

import (
	"fmt"
	"net/netip"
	"os/exec"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/out"
)

// primaryInterface is the main network interface in the buildroot-based minikube ISO.
// Update this if the ISO switches to a different distro or naming scheme.
const primaryInterface = "eth0"

// configureDNS configures static DNS servers on the VM, overriding DHCP-provided
// DNS settings. This fixes DNS resolution on managed Macs where network extensions
// block DNS traffic from the VM bridge to the host resolver.
//
// We configure systemd-resolved with two settings:
//
//  1. DNS servers (resolvectl dns eth0 8.8.8.8 1.1.1.1):
//     Sets DNS servers that are reachable from the VM, bypassing the
//     host's broken DNS path.
//
//  2. Routing domain (resolvectl domain eth0 "~."):
//     The "~." syntax tells systemd-resolved to route ALL DNS queries through
//     eth0's DNS servers. The "~" prefix marks it as a routing domain (not a
//     search domain), and "." matches the root domain (everything). Without
//     this, systemd-resolved might still try other interfaces' DNS servers.
func configureDNS(runner command.Runner, servers []netip.Addr) {
	if len(servers) == 0 {
		return
	}

	values := make([]string, len(servers))
	for i, addr := range servers {
		values[i] = addr.String()
	}
	dnsServers := strings.Join(values, " ")

	script := fmt.Sprintf(`
resolvectl dns %s %s
resolvectl domain %s "~."
resolvectl flush-caches
`, primaryInterface, dnsServers, primaryInterface)

	cmd := exec.Command("sudo", "bash", "-o", "errexit", "-c", script)
	if _, err := runner.RunCmd(cmd); err != nil {
		klog.Warningf("Failed to configure static DNS servers: %v", err)
		out.WarningT("Failed to configure static DNS servers")
		return
	}

	klog.Infof("Configured static DNS servers: %s", dnsServers)
}

// configureMDNS enables mDNS (.local address resolution) on the VM by configuring
// systemd-resolved. This allows the guest to resolve other machines on the local
// network that advertise via mDNS.
func configureMDNS(runner command.Runner, enabled bool) {
	if !enabled {
		return
	}
	cmd := exec.Command("sudo", "resolvectl", "mdns", primaryInterface, "yes")
	if _, err := runner.RunCmd(cmd); err != nil {
		klog.Warningf("Failed to enable mDNS on %s: %v", primaryInterface, err)
		out.WarningT("Failed to enable mDNS on {{.iface}}", out.V{"iface": primaryInterface})
		return
	}
	klog.Infof("Enabled mDNS on %s", primaryInterface)
}
