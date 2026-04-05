//go:build darwin

/*
Copyright 2024 The Kubernetes Authors All rights reserved.

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

package krunkit

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/google/uuid"
	"k8s.io/minikube/pkg/libmachine/drivers"

	"k8s.io/minikube/pkg/drivers/common/virtiofs"
	"k8s.io/minikube/pkg/drivers/common/vmnet"
	"k8s.io/minikube/pkg/drivers/krunkit"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
	"k8s.io/minikube/pkg/minikube/run"
)

const (
	docURL = "https://minikube.sigs.k8s.io/docs/reference/drivers/krunkit/"
)

func init() {
	if err := registry.Register(registry.DriverDef{
		Name: driver.Krunkit,
		Init: func(options *run.CommandOptions) drivers.Driver {
			return krunkit.NewDriver("", "", options)
		},
		Config:   configure,
		Status:   status,
		Default:  true,
		Priority: registry.Experimental,
	}); err != nil {
		panic(fmt.Sprintf("register failed: %v", err))
	}
}

func configure(cfg config.ClusterConfig, n config.Node) (interface{}, error) {
	machineName := config.MachineName(cfg, n)
	storePath := localpath.MiniPath()

	// We generate a random UUID (or use a user provided one). vment-helper will
	// obtain a mac address from the vmnet framework using the UUID.
	u := cfg.UUID
	if u == "" {
		u = uuid.NewString()
	}

	mounts, err := virtiofs.ValidateMountString(cfg.MountString)
	if err != nil {
		return nil, err
	}

	return &krunkit.Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: machineName,
			StorePath:   storePath,
			SSHUser:     "docker",
		},
		Boot2DockerURL: download.LocalISOResource(cfg.MinikubeISO),
		DiskSize:       cfg.DiskSize,
		Memory:         cfg.Memory,
		CPU:            cfg.CPUs,
		ExtraDisks:     cfg.ExtraDisks,
		VirtiofsMounts: mounts,
		VmnetHelper: vmnet.Helper{
			MachineDir:  filepath.Join(storePath, "machines", machineName),
			InterfaceID: u,
		},
	}, nil
}

func status(options *run.CommandOptions) registry.State {
	if runtime.GOOS != "darwin" && runtime.GOARCH != "arm64" {
		err := errors.New("the krunkit driver is only supported on macOS arm64 machines")
		return registry.State{Error: err, Fix: "Use another driver", Doc: docURL}
	}
	if _, err := exec.LookPath("krunkit"); err != nil {
		return registry.State{Error: err, Fix: "Run 'brew tap slp/krunkit && brew install krunkit'", Doc: docURL}
	}
	if err := vmnet.ValidateHelper(options); err != nil {
		vmnetErr := err.(*vmnet.Error)
		return registry.State{Error: vmnetErr.Err, Fix: "Install and configure vment-helper", Doc: docURL}
	}
	return registry.State{Installed: true, Healthy: true, Running: true}
}
