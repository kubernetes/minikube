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

import "regexp"

// re is a shortcut around regexp.MustCompile
func re(s string) *regexp.Regexp {
	return regexp.MustCompile(s)
}

// vmProblems are VM related problems
var vmProblems = map[string]match{
	"VBOX_NOT_FOUND": {
		Regexp: re(`VBoxManage not found. Make sure VirtualBox is installed and VBoxManage is in the path`),
		Advice: "Install VirtualBox, ensure that VBoxManage is executable and in path, or select an alternative value for --vm-driver",
		URL:    "https://www.virtualbox.org/wiki/Downloads",
		Issues: []int{3784, 3776},
	},
	"VBOX_VTX_DISABLED": {
		Regexp: re(`This computer doesn't have VT-X/AMD-v enabled`),
		Advice: "In some environments, this message is incorrect. Try 'minikube start --no-vtx-check'",
		Issues: []int{3900},
	},
	"VBOX_THIRD_PARTY": {
		Regexp: re(`The virtual machine * has terminated unexpectedly during startup with exit code 1`),
		Advice: "A third-party program may be interfering with VirtualBox. Try disabling any real-time antivirus software, reinstalling VirtualBox and rebooting.",
		Issues: []int{3910},
	},
	"KVM2_NOT_FOUND": {
		Regexp: re(`Driver "kvm2" not found. Do you have the plugin binary .* accessible in your PATH`),
		Advice: "Please install the minikube kvm2 VM driver, or select an alternative --vm-driver",
		URL:    "https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#kvm2-driver",
	},
	"KVM2_NO_IP": {
		Regexp: re(`Error starting stopped host: Machine didn't return an IP after 120 seconds`),
		Advice: "The KVM driver is unable to resurrect this old VM. Please run `minikube delete` to delete it and try again.",
		Issues: []int{3901, 3566, 3434},
	},
	"VM_DOES_NOT_EXIST": {
		Regexp: re(`Error getting state for host: machine does not exist`),
		Advice: "Your system no longer knows about the VM previously created by minikube. Run 'minikube delete' to reset your local state.",
		Issues: []int{3864},
	},
	"VM_IP_NOT_FOUND": {
		Regexp: re(`Error getting ssh host name for driver: IP not found`),
		Advice: "The minikube VM is offline. Please run 'minikube start' to start it again.",
		Issues: []int{3849, 3648},
	},
	"VM_BOOT_FAILED_HYPERV_ENABLED": {
		Regexp: re(`VirtualBox won't boot a 64bits VM when Hyper-V is activated`),
		Advice: "Disable Hyper-V when you want to run VirtualBox to boot the VM",
		Issues: []int{4051},
	},
	"HYPERV_NO_VSWITCH": {
		Regexp: re(`no External vswitch found. A valid vswitch must be available for this command to run.`),
		Advice: "Configure an external network switch following the official documentation, then add `--hyperv-virtual-switch=<switch-name>` to `minikube start`",
		URL:    "https://docs.docker.com/machine/drivers/hyper-v/",
	},
}

// proxyDoc is the URL to proxy documentation
const proxyDoc = "https://github.com/kubernetes/minikube/blob/master/docs/http_proxy.md"

// netProblems are network related problems.
var netProblems = map[string]match{
	"GCR_UNAVAILABLE": {
		Regexp: re(`gcr.io.*443: connect: invalid argument`),
		Advice: "minikube is unable to access the Google Container Registry. You may need to configure it to use a HTTP proxy.",
		URL:    proxyDoc,
		Issues: []int{3860},
	},
	"DOWNLOAD_RESET_BY_PEER": {
		Regexp: re(`Error downloading .*connection reset by peer`),
		Advice: "A firewall is likely blocking minikube from reaching the internet. You may need to configure minikube to use a proxy.",
		URL:    proxyDoc,
		Issues: []int{3909},
	},
	"DOWNLOAD_IO_TIMEOUT": {
		Regexp: re(`Error downloading .*timeout`),
		Advice: "A firewall is likely blocking minikube from reaching the internet. You may need to configure minikube to use a proxy.",
		URL:    proxyDoc,
		Issues: []int{3846},
	},
	"DOWNLOAD_TLS_OVERSIZED": {
		Regexp: re(`failed to download.*tls: oversized record received with length`),
		Advice: "A firewall is interfering with minikube's ability to make outgoing HTTPS requests. You may need to configure minikube to use a proxy.",
		URL:    proxyDoc,
		Issues: []int{3857, 3759},
	},
	"ISO_DOWNLOAD_FAILED": {
		Regexp: re(`iso: failed to download`),
		Advice: "A firewall is likely blocking minikube from reaching the internet. You may need to configure minikube to use a proxy.",
		URL:    proxyDoc,
		Issues: []int{3922},
	},
	"PULL_TIMEOUT_EXCEEDED": {
		Regexp: re(`failed to pull image k8s.gcr.io.*Client.Timeout exceeded while awaiting headers`),
		Advice: "A firewall is blocking Docker within the minikube VM from reaching the internet. You may need to configure it to use a proxy.",
		URL:    proxyDoc,
		Issues: []int{3898},
	},
	"SSH_AUTH_FAILURE": {
		Regexp: re(`ssh: handshake failed: ssh: unable to authenticate.*, no supported methods remain`),
		Advice: "Your host is failing to route packets to the minikube VM. If you have VPN software, try turning it off or configuring it so that it does not re-route traffic to the VM IP. If not, check your VM environment routing options.",
		Issues: []int{3930},
	},
	"SSH_TCP_FAILURE": {
		Regexp: re(`dial tcp .*:22: connectex: A connection attempt failed because the connected party did not properly respond`),
		Advice: "Your host is failing to route packets to the minikube VM. If you have VPN software, try turning it off or configuring it so that it does not re-route traffic to the VM IP. If not, check your VM environment routing options.",
		Issues: []int{3388},
	},
	"INVALID_PROXY_HOSTNAME": {
		Regexp: re(`dial tcp: lookup.*: no such host`),
		Advice: "Verify that your HTTP_PROXY and HTTPS_PROXY environment variables are set correctly.",
		URL:    proxyDoc,
	},
}

// deployProblems are Kubernetes deployment problems.
var deployProblems = map[string]match{
	"DOCKER_UNAVAILABLE": {
		Regexp: re(`Error configuring auth on host: OS type not recognized`),
		Advice: "Docker inside the VM is unavailable. Try running 'minikube delete' to reset the VM.",
		Issues: []int{3952},
	},
	"INVALID_KUBERNETES_VERSION": {
		Regexp: re(`No Major.Minor.Patch elements found`),
		Advice: "Specify --kubernetes-version in v<major>.<minor.<build> form. example: 'v1.1.14'",
	},
	"KUBERNETES_VERSION_MISSING_V": {
		Regexp: re(`strconv.ParseUint: parsing "": invalid syntax`),
		Advice: "Check that your --kubernetes-version has a leading 'v'. For example: 'v1.1.14'",
	},
}

// osProblems are operating-system specific issues
var osProblems = map[string]match{
	"NON_C_DRIVE": {
		Regexp: re(`.iso: The system cannot find the path specified.`),
		Advice: "Run minikube from the C: drive.",
		Issues: []int{1574},
	},
}
