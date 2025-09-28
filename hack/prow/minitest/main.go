package main

import (
	"minitest/deployer"

	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	config := &deployer.MiniTestBoskosConfig{
		GCPZone:                        "us-central1-b",
		InstanceImage:                  "ubuntu-os-cloud/ubuntu-2404-lts-amd64",
		InstanceType:                   "n2-standard-4",
		DiskGiB:                        300,
		BoskosAcquireTimeoutSeconds:    3 * 60,
		BoskosHeartbeatIntervalSeconds: 10,
		BoskosLocation:                 "http://boskos.test-pods.svc.cluster.local",
	}
	klog.Infof("Startring deployer with config %v", config)

	gcpDeployer := deployer.NewMiniTestBosKosDeployer(config)
	if err := gcpDeployer.Up(); err != nil {
		klog.Errorf("failed to start deployer: %v", err)
	}
	if err := gcpDeployer.Execute("sudo", "apt", "install", "-y", "rsync"); err != nil {
		klog.Errorf("failed to execute command in docker deployer: %v", err)
	}
	if err := gcpDeployer.Sync(".", "~/minikube"); err != nil {
		klog.Errorf("failed to sync file in docker deployer: %v", err)
	}
	if err := gcpDeployer.Execute("ls"); err != nil {
		klog.Errorf("failed to execute command in docker deployer: %v", err)
	}
	if err := gcpDeployer.Down(); err != nil {
		klog.Errorf("failed to stop deployer: %v", err)
	}

	// config := &deployer.MiniTestDockerConfig{
	// 	Image: "debian",
	// }
	// dockerDeployer := deployer.NewMiniTestDockerDeployer(config)
	// if err := dockerDeployer.Up(); err != nil {
	// 	klog.Errorf("failed to start docker deployer: %v", err)
	// }

	// if err := dockerDeployer.Execute("whoami"); err != nil {
	// 	klog.Errorf("failed to execute command in docker deployer: %v", err)
	// }
	// if err := dockerDeployer.Sync(".", "~/minikube"); err != nil {
	// 	klog.Errorf("failed to sync file in docker deployer: %v", err)
	// }

	// if err := dockerDeployer.Down(); err != nil {
	// 	klog.Errorf("failed to stop docker deployer: %v", err)
	// }
	// klog.Infof("Startring docker deployer with config %v", config)
}
