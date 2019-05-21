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
	"HYPERKIT_NO_IP": {
		Regexp: re(`IP address never found in dhcp leases file Temporary Error: Could not find an IP address for`),
		Advice: "Install the latest minikube hyperkit driver, and run 'minikube delete'",
		URL:    "https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#hyperkit-driver",
		Issues: []int{1926, 4206},
	},
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
	"VBOX_VERR_VMX_NO_VMX": {
		Regexp: re(`VT-x is not available.*VERR_VMX_NO_VMX`),
		Advice: "Please check your BIOS, and ensure that you are running without HyperV or other nested virtualization that may interfere",
		Issues: []int{1994},
	},
	"VBOX_BLOCKED": {
		Regexp: re(`NS_ERROR_FAILURE.*0x80004005`),
		Advice: "Reinstall VirtualBox and verify that it is not blocked: System Preferences -> Security & Privacy -> General -> Some system software was blocked from loading",
		Issues: []int{4107},
		GOOS:   "darwin",
	},
	"VBOX_DRV_NOT_LOADED": {
		Regexp: re(`The vboxdrv kernel module is not loaded`),
		Advice: "Run 'sudo modprobe vboxdrv' and reinstall VirtualBox if it fails.",
		Issues: []int{4043},
	},
	"VBOX_DEVICE_MISSING": {
		Regexp: re(`/dev/vboxdrv does not exist`),
		Advice: "Run 'sudo modprobe vboxdrv' and reinstall VirtualBox if it fails.",
		Issues: []int{3974},
	},
	"VBOX_HARDENING": {
		Regexp: re(`terminated unexpectedly.*VBoxHardening`),
		Advice: "Disable real-time anti-virus software, reboot, and reinstall VirtualBox if the problem continues.",
		Issues: []int{3859, 3910},
		URL:    "https://forums.virtualbox.org/viewtopic.php?f=25&t=82106",
		GOOS:   "windows",
	},
	"VBOX_HOST_ADAPTER": {
		Regexp: re(`The host-only adapter we just created is not visible`),
		Advice: "Reboot to complete VirtualBox installation, and verify that VirtualBox is not blocked by your system",
		Issues: []int{3614, 4222},
		URL:    "https://stackoverflow.com/questions/52277019/how-to-fix-vm-issue-with-minikube-start",
	},
	"VBOX_KERNEL_MODULE_NOT_LOADED": {
		Regexp: re(`The vboxdrv kernel module is not loaded`),
		Advice: "Run 'sudo modprobe vboxdrv' and reinstall VirtualBox if it fails.",
		Issues: []int{4043},
	},
	"KVM2_NOT_FOUND": {
		Regexp: re(`Driver "kvm2" not found. Do you have the plugin binary .* accessible in your PATH`),
		Advice: "Please install the minikube kvm2 VM driver, or select an alternative --vm-driver",
		URL:    "https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#kvm2-driver",
	},
	"KVM2_RESTART_NO_IP": {
		Regexp: re(`Error starting stopped host: Machine didn't return an IP after 120 seconds`),
		Advice: "The KVM driver is unable to resurrect this old VM. Please run `minikube delete` to delete it and try again.",
		Issues: []int{3901, 3434},
	},
	"KVM2_START_NO_IP": {
		Regexp: re(`Error in driver during machine creation: Machine didn't return an IP after 120 seconds`),
		Advice: "Install the latest kvm2 driver and run 'virt-host-validate'",
		URL:    "https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#kvm2-driver",
		Issues: []int{4249, 3566},
	},
	"KVM2_NETWORK_DEFINE_XML": {
		Regexp: re(`not supported by the connection driver: virNetworkDefineXML`),
		Advice: "Rebuild libvirt with virt-network support",
		URL:    "https://forums.gentoo.org/viewtopic-t-981692-start-0.html",
		Issues: []int{4195},
	},
	"KVM2_QEMU_MONITOR": {
		Regexp: re(`qemu unexpectedly closed the monitor`),
		Advice: "Upgrade to QEMU v3.1.0+, run 'virt-host-validate', or ensure that you are not running in a nested VM environment.",
		Issues: []int{4277},
	},
	"KVM_UNAVAILABLE": {
		Regexp: re(`invalid argument: could not find capabilities for domaintype=kvm`),
		Advice: "Your host does not support KVM virtualization. Ensure that qemu-kvm is installed, and run 'virt-host-validate' to debug the problem",
		URL:    "http://mikko.repolainen.fi/documents/virtualization-with-kvm",
		Issues: []int{2991},
	},
	"DRIVER_CRASHED": {
		Regexp: re(`Error attempting to get plugin server address for RPC`),
		Advice: "The VM driver exited with an error, and may be corrupt. Run 'minikube start' with --alsologtostderr -v=8 to see the error",
		URL:    "https://github.com/kubernetes/minikube/blob/master/docs/drivers.md",
	},
	"DRIVER_EXITED": {
		Regexp: re(`Unable to start VM: create: creating: exit status 1`),
		Advice: "Re-run 'minikube start' with --alsologtostderr -v=8 to see the VM driver error message",
		URL:    "https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#troubleshooting",
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
	"HOST_CIDR_CONFLICT": {
		Regexp: re(`host-only cidr conflicts with the network address of a host interface`),
		Advice: "Specify an alternate --host-only-cidr value, such as 172.16.0.1/24",
		Issues: []int{3594},
	},
	"OOM_KILL_SSH": {
		Regexp: re(`Process exited with status 137 from signal KILL`),
		Advice: "Disable dynamic memory in your VM manager, or pass in a larger --memory value",
		Issues: []int{1766},
	},
	"OOM_KILL_SCP": {
		Regexp: re(`An existing connection was forcibly closed by the remote host`),
		Advice: "Disable dynamic memory in your VM manager, or pass in a larger --memory value",
		Issues: []int{1766},
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
		Regexp: re(`tls: oversized record received with length`),
		Advice: "A firewall is interfering with minikube's ability to make outgoing HTTPS requests. You may need to change the value of the HTTPS_PROXY environment variable.",
		URL:    proxyDoc,
		Issues: []int{3857, 3759, 4252},
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
	"APISERVER_TIMEOUT": {
		Regexp: re(`wait: waiting for component=kube-apiserver: timed out waiting for the condition`),
		Advice: "Run 'minikube delete'. If the problem persists, check your proxy or firewall configuration",
		Issues: []int{4202, 3836, 4221},
	},
}

// osProblems are operating-system specific issues
var osProblems = map[string]match{
	"NON_C_DRIVE": {
		Regexp: re(`.iso: The system cannot find the path specified.`),
		Advice: "Run minikube from the C: drive.",
		Issues: []int{1574},
	},
	"SYSTEMCTL_EXIT_1": {
		Regexp: re(`Failed to enable container runtime: .*sudo systemctl start docker: exit status 1`),
		Advice: "If using the none driver, ensure that systemctl is installed",
		URL:    "https://github.com/kubernetes/minikube/blob/master/docs/vmdriver-none.md",
		Issues: []int{2704},
	},
}

// stateProblems are issues relating to local state
var stateProblems = map[string]match{
	"MACHINE_DOES_NOT_EXST": {
		Regexp: re(`Error getting state for host: machine does not exist`),
		Advice: "Run 'minikube delete' to delete the stale VM",
		Issues: []int{3864},
	},
	"IP_NOT_FOUND": {
		Regexp: re(`Error getting ssh host name for driver: IP not found`),
		Advice: "The minikube VM is offline. Please run 'minikube start' to start it again.",
		Issues: []int{3849, 3648},
	},
}
