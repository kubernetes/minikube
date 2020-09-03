/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

/*
Copyright 2019 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY matchND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package reason

import (
	"regexp"

	"k8s.io/minikube/pkg/minikube/style"
)

// links used by multiple known issues
const (
	proxyDoc = "https://minikube.sigs.k8s.io/docs/handbook/vpn_and_proxy/"
	vpnDoc   = "https://minikube.sigs.k8s.io/docs/handbook/vpn_and_proxy/"
)

// re is a shortcut around regexp.MustCompile
func re(s string) *regexp.Regexp {
	return regexp.MustCompile(s)
}

// programIssues are issues with the minikube binary
var programIssues = []match{
	{
		Kind: Kind{
			ID:       "MK_KVERSION_USAGE",
			ExitCode: ExProgramUsage,
			Advice:   "Specify --kubernetes-version in v<major>.<minor.<build> form. example: 'v1.1.14'",
		},

		Regexp: re(`No Major.Minor.Patch elements found`),
	},
}

// resourceIssues are failures due to resource constraints
var resourceIssues = []match{
	{
		Kind: Kind{
			ID:       "RSRC_KVM_OOM",
			ExitCode: ExInsufficientMemory,
			Advice:   "Choose a smaller value for --memory, such as 2000",
			Issues:   []int{6366},
		},
		Regexp: re(`cannot set up guest memory.*Cannot allocate memory`),
		GOOS:   []string{"linux"},
	},
	{
		Kind: Kind{
			ID:       "RSRC_SSH_OOM",
			ExitCode: ExInsufficientMemory,
			Advice:   "Disable dynamic memory in your VM manager, or pass in a larger --memory value",
			Issues:   []int{1766},
		},
		Regexp: re(`Process exited with status 137 from signal matchLL`),
	},
	{
		Kind: Kind{
			ID:       "RSRC_SCP_OOM",
			ExitCode: ExInsufficientMemory,
			Advice:   "Disable dynamic memory in your VM manager, or pass in a larger --memory value",
			Issues:   []int{1766},
		},
		Regexp: re(`An existing connection was forcibly closed by the remote host`),
	},
	{
		// Fallback to deliver a good error message even if internal checks are not run
		Kind: Kind{
			ID:       "RSRC_INSUFFICIENT_CORES",
			ExitCode: ExInsufficientCores,
			Advice:   "Kubernetes requires at least 2 CPU's to start",
			Issues:   []int{7905},
			URL:      "https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/",
		},
		Regexp: re(`ERROR.*the number of available CPUs 1 is less than the required 2`),
	},
}

// hostIssues are related to the host operating system or BIOS
var hostIssues = []match{
	{
		Kind: Kind{
			ID:       "HOST_VIRT_UNAVAILABLE",
			ExitCode: ExHostConfig,
			Advice:   "Virtualization support is disabled on your computer. If you are running minikube within a VM, try '--driver=docker'. Otherwise, consult your systems BIOS manual for how to enable virtualization.",
			Issues:   []int{3900, 4730},
		},
		Regexp: re(`This computer doesn't have VT-X/AMD-v enabled`),
	},
	{
		Kind: Kind{
			ID:       "HOST_VTX_DISABLED",
			ExitCode: ExHostConfig,
			Advice:   "Virtualization support is disabled on your computer. If you are running minikube within a VM, try '--driver=docker'. Otherwise, consult your systems BIOS manual for how to enable virtualization.",
			Issues:   []int{5282, 5456},
		},
		Regexp: re(`VT-x is disabled.*VERR_VMX_MSR_ALL_VMX_DISABLED`),
	},
	{
		Kind: Kind{
			ID:       "HOST_VTX_UNAVAILABLE",
			ExitCode: ExHostConfig,
			Advice:   "Your host does not support virtualization. If you are running minikube within a VM, try '--driver=docker'. Otherwise, enable virtualization in your BIOS",
			Issues:   []int{1994, 5326},
		},
		Regexp: re(`VT-x is not available.*VERR_VMX_NO_VMX`),
	},
	{
		Kind: Kind{
			ID:       "HOST_SVM_DISABLED",
			ExitCode: ExHostConfig,
			Advice:   "Your host does not support virtualization. If you are running minikube within a VM, try '--driver=docker'. Otherwise, enable virtualization in your BIOS",
			Issues:   []int{7074},
		},
		Regexp: re(`VERR_SVM_DISABLED`),
	},
	{
		Kind: Kind{
			ID:       "HOST_NON_C_DRIVE",
			ExitCode: ExHostUsage,
			Advice:   "Run minikube from the C: drive.",
			Issues:   []int{1574},
		},
		Regexp: re(`.iso: The system cannot find the path specified.`),
	},
	{
		Kind: Kind{
			ID:       "HOST_KUBECONFIG_WRITE",
			ExitCode: ExHostPermission,
			Advice:   "Unset the KUBECONFIG environment variable, or verify that it does not point to an empty or otherwise invalid path",
			Issues:   []int{5268, 4100, 5207},
		},
		Regexp: re(`Failed to setup kubeconfig: writing kubeconfig`),
	},
	{
		Kind: Kind{
			ID:       "HOST_KUBECONFIG_PERMISSION",
			ExitCode: ExHostPermission,
			Advice:   "Run: 'sudo chown $USER $HOME/.kube/config && chmod 600 $HOME/.kube/config'",
			Issues:   []int{5714},
			Style:    style.NotAllowed,
		},
		Regexp: re(`.kube/config: permission denied`),
		GOOS:   []string{"darwin", "linux"},
	},
	{
		Kind: Kind{
			ID:       "HOST_JUJU_LOCK_PERMISSION",
			ExitCode: ExHostPermission,
			Advice:   "Run 'sudo sysctl fs.protected_regular=0', or try a driver which does not require root, such as '--driver=docker'",
			Issues:   []int{6391},
		},
		Regexp: re(`unable to open /tmp/juju.*: permission denied`),
		GOOS:   []string{"linux"},
	},
	{
		Kind: Kind{
			ID:       "HOST_DOCKER_CHROMEOS",
			ExitCode: ExHostUnsupported,
			Advice:   "ChromeOS is missing the kernel support necessary for running Kubernetes",
			Issues:   []int{6411},
		},
		Regexp: re(`Container.*is not running.*chown docker:docker`),
	},
	{
		Kind: Kind{
			ID:       "HOST_PIDS_CGROUP",
			ExitCode: ExHostUnsupported,
			Advice:   "Ensure that the required 'pids' cgroup is enabled on your host: grep pids /proc/cgroups",
			Issues:   []int{6411},
		},
		Regexp: re(`failed to find subsystem mount for required subsystem: pids`),
		GOOS:   []string{"linux"},
	},
	{
		Kind: Kind{
			ID:       "HOST_HOME_PERMISSION",
			ExitCode: ExGuestPermission,
			Advice:   "Your user lacks permissions to the minikube profile directory. Run: 'sudo chown -R $USER $HOME/.minikube; chmod -R u+wrx $HOME/.minikube' to fix",
			Issues:   []int{9165},
		},
		Regexp: re(`/.minikube/.*: permission denied`),
	},
}

// providerIssues are failures relating to a driver provider
var providerIssues = []match{
	// General
	{
		Kind: Kind{
			ID:       "PR_PRECREATE_EXIT_1",
			ExitCode: ExProviderError,
			Advice:   "The hypervisor does not appear to be configured properly. Run 'minikube start --alsologtostderr -v=1' and inspect the error code",
			Issues:   []int{6098},
		},
		Regexp: re(`precreate: exit status 1`),
	},

	// Docker environment
	{
		Kind: Kind{
			ID:       "PR_DOCKER_CGROUP_MOUNT",
			ExitCode: ExProviderError,
			Advice:   "Run: 'sudo mkdir /sys/fs/cgroup/systemd && sudo mount -t cgroup -o none,name=systemd cgroup /sys/fs/cgroup/systemd'",
			URL:      "https://github.com/microsoft/WSL/issues/4189",
			Issues:   []int{5392},
		},
		Regexp: re(`cannot find cgroup mount destination: unknown`),
		GOOS:   []string{"linux"},
	},
	{
		Kind: Kind{
			ID:       "PR_DOCKER_READONLY_VOL",
			ExitCode: ExProviderError,
			Advice:   "Restart Docker",
			Issues:   []int{6825},
		},
		Regexp: re(`mkdir /var/lib/docker/volumes.*: read-only file system`),
	},
	{
		Kind: Kind{
			ID:       "PR_DOCKER_NO_SSH",
			ExitCode: ExProviderTimeout,
			Advice:   "Restart Docker, Ensure docker is running and then run: 'minikube delete' and then 'minikube start' again",
			URL:      "https://github.com/kubernetes/minikube/issues/8163#issuecomment-652627436",
			Issues:   []int{8163},
		},
		Regexp: re(`executing "" at <index (index .NetworkSettings.Ports "22/tcp") 0>`),
	},
	{
		Kind: Kind{
			ID:       "PR_DOCKER_MOUNTS_EOF",
			ExitCode: ExProviderError,
			Advice:   "Reset Docker to factory defaults",
			Issues:   []int{8832},
			URL:      "https://docs.docker.com/docker-for-mac/#reset",
		},
		GOOS:   []string{"darwin"},
		Regexp: re(`docker:.*Mounts denied: EOF`),
	},
	{
		Kind: Kind{
			ID:       "PR_DOCKER_MOUNTS_EOF",
			ExitCode: ExProviderError,
			Advice:   "Reset Docker to factory defaults",
			Issues:   []int{8832},
			URL:      "https://docs.docker.com/docker-for-windows/#reset",
		},
		GOOS:   []string{"windows"},
		Regexp: re(`docker:.*Mounts denied: EOF`),
	},

	// Hyperkit hypervisor
	{
		Kind: Kind{
			ID:       "PR_HYPERKIT_NO_IP",
			ExitCode: ExProviderError,
			Advice:   "Install the latest hyperkit binary, and run 'minikube delete'",
			URL:      "https://minikube.sigs.k8s.io/docs/reference/drivers/hyperkit/",
			Issues:   []int{1926, 4206},
		},
		Regexp: re(`IP address never found in dhcp leases file Temporary Error: Could not find an IP address for`),
		GOOS:   []string{"darwin"},
	},
	{
		Kind: Kind{
			ID:       "PR_HYPERKIT_NOT_FOUND",
			ExitCode: ExProviderNotFound,
			Advice:   "Please install the minikube hyperkit VM driver, or select an alternative --driver",
			URL:      "https://minikube.sigs.k8s.io/docs/reference/drivers/hyperkit/",
		},
		Regexp: re(`Driver "hyperkit" not found.`),
		GOOS:   []string{"darwin"},
	},
	{
		Kind: Kind{
			ID:       "PR_HYPERKIT_VMNET_FRAMEWORK",
			ExitCode: ExProviderError,
			Advice:   "Hyperkit networking is broken. Upgrade to the latest hyperkit version and/or Docker for Desktop. Alternatively, you may choose an alternate --driver",
			Issues:   []int{6028, 5594},
		},
		Regexp: re(`error from vmnet.framework: -1`),
		GOOS:   []string{"darwin"},
	},
	{
		Kind: Kind{
			ID:       "PR_HYPERKIT_CRASHED",
			ExitCode: ExProviderError,
			Advice:   "Hyperkit is broken. Upgrade to the latest hyperkit version and/or Docker for Desktop. Alternatively, you may choose an alternate --driver",
			Issues:   []int{6079, 5780},
		},
		Regexp: re(`hyperkit crashed!`),
		GOOS:   []string{"darwin"},
	},

	// Hyper-V hypervisor
	{
		Kind: Kind{
			ID:       "PR_HYPERV_AS_ADMIN",
			ExitCode: ExProviderPermission,
			Advice:   "Right-click the PowerShell icon and select Run as Administrator to open PowerShell in elevated mode.",
			URL:      "https://rominirani.com/docker-machine-windows-10-hyper-v-troubleshooting-tips-367c1ea73c24",
			Issues:   []int{4511},
		},
		Regexp: re(`Hyper-v commands have to be run as an Administrator`),
		GOOS:   []string{"windows"},
	},
	{
		Kind: Kind{
			ID:       "PR_HYPERV_NEEDS_ESC",
			ExitCode: ExProviderPermission,
			Advice:   "Right-click the PowerShell icon and select Run as Administrator to open PowerShell in elevated mode.",
			Issues:   []int{7347},
		},
		Regexp: re(`The requested operation requires elevation.`),
		GOOS:   []string{"windows"},
	},
	{
		Kind: Kind{
			ID:       "PR_POWERSHELL_CONSTRAINED",
			ExitCode: ExProviderPermission,
			Advice:   "PowerShell is running in constrained mode, which is incompatible with Hyper-V scripting.",
			Issues:   []int{7347},
			URL:      "https://devblogs.microsoft.com/powershell/powershell-constrained-language-mode/",
		},
		Regexp: re(`MethodInvocationNotSupportedInConstrainedLanguage`),
		GOOS:   []string{"windows"},
	},
	{
		Kind: Kind{
			ID:       "PR_HYPERV_MODULE_NOT_INSTALLED",
			ExitCode: ExProviderNotFound,
			Advice:   "Run: 'Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V-Tools-All'",
			Issues:   []int{7347},
			URL:      "https://www.altaro.com/hyper-v/install-hyper-v-powershell-module/",
		},
		Regexp: re(`Hyper-V PowerShell Module is not available`),
		GOOS:   []string{"windows"},
	},

	// KVM hypervisor
	{
		Kind: Kind{
			ID:       "PR_KVM_CAPABILITIES",
			ExitCode: ExProviderUnavailable,
			Advice:   "Your host does not support KVM virtualization. Ensure that qemu-kvm is installed, and run 'virt-host-validate' to debug the problem",
			URL:      "http://mikko.repolainen.fi/documents/virtualization-with-kvm",
			Issues:   []int{2991},
		},
		Regexp: re(`invalid argument: could not find capabilities for domaintype=kvm`),
		GOOS:   []string{"linux"},
	},
	{
		Kind: Kind{
			ID:       "PR_KVM_SOCKET",
			ExitCode: ExProviderUnavailable,
			Advice:   "Check that libvirt is setup properly",
			URL:      "https://minikube.sigs.k8s.io/docs/reference/drivers/kvm2/",
		},
		Regexp: re(`error connecting to libvirt socket`),
		GOOS:   []string{"linux"},
	},
	{
		Kind: Kind{
			ID:       "PR_KVM_ISO_PERMISSION",
			ExitCode: ExProviderPermission,
			Advice:   "Ensure that the user listed in /etc/libvirt/qemu.conf has access to your home directory",
			Issues:   []int{5950},
		},
		Regexp: re(`boot2docker.iso.*Permission denied`),
		GOOS:   []string{"linux"},
	},
	{
		Kind: Kind{
			ID:       "PR_KVM_NET_XML",
			ExitCode: ExProviderConfig,
			Advice:   "Rebuild libvirt with virt-network support",
			URL:      "https://forums.gentoo.org/viewtopic-t-981692-start-0.html",
			Issues:   []int{4195},
		},
		Regexp: re(`not supported by the connection driver: virNetworkDefineXML`),
		GOOS:   []string{"linux"},
	},
	{
		Kind: Kind{
			ID:       "PR_KVM_MSR",
			ExitCode: ExProviderError,
			Advice:   "Upgrade to QEMU v3.1.0+, run 'virt-host-validate', or ensure that you are not running in a nested VM environment.",
			Issues:   []int{4277},
		},
		Regexp: re(`qemu unexpectedly closed the monitor.*failed to set MSR`),
		GOOS:   []string{"linux"},
	},
	{
		Kind: Kind{
			ID:       "PR_KVM_CREATE_BUSY",
			ExitCode: ExDriverConflict,
			Advice:   "Another hypervisor, such as VirtualBox, is conflicting with KVM. Please stop the other hypervisor, or use --driver to switch to it.",
			Issues:   []int{4913},
		},
		Regexp: re(`KVM_CREATE_VM.* failed:.* Device or resource busy`),
		GOOS:   []string{"linux"},
	},

	// VirtualBox provider
	{
		Kind: Kind{
			ID:       "PR_VBOX_BLOCKED",
			ExitCode: ExProviderPermission,
			Advice:   "Reinstall VirtualBox and verify that it is not blocked: System Preferences -> Security & Privacy -> General -> Some system software was blocked from loading",
			Issues:   []int{4107},
		},
		Regexp: re(`NS_ERROR.*0x80004005`),
		GOOS:   []string{"darwin"},
	},
	{
		Kind: Kind{
			ID:       "PR_VBOX_MODULE",
			ExitCode: ExProviderNotRunning,
			Advice:   "Reinstall VirtualBox and reboot. Alternatively, try the kvm2 driver: https://minikube.sigs.k8s.io/docs/reference/drivers/kvm2/",
			Issues:   []int{4043, 4711},
		},
		Regexp: re(`vboxdrv kernel module is not loaded`),
	},
	{
		Kind: Kind{
			ID:       "PR_VBOX_DEVICE_MISSING",
			ExitCode: ExProviderNotRunning,
			Advice:   "Reinstall VirtualBox and reboot. Alternatively, try the kvm2 driver: https://minikube.sigs.k8s.io/docs/reference/drivers/kvm2/",
			Issues:   []int{3974},
		},
		Regexp: re(`vboxdrv does not exist`),
	},
	{
		Kind: Kind{
			ID:       "PR_VBOX_HARDENING",
			ExitCode: ExProviderConflict,
			Advice:   "VirtualBox is broken. Disable real-time anti-virus software, reboot, and reinstall VirtualBox if the problem continues.",
			Issues:   []int{3859, 3910},
			URL:      "https://forums.virtualbox.org/viewtopic.php?f=25&t=82106",
		},
		Regexp: re(`terminated unexpectedly.*VBoxHardening`),
		GOOS:   []string{"windows"},
	},
	{
		Kind: Kind{
			ID:       "PR_VBOX_80004005",
			ExitCode: ExProviderError,
			Advice:   "VirtualBox is broken. Reinstall VirtualBox, reboot, and run 'minikube delete'.",
			Issues:   []int{5227},
		},
		Regexp: re(`terminated unexpectedly.*NS_ERROR.*0x80004005`),
		GOOS:   []string{"linux"},
	},
	{
		Kind: Kind{
			ID:       "PR_VBOX_HYPERV_64_BOOT",
			ExitCode: ExProviderConflict,
			Advice:   "VirtualBox and Hyper-V are having a conflict. Use '--driver=hyperv' or disable Hyper-V using: 'bcdedit /set hypervisorlaunchtype off'",
			Issues:   []int{4051, 4783},
		},
		Regexp: re(`VirtualBox won't boot a 64bits VM when Hyper-V is activated`),
	},
	{
		Kind: Kind{
			ID:       "PR_VBOX_HYPERV_CONFLICT",
			ExitCode: ExProviderConflict,
			Advice:   "VirtualBox and Hyper-V are having a conflict. Use '--driver=hyperv' or disable Hyper-V using: 'bcdedit /set hypervisorlaunchtype off'",
			Issues:   []int{4587},
		},
		Regexp: re(`vrc=VERR_NEM_VM_CREATE`),
	},
	{
		Kind: Kind{
			ID:       "PR_VBOXMANAGE_NOT_FOUND",
			ExitCode: ExProviderNotFound,
			Advice:   "Install VirtualBox and ensure it is in the path, or select an alternative value for --driver",
			URL:      "https://minikube.sigs.k8s.io/docs/start/",
			Issues:   []int{3784},
		},
		Regexp: re(`VBoxManage not found. Make sure VirtualBox is installed and VBoxManage is in the path`),
	},
}

// driverIssues are specific to a libmachine driver
var driverIssues = []match{
	// Generic VM driver
	{
		Kind: Kind{
			ID:           "DRV_CORRUPT",
			ExitCode:     ExDriverError,
			Advice:       "The VM driver exited with an error, and may be corrupt. Run 'minikube start' with --alsologtostderr -v=8 to see the error",
			URL:          "https://minikube.sigs.k8s.io/docs/reference/drivers/",
			NewIssueLink: true,
		},
		Regexp: re(`Error attempting to get plugin server address for RPC`),
	},
	{
		Kind: Kind{
			ID:           "DRV_EXITED_1",
			ExitCode:     ExDriverError,
			Advice:       "The VM driver crashed. Run 'minikube start --alsologtostderr -v=8' to see the VM driver error message",
			URL:          "https://minikube.sigs.k8s.io/docs/reference/drivers/#troubleshooting",
			NewIssueLink: true,
		},
		Regexp: re(`Unable to start VM: start: exit status 1`),
	},
	{
		Kind: Kind{
			ID:       "DRV_REGISTRY_NOT_FOUND",
			ExitCode: ExDriverUnsupported,
			Advice:   "Your minikube config refers to an unsupported driver. Erase ~/.minikube, and try again.",
			Issues:   []int{5295},
		},
		Regexp: re(`registry: driver not found`),
	},
	{
		Kind: Kind{
			ID:           "DRV_MISSING_ADDRESS",
			ExitCode:     ExDriverError,
			Advice:       "The machine-driver specified is failing to start. Try running 'docker-machine-driver-<type> version'",
			Issues:       []int{6023, 4679},
			NewIssueLink: true,
		},
		Regexp: re(`new host: dial tcp: missing address`),
	},
	{
		Kind: Kind{
			ID:       "DRV_CREATE_TIMEOUT",
			ExitCode: ExDriverTimeout,
			Advice:   "Try 'minikube delete', and disable any conflicting VPN or firewall software",
			Issues:   []int{7072},
		},
		Regexp: re(`create host timed out in \d`),
	},
	{
		Kind: Kind{
			ID:       "DRV_IMAGE_ARCH_UNSUPPORTED",
			ExitCode: ExDriverUnsupported,
			Advice:   "This driver does not yet work on your architecture. Maybe try --driver=none",
			Issues:   []int{7071},
		},
		Regexp: re(`Error: incompatible image architecture`),
		GOOS:   []string{"linux"},
	},

	// Hyper-V
	{
		Kind: Kind{
			ID:       "DRV_HYPERV_NO_VSWITCH",
			ExitCode: ExDriverConfig,
			Advice:   "Configure an external network switch following the official documentation, then add `--hyperv-virtual-switch=<switch-name>` to `minikube start`",
			URL:      "https://docs.docker.com/machine/drivers/hyper-v/",
		},
		Regexp: re(`no External vswitch found. A valid vswitch must be available for this command to run.`),
		GOOS:   []string{"windows"},
	},
	{
		Kind: Kind{
			ID:       "DRV_HYPERV_VSWITCH_NOT_FOUND",
			ExitCode: ExDriverUsage,
			Advice:   "Confirm that you have supplied the correct value to --hyperv-virtual-switch using the 'Get-VMSwitch' command",
			URL:      "https://docs.docker.com/machine/drivers/hyper-v/",
		},
		Regexp: re(`precreate: vswitch.*not found`),
		GOOS:   []string{"windows"},
	},
	{
		Kind: Kind{
			ID:       "DRV_HYPERV_POWERSHELL_NOT_FOUND",
			ExitCode: ExDriverUnavailable,
			Advice:   "To start minikube with Hyper-V, Powershell must be in your PATH`",
			URL:      "https://docs.docker.com/machine/drivers/hyper-v/",
		},
		Regexp: re(`Powershell was not found in the path`),
		GOOS:   []string{"windows"},
	},

	{
		Kind: Kind{
			ID:       "DRV_HYPERV_FILE_DELETE",
			ExitCode: ExDriverConflict,
			Advice:   "You may need to stop the Hyper-V Manager and run `minikube delete` again.",
			Issues:   []int{6804},
		},
		Regexp: re(`Unable to remove machine directory`),
		GOOS:   []string{"windows"},
	},

	// KVM
	{
		Kind: Kind{
			ID:       "DRV_KVM2_NOT_FOUND",
			ExitCode: ExDriverNotFound,
			Advice:   "Please install the minikube kvm2 VM driver, or select an alternative --driver",
			URL:      "https://minikube.sigs.k8s.io/docs/reference/drivers/kvm2/",
		},
		Regexp: re(`Driver "kvm2" not found. Do you have the plugin binary .* accessible in your PATH`),
		GOOS:   []string{"linux"},
	},

	{
		Kind: Kind{
			ID:       "DRV_RESTART_NO_IP",
			ExitCode: ExDriverTimeout,
			Advice:   "The KVM driver is unable to resurrect this old VM. Please run `minikube delete` to delete it and try again.",
			Issues:   []int{3901, 3434},
		},
		Regexp: re(`Error starting stopped host: Machine didn't return an IP after \d+ seconds`),
	},
	{
		Kind: Kind{
			ID:       "DRV_NO_IP",
			ExitCode: ExDriverTimeout,
			Advice:   "Check your firewall rules for interference, and run 'virt-host-validate' to check for KVM configuration issues. If you are running minikube within a VM, consider using --driver=none",
			URL:      "https://minikube.sigs.k8s.io/docs/reference/drivers/kvm2/",
			Issues:   []int{4249, 3566},
		},
		Regexp: re(`Error in driver during machine creation: Machine didn't return an IP after \d+ seconds`),
		GOOS:   []string{"linux"},
	},
}

// localNetworkIssues are errors communicating to the guest
var localNetworkIssues = []match{
	{
		Kind: Kind{
			ID:       "IF_SSH_AUTH",
			ExitCode: ExLocalNetworkConfig,
			Advice:   "Your host is failing to route packets to the minikube VM. If you have VPN software, try turning it off or configuring it so that it does not re-route traffic to the VM IP. If not, check your VM environment routing options.",
			URL:      vpnDoc,
			Issues:   []int{3930},
		},
		Regexp: re(`ssh: handshake failed: ssh: unable to authenticate.*, no supported methods remain`),
	},
	{
		Kind: Kind{
			ID:       "IF_SSH_NO_RESPONSE",
			ExitCode: ExLocalNetworkConfig,
			Advice:   "Your host is failing to route packets to the minikube VM. If you have VPN software, try turning it off or configuring it so that it does not re-route traffic to the VM IP. If not, check your VM environment routing options.",
			URL:      vpnDoc,
			Issues:   []int{3388},
		},
		Regexp: re(`dial tcp .*:22: connectex: A connection attempt failed because the connected party did not properly respond`),
	},
	{
		Kind: Kind{
			ID:       "IF_HOST_CIDR_CONFLICT",
			ExitCode: ExLocalNetworkConflict,
			Advice:   "Specify an alternate --host-only-cidr value, such as 172.16.0.1/24",
			Issues:   []int{3594},
		},
		Regexp: re(`host-only cidr conflicts with the network address of a host interface`),
	},
	{
		Kind: Kind{
			ID:       "IF_VBOX_NOT_VISIBLE",
			ExitCode: ExLocalNetworkNotFound,
			Advice:   "Reboot to complete VirtualBox installation, verify that VirtualBox is not blocked by your system, and/or use another hypervisor",
			Issues:   []int{3614, 4222, 5817},
			URL:      "https://stackoverflow.com/questions/52277019/how-to-fix-vm-issue-with-minikube-start",
		},
		Regexp: re(`The host-only adapter we just created is not visible`),
	},
	{
		Kind: Kind{
			ID:       "IF_VBOX_SAME_IP",
			ExitCode: ExLocalNetworkConflict,
			Advice:   "Use VirtualBox to remove the conflicting VM and/or network interfaces",
			URL:      "https://stackoverflow.com/questions/55573426/virtualbox-is-configured-with-multiple-host-only-adapters-with-the-same-ip-whe",
			Issues:   []int{3584},
		},
		Regexp: re(`VirtualBox is configured with multiple host-only adapters with the same IP`),
	},
	{
		Kind: Kind{
			ID:       "IF_VBOX_NOT_FOUND",
			ExitCode: ExLocalNetworkNotFound,
			Advice:   "VirtualBox is unable to find its network interface. Try upgrading to the latest release and rebooting.",
			Issues:   []int{6036},
		},
		Regexp: re(`ERR_INTNET_FLT_IF_NOT_FOUND`),
	},
	{
		Kind: Kind{
			ID:       "IF_VBOX_UNSPECIFIED",
			ExitCode: ExLocalNetworkConflict,
			Advice:   "VirtualBox cannot create a network, probably because it conflicts with an existing network that minikube no longer knows about. Try running 'minikube delete'",
			Issues:   []int{5260},
		},
		Regexp: re(`Error setting up host only network on machine start.*Unspecified error`),
	},
	{
		Kind: Kind{
			ID:       "IF_SSH_TIMEOUT",
			ExitCode: ExLocalNetworkTimeout,
			Advice:   "Try 'minikube delete', and disable any conflicting VPN or firewall software",
			Issues:   []int{4617},
		},
		Regexp: re(`waiting for SSH to be available`),
	},
}

// internetIssues are internet related problems.
var internetIssues = []match{
	{
		Kind: Kind{
			ID:       "INET_GCR_UNAVAILABLE",
			ExitCode: ExInternetUnavailable,
			Advice:   "minikube is unable to access the Google Container Registry. You may need to configure it to use a HTTP proxy.",
			URL:      proxyDoc,
			Issues:   []int{3860},
		},
		Regexp: re(`gcr.io.*443: connect: invalid argument`),
	},
	{
		Kind: Kind{
			ID:       "INET_RESET_BY_PEER",
			ExitCode: ExInternetUnavailable,
			Advice:   "A firewall is likely blocking minikube from reaching the internet. You may need to configure minikube to use a proxy.",
			URL:      proxyDoc,
			Issues:   []int{3909},
		},
		Regexp: re(`Error downloading .*connection reset by peer`),
	},
	{
		Kind: Kind{
			ID:       "INET_DOWNLOAD_TIMEOUT",
			ExitCode: ExInternetTimeout,
			Advice:   "A firewall is likely blocking minikube from reaching the internet. You may need to configure minikube to use a proxy.",
			URL:      proxyDoc,
			Issues:   []int{3846},
		},
		Regexp: re(`Error downloading .*timeout`),
	},
	{
		Kind: Kind{
			ID:       "INET_TLS_OVERSIZED",
			ExitCode: ExInternetConflict,
			Advice:   "A firewall is interfering with minikube's ability to make outgoing HTTPS requests. You may need to change the value of the HTTPS_PROXY environment variable.",
			URL:      proxyDoc,
			Issues:   []int{3857, 3759, 4252},
		},
		Regexp: re(`tls: oversized record received with length`),
	},
	{
		Kind: Kind{
			ID:       "INET_DOWNLOAD_BLOCKED",
			ExitCode: ExInternetTimeout,
			Advice:   "A firewall is likely blocking minikube from reaching the internet. You may need to configure minikube to use a proxy.",
			URL:      proxyDoc,
			Issues:   []int{3922, 6109, 6123},
		},
		Regexp: re(`iso: failed to download|download.*host has failed to respond`),
	},
	{
		Kind: Kind{
			ID:       "INET_PULL_TIMEOUT",
			ExitCode: ExInternetTimeout,
			Advice:   "A firewall is blocking Docker the minikube VM from reaching the image repository. You may need to select --image-repository, or use a proxy.",
			URL:      proxyDoc,
			Issues:   []int{3898, 6070},
		},
		Regexp: re(`ImagePull.*Timeout exceeded while awaiting headers`),
	},
	{
		Kind: Kind{
			ID:       "INET_LOOKUP_HOST",
			ExitCode: ExInternetConfig,
			Advice:   "Verify that your HTTP_PROXY and HTTPS_PROXY environment variables are set correctly.",
			URL:      proxyDoc,
		},
		Regexp: re(`dial tcp: lookup.*: no such host`),
	},
	{
		Kind: Kind{
			ID:       "INET_PROXY_CONFUSION",
			ExitCode: ExInternetConfig,
			Advice:   "Ensure that your value for HTTPS_PROXY points to an HTTPS proxy rather than an HTTP proxy",
			Issues:   []int{6107},
			URL:      proxyDoc,
		},
		Regexp: re(`http: server gave HTTP response to HTTPS client`),
	},
	{
		Kind: Kind{
			ID:       "INET_NOT_TLS",
			ExitCode: ExInternetConfig,
			Advice:   "Ensure that your value for HTTPS_PROXY points to an HTTPS proxy rather than an HTTP proxy",
			Issues:   []int{7286},
			URL:      proxyDoc,
		},
		Regexp: re(`tls: first record does not look like a TLS handshake`),
	},
	{
		Kind: Kind{
			ID:       "INET_PROXY_503",
			ExitCode: ExInternetConfig,
			Advice:   "Confirm that you have a working internet connection and that your VM has not run out of resources by using: 'minikube logs'",
			Issues:   []int{4749},
		},
		Regexp: re(`proxy.*unexpected response code: 503`),
	},
	{
		Kind: Kind{
			ID:       "INET_DEFAULT_ROUTE",
			ExitCode: ExInternetNotFound,
			Advice:   "Configure a default route on this Linux host, or use another --driver that does not require it",
			Issues:   []int{6083, 5636},
		},
		Regexp: re(`(No|from) default routes`),
		GOOS:   []string{"linux"},
	},
}

var guestIssues = []match{
	{
		Kind: Kind{
			ID:       "GUEST_KVM2_NO_DOMAIN",
			ExitCode: ExGuestNotFound,
			Advice:   "The VM that minikube is configured for no longer exists. Run 'minikube delete'",
			Issues:   []int{3636},
		},
		Regexp: re(`no domain with matching name`),
		GOOS:   []string{"linux"},
	},
	{
		Kind: Kind{
			ID:       "GUEST_PORT_IN_USE",
			ExitCode: ExGuestConflict,
			Advice:   "kubeadm detected a TCP port conflict with another process: probably another local Kubernetes installation. Run lsof -p<port> to find the process and kill it",
			Issues:   []int{5484},
		},
		Regexp: re(`ERROR Port-.*is in use`),
		GOOS:   []string{"linux"},
	},

	{
		Kind: Kind{
			ID:       "GUEST_DOES_NOT_EXIST",
			ExitCode: ExGuestNotFound,
			Advice:   "Run 'minikube delete' to delete the stale VM, or and ensure that minikube is running as the same user you are issuing this command with",
			Issues:   []int{3864, 6087},
		},
		Regexp: re(`machine does not exist`),
	},
	{
		Kind: Kind{
			ID:       "GUEST_NOT_FOUND",
			ExitCode: ExGuestNotFound,
			Advice:   "Your minikube vm is not running, try minikube start.",
			Issues:   []int{4889},
		},
		Regexp: re(`Machine does not exist for api.Exists`),
	},
	{
		Kind: Kind{
			ID:       "GUEST_IP_NOT_FOUND",
			ExitCode: ExGuestNotRunning,
			Advice:   "The minikube VM is offline. Please run 'minikube start' to start it again.",
			Issues:   []int{3849, 3648},
		},
		Regexp: re(`Error getting ssh host name for driver: IP not found`),
	},
	{
		Kind: Kind{
			ID:       "GUEST_UNSIGNED_CERT",
			ExitCode: ExGuestConfig,
			Advice:   "Try 'minikube delete' to force new SSL certificates to be installed",
			Issues:   []int{6596},
		},
		Regexp: re(`not signed by CA certificate ca: crypto/rsa: verification error`),
	},
	{
		Kind: Kind{
			ID:       "GUEST_VBOX_NO_VM",
			ExitCode: ExGuestNotFound,
			Advice:   "The VM that minikube is configured for no longer exists. Run 'minikube delete'",
			Issues:   []int{4694},
		},
		Regexp: re(`Could not find a registered machine named`),
	},
	{
		Kind: Kind{
			ID:       "GUEST_FILE_IN_USE",
			ExitCode: ExGuestConflict,
			Advice:   "Another program is using a file required by minikube. If you are using Hyper-V, try stopping the minikube VM from within the Hyper-V manager",
			URL:      "https://docs.docker.com/machine/drivers/hyper-v/",
			Issues:   []int{7300},
		},
		Regexp: re(`The process cannot access the file because it is being used by another process`),
		GOOS:   []string{"windows"},
	},
	{
		Kind: Kind{
			ID:       "GUEST_NOT_FOUND",
			ExitCode: ExGuestNotFound,
			Advice:   "minikube is missing files relating to your guest environment. This can be fixed by running 'minikube delete'",
			Issues:   []int{9130},
		},
		Regexp: re(`config.json: The system cannot find the file specified`),
	},
	{
		Kind: Kind{
			ID:       "GUEST_SSH_CERT_NOT_FOUND",
			ExitCode: ExGuestNotFound,
			Advice:   "minikube is missing files relating to your guest environment. This can be fixed by running 'minikube delete'",
			Issues:   []int{9130},
		},
		Regexp: re(`id_rsa: no such file or directory`),
	},
	{
		Kind: Kind{
			ID:       "GUEST_CONFIG_CORRUPT",
			ExitCode: ExGuestConfig,
			Advice:   "The existing node configuration appears to be corrupt. Run 'minikube delete'",
			Issues:   []int{9175},
		},
		Regexp: re(`configuration.*corrupt`),
	},
}

// runtimeIssues are container runtime issues (containerd, docker, etc)
var runtimeIssues = []match{
	{
		Kind: Kind{
			ID:       "RT_DOCKER_RESTART",
			ExitCode: ExRuntimeError,
			Advice:   "Remove the invalid --docker-opt or --inecure-registry flag if one was provided",
			Issues:   []int{7070},
		},
		Regexp: re(`systemctl -f restart docker`),
	},
	{
		Kind: Kind{
			ID:       "RT_DOCKER_UNAVAILABLE",
			ExitCode: ExRuntimeUnavailable,
			Advice:   "Docker inside the VM is unavailable. Try running 'minikube delete' to reset the VM.",
			Issues:   []int{3952},
		},
		Regexp: re(`Error configuring auth on host: OS type not recognized`),
	},
	{
		Kind: Kind{
			ID:       "RT_DOCKER_EXIT_1",
			ExitCode: ExRuntimeNotFound,
			Advice:   "Either systemctl is not installed, or Docker is broken. Run 'sudo systemctl start docker' and 'journalctl -u docker'",
			URL:      "https://minikube.sigs.k8s.io/docs/reference/drivers/none",
			Issues:   []int{2704, 4498},
		},
		Regexp: re(`sudo systemctl start docker: exit status 1`),
		GOOS:   []string{"linux"},
	},
	{
		Kind: Kind{
			ID:       "RT_DOCKER_EXIT_5",
			ExitCode: ExRuntimeUnavailable,
			Advice:   "Ensure that Docker is installed and healthy: Run 'sudo systemctl start docker' and 'journalctl -u docker'. Alternatively, select another value for --driver",
			URL:      "https://minikube.sigs.k8s.io/docs/reference/drivers/none",
			Issues:   []int{5532},
		},
		Regexp: re(`sudo systemctl start docker: exit status 5`),
		GOOS:   []string{"linux"},
	},
	{
		Kind: Kind{
			ID:       "RT_CRIO_EXIT_5",
			ExitCode: ExRuntimeUnavailable,
			Advice:   "Ensure that CRI-O is installed and healthy: Run 'sudo systemctl start crio' and 'journalctl -u crio'. Alternatively, use --container-runtime=docker",
			URL:      "https://minikube.sigs.k8s.io/docs/reference/drivers/none",
			Issues:   []int{5532},
		},
		Regexp: re(`sudo systemctl restart crio: exit status 5`),
		GOOS:   []string{"linux"},
	},
}

// controlPlaneIssues are Kubernetes deployment issues
var controlPlaneIssues = []match{
	{
		Kind: Kind{
			ID:       "K8S_APISERVER_MISSING",
			ExitCode: ExControlPlaneNotFound,
			Advice:   "Check that the provided apiserver flags are valid, and that SELinux is disabled",
			Issues:   []int{4536, 6014},
		},
		Regexp: re(`apiserver process never appeared`),
	},
	{
		Kind: Kind{
			ID:       "K8S_APISERVER_TIMEOUT",
			ExitCode: ExControlPlaneTimeout,
			Advice:   "A VPN or firewall is interfering with HTTP access to the minikube VM. Alternatively, try a different VM driver: https://minikube.sigs.k8s.io/docs/start/",
			URL:      vpnDoc,
			Issues:   []int{4302},
		},
		Regexp: re(`apiserver: timed out waiting for the condition`),
	},
	{
		Kind: Kind{
			ID:       "K8S_DNS_TIMEOUT",
			ExitCode: ExControlPlaneTimeout,
			Advice:   "Run 'kubectl describe pod coredns -n kube-system' and check for a firewall or DNS conflict",
			URL:      vpnDoc,
		},
		Regexp: re(`dns: timed out waiting for the condition`),
	},
	{
		Kind: Kind{
			ID:       "K8S_KUBELET_NOT_RUNNING",
			ExitCode: ExControlPlaneUnavailable,
			Advice:   "Check output of 'journalctl -xeu kubelet', try passing --extra-config=kubelet.cgroup-driver=systemd to minikube start",
			Issues:   []int{4172},
		},
		Regexp: re(`The kubelet is not running|kubelet isn't running`),
		GOOS:   []string{"linux"},
	},
	{
		Kind: Kind{
			ID:       "K8S_INVALID_DNS_DOMAIN",
			ExitCode: ExControlPlaneConfig,
			Advice:   "Select a valid value for --dnsdomain",
		},
		Regexp: re(`dnsDomain: Invalid`),
	},
	{
		Kind: Kind{
			ID:           "K8S_INVALID_CERT_HOSTNAME",
			ExitCode:     ExControlPlaneConfig,
			Advice:       "The certificate hostname provided appears to be invalid (may be a minikube bug)",
			NewIssueLink: true,
		},
		Regexp: re(`apiServer.certSANs: Invalid value`),
	},
}

// serviceIssues are issues with services running on top of Kubernetes
var serviceIssues = []match{
	{
		Kind: Kind{
			ID:       "SVC_ENDPOINT_NOT_FOUND",
			ExitCode: ExSvcNotFound,
			Advice:   "Please make sure the service you are looking for is deployed or is in the correct namespace.",
			Issues:   []int{4599},
		},
		Regexp: re(`Could not find finalized endpoint being pointed to by`),
	},
	{
		Kind: Kind{
			ID:       "SVC_OPEN_NOT_FOUND",
			ExitCode: ExSvcNotFound,
			Advice:   "Use 'kubect get po -A' to find the correct and namespace name",
			Issues:   []int{5836},
		},
		Regexp: re(`Error opening service.*not found`),
	},
	{
		Kind: Kind{
			ID:       "SVC_DASHBOARD_ROLE_REF",
			ExitCode: ExSvcPermission,
			Advice:   "Run: 'kubectl delete clusterrolebinding kubernetes-dashboard'",
			Issues:   []int{7256},
		},
		Regexp: re(`dashboard.*cannot change roleRef`),
	},
}
