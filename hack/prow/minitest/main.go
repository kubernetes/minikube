package main

import (
	"minitest/deployer"
	"minitest/tester"

	"os"
	"flag"

	"k8s.io/klog/v2"
)

var deployers = map[string]func(string) deployer.MiniTestDeployer{
	"boskos": deployer.NewMiniTestBosKosDeployerFromConfigFile,
	"docker": deployer.NewMiniTestDockerDeployerFromConfigFile,
}
var testers = map[string]tester.MiniTestTester{
	"kvm-integration": &tester.KVMIntegrationTester{},
}

func main() {

	flagSet := flag.CommandLine
	deployerName := flagSet.String("deployer", "boskos", "deployer to use. Options: [boskos, docker]")
	config := flagSet.String("config", "", "path to deployer config file")
	testerName := flagSet.String("tester", "kvm-integration", "tester to use. Options: [kvm-integration]")
	klog.InitFlags(flagSet)
	flagSet.Parse(os.Args[1:])



	dep := getDeployer(*deployerName)(*config)
	tester := getTester(*testerName)

	if err := dep.Up(); err != nil {
		klog.Errorf("failed to start deployer: %v", err)
	}

	tester.Run(dep)
	
	if err := dep.Down(); err != nil {
		klog.Errorf("failed to stop deployer: %v", err)
	}

	// config := &deployer.MiniTestBoskosConfig{
	// 	GCPZone:                        "us-central1-b",
	// 	InstanceImage:                  "ubuntu-os-cloud/ubuntu-2404-lts-amd64",
	// 	InstanceType:                   "n2-standard-8",
	// 	DiskGiB:                        300,
	// 	BoskosAcquireTimeoutSeconds:    3 * 60,
	// 	BoskosHeartbeatIntervalSeconds: 10,
	// 	BoskosLocation:                 "http://boskos.test-pods.svc.cluster.local",
	// }
	// klog.Infof("Startring deployer with config %v", config)
	// if err := dep.Execute("sudo", "apt", "install", "-y", "rsync"); err != nil {
	// 	klog.Errorf("failed to execute command in docker deployer: %v", err)
	// }

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

func getDeployer(name string) func(string) deployer.MiniTestDeployer {
	d, ok := deployers[name]
	if !ok {
		klog.Fatalf("deployer %s not found. Available deployers: %v", name, deployers)
	}
	return d
}

func getTester(name string) tester.MiniTestTester {
	t, ok := testers[name]
	if !ok {
		klog.Fatalf("tester %s not found. Available testers: %v", name, testers)
	}
	return t
}
