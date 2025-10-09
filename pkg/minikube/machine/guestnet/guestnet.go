// Copyright 2025 The Kubernetes Authors
// SPDX-License-Identifier: Apache-2.0
//
// Small helpers to detect guest IPv4/IPv6 addresses for VM drivers (e.g., Hyper-V)
// and to enforce IPv6 presence when the user requested --ip-family=ipv6/dual.

package guestnet

import (
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
)

// DetectIPs returns the first global-scoped IPv4 and IPv6 found inside the guest.
// It relies on standard Linux 'ip' output and ignores link-local IPv6 (fe80::/10).
func DetectIPs(r command.Runner) (ipv4 string, ipv6 string, err error) {
	// IPv4: first global address
	cmd4 := exec.Command("sudo", "sh", "-c", `ip -o -4 addr show scope global | awk '{print $4}' | cut -d/ -f1 | head -n1`)
	rr4, err4 := r.RunCmd(cmd4)
	if err4 == nil {
		ipv4 = strings.TrimSpace(rr4.Stdout.String())
	} else {
		klog.V(2).Infof("DetectIPs: ipv4 probe failed: %v", err4)
	}

	// IPv6: first global (non-link-local) address
	cmd6 := exec.Command("sudo", "sh", "-c", `ip -o -6 addr show scope global | awk '{print $4}' | cut -d/ -f1 | grep -v '^fe80' | head -n1`)
	rr6, err6 := r.RunCmd(cmd6)
	if err6 == nil {
		ipv6 = strings.TrimSpace(rr6.Stdout.String())
	} else {
		klog.V(2).Infof("DetectIPs: ipv6 probe failed: %v", err6)
	}

	// Only return an error if both probes failed to execute;
	// missing one family is not an execution error.
	if err4 != nil && err6 != nil {
		return ipv4, ipv6, errors.Errorf("failed to probe guest IPs (ipv4: %v, ipv6: %v)", err4, err6)
	}
	return ipv4, ipv6, nil
}

// RequireIPv6IfRequested returns an error if the cluster ip-family implies IPv6
// but the guest has no detected global IPv6 address. Caller may decide to exit early.
func RequireIPv6IfRequested(cc config.ClusterConfig, detectedIPv6 string) error {
	fam := strings.ToLower(strings.TrimSpace(cc.KubernetesConfig.IPFamily))
	if fam == "ipv6" || fam == "dual" {
		if strings.TrimSpace(detectedIPv6) == "" {
			return errors.Errorf(
				"IPv6/dual-stack requested (--ip-family=%s), but no global IPv6 was detected inside the VM. "+
					"Ensure your Hyper-V/VM switch provides IPv6 and the guest receives a non-link-local IPv6 address.",
				fam,
			)
		}
	}
	return nil
}

