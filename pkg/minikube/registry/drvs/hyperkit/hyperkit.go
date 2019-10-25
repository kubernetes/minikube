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

package hyperkit

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/pborman/uuid"

	"k8s.io/minikube/pkg/drivers/hyperkit"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
)

const (
	docURL = "https://minikube.sigs.k8s.io/docs/reference/drivers/hyperkit/"
)

func init() {
	if err := registry.Register(registry.DriverDef{
		Name:     driver.HyperKit,
		Config:   configure,
		Status:   status,
		Priority: registry.Preferred,
	}); err != nil {
		panic(fmt.Sprintf("register: %v", err))
	}
}

func configure(config cfg.MachineConfig) interface{} {
	u := config.UUID
	if u == "" {
		u = uuid.NewUUID().String()
	}

	return &hyperkit.Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: cfg.GetMachineName(),
			StorePath:   localpath.MiniPath(),
			SSHUser:     "docker",
		},
		Boot2DockerURL: config.Downloader.GetISOFileURI(config.MinikubeISO),
		DiskSize:       config.DiskSize,
		Memory:         config.Memory,
		CPU:            config.CPUs,
		NFSShares:      config.NFSShare,
		NFSSharesRoot:  config.NFSSharesRoot,
		UUID:           u,
		VpnKitSock:     config.HyperkitVpnKitSock,
		VSockPorts:     config.HyperkitVSockPorts,
		Cmdline:        "loglevel=3 console=ttyS0 console=tty0 noembed nomodeset norestore waitusb=10 systemd.legacy_systemd_cgroup_controller=yes random.trust_cpu=on hw_rng_model=virtio base host=" + cfg.GetMachineName(),
	}
}

func status() registry.State {
	path, err := exec.LookPath("hyperkit")
	if err != nil {
		return registry.State{Error: err, Fix: "Run 'brew install hyperkit'", Doc: docURL}
	}

	cmd := exec.Command(path, "-v")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return registry.State{Installed: true, Error: fmt.Errorf("%s failed:\n%s", strings.Join(cmd.Args, " "), out), Fix: "Run 'brew install hyperkit'", Doc: docURL}
	}

	return registry.State{Installed: true, Healthy: true}
}
