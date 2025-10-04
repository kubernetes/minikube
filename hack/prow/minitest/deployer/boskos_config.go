package deployer

type MiniTestBoskosConfig struct {
	GCPZone       string  
	InstanceImage string  
	InstanceType  string  
	DiskGiB       int     
	// Boskos flags correspond to https://github.com/kubernetes-sigs/kubetest2/blob/71238a9645df6fbd7eaac9a36f635c22f1566168/kubetest2-gce/deployer/deployer.go
	BoskosAcquireTimeoutSeconds    int     
	BoskosHeartbeatIntervalSeconds int    
	BoskosLocation                 string 
}
