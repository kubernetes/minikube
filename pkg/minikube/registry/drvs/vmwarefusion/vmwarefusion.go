// +build darwin

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

package vmwarefusion

import (
	"fmt"
	"os/exec"

	"github.com/docker/machine/drivers/vmwarefusion"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/pkg/errors"

	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
)

func init() {
	if err := registry.Register(registry.DriverDef{
		Name:     driver.VMwareFusion,
		Config:   configure,
		Status:   status,
		Init:     func() drivers.Driver { return vmwarefusion.NewDriver("", "") },
		Priority: registry.Deprecated,
	}); err != nil {
		panic(fmt.Sprintf("register: %v", err))
	}
}

func configure(config cfg.MachineConfig) interface{} {
	d := vmwarefusion.NewDriver(cfg.GetMachineName(), localpath.MiniPath()).(*vmwarefusion.Driver)
	d.Boot2DockerURL = config.Downloader.GetISOFileURI(config.MinikubeISO)
	d.Memory = config.Memory
	d.CPU = config.CPUs
	d.DiskSize = config.DiskSize

	// TODO(philips): push these defaults upstream to fixup this driver
	d.SSHPort = 22
	d.ISO = d.ResolveStorePath("boot2docker.iso")
	return d
}

func status() registry.State {
	_, err := exec.LookPath("vmrun")
	if err != nil {
		return registry.State{Error: errors.Wrap(err, "vmrun path check"), Fix: "Install VMWare Fusion", Doc: "https://minikube.sigs.k8s.io/docs/reference/drivers/vmwarefusion/"}
	}
	return registry.State{Installed: true, Healthy: true}
}
