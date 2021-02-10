// +build darwin
// -build arm64

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
	"fmt"
	"os"

	"github.com/docker/machine/libmachine/drivers/plugin"
	"k8s.io/minikube/pkg/drivers/hyperkit"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println("version:", hyperkit.GetVersion())
		fmt.Println("commit:", hyperkit.GetGitCommitID())
		return
	}

	plugin.RegisterDriver(hyperkit.NewDriver("", ""))
}
