package main

import (
	"minitest/deployer"

	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	config := &deploy.MiniTestConfig{
		GCPZone:                        "us-central1-b",
		InstanceImage:                  "ubuntu-os-cloud/ubuntu-2404-lts-amd64",
		InstanceType:                   "n2-standard-4",
		DiskGiB:                        300,
		BoskosAcquireTimeoutSeconds:    3 * 60,
		BoskosHeartbeatIntervalSeconds: 10,
		BoskosLocation:                 "http://boskos.test-pods.svc.cluster.local",
	}
	klog.Infof("Startring deployer with config %v",config)

	deployer := deploy.NewMiniTestBosKosDeployer(config)
	if err := deployer.Up(); err != nil {
		klog.Errorf("failed to start deployer: %v", err)
	}
	if err := deployer.Down(); err != nil {
		klog.Errorf("failed to stop deployer: %v", err)
	}
}
