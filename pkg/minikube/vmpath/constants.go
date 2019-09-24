package vmpath

const (
	// GuestAddonsDir is the default path of the addons configration
	GuestAddonsDir = "/etc/kubernetes/addons"
	// GuestManifestsDir is where the kubelet should look for static Pod manifests
	GuestManifestsDir = "/etc/kubernetes/manifests"
	// GuestEphemeralDir is the path where ephemeral data should be stored within the VM
	GuestEphemeralDir = "/var/tmp/minikube"
	// GuestPersistentDir is the path where persistent data should be stored within the VM (not tmpfs)
	GuestPersistentDir = "/var/lib/minikube"
	// GuestCertsDir are where Kubernetes certificates are kept on the guest
	GuestCertsDir = GuestPersistentDir + "/certs"
)
