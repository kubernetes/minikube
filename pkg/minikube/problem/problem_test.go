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

package problem

import (
	"fmt"
	"testing"
)

func TestFromError(t *testing.T) {
	var tests = []struct {
		issue int
		os    string
		want  string
		err   string
	}{
		{0, "", "", "this is just a lame error message with no matches."},
		{2991, "", "KVM_UNAVAILABLE", "Unable to start VM: create: Error creating machine: Error in driver during machine creation: creating domain: Error defining domain xml:\n\n: virError(Code=8, Domain=44, Message='invalid argument: could not find capabilities for domaintype=kvm ')"},
		{3594, "", "HOST_CIDR_CONFLICT", "Error starting host: Error starting stopped host: Error setting up host only network on machine start: host-only cidr conflicts with the network address of a host interface."},
		{3614, "", "VBOX_HOST_ADAPTER", "Error starting host:  Error starting stopped host: Error setting up host only network on machine start: The host-only adapter we just created is not visible. This is a well known VirtualBox bug. You might want to uninstall it and reinstall at least version 5.0.12 that is is supposed to fix this issue"},
		{3784, "", "VBOX_NOT_FOUND", "create: precreate: VBoxManage not found. Make sure VirtualBox is installed and VBoxManage is in the path"},
		{3849, "", "IP_NOT_FOUND", "bootstrapper: Error creating new ssh host from driver: Error getting ssh host name for driver: IP not found"},
		{3859, "windows", "VBOX_HARDENING", `Unable to start VM: create: creating: Unable to start the VM: C:\Program Files\Oracle\VirtualBox\VBoxManage.exe startvm minikube --type headless failed:
VBoxManage.exe: error: The virtual machine 'minikube' has terminated unexpectedly during startup with exit code -1073741819 (0xc0000005). More details may be available in 'C:\Users\pabitra_b.minikube\machines\minikube\minikube\Logs\VBoxHardening.log'
VBoxManage.exe: error: Details: code E_FAIL (0x80004005), component MachineWrap, interface IMachine`},
		{3922, "", "ISO_DOWNLOAD_FAILED", `unable to cache ISO: https://storage.googleapis.com/minikube/iso/minikube-v0.35.0.iso: failed to download: failed to download to temp file: download failed: 5 error(s) occurred:
* Temporary download error: Get https://storage.googleapis.com/minikube/iso/minikube-v0.35.0.iso: dial tcp 216.58.207.144:443: connectex: A connection attempt failed because the connected party did not properly respond after a period of time, or established connection failed because connected host has failed to respond.`},
		{4107, "darwin", "VBOX_BLOCKED", "Result Code: NS_ERROR_FAILURE (0x80004005)"},
		{4202, "", "APISERVER_TIMEOUT", "Error restarting cluster: wait: waiting for component=kube-apiserver: timed out waiting for the condition"},
		{4252, "", "DOWNLOAD_TLS_OVERSIZED", "Failed to update cluster: downloading binaries: downloading kubeadm: Error downloading kubeadm v1.14.1: failed to download: failed to download to temp file: download failed: 5 error(s) occurred:\n\nTemporary download error: Get https://storage.googleapis.com/kubernetes-release/release/v1.14.1/bin/linux/amd64/kubeadm: proxyconnect tcp: tls: oversized record received with length 20527"},
		{4222, "", "VBOX_HOST_ADAPTER", "Unable to start VM: create: creating: Error setting up host only network on machine start: The host-only adapter we just created is not visible. This is a well known VirtualBox bug. You might want to uninstall it and reinstall at least version 5.0.12 that is is supposed to fix this issue"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			got := FromError(fmt.Errorf(tc.err), tc.os)
			if got == nil {
				if tc.want != "" {
					t.Errorf("FromError(%q)=nil, want %s", tc.err, tc.want)
				}
				return
			}
			if got.ID != tc.want {
				t.Errorf("FromError(%q)=%s, want %s", tc.err, got.ID, tc.want)
			}

			found := false
			for _, i := range got.Issues {
				if i == tc.issue {
					found = true
				}
			}
			if !found {
				t.Errorf("Issue %d is not listed in %+v", tc.issue, got.Issues)
			}
		})
	}
}
