/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package qemu

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/docker/machine/libmachine/drivers"
	drvqemu "github.com/machine-drivers/docker-machine-driver-qemu"

	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
)

const (
	docURL = "https://minikube.sigs.k8s.io/docs/reference/drivers/qemu/"
)

func init() {
	if err := registry.Register(registry.DriverDef{
		Name:     driver.QEMU,
		Config:   configure,
		Status:   status,
		Default:  true,
		Priority: registry.Experimental,
	}); err != nil {
		panic(fmt.Sprintf("register failed: %v", err))
	}
}

func configure(cc config.ClusterConfig, n config.Node) (interface{}, error) {
	name := config.MachineName(cc, n)
	return drvqemu.Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: name,
			StorePath:   localpath.MiniPath(),
			SSHUser:     "docker",
		},
		Boot2DockerURL: download.LocalISOResource(cc.MinikubeISO),
		DiskSize:       cc.DiskSize,
		Memory:         cc.Memory,
		CPU:            cc.CPUs,
	}, nil
}

func status() registry.State {
	var qemuSystem string
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		qemuSystem = "qemu-system-x86_64"
	case "arm64":
		qemuSystem = "qemu-system-aarch64"
	default:
		return registry.State{Error: fmt.Errorf("unknown arch: %s", arch), Doc: docURL}
	}

	_, err := exec.LookPath(qemuSystem)
	if err != nil {
		return registry.State{Error: err, Fix: "Install qemu-system", Doc: docURL}
	}

	return registry.State{Installed: true, Healthy: true, Running: true}
}
