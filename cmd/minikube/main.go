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
	"flag"
	"os"

	"github.com/pkg/profile"
	"github.com/spf13/pflag"
	"k8s.io/klog"
	"k8s.io/minikube/cmd/minikube/cmd"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	_ "k8s.io/minikube/pkg/provision"
)

const minikubeEnableProfile = "MINIKUBE_ENABLE_PROFILING"

func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "false")
	flag.Parse()

	klog.CopyStandardLogTo("INFO")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	if os.Getenv(minikubeEnableProfile) == "1" {
		defer profile.Start(profile.TraceProfile).Stop()
	}
	if os.Getenv(constants.IsMinikubeChildProcess) == "" {
		machine.StartDriver()
	}
	out.SetOutFile(os.Stdout)
	out.SetErrFile(os.Stderr)
	cmd.Execute()
	klog.Flush()
}
