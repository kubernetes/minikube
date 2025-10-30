/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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

package main

import (
	"minitest/deployer"
	"minitest/tester"

	"flag"
	"os"

	"k8s.io/klog/v2"
)

var deployers = map[string]func(string) deployer.MiniTestDeployer{
	// boskos deployer will require a gcp project and start instances in it to run the test
	// the whole gcp project will be cleaned up after tests are done
	"boskos": deployer.NewMiniTestBosKosDeployerFromConfigFile,
	//docker deployer is for testing minitest. This should never be used for testing minikube
	"docker": deployer.NewMiniTestDockerDeployerFromConfigFile,
}
var testers = map[string]tester.MiniTestTester{
	"kvm-docker-linux-amd64-integration": &tester.KVMDockerLinuxAmd64IntegrationTester{},
}

func main() {

	flagSet := flag.CommandLine
	deployerName := flagSet.String("deployer", "boskos", "deployer to use. Options: [boskos, docker]")
	config := flagSet.String("config", "", "path to deployer config file")
	testerName := flagSet.String("tester", "kvm-docker-linux-amd64-integration", "tester to use. Options: [kvm-docker-linux-amd64-integration]")
	klog.InitFlags(flagSet)
	flagSet.Parse(os.Args[1:])

	dep := getDeployer(*deployerName)(*config)
	tester := getTester(*testerName)

	if err := dep.Up(); err != nil {
		klog.Fatalf("failed to start deployer: %v", err)
	}
	var testErr error
	if testErr = tester.Run(dep); testErr != nil {
		klog.Errorf("failed to run tests: %v", testErr)
	}

	if err := dep.Down(); err != nil {
		klog.Fatalf("failed to stop deployer: %v", err)
	}
	if testErr != nil {
		os.Exit(1)
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
