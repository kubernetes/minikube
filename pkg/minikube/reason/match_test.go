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

package reason

import (
	"fmt"
	"testing"
)

func TestFromError(t *testing.T) {
	tests := []struct {
		issue int
		os    string
		want  string
		err   string
	}{
		{0, "", "", "this is just a lame error message with no matches."},
		{2991, "linux", "PR_KVM_CAPABILITIES", "Unable to start VM: create: Error creating machine: Error in driver during machine creation: creating domain: Error defining domain xml:\n\n: virError(Code=8, Domain=44, Message='invalid argument: could not find capabilities for domaintype=kvm ')"},
		{3594, "", "IF_HOST_CIDR_CONFLICT", "Error starting host: Error starting stopped host: Error setting up host only network on machine start: host-only cidr conflicts with the network address of a host interface."},
		{3614, "", "IF_VBOX_NOT_VISIBLE", "Error starting host:  Error starting stopped host: Error setting up host only network on machine start: The host-only adapter we just created is not visible. This is a well known VirtualBox bug. You might want to uninstall it and reinstall at least version 5.0.12 that is supposed to fix this issue"},
		{3784, "", "PR_VBOXMANAGE_NOT_FOUND", "create: precreate: VBoxManage not found. Make sure VirtualBox is installed and VBoxManage is in the path"},
		{3849, "", "GUEST_IP_NOT_FOUND", "bootstrapper: Error creating new ssh host from driver: Error getting ssh host name for driver: IP not found"},
		{3859, "windows", "PR_VBOX_HARDENING", `Unable to start VM: create: creating: Unable to start the VM: C:\Program Files\Oracle\VirtualBox\VBoxManage.exe startvm minikube --type headless failed:
VBoxManage.exe: error: The virtual machine 'minikube' has terminated unexpectedly during startup with exit code -1073741819 (0xc0000005). More details may be available in 'C:\Users\pabitra_b.minikube\machines\minikube\minikube\Logs\VBoxHardening.log'
VBoxManage.exe: error: Details: code E_FAIL (0x80004005), component MachineWrap, interface IMachine`},
		{3922, "", "INET_DOWNLOAD_BLOCKED", `unable to cache ISO: https://storage.googleapis.com/minikube/iso/minikube-v0.35.0.iso: failed to download: failed to download to temp file: download failed: 5 error(s) occurred:
* Temporary download error: Get https://storage.googleapis.com/minikube/iso/minikube-v0.35.0.iso: dial tcp 216.58.207.144:443: connectex: A connection attempt failed because the connected party did not properly respond after a period of time, or established connection failed because connected host has failed to respond.`},
		{4107, "darwin", "PR_VBOX_BLOCKED", "Result Code: NS_ERROR (0x80004005)"},
		{4302, "", "K8S_APISERVER_TIMEOUT", "apiserver: timed out waiting for the condition"},
		{4252, "", "INET_TLS_OVERSIZED", "Failed to update cluster: downloading binaries: downloading kubeadm: Error downloading kubeadm v1.14.1: failed to download: failed to download to temp file: download failed: 5 error(s) occurred:\n\nTemporary download error: Get https://storage.googleapis.com/kubernetes-release/release/v1.14.1/bin/linux/amd64/kubeadm: proxyconnect tcp: tls: oversized record received with length 20527"},
		{4222, "", "IF_VBOX_NOT_VISIBLE", "Unable to start VM: create: creating: Error setting up host only network on machine start: The host-only adapter we just created is not visible. This is a well known VirtualBox bug. You might want to uninstall it and reinstall at least version 5.0.12 that is supposed to fix this issue"},
		{6014, "linux", "K8S_APISERVER_MISSING", "Error restarting cluster: waiting for apiserver: apiserver process never appeared"},
		{5836, "", "SVC_OPEN_NOT_FOUND", `Error opening service: Service newservice was not found in "unknown" namespace. You may select another namespace by using 'minikube service newservice -n : Temporary Error: Error getting service newservice: services "newservice" not found`},
		{6087, "", "GUEST_DOES_NOT_EXIST", `Error getting machine status: state: machine does not exist`},
		{5714, "darwin", "HOST_KUBECONFIG_PERMISSION", `Failed to setup kubeconfig: writing kubeconfig: Error writing file /Users/matthewgleich/.kube/config: error writing file /Users/matthewgleich/.kube/config: open /Users/matthewgleich/.kube/config: permission denied`},
		{5532, "linux", "RT_DOCKER_EXIT_5", `Failed to enable container runtime: running command: sudo systemctl start docker: exit status 5`},
		{5532, "linux", "RT_CRIO_EXIT_5", `Failed to enable container runtime: running command: sudo systemctl restart crio: exit status 5`},
		{5484, "linux", "GUEST_PORT_IN_USE", `[ERROR Port-10252]: Port 10252 is in use`},
		{4913, "linux", "PR_KVM_CREATE_BUSY", `Unable to start VM: create: Error creating machine: Error in driver during machine creation: error creating VM: virError(Code=1, Domain=10, Message='internal error: process exited while connecting to monitor: ioctl(KVM_CREATE_VM) failed: 16 Device or resource busy`},
		{5950, "linux", "PR_KVM_ISO_PERMISSION", `Retriable failure: create: Error creating machine: Error in driver during machine creation: error creating VM: virError(Code=1, Domain=10, Message='internal error: qemu unexpectedly closed the monitor: 2019-11-19T16:08:16.757609Z qemu-kvm: -drive file=/home/lnicotra/.minikube/machines/minikube/boot2docker.iso,format=raw,if=none,id=drive-scsi0-0-0-2,readonly=on: could not open disk image /home/lnicotra/.minikube/machines/minikube/boot2docker.iso: Could not open '/home/lnicotra/.minikube/machines/minikube/boot2docker.iso': Permission denied'`},
		{5836, "", "SVC_OPEN_NOT_FOUND", `Error opening service: Service kubernetes-bootcamp was not found in "default" namespace. You may select another namespace by using 'minikube service kubernetes-bootcamp -n : Temporary Error: Error getting service kubernetes-bootcamp: services "kubernetes-bootcamp" not found`},
		{3898, "", "INET_PULL_TIMEOUT", `[ERROR ImagePull]: failed to pull image k8s.gcr.io/kube-controller-manager:v1.17.0: output: Error response from daemon: Get https://k8s.gcr.io/v2/: net/http: request canceled while waiting for connection (Client.Timeout exceeded while awaiting headers)`},
		{6079, "darwin", "PR_HYPERKIT_CRASHED", `Error creating machine: Error in driver during machine creation: hyperkit crashed! command line:`},
		{5636, "linux", "INET_DEFAULT_ROUTE", `Unable to get VM IP address: unable to select an IP from default routes.`},
		{6087, "", "GUEST_DOES_NOT_EXIST", `Error getting host status: state: machine does not exist`},
		{6098, "windows", "PR_PRECREATE_EXIT_1", `Retriable failure: create: precreate: exit status 1`},
		{6107, "", "INET_PROXY_CONFUSION", `http: server gave HTTP response to HTTPS client`},
		{6109, "", "INET_DOWNLOAD_BLOCKED", `Failed to update cluster: downloading binaries: downloading kubelet: Error downloading kubelet v1.16.2: failed to download: failed to download to temp file: failed to copy contents: read tcp 192.168.0.106:61314->172.217.166.176:443: wsarecv: A connection attempt failed because the connected party did not properly respond after a period of time, or established connection failed because connected host has failed to respond.`},
		{6109, "", "INET_DOWNLOAD_BLOCKED", `Failed to update cluster: downloading binaries: downloading kubeadm: Error downloading kubeadm v1.17.0: failed to download: failed to download to temp file: failed to copy contents: read tcp [2606:a000:81c5:1e00:349a:26c0:7ea6:bbf1]:55317->[2607:f8b0:4004:815::2010]:443: wsarecv: A connection attempt failed because the connected party did not properly respond after a period of time, or established connection failed because connected host has failed to respond.`},
		{4277, "linux", "PR_KVM_MSR", `Unable to start VM: start: Error creating VM: virError(Code=1, Domain=10, Message='internal error: qemu unexpectedly closed the monitor: 2019-05-17T02:20:07.980140Z qemu-system-x86_64: error: failed to set MSR 0x38d to 0x0 qemu-system-x86_64: /build/qemu-lXHhGe/qemu-2.11+dfsg/target/i386/kvm.c:1807: kvm_put_msrs: Assertion ret == cpu->kvm_msr_buf->nmsrs failed.`},
		{8832, "macos", "PR_DOCKER_MOUNTS_EOF", `docker: Error response from daemon: Mounts denied: EOF.`},
		{9165, "", "HOST_HOME_PERMISSION", `open /Users/conradwt/.minikube/profiles/gcloud-local-dev/config.json: permission denied`},
		{9175, "", "GUEST_CONFIG_CORRUPT", " updating control plane: generating kubeadm cfg: generating extra component config for kubeadm: controlPlane configuration is corrupt: no name: {Name: IP: Port:8443 KubernetesVersion:v1.19.0 ControlPlane:true Worker:true}"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			got := MatchKnownIssue(Kind{}, fmt.Errorf(tc.err), tc.os)
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
