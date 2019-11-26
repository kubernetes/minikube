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
	// Generic VM driver
	"DRIVER_CORRUPT": {
		Regexp: re(`Error attempting to get plugin server address for RPC`),
		Advice: "The VM driver exited with an error, and may be corrupt. Run 'minikube start' with --alsologtostderr -v=8 to see the error",
		URL:    "https://minikube.sigs.k8s.io/docs/reference/drivers/",
	},
	"DRIVER_EXITED": {
		Regexp: re(`Unable to start VM: start: exit status 1`),
		Advice: "The VM driver crashed. Run 'minikube start --alsologtostderr -v=8' to see the VM driver error message",
		URL:    "https://minikube.sigs.k8s.io/docs/reference/drivers/#troubleshooting",
	},
	"DRIVER_NOT_FOUND": {
		Regexp:         re(`registry: driver not found`),
		Advice:         "Your minikube config refers to an unsupported driver. Erase ~/.minikube, and try again.",
		Issues:         []int{5295},
		HideCreateLink: true,
	},
	"MACHINE_NOT_FOUND": {
		Regexp: re(`Machine does not exist for api.Exists`),
		Advice: "Your minikube vm is not running, try minikube start.",
		Issues: []int{4889},
	},

	// Hyperkit
	"HYPERKIT_NO_IP": {
		Regexp: re(`IP address never found in dhcp leases file Temporary Error: Could not find an IP address for`),
		Advice: "Install the latest hyperkit binary, and run 'minikube delete'",
		URL:    "https://minikube.sigs.k8s.io/docs/reference/drivers/hyperkit/",
		Issues: []int{1926, 4206},
	},
	"HYPERKIT_NOT_FOUND": {
		Regexp:         re(`Driver "hyperkit" not found.`),
		Advice:         "Please install the minikube hyperkit VM driver, or select an alternative --vm-driver",
		URL:            "https://minikube.sigs.k8s.io/docs/reference/drivers/hyperkit/",
		HideCreateLink: true,
	},

	// Hyper-V
	"HYPERV_NO_VSWITCH": {
		Regexp:         re(`no External vswitch found. A valid vswitch must be available for this command to run.`),
		Advice:         "Configure an external network switch following the official documentation, then add `--hyperv-virtual-switch=<switch-name>` to `minikube start`",
		URL:            "https://docs.docker.com/machine/drivers/hyper-v/",
		HideCreateLink: true,
	},
	"HYPERV_VSWITCH_NOT_FOUND": {
		Regexp:         re(`precreate: vswitch.*not found`),
		Advice:         "Confirm that you have supplied the correct value to --hyperv-virtual-switch using the 'Get-VMSwitch' command",
		URL:            "https://docs.docker.com/machine/drivers/hyper-v/",
		HideCreateLink: true,
	},
	"HYPERV_POWERSHELL_NOT_FOUND": {
		Regexp:         re(`Powershell was not found in the path`),
		Advice:         "To start minikube with HyperV Powershell must be in your PATH`",
		URL:            "https://docs.docker.com/machine/drivers/hyper-v/",
		HideCreateLink: true,
	},
	"HYPERV_AS_ADMIN": {
		Regexp:         re(`Hyper-v commands have to be run as an Administrator`),
		Advice:         "Run the minikube command as an Administrator",
		URL:            "https://rominirani.com/docker-machine-windows-10-hyper-v-troubleshooting-tips-367c1ea73c24",
		Issues:         []int{4511},
		HideCreateLink: true,
	},

	// KVM
	"KVM2_NOT_FOUND": {
		Regexp: re(`Driver "kvm2" not found. Do you have the plugin binary .* accessible in your PATH`),
		Advice: "Please install the minikube kvm2 VM driver, or select an alternative --vm-driver",
		URL:    "https://minikube.sigs.k8s.io/docs/reference/drivers/kvm2/",
	},
	"KVM2_NO_DOMAIN": {
		Regexp: re(`no domain with matching name`),
		Advice: "The VM that minikube is configured for no longer exists. Run 'minikube delete'",
		Issues: []int{3636},
	},
	"KVM_CREATE_CONFLICT": {
		Regexp: re(`KVM_CREATE_VM.* failed:.* Device or resource busy`),
		Advice: "There appears to be another hypervisor conflicting with KVM. Please stop the other hypervisor, or use --vm-driver to switch to it.",
		Issues: []int{4913},
	},
	"KVM2_RESTART_NO_IP": {
		Regexp: re(`Error starting stopped host: Machine didn't return an IP after 120 seconds`),
		Advice: "The KVM driver is unable to resurrect this old VM. Please run `minikube delete` to delete it and try again.",
		Issues: []int{3901, 3434},
	},
	"KVM2_START_NO_IP": {
		Regexp: re(`Error in driver during machine creation: Machine didn't return an IP after 120 seconds`),
		Advice: "Check your firewall rules for interference, and run 'virt-host-validate' to check for KVM configuration issues. If you are running minikube within a VM, consider using --vm-driver=none",
		URL:    "https://minikube.sigs.k8s.io/docs/reference/drivers/kvm2/",
		Issues: []int{4249, 3566},
	},
	"KVM2_NETWORK_DEFINE_XML": {
		Regexp:         re(`not supported by the connection driver: virNetworkDefineXML`),
		Advice:         "Rebuild libvirt with virt-network support",
		URL:            "https://forums.gentoo.org/viewtopic-t-981692-start-0.html",
		Issues:         []int{4195},
		HideCreateLink: true,
	},
	"KVM2_QEMU_MONITOR": {
		Regexp:         re(`qemu unexpectedly closed the monitor`),
		Advice:         "Upgrade to QEMU v3.1.0+, run 'virt-host-validate', or ensure that you are not running in a nested VM environment.",
		Issues:         []int{4277},
		HideCreateLink: true,
	},
	"KVM_UNAVAILABLE": {
		Regexp:         re(`invalid argument: could not find capabilities for domaintype=kvm`),
		Advice:         "Your host does not support KVM virtualization. Ensure that qemu-kvm is installed, and run 'virt-host-validate' to debug the problem",
		URL:            "http://mikko.repolainen.fi/documents/virtualization-with-kvm",
		Issues:         []int{2991},
		HideCreateLink: true,
	},
	"KVM_CONNECTION_ERROR": {
		Regexp:         re(`error connecting to libvirt socket`),
		Advice:         "Have you set up libvirt correctly?",
		URL:            "https://minikube.sigs.k8s.io/docs/reference/drivers/kvm2/",
		HideCreateLink: true,
	},

	// None
	"SYSTEMCTL_EXIT_1": {
		Regexp:         re(`sudo systemctl start docker: exit status 1`),
		Advice:         "Either systemctl is not installed, or Docker is broken. Run 'sudo systemctl start docker' and 'journalctl -u docker'",
		URL:            "https://minikube.sigs.k8s.io/docs/reference/drivers/none",
		Issues:         []int{4498},
		HideCreateLink: true,
	},

	// VirtualBox
	"VBOX_BLOCKED": {
		Regexp:         re(`NS_ERROR_FAILURE.*0x80004005`),
		Advice:         "Reinstall VirtualBox and verify that it is not blocked: System Preferences -> Security & Privacy -> General -> Some system software was blocked from loading",
		Issues:         []int{4107},
		GOOS:           "darwin",
		HideCreateLink: true,
	},
	"VBOX_DRV_NOT_LOADED": {
		Regexp:         re(`vboxdrv kernel module is not loaded`),
		Advice:         "Reinstall VirtualBox and reboot. Alternatively, try the kvm2 driver: https://minikube.sigs.k8s.io/docs/reference/drivers/kvm2/",
		HideCreateLink: true,
		Issues:         []int{4043, 4711},
	},
	"VBOX_DEVICE_MISSING": {
		Regexp:         re(`vboxdrv does not exist`),
		Advice:         "Reinstall VirtualBox and reboot. Alternatively, try the kvm2 driver: https://minikube.sigs.k8s.io/docs/reference/drivers/kvm2/",
		Issues:         []int{3974},
		HideCreateLink: true,
	},
	"VBOX_HARDENING": {
		Regexp:         re(`terminated unexpectedly.*VBoxHardening`),
		Advice:         "VirtualBox is broken. Disable real-time anti-virus software, reboot, and reinstall VirtualBox if the problem continues.",
		Issues:         []int{3859, 3910},
		URL:            "https://forums.virtualbox.org/viewtopic.php?f=25&t=82106",
		GOOS:           "windows",
		HideCreateLink: true,
	},
	"VBOX_NS_ERRROR": {
		Regexp:         re(`terminated unexpectedly.*NS_ERROR_FAILURE.*0x80004005`),
		Advice:         "VirtualBox is broken. Reinstall VirtualBox, reboot, and run 'minikube delete'.",
		Issues:         []int{5227},
		GOOS:           "linux",
		HideCreateLink: true,
	},
	"VBOX_HOST_ADAPTER": {
		Regexp:         re(`The host-only adapter we just created is not visible`),
		Advice:         "Reboot to complete VirtualBox installation, and verify that VirtualBox is not blocked by your system",
		Issues:         []int{3614, 4222},
		URL:            "https://stackoverflow.com/questions/52277019/how-to-fix-vm-issue-with-minikube-start",
		HideCreateLink: true,
	},
	"VBOX_IP_CONFLICT": {
		Regexp: re(`VirtualBox is configured with multiple host-only adapters with the same IP`),
		Advice: "Use VirtualBox to remove the conflicting VM and/or network interfaces",
		URL:    "https://stackoverflow.com/questions/55573426/virtualbox-is-configured-with-multiple-host-only-adapters-with-the-same-ip-whe",
		Issues: []int{3584},
	},
	"VBOX_HYPERV_64_BOOT": {
		Regexp:         re(`VirtualBox won't boot a 64bits VM when Hyper-V is activated`),
		Advice:         "VirtualBox and Hyper-V are having a conflict. Use '--vm-driver=hyperv' or disable Hyper-V using: 'bcdedit /set hypervisorlaunchtype off'",
		Issues:         []int{4051, 4783},
		HideCreateLink: true,
	},
	"VBOX_HYPERV_NEM_VM": {
		Regexp:         re(`vrc=VERR_NEM_VM_CREATE_FAILED`),
		Advice:         "VirtualBox and Hyper-V are having a conflict. Use '--vm-driver=hyperv' or disable Hyper-V using: 'bcdedit /set hypervisorlaunchtype off'",
		Issues:         []int{4587},
		HideCreateLink: true,
	},
	"VBOX_NOT_FOUND": {
		Regexp:         re(`VBoxManage not found. Make sure VirtualBox is installed and VBoxManage is in the path`),
		Advice:         "Install VirtualBox, or select an alternative value for --vm-driver",
		URL:            "https://minikube.sigs.k8s.io/docs/start/",
		Issues:         []int{3784},
		HideCreateLink: true,
	},
	"VBOX_NO_VM": {
		Regexp:         re(`Could not find a registered machine named`),
		Advice:         "The VM that minikube is configured for no longer exists. Run 'minikube delete'",
		Issues:         []int{4694},
		HideCreateLink: true,
	},
	"VBOX_VTX_DISABLED": {
		Regexp:         re(`This computer doesn't have VT-X/AMD-v enabled`),
		Advice:         "Virtualization support is disabled on your computer. If you are running minikube within a VM, try '--vm-driver=none'. Otherwise, consult your systems BIOS manual for how to enable virtualization.",
		Issues:         []int{3900, 4730},
		HideCreateLink: true,
	},
	"VERR_VERR_VMX_DISABLED": {
		Regexp:         re(`VT-x is disabled.*VERR_VMX_MSR_ALL_VMX_DISABLED`),
		Advice:         "Virtualization support is disabled on your computer. If you are running minikube within a VM, try '--vm-driver=none'. Otherwise, consult your systems BIOS manual for how to enable virtualization.",
		Issues:         []int{5282, 5456},
		HideCreateLink: true,
	},
	"VBOX_VERR_VMX_NO_VMX": {
		Regexp:         re(`VT-x is not available.*VERR_VMX_NO_VMX`),
		Advice:         "Your host does not support virtualization. If you are running minikube within a VM, try '--vm-driver=none'. Otherwise, enable virtualization in your BIOS",
		Issues:         []int{1994, 5326},
		HideCreateLink: true,
	},
	"VBOX_HOST_NETWORK": {
		Regexp: re(`Error setting up host only network on machine start.*Unspecified error`),
		Advice: "VirtualBox cannot create a network, probably because it conflicts with an existing network that minikube no longer knows about. Try running 'minikube delete'",
		Issues: []int{5260},
	},
}

// proxyDoc is the URL to proxy documentation
const proxyDoc = "https://minikube.sigs.k8s.io/docs/reference/networking/proxy/"
const vpnDoc = "https://minikube.sigs.k8s.io/docs/reference/networking/vpn/"

// netProblems are network related problems.
var netProblems = map[string]match{
	"GCR_UNAVAILABLE": {
		Regexp:         re(`gcr.io.*443: connect: invalid argument`),
		Advice:         "minikube is unable to access the Google Container Registry. You may need to configure it to use a HTTP proxy.",
		URL:            proxyDoc,
		Issues:         []int{3860},
		HideCreateLink: true,
	},
	"DOWNLOAD_RESET_BY_PEER": {
		Regexp:         re(`Error downloading .*connection reset by peer`),
		Advice:         "A firewall is likely blocking minikube from reaching the internet. You may need to configure minikube to use a proxy.",
		URL:            proxyDoc,
		Issues:         []int{3909},
		HideCreateLink: true,
	},
	"DOWNLOAD_IO_TIMEOUT": {
		Regexp:         re(`Error downloading .*timeout`),
		Advice:         "A firewall is likely blocking minikube from reaching the internet. You may need to configure minikube to use a proxy.",
		URL:            proxyDoc,
		Issues:         []int{3846},
		HideCreateLink: true,
	},
	"DOWNLOAD_TLS_OVERSIZED": {
		Regexp:         re(`tls: oversized record received with length`),
		Advice:         "A firewall is interfering with minikube's ability to make outgoing HTTPS requests. You may need to change the value of the HTTPS_PROXY environment variable.",
		URL:            proxyDoc,
		Issues:         []int{3857, 3759, 4252},
		HideCreateLink: true,
	},
	"ISO_DOWNLOAD_FAILED": {
		Regexp:         re(`iso: failed to download`),
		Advice:         "A firewall is likely blocking minikube from reaching the internet. You may need to configure minikube to use a proxy.",
		URL:            proxyDoc,
		Issues:         []int{3922},
		HideCreateLink: true,
	},
	"PULL_TIMEOUT_EXCEEDED": {
		Regexp:         re(`failed to pull image k8s.gcr.io.*Client.Timeout exceeded while awaiting headers`),
		Advice:         "A firewall is blocking Docker within the minikube VM from reaching the internet. You may need to configure it to use a proxy.",
		URL:            proxyDoc,
		Issues:         []int{3898},
		HideCreateLink: true,
	},
	"SSH_AUTH_FAILURE": {
		Regexp: re(`ssh: handshake failed: ssh: unable to authenticate.*, no supported methods remain`),
		Advice: "Your host is failing to route packets to the minikube VM. If you have VPN software, try turning it off or configuring it so that it does not re-route traffic to the VM IP. If not, check your VM environment routing options.",
		URL:    vpnDoc,
		Issues: []int{3930},
	},
	"SSH_TCP_FAILURE": {
		Regexp: re(`dial tcp .*:22: connectex: A connection attempt failed because the connected party did not properly respond`),
		Advice: "Your host is failing to route packets to the minikube VM. If you have VPN software, try turning it off or configuring it so that it does not re-route traffic to the VM IP. If not, check your VM environment routing options.",
		URL:    vpnDoc,
		Issues: []int{3388},
	},
	"INVALID_PROXY_HOSTNAME": {
		Regexp: re(`dial tcp: lookup.*: no such host`),
		Advice: "Verify that your HTTP_PROXY and HTTPS_PROXY environment variables are set correctly.",
		URL:    proxyDoc,
	},
	"HOST_CIDR_CONFLICT": {
		Regexp:         re(`host-only cidr conflicts with the network address of a host interface`),
		Advice:         "Specify an alternate --host-only-cidr value, such as 172.16.0.1/24",
		Issues:         []int{3594},
		HideCreateLink: true,
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
		Regexp:         re(`No Major.Minor.Patch elements found`),
		Advice:         "Specify --kubernetes-version in v<major>.<minor.<build> form. example: 'v1.1.14'",
		HideCreateLink: true,
	},
	"KUBERNETES_VERSION_MISSING_V": {
		Regexp: re(`strconv.ParseUint: parsing "": invalid syntax`),
		Advice: "Check that your --kubernetes-version has a leading 'v'. For example: 'v1.1.14'",
	},
	"APISERVER_NEVER_APPEARED": {
		Regexp: re(`apiserver process never appeared`),
		Advice: "Check that your apiserver flags are valid, or run 'minikube delete'",
		Issues: []int{4536},
	},
	"APISERVER_TIMEOUT": {
		Regexp: re(`apiserver: timed out waiting for the condition`),
		Advice: "A VPN or firewall is interfering with HTTP access to the minikube VM. Alternatively, try a different VM driver: https://minikube.sigs.k8s.io/docs/start/",
		URL:    vpnDoc,
		Issues: []int{4302},
	},
	"DNS_TIMEOUT": {
		Regexp: re(`dns: timed out waiting for the condition`),
		Advice: "Run 'kubectl describe pod coredns -n kube-system' and check for a firewall or DNS conflict",
		URL:    vpnDoc,
	},
	"SERVICE_NOT_FOUND": {
		Regexp: re(`Could not find finalized endpoint being pointed to by`),
		Advice: "Please make sure the service you are looking for is deployed or is in the correct namespace.",
		Issues: []int{4599},
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
	"PROXY_UNEXPECTED_503": {
		Regexp: re(`proxy.*unexpected response code: 503`),
		Advice: "Confirm that you have a working internet connection and that your VM has not run out of resources by using: 'minikube logs'",
		Issues: []int{4749},
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
		Regexp:         re(`Failed to enable container runtime: .*sudo systemctl start docker: exit status 1`),
		Advice:         "If using the none driver, ensure that systemctl is installed",
		URL:            "https://minikube.sigs.k8s.io/docs/reference/drivers/none/",
		Issues:         []int{2704},
		HideCreateLink: true,
	},
	"KUBECONFIG_WRITE_FAIL": {
		Regexp: re(`Failed to setup kubeconfig: writing kubeconfig`),
		Advice: "Unset the KUBECONFIG environment variable, or verify that it does not point to an empty or otherwise invalid path",
		Issues: []int{5268, 4100, 5207},
	},
}

// stateProblems are issues relating to local state
var stateProblems = map[string]match{
	"MACHINE_DOES_NOT_EXST": {
		Regexp:         re(`Error getting state for host: machine does not exist`),
		Advice:         "Run 'minikube delete' to delete the stale VM",
		Issues:         []int{3864},
		HideCreateLink: true,
	},
	"IP_NOT_FOUND": {
		Regexp: re(`Error getting ssh host name for driver: IP not found`),
		Advice: "The minikube VM is offline. Please run 'minikube start' to start it again.",
		Issues: []int{3849, 3648},
	},
}
