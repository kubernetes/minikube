package deployer

type MiniTestBoskosConfig struct {
	GCPZone       string `desc:"GCP zone"`
	InstanceImage string `desc:"Instance image"`
	InstanceType  string `desc:"Instance type"`
	DiskGiB       int    `flag:"~disk-gib" desc:"Disk size in GiB"`

	// Boskos flags correspond to https://github.com/kubernetes-sigs/kubetest2/blob/71238a9645df6fbd7eaac9a36f635c22f1566168/kubetest2-gce/deployer/deployer.go
	BoskosAcquireTimeoutSeconds    int    `desc:"How long (in seconds) to hang on a request to Boskos to acquire a resource before erroring."`
	BoskosHeartbeatIntervalSeconds int    `desc:"How often (in seconds) to send a heartbeat to Boskos to hold the acquired resource. 0 means no heartbeat."`
	BoskosLocation                 string `desc:"If set, manually specifies the location of the boskos server. If unset and boskos is needed, defaults to http://boskos.test-pods.svc.cluster.local"`
}
