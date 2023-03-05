/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
	"os"

	"k8s.io/klog/v2"

	"k8s.io/minikube/cmd/minikube/util"
	"k8s.io/minikube/cmd/sudominikube/cmd"

	// Register drivers
	_ "k8s.io/minikube/pkg/minikube/registry/drvs"

	// Force exp dependency
	_ "golang.org/x/exp/ebnf"

	"github.com/google/slowjam/pkg/stacklog"
	"github.com/pkg/profile"

	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	_ "k8s.io/minikube/pkg/provision"
)

func main() {
	util.BridgeLogMessages()
	defer klog.Flush()

	util.PropagateDockerContextToEnv()

	util.SetFlags(true)

	s := stacklog.MustStartFromEnv("STACKLOG_PATH")
	defer s.Stop()

	if os.Getenv(util.MinikubeEnableProfile) == "1" {
		defer profile.Start(profile.TraceProfile).Stop()
	}
	if os.Getenv(constants.IsMinikubeChildProcess) == "" {
		machine.StartDriver()
	}
	out.SetOutFile(os.Stdout)
	out.SetErrFile(os.Stderr)
	cmd.Execute()
}
