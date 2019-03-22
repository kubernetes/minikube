package problem

import "regexp"

// r is a shortcut around regexp.MustCompile
func r(s string) *regexp.Regexp {
	return regexp.MustCompile(s)
}

// vmProblems are VM related problems
var vmProblems = map[string]match{
	"VBOX_NOT_FOUND": match{
		Regexp:   r(`VBoxManage not found. Make sure VirtualBox is installed and VBoxManage is in the path`),
		Solution: "Install VirtualBox, ensure that VBoxManage is executable and in path, or select an alternative value for --vm-driver",
		URL:      "https://www.virtualbox.org/wiki/Downloads",
		Issues:   []int{3784, 3776},
	},
	"VBOX_VTX_DISABLED": match{
		Regexp:   r(`This computer doesn't have VT-X/AMD-v enabled`),
		Solution: "In some environments, this message is incorrect. Try 'minikube start --no-vtx-check'",
		Issues:   []int{3900},
	},
	"KVM2_NOT_FOUND": match{
		Regexp:   r(`Driver "kvm2" not found. Do you have the plugin binary .* accessible in your PATH`),
		Solution: "Please install the minikube kvm2 VM driver, or select an alternative --vm-driver",
		URL:      "https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#kvm2-driver",
	},
	"KVM2_NO_IP": match{
		Regexp:   r(`Error starting stopped host: Machine didn't return an IP after 120 seconds`),
		Solution: "The KVM driver is unable to ressurect this old VM. Please run `minikube delete` to delete it and try again.",
		Issues:   []int{3901, 3566, 3434},
	},
	"VM_DOES_NOT_EXIST": match{
		Regexp:   r(`Error getting state for host: machine does not exist`),
		Solution: "Your system no longer knows about the VM previously created by minikube. Run 'minikube delete' to reset your local state.",
		Issues:   []int{3864},
	},
	"VM_IP_NOT_FOUND": match{
		Regexp:   r(`Error getting ssh host name for driver: IP not found`),
		Solution: "The minikube VM is offline. Please run 'minikube start' to start it again.",
		Issues:   []int{3849, 3648},
	},
}

// proxyDoc is the URL to proxy documentation
const proxyDoc = "https://github.com/kubernetes/minikube/blob/master/docs/http_proxy.md"

// netProblems are network related problems.
var netProblems = map[string]match{
	"GCR_UNAVAILABLE": match{
		Regexp:   r(`gcr.io.*443: connect: invalid argument`),
		Solution: "minikube is unable to access the Google Container Registry. You may need to configure it to use a HTTP proxy.",
		URL:      proxyDoc,
		Issues:   []int{3860},
	},
	"DOWNLOAD_RESET_BY_PEER": match{
		Regexp:   r(`Error downloading .*connection reset by peer`),
		Solution: "A firewall is likely blocking minikube from reaching the internet. You may need to configure minikube to use a proxy.",
		URL:      proxyDoc,
		Issues:   []int{3909},
	},
	"DOWNLOAD_IO_TIMEOUT": match{
		Regexp:   r(`Error downloading .*timeout`),
		Solution: "A firewall is likely blocking minikube from reaching the internet. You may need to configure minikube to use a proxy.",
		URL:      proxyDoc,
		Issues:   []int{3846},
	},
	"DOWNLOAD_TLS_OVERSIZED": match{
		Regexp:   r(`failed to download.*tls: oversized record received with length`),
		Solution: "A firewall is interfering with minikube's ability to make outgoing HTTPS requests. You may need to configure minikube to use a proxy.",
		URL:      proxyDoc,
		Issues:   []int{3857, 3759},
	},
	"ISO_DOWNLOAD_FAILED": match{
		Regexp:   r(`iso: failed to download`),
		Solution: "A firewall is likely blocking minikube from reaching the internet. You may need to configure minikube to use a proxy.",
		URL:      proxyDoc,
		Issues:   []int{3922},
	},
	"PULL_TIMEOUT_EXCEEDED": match{
		Regexp:   r(`failed to pull image k8s.gcr.io.*Client.Timeout exceeded while awaiting headers`),
		Solution: "A firewall is blocking Docker within the minikube VM from reaching the internet. You may need to configure it to use a proxy.",
		URL:      proxyDoc,
		Issues:   []int{3898},
	},
	"SSH_AUTH_FAILURE": match{
		Regexp:   r(`ssh: handshake failed: ssh: unable to authenticate.*, no supported methods remain`),
		Solution: "Your host is failing to route packets to the minikube VM. If you have VPN software, try turning it off or configuring it so that it does not re-route traffic to the VM IP. If not, check your VM environment routing options.",
		Issues:   []int{3930},
	},
	"SSH_TCP_FAILURE": match{
		Regexp:   r(`dial tcp .*:22: connectex: A connection attempt failed because the connected party did not properly respond`),
		Solution: "Your host is failing to route packets to the minikube VM. If you have VPN software, try turning it off or configuring it so that it does not re-route traffic to the VM IP. If not, check your VM environment routing options.",
		Issues:   []int{3388},
	},
}

// deployProblems are Kubernetes deployment problems.
var deployProblems = map[string]match{
	// This space intentionally left empty.

}
