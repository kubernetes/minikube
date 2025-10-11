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
		klog.Fatalf("failed to start deployer: %v", err)
	}

	if err:=tester.Run(dep);err!=nil{
		klog.Fatalf("failed to run tests: %v", err)
	}
	
	if err := dep.Down(); err != nil {
		klog.Fatalf("failed to stop deployer: %v", err)
	}

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
