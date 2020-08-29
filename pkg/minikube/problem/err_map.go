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
		Regexp:        re(`Error attempting to get plugin server address for RPC`),
		Advice:        "The VM driver exited with an error, and may be corrupt. Run 'minikube start' with --alsologtostderr -v=8 to see the error",
		URL:           "https://minikube.sigs.k8s.io/docs/reference/drivers/",
		ShowIssueLink: true,
	},
	"DRIVER_EXITED": {
		Regexp:        re(`Unable to start VM: start: exit status 1`),
		Advice:        "The VM driver crashed. Run 'minikube start --alsologtostderr -v=8' to see the VM driver error message",
		URL:           "https://minikube.sigs.k8s.io/docs/reference/drivers/#troubleshooting",
		ShowIssueLink: true,
	},
	"DRIVER_NOT_FOUND": {
		Regexp: re(`registry: driver not found`),
		Advice: "Your minikube config refers to an unsupported driver. Erase ~/.minikube, and try again.",
		Issues: []int{5295},
	},
	"DRIVER_MISSING_ADDRESS": {
		Regexp:        re(`new host: dial tcp: missing address`),
		Advice:        "The machine-driver specified is failing to start. Try running 'docker-machine-driver-<type> version'",
		Issues:        []int{6023, 4679},
		ShowIssueLink: true,
	},
	"PRECREATE_EXIT_1": {
		Regexp:        re(`precreate: exit status 1`),
		Advice:        "The hypervisor does not appear to be configured properly. Run 'minikube start --alsologtostderr -v=1' and inspect the error code",
		Issues:        []int{6098},
		ShowIssueLink: true,
	},
	"FILE_IN_USE": {
		Regexp: re(`The process cannot access the file because it is being used by another process`),
		Advice: "Another program is using a file required by minikube. If you are using Hyper-V, try stopping the minikube VM from within the Hyper-V manager",
		URL:    "https://docs.docker.com/machine/drivers/hyper-v/",
		GOOS:   []string{"windows"},
		Issues: []int{7300},
	},
	"CREATE_TIMEOUT": {
		Regexp: re(`create host timed out in \d`),
		Advice: "Try 'minikube delete', and disable any conflicting VPN or firewall software",
		Issues: []int{7072},
	},
	"IMAGE_ARCH": {
		Regexp: re(`Error: incompatible image architecture`),
		Advice: "This driver does not yet work on your architecture. Maybe try --driver=none",
		GOOS:   []string{"linux"},
		Issues: []int{7071},
	},
	// Docker
	"DOCKER_WSL2_MOUNT": {
		Regexp: re(`cannot find cgroup mount destination: unknown`),
		Advice: "Run: 'sudo mkdir /sys/fs/cgroup/systemd && sudo mount -t cgroup -o none,name=systemd cgroup /sys/fs/cgroup/systemd'",
		URL:    "https://github.com/microsoft/WSL/issues/4189",
		Issues: []int{5392},
		GOOS:   []string{"linux"},
	},
	"DOCKER_READONLY": {
		Regexp: re(`mkdir /var/lib/docker/volumes.*: read-only file system`),
		Advice: "Restart Docker",
		Issues: []int{6825},
	},
	"DOCKER_CHROMEOS": {
		Regexp: re(`Container.*is not running.*chown docker:docker`),
		Advice: "minikube is not yet compatible with ChromeOS",
		Issues: []int{6411},
	},
	"DOCKER_PROVISION_STUCK_CONTAINER": {
		Regexp: re(`executing "" at <index (index .NetworkSettings.Ports "22/tcp") 0>`),
		Advice: "Restart Docker, Ensure docker is running and then run: 'minikube delete' and then 'minikube start' again",
		URL:    "https://github.com/kubernetes/minikube/issues/8163#issuecomment-652627436",
		Issues: []int{8163},
	},
	// Hyperkit
	"HYPERKIT_NO_IP": {
		Regexp: re(`IP address never found in dhcp leases file Temporary Error: Could not find an IP address for`),
		Advice: "Install the latest hyperkit binary, and run 'minikube delete'",
		URL:    "https://minikube.sigs.k8s.io/docs/reference/drivers/hyperkit/",
		Issues: []int{1926, 4206},
		GOOS:   []string{"darwin"},
	},
	"HYPERKIT_NOT_FOUND": {
		Regexp: re(`Driver "hyperkit" not found.`),
		Advice: "Please install the minikube hyperkit VM driver, or select an alternative --driver",
		URL:    "https://minikube.sigs.k8s.io/docs/reference/drivers/hyperkit/",
		GOOS:   []string{"darwin"},
	},
	"HYPERKIT_VMNET_FRAMEWORK": {
		Regexp: re(`error from vmnet.framework: -1`),
		Advice: "Hyperkit networking is broken. Upgrade to the latest hyperkit version and/or Docker for Desktop. Alternatively, you may choose an alternate --driver",
		Issues: []int{6028, 5594},
		GOOS:   []string{"darwin"},
	},
	"HYPERKIT_CRASHED": {
		Regexp: re(`hyperkit crashed!`),
		Advice: "Hyperkit is broken. Upgrade to the latest hyperkit version and/or Docker for Desktop. Alternatively, you may choose an alternate --driver",
		Issues: []int{6079, 5780},
		GOOS:   []string{"darwin"},
	},
	// Hyper-V
	"HYPERV_NO_VSWITCH": {
		Regexp: re(`no External vswitch found. A valid vswitch must be available for this command to run.`),
		Advice: "Configure an external network switch following the official documentation, then add `--hyperv-virtual-switch=<switch-name>` to `minikube start`",
		URL:    "https://docs.docker.com/machine/drivers/hyper-v/",
		GOOS:   []string{"windows"},
	},
	"HYPERV_VSWITCH_NOT_FOUND": {
		Regexp: re(`precreate: vswitch.*not found`),
		Advice: "Confirm that you have supplied the correct value to --hyperv-virtual-switch using the 'Get-VMSwitch' command",
		URL:    "https://docs.docker.com/machine/drivers/hyper-v/",
		GOOS:   []string{"windows"},
	},
	"HYPERV_POWERSHELL_NOT_FOUND": {
		Regexp: re(`Powershell was not found in the path`),
		Advice: "To start minikube with Hyper-V, Powershell must be in your PATH`",
		URL:    "https://docs.docker.com/machine/drivers/hyper-v/",
		GOOS:   []string{"windows"},
	},
	"HYPERV_AS_ADMIN": {
		Regexp: re(`Hyper-v commands have to be run as an Administrator`),
		Advice: "Right-click the PowerShell icon and select Run as Administrator to open PowerShell in elevated mode.",
		URL:    "https://rominirani.com/docker-machine-windows-10-hyper-v-troubleshooting-tips-367c1ea73c24",
		Issues: []int{4511},
		GOOS:   []string{"windows"},
	},
	"HYPERV_NEEDS_ESC": {
		Regexp: re(`The requested operation requires elevation.`),
		Advice: "Right-click the PowerShell icon and select Run as Administrator to open PowerShell in elevated mode.",
		Issues: []int{7347},
		GOOS:   []string{"windows"},
	},
	"HYPERV_FILE_DELETE_FAILURE": {
		Regexp: re(`Unable to remove machine directory`),
		Advice: "You may need to stop the Hyper-V Manager and run `minikube delete` again.",
		Issues: []int{6804},
		GOOS:   []string{"windows"},
	},
	// KVM
	"KVM2_NOT_FOUND": {
		Regexp: re(`Driver "kvm2" not found. Do you have the plugin binary .* accessible in your PATH`),
		Advice: "Please install the minikube kvm2 VM driver, or select an alternative --driver",
		URL:    "https://minikube.sigs.k8s.io/docs/reference/drivers/kvm2/",
		GOOS:   []string{"linux"},
	},
	"KVM2_NO_DOMAIN": {
		Regexp: re(`no domain with matching name`),
		Advice: "The VM that minikube is configured for no longer exists. Run 'minikube delete'",
		Issues: []int{3636},
		GOOS:   []string{"linux"},
	},
	"KVM_CREATE_CONFLICT": {
		Regexp: re(`KVM_CREATE_VM.* failed:.* Device or resource busy`),
		Advice: "Another hypervisor, such as VirtualBox, is conflicting with KVM. Please stop the other hypervisor, or use --driver to switch to it.",
		Issues: []int{4913},
		GOOS:   []string{"linux"},
	},
	"KVM2_RESTART_NO_IP": {
		Regexp: re(`Error starting stopped host: Machine didn't return an IP after \d+ seconds`),
		Advice: "The KVM driver is unable to resurrect this old VM. Please run `minikube delete` to delete it and try again.",
		Issues: []int{3901, 3434},
	},
	"KVM2_START_NO_IP": {
		Regexp: re(`Error in driver during machine creation: Machine didn't return an IP after \d+ seconds`),
		Advice: "Check your firewall rules for interference, and run 'virt-host-validate' to check for KVM configuration issues. If you are running minikube within a VM, consider using --driver=none",
		URL:    "https://minikube.sigs.k8s.io/docs/reference/drivers/kvm2/",
		Issues: []int{4249, 3566},
		GOOS:   []string{"linux"},
	},
	"KVM2_NETWORK_DEFINE_XML": {
		Regexp: re(`not supported by the connection driver: virNetworkDefineXML`),
		Advice: "Rebuild libvirt with virt-network support",
		URL:    "https://forums.gentoo.org/viewtopic-t-981692-start-0.html",
		Issues: []int{4195},
		GOOS:   []string{"linux"},
	},
	"KVM2_FAILED_MSR": {
		Regexp: re(`qemu unexpectedly closed the monitor.*failed to set MSR`),
		Advice: "Upgrade to QEMU v3.1.0+, run 'virt-host-validate', or ensure that you are not running in a nested VM environment.",
		Issues: []int{4277},
		GOOS:   []string{"linux"},
	},
	"KVM_UNAVAILABLE": {
		Regexp: re(`invalid argument: could not find capabilities for domaintype=kvm`),
		Advice: "Your host does not support KVM virtualization. Ensure that qemu-kvm is installed, and run 'virt-host-validate' to debug the problem",
		URL:    "http://mikko.repolainen.fi/documents/virtualization-with-kvm",
		Issues: []int{2991},
		GOOS:   []string{"linux"},
	},
	"KVM_CONNECTION_ERROR": {
		Regexp: re(`error connecting to libvirt socket`),
		Advice: "Have you set up libvirt correctly?",
		URL:    "https://minikube.sigs.k8s.io/docs/reference/drivers/kvm2/",
		GOOS:   []string{"linux"},
	},
	"KVM_ISO_PERMISSION": {
		Regexp: re(`boot2docker.iso.*Permission denied`),
		Advice: "Ensure that the user listed in /etc/libvirt/qemu.conf has access to your home directory",
		GOOS:   []string{"linux"},
		Issues: []int{5950},
	},
	"KVM_OOM": {
		Regexp: re(`cannot set up guest memory.*Cannot allocate memory`),
		Advice: "Choose a smaller value for --memory, such as 2000",
		GOOS:   []string{"linux"},
		Issues: []int{6366},
	},
	// None
	"NONE_APISERVER_MISSING": {
		Regexp: re(`apiserver process never appeared`),
		Advice: "Check that SELinux is disabled, and that the provided apiserver flags are valid",
		Issues: []int{6014, 4536},
		GOOS:   []string{"linux"},
	},
	"NONE_DOCKER_EXIT_1": {
		Regexp: re(`sudo systemctl start docker: exit status 1`),
		Advice: "Either systemctl is not installed, or Docker is broken. Run 'sudo systemctl start docker' and 'journalctl -u docker'",
		URL:    "https://minikube.sigs.k8s.io/docs/reference/drivers/none",
		Issues: []int{4498},
		GOOS:   []string{"linux"},
	},
	"NONE_DOCKER_EXIT_5": {
		Regexp: re(`sudo systemctl start docker: exit status 5`),
		Advice: "Ensure that Docker is installed and healthy: Run 'sudo systemctl start docker' and 'journalctl -u docker'. Alternatively, select another value for --driver",
		URL:    "https://minikube.sigs.k8s.io/docs/reference/drivers/none",
		Issues: []int{5532},
		GOOS:   []string{"linux"},
	},
	"NONE_CRIO_EXIT_5": {
		Regexp: re(`sudo systemctl restart crio: exit status 5`),
		Advice: "Ensure that CRI-O is installed and healthy: Run 'sudo systemctl start crio' and 'journalctl -u crio'. Alternatively, use --container-runtime=docker",
		URL:    "https://minikube.sigs.k8s.io/docs/reference/drivers/none",
		Issues: []int{5532},
		GOOS:   []string{"linux"},
	},
	"NONE_PORT_IN_USE": {
		Regexp: re(`ERROR Port-.*is in use`),
		Advice: "kubeadm detected a TCP port conflict with another process: probably another local Kubernetes installation. Run lsof -p<port> to find the process and kill it",
		Issues: []int{5484},
		GOOS:   []string{"linux"},
	},
	"NONE_KUBELET": {
		Regexp: re(`The kubelet is not running`),
		Advice: "Check output of 'journalctl -xeu kubelet', try passing --extra-config=kubelet.cgroup-driver=systemd to minikube start",
		Issues: []int{4172},
		GOOS:   []string{"linux"},
	},
	"NONE_DEFAULT_ROUTE": {
		Regexp: re(`(No|from) default routes`),
		Advice: "Configure a default route on this Linux host, or use another --driver that does not require it",
		Issues: []int{6083, 5636},
		GOOS:   []string{"linux"},
	},
	// VirtualBox
	"VBOX_BLOCKED": {
		Regexp: re(`NS_ERROR_FAILURE.*0x80004005`),
		Advice: "Reinstall VirtualBox and verify that it is not blocked: System Preferences -> Security & Privacy -> General -> Some system software was blocked from loading",
		Issues: []int{4107},
		GOOS:   []string{"darwin"},
	},
	"VBOX_DRV_NOT_LOADED": {
		Regexp: re(`vboxdrv kernel module is not loaded`),
		Advice: "Reinstall VirtualBox and reboot. Alternatively, try the kvm2 driver: https://minikube.sigs.k8s.io/docs/reference/drivers/kvm2/",

		Issues: []int{4043, 4711},
	},
	"VBOX_DEVICE_MISSING": {
		Regexp: re(`vboxdrv does not exist`),
		Advice: "Reinstall VirtualBox and reboot. Alternatively, try the kvm2 driver: https://minikube.sigs.k8s.io/docs/reference/drivers/kvm2/",
		Issues: []int{3974},
	},
	"VBOX_HARDENING": {
		Regexp: re(`terminated unexpectedly.*VBoxHardening`),
		Advice: "VirtualBox is broken. Disable real-time anti-virus software, reboot, and reinstall VirtualBox if the problem continues.",
		Issues: []int{3859, 3910},
		URL:    "https://forums.virtualbox.org/viewtopic.php?f=25&t=82106",
		GOOS:   []string{"windows"},
	},
	"VBOX_NS_ERRROR": {
		Regexp: re(`terminated unexpectedly.*NS_ERROR_FAILURE.*0x80004005`),
		Advice: "VirtualBox is broken. Reinstall VirtualBox, reboot, and run 'minikube delete'.",
		Issues: []int{5227},
		GOOS:   []string{"linux"},
	},
	"VBOX_HOST_ADAPTER": {
		Regexp: re(`The host-only adapter we just created is not visible`),
		Advice: "Reboot to complete VirtualBox installation, verify that VirtualBox is not blocked by your system, and/or use another hypervisor",
		Issues: []int{3614, 4222, 5817},
		URL:    "https://stackoverflow.com/questions/52277019/how-to-fix-vm-issue-with-minikube-start",
	},
	"VBOX_IP_CONFLICT": {
		Regexp: re(`VirtualBox is configured with multiple host-only adapters with the same IP`),
		Advice: "Use VirtualBox to remove the conflicting VM and/or network interfaces",
		URL:    "https://stackoverflow.com/questions/55573426/virtualbox-is-configured-with-multiple-host-only-adapters-with-the-same-ip-whe",
		Issues: []int{3584},
	},
	"VBOX_HYPERV_64_BOOT": {
		Regexp: re(`VirtualBox won't boot a 64bits VM when Hyper-V is activated`),
		Advice: "VirtualBox and Hyper-V are having a conflict. Use '--driver=hyperv' or disable Hyper-V using: 'bcdedit /set hypervisorlaunchtype off'",
		Issues: []int{4051, 4783},
	},
	"VBOX_HYPERV_NEM_VM": {
		Regexp: re(`vrc=VERR_NEM_VM_CREATE_FAILED`),
		Advice: "VirtualBox and Hyper-V are having a conflict. Use '--driver=hyperv' or disable Hyper-V using: 'bcdedit /set hypervisorlaunchtype off'",
		Issues: []int{4587},
	},
	"VBOX_NOT_FOUND": {
		Regexp: re(`VBoxManage not found. Make sure VirtualBox is installed and VBoxManage is in the path`),
		Advice: "Install VirtualBox and ensure it is in the path, or select an alternative value for --driver",
		URL:    "https://minikube.sigs.k8s.io/docs/start/",
		Issues: []int{3784},
	},
	"VBOX_NO_VM": {
		Regexp: re(`Could not find a registered machine named`),
		Advice: "The VM that minikube is configured for no longer exists. Run 'minikube delete'",
		Issues: []int{4694},
	},
	"VBOX_VTX_DISABLED": {
		Regexp: re(`This computer doesn't have VT-X/AMD-v enabled`),
		Advice: "Virtualization support is disabled on your computer. If you are running minikube within a VM, try '--driver=docker'. Otherwise, consult your systems BIOS manual for how to enable virtualization.",
		Issues: []int{3900, 4730},
	},
	"VERR_VERR_VMX_DISABLED": {
		Regexp: re(`VT-x is disabled.*VERR_VMX_MSR_ALL_VMX_DISABLED`),
		Advice: "Virtualization support is disabled on your computer. If you are running minikube within a VM, try '--driver=docker'. Otherwise, consult your systems BIOS manual for how to enable virtualization.",
		Issues: []int{5282, 5456},
	},
	"VBOX_VERR_VMX_NO_VMX": {
		Regexp: re(`VT-x is not available.*VERR_VMX_NO_VMX`),
		Advice: "Your host does not support virtualization. If you are running minikube within a VM, try '--driver=docker'. Otherwise, enable virtualization in your BIOS",
		Issues: []int{1994, 5326},
	},
	"VERR_SVM_DISABLED": {
		Regexp: re(`VERR_SVM_DISABLED`),
		Advice: "Your host does not support virtualization. If you are running minikube within a VM, try '--driver=docker'. Otherwise, enable virtualization in your BIOS",
		Issues: []int{7074},
	},
	"VBOX_HOST_NETWORK": {
		Regexp: re(`Error setting up host only network on machine start.*Unspecified error`),
		Advice: "VirtualBox cannot create a network, probably because it conflicts with an existing network that minikube no longer knows about. Try running 'minikube delete'",
		Issues: []int{5260},
	},
	"VBOX_INTERFACE_NOT_FOUND": {
		Regexp: re(`ERR_INTNET_FLT_IF_NOT_FOUND`),
		Advice: "VirtualBox is unable to find its network interface. Try upgrading to the latest release and rebooting.",
		Issues: []int{6036},
	},
}

// proxyDoc is the URL to proxy documentation
const (
	proxyDoc = "https://minikube.sigs.k8s.io/docs/handbook/vpn_and_proxy/"
	vpnDoc   = "https://minikube.sigs.k8s.io/docs/handbook/vpn_and_proxy/"
)

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
	"DOWNLOAD_BLOCKED": {
		Regexp: re(`iso: failed to download|download.*host has failed to respond`),
		Advice: "A firewall is likely blocking minikube from reaching the internet. You may need to configure minikube to use a proxy.",
		URL:    proxyDoc,
		Issues: []int{3922, 6109, 6123},
	},
	"PULL_TIMEOUT_EXCEEDED": {
		Regexp: re(`ImagePull.*Timeout exceeded while awaiting headers`),
		Advice: "A firewall is blocking Docker the minikube VM from reaching the image repository. You may need to select --image-repository, or use a proxy.",
		URL:    proxyDoc,
		Issues: []int{3898, 6070},
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
		Regexp: re(`host-only cidr conflicts with the network address of a host interface`),
		Advice: "Specify an alternate --host-only-cidr value, such as 172.16.0.1/24",
		Issues: []int{3594},
	},
	"HTTP_HTTPS_RESPONSE": {
		Regexp: re(`http: server gave HTTP response to HTTPS client`),
		Advice: "Ensure that your value for HTTPS_PROXY points to an HTTPS proxy rather than an HTTP proxy",
		Issues: []int{6107},
		URL:    proxyDoc,
	},
	"NOT_A_TLS_HANDSHAKE": {
		Regexp: re(`tls: first record does not look like a TLS handshake`),
		Advice: "Ensure that your value for HTTPS_PROXY points to an HTTPS proxy rather than an HTTP proxy",
		Issues: []int{7286},
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
	"APISERVER_MISSING": {
		Regexp: re(`apiserver process never appeared`),
		Advice: "Check that the provided apiserver flags are valid, and that SELinux is disabled",
		Issues: []int{4536, 6014},
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
	"OPEN_SERVICE_NOT_FOUND": {
		Regexp: re(`Error opening service.*not found`),
		Advice: "Use 'kubect get po -A' to find the correct and namespace name",
		Issues: []int{5836},
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
	"CERT_NOT_SIGNED_BY_CA": {
		Regexp: re(`not signed by CA certificate ca: crypto/rsa: verification error`),
		Advice: "Try 'minikube delete' to force new SSL certificates to be installed",
		Issues: []int{6596},
	},
	"DOCKER_RESTART_FAILED": {
		Regexp: re(`systemctl -f restart docker`),
		Advice: "Remove the incompatible --docker-opt flag if one was provided",
		Issues: []int{7070},
	},
	"WAITING_FOR_SSH": {
		Regexp: re(`waiting for SSH to be available`),
		Advice: "Try 'minikube delete', and disable any conflicting VPN or firewall software",
		Issues: []int{4617},
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
		URL:    "https://minikube.sigs.k8s.io/docs/reference/drivers/none/",
		Issues: []int{2704},
	},
	"KUBECONFIG_WRITE_FAIL": {
		Regexp: re(`Failed to setup kubeconfig: writing kubeconfig`),
		Advice: "Unset the KUBECONFIG environment variable, or verify that it does not point to an empty or otherwise invalid path",
		Issues: []int{5268, 4100, 5207},
	},
	"KUBECONFIG_DENIED": {
		Regexp: re(`.kube/config: permission denied`),
		Advice: "Run: 'chmod 600 $HOME/.kube/config'",
		GOOS:   []string{"darwin", "linux"},
		Issues: []int{5714},
	},
	"JUJU_LOCK_DENIED": {
		Regexp: re(`unable to open /tmp/juju.*: permission denied`),
		Advice: "Run 'sudo sysctl fs.protected_regular=0', or try a driver which does not require root, such as '--driver=docker'",
		GOOS:   []string{"linux"},
		Issues: []int{6391},
	},
}

// stateProblems are issues relating to local state
var stateProblems = map[string]match{
	"MACHINE_DOES_NOT_EXIST": {
		Regexp: re(`machine does not exist`),
		Advice: "Run 'minikube delete' to delete the stale VM, or and ensure that minikube is running as the same user you are issuing this command with",
		Issues: []int{3864, 6087},
	},
	"MACHINE_NOT_FOUND": {
		Regexp: re(`Machine does not exist for api.Exists`),
		Advice: "Your minikube vm is not running, try minikube start.",
		Issues: []int{4889},
	},
	"IP_NOT_FOUND": {
		Regexp: re(`Error getting ssh host name for driver: IP not found`),
		Advice: "The minikube VM is offline. Please run 'minikube start' to start it again.",
		Issues: []int{3849, 3648},
	},
	"DASHBOARD_ROLE_REF": {
		Regexp: re(`dashboard.*cannot change roleRef`),
		Advice: "Run: 'kubectl delete clusterrolebinding kubernetes-dashboard'",
		Issues: []int{7256},
	},
}

// dockerProblems are issues relating to issues with the docker driver
var dockerProblems = map[string]match{
	"NO_SPACE_ON_DEVICE": {
		Regexp: re(`.*docker.*No space left on device.*`),
		Advice: `Try at least one of the following to free up space on the device:

	1. Run "docker system prune" to remove unused docker data
	2. Increase the amount of memory allocated to Docker for Desktop via 
		Docker icon > Preferences > Resources > Disk Image Size
	3. Run "minikube ssh -- docker system prune" if using the docker container runtime
`,
		Issues: []int{9024},
	},
}
